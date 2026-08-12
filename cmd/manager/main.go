// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/reconciler/apigatewayv2config"
	"github.com/kropath/kropath-controller/internal/reconciler/autoscalingconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/ecsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/efsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/eksconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/cloudwatchlogsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/dynamodbconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/elbconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/eventbridgeconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/iamconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/kmsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/labeloperator"
	"github.com/kropath/kropath-controller/internal/reconciler/policydocument"
	"github.com/kropath/kropath-controller/internal/reconciler/ec2config"
	"github.com/kropath/kropath-controller/internal/reconciler/rdsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/s3config"
	"github.com/kropath/kropath-controller/internal/reconciler/secretsmanagerconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/snsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/sqsconfig"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	metricsAddr              string
	probeAddr                string
	enablePolicyDoc          bool
	enableKMSCascade         bool
	enableSQSCascade         bool
	enableSMCascade          bool
	enableLabelOperator      bool
	enableSNSCascade           bool
	enableDynamoDBCascade      bool
	enableEventBridgeCascade      bool
	enableCloudWatchLogsCascade   bool
	enableELBCascade              bool
	enableRDSCascade              bool
	enableAutoScalingCascade      bool
	enableECSCascade              bool
	enableEKSCascade              bool
	enableEC2Cascade              bool
	enableApiGatewayV2Cascade     bool
	enableEFSCascade              bool
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
	flag.BoolVar(&enableSNSCascade, "enable-sns-cascade", false, "Enable the SNSConfig cascade reconciler.")
	flag.BoolVar(&enableDynamoDBCascade, "enable-dynamodb-cascade", false, "Enable the DynamoDBConfig cascade reconciler.")
	flag.BoolVar(&enableEventBridgeCascade, "enable-eventbridge-cascade", false, "Enable the EventBridgeConfig cascade reconciler.")
	flag.BoolVar(&enableCloudWatchLogsCascade, "enable-cloudwatchlogs-cascade", false, "Enable the CloudWatchLogsConfig cascade reconciler.")
	flag.BoolVar(&enableELBCascade, "enable-elb-cascade", false, "Enable the AWSELBConfig cascade reconciler.")
	flag.BoolVar(&enableRDSCascade, "enable-rds-cascade", false, "Enable the RDSConfig cascade reconciler.")
	flag.BoolVar(&enableAutoScalingCascade, "enable-autoscaling-cascade", false, "Enable the AutoScalingConfig cascade reconciler.")
	flag.BoolVar(&enableECSCascade, "enable-ecs-cascade", false, "Enable the ECSConfig cascade reconciler.")
	flag.BoolVar(&enableEKSCascade, "enable-eks-cascade", false, "Enable the EKSConfig cascade reconciler.")
	flag.BoolVar(&enableEC2Cascade, "enable-ec2-cascade", false, "Enable the EC2Config cascade reconciler.")
	flag.BoolVar(&enableApiGatewayV2Cascade, "enable-apigatewayv2-cascade", false, "Enable the ApiGatewayV2Config cascade reconciler.")
	flag.BoolVar(&enableEFSCascade, "enable-efs-cascade", false, "Enable the EFSConfig cascade reconciler.")
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

	if enableSNSCascade {
		if err := (&snsconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("SNSConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create SNSConfig reconciler")
			os.Exit(1)
		}
	}

	if enableDynamoDBCascade {
		if err := (&dynamodbconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("DynamoDBConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create DynamoDBConfig reconciler")
			os.Exit(1)
		}
	}

	if enableEventBridgeCascade {
		if err := (&eventbridgeconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("EventBridgeConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create EventBridgeConfig reconciler")
			os.Exit(1)
		}
	}

	if enableCloudWatchLogsCascade {
		if err := (&cloudwatchlogsconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("CloudWatchLogsConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create CloudWatchLogsConfig reconciler")
			os.Exit(1)
		}
	}

	if enableELBCascade {
		if err := (&elbconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("AWSELBConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create AWSELBConfig reconciler")
			os.Exit(1)
		}
	}

	if enableRDSCascade {
		if err := (&rdsconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("RDSConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create RDSConfig reconciler")
			os.Exit(1)
		}
	}

	if enableAutoScalingCascade {
		if err := (&autoscalingconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("AutoScalingConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create AutoScalingConfig reconciler")
			os.Exit(1)
		}
	}

	if enableECSCascade {
		if err := (&ecsconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("ECSConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create ECSConfig reconciler")
			os.Exit(1)
		}
	}

	if enableEKSCascade {
		if err := (&eksconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("EKSConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create EKSConfig reconciler")
			os.Exit(1)
		}
	}

	if enableEC2Cascade {
		if err := (&ec2config.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("EC2Config"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create EC2Config reconciler")
			os.Exit(1)
		}
	}

	if enableApiGatewayV2Cascade {
		if err := (&apigatewayv2config.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("ApiGatewayV2Config"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create ApiGatewayV2Config reconciler")
			os.Exit(1)
		}
	}

	if enableEFSCascade {
		if err := (&efsconfig.Reconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("EFSConfig"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			ctrl.Log.Error(err, "unable to create EFSConfig reconciler")
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

	ctrl.Log.Info("starting manager", "metrics", metricsAddr, "probes", probeAddr, "enable_poldoc", enablePolicyDoc, "enable_kms_cascade", enableKMSCascade, "enable_sqs_cascade", enableSQSCascade, "enable_secretsmanager_cascade", enableSMCascade, "enable_label_operator", enableLabelOperator, "enable_sns_cascade", enableSNSCascade, "enable_dynamodb_cascade", enableDynamoDBCascade, "enable_eventbridge_cascade", enableEventBridgeCascade, "enable_elb_cascade", enableELBCascade, "enable_rds_cascade", enableRDSCascade, "enable_autoscaling_cascade", enableAutoScalingCascade, "enable_ecs_cascade", enableECSCascade, "enable_eks_cascade", enableEKSCascade, "enable_ec2_cascade", enableEC2Cascade, "enable_apigatewayv2_cascade", enableApiGatewayV2Cascade, "enable_efs_cascade", enableEFSCascade)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
