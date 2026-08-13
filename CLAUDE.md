@docs/STANDARDS.md

## This Repo

**kropath-controller** — Multi-reconciler `kropath-operator` binary. One `controller-runtime` Manager, one Reconciler struct per feature. Every reconciler registered in `internal/features.All` runs unconditionally — there are no per-feature flags. See `kropath-core/docs/design/policy-document-crd.md` §11 (P-1) for the architecture decision.

### Reconciler 1 — Config merge (ADR-010)

- Watches `KropathConfig` (global + namespace) and `ResourceConfig` (global + namespace) CRs
- Merges all four sources and writes `status.effectiveConfig` onto namespaced ResourceConfig CRs
- Merge priority (mandatory): M-1 globalKropathCfg → M-2 → M-3 → M-4 → M-5 → M-6 localResourceCfg
- Merge priority (defaults): D-1 localResourceCfg → D-2 → D-3 → D-4 → D-5 → D-6 globalKropathCfg
- Also copies provider identity (`aws.accountId`, `aws.region`) into `effCfg.aws.*` (ADR-010 D-3)
- Unit tests required for merge logic; integration tests required for full controller reconcile path

### Reconciler 2 — PolicyDocument

- Watches `PolicyDocument` CRs and `KropathConfig` CRs (account ID / region lookup)
- Resolves `spec.statements[].principals[].ref` and `spec.statements[].resources[].ref` to ARNs using ARN prediction; falls back to `status.arn` for non-predictable kinds
- Merges Statement arrays from `spec.sources` (left-ordered concatenation); detects Sid conflicts
- Serializes resolved document to `status.resolvedDocumentJSON`; sets `Ready`, `SidConflict`, `SourceNotReady` conditions
- Raw `spec.documentJSON` path: validate JSON and pass through; not mergeable as a source
- Exposes Prometheus metrics: `kropath_poldoc_reconcile_total`, `kropath_poldoc_unresolved_refs`, etc.
- Test suite at `tests/policy/` (three phases: CRD validation, ref resolution, source merge)

### Planned reconcilers (not yet implemented)

- **Composite RGD sequencer** — apply child CRs in dependency order
- **Config cascade validator** — surface conflicts across `KropathConfig` / `ResourceConfig` namespace hierarchy
- **Cross-RGD status aggregator** — roll up child CR ready conditions to parent status

### Shared operator concerns

- Supports multiple replicas with leader election
- Single Deployment, single RBAC manifest, single `/metrics` endpoint (port 8080)
- Health probes: `/healthz` port 8081 (manager alive), `/readyz` port 8081 (leader lease + watches established)
- `/features` endpoint on `:8080` — returns version, git commit, and the live reconciler list as JSON
- **No per-feature flags.** Every reconciler in `internal/features.All` runs unconditionally. To add a reconciler: create its package under `internal/reconciler/<pkg>/`, add an entry to `features.All`, and run `make generate-features`. Missing registrations fail `TestRegistryCoversAllPackages`.
- **Every watched kind must have a CRD in `tests/fixtures/crds/`.** With the flags retired, a reconciler whose CRD is missing from the test cluster takes down the **whole manager** two minutes after startup, which surfaces as unrelated Chainsaw suites timing out. `TestEveryReconcilerHasCRDFixture` catches it at unit-test time; `make chainsaw-setup` derives its `kubectl wait` list from the fixture directory. See `docs/frequent-chainsaw-errors.md` §1.

### Chainsaw tests

**Read `docs/frequent-chainsaw-errors.md` before writing or debugging a Chainsaw suite.**
It covers the failure modes that are specific to this repo and expensive to rediscover:
a missing CRD killing the manager mid-run, why suites use a unique resource name per
step, why `skipDelete` does *not* transfer from kropath-aws, why a dropped `expect:`
block masquerades as a rate-limiter error, and order-stable list assertions.

The short version for a new suite: give every acceptance criterion its own resource
name (`ac<N>-<family>-policy`), delete nothing between steps, and use one
`ac<NN>-setup.yaml` plus one `ac<NN>-assert.yaml` per AC. `tests/eks/ctrl-eks-01/` is
the reference layout. The controller pairs a namespaced config with its `kro-system`
counterpart *by name*, so a per-step name is a complete isolation boundary.

Dated write-ups of specific incidents live in `docs/troubleshooting-logs/`.

### Before Creating a PR

Run these and confirm all pass before opening a pull request:

```bash
make lint
make test-cover
make test-chainsaw
make security
```
