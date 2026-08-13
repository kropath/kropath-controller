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
- **Every watched kind must have a CRD in `tests/fixtures/crds/`.** Since the flags were retired, a reconciler whose CRD is absent from the test cluster is no longer harmless: its informer never syncs and controller-runtime **kills the whole manager** once the 2-minute cache-sync timeout elapses. The operator serves traffic for two minutes and then exits, so the symptom is that every Chainsaw suite running after the ~2-minute mark fails on a 30s assert timeout while earlier suites pass — the set of "failing" families is decided by machine speed and test order, not by anything wrong with those families. `TestEveryReconcilerHasCRDFixture` now catches this at unit-test time. `make chainsaw-setup` derives its `kubectl wait` list from the fixture directory, so adding the CRD file is the only step required.

### Adding a Chainsaw suite — unique name per acceptance criterion

**Give every AC its own resource name (`ac<N>-<family>-policy`) and do not delete anything between steps.**

**Why:** the controller resolves a config's global counterpart **by name** — `payments-prod/ac3-eks-policy` merges only `kro-system/ac3-eks-policy` — so a per-AC name isolates each AC completely and no AC can shadow another in either direction. Suites that reuse one name across ACs need interleaved `delete` "reset" steps to stop the previous AC's values leaking into the next, and Chainsaw defers `cleanup:` blocks to the end of the file (LIFO), so those resets have to be inline `try:` steps. That is pure overhead and makes the suite order-fragile. Converting `tests/eks/ctrl-eks-01` from one reused `general-policy` name plus 16 inline deletes to unique per-AC names with zero deletes cut its cold runtime from 7.8s to 3.1s; across all 14 converted suites the full run went from 205s to 179s with 439 per-step deletes removed.

**How to apply:** one `ac<NN>-setup.yaml` (all inputs for that AC as a multi-document YAML) plus one `ac<NN>-assert.yaml` per AC. See `tests/eks/ctrl-eks-01/` for the reference layout and kropath-aws `docs/frequent-rgd-errors.md` §6 for the fuller rationale.

Two deliberate deviations from the kropath-aws reference pattern:

- **Do not set `spec.skipDelete: true`.** It exists there to dodge cleanup-phase timeouts caused by ACK finalizers, which this repo has none of — these are plain config CRs with no finalizers and no cloud calls, so Chainsaw's end-of-test cleanup is cheap. Letting it reclaim each suite's objects keeps the shared kind cluster small.
- **A negative-path step keeps its own file.** An `apply:` carrying `expect:` (a deliberately invalid manifest asserting the API server rejects it) must not be merged into a shared `ac<NN>-setup.yaml`, because `expect:` applies per-manifest. Give those a dedicated `ac<NN>-setup-<i>.yaml`. Dropping the `expect:` block does not fail loudly — the rejected apply just fails the step, and the retry loop surfaces as a confusing `client rate limiter ... context deadline exceeded` rather than a validation error.

### Chainsaw Test Assertion Stability

**Never assert a list/array field by exact position when its element order is not guaranteed deterministic — use item-level `(?...)` matching instead.**

**Why:** Any list built by iterating a Go map (or, in the sibling `kropath-aws` repo, a kro CEL `.merge().transformList()` chain) has iteration order that is not guaranteed stable across runs. A positional chainsaw `assert` on such a list is a latent flake — it can pass locally and fail in CI, or pass in isolation and fail under parallel execution, purely from map-order nondeterminism.

**How to apply:** This repo's current `effectiveConfig.mandatory.tags` / `.defaults.tags` / `syncedLabels` / `syncedAnnotations` fields are all **maps**, not lists, so today's asserts (e.g. `tests/kms/ctrl-kms-01/33-assert-ac15.yaml`) are order-stable and need no change. `PolicyDocument` statement merges are also stable — `spec.sources` are concatenated in source order, not iterated from a map. But if a future change adds or asserts a **list built from a map** (e.g. a config field ever serialized as `[]{key,value}`, or a merge algorithm that starts ranging over a map), pair a length check with a per-item match instead of a positional list:

```yaml
spec:
  (length(tags)): 2
  tags:
    - (key == 'cost-centre'): true
      value: platform
    - (key == 'environment'): true
      value: mandatory
```

See `kropath-aws/docs/frequent-rgd-errors.md` §6 "Flaky List/Array Asserts — CEL Map-to-List Transforms Have Unstable Order" for the fuller writeup and the pattern applied to that repo's kro RGD chainsaw tests.

### Before Creating a PR

Run these and confirm all pass before opening a pull request:

```bash
make lint
make test-cover
make test-chainsaw
make security
```
