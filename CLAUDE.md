@docs/STANDARDS.md

## This Repo

**kropath-controller** — Go controller that implements the ADR-010 config merge logic.

- Watches `KropathConfig` (global + namespace) and `ResourceConfig` (global + namespace) CRs
- Merges all four sources and writes `status.effectiveConfig` onto namespaced ResourceConfig CRs
- Merge priority (mandatory): M-1 globalKropathCfg → M-2 → M-3 → M-4 → M-5 → M-6 localResourceCfg
- Merge priority (defaults): D-1 localResourceCfg → D-2 → D-3 → D-4 → D-5 → D-6 globalKropathCfg
- Also copies provider identity (`aws.accountId`, `aws.region`) into `effCfg.aws.*` (ADR-010 D-3)
- Supports multiple replicas with leader election
- Unit tests required for merge logic; integration tests required for full controller reconcile path
