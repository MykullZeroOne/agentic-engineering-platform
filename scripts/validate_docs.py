#!/usr/bin/env python3
"""Validate the canonical documentation corpus.

Enforces the contracts in docs/spec/DOCUMENT_LIFECYCLE.md, docs/spec/AUTHORITY_MODEL.md,
and docs/spec/REGISTRIES.md. This is the `validation.docs` hook binding in
.agentic/hooks/hooks.yaml, running today at enforcement stage `check`.

Usage:  python3 scripts/validate_docs.py [--quiet]
Exit:   0 clean, 1 violations found.
"""
from __future__ import annotations
import pathlib, re, sys

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
REG = ROOT / ".agentic" / "registries"

# Entry points and illustrative fragments, per DOCUMENT_LIFECYCLE.md "Exemptions".
EXEMPT_FILES = {"README.md", "CLAUDE.md", "examples/project.yaml"}
EXEMPT_DIRS = {"docs/schemas", "scripts", ".github"}

REQUIRED = ["id", "type", "tier", "status", "version", "owner",
            "human_approved", "approved_by", "approved_on",
            "supersedes", "superseded_by", "last_reviewed"]

TYPE_TIER = {"principle": 0, "spec": 0, "adr": 1, "prd": 2, "ads": 2, "design": 2,
             "policy": 3, "standard": 3, "skill": 4,
             "guide": None, "schema": None, "example": None}

# `accepted` is the ADR spelling of `approved`; only ADRs may use it.
AUTHORITATIVE = {"approved", "accepted"}
ID_PATTERN = {"adr": r"^ADR-\d{3}$", "prd": r"^PRD-\d{3}$", "ads": r"^ADS-\d{3}$"}

ENFORCEMENT_STAGES = {"prose", "check", "hook"}

errors: list[str] = []
warnings: list[str] = []


def err(path, msg):
    errors.append(f"{path}: {msg}")


def load_registry(name):
    return yaml.safe_load((REG / name).read_text())


def parse_front_matter(path: pathlib.Path):
    """Return (front_matter, ok). Markdown uses --- delimiters; YAML uses top-level keys."""
    text = path.read_text()
    if path.suffix in (".yaml", ".yml"):
        data = yaml.safe_load(text)
        return (data, True) if isinstance(data, dict) else (None, False)
    if not text.startswith("---\n"):
        return None, False
    end = text.find("\n---\n", 4)
    if end == -1:
        return None, False
    return yaml.safe_load(text[4:end]), True


def is_exempt(rel: str) -> bool:
    if rel in EXEMPT_FILES:
        return True
    return any(rel.startswith(d + "/") for d in EXEMPT_DIRS)


def collect() -> list[tuple[pathlib.Path, str]]:
    out = []
    for pattern in ("docs/**/*.md", "docs/**/*.yaml", "examples/**/*.md", "examples/**/*.yaml"):
        for p in sorted(ROOT.glob(pattern)):
            rel = str(p.relative_to(ROOT))
            if not is_exempt(rel):
                out.append((p, rel))
    return out


def check_documents(docs, states):
    statuses = {v["token"] for v in states["axes"]["document_status"]["values"]}
    by_id: dict[str, str] = {}
    fms: dict[str, dict] = {}

    for path, rel in docs:
        fm, ok = parse_front_matter(path)
        if not ok or fm is None:
            err(rel, "missing or unparseable front matter")
            continue

        missing = [f for f in REQUIRED if f not in fm]
        if missing:
            err(rel, f"missing required field(s): {', '.join(missing)}")
            continue

        fid, typ, tier, status = fm["id"], fm["type"], fm["tier"], fm["status"]

        if fid in by_id:
            err(rel, f"duplicate id {fid} (also {by_id[fid]})")
        by_id[fid] = rel
        fms[fid] = fm

        if typ not in TYPE_TIER:
            err(rel, f"unknown type {typ!r}")
        elif tier != TYPE_TIER[typ]:
            err(rel, f"type {typ!r} requires tier {TYPE_TIER[typ]}, found {tier!r}")

        if status not in statuses:
            err(rel, f"status {status!r} not in states.yaml document_status")
        if status == "accepted" and typ != "adr":
            err(rel, "status 'accepted' is reserved for ADRs; use 'approved'")

        if pat := ID_PATTERN.get(typ):
            if not re.match(pat, str(fid)):
                err(rel, f"id {fid!r} does not match {pat} for type {typ!r}")
            elif not pathlib.PurePath(rel).name.startswith(str(fid)):
                err(rel, f"filename must start with its id {fid!r}")

        # An artifact carries its tier's authority only once its gate is closed.
        if status in AUTHORITATIVE:
            if not fm["human_approved"]:
                err(rel, f"status {status!r} requires human_approved: true")
            if not fm["approved_by"] or not fm["approved_on"]:
                err(rel, f"status {status!r} requires approved_by and approved_on")
        elif fm["human_approved"]:
            err(rel, f"human_approved: true is invalid for status {status!r}")

        if "enforcement" in fm and fm["enforcement"] not in ENFORCEMENT_STAGES:
            err(rel, f"enforcement {fm['enforcement']!r} not in {sorted(ENFORCEMENT_STAGES)}")

    # Supersession must be symmetric, or provenance breaks.
    for fid, fm in fms.items():
        target = fm.get("supersedes")
        if target:
            if target not in fms:
                err(by_id[fid], f"supersedes unknown id {target!r}")
            elif fms[target].get("superseded_by") != fid:
                err(by_id[fid], f"{target} does not declare superseded_by: {fid}")
            elif fms[target]["status"] != "superseded":
                err(by_id[target], f"superseded by {fid} but status is {fms[target]['status']!r}")
    return by_id


