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
IMAGE_TAG        ?= $(shell git rev-parse --short HEAD)
REPORT_DIR       := test-results
CONTROLLER_LOG   := /tmp/kropath-controller/controller.log
CONTROLLER_PID   := /tmp/kropath-controller/pid

# ─── Test config ───────────────────────────────────────────────────────────────
KIND_CLUSTER     := kropath-controller-test
HEALTH_PORT      := 18081
TEST_NAMESPACES  := kro-system payments-prod
CHAINSAW         ?= chainsaw
GOLANGCI         ?= golangci-lint

CHAINSAW_FLAGS   := --parallel 1 --report-format JUNIT-TEST --report-path $(REPORT_DIR)/

.PHONY: all build test test-cover vet fmt lint \
        docker-build docker-push \
        kind-up kind-down \
        chainsaw-setup chainsaw-start chainsaw-wait chainsaw-stop \
        test-iam test-s3 test-kms test-policy test-chainsaw \
        install-tools gosec vulncheck security \
        help default

default: help

all: build ## Build the operator binary (default target).

# ─── Build ─────────────────────────────────────────────────────────────────────

build: ## Compile the operator binary → bin/kropath-operator.
	@mkdir -p bin
	go build -o $(BINARY) $(MAIN_PKG)

# ─── Container image ────────────────────────────────────────────────────────────

docker-build: ## Build the container image, tagged with the short git SHA and 'latest'.
	docker build -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
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

lint: vet ## Run golangci-lint (required gate before every commit).
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

chainsaw-setup: build kind-up ## Create kind cluster, apply CRDs, create test namespaces.
	kubectl apply -f tests/fixtures/crds/
	kubectl wait --for=condition=Established --timeout=60s \
		crd/awskropathconfigs.kropath.run \
		crd/awsiamconfigs.kropath.run \
		crd/awss3configs.kropath.run \
		crd/awskmsconfigs.kropath.run \
		crd/awspolicydocuments.kropath.run \
		crd/awsiamroles.kropath.run \
		crd/awss3buckets.kropath.run \
		crd/awslambdafunctions.kropath.run \
		crd/awssqsqueues.kropath.run \
		crd/awskmskeys.kropath.run \
		crd/awssecretsmanagersecrets.kropath.run
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
			--enable-kms-cascade \
			--enable-poldoc \
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
