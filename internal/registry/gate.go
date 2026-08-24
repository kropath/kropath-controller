// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// KropathConfigGVK is the GVK that must be served before any reconciler starts.
// It is an install prerequisite — the operator exits non-zero if absent.
var KropathConfigGVK = schema.GroupVersionKind{
	Group:   "aws.kropath.run",
	Version: "v1alpha1",
	Kind:    "KropathConfig",
}

// ResourceLister is the subset of discovery.DiscoveryInterface used by GatherServedGVKs.
// Using a narrow interface makes the function easy to stub in tests.
type ResourceLister interface {
	ServerPreferredResources() ([]*metav1.APIResourceList, error)
}

// GatherServedGVKs queries the API server and returns the set of GVKs that are
// currently served (i.e. have an established CRD with this version available).
// Partial failures from ServerPreferredResources are ignored as long as at
// least one resource list is returned.
func GatherServedGVKs(dc ResourceLister) (map[schema.GroupVersionKind]bool, error) {
	resourceLists, err := dc.ServerPreferredResources()
	if err != nil && len(resourceLists) == 0 {
		return nil, fmt.Errorf("discovery: listing server resources: %w", err)
	}
	served := make(map[schema.GroupVersionKind]bool, len(resourceLists)*10)
	for _, rl := range resourceLists {
		gv, parseErr := schema.ParseGroupVersion(rl.GroupVersion)
		if parseErr != nil {
			continue
		}
		for _, r := range rl.APIResources {
			served[schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: r.Kind}] = true
		}
	}
	return served, nil
}

// RunGate registers each entry whose Required GVKs are all present in servedGVKs,
// leaving the rest pending. Before the per-entry loop it checks the KropathConfig
// precondition: if KropathConfig is not served, RunGate returns an error — the
// caller should treat this as fatal and exit non-zero.
func (c *Coordinator) RunGate(bctx BuildCtx, servedGVKs map[schema.GroupVersionKind]bool) error {
	if !servedGVKs[KropathConfigGVK] {
		return fmt.Errorf("startup gate: %s/%s %s is not served — "+
			"install the KropathConfig CRD before starting kropath-controller",
			KropathConfigGVK.Group, KropathConfigGVK.Version, KropathConfigGVK.Kind)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.entries {
		es := &c.entries[i]
		if es.active {
			continue
		}

		missing := 0
		for _, gvk := range es.entry.Required {
			if !servedGVKs[gvk] {
				missing++
			}
		}

		reconcilerMissingKinds.WithLabelValues(es.entry.Package).Set(float64(missing))

		if missing > 0 {
			reconcilerActive.WithLabelValues(es.entry.Package).Set(0)
			bctx.Log.Info("reconciler pending: required CRDs not yet served",
				"package", es.entry.Package,
				"missingCount", missing)
			continue
		}

		handle, err := es.entry.Build(bctx)
		if err != nil {
			return fmt.Errorf("registry: building %s: %w", es.entry.Package, err)
		}
		es.handle = handle
		es.active = true
		reconcilerActive.WithLabelValues(es.entry.Package).Set(1)
	}
	return nil
}