def check_index(by_id):
    """CLAUDE.md requires every document to appear in the index."""
    index = ROOT / "docs" / "DOCUMENTATION_INDEX.md"
    text = index.read_text()
    for fid, rel in sorted(by_id.items()):
        if rel == "docs/DOCUMENTATION_INDEX.md":
            continue
        if rel not in text:
            err("docs/DOCUMENTATION_INDEX.md", f"does not list {rel} ({fid})")


def check_gates(gates):
    known = {g["id"] for g in gates["gates"]}
    aliases = {a for g in gates["gates"] for a in g.get("aliases", [])}
    for name in ("project.yaml",):
        path = ROOT / ".agentic" / name
        data = yaml.safe_load(path.read_text())
        for gid in data.get("human_gates", []):
            if gid not in known:
                hint = " (is an alias; use the canonical id)" if gid in aliases else ""
                err(f".agentic/{name}", f"unknown gate id {gid!r}{hint}")
    for p in sorted(ROOT.glob("docs/ads/*.yaml")):
        data = yaml.safe_load(p.read_text())
        for gid in data.get("human_gates", []) or []:
            if gid not in known:
                err(str(p.relative_to(ROOT)), f"unknown gate id {gid!r}")


def check_hooks(points):
    valid = {p["id"] for p in points["hook_points"]}
    domains = set(points["hook_id_domains"])
    hooks = yaml.safe_load((ROOT / ".agentic" / "hooks" / "hooks.yaml").read_text())
    rel = ".agentic/hooks/hooks.yaml"

    bound = set(hooks["bindings"])
    unbound = set(hooks.get("unbound", []))
    for pt in bound:
        if pt not in valid:
            err(rel, f"binding to undefined hook point {pt!r}")
    for pt in unbound:
        if pt not in valid:
            err(rel, f"unbound lists undefined hook point {pt!r}")
    if overlap := bound & unbound:
        err(rel, f"hook point(s) both bound and unbound: {sorted(overlap)}")
    if missing := valid - bound - unbound:
        err(rel, f"hook point(s) neither bound nor listed unbound: {sorted(missing)}")

    for point, classes in hooks["bindings"].items():
        for cls, entries in classes.items():
            if cls not in ("hard", "soft"):
                err(rel, f"{point}: unknown hook class {cls!r}")
            for entry in entries:
                hid = entry["id"]
                if "." not in hid or hid.split(".", 1)[0] not in domains:
                    err(rel, f"hook id {hid!r} is outside the closed domain namespace")


def main() -> int:
    quiet = "--quiet" in sys.argv
    states = load_registry("states.yaml")
    gates = load_registry("gates.yaml")
    points = load_registry("hook-points.yaml")

    docs = collect()
    by_id = check_documents(docs, states)
    check_index(by_id)
    check_gates(gates)
    check_hooks(points)

    for w in warnings:
        print(f"warning: {w}")
    for e in errors:
        print(f"error: {e}")

    if errors:
        print(f"\n{len(errors)} violation(s) across {len(docs)} documents.")
        return 1
    if not quiet:
        print(f"OK: {len(docs)} documents, {len(gates['gates'])} gates, "
              f"{len(points['hook_points'])} hook points, "
              f"{len(states['axes'])} state axes.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
