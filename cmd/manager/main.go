// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/reconciler/iamconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/kmsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/labeloperator"
	"github.com/kropath/kropath-controller/internal/reconciler/policydocument"
	"github.com/kropath/kropath-controller/internal/reconciler/s3config"
	"github.com/kropath/kropath-controller/internal/reconciler/secretsmanagerconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/sqsconfig"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	metricsAddr         string
	probeAddr           string
	enablePolicyDoc       bool
	enableKMSCascade      bool
	enableSQSCascade      bool
	enableSMCascade       bool
	enableLabelOperator   bool
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
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enablePolicyDoc, "enable-poldoc", false, "Enable the PolicyDocument reconciler.")
	flag.BoolVar(&enableKMSCascade, "enable-kms-cascade", false, "Enable the KMSConfig cascade reconciler.")
	flag.BoolVar(&enableSQSCascade, "enable-sqs-cascade", false, "Enable the SQSConfig cascade reconciler.")
	flag.BoolVar(&enableSMCascade, "enable-secretsmanager-cascade", false, "Enable the SecretsManagerConfig cascade reconciler.")
	flag.BoolVar(&enableLabelOperator, "enable-label-operator", false, "Enable the label-operator reconciler.")
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

	if err := (&iamconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IAMConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create IAM config controller")
		os.Exit(1)
	}

	if err := (&s3config.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("S3Config"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create S3 config controller")
		os.Exit(1)
	}

	if enablePolicyDoc {
		if err := (&policydocument.Reconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create PolicyDocument reconciler")
			os.Exit(1)
		}
	}

	if enableKMSCascade {
		if err := (&kmsconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("KMSConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create KMSConfig reconciler")
			os.Exit(1)
		}
	}

	if enableSQSCascade {
		if err := (&sqsconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("SQSConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create SQSConfig reconciler")
			os.Exit(1)
		}
	}

	if enableSMCascade {
		if err := (&secretsmanagerconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("SecretsManagerConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create SecretsManagerConfig reconciler")
			os.Exit(1)
		}
	}

	if enableLabelOperator {
		if err := labeloperator.Setup(mgr, ctrl.Log.WithName("controllers").WithName("LabelOperator")); err != nil {
			ctrl.Log.Error(err, "unable to setup label-operator")
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

	ctrl.Log.Info("starting manager", "metrics", metricsAddr, "probes", probeAddr, "enable_poldoc", enablePolicyDoc, "enable_kms_cascade", enableKMSCascade, "enable_sqs_cascade", enableSQSCascade, "enable_secretsmanager_cascade", enableSMCascade, "enable_label_operator", enableLabelOperator)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
