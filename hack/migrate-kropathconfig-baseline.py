#!/usr/bin/env python3
"""Migrate a Chainsaw suite's per-AC global KropathConfig fixtures to the
singleton `baseline` name, isolated via a dedicated namespace pair per step
(KRO-1137, Option A -- docs/design/kropathconfig-singleton-chainsaw-migration.md).

For each `*setup*.yaml` file in a suite directory that contains a
`kind: KropathConfig` document:
  - renames it to `baseline` and moves it to a new global-config namespace
    `<slug>-<family>-global-ns`
  - moves every other document that lived in the OLD global namespace (e.g. a
    global-tier <ResourceFamily>Config, which resolves via the same
    global-config-namespace annotation, ADR-015 §5.6) to the new global-ns too
  - moves every document that lived in the suite's shared local namespace to
    a new `<slug>-<family>-app-ns`, and annotates that namespace with
    `<provider>.kropath.run/global-config-namespace` pointing at the new
    global namespace
  - prepends the two new `Namespace` objects (global-ns, then app-ns)
  - rewrites the matching `*assert*.yaml` file(s) for the same step

`<slug>` is derived from the setup file's AC identifier, preserving the
exact digit form used in the filename (so `ac09-setup.yaml` -> `ac09`, not
`ac9`) so the companion assert file(s) can still be found by substring glob.

This script is intentionally conservative: it SKIPS (and reports) any file
where the KropathConfig's namespace or the local-namespace set can't be
resolved unambiguously. Skipped files must be migrated by hand. Comments
and formatting are preserved via ruamel.yaml round-trip mode.

Usage:
  python3 hack/migrate-kropathconfig-baseline.py tests/acm/ctrl-acm \
      --family acm --provider aws
"""
import argparse
import pathlib
import re
import sys

from ruamel.yaml import YAML

GLOBAL_NS_CANDIDATES = {"kro-system"}

yaml = YAML()
yaml.preserve_quotes = True
yaml.indent(mapping=2, sequence=2, offset=0)
yaml.width = 4096


def ac_slug(filename: str) -> str:
    """Slug preserving the exact digit form used in the filename, so glob
    matching against companion files (e.g. an assert file) still works."""
    m = re.search(r"(ac0*[0-9]+)", filename, re.IGNORECASE)
    if m:
        return m.group(1).lower()
    m = re.match(r"([0-9]+)", filename)
    if m:
        return m.group(1)
    return pathlib.Path(filename).stem


def load_docs(path: pathlib.Path):
    with open(path) as f:
        return list(yaml.load_all(f))


def dump_docs(path: pathlib.Path, docs) -> None:
    with open(path, "w") as f:
        first = True
        for d in docs:
            if d is None:
                continue
            if not first:
                f.write("---\n")
            first = False
            yaml.dump(d, f)


def migrate_setup_file(path: pathlib.Path, family: str, provider: str):
    docs = load_docs(path)
    kpc_docs = [d for d in docs if d and d.get("kind") == "KropathConfig"]
    if not kpc_docs:
        return None
    if len(kpc_docs) > 1:
        return f"SKIP {path}: multiple KropathConfig documents in one file"

    kpc = kpc_docs[0]
    old_global_ns = kpc.get("metadata", {}).get("namespace")
    if old_global_ns not in GLOBAL_NS_CANDIDATES:
        return f"SKIP {path}: KropathConfig namespace {old_global_ns!r} is not a recognized global namespace"

    other_docs = [d for d in docs if d is not kpc]
    local_namespaces = {
        d.get("metadata", {}).get("namespace")
        for d in other_docs
        if d
        and d.get("kind") != "Namespace"
        and d.get("metadata", {}).get("namespace") != old_global_ns
    }
    local_namespaces.discard(None)
    if len(local_namespaces) > 1:
        return f"SKIP {path}: ambiguous local namespace set {local_namespaces}"
    old_local_ns = next(iter(local_namespaces), None)
    if old_local_ns is None:
        return f"SKIP {path}: no local-namespace document found alongside KropathConfig"

    slug = ac_slug(path.stem)
    app_ns = f"{slug}-{family}-app-ns"
    global_ns = f"{slug}-{family}-global-ns"
    annotation_key = f"{provider}.kropath.run/global-config-namespace"

    kpc["metadata"]["name"] = "baseline"
    kpc["metadata"]["namespace"] = global_ns
    for d in other_docs:
        ns = d.get("metadata", {}).get("namespace") if d else None
        if ns == old_local_ns:
            d["metadata"]["namespace"] = app_ns
        elif ns == old_global_ns:
            d["metadata"]["namespace"] = global_ns

    ns_global = {"apiVersion": "v1", "kind": "Namespace", "metadata": {"name": global_ns}}
    ns_app = {
        "apiVersion": "v1",
        "kind": "Namespace",
        "metadata": {"name": app_ns, "annotations": {annotation_key: global_ns}},
    }
    new_docs = [ns_global, ns_app] + docs
    dump_docs(path, new_docs)
    ns_map = {old_local_ns: app_ns, old_global_ns: global_ns}
    return {"path": path, "ns_map": ns_map, "app_ns": app_ns, "global_ns": global_ns, "slug": slug}


def migrate_assert_files(suite_dir: pathlib.Path, slug: str, setup_path: pathlib.Path, ns_map: dict):
    # Anchored so slug "ac1" cannot substring-match "ac10"/"ac11" filenames: the
    # slug must be followed by a non-digit (or end of the stem) at the match site.
    slug_re = re.compile(re.escape(slug) + r"(?!\d)")
    changed = []
    for p in sorted(suite_dir.glob("*assert*.yaml")):
        if p == setup_path or not slug_re.search(p.stem):
            continue
        docs = load_docs(p)
        touched = False
        for d in docs:
            ns = d.get("metadata", {}).get("namespace") if d else None
            if ns in ns_map:
                d["metadata"]["namespace"] = ns_map[ns]
                touched = True
        if touched:
            dump_docs(p, docs)
            changed.append(str(p))
    return changed


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("suite_dir")
    ap.add_argument("--family", required=True)
    ap.add_argument("--provider", default="aws")
    args = ap.parse_args()

    suite_dir = pathlib.Path(args.suite_dir)
    setup_files = sorted(
        p for p in suite_dir.glob("*setup*.yaml") if "assert" not in p.name
    )

    skips = []
    migrated = []
    for p in setup_files:
        result = migrate_setup_file(p, args.family, args.provider)
        if result is None:
            continue
        if isinstance(result, str):
            skips.append(result)
            continue
        changed_asserts = migrate_assert_files(
            suite_dir, result["slug"], p, result["ns_map"]
        )
        migrated.append((str(p), changed_asserts, result["app_ns"], result["global_ns"]))

    for path, asserts, app_ns, global_ns in migrated:
        print(f"MIGRATED {path} -> app_ns={app_ns} global_ns={global_ns}")
        for a in asserts:
            print(f"  also updated {a}")
    for s in skips:
        print(s, file=sys.stderr)

    if skips:
        sys.exit(1)


if __name__ == "__main__":
    main()
