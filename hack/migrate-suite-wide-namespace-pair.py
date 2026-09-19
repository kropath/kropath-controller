#!/usr/bin/env python3
"""Migrate a numbered-file Chainsaw suite (ecr/ecrpublic/ecs/eventbridge/sfn
style -- one chainsaw-test.yaml driving many NN-*.yaml fixtures that all
reuse a single object name, deleted and recreated step to step) to the
singleton `baseline` KropathConfig name (KRO-1137).

Unlike the per-AC `*setup*.yaml` suites migrate-kropathconfig-baseline.py
handles, these suites already delete-and-recreate their one shared
KropathConfig/object between steps, so no per-AC namespace pair is needed --
but the whole suite still needs its OWN dedicated global-ns/app-ns pair so
its transient `baseline` KropathConfig in kro-system can't collide with any
other already-migrated suite running in the same `make test-chainsaw` pass.

For every *.yaml file in the suite dir except chainsaw-test.yaml and any
already-written 00-namespaces.yaml:
  - every `kind: KropathConfig` document is renamed to `baseline`
  - every document (any kind) in the old global namespace (kro-system)
    moves to the new global-ns
  - every document (any kind) in the old local namespace moves to the new
    app-ns

Writes 00-namespaces.yaml with the two new Namespace objects (app-ns
annotated with the global-config-namespace pointer) and prepends an apply
of it to chainsaw-test.yaml's first step. Also plain-text-substitutes the
namespace literals used in chainsaw-test.yaml's own `script:`/`namespace:`
fields (kubectl delete commands, spec.namespace), and the shared object's
name specifically where it appears in a `kropathconfig(s)` delete command.

Usage:
  python3 hack/migrate-suite-wide-namespace-pair.py tests/ecr/ctrl-ecr-01 \
      --family ecr --provider aws --old-local-ns registry-prod \
      --old-name general-policy
"""
import argparse
import pathlib
import re

from ruamel.yaml import YAML

OLD_GLOBAL_NS = "kro-system"

yaml = YAML()
yaml.preserve_quotes = True
yaml.indent(mapping=2, sequence=2, offset=0)
yaml.width = 4096


def migrate_fixture_files(suite_dir: pathlib.Path, global_ns: str, app_ns: str, old_local_ns: str):
    changed = []
    for p in sorted(suite_dir.glob("*.yaml")):
        if p.name in ("chainsaw-test.yaml", "00-namespaces.yaml"):
            continue
        with open(p) as f:
            docs = list(yaml.load_all(f))
        touched = False
        for d in docs:
            if not d:
                continue
            if d.get("kind") == "KropathConfig" and d.get("metadata", {}).get("name") != "baseline":
                d["metadata"]["name"] = "baseline"
                touched = True
            ns = d.get("metadata", {}).get("namespace")
            if ns == OLD_GLOBAL_NS:
                d["metadata"]["namespace"] = global_ns
                touched = True
            elif ns == old_local_ns:
                d["metadata"]["namespace"] = app_ns
                touched = True
        if touched:
            with open(p, "w") as f:
                first = True
                for d in docs:
                    if d is None:
                        continue
                    if not first:
                        f.write("---\n")
                    first = False
                    yaml.dump(d, f)
            changed.append(str(p))
    return changed


def write_namespaces_file(suite_dir: pathlib.Path, global_ns: str, app_ns: str, provider: str):
    p = suite_dir / "00-namespaces.yaml"
    annotation_key = f"{provider}.kropath.run/global-config-namespace"
    content = f"""apiVersion: v1
kind: Namespace
metadata:
  name: {global_ns}
---
apiVersion: v1
kind: Namespace
metadata:
  name: {app_ns}
  annotations:
    {annotation_key}: {global_ns}
"""
    p.write_text(content)
    return p


def patch_chainsaw_test(path: pathlib.Path, global_ns: str, app_ns: str, old_local_ns: str, old_name: str):
    text = path.read_text()

    # kubectl ... -n <old-ns> --> -n <new-ns>
    text = text.replace(f"-n {OLD_GLOBAL_NS} ", f"-n {global_ns} ")
    text = re.sub(rf"-n {re.escape(OLD_GLOBAL_NS)}(?=[\s\"'])", f"-n {global_ns}", text)
    text = re.sub(rf"-n {re.escape(old_local_ns)}(?=[\s\"'])", f"-n {app_ns}", text)

    # Test-level default namespace (eventbridge/sfn style: `namespace: <old_local_ns>`)
    text = re.sub(
        rf"^(\s*namespace:\s*){re.escape(old_local_ns)}\s*$",
        rf"\g<1>{app_ns}",
        text,
        flags=re.MULTILINE,
    )

    # kropathconfig(s)[.aws.kropath.run] <old_name> -> ... baseline ...
    text = re.sub(
        rf"(kropathconfigs?(?:\.aws\.kropath\.run)?\s+){re.escape(old_name)}\b",
        rf"\1baseline",
        text,
    )

    return text


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("suite_dir")
    ap.add_argument("--family", required=True)
    ap.add_argument("--provider", default="aws")
    ap.add_argument("--old-local-ns", required=True)
    ap.add_argument("--old-name", required=True, help="the reused object name, e.g. general-policy")
    args = ap.parse_args()

    suite_dir = pathlib.Path(args.suite_dir)
    global_ns = f"{args.family}-global-ns"
    app_ns = f"{args.family}-app-ns"

    changed = migrate_fixture_files(suite_dir, global_ns, app_ns, args.old_local_ns)
    for c in changed:
        print("MIGRATED", c)

    ns_file = write_namespaces_file(suite_dir, global_ns, app_ns, args.provider)
    print("WROTE", ns_file)

    ct_path = suite_dir / "chainsaw-test.yaml"
    new_text = patch_chainsaw_test(ct_path, global_ns, app_ns, args.old_local_ns, args.old_name)
    ct_path.write_text(new_text)
    print("PATCHED", ct_path, "(text substitutions only -- step insertion done separately)")


if __name__ == "__main__":
    main()
