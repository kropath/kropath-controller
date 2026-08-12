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

.PHONY: setup test-iam test-s3 test-kms test-sqs test-policy test-label-operator test-secretsmanager test-sns test-dynamodb test-eventbridge test-cloudwatchlogs test-rds test-autoscaling test-ecs test-eks test-ec2 test-apigatewayv2 test-all

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

## test-apigatewayv2: Run API Gateway V2 cascade integration tests (ctrl-apigwv2-01).
test-apigatewayv2:
	$(CHAINSAW) test apigatewayv2/ctrl-apigwv2-01/

## test-all: Run all integration test suites.
test-all: test-iam test-s3 test-kms test-sqs test-policy test-label-operator test-secretsmanager test-sns test-dynamodb test-eventbridge test-cloudwatchlogs test-rds test-autoscaling test-ecs test-eks test-ec2 test-apigatewayv2
