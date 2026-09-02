// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

// Package features is the canonical list of reconcilers built into the
// kropath-operator binary. Every reconciler listed here runs unconditionally;
// there are no per-feature flags.
//
// When adding a new reconciler:
//  1. Create its package under internal/reconciler/<pkg>/.
//  2. Add an entry to All below.
//  3. Run `make features-gen` to regenerate docs/features.yaml.
//  4. Failing to do step 2 causes TestRegistryCoversAllPackages to fail.
package features

// Reconciler describes a single reconciler built into the kropath-operator binary.
type Reconciler struct {
	// Name is the human-readable controller name (e.g. "KMSConfig").
	Name string `json:"name" yaml:"name"`
	// Package is the Go sub-package under internal/reconciler/ (e.g. "kmsconfig").
	// It is the unique, stable machine identifier for a reconciler.
	Package string `json:"package" yaml:"package"`
	// Description is a short human-readable summary of what the reconciler does.
	Description string `json:"description" yaml:"description"`
	// Kinds lists the CRD kinds this reconciler watches
	// (e.g. ["KMSConfig", "KropathConfig"]). Empty for LabelOperator, which
	// watches every already-registered config kind rather than a fixed set.
	Kinds []string `json:"kinds,omitempty" yaml:"kinds,omitempty"`
	// SinceVersion is the semver of the first release containing this reconciler.
	SinceVersion string `json:"sinceVersion" yaml:"sinceVersion"`
	// Stability is one of "alpha", "beta", or "stable".
	Stability string `json:"stability" yaml:"stability"`
	// State is "active" when the reconciler is running or "pending" when one or more
	// required CRDs are not yet installed. Omitted when the coordinator is unavailable
	// (e.g. the offline `features` subcommand).
	State string `json:"state,omitempty" yaml:"state,omitempty"`
	// MissingKinds lists the required CRD kinds not yet served, populated only when
	// State is "pending". Omitted when empty or when the coordinator is unavailable.
	MissingKinds []string `json:"missingKinds,omitempty" yaml:"missingKinds,omitempty"`
}

// cascade builds the entry for a standard <Family>Config cascade reconciler.
// Every one of them watches its own config kind plus KropathConfig, and carries
// the same description shape, so spelling that out 20 times would be noise.
func cascade(name, pkg string) Reconciler {
	return Reconciler{
		Name:         name,
		Package:      pkg,
		Description:  "Reconciles " + name + " CRs and propagates effective config.",
		Kinds:        []string{name, "KropathConfig"},
		SinceVersion: seedVersion,
		Stability:    "stable",
	}
}

// seedVersion is the release in which every reconciler present at the
// per-feature-flag migration first shipped. Reconcilers added after the
// migration get their own later version.
const seedVersion = "v0.0.1"

// All is the canonical list of reconcilers in the kropath-operator binary.
// Every entry must correspond to a directory in internal/reconciler/.
// Adding a directory there without adding an entry here causes
// TestRegistryCoversAllPackages to fail.
var All = []Reconciler{
	cascade("IAMConfig", "iamconfig"),
	cascade("S3Config", "s3config"),
	cascade("KMSConfig", "kmsconfig"),
	cascade("SQSConfig", "sqsconfig"),
	cascade("SecretsManagerConfig", "secretsmanagerconfig"),
	{
		Name:         "PolicyDocument",
		Package:      "policydocument",
		Description:  "Reconciles PolicyDocument CRs and resolves IAM policy references.",
		Kinds:        []string{"PolicyDocument", "KropathConfig"},
		SinceVersion: seedVersion,
		Stability:    "stable",
	},
	{
		Name:         "LabelOperator",
		Package:      "labeloperator",
		Description:  "Applies provider resource-name labels to all CRs across all provider API groups.",
		SinceVersion: seedVersion,
		Stability:    "stable",
	},
	cascade("SNSConfig", "snsconfig"),
	cascade("DynamoDBConfig", "dynamodbconfig"),
	cascade("EventBridgeConfig", "eventbridgeconfig"),
	cascade("CloudWatchLogsConfig", "cloudwatchlogsconfig"),
	cascade("CloudWatchConfig", "cloudwatchconfig"),
	cascade("ELBConfig", "elbconfig"),
	cascade("RDSConfig", "rdsconfig"),
	cascade("AutoScalingConfig", "autoscalingconfig"),
	cascade("ECSConfig", "ecsconfig"),
	cascade("EKSConfig", "eksconfig"),
	cascade("EC2Config", "ec2config"),
	cascade("ApiGatewayV2Config", "apigatewayv2config"),
	cascade("APIGatewayConfig", "apigatewayconfig"),
	cascade("EFSConfig", "efsconfig"),
	cascade("ElastiCacheConfig", "elasticacheconfig"),
	cascade("ECRConfig", "ecrconfig"),
	cascade("StepFunctionsConfig", "stepfunctionsconfig"),
	cascade("MSKConfig", "mskconfig"),
	cascade("MemoryDBConfig", "memorydbconfig"),
	cascade("ACMConfig", "acmconfig"),
	cascade("EMRConfig", "emrconfig"),
	cascade("DocumentDBConfig", "documentdbconfig"),
	cascade("GlueConfig", "glueconfig"),
	cascade("AthenaConfig", "athenaconfig"),
	cascade("DSQLConfig", "dsqlconfig"),
	cascade("Route53Config", "route53config"),
	cascade("SSMConfig", "ssmconfig"),
	cascade("CognitoConfig", "cognitoconfig"),
	cascade("KinesisConfig", "kinesisconfig"),
	cascade("CloudTrailConfig", "cloudtrailconfig"),
	cascade("AppScalingConfig", "appscalingconfig"),
	cascade("KeyspacesConfig", "keyspacesconfig"),
	cascade("WAFConfig", "wafconfig"),
	cascade("BedrockConfig", "bedrockconfig"),
	cascade("SageMakerConfig", "sagemakerconfig"),
}
