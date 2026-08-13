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

# ─── Version pins — keep in sync with .github/workflows/ci.yaml ───────────────
#
# Go version is read from go.mod via setup-go; all other tool versions are
# pinned here and mirrored in ci.yaml so local and CI environments are identical.
KIND_VERSION     := v0.25.0
GOLANGCI_VERSION := v2.11.4
CHAINSAW_VERSION := v0.2.15

# ─── Paths ─────────────────────────────────────────────────────────────────────
BINARY           := bin/kropath-operator
MAIN_PKG         := ./cmd/manager
IMAGE_REGISTRY   := ghcr.io/kropath
IMAGE_NAME       := kropath-controller
IMAGE_TAG        ?= $(shell git rev-parse --short=7 HEAD)
REPORT_DIR       := test-results
CONTROLLER_LOG   := /tmp/kropath-controller/controller.log
CONTROLLER_PID   := /tmp/kropath-controller/pid

# ─── Version stamping ──────────────────────────────────────────────────────────
VERSION    ?= dev
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo none)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MODULE     := github.com/kropath/kropath-controller
LDFLAGS    := -s -w \
    -X $(MODULE)/internal/version.Version=$(VERSION) \
    -X $(MODULE)/internal/version.GitCommit=$(GIT_COMMIT) \
    -X "$(MODULE)/internal/version.BuildDate=$(BUILD_DATE)"

# ─── Test config ───────────────────────────────────────────────────────────────
KIND_CLUSTER     := kropath-controller-test
HEALTH_PORT      := 18081
TEST_NAMESPACES  := kro-system payments-prod events-prod network-prod registry-prod
CHAINSAW         ?= chainsaw
GOLANGCI         ?= golangci-lint

CHAINSAW_FLAGS   := --parallel 1 --report-format JUNIT-TEST --report-path $(REPORT_DIR)/

.PHONY: all build test test-cover vet fmt lint \
        features-gen features-verify \
        docker-build docker-push \
        kind-up kind-down \
        chainsaw-setup chainsaw-start chainsaw-wait chainsaw-stop \
        test-iam test-s3 test-kms test-policy test-label-operator \
        test-apigatewayv2 test-autoscaling test-cwl test-dynamodb test-ec2 \
        test-ecr test-ecs test-efs test-eks test-elasticache test-eventbridge \
        test-rds test-secretsmanager test-sns test-sqs test-stepfunctions \
        test-version test-features \
        test-chainsaw \
        install-tools gosec vulncheck security \
        help default

default: help

all: build ## Build the operator binary (default target).

# ─── Build ─────────────────────────────────────────────────────────────────────

build: ## Compile the operator binary → bin/kropath-operator (with version ldflags).
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(MAIN_PKG)

# ─── Feature registry ──────────────────────────────────────────────────────────

features-gen: ## Regenerate docs/features.yaml from internal/features.All.
	go run ./cmd/gen-features

features-verify: ## CI gate: fail if docs/features.yaml is stale (run make features-gen to fix).
	go run ./cmd/gen-features --check

# ─── Container image ────────────────────────────────────────────────────────────

# Single-architecture (the host's) on purpose: buildx cannot --load a multi-platform
# image into the local daemon. CI publishes linux/amd64 + linux/arm64 manifests.
docker-build: ## Build the container image for the host architecture (version-stamped), tagged with the short git SHA and 'latest'.
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		-t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_REGISTRY)/$(IMAGE_NAME):latest .

docker-push: ## Push the SHA-tagged and 'latest' images (CI use; requires prior registry login).
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):latest

# ─── Unit tests ────────────────────────────────────────────────────────────────

test: ## Run unit tests with race detection.
	go test -race ./...

test-cover: ## Run unit tests with race detection and produce an HTML coverage report.
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ─── Code quality ──────────────────────────────────────────────────────────────

vet: ## Run go vet over all packages.
	go vet ./...

fmt: ## Format Go source files with gofmt and goimports.
	gofmt -w .
	@goimports -w . 2>/dev/null || \
		echo "goimports not installed; run: go install golang.org/x/tools/cmd/goimports@latest"

lint: vet features-verify ## Run golangci-lint (required gate before every commit).
	$(GOLANGCI) run ./...

# ─── Kind cluster ──────────────────────────────────────────────────────────────

kind-up: ## Create the kind cluster (no-op if it already exists).
	@command -v kind >/dev/null 2>&1 || \
		{ echo "ERROR: kind not found. Run: make install-tools"; exit 1; }
	@if kind get clusters 2>/dev/null | grep -q "^$(KIND_CLUSTER)$$"; then \
		echo "Kind cluster '$(KIND_CLUSTER)' already exists."; \
	else \
		kind create cluster --name $(KIND_CLUSTER) --wait 120s; \
	fi

