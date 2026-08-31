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
            "human_approved", "approved_by", "approved_on", "approval_record",
            "supersedes", "superseded_by", "last_reviewed"]

def _vocab(name):
    """Terms of one controlled vocabulary, keyed by token."""
    v = yaml.safe_load((REG / "vocabularies.yaml").read_text())["vocabularies"]
    return {term["token"]: term for term in v[name]["terms"]}


# Derived from .agentic/registries/vocabularies.yaml rather than hardcoded, so the rule
# lives in the registry and the validator reads it. Previously each of these was a Python
# constant that documentation restated -- two copies, one of them authoritative by accident.
TYPE_TIER = {tok: term["tier"] for tok, term in _vocab("artifact_type").items()}

# `accepted` is the ADR spelling of `approved`; only ADRs may use it.
AUTHORITATIVE = {"approved", "accepted"}
# A retired artifact was legitimately approved once. Supersession ends its force; it does
# not rewrite the fact that a human approved it, so these statuses keep their approval fields.
RETIRED = {"superseded", "deprecated"}
ID_PATTERN = {"adr": r"^ADR-\d{3}$", "prd": r"^PRD-\d{3}$", "ads": r"^ADS-\d{3}$"}

ENFORCEMENT_STAGES = set(_vocab("enforcement_stage"))

# Hash schemes and their digest lengths, per docs/spec/CONTENT_HASHING.md. A bare
# "sha256:" prefix is the pre-ADR-016 form: accepted during the interim but warned on,
# so existing records apply visible pressure toward the migration attestation rather
# than sitting silently non-conformant.
HASH_SCHEMES = {tok: term["digest_length"] for tok, term in _vocab("hash_scheme").items()}
LEGACY_BARE_PREFIX = "sha256:"

# Local work store, per docs/spec/LOCAL_WORK_STORE.md. Types are CLAUDE.md's branch types,
# so a work item's type and its branch prefix are one token rather than two vocabularies
# that drift. work_state is not enumerated here: states.yaml is its sole source.
WORK_REQUIRED = ["id", "store", "project", "type", "work_state", "priority",
                 "title", "description"]
WORK_TYPES = set(_vocab("work_item_type"))
WORK_PRIORITIES = set(_vocab("priority"))

# Required ADR sections, per docs/spec/DOCUMENT_LIFECYCLE.md. ADR-001..008 predate the
# requirement and are accepted and immutable, so they are grandfathered rather than rewritten.
ADR_SECTIONS = ["## Context", "## Decision", "## Alternatives considered",
                "## Consequences", "## Risks"]
ADR_STRUCTURE_FROM = 9

# Approved before ADR-013 established approval records. Listed so the absence is visible
# rather than silently tolerated; they are reported as warnings, not errors.
PRE_RECORD_APPROVALS = {"PRINCIPLES", "ADR-001", "ADR-002", "ADR-003", "ADR-004",
                        "ADR-005", "ADR-006", "ADR-007", "ADR-008"}
APPROVAL_FIELDS = ["id", "gate", "surface", "approver", "approved_on",
                   "artifacts", "request", "statement"]

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
        elif fm["human_approved"] and status not in RETIRED:
            err(rel, f"human_approved: true is invalid for status {status!r}")

        if typ == "adr" and re.match(r"^ADR-\d{3}$", str(fid)):
            if int(str(fid)[4:]) >= ADR_STRUCTURE_FROM:
                text = path.read_text()
                absent = [s for s in ADR_SECTIONS if s not in text]
                if absent:
                    err(rel, f"ADR missing required section(s): {', '.join(absent)}")

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
    return by_id, fms


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


