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
	"github.com/kropath/kropath-controller/internal/reconciler/apigatewayconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/apigatewayv2config"
	"github.com/kropath/kropath-controller/internal/reconciler/autoscalingconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/cloudwatchconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/cloudwatchlogsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/dynamodbconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/ec2config"
	"github.com/kropath/kropath-controller/internal/reconciler/ecrconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/ecsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/efsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/eksconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/elasticacheconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/elbconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/memorydbconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/eventbridgeconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/iamconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/kmsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/labeloperator"
	"github.com/kropath/kropath-controller/internal/reconciler/policydocument"
	"github.com/kropath/kropath-controller/internal/reconciler/rdsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/s3config"
	"github.com/kropath/kropath-controller/internal/reconciler/secretsmanagerconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/snsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/sqsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/mskconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/stepfunctionsconfig"
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

	// Every reconciler runs unconditionally — no per-feature flags.
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

	if err := (&policydocument.Reconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create PolicyDocument reconciler")
		os.Exit(1)
	}

	if err := (&kmsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("KMSConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create KMSConfig reconciler")
		os.Exit(1)
	}

	if err := (&sqsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SQSConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create SQSConfig reconciler")
		os.Exit(1)
	}

	if err := (&secretsmanagerconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SecretsManagerConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create SecretsManagerConfig reconciler")
		os.Exit(1)
	}

	if err := labeloperator.Setup(mgr, ctrl.Log.WithName("controllers").WithName("LabelOperator")); err != nil {
		ctrl.Log.Error(err, "unable to setup label-operator")
		os.Exit(1)
	}

	if err := (&snsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SNSConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create SNSConfig reconciler")
		os.Exit(1)
	}

	if err := (&dynamodbconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DynamoDBConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create DynamoDBConfig reconciler")
		os.Exit(1)
	}

	if err := (&eventbridgeconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("EventBridgeConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create EventBridgeConfig reconciler")
		os.Exit(1)
	}

	if err := (&cloudwatchlogsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CloudWatchLogsConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create CloudWatchLogsConfig reconciler")
		os.Exit(1)
	}

	if err := (&cloudwatchconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CloudWatchConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create CloudWatchConfig reconciler")
		os.Exit(1)
	}

	if err := (&elbconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AWSELBConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create AWSELBConfig reconciler")
		os.Exit(1)
	}

	if err := (&rdsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("RDSConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create RDSConfig reconciler")
		os.Exit(1)
	}

	if err := (&autoscalingconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AutoScalingConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create AutoScalingConfig reconciler")
		os.Exit(1)
	}

	if err := (&ecsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ECSConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create ECSConfig reconciler")
		os.Exit(1)
	}

	if err := (&eksconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("EKSConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create EKSConfig reconciler")
		os.Exit(1)
	}

	if err := (&ec2config.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("EC2Config"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create EC2Config reconciler")
		os.Exit(1)
	}

	if err := (&apigatewayv2config.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ApiGatewayV2Config"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create ApiGatewayV2Config reconciler")
		os.Exit(1)
	}

	if err := (&apigatewayconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("APIGatewayConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create APIGatewayConfig reconciler")
		os.Exit(1)
	}

	if err := (&efsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("EFSConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create EFSConfig reconciler")
		os.Exit(1)
	}

	if err := (&elasticacheconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ElastiCacheConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create ElastiCacheConfig reconciler")
		os.Exit(1)
	}

	if err := (&ecrconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ECRConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create ECRConfig reconciler")
		os.Exit(1)
	}

	if err := (&stepfunctionsconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("StepFunctionsConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create StepFunctionsConfig reconciler")
		os.Exit(1)
	}

	if err := (&mskconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MSKConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create MSKConfig reconciler")
		os.Exit(1)
	}

	if err := (&memorydbconfig.Reconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MemoryDBConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create MemoryDBConfig reconciler")
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
