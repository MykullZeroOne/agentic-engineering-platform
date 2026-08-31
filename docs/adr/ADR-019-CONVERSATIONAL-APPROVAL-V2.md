---
id: ADR-019
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-30
approval_record: APR-0006
supersedes: ADR-013
superseded_by: null
last_reviewed: 2026-08-30
---

# ADR-019 — Conversational Approval, Restated

**Supersedes ADR-013 on acceptance.** The front-matter `supersedes` link is deliberately null
while this ADR is `proposed`: a proposed decision retires nothing, and ADR-013 remains in force
until a human accepts this. ADR-013's substance is retained; two clauses that later decisions
contradicted are corrected. Nothing about who may close a gate changes.

## Context

ADR-013 established that a human's explicit instruction in conversation closes a gate, provided an
approval record captures it. That holds. Two of its clauses no longer do.

**The hash scope contradicts ADR-016.** ADR-013 states `content_hash` covers "the document body,
excluding front matter." ADR-016 states it covers canonical JSON over the structured artifact,
excluding approval metadata. Both are accepted, and as written they are incompatible. ADR-016
described itself as a refinement; that was wrong, and this ADR is the correction. The two are not
in tension on *principle* — both say hash the content and exclude the approval envelope — but
ADR-013 expressed it in terms of a markdown document, which ADR-017 replaced with structured data
as the canonical artifact.

**`approved_by` is specified as a pointer and implemented as a name.** ADR-013 says `approved_by`
and `approved_on` "point at" the approval record. In practice `approved_by` holds an identity
(`MykullZeroOne`) across 24 artifacts, because that is what a reader wants to see. The pointer the
provenance chain needs was never added, so the link ADR-013 claims does not exist in the data.

## Decision

**ADR-013 is restated with those two clauses corrected. Everything else is carried forward
unchanged.**

Carried forward, verbatim in effect:

1. A human's explicit, unambiguous instruction in conversation is a valid gate closure, provided an
   approval record is captured.
2. The record — not the front-matter fields — is the provenance.
3. Valid surfaces are `chat`, `github_review`, and `ui`.
4. **Merging a pull request is not a gate closure.** Squash merge (ADR-009) leaves no granularity
   to approve with and names no scope.
5. An approval binds to the artifact version it names.
6. An approval names its scope; what the human did not address is not approved.
7. **Ambiguity is not approval.** An agent must not infer a closure from acquiescence, silence,
   enthusiasm, or an instruction about adjacent work.
8. An agent may never close a gate, only record a human's closure of one.

Corrected:

9. **Hash scope is a principle, not a format.** An approval binds to a digest over the artifact's
   *content*, excluding the approval envelope — the fields that carry the approval itself, which
   would otherwise make every approval self-invalidating on write. The computation is defined by
   ADR-016 and `docs/spec/CONTENT_HASHING.md`, and is versioned there. This ADR does not restate
   it, so the two cannot drift apart again.

10. **Identity and pointer are separate fields.** `approved_by` holds the authenticated identity
    that closed the gate — never a role, per ADR-013's original and retained rule.
    `approval_record` holds the record id. The record is still the provenance; `approved_by` is a
    denormalised convenience whose agreement with the record is validator-enforced.

## Alternatives considered

**Partial supersession — retire only ADR-013's two bad clauses.** Rejected. It would require a
clause-level supersession mechanism the model does not have, and an ADR whose force depends on
chasing clause links to discover what is still live is unreadable. Whole-document supersession with
restatement is the standard practice precisely because the reader can trust what is in front of
them.

**Leave both accepted and let ADR-016 win by recency.** Rejected. Nothing in the authority model
resolves same-tier conflicts by recency — `AUTHORITY_MODEL.md` rule 3 says supersession is the
*only* ordering within a tier, and that two conflicting approved artifacts without a supersession
link are a finding, not something to resolve by judgment. Doing it informally here would quietly
break that rule for everything else.

**Change `approved_by` to hold the record id.** Rejected. It makes the common case — "who approved
this?" — require a second lookup, and the field name would then mean something other than what it
says. Adding the pointer costs one field and keeps both readable.

**Supersede ADR-016 as well and merge both into one ADR.** Rejected. ADR-016 is correct under
ADR-017 and nothing in it needs restating. Superseding a correct decision to tidy the numbering
would lose its alternatives and risks section for no gain.

## Consequences

- ADR-013 gains `superseded_by: ADR-019` and is retained for provenance.
- `docs/spec/DOCUMENT_LIFECYCLE.md` must add `approval_record` to the front-matter contract, and
  the 24 artifacts carrying `approved_by` must gain the pointer. **That is a material change to
  DOCUMENT_LIFECYCLE, so accepting this ADR re-opens its `constitutional_change` gate** — it is
  approved as it stands today, not as it will stand after.
- `scripts/validate_docs.py` gains an `approval_record` agreement check alongside the existing
  `approved_by` one.
- ADR-016 stands unmodified. Its self-description as a refinement of ADR-013 is now historical:
  what it contradicted is superseded.

## Risks

- **Restatement can drift from the original.** Clauses 1-8 are carried forward by hand, so a
  transcription error would silently change a rule nobody thinks changed. Mitigated by ADR-013
  remaining readable in the corpus; not eliminated.
- **A superseded ADR is still cited.** Existing records — `APR-0002` closed a gate on ADR-013
  itself — reference it. Supersession does not rewrite history, and readers following an old
  citation land on a document that no longer governs. The `superseded_by` link is the only
  mitigation.
- **This is the second correction to the same mechanism in one day.** The approval machinery has
  now been specified, contradicted, and restated in quick succession, which is weak evidence that
  it is being designed faster than it is being understood.
