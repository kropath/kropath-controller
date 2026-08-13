// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package features_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/kropath/kropath-controller/internal/features"
)

// TestRegistryCoversAllPackages ensures that every directory under
// internal/reconciler/ is listed in features.All, and vice versa.
// This is the CI gate that makes "adding a reconciler without registering it"
// a failing test rather than a silent operational gap.
func TestRegistryCoversAllPackages(t *testing.T) {
	entries, err := os.ReadDir("../reconciler")
	if err != nil {
		t.Fatalf("reading internal/reconciler: %v", err)
	}

	dirs := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			dirs[e.Name()] = true
		}
	}

	registered := map[string]bool{}
	for _, r := range features.All {
		registered[r.Package] = true
	}

	for pkg := range dirs {
		if !registered[pkg] {
			t.Errorf("package %q exists in internal/reconciler/ but is missing from features.All — add it", pkg)
		}
	}
	for pkg := range registered {
		if !dirs[pkg] {
			t.Errorf("features.All lists package %q but internal/reconciler/%s/ does not exist — remove it or create the package", pkg, pkg)
		}
	}
}

// packagesWithoutOwnCRD lists reconcilers in features.All that do not own a
// primary <package>s.aws.kropath.run CRD. The label operator watches every
// already-registered config kind rather than a kind of its own.
var packagesWithoutOwnCRD = map[string]bool{
	"labeloperator": true,
}

// crdNameRE matches the `  name: <crd>` line of a CRD's metadata block.
var crdNameRE = regexp.MustCompile(`(?m)^  name: ([a-z0-9.]+)$`)

// TestEveryReconcilerHasCRDFixture ensures every reconciler in features.All has its
// CRD in tests/fixtures/crds/, which is what `make chainsaw-setup` applies to the
// kind cluster.
//
// This guards a failure mode that is expensive to diagnose from its symptom. Since
// per-feature flags were retired (KRO-635), every reconciler starts unconditionally.
// If a reconciler watches a kind whose CRD is absent from the cluster, its informer
// never syncs, and controller-runtime aborts the *entire manager* once the 2-minute
// cache-sync timeout elapses. The operator therefore serves traffic normally for two
// minutes and then exits, so the visible symptom is that every Chainsaw suite which
// happens to run after the ~2-minute mark fails on a 30s assert timeout — while the
// suites that ran before it pass. Which suites fail depends only on machine speed and
// test ordering, not on the resource families involved, which makes the failure look
// unrelated to its actual cause.
func TestEveryReconcilerHasCRDFixture(t *testing.T) {
	fixtures, err := filepath.Glob("../../tests/fixtures/crds/*.yaml")
	if err != nil {
		t.Fatalf("globbing tests/fixtures/crds: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no CRD fixtures found under tests/fixtures/crds/")
	}

	present := map[string]bool{}
	for _, f := range fixtures {
		data, err := os.ReadFile(f) //nolint:gosec // test-only read of a repo-relative fixture path
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		for _, m := range crdNameRE.FindAllStringSubmatch(string(data), -1) {
			present[m[1]] = true
		}
	}

	for _, r := range features.All {
		if packagesWithoutOwnCRD[r.Package] {
			continue
		}
		crd := r.Package + "s.aws.kropath.run"
		if !present[crd] {
			t.Errorf("reconciler %q (package %q) watches a kind with no CRD fixture: expected %q under tests/fixtures/crds/.\n"+
				"Without it the manager exits after the 2-minute cache-sync timeout and later Chainsaw suites fail on assert timeouts.\n"+
				"Copy the CRD from kropath-aws/crds/ into tests/fixtures/crds/, or add %q to packagesWithoutOwnCRD if it genuinely owns no CRD.",
				r.Name, r.Package, crd, r.Package)
		}
	}

	// Every config reconciler also reads KropathConfig, so that CRD must be present too.
	for _, crd := range []string{"kropathconfigs.aws.kropath.run", "kropathconfigs.kropath.run"} {
		if !present[crd] {
			t.Errorf("missing CRD fixture %q under tests/fixtures/crds/ — every config reconciler watches KropathConfig", crd)
		}
	}
}

// --- KRO-637 acceptance criteria -------------------------------------------
//
// The tests below restore guarantees specified by KRO-637 that were not
// carried over when KRO-635 landed the simpler static-slice registry.

