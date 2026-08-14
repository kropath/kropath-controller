# kropath-controller

Go controller for [kropath](https://github.com/kropath) (kro + golden path), a multi-cloud golden
path platform. `kropath-controller` is a multi-reconciler `kropath-operator` binary built on a
single `controller-runtime` Manager — one Reconciler struct per feature.

See `docs/STANDARDS.md` for the standards that bind this repo.

## Reconcilers

All reconcilers are always enabled. Feature availability is determined by which image version is deployed, not by runtime flags. For the full list of reconcilers active in a given image, query the `/features` endpoint or see the generated `docs/features.yaml`.

| Reconciler | CR(s) watched | Purpose |
|---|---|---|
| IAM config merge | `KropathConfig`, `IAMConfig` | Merges org/namespace/instance config layers, writes `status.effectiveConfig` (ADR-010) |
| S3 config merge | `KropathConfig`, `S3Config` | Same merge pattern, scoped to S3 resource config |
| KMS config cascade | `KropathConfig`, `KMSConfig` | Cascades KMS key-spec and policy config to `status.effectiveConfig` |
| PolicyDocument | `PolicyDocument`, `KropathConfig` | Resolves principal/resource refs to ARNs, merges statement sources, writes `status.resolvedDocumentJSON` |

Config-merge reconcilers write the single `status.effectiveConfig` object that kro RGDs read via
one `externalRef` lookup per RGD (see the CEL cascade pattern in `docs/STANDARDS.md`).

## Requirements

- Go 1.26.6 (pinned in `go.mod`; keep in sync with `Makefile` / `.github/workflows/ci.yaml`)
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
  --health-probe-bind-address=:8081
```

| Flag | Default | Description |
|---|---|---|
| `--metrics-bind-address` | `:8080` | Prometheus `/metrics` endpoint |
| `--health-probe-bind-address` | `:8081` | `/healthz` and `/readyz` endpoints |

All reconcilers start automatically. To see which reconcilers are active, use the `/features` endpoint or the `features` subcommand (see below).

The manager runs with leader election on by default (`LEADER_ELECTION_NAMESPACE` or
`POD_NAMESPACE` env var selects the lease namespace; defaults to `default`).

## `/features` endpoint

`GET /features` on the metrics listener (`:8080`) returns version metadata and the live reconciler list as JSON:

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

## `kropath-operator features` subcommand

Prints the same JSON as `GET /features` and exits — no kubeconfig or cluster connection needed. Useful for inspecting an image before deploying it:

```bash
docker run --rm ghcr.io/kropath/kropath-controller:v0.1.0 features
```

or locally:

```bash
./bin/kropath-operator features | jq '.features | length'
```

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

`.github/workflows/ci.yaml` runs on pull requests and on push to `main`, with two exceptions:

- **Chainsaw integration tests** run on pull requests only. A push to `main` runs lint, unit
  tests, feature-registry drift check, security scans, and the image build — it does not
  re-run Chainsaw against an identical tree that already passed on the PR head.
- **Markdown-only changes** skip CI entirely. A commit or PR that touches only `*.md` files
  produces no workflow run. (Branch protection is not enabled on `main`, so skipped runs do
  not block merges.)

On push to `main`, the image build publishes `ghcr.io/kropath/kropath-controller:latest` and
`ghcr.io/kropath/kropath-controller:sha-<short>`.

`.github/workflows/pr-title.yaml` runs on every pull request — including Markdown-only ones, which
is why it is a separate workflow from `ci.yaml`.

## Releases

Releases are fully automated via [release-please](https://github.com/googleapis/release-please):

1. Merge a `feat(...)` or `fix(...)` PR to `main`.
2. release-please opens a release PR accumulating unreleased changes and proposing the next
   semver version.
3. A human reviews and merges the release PR.
4. Merging the release PR creates the git tag, the GitHub Release, and the `CHANGELOG.md` entry.
5. The tag triggers `release.yaml`, which builds and pushes the versioned image.

The seed tag `v0.0.1` establishes the baseline; the first automated release will be `v0.1.0`
(for a `feat`) or `v0.0.2` (for a `fix`).

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
non-releasable commits — a batch of doc-only merges opens no release PR. A release is cut when the
accumulated release PR is merged; `build-release-image` then publishes the versioned image.

Keep the type list in `pr-title.yaml` in sync with `changelog-sections` in
`release-please-config.json`.

### Image tags

| Trigger | Tags pushed |
|---|---|
| Push to `main` | `:latest`, `:sha-<short7>` |
| Release tag `vX.Y.Z` | `:vX.Y.Z`, `:sha-<short7>`, `:latest` |

## Repository layout

```
api/v1alpha1/          CRD Go types + scheme registration
cmd/manager/           main.go — flag parsing, manager wiring
internal/reconciler/   one package per reconciler (iamconfig, s3config, kmsconfig, policydocument)
internal/cascade/      shared config-merge helpers
tests/                 Chainsaw integration suites (iam, s3, kms, policy)
docs/                  engineering standards and design docs
```
