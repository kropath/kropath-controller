// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ---- helpers ---------------------------------------------------------------

func metricsNoopBuild(_ BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
	return nil, nil
}

func metricsBuildCtx() BuildCtx {
	return BuildCtx{Manager: nil, Log: logr.Discard()}
}

func readGauge(t *testing.T, vec *prometheus.GaugeVec, labels prometheus.Labels) float64 {
	t.Helper()
	g, err := vec.GetMetricWith(labels)
	if err != nil {
		t.Fatalf("GetMetricWith: %v", err)
	}
	var m dto.Metric
	if err := g.Write(&m); err != nil {
		t.Fatalf("metric Write: %v", err)
	}
	return m.GetGauge().GetValue()
}

func readCounter(t *testing.T, vec *prometheus.CounterVec, labels prometheus.Labels) float64 {
	t.Helper()
	c, err := vec.GetMetricWith(labels)
	if err != nil {
		t.Fatalf("GetMetricWith: %v", err)
	}
	var m dto.Metric
	if err := c.Write(&m); err != nil {
		t.Fatalf("metric Write: %v", err)
	}
	return m.GetCounter().GetValue()
}

// ---- tests -----------------------------------------------------------------

// TestMetrics_RunGate_PendingGauge verifies that RunGate sets
// kropath_reconciler_active=0 and kropath_reconciler_missing_kinds>0 for an
// entry whose Required GVKs are absent.
func TestMetrics_RunGate_PendingGauge(t *testing.T) {
	pkg := "mtest-pending-gauge"
	gvkBase := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestPending"}
	gvkMissing := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestMissing"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{gvkBase, gvkMissing},
		Build:    metricsNoopBuild,
	})

	servedGVKs := map[schema.GroupVersionKind]bool{
		KropathConfigGVK: true,
		gvkBase:          true,
		// gvkMissing is absent
	}
	if err := coord.RunGate(metricsBuildCtx(), servedGVKs); err != nil {
		t.Fatalf("RunGate: %v", err)
	}

	if got := readGauge(t, reconcilerActive, prometheus.Labels{"package": pkg}); got != 0 {
		t.Errorf("reconcilerActive: want 0 (pending), got %v", got)
	}
	if got := readGauge(t, reconcilerMissingKinds, prometheus.Labels{"package": pkg}); got != 1 {
		t.Errorf("reconcilerMissingKinds: want 1, got %v", got)
	}
}

// TestMetrics_RunGate_ActiveGauge verifies that RunGate sets
// kropath_reconciler_active=1 for an entry whose Required GVKs are all served.
func TestMetrics_RunGate_ActiveGauge(t *testing.T) {
	pkg := "mtest-active-gauge"
	gvk := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestActive"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{gvk},
		Build:    metricsNoopBuild,
	})

	servedGVKs := map[schema.GroupVersionKind]bool{
		KropathConfigGVK: true,
		gvk:              true,
	}
	if err := coord.RunGate(metricsBuildCtx(), servedGVKs); err != nil {
		t.Fatalf("RunGate: %v", err)
	}

	if got := readGauge(t, reconcilerActive, prometheus.Labels{"package": pkg}); got != 1 {
		t.Errorf("reconcilerActive: want 1 (active), got %v", got)
	}
	if got := readGauge(t, reconcilerMissingKinds, prometheus.Labels{"package": pkg}); got != 0 {
		t.Errorf("reconcilerMissingKinds: want 0, got %v", got)
	}
}

// TestMetrics_OnGVKServable_ActivationCounter verifies that OnGVKServable
// increments kropath_reconciler_activations_total and sets active=1 on activation.
func TestMetrics_OnGVKServable_ActivationCounter(t *testing.T) {
	pkg := "mtest-activation-ctr"
	gvk := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestActivated"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{gvk},
		Build:    metricsNoopBuild,
	})

	before := readCounter(t, reconcilerActivationsTotal, prometheus.Labels{"package": pkg})

	if err := coord.OnGVKServable(metricsBuildCtx(), gvk); err != nil {
		t.Fatalf("OnGVKServable: %v", err)
	}

	if got := readGauge(t, reconcilerActive, prometheus.Labels{"package": pkg}); got != 1 {
		t.Errorf("reconcilerActive: want 1, got %v", got)
	}
	if got := readGauge(t, reconcilerMissingKinds, prometheus.Labels{"package": pkg}); got != 0 {
		t.Errorf("reconcilerMissingKinds: want 0, got %v", got)
	}
	if delta := readCounter(t, reconcilerActivationsTotal, prometheus.Labels{"package": pkg}) - before; delta != 1 {
		t.Errorf("reconcilerActivationsTotal delta: want 1, got %v", delta)
	}
}

// TestMetrics_AddKindWatch_ErrorCounter verifies that a failing AddKindWatch
// increments kropath_crd_watch_errors_total{reason="watch_failed"}.
func TestMetrics_AddKindWatch_ErrorCounter(t *testing.T) {
	pkg := "mtest-watch-err"
	reqGVK := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestBase"}
	optGVK := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestOpt"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{reqGVK},
		Optional: []schema.GroupVersionKind{optGVK},
		Build:    metricsNoopBuild,
		AddKindWatch: func(_ controller.Controller, _ schema.GroupVersionKind) error {
			return fmt.Errorf("simulated watch attachment failure")
		},
	})

	// Activate the entry first.
	if err := coord.OnGVKServable(metricsBuildCtx(), reqGVK); err != nil {
		t.Fatalf("OnGVKServable (activate): %v", err)
	}

	before := readCounter(t, crdWatchErrorsTotal, prometheus.Labels{"reason": errReasonWatchFailed})

	// Trigger AddKindWatch which is expected to fail.
	err := coord.OnGVKServable(metricsBuildCtx(), optGVK)
	if err == nil {
		t.Fatal("expected error from failing AddKindWatch")
	}

	if delta := readCounter(t, crdWatchErrorsTotal, prometheus.Labels{"reason": errReasonWatchFailed}) - before; delta != 1 {
		t.Errorf("crdWatchErrorsTotal delta: want 1, got %v", delta)
	}
}