// TestNoDuplicateEntries asserts that no Name or Package appears twice in
// features.All (KRO-637 AC-6).
//
// KRO-637's map-backed registry got this for free: Register panicked on a
// duplicate key. A static slice has no such guard, so a copy-paste mistake
// would otherwise be invisible — TestRegistryCoversAllPackages compares
// against a map and therefore collapses duplicates silently, and the
// duplicate would surface only as a reconciler started twice at runtime.
func TestNoDuplicateEntries(t *testing.T) {
	seenName := map[string]bool{}
	seenPkg := map[string]bool{}
	for _, r := range features.All {
		if seenName[r.Name] {
			t.Errorf("duplicate Name %q in features.All", r.Name)
		}
		if seenPkg[r.Package] {
			t.Errorf("duplicate Package %q in features.All", r.Package)
		}
		seenName[r.Name] = true
		seenPkg[r.Package] = true
	}
}

// TestAllOrderIsDeterministic asserts that repeated reads of features.All
// observe the same order (KRO-637 AC-5).
//
// AC-5 called for sorting by Name because KRO-637 backed the registry with a
// map, and any slice derived from Go map iteration has unstable order (see
// "Chainsaw Test Assertion Stability" in CLAUDE.md). The shipped registry is a
// package-level slice literal, so declaration order already gives the
// stability AC-5 exists to guarantee — docs/features.yaml and the /features
// JSON are byte-identical across runs. This test pins that invariant so a
// future change to a map-backed or lazily-built registry fails here rather
// than as an intermittent diff in generated output.
func TestAllOrderIsDeterministic(t *testing.T) {
	// Compared by Package, the unique stable key: features.Reconciler contains a
	// slice field and so is not comparable with ==.
	first := make([]string, len(features.All))
	for i, r := range features.All {
		first[i] = r.Package
	}
	for i := 0; i < 3; i++ {
		for j, r := range features.All {
			if r.Package != first[j] {
				t.Fatalf("features.All order changed between reads: index %d was %q, now %q", j, first[j], r.Package)
			}
		}
	}
}

// validStability is the closed set of Stability values from KRO-637's design.
var validStability = map[string]bool{"alpha": true, "beta": true, "stable": true}

// TestEveryReconcilerHasCompleteMetadata asserts every entry carries the
// descriptive fields KRO-637 specified, so /features and docs/features.yaml
// stay self-describing as reconcilers are added.
func TestEveryReconcilerHasCompleteMetadata(t *testing.T) {
	for _, r := range features.All {
		if r.Description == "" {
			t.Errorf("reconciler %q has an empty Description", r.Name)
		}
		if r.SinceVersion == "" {
			t.Errorf("reconciler %q has an empty SinceVersion", r.Name)
		}
		if !validStability[r.Stability] {
			t.Errorf("reconciler %q has Stability %q — want one of alpha/beta/stable", r.Name, r.Stability)
		}
	}
}

// wiredRE matches the two shapes main.go uses to start a reconciler:
// `&<pkg>.Reconciler{` for the struct-based ones and `<pkg>.Setup(` for
// labeloperator, which KRO-637 explicitly said not to refactor for uniformity.
func wiredRE(pkg string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(pkg) + `\.(Reconciler\{|Setup\()`)
}

// TestEveryRegisteredReconcilerIsWired asserts that every entry in features.All
// is actually started by cmd/manager/main.go.
//
// This closes the one functional gap left by swapping KRO-637's design for the
// static slice. In KRO-637 the registry *was* the wiring: All held a Setup
// closure and main.go called SetupAll, so /features could not disagree with
// what the binary ran. In the shipped design the slice is reporting-only and
// main.go wires each reconciler by hand, so an entry that is listed and has a
// package directory — passing TestRegistryCoversAllPackages — can still never
// be started. The endpoint would then advertise a reconciler that does not run.
func TestEveryRegisteredReconcilerIsWired(t *testing.T) {
	src, err := os.ReadFile("../../cmd/manager/main.go")
	if err != nil {
		t.Fatalf("reading cmd/manager/main.go: %v", err)
	}
	main := string(src)

	for _, r := range features.All {
		if !wiredRE(r.Package).MatchString(main) {
			t.Errorf("reconciler %q (package %q) is listed in features.All but never started in cmd/manager/main.go.\n"+
				"/features and docs/features.yaml would advertise a reconciler the binary does not run.\n"+
				"Add its setup call to main.go, or remove the entry from features.All.",
				r.Name, r.Package)
		}
	}
}
