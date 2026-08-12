// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package features_test

import (
	"os"
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
