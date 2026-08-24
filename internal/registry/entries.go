// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/internal/reconciler/acmconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/documentdbconfig"
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
	"github.com/kropath/kropath-controller/internal/reconciler/emrconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/eventbridgeconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/iamconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/kmsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/labeloperator"
	"github.com/kropath/kropath-controller/internal/reconciler/memorydbconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/mskconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/policydocument"
	"github.com/kropath/kropath-controller/internal/reconciler/rdsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/s3config"
	"github.com/kropath/kropath-controller/internal/reconciler/secretsmanagerconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/snsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/sqsconfig"
	"github.com/kropath/kropath-controller/internal/reconciler/stepfunctionsconfig"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const awsGroup = "aws.kropath.run"
const awsVersion = "v1alpha1"

func awsGVK(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: awsGroup, Version: awsVersion, Kind: kind}
}

// policyDocumentRefGVKs are the six optional GVKs that policydocument watches for
// ref resolution. Commit 2 (gate) will gate some of these on CRD availability.
var policyDocumentRefGVKs = []schema.GroupVersionKind{
	awsGVK("AWSIAMRole"),
	awsGVK("AWSS3Bucket"),
	awsGVK("AWSLambdaFunction"),
	awsGVK("AWSSQSQueue"),
	awsGVK("AWSKMSKey"),
	awsGVK("AWSSecretsManagerSecret"),
}

// All returns the complete registry in the same order as features.All.
func All() []Entry {
	// pdr is declared here so both the Build and AddKindWatch closures for
	// policydocument capture the same pointer variable. Build sets it on first
	// call; AddKindWatch reads it thereafter. The coordinator mutex guarantees
	// that Build completes before any AddKindWatch call is dispatched.
	var pdr *policydocument.Reconciler

	// labelop* vars are captured by both the labeloperator Build and AddKindWatch
	// closures. Build populates all three; AddKindWatch reads them.
	var (
		labelopHandles map[string]controller.Controller
		labelopMgr     ctrl.Manager
		labelopLog     logr.Logger
	)

	return []Entry{
		cascadeEntry("iamconfig", "IAMConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&iamconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("IAMConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("s3config", "S3Config", func(bctx BuildCtx) (controller.Controller, error) {
			return (&s3config.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("S3Config"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("kmsconfig", "KMSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&kmsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("KMSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("sqsconfig", "SQSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&sqsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SQSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("secretsmanagerconfig", "SecretsManagerConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&secretsmanagerconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SecretsManagerConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		// PolicyDocument watches its own kind plus six optional ref GVKs.
		// The closure captures pdr declared above so AddKindWatch can find the
		// already-built reconciler without a separate lookup.
		{
			Package:  "policydocument",
			Required: []schema.GroupVersionKind{awsGVK("PolicyDocument"), awsGVK("KropathConfig")},
			Optional: policyDocumentRefGVKs,
			Build: func(bctx BuildCtx, servedOptional []schema.GroupVersionKind) (controller.Controller, error) {
				pdr = &policydocument.Reconciler{}
				return pdr.BuildWithManager(bctx.Manager, servedOptional)
			},
			AddKindWatch: func(c controller.Controller, gvk schema.GroupVersionKind) error {
				if pdr == nil {
					return fmt.Errorf("policydocument: AddKindWatch called before Build")
				}
				return pdr.AddKindWatch(c, gvk)
			},
		},
		// LabelOperator creates one controller per provider GVK at startup, and registers
		// new controllers at runtime via AddKindWatch when new provider CRDs appear.
		// Optional == nil signals wildcard: any provider GVK triggers AddKindWatch.
		{
			Package:  "labeloperator",
			Required: nil,
			Optional: nil, // wildcard: handled by registry.OnGVKServable
			Build: func(bctx BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
				log := bctx.Log.WithName("controllers").WithName("LabelOperator")
				handles, err := labeloperator.Setup(bctx.Manager, log)
				if err != nil {
					return nil, fmt.Errorf("label-operator setup: %w", err)
				}
				labelopHandles = handles
				labelopMgr = bctx.Manager
				labelopLog = log
				return nil, nil
			},
			AddKindWatch: func(_ controller.Controller, gvk schema.GroupVersionKind) error {
				if labelopHandles == nil {
					return fmt.Errorf("labeloperator: AddKindWatch called before Build")
				}
				if labeloperator.LabelKeyForGroup(gvk.Group) == "" {
					return nil // not a watched provider group
				}
				key := gvk.String()
				if labelopHandles[key] != nil {
					return nil // controller already registered at startup
				}
				handle, err := labeloperator.RegisterController(
					labelopMgr, labeloperator.ControllerName(gvk), gvk, labelopLog,
				)
				if err != nil {
					return fmt.Errorf("label-operator: register runtime controller for %v: %w", gvk, err)
				}
				labelopHandles[key] = handle
				return nil
			},
		},
		cascadeEntry("snsconfig", "SNSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&snsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SNSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("dynamodbconfig", "DynamoDBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&dynamodbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("DynamoDBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("eventbridgeconfig", "EventBridgeConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&eventbridgeconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EventBridgeConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("cloudwatchlogsconfig", "CloudWatchLogsConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&cloudwatchlogsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CloudWatchLogsConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("cloudwatchconfig", "CloudWatchConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&cloudwatchconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CloudWatchConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("elbconfig", "ELBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&elbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("AWSELBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("rdsconfig", "RDSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&rdsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("RDSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("autoscalingconfig", "AutoScalingConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&autoscalingconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("AutoScalingConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ecsconfig", "ECSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ecsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ECSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("eksconfig", "EKSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&eksconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EKSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ec2config", "EC2Config", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ec2config.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EC2Config"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("apigatewayv2config", "ApiGatewayV2Config", func(bctx BuildCtx) (controller.Controller, error) {
			return (&apigatewayv2config.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ApiGatewayV2Config"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("apigatewayconfig", "APIGatewayConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&apigatewayconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("APIGatewayConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("efsconfig", "EFSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&efsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EFSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("elasticacheconfig", "ElastiCacheConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&elasticacheconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ElastiCacheConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ecrconfig", "ECRConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ecrconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ECRConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("stepfunctionsconfig", "StepFunctionsConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&stepfunctionsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("StepFunctionsConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("mskconfig", "MSKConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&mskconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("MSKConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("memorydbconfig", "MemoryDBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&memorydbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("MemoryDBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("acmconfig", "ACMConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&acmconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ACMConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("emrconfig", "EMRConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&emrconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EMRConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("documentdbconfig", "DocumentDBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&documentdbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("DocumentDBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
	}
}

// cascadeEntry builds a standard cascade reconciler entry. Every cascade reconciler
// watches its own <Family>Config plus KropathConfig. The inner build function does
// not receive servedOptional because cascade reconcilers have no Optional GVKs.
func cascadeEntry(pkg, kind string, build func(BuildCtx) (controller.Controller, error)) Entry {
	return Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{awsGVK(kind), awsGVK("KropathConfig")},
		Optional: nil,
		Build: func(bctx BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
			return build(bctx)
		},
		AddKindWatch: nil,
	}
}
