# Copyright 2026 kropath Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

CHAINSAW ?= chainsaw

.PHONY: setup test-iam test-s3 test-kms test-sqs test-policy test-label-operator test-secretsmanager test-sns test-dynamodb test-eventbridge test-cloudwatchlogs test-rds test-autoscaling test-ecs test-eks test-ec2 test-apigateway test-apigatewayv2 test-efs test-elasticache test-ecr test-stepfunctions test-msk test-memorydb test-acm test-emr test-athena test-dsql test-version test-features test-dyn test-dyn-01 test-dyn-02 test-dyn-03 test-dyn-04 test-dyn-05 test-all

## setup: Apply all CRDs to the cluster. Run before any test target.
setup:
	kubectl apply -f fixtures/crds/
	@echo "CRDs applied. Verify with: kubectl get crd | grep kropath"

## test-iam: Run IAM cascade integration tests (ctrl-iam-01).
test-iam:
	$(CHAINSAW) test iam/ctrl-iam-01/

## test-s3: Run S3 cascade integration tests (ctrl-s3-01).
test-s3:
	$(CHAINSAW) test s3/ctrl-s3-01/

## test-kms: Run KMS cascade integration tests (ctrl-kms-01).
test-kms:
	$(CHAINSAW) test kms/ctrl-kms-01/

## test-sqs: Run SQS cascade integration tests (ctrl-sqs-01).
test-sqs:
	$(CHAINSAW) test sqs/ctrl-sqs-01/

## test-policy: Run policy document integration tests.
test-policy:
	$(CHAINSAW) test policy/phase2-refs/ policy/phase3-merge/

## test-label-operator: Run label-operator integration tests (ctrl-label-op-01).
test-label-operator:
	$(CHAINSAW) test label-operator/ctrl-label-op-01/

## test-secretsmanager: Run Secrets Manager cascade integration tests (ctrl-secretsmanager-01).
test-secretsmanager:
	$(CHAINSAW) test secretsmanager/ctrl-secretsmanager-01/

## test-sns: Run SNS cascade integration tests (ctrl-sns-01).
test-sns:
	$(CHAINSAW) test sns/ctrl-sns-01/

## test-dynamodb: Run DynamoDB cascade integration tests (ctrl-dynamodb-01).
test-dynamodb:
	$(CHAINSAW) test dynamodb/ctrl-dynamodb-01/

## test-eventbridge: Run EventBridge cascade integration tests (ctrl-eventbridge-01).
test-eventbridge:
	$(CHAINSAW) test eventbridge/ctrl-eventbridge-01/

## test-cloudwatchlogs: Run CloudWatch Logs cascade integration tests (ctrl-cwl-01).
test-cloudwatchlogs:
	$(CHAINSAW) test cloudwatchlogs/ctrl-cwl-01/

## test-rds: Run RDS cascade integration tests (ctrl-rds-01).
test-rds:
	$(CHAINSAW) test rds/ctrl-rds-01/

## test-autoscaling: Run Auto Scaling cascade integration tests (ctrl-autoscaling-01).
test-autoscaling:
	$(CHAINSAW) test autoscaling/ctrl-autoscaling-01/

## test-ecs: Run ECS cascade integration tests (ctrl-ecs-01).
test-ecs:
	$(CHAINSAW) test ecs/ctrl-ecs-01/

## test-eks: Run EKS cascade integration tests (ctrl-eks-01).
test-eks:
	$(CHAINSAW) test eks/ctrl-eks-01/

## test-ec2: Run EC2 cascade integration tests (ctrl-ec2-01).
test-ec2:
	$(CHAINSAW) test ec2/ctrl-ec2-01/

## test-apigateway: Run API Gateway cascade integration tests (ctrl-apigw-01).
test-apigateway:
	$(CHAINSAW) test apigateway/ctrl-apigw-01/

## test-apigatewayv2: Run API Gateway V2 cascade integration tests (ctrl-apigwv2-01).
test-apigatewayv2:
	$(CHAINSAW) test apigatewayv2/ctrl-apigwv2-01/

## test-efs: Run EFS cascade integration tests (ctrl-efs-01).
test-efs:
	$(CHAINSAW) test efs/ctrl-efs-01/

