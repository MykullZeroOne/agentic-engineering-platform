---
id: SPEC-AUTHORITY
type: spec
tier: 0
status: approved
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-30
approval_record: APR-0005
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Authority Model

Defines `authority_level` — the field `docs/context/CONTEXT_PROVENANCE.md` records for every
context item and `docs/context/CONTEXT_RETRIEVAL.md` ranks by first. Every artifact in this
repository declares a tier. When two artifacts conflict, this document decides which wins.

## Tiers

| Tier | Name | Artifacts | Nature |
| --- | --- | --- | --- |
| 0 | Principles | `docs/vision/PRINCIPLES.md`, this document, `docs/spec/**` | Constitutional |
| 1 | Decisions | ADR | Normative |
| 2 | Requirements | PRD, ADS, design documentation | Normative |
| 3 | Rules | Policy, Standard | Normative |
| 4 | Procedures | Skill | Normative |
| 5 | Execution state | GitHub issues, PRs, checks, releases | Factual |
| 6 | Learned | Memory | Derived |

Tiers 0–4 are the **normative stack**: they state what *ought* to be. Tier 5 is **factual**: it
states what *is*. Tier 6 is **derived**: it states what has been *observed*.

## Precedence

1. **Within the normative stack, the lower tier wins.** A principle beats an ADR; an ADR beats a
   PRD; a policy beats a skill. An artifact may constrain lower tiers but never contradict a
   higher-authority one.
2. **Only `approved` or `accepted` artifacts carry their tier's authority.** A `draft` ADR
   outranks nothing. See `docs/spec/DOCUMENT_LIFECYCLE.md`.
3. **Within a tier, supersession is the only ordering.** If artifact B declares
   `supersedes: A`, B wins. Two approved artifacts in the same tier that conflict without a
   supersession link are a **finding**, not a thing to resolve by judgment — raise it per
   `docs/workflows/QUESTION_ESCALATION.md`.
4. **Tier 5 is authoritative only for execution state.** Per ADR-001, GitHub owns whether an
   issue is open, a check passed, or a PR merged. It has no authority over what is *required* —
   a green check does not satisfy a requirement the tier-2 artifact does not list.
5. **Tier 6 never outranks tiers 0–4.** A memory that contradicts approved canonical knowledge
   is set to `challenged` (`.agentic/registries/states.yaml`, `memory_status`) and excluded from
   retrieval as authoritative. This implements `docs/memory/MEMORY_ARCHITECTURE.md`.

## Why memory sits at the bottom

Memory is evidence-backed observation, not truth. `docs/memory/CONSOLIDATION.md` defines a
promotion ladder — observation → candidate → validated pattern → procedure → standard. That
ladder is exactly a climb *up* this table: a memory gains authority only by being promoted into
tier 4 or 3, at which point the memory itself is marked `promoted` and the authority lives at the
new tier. Nothing skips a rung.

## Human gates cut across tiers

A gate is not a tier. Gates (`.agentic/registries/gates.yaml`) determine *who may change an
artifact*; tiers determine *which artifact wins*. Every artifact in tiers 0–3 is gated; tier 4
and below may be agent-authored within policy.

## Enforcement stages

Every rule-bearing artifact declares how it is currently enforced:

| Stage | Meaning |
| --- | --- |
| `prose` | An agent must read the rule and honor it. |
| `check` | A validator or CI check enforces it. |
| `hook` | An AEP lifecycle hook enforces it deterministically. |

Rules begin at `prose` and migrate rightward as the platform is built. This is deliberate:
ADR-007 requires that mandatory behavior become deterministic, and recording the current stage
turns that principle into a measurable backlog rather than a standing violation. **The set of
rules still at `prose` is the acceptance criteria for the hook engine.**