def check_approvals(gates, by_id, fms):
    """Approval records are the provenance behind every closed gate (ADR-013)."""
    known_gates = {g["id"] for g in gates["gates"]}
    approved_ids: set[str] = set()

    for path in sorted((ROOT / ".agentic" / "approvals").glob("APR-*.yaml")):
        rel = str(path.relative_to(ROOT))
        rec = yaml.safe_load(path.read_text())

        for field in APPROVAL_FIELDS:
            if field not in rec:
                err(rel, f"approval record missing required field {field!r}")
        if not re.match(r"^APR-\d{4}$", str(rec.get("id", ""))):
            err(rel, f"approval id {rec.get('id')!r} does not match APR-NNNN")
        if rec.get("gate") not in known_gates:
            err(rel, f"unknown gate id {rec.get('gate')!r}")
        if rec.get("surface") not in set(_vocab("approval_surface")):
            err(rel, f"surface {rec.get('surface')!r} is not a valid closure surface")

        for art in rec.get("artifacts", []) or []:
            aid = art.get("id")
            if aid not in by_id:
                err(rel, f"approves unknown artifact {aid!r}")
                continue
            h = str(art.get("content_hash", ""))
            if h.startswith(LEGACY_BARE_PREFIX):
                # Pre-ADR-016 form: no scheme prefix, so the computation is ambiguous.
                warnings.append(f"{rel}: {aid} carries a bare {LEGACY_BARE_PREFIX} hash; "
                                f"needs re-anchoring by attestation (ADR-016)")
            else:
                scheme, sep, digest = h.partition(":sha256:")
                if not sep or scheme not in HASH_SCHEMES:
                    err(rel, f"artifact {aid} hash {h!r} names no known scheme "
                             f"(expected one of {sorted(HASH_SCHEMES)})")
                elif len(digest) != HASH_SCHEMES[scheme]:
                    err(rel, f"artifact {aid} scheme {scheme!r} requires a "
                             f"{HASH_SCHEMES[scheme]}-character digest, got {len(digest)}")
            # A record naming an artifact that has since left `approved` is history:
            # it approved an earlier version, and a material change re-opened the gate
            # (ADR-016 clause 5). Surface it, but do not enforce agreement against it.
            if fms[aid]["status"] not in AUTHORITATIVE:
                warnings.append(f"{rel}: approves {aid}, whose approval has since lapsed "
                                f"(now {fms[aid]['status']!r})")
                continue

            # ADR-019: approved_by is the identity, approval_record the pointer. Both
            # must agree with the record, or the provenance chain is decorative.
            claimed = fms[aid].get("approved_by")
            if claimed != rec.get("approver"):
                err(by_id[aid], f"approved_by {claimed!r} does not match "
                                f"{rec['id']} approver {rec.get('approver')!r}")
            pointer = fms[aid].get("approval_record")
            if pointer != rec["id"]:
                err(by_id[aid], f"approval_record {pointer!r} does not name "
                                f"{rec['id']}, which approves this artifact")
            approved_ids.add(aid)

    # An artifact carrying authority should be able to name the approval that granted it.
    for fid, fm in fms.items():
        if fm["status"] not in AUTHORITATIVE:
            # A retired artifact keeps its provenance for the same reason it keeps
            # human_approved: supersession ends its force, not the fact of its approval.
            if fm.get("approval_record") and fm["status"] not in RETIRED:
                err(by_id[fid], f"approval_record set on a {fm['status']!r} artifact")
            continue
        if fid in PRE_RECORD_APPROVALS:
            if fm.get("approval_record"):
                err(by_id[fid], "predates approval records; approval_record must be null")
            continue
        if fid not in approved_ids:
            warnings.append(f"{by_id[fid]}: {fm['status']} with no approval record (ADR-019)")
        elif not fm.get("approval_record"):
            err(by_id[fid], "approved artifact does not name its approval_record (ADR-019)")


