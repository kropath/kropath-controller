// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/reconciler/iamconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/policydocument"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	metricsAddr     string
	probeAddr       string
	enablePolicyDoc bool
)

func main() {
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enablePolicyDoc, "enable-poldoc", false, "Enable the AWSPolicyDocument reconciler.")
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

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
		Scheme:                 sch,
		Metrics:                server.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         true,
		LeaderElectionID:       "kropath-controller.kropath.run",
	})
	if err != nil {
		ctrl.Log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&iamconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AWSIAMConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create IAM config controller")
		os.Exit(1)
	}

	if enablePolicyDoc {
		if err := (&policydocument.Reconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create AWSPolicyDocument reconciler")
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	ctrl.Log.Info("starting manager", "metrics", metricsAddr, "probes", probeAddr, "enable_poldoc", enablePolicyDoc)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
