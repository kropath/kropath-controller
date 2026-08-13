# Unique-Name-Per-Step Rollout Across 14 Chainsaw Suites

**Date:** 2026-08-13
**Issue:** KRO-635 (PR #35)
**Scope:** `tests/eks`, `tests/autoscaling`, `tests/cloudwatchlogs`, `tests/dynamodb`,
`tests/ec2`, `tests/efs`, `tests/elasticache`, `tests/iam`, `tests/kms`, `tests/rds`,
`tests/s3`, `tests/secretsmanager`, `tests/sns`, `tests/sqs`

## Starting point

Every suite reused a single resource name (`general-policy`, `general-table`,
`general-bus`, …) across all its acceptance criteria, and used deletes to reset state
between them. Across `tests/` that came to **439 per-step `cleanup:` deletes**, 13
inline `delete:` reset steps, and a set of `kubectl delete` reset scripts including
three `purge-stale-leftovers` steps.

`tests/eks/ctrl-eks-01` was the worst case: one `general-policy` name across all 16
ACs, and **16 interleaved inline `delete` steps** — each commented as working around
an SSA-PATCH retry loop — because Chainsaw defers `cleanup:` blocks to the end of the
file in LIFO order, so a mid-suite reset has to be a `try:` step.

## Why unique names work here

The controller resolves a config's global counterpart **by name**:

```go
globalKropath, _ := r.loadKropathConfig(ctx, kroSystemNamespace, cfg.Name)
globalCWL, _     := r.loadCloudWatchLogsConfig(ctx, kroSystemNamespace, cfg.Name)
```

`payments-prod/ac3-eks-policy` merges only `kro-system/ac3-eks-policy`. A per-AC name
is therefore a complete isolation boundary in both directions, and every reason for
the deletes disappears. This is the pattern from kropath-aws
`docs/frequent-rgd-errors.md` §6, adapted — see the deviations below.

Layout per step: one `ac<NN>-setup.yaml` holding that step's inputs as a
multi-document YAML, plus one `ac<NN>-assert.yaml`.

## Result

| | Before | After |
|---|---|---|
| `tests/eks` cold runtime | 7.8s | 3.1s |
| Full `make test-chainsaw` | 205s | 179s |
| Per-step deletes across `tests/` | 439 | 0 |
| Suites passing | 22/22 | 22/22 |

The EKS figure is a true A/B: both measured cold on the same cluster with the suite's
objects deleted beforehand, with the old suite restored from git in between. An
earlier 18.7s reading for the new suite was a measurement error — it was the first run
after a controller restart, so it included manager warm-up, not suite cost.

Coverage was verified rather than assumed: `assert` and `expect` counts are identical
before and after for all 13 rolled-out suites. The only step-count changes are the
three dropped `purge-stale-leftovers` steps.

## Two deviations from the kropath-aws pattern

**`skipDelete: true` is not used here.** In kropath-aws it exists to avoid
cleanup-phase timeouts caused by ACK finalizers that no controller ever removes. This
repo has no ACK controllers *and* no finalizers on these objects — they are plain
config CRs with no cloud calls — so Chainsaw's end-of-test cleanup is cheap. It was
tried anyway, copied from the reference, and made things worse: leaving every suite's
objects in the cluster for a whole run put enough steady reconcile load on the API
server that later suites failed with
`client rate limiter Wait returned an error: would exceed context deadline`.

**Negative-path applies keep their own file.** See below.

## Three mistakes made during the conversion

Recorded because two of them fail in misleading ways.

### 1. Reset scripts were renamed instead of removed

The conversion renamed resource names inside `script:` bodies to keep `kubectl`
references consistent. Applied to a *reset* script, that is actively harmful:

```yaml
# before — deletes the previous step's leftovers
- script:
    content: kubectl delete kropathconfigs.aws.kropath.run general-policy -n kro-system --ignore-not-found=true
# after renaming — deletes the config this step just applied
- script:
    content: kubectl delete kropathconfigs.aws.kropath.run ac11-general-policy -n kro-system --ignore-not-found=true
```

`ec2` AC-11 then asserted `flowLogsRequired: true` against a config whose
`KropathConfig` had just been deleted, and got `false`. Reset scripts must be
*dropped*, exactly like the inline `delete:` steps — a script whose every line is
`kubectl delete` is a reset, never an assertion.

### 2. `expect:` blocks were dropped from negative-path applies

This was the expensive one. Several suites assert that a deliberately invalid manifest
is rejected:

```yaml
- apply:
    file: 17-kropathconfig-both-tiers-boundary.yaml
    expect:
      - match: {apiVersion: aws.kropath.run/v1alpha1, kind: KropathConfig}
        check: {($error != null): true}
```

Collapsing a step's applies into one `ac<NN>-setup.yaml` kept only `file:` and
discarded `expect:`. The failure mode is badly misleading: the apply is *supposed* to
fail, so without `expect:` the step fails, Chainsaw retries the apply, and the retry
loop surfaces as

```
client rate limiter Wait returned an error: rate: Wait(n=1) would exceed context deadline
```

in three of the four affected suites. Only `iam` surfaced the real message
(`permissionsBoundaryArn must be set in either mandatory or defaults, not both`), and
only after `--apply-timeout` had been raised enough for the underlying error to win
the race against the deadline.

That misread cost a wrong fix: raising `--apply-timeout`/`--delete-timeout` in
`CHAINSAW_FLAGS` on the theory that the extra objects were queueing behind client-go's
default 5 QPS. It fixed nothing and doubled the run to 414s. Reverted.

Correct handling: a step containing any apply with options beyond `file:` keeps one
file per apply (`ac<NN>-setup-<i>.yaml`) so each keeps its own `expect:`.

### 3. Verb order was not preserved

The first converter emitted all applies, then other verbs, then asserts. Where a step
had a `script:` before its applies, that reordering changed behaviour. Fixed by
recording each verb's original index and emitting in place.

## Suites deliberately left unconverted

`apigatewayv2`, `ecr`, `ecs`, `eventbridge`, `stepfunctions`, `label-operator`, and
both `policy` suites. All pass and are unchanged.

- The four script-driven suites (`apigatewayv2`, `eventbridge`, `stepfunctions`,
  `ecr`) drive assertions from `script:` bodies with resource names hardcoded in
  `kubectl`/`jq`. They still carry ~30 `kubectl delete` reset scripts between steps —
  this is the remaining cleanup if the pattern is finished later.
- The two `policy` suites are the risky ones: `PolicyDocument` resolves
  `spec.statements[].principals[].ref` and `.resources[].ref` **by name**, so renaming
  a resource without renaming every ref pointing at it breaks resolution silently.
  These need hand conversion, not a mechanical one.

## Tooling note

The conversion was done by a script (parse `chainsaw-test.yaml`, group each step's
files, rewrite names by whole-token substitution longest-first so `general-policy`
cannot corrupt `general-policy-defaults`). It is not committed — it is a one-shot
migration tool, and re-running it against already-converted suites is meaningless.
The pattern it produces is documented in `docs/frequent-chainsaw-errors.md`.
