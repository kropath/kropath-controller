// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	goruntime "runtime"
	"time"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/features"
	"github.com/kropath/kropath-controller/internal/registry"
	"github.com/kropath/kropath-controller/internal/version"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	metricsAddr string
	probeAddr   string
)

func leaderElectionNamespace() string {
	if ns := os.Getenv("LEADER_ELECTION_NAMESPACE"); ns != "" {
		return ns
	}
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	return "default"
}

func main() {
	// "features" subcommand: print the same JSON as GET /features, then exit.
	// No manager, no kubeconfig, no cluster connection — safe for CI and laptops.
	if len(os.Args) > 1 && os.Args[1] == "features" {
		resp := features.Response{
			Version:   version.Version,
			GitCommit: version.GitCommit,
			BuildDate: version.BuildDate,
			GoVersion: goruntime.Version(),
			Features:  features.All,
		}
		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		return
	}

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrl.Log.Info("starting manager",
		"version", version.Version,
		"gitCommit", version.GitCommit,
		"buildDate", version.BuildDate,
		"metrics", metricsAddr,
		"probes", probeAddr,
		"reconcilers", len(features.All),
	)

	sch := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(sch); err != nil {
		ctrl.Log.Error(err, "unable to add client-go scheme")
		os.Exit(1)
	}
	if err := v1alpha1.AddToScheme(sch); err != nil {
		ctrl.Log.Error(err, "unable to add api scheme")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  sch,
		Metrics:                 server.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          true,
		LeaderElectionNamespace: leaderElectionNamespace(),
		LeaderElectionID:        "kropath-controller.aws.kropath.run",
	})
	if err != nil {
		ctrl.Log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// /features endpoint: returns version + live reconciler list as JSON.
	if err := mgr.AddMetricsServerExtraHandler("/features", features.Handler(
		version.Version, version.GitCommit, version.BuildDate, goruntime.Version(), features.All,
	)); err != nil {
		ctrl.Log.Error(err, "unable to register /features handler")
		os.Exit(1)
	}

	// Build the reconciler registry and register every entry unconditionally.
	// Commit 2 (KRO-848) replaces RunUnconditional with RunGate once the
	// CRD availability startup gate is in place.
	coord := &registry.Coordinator{}
	for _, e := range registry.All() {
		coord.Add(e)
	}
	bctx := registry.BuildCtx{
		Manager: mgr,
		Log:     ctrl.Log,
	}
	if err := coord.RunUnconditional(bctx); err != nil {
		ctrl.Log.Error(err, "unable to register reconcilers")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", func(req *http.Request) error {
		ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
		defer cancel()
		if !mgr.GetCache().WaitForCacheSync(ctx) {
			return fmt.Errorf("informer cache not synced")
		}
		return nil
	}); err != nil {
		ctrl.Log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
