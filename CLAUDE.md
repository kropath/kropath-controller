@docs/STANDARDS.md

## This Repo

**kropath-controller** — Multi-reconciler `kropath-operator` binary. One `controller-runtime` Manager, one Reconciler struct per feature, each enabled by a feature flag. See `kropath-core/docs/design/policy-document-crd.md` §11 (P-1) for the architecture decision.

### Reconciler 1 — Config merge (ADR-010)

- Watches `KropathConfig` (global + namespace) and `ResourceConfig` (global + namespace) CRs
- Merges all four sources and writes `status.effectiveConfig` onto namespaced ResourceConfig CRs
- Merge priority (mandatory): M-1 globalKropathCfg → M-2 → M-3 → M-4 → M-5 → M-6 localResourceCfg
- Merge priority (defaults): D-1 localResourceCfg → D-2 → D-3 → D-4 → D-5 → D-6 globalKropathCfg
- Also copies provider identity (`aws.accountId`, `aws.region`) into `effCfg.aws.*` (ADR-010 D-3)
- Unit tests required for merge logic; integration tests required for full controller reconcile path

### Reconciler 2 — PolicyDocument (`--enable-poldoc`)

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
- Feature flags: `--enable-poldoc`, `--enable-sequencer` (future) — allows incremental rollout

### Before Creating a PR

Run these and confirm all pass before opening a pull request:

```bash
make lint
make test-cover
make test-chainsaw
make security
```
