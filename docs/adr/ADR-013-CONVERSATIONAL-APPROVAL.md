---
id: ADR-013
type: adr
tier: 1
status: superseded
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
supersedes: null
superseded_by: ADR-019
last_reviewed: 2026-08-29
---

# ADR-013 — Conversational Approval Closes Human Gates

## Context

Principle 1 makes the human the sole closer of every gate in `.agentic/registries/gates.yaml`, and
`docs/security/GOVERNANCE.md` requires material decisions and human interventions to be durable
events with provenance. Neither says *how* a human closes a gate.

In practice the human interacts with agents primarily through conversation. Requiring a GitHub
review or a UI form for every approval would either block on interfaces that do not exist yet, or
push the human into a ceremony that does not match how the work actually happens. Meanwhile the
current state is worse than either: front matter asserts `approved_by` and `approved_on` with
nothing linking the assertion to the evidence that produced it, which is an approval that cannot be
audited. This was recorded as a known gap; this ADR closes it.

## Decision

**A human's explicit, unambiguous instruction in a conversation with an agent is a valid gate
closure, provided an approval record is captured.**

The approval record — not the front-matter fields — is the provenance. `approved_by` and
`approved_on` become a pointer to it.

### Approval record

| Field | Purpose |
| --- | --- |
| `id` | `APR-NNNN`, sequential, never reused |
| `gate` | Gate ID from `.agentic/registries/gates.yaml` |
| `artifacts` | IDs approved, each with the `content_hash` at approval time |
| `approver` | The specific authenticated identity — never "the human" |
| `approved_on` | Timestamp of the approving statement |
| `surface` | `chat`, `github_review`, or `ui` |
| `session` | Session or run identifier the approval occurred in |
| `request` | What the agent presented — the summary the human was responding to |
| `statement` | Verbatim quote of the approving words |
| `scope_note` | Anything the approval explicitly did *not* cover |

`content_hash` covers the document **body, excluding front matter**. Front matter carries the
approval itself, so hashing it would make every approval self-invalidating the moment it was
recorded.

Three rules give the record its force:

1. **An approval binds to a version, not a name.** The record hashes the artifacts as they stood.
   A material change re-opens the gate (`docs/spec/DOCUMENT_LIFECYCLE.md`), and the prior record
   does not carry forward.
2. **An approval names its scope.** What the human did not address is not approved. `scope_note`
   records the boundary explicitly rather than leaving it to inference.
3. **Ambiguity is not approval.** An agent must not infer a closure from acquiescence, silence,
   enthusiasm, or an instruction about adjacent work. If the human's statement does not identify
   what is being approved, the agent asks. Recording `request` alongside `statement` makes an
   over-read visible to the human afterward rather than discoverable only in consequences.

### Merging is not approval

A merge does not close a gate. Squash merge (ADR-009) collapses a pull request into a single
commit, so a merge carries no granularity — a pull request holding five ADRs cannot express
acceptance of four of them — and it names no scope, which the scope rule requires. A pull request
may merge with `proposed` artifacts in it; they carry no authority until a record closes their
gate.

A pull request *review* is different, and remains a valid surface: it can name what it approves,
and a record is written from it.

## Alternatives considered

**GitHub PR review as the only valid closure.** Rejected as *sole* mechanism, retained as a valid
surface. It produces an excellent record, but many gates do not correspond to a diff — accepting a
privacy risk, answering an escalated legal finding, approving product intent before any code
exists. Forcing those through a PR would either delay them or manufacture empty PRs to hold them.

**A dedicated approvals UI.** Rejected as a precondition, planned as a surface. PRD-005 will
provide one, but requiring it means no gate can close until Phase 1 ships. The record schema is
deliberately surface-independent so the UI becomes another writer of the same record.

**Treat conversational approval as advisory, pending later ratification.** Rejected. It creates a
window in which work proceeds on an unratified approval — exactly the ambiguity gates exist to
eliminate — and in practice the ratification step would be skipped.

## Consequences

- Approval records are immutable events in the sense of ADR-004, and are the "durable events with
  provenance" `GOVERNANCE.md` requires. Until the event store exists they are files under
  `.agentic/approvals/`.
- An agent may set `approved_by` and `approved_on` **only** when writing a corresponding record.
  Setting them without one remains a Principle 1 violation, not a shortcut.
- The human can audit any approval by reading what was presented and what they said, without
  reconstructing a conversation from memory.
- Approvals become reviewable as a set: which gates closed, on what, by whom, on which version.

## Risks

- **An agent over-reads a casual remark as approval.** The principal risk of this decision.
  Mitigated by the ambiguity rule, by requiring the agent's request to have been explicit, and by
  recording both sides so the human can see what was taken as consent.
- **A conversational record is easier to fabricate than a signed review.** Mitigated by capturing
  the verbatim statement, session identifier, and artifact hashes, and by the record being visible
  to the human. This is weaker than a cryptographic signature and is accepted deliberately;
  strengthening it is a candidate for a superseding ADR once identity infrastructure exists.
- **Session transcripts may not be durably retrievable.** Mitigated by capturing the quote *in* the
  record rather than referencing an external transcript.
