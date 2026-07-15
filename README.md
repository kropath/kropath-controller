# kropath-controller

Go controller for [kropath](https://github.com/kropath) (kro + golden path), a multi-cloud golden
path platform. `kropath-controller` is a multi-reconciler `kropath-operator` binary built on a
single `controller-runtime` Manager — one Reconciler struct per feature, each independently
enabled by a feature flag.

See `docs/STANDARDS.md` and ADR-015 (`kropath-core`) for the full architecture reference.

## Reconcilers

| Reconciler | CR(s) watched | Flag | Purpose |
|---|---|---|---|
| IAM config merge | `KropathConfig`, `IAMConfig` | always on | Merges org/namespace/instance config layers, writes `status.effectiveConfig` (ADR-010) |
| S3 config merge | `KropathConfig`, `S3Config` | always on | Same merge pattern, scoped to S3 resource config |
| KMS config cascade | `KropathConfig`, `KMSConfig` | `--enable-kms-cascade` | Cascades KMS key-spec and policy config to `status.effectiveConfig` |
| PolicyDocument | `PolicyDocument`, `KropathConfig` | `--enable-poldoc` | Resolves principal/resource refs to ARNs, merges statement sources, writes `status.resolvedDocumentJSON` |

Config-merge reconcilers write the single `status.effectiveConfig` object that kro RGDs read via
one `externalRef` lookup per RGD (see the CEL cascade pattern in `docs/STANDARDS.md`).

## Requirements

- Go 1.26.5 (pinned in `go.mod`; keep in sync with `Makefile` / `.github/workflows/ci.yaml`)
- [kind](https://kind.sigs.k8s.io/) v0.25.0 — local integration-test cluster
- [Chainsaw](https://kyverno.github.io/chainsaw/) v0.2.15 — integration test runner
- [golangci-lint](https://golangci-lint.run/) v2.11.4
- Docker — for building the container image

Install the pinned tool versions with:

```bash
make install-tools
```

## Building

```bash
make build          # compile bin/kropath-operator
make docker-build    # build the container image (tag = short git SHA + latest)
```

## Running locally

```bash
./bin/kropath-operator \
  --metrics-bind-address=:8080 \
  --health-probe-bind-address=:8081 \
  --enable-poldoc \
  --enable-kms-cascade
```

| Flag | Default | Description |
|---|---|---|
| `--metrics-bind-address` | `:8080` | Prometheus `/metrics` endpoint |
| `--health-probe-bind-address` | `:8081` | `/healthz` and `/readyz` endpoints |
| `--enable-poldoc` | `false` | Enable the PolicyDocument reconciler |
| `--enable-kms-cascade` | `false` | Enable the KMSConfig cascade reconciler |

The manager runs with leader election on by default (`LEADER_ELECTION_NAMESPACE` or
`POD_NAMESPACE` env var selects the lease namespace; defaults to `default`).

## Testing

```bash
make test           # unit tests, race detector
make test-cover      # unit tests + HTML coverage report
make lint            # go vet + golangci-lint (required before every commit)
```

### Integration tests (Chainsaw + kind)

```bash
make test-chainsaw   # kind-up → build → apply CRDs → start operator → run all suites → stop
```

To iterate on a single suite without restarting the operator:

```bash
make chainsaw-start chainsaw-wait
make test-iam        # or test-s3, test-kms, test-policy
make chainsaw-stop
```

On a failed run, clean up manually with `make chainsaw-stop kind-down`.

## Security scans

Run only after implementation is complete and the image has been built — not during active
development:

```bash
make security        # gosec (SAST) + govulncheck (dependency CVEs)
```

## CI

`.github/workflows/ci.yaml` runs lint, unit tests, security scans, Chainsaw integration tests, and
(on push to `main`) builds and pushes the container image to
`ghcr.io/kropath/kropath-controller`.

## Repository layout

```
api/v1alpha1/          CRD Go types + scheme registration
cmd/manager/           main.go — flag parsing, manager wiring
internal/reconciler/   one package per reconciler (iamconfig, s3config, kmsconfig, policydocument)
internal/cascade/      shared config-merge helpers
tests/                 Chainsaw integration suites (iam, s3, kms, policy)
docs/                  engineering standards and design docs
```
