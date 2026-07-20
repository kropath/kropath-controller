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

.PHONY: setup test-iam test-s3 test-kms test-sqs test-policy test-label-operator test-all

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

## test-all: Run all integration test suites.
test-all: test-iam test-s3 test-kms test-sqs test-policy test-label-operator
