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
