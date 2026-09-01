---
id: SPEC-EVIDENCE-PACKAGE
type: spec
tier: 0
status: approved
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-09-01
approval_record: APR-0016
supersedes: null
superseded_by: null
last_reviewed: 2026-09-01
enforcement: prose
---

# Evidence Package

`docs/spec/REGISTRIES.md` records this as a known gap: the package has a *shape* —
`.github/pull_request_template.md` — and no *schema*. Under ADR-017 an evidence package is
structured data of which the pull request body is one projection, so the template is an interim
human-facing form. QA and review cannot consume it mechanically.

CLAUDE.md merge requirement 3 demands an evidence package on every pull request. Until now
nothing said what one is.

## Origin

The convention is adapted from Kairo's `docs/v2/27-traceability/acceptance-criteria-convention.md`
(same owner, MIT), which solved the same problem for the same reason. Its four rules are kept
because they are the load-bearing part; the schema below is AEP's.

## Rules

1. **Automate first.** If a check can run in CI or a local script without human judgment, it
   must be automated. A manual proof for something automatable is a defect in the package.
2. **Manual only when necessary**, and then as Given / When / Then. UI inspection, one-time
   bootstrap review, and outcomes requiring human judgment qualify. Nothing else does.
3. **Preconditions are explicit.** Environment variables, running services, seed data and
   config files are listed, with the commands that create them, before the scenario runs.
4. **One proof per claim.** Every claim cites exactly one primary proof: a command, a test
   path, or a manual scenario id. A claim citing two proofs is two claims.

## Schema

`evidence/v1`. Four required blocks and one that is required to be present even when empty.

### Identity
`evidence_id`, `schema`, `work_item`, `head_commit`, `recorded_on`. `head_commit` is not
optional: **a proof that does not name the commit it ran on proves nothing about the head**,
and a package assembled across a rebase is the failure this field exists to catch.

### `claims`
One entry per material change. Each carries `change` (what), `satisfies` (the requirement, ADR,
ADS criterion or work item it serves), and `proof` (exactly one id, per rule 4).

`satisfies` may not be empty. "Refactor" and "cleanup" are not requirements — a change that
satisfies nothing is stated as such in `open`, with a justification, rather than given a
decorative satisfier.

### `proofs`
One entry per id. `kind` is `automated` or `manual`.

- **automated** — `command`, `expect`, `ran_on`, `result`. `expect` is the observable outcome
  (`exit 0`, a named test passing), not a description of intent.
- **manual** — `scenario`, `preconditions`, `given`, `when`, `then`, `performed_by`. A manual
  proof also carries `why_not_automated`, because rule 1 makes that the exception requiring a
  reason.

`result` is recorded, never inferred. A proof with no `result` is an unrun proof.

### `gates`
Every human gate the change triggers, from `.agentic/registries/gates.yaml`, with `state` and
`approval_record`. **Merging closes no gate (ADR-019)** — only a record does — so `state:
crossed` with `approval_record: null` is a valid and blocking entry, not an error in the
package.

### `open`
What is **not** proven. Required to be present even when empty, because an absent section and
an empty one read identically and only one of them is a claim.

This is where a suspended requirement lives. AEP's independent review is suspended, not
satisfied, and has been recorded that way on every pull request since PR #15; that is exactly
the shape this block is for.

## Relationship to the pull request

The pull request body is a **projection** of the package, per ADR-017 — the human-facing
rendering, not the artifact. `.github/pull_request_template.md` stays as the authoring surface
until a generator exists. Where the two disagree, the structured package is the artifact.

## Enforcement

`prose` today. Nothing validates a package because no package exists to validate; a checker is
the next rung and is not part of this specification. The ladder is `prose → check → hook`, and
claiming `check` before a check exists would make this document assert a stage the pipeline is
not at.

## What this does not do

- It does not define the *skill* schema, the other half of the gap SPEC-REGISTRIES records.
- It does not decide where packages are stored. `.agentic/` is the obvious home and that is a
  `platform_config` change with its own gate.
- It does not close the tier 3 question (WI-0006). This is a tier 0 spec describing a data
  shape, not a Rule about how work is done.
