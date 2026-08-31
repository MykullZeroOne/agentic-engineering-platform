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
from validate_docs import REQUIRED as FRONT_MATTER, PRE_RECORD_APPROVALS  # noqa: E402

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


# An artifact carries authority only at these statuses (AUTHORITY_MODEL precedence rule 2).
AUTHORITATIVE = {"approved", "accepted"}


def front_matter_at(rel: str, ref: str) -> dict | None:
    """The artifact's front matter at `ref`, or None if it has none that parses."""
    raw = at_ref(rel, ref)
    if raw is None:
        return None
    fm = raw
    if rel.endswith(".md"):
        if not raw.startswith("---\n"):
            return None
        end = raw.find("\n---\n", 4)
        if end == -1:
            return None
        fm = raw[4:end]
    try:
        data = yaml.safe_load(fm)
    except yaml.YAMLError:
        return None
    return data if isinstance(data, dict) else None


def status_at(rel: str, ref: str) -> str | None:
    """The artifact's `status` at `ref`.

    None is deliberately treated as authoritative by the caller: a file with no lifecycle
    -- a registry, a migration -- is always in force, so the conservative reading is that
    the gate applies.
    """
    fm = front_matter_at(rel, ref)
    return fm.get("status") if fm else None


def grandfathered(rel: str) -> bool:
    """Whether this artifact predates approval records and so can never carry one.

    PRINCIPLES and ADR-001..008 are approved, and validate_docs.py requires their
    approval_record to be null -- writing one fails the build. Under --strict they would
    otherwise be permanently unmergeable, which is enforcement turning into a wall rather
    than a gate. Reported, never blocking; ADR-016 attestation (WI-0005) is the real fix.
    """
    fm = front_matter_at(rel, "HEAD") or front_matter_at(rel, "HEAD~1")
    return bool(fm) and str(fm.get("id")) in PRE_RECORD_APPROVALS


def authority(rel: str, base: str) -> tuple[bool, bool]:
    """Whether this path carries authority at base, and at HEAD.

    A file with no readable front matter -- a registry, a migration -- counts as
    authoritative: it has no draft state, so it is always in force.
    """
    out = []
    for ref in (base, "HEAD"):
        if at_ref(rel, ref) is None:
            out.append(False)  # absent here; the other side decides
            continue
        st = status_at(rel, ref)
        out.append(st is None or st in AUTHORITATIVE)
    return out[0], out[1]


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


def file_hash(rel: str) -> str | None:
    """`legacy/file-32`: sha256 over the whole file at HEAD, for a config artifact.

    Raw bytes, matching how the digest in a record is computed. A config artifact has no
    markdown body, which is why truncated-32 cannot describe one.
    """
    proc = subprocess.run(["git", "show", f"HEAD:{rel}"], cwd=ROOT, capture_output=True)
    if proc.returncode != 0:
        return None
    return hashlib.sha256(proc.stdout).hexdigest()[:32]


def load_records() -> list[dict]:
    return [yaml.safe_load(p.read_text()) for p in sorted(APPROVALS.glob("APR-*.yaml"))]


def cover(gate_id: str, rel: str, records: list[dict]) -> tuple[str, str]:
    """Best coverage this path has for this gate: (state, detail)."""
    stale = None
    for rec in records:
        if rec.get("gate") != gate_id:
            continue
        for art in rec.get("artifacts", []) or []:
            if art.get("path") != rel:
                continue
            declared = str(art.get("content_hash", "")).rpartition(":")[2]
            # Which digest to recompute comes off the record, not off the extension:
            # a config artifact hashes whole, a document hashes its body. Guessing here
            # would let the two drift apart the moment a scheme changes.
            actual = file_hash(rel) if art.get("kind") == "config" else body_hash(rel)
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

    open_gates, unverified, grand_count, lines, unevaluated = 0, 0, 0, [], []

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

        # A gate scoped `authoritative` does not fire on drafts: the document asserts the
        # same nothing before and after, so there is no version for a human to approve.
        # The moment that does matter -- a tier-0 artifact reaching `approved` -- is the
        # gate's artifact_status trigger, and carries_authority() sees it as the head side.
        scope = gate.get("applies_when", "always")

        gate_open = False
        detail = []
        for rel in matched:
            was, now = authority(rel, args.base)

            if scope == "authoritative" and not (was or now):
                detail.append(("draft", rel, "carries no authority on either side"))
                continue

            # Granting or removing authority is the change this gate exists for, and it is
            # always front-matter-only -- `status: draft` to `approved` touches no body.
            # The unchanged-body rule below would therefore skip the single most important
            # case, and did: PR #9 approved three tier-0 specs and this check reported no
            # gate. A transition is never "unchanged", whatever the body says.
            if was != now:
                state, why = cover(gate["id"], rel, records)
                if state != "covered":
                    gate_open = True
                verb = "granted" if now else "removed"
                detail.append((state, rel, f"authority {verb}; {why}"))
                continue
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
            if state != "covered" and grandfathered(rel):
                # Approved before approval records existed and unable to carry one.
                # Reported so the gap stays visible, never blocking.
                grand_count += 1
                detail.append(("grand", rel, "predates approval records; needs ADR-016 "
                                             "attestation (WI-0005), not a record"))
                continue
            if state == "covered" and why.endswith("(currency unverifiable)"):
                unverified += 1
            if state != "covered":
                gate_open = True
            detail.append((state, rel, why))

        if not any(s not in ("unchanged", "draft") for s, _, _ in detail):
            continue

        status = "OPEN" if gate_open else "closed"
        if gate_open:
            open_gates += 1
        lines.append(f"  {gate['id']}  [{status}]  approver: {gate.get('approver')}")
        for state, rel, why in detail:
            if state in ("unchanged", "draft"):
                continue
            if state == "grand":
                lines.append(f"    GRAND {rel}  -- {why}")
                continue
            mark = {"covered": "ok   ", "stale": "STALE", "open": "OPEN "}[state]
            lines.append(f"    {mark} {rel}  -- {why}")
        skipped = sum(1 for s, _, _ in detail if s == "unchanged")
        if skipped:
            lines.append(f"    ({skipped} path(s) touched with an unchanged body, not counted)")
        drafts = sum(1 for s, _, _ in detail if s == "draft")
        if drafts:
            lines.append(f"    ({drafts} draft path(s) not counted; this gate is scoped "
                         f"`authoritative`)")
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
                  f"could not be checked, so that approval never re-opens on edit. A "
                  f"non-markdown artifact needs `kind: config` on its record entry to be "
                  f"hashed under legacy/file-32; without it there is no digest to compare.")
        if unevaluated:
            print(f"\nNot evaluable from a diff: {'; '.join(unevaluated)}")

    if open_gates:
        # An open gate is a human's to close (Principle 1). The check names it and stops;
        # it never records an approval, and a passing run is not one.
        print(f"\n{open_gates} gate(s) OPEN. A human must approve and a record must be "
              f"written to {APPROVALS.relative_to(ROOT)}/ before this merges.")
        return 1 if args.strict else 0

    if grand_count:
        # Saying "every gate is closed" while something was waved through is the kind of
        # overclaim this check exists to catch.
        print(f"\nWARNING: {grand_count} path(s) passed as grandfathered, not approved. "
              f"They predate approval records and cannot carry one, so nothing here "
              f"verifies them; ADR-016 attestation (WI-0005) is what would.")
    if lines:
        print("\nOK: every triggered gate is closed by a current approval record"
              + (", except the grandfathered path(s) above." if grand_count else "."))
    else:
        print("\nOK: this change crosses no human gate.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
