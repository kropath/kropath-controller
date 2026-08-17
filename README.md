# kropath-controller

[![CI](https://github.com/kropath/kropath-controller/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/kropath/kropath-controller/actions/workflows/ci.yaml)
[![Release](https://github.com/kropath/kropath-controller/actions/workflows/release.yaml/badge.svg)](https://github.com/kropath/kropath-controller/actions/workflows/release.yaml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

Go controller for [kropath](https://github.com/kropath) (kro + golden path), a multi-cloud golden
path platform. `kropath-controller` is a multi-reconciler `kropath-operator` binary built on a
single `controller-runtime` Manager — one Reconciler struct per feature.

It is the **config-resolution half** of kropath: it watches the governance CRs a platform team
writes (`KropathConfig` and per-service `<Service>Config`), merges the layers under a strict
`mandatory > spec > defaults` precedence, and writes a single `status.effectiveConfig` object.
The kro ResourceGraphDefinitions in [kropath-aws](https://github.com/kropath/kropath-aws) read
that object through one `externalRef` lookup per RGD and project it onto ACK resources.

> ### ⚠️ EXPERIMENTAL
>
> **This project is under active development and is not production-ready.** CRD schemas,
> `status.effectiveConfig` shapes, and reconciler behaviour may change without notice, and there
> are no compatibility guarantees between releases.
>
> CI runs unit tests and Chainsaw integration tests against a local
> [kind](https://kind.sigs.k8s.io/) cluster with the operator running out-of-cluster. No AWS
> account or credentials are involved, and no ACK controllers are installed. **Integration
> testing against a live AWS environment is currently in progress and runs outside CI**, so
> cloud-side behaviour is not yet covered by the automated suite in this repository.

---

## How it works

```
   ┌──────────────────────┐        ┌───────────────────────┐
   │   KropathConfig      │        │  <Service>Config      │
   │  (org / namespace)   │        │  (per resource type)  │
   └──────────┬───────────┘        └───────────┬───────────┘
              │      mandatory / defaults / aws│
              └───────────────┬────────────────┘
                              ▼
                 ┌───────────────────────────┐
                 │   kropath-controller      │   ← this repo
                 │   <Service>Config          │
                 │   reconciler (cascade)    │
                 └─────────────┬─────────────┘
                               │  status.effectiveConfig
                               ▼
                 ┌───────────────────────────┐
                 │   kro RGD (kropath-aws)   │──▶ ACK CR ──▶ AWS
                 └───────────────────────────┘
```

Every config reconciler follows the same shape: watch `<Service>Config` plus `KropathConfig`,
run the merge helper in `internal/cascade`, and write `status.effectiveConfig` (ADR-010). The
merge is map-based and last-writer-wins per tier, so `mandatory` always beats a user's `spec`.

Two reconcilers are not config cascades:

- **PolicyDocument** resolves principal/resource refs to ARNs, merges statement sources, and
  writes `status.resolvedDocumentJSON`.
- **LabelOperator** applies provider resource-name labels (`aws.kropath.run/resource-name`, …)
  to CRs across all provider API groups, which is what makes the RGDs' `selector.matchLabels`
  config lookup resolve.

All reconcilers are always enabled. Feature availability is determined by which image version is
deployed, not by runtime flags — query `/features` or the generated
[`docs/features.yaml`](docs/features.yaml) to see what a given image contains.

## Implementation status

**23 reconcilers**, **23 CRD types** registered in `api/v1alpha1`, **25 Chainsaw suites**
covering **274 steps**, and **39 Go test files**.

**Suite** is the Chainsaw suite under `tests/`; **Steps** counts its named steps.
**AWS integration** tracks end-to-end validation against a live AWS account with real ACK
controllers — that work is in progress outside CI, so every entry is currently `⏳ Pending`.

| Reconciler | CR(s) watched | Suite | Steps | AWS integration |
|---|---|---|---|---|
| `ApiGatewayConfig` | `ApiGatewayConfig`, `KropathConfig` | `apigateway/ctrl-apigw-01` | 6 | ⏳ Pending |
| `ApiGatewayV2Config` | `ApiGatewayV2Config`, `KropathConfig` | `apigatewayv2/ctrl-apigwv2-01` | 10 | ⏳ Pending |
| `AutoScalingConfig` | `AutoScalingConfig`, `KropathConfig` | `autoscaling/ctrl-autoscaling-01` | 9 | ⏳ Pending |
| `CloudWatchLogsConfig` | `CloudWatchLogsConfig`, `KropathConfig` | `cloudwatchlogs/ctrl-cwl-01` | 11 | ⏳ Pending |
| `DynamoDBConfig` | `DynamoDBConfig`, `KropathConfig` | `dynamodb/ctrl-dynamodb-01` | 14 | ⏳ Pending |
| `EC2Config` | `EC2Config`, `KropathConfig` | `ec2/ctrl-ec2-01` | 15 | ⏳ Pending |
| `ECRConfig` | `ECRConfig`, `KropathConfig` | `ecr/ctrl-ecr-01` | 22 | ⏳ Pending |
| `ECSConfig` | `ECSConfig`, `KropathConfig` | `ecs/ctrl-ecs-01` | 3 | ⏳ Pending |
| `EFSConfig` | `EFSConfig`, `KropathConfig` | `efs/ctrl-efs-01` | 10 | ⏳ Pending |
| `EKSConfig` | `EKSConfig`, `KropathConfig` | `eks/ctrl-eks-01` | 16 | ⏳ Pending |
| `ElastiCacheConfig` | `ElastiCacheConfig`, `KropathConfig` | `elasticache/ctrl-elasticache-01` | 13 | ⏳ Pending |
| `ELBConfig` | `ELBConfig`, `KropathConfig` | **none** — see [Known gaps](#known-gaps) | — | ⏳ Pending |
| `EventBridgeConfig` | `EventBridgeConfig`, `KropathConfig` | `eventbridge/ctrl-eventbridge-01` | 10 | ⏳ Pending |
| `IAMConfig` | `IAMConfig`, `KropathConfig` | `iam/ctrl-iam-01` | 7 | ⏳ Pending |
| `KMSConfig` | `KMSConfig`, `KropathConfig` | `kms/ctrl-kms-01` | 15 | ⏳ Pending |
| `LabelOperator` | all provider API groups | `label-operator/ctrl-label-op-01` | 8 | ⏳ Pending |
| `PolicyDocument` | `PolicyDocument`, `KropathConfig` | `policy/phase2-refs`, `policy/phase3-merge` | 18 | ⏳ Pending |
| `RDSConfig` | `RDSConfig`, `KropathConfig` | `rds/ctrl-rds-01` | 16 | ⏳ Pending |
| `S3Config` | `S3Config`, `KropathConfig` | `s3/ctrl-s3-01` | 12 | ⏳ Pending |
| `SecretsManagerConfig` | `SecretsManagerConfig`, `KropathConfig` | `secretsmanager/ctrl-secretsmanager-01` | 15 | ⏳ Pending |
| `SNSConfig` | `SNSConfig`, `KropathConfig` | `sns/ctrl-sns-01` | 13 | ⏳ Pending |
| `SQSConfig` | `SQSConfig`, `KropathConfig` | `sqs/ctrl-sqs-01` | 13 | ⏳ Pending |
| `StepFunctionsConfig` | `StepFunctionsConfig`, `KropathConfig` | `stepfunctions/ctrl-sfn-01` | 11 | ⏳ Pending |

Two further suites cover the binary rather than a reconciler: `features/ctrl-features-01` (5
steps) exercises the `/features` endpoint and `version/ctrl-version-01` (2 steps) the build-info
metrics.

`docs/features.yaml` is generated from the registry by `make features-gen` and CI fails if it
drifts from the code (the **Feature registry drift gate** job).

### Known gaps

- **`ELBConfig` has no Chainsaw suite.** The reconciler, CRD type, cascade helper, and
  `tests/fixtures/crds/awselbconfig.yaml` all exist, but there is no `tests/elb/` directory and
  no `test-elb` Make target — it is the only reconciler with no integration coverage. (See
  `docs/troubleshooting-logs/2026-08-13-elbconfig-missing-crd-manager-crash.md` for the incident
  that followed from its CRD not being applied.)
- **Cascade helpers without reconcilers.** `internal/cascade/cloudfront.go` and
  `internal/cascade/lambda.go` are implemented and unit-tested, but no `CloudFrontConfig` or
  `LambdaConfig` reconciler or CRD type exists yet, so nothing calls them at runtime.
- **`make test-apigateway` is missing from the root `Makefile`.** `tests/apigateway/ctrl-apigw-01`
  runs under `make test-chainsaw` (which runs `chainsaw test tests/`) and via
  `tests/Makefile`, but there is no single-suite target at the repo root the way every other
  service has one.

## Requirements

| Tool | Version | Notes |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.26.6 | pinned in `go.mod`; keep in sync with `Makefile` and `.github/workflows/ci.yaml` |
| [kind](https://kind.sigs.k8s.io/) | v0.25.0 | local integration-test cluster |
| [Chainsaw](https://kyverno.github.io/chainsaw/) | v0.2.15 | integration test runner |
| [golangci-lint](https://golangci-lint.run/) | v2.11.4 | |
| Docker | — | container image build |

Install the pinned tool versions with:

```bash
make install-tools
```

## Building

```bash
make build           # compile bin/kropath-operator
make docker-build    # build the container image (tag = short git SHA + latest)
```

## Running locally

```bash
./bin/kropath-operator \
  --metrics-bind-address=:8080 \
  --health-probe-bind-address=:8081
```

| Flag | Default | Description |
|---|---|---|
| `--metrics-bind-address` | `:8080` | Prometheus `/metrics` endpoint |
| `--health-probe-bind-address` | `:8081` | `/healthz` and `/readyz` endpoints |

All reconcilers start automatically. The manager runs with leader election on by default
(`LEADER_ELECTION_NAMESPACE` or `POD_NAMESPACE` selects the lease namespace; defaults to
`default`).

## `/features` endpoint

`GET /features` on the metrics listener (`:8080`) returns version metadata and the live
reconciler list as JSON:

```bash
curl http://localhost:8080/features
```

```json
{
  "version": "v0.1.0",
  "gitCommit": "a1b2c3d",
  "buildDate": "2026-08-13T00:00:00Z",
  "goVersion": "go1.26.6",
  "features": [
    {
      "name": "IAMConfig",
      "package": "iamconfig",
      "description": "Reconciles IAMConfig CRs and propagates effective config.",
      "kinds": ["IAMConfig", "KropathConfig"],
      "sinceVersion": "v0.0.1",
      "stability": "stable"
    }
  ]
}
```

Filter by package name with `?name=<package>`:

```bash
# Returns the single kmsconfig entry or HTTP 404 if unknown.
curl 'http://localhost:8080/features?name=kmsconfig'
```

Only `GET` and `HEAD` are accepted; any other method returns `405`.

### `kropath-operator features` subcommand

Prints the same JSON as `GET /features` and exits — no kubeconfig or cluster connection needed.
Useful for inspecting an image before deploying it:

```bash
docker run --rm ghcr.io/kropath/kropath-controller:v0.1.0 features
```

or locally:

```bash
./bin/kropath-operator features | jq '.features | length'
```

## Testing

```bash
make test            # unit tests, race detector
make test-cover      # unit tests + HTML coverage report
make lint            # go vet + golangci-lint (required before every commit)
make features-verify # fail if docs/features.yaml has drifted from the registry
```

### Integration tests (Chainsaw + kind)

```bash
make test-chainsaw   # kind-up → build → apply CRDs → start operator → run all suites → stop
```

To iterate on a single suite without restarting the operator:

```bash
make chainsaw-start chainsaw-wait
make test-iam        # one target per service — see `make help` for the full list
make chainsaw-stop
```

The operator runs **out-of-cluster** against the kind cluster; CRDs come from
`tests/fixtures/crds/`. On a failed run, clean up manually with `make chainsaw-stop kind-down`.

Suites follow the canonical **unique-resource-name-per-step** pattern — see
[`docs/frequent-chainsaw-errors.md`](docs/frequent-chainsaw-errors.md) before writing or fixing
one.

## Security scans

Run only after implementation is complete and the image has been built — not during active
development:

```bash
make security        # gosec (SAST) + govulncheck (dependency CVEs)
```

## CI

`.github/workflows/ci.yaml` runs on pull requests and on push to `main`, with two exceptions:

- **Chainsaw integration tests** run on pull requests only. A push to `main` runs lint, unit
  tests, feature-registry drift check, security scans, and the image build — it does not
  re-run Chainsaw against an identical tree that already passed on the PR head.
- **Markdown-only changes** skip the expensive work but still report. Every job starts and
  reports a `success`; a `changes` job diffs the PR against its merge base and, when nothing
  outside `*.md` moved, each downstream job skips its own steps.

The second point is load-bearing and easy to get wrong. The `main` ruleset requires the `Unit
tests` and `Chainsaw integration tests` contexts, and a required context is only satisfied by a
run that reports. Filtering at the trigger — `on.pull_request.paths-ignore: '**/*.md'` — stops
the workflow from starting at all, so those contexts are never reported and a docs-only PR sits
on "Expected — Waiting for status to be reported" with no way to merge it. Hence the path check
lives inside the jobs.

Two consequences worth knowing:

- Changing a job's `name:` renames its status context and silently un-satisfies the ruleset.
  `Unit tests` and `Chainsaw integration tests` are load-bearing strings.
- The path check fails open: a missing or all-zero base SHA (new branch, force push) runs the
  full suite rather than assuming docs-only.

On push to `main`, the image build publishes to `ghcr.io/kropath/kropath-controller` — see
[Image tags](#image-tags) for the full tag matrix.

`.github/workflows/pr-title.yaml` is a separate workflow so that it carries no path filter of its
own — a docs-only PR is exactly the case that must be typed correctly to stay out of a release.

## Releases

Releases are fully automated via [release-please](https://github.com/googleapis/release-please):

1. Merge a `feat(...)` or `fix(...)` PR to `main`.
2. release-please opens a release PR accumulating unreleased changes and proposing the next
   semver version.
3. A human reviews and merges the release PR.
4. Merging the release PR creates the git tag, the GitHub Release, and the `CHANGELOG.md` entry.
5. The tag triggers `release.yaml`, which builds and pushes the versioned image.

PRs are squash-merged, so the **PR title becomes the commit subject on `main`** and is the only
input release-please parses. Titles must be conventional commits with the ticket id as the scope:

```
feat(KRO-637): restore feature-registry metadata
fix(KRO-641): guard nil effectiveConfig on first reconcile
docs(KRO-650): document blocked_by metadata format
```

`pr-title.yaml` enforces this. The older `[KRO-637]: feat: …` form is rejected: the bracketed
prefix breaks the conventional-commit header regex, so the commit parses with no type, never bumps
the version, and never reaches `CHANGELOG.md`.

Which types cut a release:

| Type | Effect |
|---|---|
| `feat` | minor bump |
| `fix`, `perf`, `deps` | patch bump |
| `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `revert` | **no release** |

`release.yaml` keeps its unfiltered push trigger, but release-please is a no-op on a run of
non-releasable commits — a batch of doc-only merges opens no release PR, so step 2 above never
fires and nothing is published.

Keep the type list in `pr-title.yaml` in sync with `changelog-sections` in
`release-please-config.json`.

### Image tags

| Trigger | Tags pushed |
|---|---|
| Push to `main` | `:latest`, `:sha-<short7>` |
| Release tag `vX.Y.Z` | `:vX.Y.Z`, `:sha-<short7>`, `:latest` |

## Repository layout

| Path | Contents |
|---|---|
| `api/v1alpha1/` | CRD Go types (23 kinds) + scheme registration — group `aws.kropath.run` |
| `cmd/manager/` | `main.go` — flag parsing, manager wiring, `features` subcommand |
| `cmd/gen-features/` | generates `docs/features.yaml` from the reconciler registry |
| `internal/reconciler/` | one package per reconciler (23 packages + `util`) |
| `internal/cascade/` | shared config-merge helpers, one file per service |
| `internal/features/` | the feature registry and the `/features` HTTP handler |
| `internal/version/` | build-info and feature-enabled Prometheus metrics |
| `config/rbac/` | ClusterRole manifests for the manager, PolicyDocument, and LabelOperator |
| `tests/` | Chainsaw integration suites + `tests/fixtures/crds/` |
| `docs/` | engineering standards, the Chainsaw error catalog, and dated troubleshooting logs |

## Documentation

| Doc | Purpose |
|---|---|
| [`docs/STANDARDS.md`](docs/STANDARDS.md) | The engineering standards that bind this repo |
| [`docs/frequent-chainsaw-errors.md`](docs/frequent-chainsaw-errors.md) | Catalog of Chainsaw/controller test traps already hit — read before fixing a suite |
| [`docs/features.yaml`](docs/features.yaml) | Generated reconciler registry (do not edit by hand) |
| [`docs/troubleshooting-logs/`](docs/troubleshooting-logs/) | Dated per-incident fix logs |
| [`CLAUDE.md`](CLAUDE.md) | Repo conventions and working loop, for both humans and coding agents |

## Related repositories

| Repo | Role |
|---|---|
| [kropath-aws](https://github.com/kropath/kropath-aws) | kro RGDs and governance CRD manifests that consume `status.effectiveConfig` |

## Contributing

Bug fixes and small changes are welcome as pull requests. Feature requests and architectural
changes should be raised as a GitHub Issue — accepted requests go onto the development roadmap and
are implemented by the maintainers; feature PRs are not being accepted yet. See
[CONTRIBUTION.md](CONTRIBUTION.md).

## License

Apache License 2.0 — see [LICENSE](LICENSE).