kind-down: ## Delete the kind cluster.
	kind delete cluster --name $(KIND_CLUSTER)

# ─── Chainsaw integration-test helpers ────────────────────────────────────────
#
# Typical workflow:
#   make test-chainsaw          # end-to-end: kind-up → build → CRDs → start → run all → stop
#
# To iterate on a single suite without restarting the controller:
#   make chainsaw-start chainsaw-wait
#   make test-kms
#   make chainsaw-stop
#
# On failure mid-run, clean up manually: make chainsaw-stop kind-down

# Every CRD shipped under tests/fixtures/crds/, as `crd/<metadata.name>` arguments.
# Derived from the fixture files rather than hand-listed: a kind that a reconciler
# watches but whose CRD is absent from the cluster makes controller-runtime's informer
# never sync, and the manager exits with a fatal error after the 2-minute cache-sync
# timeout. That surfaces as every Chainsaw suite after the ~2-minute mark timing out,
# which is confusing to diagnose and has nothing to do with those suites. Deriving the
# list keeps it from drifting out of sync with the fixtures again (KRO-635).
CRD_WAIT_TARGETS := $(shell awk '/^  name: [a-z0-9.]+$$/ {print "crd/" $$2}' tests/fixtures/crds/*.yaml | sort -u)

chainsaw-setup: build kind-up ## Create kind cluster, apply CRDs, create test namespaces.
	kubectl apply -f tests/fixtures/crds/
	kubectl wait --for=condition=Established --timeout=60s $(CRD_WAIT_TARGETS)
	@for ns in $(TEST_NAMESPACES); do \
		kubectl create namespace "$$ns" --dry-run=client -o yaml | kubectl apply -f -; \
	done

chainsaw-start: chainsaw-setup ## Build, set up CRDs, and start the operator in the background.
	@mkdir -p /tmp/kropath-controller
	@if [ -f $(CONTROLLER_PID) ] && kill -0 "$$(cat $(CONTROLLER_PID))" 2>/dev/null; then \
		echo "Controller already running (PID $$(cat $(CONTROLLER_PID)))."; \
	else \
		KUBECONFIG="$${HOME}/.kube/config" $(BINARY) \
			--health-probe-bind-address=:$(HEALTH_PORT) \
			> $(CONTROLLER_LOG) 2>&1 & echo $$! > $(CONTROLLER_PID); \
		echo "Controller started (PID $$(cat $(CONTROLLER_PID)))."; \
	fi

chainsaw-wait: ## Block until the controller /readyz endpoint responds (up to 60 s).
	@echo "Waiting for /readyz on :$(HEALTH_PORT)..."
	@for i in $$(seq 1 30); do \
		if curl -fsS http://127.0.0.1:$(HEALTH_PORT)/readyz >/dev/null 2>&1; then \
			echo "Controller is ready."; exit 0; \
		fi; \
		sleep 2; \
	done; \
	echo "ERROR: controller did not become ready. Log:"; \
	cat $(CONTROLLER_LOG); \
	exit 1

chainsaw-stop: ## Stop the background controller process.
	@if [ -f $(CONTROLLER_PID) ]; then \
		kill "$$(cat $(CONTROLLER_PID))" 2>/dev/null && echo "Controller stopped." || true; \
		rm -f $(CONTROLLER_PID); \
	else \
		echo "No controller PID file found (already stopped?)."; \
	fi

# ─── Chainsaw individual suites ────────────────────────────────────────────────
# These targets assume the controller is already running (chainsaw-start + chainsaw-wait).

test-iam: ## Run IAM cascade Chainsaw suite (ctrl-iam-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/iam/ctrl-iam-01/ $(CHAINSAW_FLAGS)

test-s3: ## Run S3 cascade Chainsaw suite (ctrl-s3-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/s3/ctrl-s3-01/ $(CHAINSAW_FLAGS)

test-kms: ## Run KMS cascade Chainsaw suite (ctrl-kms-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/kms/ctrl-kms-01/ $(CHAINSAW_FLAGS)

test-policy: ## Run policy document Chainsaw suites (phase2-refs + phase3-merge).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/policy/phase2-refs/ tests/policy/phase3-merge/ $(CHAINSAW_FLAGS)

test-label-operator: ## Run label-operator Chainsaw suite (ctrl-label-op-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/label-operator/ctrl-label-op-01/ $(CHAINSAW_FLAGS)

test-apigatewayv2: ## Run API Gateway v2 cascade Chainsaw suite (ctrl-apigwv2-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/apigatewayv2/ctrl-apigwv2-01/ $(CHAINSAW_FLAGS)

test-autoscaling: ## Run Auto Scaling cascade Chainsaw suite (ctrl-autoscaling-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/autoscaling/ctrl-autoscaling-01/ $(CHAINSAW_FLAGS)

test-cwl: ## Run CloudWatch Logs cascade Chainsaw suite (ctrl-cwl-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/cloudwatchlogs/ctrl-cwl-01/ $(CHAINSAW_FLAGS)

test-dynamodb: ## Run DynamoDB cascade Chainsaw suite (ctrl-dynamodb-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/dynamodb/ctrl-dynamodb-01/ $(CHAINSAW_FLAGS)

test-ec2: ## Run EC2 cascade Chainsaw suite (ctrl-ec2-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ec2/ctrl-ec2-01/ $(CHAINSAW_FLAGS)

test-ecr: ## Run ECR cascade Chainsaw suite (ctrl-ecr-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ecr/ctrl-ecr-01/ $(CHAINSAW_FLAGS)

test-ecs: ## Run ECS cascade Chainsaw suite (ctrl-ecs-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ecs/ctrl-ecs-01/ $(CHAINSAW_FLAGS)

test-efs: ## Run EFS cascade Chainsaw suite (ctrl-efs-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/efs/ctrl-efs-01/ $(CHAINSAW_FLAGS)

test-eks: ## Run EKS cascade Chainsaw suite (ctrl-eks-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/eks/ctrl-eks-01/ $(CHAINSAW_FLAGS)

test-elasticache: ## Run ElastiCache cascade Chainsaw suite (ctrl-elasticache-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/elasticache/ctrl-elasticache-01/ $(CHAINSAW_FLAGS)

test-eventbridge: ## Run EventBridge cascade Chainsaw suite (ctrl-eventbridge-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/eventbridge/ctrl-eventbridge-01/ $(CHAINSAW_FLAGS)

test-rds: ## Run RDS cascade Chainsaw suite (ctrl-rds-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/rds/ctrl-rds-01/ $(CHAINSAW_FLAGS)

test-secretsmanager: ## Run Secrets Manager cascade Chainsaw suite (ctrl-secretsmanager-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/secretsmanager/ctrl-secretsmanager-01/ $(CHAINSAW_FLAGS)

test-sns: ## Run SNS cascade Chainsaw suite (ctrl-sns-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/sns/ctrl-sns-01/ $(CHAINSAW_FLAGS)

test-sqs: ## Run SQS cascade Chainsaw suite (ctrl-sqs-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/sqs/ctrl-sqs-01/ $(CHAINSAW_FLAGS)

test-stepfunctions: ## Run Step Functions cascade Chainsaw suite (ctrl-sfn-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/stepfunctions/ctrl-sfn-01/ $(CHAINSAW_FLAGS)

test-version: ## Run build-info and feature-enabled metrics Chainsaw suite (ctrl-version-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/version/ctrl-version-01/ $(CHAINSAW_FLAGS)

test-features: ## Run /features endpoint Chainsaw suite (ctrl-features-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/features/ctrl-features-01/ $(CHAINSAW_FLAGS)

test-chainsaw: chainsaw-stop chainsaw-start chainsaw-wait ## Stop any stale controller, start fresh, run ALL Chainsaw suites, then stop it.
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/     $(CHAINSAW_FLAGS)
# 	$(CHAINSAW) test tests/s3/ctrl-s3-01/       $(CHAINSAW_FLAGS)
# 	$(CHAINSAW) test tests/kms/ctrl-kms-01/     $(CHAINSAW_FLAGS)
# 	$(CHAINSAW) test tests/policy/phase2-refs/ tests/policy/phase3-merge/ $(CHAINSAW_FLAGS)
	$(MAKE) chainsaw-stop

# ─── Tool installation ─────────────────────────────────────────────────────────

install-tools: ## Install kind, chainsaw, golangci-lint, and goimports locally.
	go install sigs.k8s.io/kind@$(KIND_VERSION)
	go install github.com/kyverno/chainsaw@$(CHAINSAW_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install golang.org/x/tools/cmd/goimports@latest

# ─── Security scans ────────────────────────────────────────────────────────────
# Run only when implementation is complete. Do not run during active development.

gosec: ## Run gosec SAST scanner.
	gosec ./...

vulncheck: ## Run govulncheck dependency CVE scanner.
	govulncheck ./...

security: gosec vulncheck ## Run all security scans (gosec + govulncheck).

# ─── Help ──────────────────────────────────────────────────────────────────────

help: ## Display this help message.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'
