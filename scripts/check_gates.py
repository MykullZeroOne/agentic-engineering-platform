#!/usr/bin/env python3
"""Evaluate which human gates a change triggers, and whether each one is closed.

Merge requirement 5 in CLAUDE.md -- "an agent must never merge a PR that crosses a
human gate" -- was prose with nothing behind it, and PR #4 and #5 both merged with
open gates to prove it. This is the `check` stage of that rule: it reads the path
triggers in .agentic/registries/gates.yaml, matches them against a diff, and demands
an approval record covering every changed path the gate claims.

Merging is not a closure surface (ADR-019), so nothing else in the pipeline can
notice an unapproved gated change. This check is it.

A record covers a path when it names the same gate, names that path, and -- where the
hash scheme is computable -- still matches the file's current content. That last clause
is what makes an approval bind to a version rather than to a filename: edit the file
after approval and the gate re-opens on its own.

Usage:  python3 scripts/check_gates.py [--base REF] [--strict] [--quiet]
Exit:   0 clean, or advisory; 1 when --strict and a gate is open.
"""
from __future__ import annotations
import argparse, hashlib, pathlib, re, subprocess, sys

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
REG = ROOT / ".agentic" / "registries"
APPROVALS = ROOT / ".agentic" / "approvals"

# The front-matter contract, imported rather than restated: two copies of this list would
# drift the moment DOCUMENT_LIFECYCLE.md gains a field, and one of them would be wrong.
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from validate_docs import REQUIRED as FRONT_MATTER  # noqa: E402

# Only `path` triggers are evaluable from a diff. The others -- classification, finding,
# lifecycle, capability -- need runtime signals that do not exist yet, so the gates
# carrying them stay at enforcement stage `prose` and are reported as unevaluated rather
# than silently passed.
EVALUABLE_KIND = "path"


def glob_to_re(pattern: str) -> re.Pattern:
    """Translate a registry path glob. `**` crosses separators, `*` does not."""
    out, i = [], 0
    while i < len(pattern):
        if pattern.startswith("**", i):
            out.append(".*")
            i += 2
        elif pattern[i] == "*":
            out.append("[^/]*")
            i += 1
        else:
            out.append(re.escape(pattern[i]))
            i += 1
    return re.compile("^" + "".join(out) + "$")


def changed_paths(base: str) -> list[str]:
    """Files this branch changes relative to the merge base with `base`."""
    proc = subprocess.run(
        ["git", "diff", "--name-only", f"{base}...HEAD"],
        cwd=ROOT, capture_output=True, text=True,
    )
    if proc.returncode != 0:
        sys.exit(f"error: cannot diff against {base!r}: {proc.stderr.strip()}")
    return [p for p in proc.stdout.splitlines() if p]


def hash_text(rel: str, raw: str | None) -> str | None:
    """`legacy/truncated-32`: sha256 over the markdown body, front matter excluded.

    Returns None when the scheme does not define a body for this file -- registry and
    role YAML among them. Those paths still require a record; only its currency is
    unverifiable, which is one of the concrete costs of `canonical/v2` not existing yet.
    """
    if raw is None or not rel.endswith(".md"):
        return None
    end = raw.find("\n---\n", 4)
    if not raw.startswith("---\n") or end == -1:
        return None
    return hashlib.sha256(raw[end + 5:].encode()).hexdigest()[:32]


def body_hash(rel: str) -> str | None:
    """Hash of the committed version at HEAD.

    HEAD, not the working tree: changed_paths() diffs commits, so reading the worktree
    here would mix two snapshots and let an uncommitted edit change the verdict about a
    commit that does not contain it.
    """
    return hash_text(rel, at_ref(rel, "HEAD"))


def at_ref(rel: str, ref: str) -> str | None:
    """File content at `ref`, or None if it does not exist there."""
    proc = subprocess.run(["git", "show", f"{ref}:{rel}"],
                          cwd=ROOT, capture_output=True, text=True)
    return proc.stdout if proc.returncode == 0 else None


def content_only(rel: str, raw: str | None):
    """Everything but the front matter -- the part an approval is actually about.

    Distinct from hash_text: that computes `legacy/truncated-32` for comparison against a
    record, and is defined only over a markdown body. This answers the different question
    "did the diff touch gated content", which a YAML artifact needs too, since its front
    matter is structural keys rather than a delimited block. It is never written into a
    record, so it asserts no hash scheme.
    """
    if raw is None:
        return None
    if rel.endswith(".md"):
        end = raw.find("\n---\n", 4)
        return raw[end + 5:] if raw.startswith("---\n") and end != -1 else raw
    if rel.endswith((".yaml", ".yml")):
        # Textual, not parsed. Parsing would drop comments, and in the registries the
        # comments carry the rationale a human is approving -- a comment-only edit to
        # gates.yaml is a real change to it. Only top-level front-matter keys are
        # removed; every one of them holds a scalar, so a line filter is sufficient.
        drop = tuple(f"{k}:" for k in FRONT_MATTER)
        return "\n".join(ln for ln in raw.splitlines() if not ln.startswith(drop))
    return raw