def check_capabilities(gates):
    """Capability tokens are closed, and gated_by must name a real gate (SPEC-CAPABILITIES)."""
    caps = _vocab("capability")
    known_gates = {g["id"] for g in gates["gates"]}
    rel = ".agentic/registries/vocabularies.yaml"

    for token, term in caps.items():
        if "." not in token:
            err(rel, f"capability {token!r} is not <domain>.<action>")
        gate = term.get("gated_by")
        if gate and gate not in known_gates:
            err(rel, f"capability {token!r} is gated_by unknown gate {gate!r}")

    # Every grant in a role definition must name a capability that exists. A wildcard is
    # expanded here the same way a grant would be: against the registry as it stands.
    for path in sorted(ROOT.glob("docs/schemas/*.yaml")):
        data = yaml.safe_load(path.read_text())
        if not isinstance(data, dict) or "tools" not in data:
            continue
        srel = str(path.relative_to(ROOT))
        for kind in ("allow", "deny"):
            for grant in (data["tools"].get(kind) or []):
                if grant.endswith(".*"):
                    domain = grant[:-2]
                    if not any(c.startswith(domain + ".") for c in caps):
                        err(srel, f"{kind} grant {grant!r} expands to no capability")
                elif grant not in caps:
                    err(srel, f"{kind} grant {grant!r} is not a known capability")


def check_work_items(states, gates):
    """Validate the local work store (docs/spec/LOCAL_WORK_STORE.md)."""
    work_dir = ROOT / ".agentic" / "work"
    if not work_dir.is_dir():
        return

    valid_states = {v["token"] for v in states["axes"]["work_state"]["values"]}
    known_gates = {g["id"] for g in gates["gates"]}
    items: dict[str, dict] = {}

    for path in sorted(work_dir.glob("WI-*.yaml")):
        rel = str(path.relative_to(ROOT))
        item = yaml.safe_load(path.read_text())
        if not isinstance(item, dict):
            err(rel, "work item is not a mapping")
            continue

        for field in WORK_REQUIRED:
            if field not in item:
                err(rel, f"work item missing required field {field!r}")
        wid = str(item.get("id", ""))
        if not re.match(r"^WI-\d{4}$", wid):
            err(rel, f"id {wid!r} does not match WI-NNNN")
        elif path.stem != wid:
            err(rel, f"filename must match its id {wid!r}")
        if wid in items:
            err(rel, f"duplicate work item id {wid}")
        items[wid] = item

        if item.get("type") not in WORK_TYPES:
            err(rel, f"type {item.get('type')!r} not in {sorted(WORK_TYPES)}")
        if item.get("priority") not in WORK_PRIORITIES:
            err(rel, f"priority {item.get('priority')!r} not in {sorted(WORK_PRIORITIES)}")
        if item.get("work_state") not in valid_states:
            err(rel, f"work_state {item.get('work_state')!r} not in states.yaml work_state axis")
        for gid in item.get("required_gates", []) or []:
            if gid not in known_gates:
                err(rel, f"unknown gate id {gid!r}")

    # Links must resolve, or the dependency graph is decorative.
    for wid, item in items.items():
        rel = f".agentic/work/{wid}.yaml"
        for field in ("dependencies", "parent"):
            refs = item.get(field) or []
            if isinstance(refs, str):
                refs = [refs]
            for ref in refs:
                if ref not in items:
                    err(rel, f"{field} references unknown work item {ref!r}")
                elif ref == wid:
                    err(rel, f"{field} references itself")


def main() -> int:
    quiet = "--quiet" in sys.argv
    states = load_registry("states.yaml")
    gates = load_registry("gates.yaml")
    points = load_registry("hook-points.yaml")

    docs = collect()
    by_id, fms = check_documents(docs, states)
    check_index(by_id)
    check_gates(gates)
    check_hooks(points)
    check_approvals(gates, by_id, fms)
    check_work_items(states, gates)
    check_capabilities(gates)

    for w in warnings:
        print(f"warning: {w}")
    for e in errors:
        print(f"error: {e}")

    if errors:
        print(f"\n{len(errors)} violation(s) across {len(docs)} documents.")
        return 1
    if not quiet:
        n_work = len(list((ROOT / ".agentic" / "work").glob("WI-*.yaml")))
        print(f"OK: {len(docs)} documents, {len(gates['gates'])} gates, "
              f"{len(points['hook_points'])} hook points, "
              f"{len(states['axes'])} state axes, {n_work} work items.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
