<!--
  This IS the evidence package required by CLAUDE.md merge requirement 3.

  It was cut down on 2026-09-01. The previous version asked for seven sections and a
  six-box checklist on every change, including one box that was false on every pull
  request because independent review is suspended repository-wide. That produced long
  bodies nobody consumed and taught people to skim the list.

  Keep it proportionate: a one-line fix does not need the same body as an architecture
  decision. Delete a heading that does not apply rather than filling it.
-->

## What and why

<!-- The change and what forced it. Two or three sentences for ordinary work. -->

## Verification

<!--
  The command and its actual output, not a claim that it passed. For a fix, show the
  failure first and then the same check passing.
-->

```
```

## Gates

<!--
  Only if the change crosses one. `check_gates.py --strict` decides; paste its verdict.
  Merging closes no gate (ADR-019) -- an approval record does. Delete this heading when
  nothing is crossed.
-->

## Risks

<!-- What you are least sure of, and anything you did not do. One or two lines. -->

---

<!--
  Expand into the full package -- traceability table mapping each change to what it
  satisfies, per-proof verification, an explicit gate table -- when the change is an ADR,
  a tier-0 spec, or anything crossing a gate. `docs/spec/EVIDENCE_PACKAGE.md` defines the
  structure; this body is a projection of it.

  Merge requirements live in CLAUDE.md and are not restated here. Independent review is
  suspended repository-wide (WI-0016), which is why it is no longer a per-PR box.
-->
