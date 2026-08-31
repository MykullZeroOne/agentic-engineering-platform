<!--
  This template IS the evidence package required by CLAUDE.md merge requirement #3.
  A pull request is an evidence artifact, not just a diff (END_TO_END_SDLC, phases 5-6;
  ENGINEERING_LOOP, step 7).

  Delete sections that genuinely do not apply, and say why rather than leaving them blank.
  An empty section reads as "nothing to report"; a deleted one with a reason reads as a
  decision.
-->

## What changed

<!-- The change itself, in a few sentences. Not a file list — the diff already is one. -->

## Why

<!-- What forced this now, and what breaks without it. -->

## Traceability

<!--
  Map each material change to the thing it satisfies. Use requirement IDs (REQ-NNN-NNN),
  ADR IDs, or an ADS acceptance criterion. "Refactor" and "cleanup" are not requirements —
  if a change satisfies nothing, say so and justify it.
-->

| Change | Satisfies | Verified by |
| --- | --- | --- |
|  |  |  |

## Verification

<!--
  How you know it works. Paste the command and its actual output, not a claim that it
  passed. For a fix, show the failure reproduced first and then the same check passing.
-->

```
$ python3 scripts/validate_docs.py
```

## Human gates

<!--
  Gate IDs resolve against .agentic/registries/gates.yaml. Merging does NOT close a gate
  (ADR-013) — an approval record under .agentic/approvals/ does. List every gate this
  change crosses and the record that closed it, or "not closed" if it is still open.
-->

| Gate | Approval record | Status |
| --- | --- | --- |
|  |  |  |

## Risks and what could go wrong

<!-- What you are least sure of. A PR with no stated risk is usually an unexamined one. -->

---

### Merge requirements

Per `CLAUDE.md`. A PR may merge only when all of these hold.

- [ ] **1. Names its work item and closes it.** Work items live in `.agentic/work/`
      (`docs/spec/LOCAL_WORK_STORE.md`). Put `Closes WI-NNNN` in the squash commit body.
- [ ] **2. All required checks pass.** Never merge on a red or skipped required check, and
      never weaken a check to make a PR mergeable.
- [ ] **3. Carries its evidence package.** The sections above.
- [ ] **4. Independent review.** The agent that wrote the change never approves it.
- [ ] **5. Every human gate has explicit approval**, recorded under `.agentic/approvals/`.
      An agent must never merge a PR whose gates are unclosed.
- [ ] **6. Up to date with `main`** — rebased, not merged (ADR-009).
