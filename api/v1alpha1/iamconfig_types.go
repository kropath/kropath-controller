// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"github.com/kropath/kropath-controller/internal/cascade"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ProviderIdentity struct {
	AccountID string `json:"accountId,omitempty"`
	Region    string `json:"region,omitempty"`
}

type KropathConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KropathConfigSpec `json:"spec,omitempty"`
}

type KropathConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KropathConfig `json:"items"`
}

type KropathConfigSpec struct {
	Mandatory KropathConfigTier `json:"mandatory,omitempty"`
	Defaults  KropathConfigTier `json:"defaults,omitempty"`
	AWS       ProviderIdentity  `json:"aws,omitempty"`
}

type KropathConfigTier struct {
	IAM            cascade.IAMSection                   `json:"iam,omitempty"`
	S3             cascade.S3Section                    `json:"s3,omitempty"`
	KMS            cascade.KMSKropathSection            `json:"kms,omitempty"`
	SQS            cascade.SQSKropathSection            `json:"sqs,omitempty"`
	SecretsManager cascade.SMKropathSection             `json:"secretsManager,omitempty"`
	SNS            cascade.SNSKropathSection            `json:"sns,omitempty"`
	DynamoDB       cascade.DynamoDBKropathSection       `json:"dynamodb,omitempty"`
	EventBridge    cascade.EventBridgeKropathSection    `json:"eventbridge,omitempty"`
	CloudWatchLogs cascade.CloudWatchLogsKropathSection `json:"cloudwatchlogs,omitempty"`
	CloudWatch     cascade.CloudWatchKropathSection     `json:"cloudwatch,omitempty"`
	ELB            cascade.ELBKropathSection            `json:"elb,omitempty"`
	RDS            cascade.RDSKropathSection            `json:"rds,omitempty"`
	AutoScaling       cascade.AutoScalingKropathSection    `json:"autoscaling,omitempty"`
	ECS               cascade.ECSKropathSection            `json:"ecs,omitempty"`
	EKS               cascade.EKSKropathSection            `json:"eks,omitempty"`
	EC2               cascade.EC2KropathSection            `json:"ec2,omitempty"`
	ApiGatewayV2      cascade.ApiGatewayV2KropathSection      `json:"apigatewayv2,omitempty"`
	ApiGateway        cascade.ApiGatewayKropathSection        `json:"apigateway,omitempty"`
	ElastiCache       cascade.ElastiCacheKropathSection       `json:"elasticache,omitempty"`
	ECR               cascade.ECRKropathSection               `json:"ecr,omitempty"`
	StepFunctions     cascade.StepFunctionsKropathSection     `json:"stepfunctions,omitempty"`
	MSK               cascade.MSKKropathSection               `json:"msk,omitempty"`
	MemoryDB          cascade.MemoryDBKropathSection          `json:"memorydb,omitempty"`
	CertificateManager cascade.ACMKropathSection             `json:"certificateManager,omitempty"`
	EMR               cascade.EMRKropathSection               `json:"emr,omitempty"`
	DocumentDB        cascade.DocumentDBKropathSection        `json:"documentdb,omitempty"`
	Glue              cascade.GlueKropathSection              `json:"glue,omitempty"`
	Athena            cascade.AthenaKropathSection            `json:"athena,omitempty"`
	DSQL              cascade.DSQLKropathSection              `json:"dsql,omitempty"`
	Route53           cascade.Route53KropathSection           `json:"route53,omitempty"`
	SSM               cascade.SSMKropathSection               `json:"ssm,omitempty"`
	Cognito           cascade.CognitoKropathSection           `json:"cognito,omitempty"`
	Kinesis           cascade.KinesisKropathSection           `json:"kinesis,omitempty"`
	CloudTrail        cascade.CloudTrailKropathSection        `json:"cloudtrail,omitempty"`
	AppScaling        cascade.AppScalingKropathSection        `json:"appScaling,omitempty"`
	Keyspaces         cascade.KeyspacesKropathSection         `json:"keyspaces,omitempty"`
	WAF               cascade.WAFKropathSection               `json:"waf,omitempty"`
	Bedrock           cascade.BedrockKropathSection           `json:"bedrock,omitempty"`
	SageMaker         cascade.SageMakerKropathSection         `json:"sagemaker,omitempty"`
	OpenSearch        cascade.OpenSearchKropathSection        `json:"opensearch,omitempty"`
	Pipes             cascade.PipesKropathSection             `json:"pipes,omitempty"`
	CodeArtifact      cascade.CodeArtifactKropathSection      `json:"codeartifact,omitempty"`
	MWAA              cascade.MWAAKropathSection              `json:"mwaa,omitempty"`
	Tags              map[string]string                       `json:"tags,omitempty"`
	SyncedLabels      map[string]string                    `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string                    `json:"syncedAnnotations,omitempty"`
}

type IAMConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMConfigSpec   `json:"spec,omitempty"`
	Status IAMConfigStatus `json:"status,omitempty"`
}

type IAMConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []IAMConfig `json:"items"`
}

type IAMConfigSpec struct {
	Mandatory cascade.IAMSection `json:"mandatory,omitempty"`
	Defaults  cascade.IAMSection `json:"defaults,omitempty"`
}

type IAMConfigStatus struct {
	EffectiveConfig    EffectiveIAMConfig `json:"effectiveConfig,omitempty"`
	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                `json:"syncedTimestamp,omitempty"`
}

type EffectiveIAMConfig struct {
	AWS       ProviderIdentity `json:"aws,omitempty"`
	Mandatory cascade.IAMSection  `json:"mandatory,omitempty"`
	Defaults  cascade.IAMSection  `json:"defaults,omitempty"`
}

func (in *KropathConfig) DeepCopyInto(out *KropathConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	if in.Spec.Mandatory.KMS.AllowedKeySpecs != nil {
		out.Spec.Mandatory.KMS.AllowedKeySpecs = make([]string, len(in.Spec.Mandatory.KMS.AllowedKeySpecs))
		copy(out.Spec.Mandatory.KMS.AllowedKeySpecs, in.Spec.Mandatory.KMS.AllowedKeySpecs)
	}
	if in.Spec.Mandatory.KMS.Tags != nil {
		out.Spec.Mandatory.KMS.Tags = make(map[string]string, len(in.Spec.Mandatory.KMS.Tags))
		for k, v := range in.Spec.Mandatory.KMS.Tags {
			out.Spec.Mandatory.KMS.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.Tags != nil {
		out.Spec.Mandatory.Tags = make(map[string]string, len(in.Spec.Mandatory.Tags))
		for k, v := range in.Spec.Mandatory.Tags {
			out.Spec.Mandatory.Tags[k] = v
		}
	}
	if in.Spec.Defaults.KMS.AllowedKeySpecs != nil {
		out.Spec.Defaults.KMS.AllowedKeySpecs = make([]string, len(in.Spec.Defaults.KMS.AllowedKeySpecs))
		copy(out.Spec.Defaults.KMS.AllowedKeySpecs, in.Spec.Defaults.KMS.AllowedKeySpecs)
	}
	if in.Spec.Defaults.KMS.Tags != nil {
		out.Spec.Defaults.KMS.Tags = make(map[string]string, len(in.Spec.Defaults.KMS.Tags))
		for k, v := range in.Spec.Defaults.KMS.Tags {
			out.Spec.Defaults.KMS.Tags[k] = v
		}
	}
	if in.Spec.Defaults.Tags != nil {
		out.Spec.Defaults.Tags = make(map[string]string, len(in.Spec.Defaults.Tags))
		for k, v := range in.Spec.Defaults.Tags {
			out.Spec.Defaults.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.SQS.Tags != nil {
		out.Spec.Mandatory.SQS.Tags = make(map[string]string, len(in.Spec.Mandatory.SQS.Tags))
		for k, v := range in.Spec.Mandatory.SQS.Tags {
			out.Spec.Mandatory.SQS.Tags[k] = v
		}
	}
	if in.Spec.Defaults.SQS.Tags != nil {
		out.Spec.Defaults.SQS.Tags = make(map[string]string, len(in.Spec.Defaults.SQS.Tags))
		for k, v := range in.Spec.Defaults.SQS.Tags {
			out.Spec.Defaults.SQS.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.SecretsManager.Tags != nil {
		out.Spec.Mandatory.SecretsManager.Tags = make(map[string]string, len(in.Spec.Mandatory.SecretsManager.Tags))
		for k, v := range in.Spec.Mandatory.SecretsManager.Tags {
			out.Spec.Mandatory.SecretsManager.Tags[k] = v
		}
	}
	if in.Spec.Defaults.SecretsManager.Tags != nil {
		out.Spec.Defaults.SecretsManager.Tags = make(map[string]string, len(in.Spec.Defaults.SecretsManager.Tags))
		for k, v := range in.Spec.Defaults.SecretsManager.Tags {
			out.Spec.Defaults.SecretsManager.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.DynamoDB.Tags != nil {
		out.Spec.Mandatory.DynamoDB.Tags = make(map[string]string, len(in.Spec.Mandatory.DynamoDB.Tags))
		for k, v := range in.Spec.Mandatory.DynamoDB.Tags {
			out.Spec.Mandatory.DynamoDB.Tags[k] = v
		}
	}
	if in.Spec.Defaults.DynamoDB.Tags != nil {
		out.Spec.Defaults.DynamoDB.Tags = make(map[string]string, len(in.Spec.Defaults.DynamoDB.Tags))
		for k, v := range in.Spec.Defaults.DynamoDB.Tags {
			out.Spec.Defaults.DynamoDB.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.EventBridge.Tags != nil {
		out.Spec.Mandatory.EventBridge.Tags = make(map[string]string, len(in.Spec.Mandatory.EventBridge.Tags))
		for k, v := range in.Spec.Mandatory.EventBridge.Tags {
			out.Spec.Mandatory.EventBridge.Tags[k] = v
		}
	}
	if in.Spec.Defaults.EventBridge.Tags != nil {
		out.Spec.Defaults.EventBridge.Tags = make(map[string]string, len(in.Spec.Defaults.EventBridge.Tags))
		for k, v := range in.Spec.Defaults.EventBridge.Tags {
			out.Spec.Defaults.EventBridge.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.CloudWatchLogs.Tags != nil {
		out.Spec.Mandatory.CloudWatchLogs.Tags = make(map[string]string, len(in.Spec.Mandatory.CloudWatchLogs.Tags))
		for k, v := range in.Spec.Mandatory.CloudWatchLogs.Tags {
			out.Spec.Mandatory.CloudWatchLogs.Tags[k] = v
		}
	}
	if in.Spec.Defaults.CloudWatchLogs.Tags != nil {
		out.Spec.Defaults.CloudWatchLogs.Tags = make(map[string]string, len(in.Spec.Defaults.CloudWatchLogs.Tags))
		for k, v := range in.Spec.Defaults.CloudWatchLogs.Tags {
			out.Spec.Defaults.CloudWatchLogs.Tags[k] = v
		}
	}
	// CloudWatch.ActionsEnabled is *bool — deep-copy the pointer.
	if in.Spec.Mandatory.CloudWatch.ActionsEnabled != nil {
		v := *in.Spec.Mandatory.CloudWatch.ActionsEnabled
		out.Spec.Mandatory.CloudWatch.ActionsEnabled = &v
	}
	if in.Spec.Mandatory.CloudWatch.Tags != nil {
		out.Spec.Mandatory.CloudWatch.Tags = make(map[string]string, len(in.Spec.Mandatory.CloudWatch.Tags))
		for k, v := range in.Spec.Mandatory.CloudWatch.Tags {
			out.Spec.Mandatory.CloudWatch.Tags[k] = v
		}
	}
	if in.Spec.Defaults.CloudWatch.ActionsEnabled != nil {
		v := *in.Spec.Defaults.CloudWatch.ActionsEnabled
		out.Spec.Defaults.CloudWatch.ActionsEnabled = &v
	}
	if in.Spec.Defaults.CloudWatch.Tags != nil {
		out.Spec.Defaults.CloudWatch.Tags = make(map[string]string, len(in.Spec.Defaults.CloudWatch.Tags))
		for k, v := range in.Spec.Defaults.CloudWatch.Tags {
			out.Spec.Defaults.CloudWatch.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.ELB.Tags != nil {
		out.Spec.Mandatory.ELB.Tags = make(map[string]string, len(in.Spec.Mandatory.ELB.Tags))
		for k, v := range in.Spec.Mandatory.ELB.Tags {
			out.Spec.Mandatory.ELB.Tags[k] = v
		}
	}
	if in.Spec.Defaults.ELB.Tags != nil {
		out.Spec.Defaults.ELB.Tags = make(map[string]string, len(in.Spec.Defaults.ELB.Tags))
		for k, v := range in.Spec.Defaults.ELB.Tags {
			out.Spec.Defaults.ELB.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.RDS.Tags != nil {
		out.Spec.Mandatory.RDS.Tags = make(map[string]string, len(in.Spec.Mandatory.RDS.Tags))
		for k, v := range in.Spec.Mandatory.RDS.Tags {
			out.Spec.Mandatory.RDS.Tags[k] = v
		}
	}
	if in.Spec.Defaults.RDS.Tags != nil {
		out.Spec.Defaults.RDS.Tags = make(map[string]string, len(in.Spec.Defaults.RDS.Tags))
		for k, v := range in.Spec.Defaults.RDS.Tags {
			out.Spec.Defaults.RDS.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.AutoScaling.Tags != nil {
		out.Spec.Mandatory.AutoScaling.Tags = make(map[string]string, len(in.Spec.Mandatory.AutoScaling.Tags))
		for k, v := range in.Spec.Mandatory.AutoScaling.Tags {
			out.Spec.Mandatory.AutoScaling.Tags[k] = v
		}
	}
	if in.Spec.Defaults.AutoScaling.Tags != nil {
		out.Spec.Defaults.AutoScaling.Tags = make(map[string]string, len(in.Spec.Defaults.AutoScaling.Tags))
		for k, v := range in.Spec.Defaults.AutoScaling.Tags {
			out.Spec.Defaults.AutoScaling.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.ECS.Tags != nil {
		out.Spec.Mandatory.ECS.Tags = make(map[string]string, len(in.Spec.Mandatory.ECS.Tags))
		for k, v := range in.Spec.Mandatory.ECS.Tags {
			out.Spec.Mandatory.ECS.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.ECS.SyncedLabels != nil {
		out.Spec.Mandatory.ECS.SyncedLabels = make(map[string]string, len(in.Spec.Mandatory.ECS.SyncedLabels))
		for k, v := range in.Spec.Mandatory.ECS.SyncedLabels {
			out.Spec.Mandatory.ECS.SyncedLabels[k] = v
		}
	}
	if in.Spec.Mandatory.ECS.SyncedAnnotations != nil {
		out.Spec.Mandatory.ECS.SyncedAnnotations = make(map[string]string, len(in.Spec.Mandatory.ECS.SyncedAnnotations))
		for k, v := range in.Spec.Mandatory.ECS.SyncedAnnotations {
			out.Spec.Mandatory.ECS.SyncedAnnotations[k] = v
		}
	}
	if in.Spec.Defaults.ECS.Tags != nil {
		out.Spec.Defaults.ECS.Tags = make(map[string]string, len(in.Spec.Defaults.ECS.Tags))
		for k, v := range in.Spec.Defaults.ECS.Tags {
			out.Spec.Defaults.ECS.Tags[k] = v
		}
	}
	if in.Spec.Defaults.ECS.SyncedLabels != nil {
		out.Spec.Defaults.ECS.SyncedLabels = make(map[string]string, len(in.Spec.Defaults.ECS.SyncedLabels))
		for k, v := range in.Spec.Defaults.ECS.SyncedLabels {
			out.Spec.Defaults.ECS.SyncedLabels[k] = v
		}
	}
	if in.Spec.Defaults.ECS.SyncedAnnotations != nil {
		out.Spec.Defaults.ECS.SyncedAnnotations = make(map[string]string, len(in.Spec.Defaults.ECS.SyncedAnnotations))
		for k, v := range in.Spec.Defaults.ECS.SyncedAnnotations {
			out.Spec.Defaults.ECS.SyncedAnnotations[k] = v
		}
	}
	if in.Spec.Mandatory.EKS.LoggingTypes != nil {
		out.Spec.Mandatory.EKS.LoggingTypes = make([]string, len(in.Spec.Mandatory.EKS.LoggingTypes))
		copy(out.Spec.Mandatory.EKS.LoggingTypes, in.Spec.Mandatory.EKS.LoggingTypes)
	}
	if in.Spec.Mandatory.EKS.Tags != nil {
		out.Spec.Mandatory.EKS.Tags = make(map[string]string, len(in.Spec.Mandatory.EKS.Tags))
		for k, v := range in.Spec.Mandatory.EKS.Tags {
			out.Spec.Mandatory.EKS.Tags[k] = v
		}
	}
	if in.Spec.Defaults.EKS.LoggingTypes != nil {
		out.Spec.Defaults.EKS.LoggingTypes = make([]string, len(in.Spec.Defaults.EKS.LoggingTypes))
		copy(out.Spec.Defaults.EKS.LoggingTypes, in.Spec.Defaults.EKS.LoggingTypes)
	}
	if in.Spec.Defaults.EKS.Tags != nil {
		out.Spec.Defaults.EKS.Tags = make(map[string]string, len(in.Spec.Defaults.EKS.Tags))
		for k, v := range in.Spec.Defaults.EKS.Tags {
			out.Spec.Defaults.EKS.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.EC2.Tags != nil {
		out.Spec.Mandatory.EC2.Tags = make(map[string]string, len(in.Spec.Mandatory.EC2.Tags))
		for k, v := range in.Spec.Mandatory.EC2.Tags {
			out.Spec.Mandatory.EC2.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.EC2.SyncedLabels != nil {
		out.Spec.Mandatory.EC2.SyncedLabels = make(map[string]string, len(in.Spec.Mandatory.EC2.SyncedLabels))
		for k, v := range in.Spec.Mandatory.EC2.SyncedLabels {
			out.Spec.Mandatory.EC2.SyncedLabels[k] = v
		}
	}
	if in.Spec.Mandatory.EC2.SyncedAnnotations != nil {
		out.Spec.Mandatory.EC2.SyncedAnnotations = make(map[string]string, len(in.Spec.Mandatory.EC2.SyncedAnnotations))
		for k, v := range in.Spec.Mandatory.EC2.SyncedAnnotations {
			out.Spec.Mandatory.EC2.SyncedAnnotations[k] = v
		}
	}
	if in.Spec.Defaults.EC2.Tags != nil {
		out.Spec.Defaults.EC2.Tags = make(map[string]string, len(in.Spec.Defaults.EC2.Tags))
		for k, v := range in.Spec.Defaults.EC2.Tags {
			out.Spec.Defaults.EC2.Tags[k] = v
		}
	}
	if in.Spec.Defaults.EC2.SyncedLabels != nil {
		out.Spec.Defaults.EC2.SyncedLabels = make(map[string]string, len(in.Spec.Defaults.EC2.SyncedLabels))
		for k, v := range in.Spec.Defaults.EC2.SyncedLabels {
			out.Spec.Defaults.EC2.SyncedLabels[k] = v
		}
	}
	if in.Spec.Defaults.EC2.SyncedAnnotations != nil {
		out.Spec.Defaults.EC2.SyncedAnnotations = make(map[string]string, len(in.Spec.Defaults.EC2.SyncedAnnotations))
		for k, v := range in.Spec.Defaults.EC2.SyncedAnnotations {
			out.Spec.Defaults.EC2.SyncedAnnotations[k] = v
		}
	}
	if in.Spec.Mandatory.ApiGatewayV2.Tags != nil {
		out.Spec.Mandatory.ApiGatewayV2.Tags = make(map[string]string, len(in.Spec.Mandatory.ApiGatewayV2.Tags))
		for k, v := range in.Spec.Mandatory.ApiGatewayV2.Tags {
			out.Spec.Mandatory.ApiGatewayV2.Tags[k] = v
		}
	}
	if in.Spec.Defaults.ApiGatewayV2.Tags != nil {
		out.Spec.Defaults.ApiGatewayV2.Tags = make(map[string]string, len(in.Spec.Defaults.ApiGatewayV2.Tags))
		for k, v := range in.Spec.Defaults.ApiGatewayV2.Tags {
			out.Spec.Defaults.ApiGatewayV2.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.ApiGateway.Tags != nil {
		out.Spec.Mandatory.ApiGateway.Tags = make(map[string]string, len(in.Spec.Mandatory.ApiGateway.Tags))
		for k, v := range in.Spec.Mandatory.ApiGateway.Tags {
			out.Spec.Mandatory.ApiGateway.Tags[k] = v
		}
	}
	if in.Spec.Defaults.ApiGateway.Tags != nil {
		out.Spec.Defaults.ApiGateway.Tags = make(map[string]string, len(in.Spec.Defaults.ApiGateway.Tags))
		for k, v := range in.Spec.Defaults.ApiGateway.Tags {
			out.Spec.Defaults.ApiGateway.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.ElastiCache.Tags != nil {
		out.Spec.Mandatory.ElastiCache.Tags = make(map[string]string, len(in.Spec.Mandatory.ElastiCache.Tags))
		for k, v := range in.Spec.Mandatory.ElastiCache.Tags {
			out.Spec.Mandatory.ElastiCache.Tags[k] = v
		}
	}
	if in.Spec.Defaults.ElastiCache.Tags != nil {
		out.Spec.Defaults.ElastiCache.Tags = make(map[string]string, len(in.Spec.Defaults.ElastiCache.Tags))
		for k, v := range in.Spec.Defaults.ElastiCache.Tags {
			out.Spec.Defaults.ElastiCache.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.ECR.Tags != nil {
		out.Spec.Mandatory.ECR.Tags = make(map[string]string, len(in.Spec.Mandatory.ECR.Tags))
		for k, v := range in.Spec.Mandatory.ECR.Tags {
			out.Spec.Mandatory.ECR.Tags[k] = v
		}
	}
	if in.Spec.Defaults.ECR.Tags != nil {
		out.Spec.Defaults.ECR.Tags = make(map[string]string, len(in.Spec.Defaults.ECR.Tags))
		for k, v := range in.Spec.Defaults.ECR.Tags {
			out.Spec.Defaults.ECR.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.StepFunctions.Tags != nil {
		out.Spec.Mandatory.StepFunctions.Tags = make(map[string]string, len(in.Spec.Mandatory.StepFunctions.Tags))
		for k, v := range in.Spec.Mandatory.StepFunctions.Tags {
			out.Spec.Mandatory.StepFunctions.Tags[k] = v
		}
	}
	if in.Spec.Defaults.StepFunctions.Tags != nil {
		out.Spec.Defaults.StepFunctions.Tags = make(map[string]string, len(in.Spec.Defaults.StepFunctions.Tags))
		for k, v := range in.Spec.Defaults.StepFunctions.Tags {
			out.Spec.Defaults.StepFunctions.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.MemoryDB.AllowedNodeTypes != nil {
		out.Spec.Mandatory.MemoryDB.AllowedNodeTypes = make([]string, len(in.Spec.Mandatory.MemoryDB.AllowedNodeTypes))
		copy(out.Spec.Mandatory.MemoryDB.AllowedNodeTypes, in.Spec.Mandatory.MemoryDB.AllowedNodeTypes)
	}
	if in.Spec.Mandatory.MemoryDB.Tags != nil {
		out.Spec.Mandatory.MemoryDB.Tags = make(map[string]string, len(in.Spec.Mandatory.MemoryDB.Tags))
		for k, v := range in.Spec.Mandatory.MemoryDB.Tags {
			out.Spec.Mandatory.MemoryDB.Tags[k] = v
		}
	}
	if in.Spec.Defaults.MemoryDB.AllowedNodeTypes != nil {
		out.Spec.Defaults.MemoryDB.AllowedNodeTypes = make([]string, len(in.Spec.Defaults.MemoryDB.AllowedNodeTypes))
		copy(out.Spec.Defaults.MemoryDB.AllowedNodeTypes, in.Spec.Defaults.MemoryDB.AllowedNodeTypes)
	}
	if in.Spec.Defaults.MemoryDB.Tags != nil {
		out.Spec.Defaults.MemoryDB.Tags = make(map[string]string, len(in.Spec.Defaults.MemoryDB.Tags))
		for k, v := range in.Spec.Defaults.MemoryDB.Tags {
			out.Spec.Defaults.MemoryDB.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.CertificateManager.Tags != nil {
		out.Spec.Mandatory.CertificateManager.Tags = make(map[string]string, len(in.Spec.Mandatory.CertificateManager.Tags))
		for k, v := range in.Spec.Mandatory.CertificateManager.Tags {
			out.Spec.Mandatory.CertificateManager.Tags[k] = v
		}
	}
	if in.Spec.Defaults.CertificateManager.Tags != nil {
		out.Spec.Defaults.CertificateManager.Tags = make(map[string]string, len(in.Spec.Defaults.CertificateManager.Tags))
		for k, v := range in.Spec.Defaults.CertificateManager.Tags {
			out.Spec.Defaults.CertificateManager.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.DocumentDB.AllowedInstanceClasses != nil {
		out.Spec.Mandatory.DocumentDB.AllowedInstanceClasses = make([]string, len(in.Spec.Mandatory.DocumentDB.AllowedInstanceClasses))
		copy(out.Spec.Mandatory.DocumentDB.AllowedInstanceClasses, in.Spec.Mandatory.DocumentDB.AllowedInstanceClasses)
	}
	if in.Spec.Mandatory.DocumentDB.Tags != nil {
		out.Spec.Mandatory.DocumentDB.Tags = make(map[string]string, len(in.Spec.Mandatory.DocumentDB.Tags))
		for k, v := range in.Spec.Mandatory.DocumentDB.Tags {
			out.Spec.Mandatory.DocumentDB.Tags[k] = v
		}
	}
	if in.Spec.Defaults.DocumentDB.AllowedInstanceClasses != nil {
		out.Spec.Defaults.DocumentDB.AllowedInstanceClasses = make([]string, len(in.Spec.Defaults.DocumentDB.AllowedInstanceClasses))
		copy(out.Spec.Defaults.DocumentDB.AllowedInstanceClasses, in.Spec.Defaults.DocumentDB.AllowedInstanceClasses)
	}
	if in.Spec.Defaults.DocumentDB.Tags != nil {
		out.Spec.Defaults.DocumentDB.Tags = make(map[string]string, len(in.Spec.Defaults.DocumentDB.Tags))
		for k, v := range in.Spec.Defaults.DocumentDB.Tags {
			out.Spec.Defaults.DocumentDB.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.Glue.Tags != nil {
		out.Spec.Mandatory.Glue.Tags = make(map[string]string, len(in.Spec.Mandatory.Glue.Tags))
		for k, v := range in.Spec.Mandatory.Glue.Tags {
			out.Spec.Mandatory.Glue.Tags[k] = v
		}
	}
	if in.Spec.Defaults.Glue.Tags != nil {
		out.Spec.Defaults.Glue.Tags = make(map[string]string, len(in.Spec.Defaults.Glue.Tags))
		for k, v := range in.Spec.Defaults.Glue.Tags {
			out.Spec.Defaults.Glue.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.Athena.Tags != nil {
		out.Spec.Mandatory.Athena.Tags = make(map[string]string, len(in.Spec.Mandatory.Athena.Tags))
		for k, v := range in.Spec.Mandatory.Athena.Tags {
			out.Spec.Mandatory.Athena.Tags[k] = v
		}
	}
	if in.Spec.Defaults.Athena.Tags != nil {
		out.Spec.Defaults.Athena.Tags = make(map[string]string, len(in.Spec.Defaults.Athena.Tags))
		for k, v := range in.Spec.Defaults.Athena.Tags {
			out.Spec.Defaults.Athena.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.DSQL.Tags != nil {
		out.Spec.Mandatory.DSQL.Tags = make(map[string]string, len(in.Spec.Mandatory.DSQL.Tags))
		for k, v := range in.Spec.Mandatory.DSQL.Tags {
			out.Spec.Mandatory.DSQL.Tags[k] = v
		}
	}
	if in.Spec.Defaults.DSQL.Tags != nil {
		out.Spec.Defaults.DSQL.Tags = make(map[string]string, len(in.Spec.Defaults.DSQL.Tags))
		for k, v := range in.Spec.Defaults.DSQL.Tags {
			out.Spec.Defaults.DSQL.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.Route53.Tags != nil {
		out.Spec.Mandatory.Route53.Tags = make(map[string]string, len(in.Spec.Mandatory.Route53.Tags))
		for k, v := range in.Spec.Mandatory.Route53.Tags {
			out.Spec.Mandatory.Route53.Tags[k] = v
		}
	}
	if in.Spec.Defaults.Route53.Tags != nil {
		out.Spec.Defaults.Route53.Tags = make(map[string]string, len(in.Spec.Defaults.Route53.Tags))
		for k, v := range in.Spec.Defaults.Route53.Tags {
			out.Spec.Defaults.Route53.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.SSM.AllowedDocumentTypes != nil {
		out.Spec.Mandatory.SSM.AllowedDocumentTypes = make([]string, len(in.Spec.Mandatory.SSM.AllowedDocumentTypes))
		copy(out.Spec.Mandatory.SSM.AllowedDocumentTypes, in.Spec.Mandatory.SSM.AllowedDocumentTypes)
	}
	if in.Spec.Defaults.SSM.AllowedDocumentTypes != nil {
		out.Spec.Defaults.SSM.AllowedDocumentTypes = make([]string, len(in.Spec.Defaults.SSM.AllowedDocumentTypes))
		copy(out.Spec.Defaults.SSM.AllowedDocumentTypes, in.Spec.Defaults.SSM.AllowedDocumentTypes)
	}
	if in.Spec.Mandatory.SSM.Tags != nil {
		out.Spec.Mandatory.SSM.Tags = make(map[string]string, len(in.Spec.Mandatory.SSM.Tags))
		for k, v := range in.Spec.Mandatory.SSM.Tags {
			out.Spec.Mandatory.SSM.Tags[k] = v
		}
	}
	if in.Spec.Defaults.SSM.Tags != nil {
		out.Spec.Defaults.SSM.Tags = make(map[string]string, len(in.Spec.Defaults.SSM.Tags))
		for k, v := range in.Spec.Defaults.SSM.Tags {
			out.Spec.Defaults.SSM.Tags[k] = v
		}
	}
	// Cognito.PasswordPolicy *bool fields — deep-copy each pointer.
	if in.Spec.Mandatory.Cognito.PasswordPolicy.RequireLowercase != nil {
		v := *in.Spec.Mandatory.Cognito.PasswordPolicy.RequireLowercase
		out.Spec.Mandatory.Cognito.PasswordPolicy.RequireLowercase = &v
	}
	if in.Spec.Mandatory.Cognito.PasswordPolicy.RequireNumbers != nil {
		v := *in.Spec.Mandatory.Cognito.PasswordPolicy.RequireNumbers
		out.Spec.Mandatory.Cognito.PasswordPolicy.RequireNumbers = &v
	}
	if in.Spec.Mandatory.Cognito.PasswordPolicy.RequireSymbols != nil {
		v := *in.Spec.Mandatory.Cognito.PasswordPolicy.RequireSymbols
		out.Spec.Mandatory.Cognito.PasswordPolicy.RequireSymbols = &v
	}
	if in.Spec.Mandatory.Cognito.PasswordPolicy.RequireUppercase != nil {
		v := *in.Spec.Mandatory.Cognito.PasswordPolicy.RequireUppercase
		out.Spec.Mandatory.Cognito.PasswordPolicy.RequireUppercase = &v
	}
	if in.Spec.Mandatory.Cognito.Tags != nil {
		out.Spec.Mandatory.Cognito.Tags = make(map[string]string, len(in.Spec.Mandatory.Cognito.Tags))
		for k, v := range in.Spec.Mandatory.Cognito.Tags {
			out.Spec.Mandatory.Cognito.Tags[k] = v
		}
	}
	if in.Spec.Defaults.Cognito.PasswordPolicy.RequireLowercase != nil {
		v := *in.Spec.Defaults.Cognito.PasswordPolicy.RequireLowercase
		out.Spec.Defaults.Cognito.PasswordPolicy.RequireLowercase = &v
	}
	if in.Spec.Defaults.Cognito.PasswordPolicy.RequireNumbers != nil {
		v := *in.Spec.Defaults.Cognito.PasswordPolicy.RequireNumbers
		out.Spec.Defaults.Cognito.PasswordPolicy.RequireNumbers = &v
	}
	if in.Spec.Defaults.Cognito.PasswordPolicy.RequireSymbols != nil {
		v := *in.Spec.Defaults.Cognito.PasswordPolicy.RequireSymbols
		out.Spec.Defaults.Cognito.PasswordPolicy.RequireSymbols = &v
	}
	if in.Spec.Defaults.Cognito.PasswordPolicy.RequireUppercase != nil {
		v := *in.Spec.Defaults.Cognito.PasswordPolicy.RequireUppercase
		out.Spec.Defaults.Cognito.PasswordPolicy.RequireUppercase = &v
	}
	if in.Spec.Defaults.Cognito.Tags != nil {
		out.Spec.Defaults.Cognito.Tags = make(map[string]string, len(in.Spec.Defaults.Cognito.Tags))
		for k, v := range in.Spec.Defaults.Cognito.Tags {
			out.Spec.Defaults.Cognito.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.CloudTrail.Tags != nil {
		out.Spec.Mandatory.CloudTrail.Tags = make(map[string]string, len(in.Spec.Mandatory.CloudTrail.Tags))
		for k, v := range in.Spec.Mandatory.CloudTrail.Tags {
			out.Spec.Mandatory.CloudTrail.Tags[k] = v
		}
	}
	if in.Spec.Defaults.CloudTrail.Tags != nil {
		out.Spec.Defaults.CloudTrail.Tags = make(map[string]string, len(in.Spec.Defaults.CloudTrail.Tags))
		for k, v := range in.Spec.Defaults.CloudTrail.Tags {
			out.Spec.Defaults.CloudTrail.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.AppScaling.Tags != nil {
		out.Spec.Mandatory.AppScaling.Tags = make(map[string]string, len(in.Spec.Mandatory.AppScaling.Tags))
		for k, v := range in.Spec.Mandatory.AppScaling.Tags {
			out.Spec.Mandatory.AppScaling.Tags[k] = v
		}
	}
	if in.Spec.Defaults.AppScaling.Tags != nil {
		out.Spec.Defaults.AppScaling.Tags = make(map[string]string, len(in.Spec.Defaults.AppScaling.Tags))
		for k, v := range in.Spec.Defaults.AppScaling.Tags {
			out.Spec.Defaults.AppScaling.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.Keyspaces.Tags != nil {
		out.Spec.Mandatory.Keyspaces.Tags = make(map[string]string, len(in.Spec.Mandatory.Keyspaces.Tags))
		for k, v := range in.Spec.Mandatory.Keyspaces.Tags {
			out.Spec.Mandatory.Keyspaces.Tags[k] = v
		}
	}
	if in.Spec.Defaults.Keyspaces.Tags != nil {
		out.Spec.Defaults.Keyspaces.Tags = make(map[string]string, len(in.Spec.Defaults.Keyspaces.Tags))
		for k, v := range in.Spec.Defaults.Keyspaces.Tags {
			out.Spec.Defaults.Keyspaces.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.SageMaker.Tags != nil {
		out.Spec.Mandatory.SageMaker.Tags = make(map[string]string, len(in.Spec.Mandatory.SageMaker.Tags))
		for k, v := range in.Spec.Mandatory.SageMaker.Tags {
			out.Spec.Mandatory.SageMaker.Tags[k] = v
		}
	}
	if in.Spec.Defaults.SageMaker.Tags != nil {
		out.Spec.Defaults.SageMaker.Tags = make(map[string]string, len(in.Spec.Defaults.SageMaker.Tags))
		for k, v := range in.Spec.Defaults.SageMaker.Tags {
			out.Spec.Defaults.SageMaker.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.OpenSearch.Tags != nil {
		out.Spec.Mandatory.OpenSearch.Tags = make(map[string]string, len(in.Spec.Mandatory.OpenSearch.Tags))
		for k, v := range in.Spec.Mandatory.OpenSearch.Tags {
			out.Spec.Mandatory.OpenSearch.Tags[k] = v
		}
	}
	if in.Spec.Defaults.OpenSearch.Tags != nil {
		out.Spec.Defaults.OpenSearch.Tags = make(map[string]string, len(in.Spec.Defaults.OpenSearch.Tags))
		for k, v := range in.Spec.Defaults.OpenSearch.Tags {
			out.Spec.Defaults.OpenSearch.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.CodeArtifact.Tags != nil {
		out.Spec.Mandatory.CodeArtifact.Tags = make(map[string]string, len(in.Spec.Mandatory.CodeArtifact.Tags))
		for k, v := range in.Spec.Mandatory.CodeArtifact.Tags {
			out.Spec.Mandatory.CodeArtifact.Tags[k] = v
		}
	}
	if in.Spec.Defaults.CodeArtifact.Tags != nil {
		out.Spec.Defaults.CodeArtifact.Tags = make(map[string]string, len(in.Spec.Defaults.CodeArtifact.Tags))
		for k, v := range in.Spec.Defaults.CodeArtifact.Tags {
			out.Spec.Defaults.CodeArtifact.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.MWAA.AirflowConfigurationOptions != nil {
		out.Spec.Mandatory.MWAA.AirflowConfigurationOptions = make(map[string]string, len(in.Spec.Mandatory.MWAA.AirflowConfigurationOptions))
		for k, v := range in.Spec.Mandatory.MWAA.AirflowConfigurationOptions {
			out.Spec.Mandatory.MWAA.AirflowConfigurationOptions[k] = v
		}
	}
	if in.Spec.Defaults.MWAA.AirflowConfigurationOptions != nil {
		out.Spec.Defaults.MWAA.AirflowConfigurationOptions = make(map[string]string, len(in.Spec.Defaults.MWAA.AirflowConfigurationOptions))
		for k, v := range in.Spec.Defaults.MWAA.AirflowConfigurationOptions {
			out.Spec.Defaults.MWAA.AirflowConfigurationOptions[k] = v
		}
	}
	if in.Spec.Mandatory.SyncedLabels != nil {
		out.Spec.Mandatory.SyncedLabels = make(map[string]string, len(in.Spec.Mandatory.SyncedLabels))
		for k, v := range in.Spec.Mandatory.SyncedLabels {
			out.Spec.Mandatory.SyncedLabels[k] = v
		}
	}
	if in.Spec.Mandatory.SyncedAnnotations != nil {
		out.Spec.Mandatory.SyncedAnnotations = make(map[string]string, len(in.Spec.Mandatory.SyncedAnnotations))
		for k, v := range in.Spec.Mandatory.SyncedAnnotations {
			out.Spec.Mandatory.SyncedAnnotations[k] = v
		}
	}
	if in.Spec.Defaults.SyncedLabels != nil {
		out.Spec.Defaults.SyncedLabels = make(map[string]string, len(in.Spec.Defaults.SyncedLabels))
		for k, v := range in.Spec.Defaults.SyncedLabels {
			out.Spec.Defaults.SyncedLabels[k] = v
		}
	}
	if in.Spec.Defaults.SyncedAnnotations != nil {
		out.Spec.Defaults.SyncedAnnotations = make(map[string]string, len(in.Spec.Defaults.SyncedAnnotations))
		for k, v := range in.Spec.Defaults.SyncedAnnotations {
			out.Spec.Defaults.SyncedAnnotations[k] = v
		}
	}
}

func (in *KropathConfig) DeepCopy() *KropathConfig {
	if in == nil {
		return nil
	}
	out := new(KropathConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *KropathConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *KropathConfigList) DeepCopyInto(out *KropathConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]KropathConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *KropathConfigList) DeepCopy() *KropathConfigList {
	if in == nil {
		return nil
	}
	out := new(KropathConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *KropathConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *IAMConfig) DeepCopyInto(out *IAMConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *IAMConfig) DeepCopy() *IAMConfig {
	if in == nil {
		return nil
	}
	out := new(IAMConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *IAMConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *IAMConfigList) DeepCopyInto(out *IAMConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]IAMConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *IAMConfigList) DeepCopy() *IAMConfigList {
	if in == nil {
		return nil
	}
	out := new(IAMConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *IAMConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