def load_records() -> list[dict]:
    return [yaml.safe_load(p.read_text()) for p in sorted(APPROVALS.glob("APR-*.yaml"))]


def cover(gate_id: str, rel: str, records: list[dict]) -> tuple[str, str]:
    """Best coverage this path has for this gate: (state, detail)."""
    actual = body_hash(rel)
    stale = None
    for rec in records:
        if rec.get("gate") != gate_id:
            continue
        for art in rec.get("artifacts", []) or []:
            if art.get("path") != rel:
                continue
            declared = str(art.get("content_hash", "")).rpartition(":")[2]
            if actual is None:
                return "covered", f"{rec['id']} (currency unverifiable)"
            if declared == actual:
                return "covered", rec["id"]
            stale = f"{rec['id']} approved an earlier version"
    return ("stale", stale) if stale else ("open", "no approval record names this path")


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--base", default="origin/main", help="ref to diff against")
    ap.add_argument("--strict", action="store_true", help="exit 1 when a gate is open")
    ap.add_argument("--quiet", action="store_true")
    args = ap.parse_args()

    gates = yaml.safe_load((REG / "gates.yaml").read_text())["gates"]
    records = load_records()
    changed = changed_paths(args.base)
    if not changed:
        print(f"OK: no changes against {args.base}.")
        return 0

    open_gates, unverified, lines, unevaluated = 0, 0, [], []

    for gate in gates:
        triggers = gate.get("triggers", []) or []
        globs = [t["condition"] for t in triggers if t.get("kind") == EVALUABLE_KIND]
        others = sorted({t.get("kind") for t in triggers if t.get("kind") != EVALUABLE_KIND})
        if not globs:
            if others:
                unevaluated.append(f"{gate['id']} (triggers: {', '.join(others)})")
            continue

        matched = sorted({p for p in changed
                          for g in globs if glob_to_re(g).match(p)})
        if not matched:
            continue

        gate_open = False
        detail = []
        for rel in matched:
            # An approval binds to content, not to a filename (ADR-016). If the content
            # is identical on both sides, this diff moved metadata only: whatever approval
            # state the path had, it still has, so the gate does not re-open.
            head = at_ref(rel, "HEAD")
            before = content_only(rel, at_ref(rel, args.base))
            if before is not None and head is None:
                # Removing a gated artifact is a change to it. No record can cover a
                # file that is gone, and the currency check below would wave it through.
                gate_open = True
                detail.append(("open", rel, "deleted; removal of a gated artifact"))
                continue
            if before is not None and before == content_only(rel, head):
                detail.append(("unchanged", rel, "body unchanged; front matter only"))
                continue
            state, why = cover(gate["id"], rel, records)
            if state == "covered" and why.endswith("(currency unverifiable)"):
                unverified += 1
            if state != "covered":
                gate_open = True
            detail.append((state, rel, why))

        if not any(s != "unchanged" for s, _, _ in detail):
            continue

        status = "OPEN" if gate_open else "closed"
        if gate_open:
            open_gates += 1
        lines.append(f"  {gate['id']}  [{status}]  approver: {gate.get('approver')}")
        for state, rel, why in detail:
            if state == "unchanged":
                continue
            mark = {"covered": "ok   ", "stale": "STALE", "open": "OPEN "}[state]
            lines.append(f"    {mark} {rel}  -- {why}")
        skipped = sum(1 for s, _, _ in detail if s == "unchanged")
        if skipped:
            lines.append(f"    ({skipped} path(s) touched with an unchanged body, not counted)")
        if others:
            lines.append(f"    note: {', '.join(others)} triggers on this gate are not evaluated")
        lines.append("")

    if not args.quiet:
        if lines:
            print(f"Gates triggered by {len(changed)} changed file(s) against {args.base}:\n")
            print("\n".join(lines))
        else:
            print(f"No path-triggered gate matches the {len(changed)} changed file(s).")
        if unverified:
            # Named the loudest way available: a record accepted here can never go stale,
            # because nothing can tell whether the file still matches what was approved.
            print(f"\nWARNING: {unverified} path(s) accepted on a record whose currency "
                  f"could not be checked. legacy/truncated-32 is defined over a markdown "
                  f"body, so a YAML artifact's approval never re-opens on edit. "
                  f"canonical/v2 (ADR-016) is what closes this.")
        if unevaluated:
            print(f"\nNot evaluable from a diff: {'; '.join(unevaluated)}")

    if open_gates:
        # An open gate is a human's to close (Principle 1). The check names it and stops;
        # it never records an approval, and a passing run is not one.
        print(f"\n{open_gates} gate(s) OPEN. A human must approve and a record must be "
              f"written to {APPROVALS.relative_to(ROOT)}/ before this merges.")
        return 1 if args.strict else 0

    if lines:
        print("\nOK: every triggered gate is closed by a current approval record.")
    else:
        print("\nOK: this change crosses no human gate.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