## test-elasticache: Run ElastiCache cascade integration tests (ctrl-elasticache-01).
test-elasticache:
	$(CHAINSAW) test elasticache/ctrl-elasticache-01/

## test-ecr: Run ECR cascade integration tests (ctrl-ecr-01).
test-ecr:
	$(CHAINSAW) test ecr/ctrl-ecr-01/

## test-stepfunctions: Run Step Functions cascade integration tests (ctrl-sfn-01).
test-stepfunctions:
	$(CHAINSAW) test stepfunctions/ctrl-sfn-01/

## test-msk: Run MSK cascade integration tests (ctrl-msk-01).
test-msk:
	$(CHAINSAW) test msk/mskconfig/ctrl-msk-01/

## test-memorydb: Run MemoryDB cascade integration tests (ctrl-memorydb-01).
test-memorydb:
	$(CHAINSAW) test memorydb/ctrl-memorydb-01/

## test-acm: Run Certificate Manager cascade integration tests (ctrl-acm).
test-acm:
	$(CHAINSAW) test acm/ctrl-acm/

## test-emr: Run EMR cascade integration tests (ctrl-emr-01).
test-emr:
	$(CHAINSAW) test emr/ctrl-emr-01/

## test-athena: Run Athena cascade integration tests (ctrl-athena-01).
test-athena:
	$(CHAINSAW) test athena/ctrl-athena-01/

## test-dsql: Run Aurora DSQL cascade integration tests (ctrl-dsql-01).
test-dsql:
	$(CHAINSAW) test dsql/ctrl-dsql-01/

## test-version: Run build-info and feature-enabled metrics integration tests (ctrl-version-01).
test-version:
	$(CHAINSAW) test version/ctrl-version-01/

## test-features: Run /features endpoint integration tests (ctrl-features-01).
test-features:
	$(CHAINSAW) test features/ctrl-features-01/

## test-dyn-01: Run startup gate tests (ctrl-dyn-01). ELBConfig CRD must be ABSENT.
test-dyn-01:
	$(CHAINSAW) test ctrl-dyn-01/

## test-dyn-02: Run dynamic CRD activation tests (ctrl-dyn-02).
## Run after test-dyn-01. Installs the ELBConfig CRD mid-test and verifies activation.
test-dyn-02:
	$(CHAINSAW) test ctrl-dyn-02/

## test-dyn-03: Run CRD removal resilience tests (ctrl-dyn-03).
## Run after test-dyn-02. Deletes and reinstalls the ELBConfig CRD.
test-dyn-03:
	$(CHAINSAW) test ctrl-dyn-03/

## test-dyn-04: Run policydocument partial optional kinds tests (ctrl-dyn-04).
## AWSLambdaFunction CRD must be ABSENT before this target runs (lives in crds-optional/).
test-dyn-04:
	$(CHAINSAW) test ctrl-dyn-04/

## test-dyn-05: Run labeloperator late kind tests (ctrl-dyn-05).
## CloudSpannerInstanceConfig CRD must be ABSENT before this target runs (lives in crds-optional/).
test-dyn-05:
	$(CHAINSAW) test ctrl-dyn-05/

## test-dyn: Run all dynamic CRD detection integration tests in order (dyn-01 → dyn-02 → dyn-03 → dyn-04 → dyn-05).
## Requires ELBConfig, AWSLambdaFunction, and CloudSpannerInstanceConfig CRDs to be ABSENT before this target runs.
## The controller must already be running via chainsaw-start chainsaw-wait.
test-dyn: test-dyn-01 test-dyn-02 test-dyn-03 test-dyn-04 test-dyn-05

## test-all: Run all integration test suites.
test-all: test-iam test-s3 test-kms test-sqs test-policy test-label-operator test-secretsmanager test-sns test-dynamodb test-eventbridge test-cloudwatchlogs test-rds test-autoscaling test-ecs test-eks test-ec2 test-apigateway test-apigatewayv2 test-efs test-elasticache test-ecr test-stepfunctions test-msk test-memorydb test-acm test-emr test-athena test-dsql test-version test-features test-dyn
