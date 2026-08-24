# tests/fixtures/crds-optional

CRD fixtures that must be **absent** at operator startup for specific dynamic-detection suites go here.

## The rule

A CRD fixture belongs in `tests/fixtures/crds/` (the default) **unless** a dynamic-detection suite requires that CRD to be missing when the operator starts. In that case, place the fixture in this directory instead.

| Directory | Applied by `chainsaw-setup` | In `kubectl wait` list |
|---|---|---|
| `tests/fixtures/crds/` | Yes | Yes |
| `tests/fixtures/crds-optional/` | **No** | **No** |

## How suites use optional CRDs

A suite that needs to install an optional CRD at a specific point during the test calls:

```makefile
make chainsaw-install-optional-crd CRD_FILE=tests/fixtures/crds-optional/<name>.yaml
```

This applies the CRD and waits for `Established` before returning.

## Unit test coverage

Both `TestEveryReconcilerHasCRDFixture` and `TestReconcilerKindMatchesFixtureAndScheme` in
`internal/features/features_test.go` glob **both** directories, so a fixture placed here
satisfies those checks exactly as one in `crds/` would.
