// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

// Package features is the canonical list of reconcilers built into the
// kropath-operator binary. Every reconciler listed here runs unconditionally;
// there are no per-feature flags.
//
// When adding a new reconciler:
//  1. Create its package under internal/reconciler/<pkg>/.
//  2. Add an entry to All below.
//  3. Run `make generate-features` to regenerate docs/features.yaml.
//  4. Failing to do step 2 causes TestRegistryCoversAllPackages to fail.
package features

// Reconciler describes a single reconciler built into the kropath-operator binary.
type Reconciler struct {
	// Name is the human-readable controller name (e.g. "KMSConfig").
	Name string `json:"name" yaml:"name"`
	// Package is the Go sub-package under internal/reconciler/ (e.g. "kmsconfig").
	Package string `json:"package" yaml:"package"`
}

// All is the canonical list of reconcilers in the kropath-operator binary.
// Every entry must correspond to a directory in internal/reconciler/.
// Adding a directory there without adding an entry here causes
// TestRegistryCoversAllPackages to fail.
var All = []Reconciler{
	{Name: "IAMConfig", Package: "iamconfig"},
	{Name: "S3Config", Package: "s3config"},
	{Name: "KMSConfig", Package: "kmsconfig"},
	{Name: "SQSConfig", Package: "sqsconfig"},
	{Name: "SecretsManagerConfig", Package: "secretsmanagerconfig"},
	{Name: "PolicyDocument", Package: "policydocument"},
	{Name: "LabelOperator", Package: "labeloperator"},
	{Name: "SNSConfig", Package: "snsconfig"},
	{Name: "DynamoDBConfig", Package: "dynamodbconfig"},
	{Name: "EventBridgeConfig", Package: "eventbridgeconfig"},
	{Name: "CloudWatchLogsConfig", Package: "cloudwatchlogsconfig"},
	{Name: "ELBConfig", Package: "elbconfig"},
	{Name: "RDSConfig", Package: "rdsconfig"},
	{Name: "AutoScalingConfig", Package: "autoscalingconfig"},
	{Name: "ECSConfig", Package: "ecsconfig"},
	{Name: "EKSConfig", Package: "eksconfig"},
	{Name: "EC2Config", Package: "ec2config"},
	{Name: "ApiGatewayV2Config", Package: "apigatewayv2config"},
	{Name: "EFSConfig", Package: "efsconfig"},
	{Name: "ElastiCacheConfig", Package: "elasticacheconfig"},
	{Name: "ECRConfig", Package: "ecrconfig"},
	{Name: "StepFunctionsConfig", Package: "stepfunctionsconfig"},
}
