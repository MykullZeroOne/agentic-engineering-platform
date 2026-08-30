---
id: ADR-016
type: adr
tier: 1
status: proposed
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-30
---

# ADR-016 — Canonical Content Hashing and Migration Attestation

This ADR **refines** ADR-013; it does not supersede it. ADR-013 established that an approval binds
to a content hash. It left the computation undefined, and ADR-015 turned that omission into a
blocker.

## Context

ADR-013 binds every approval to a `content_hash` of the document body. It never said how to compute
one, so today the hash is `sha256` of the raw markdown after front matter, truncated to 32 hex
characters. That is a property of one serialization, not of the document.

ADR-015 makes stores pluggable, which makes the omission load-bearing. This corpus hard-wraps at
roughly 100 columns; the same document from Confluence has no hard wraps. A byte-level hash of the
two differs, so migrating a store would silently invalidate every approval on every document moved
— without anything having changed in meaning.

Reformatting inside a single store has the same effect today. The problem is not specific to
migration; migration only makes it unavoidable.

## Decision

**Approvals bind to a canonical hash computed over a document model, and a hash that legitimately
changes is re-anchored by human attestation rather than recomputed.**

1. **The canonical form is defined over a parsed block model, not over text.** Specified in
   `docs/spec/CONTENT_HASHING.md`. Hard-wrapping, list markers, emphasis characters, table padding
   and fence style are representation and are normalized away; text, heading level, list structure,
   link destinations and code content are meaning and are preserved.
2. **Two hashes.** `content_hash` over the canonical form is the approval anchor. `source_hash`
   over raw stored bytes detects in-place edits and is diagnostic only.
3. **The algorithm is versioned**, `canonical/v1`, and the version prefixes every hash.
4. **A change in `content_hash` is a material change** — version increments, status returns to
   `proposed`, the gate re-opens. An editorial exemption is a human act, never an agent's.
5. **Migration attestation.** When representation changes legitimately, a human attests that
   meaning did not, naming the method used to establish it. The attestation re-anchors existing
   approvals to new hashes. It is not a re-approval, and the record keeps the two distinguishable.
6. **Full 64-character digests.** The current 32-character truncation is replaced.

## Alternatives considered

**Keep hashing raw bytes and accept re-approval on every migration.** Rejected. It is safe against
false stability but produces mass re-approval events, and mass re-approval is how a human learns to
click through without reading — the automation-complacency failure ADR-014 already identifies as
un-trainable-away. A mechanism that manufactures rubber-stamping to protect integrity has traded
one failure for a worse one.

**Text-level normalization** — normalize line endings, trailing whitespace, blank runs, then hash.
Rejected on a concrete ground: it cannot undo hard-wrapping, and unwrapping requires knowing which
newlines are paragraph soft breaks and which are structural, inside code blocks, tables and lists.
That requires parsing. A regex normalizer would appear to work on this corpus and fail on the first
Confluence document.

**Hash a semantic summary or embedding.** Rejected outright — approximate matching cannot
distinguish "reworded" from "reversed," which is precisely the distinction an approval depends on.

**Carry approvals across migration automatically when a diff looks equivalent.** Rejected. It puts
an agent in the position of deciding that a human's approval still holds, which Principle 1
forbids. Attestation keeps the judgment with the human while keeping the cost to one act per
migration rather than one per document.

**Never migrate — pin each document to its original store forever.** Rejected: it makes ADR-015's
provider model unusable in practice and traps projects in whichever tool they started with.

## Consequences

- `scripts/validate_docs.py` and the approval-writing path must implement `canonical/v1`, which
  requires a CommonMark parser rather than string handling.
- Approval records gain `source_hash` and an algorithm-qualified `content_hash`.
- A new record type, `ATT-NNNN`, lives under `.agentic/attestations/`.
- **`APR-0001` through `APR-0003` must be re-anchored by the first attestation.** They carry
  truncated digests over raw bodies; adopting `canonical/v1` changes all of them. The mechanism's
  first use is on its own introduction.
- Materiality stops being a judgment call, so `DOCUMENT_LIFECYCLE.md`'s prose definition becomes
  the fallback for the editorial case rather than the primary rule.
- Every non-local provider must produce the canonical form, which raises the bar for provider
  eligibility set in ADR-015.

## Risks

- **The canonical form can be wrong in either direction.** A rule that normalizes away something
  meaningful causes false stability — an approval covering content nobody approved. The block model
  is therefore conservative: anything not explicitly listed as representation stays significant.
- **Parser dependence.** Two CommonMark implementations can disagree on edge cases, so the same
  document could hash differently in different tools. Mitigated by pinning one parser and treating
  a parser change as an algorithm change requiring attestation — but this is a real coupling, and
  the strictest reading is that the hash depends on a library, not only on a specification.
- **Attestation can become the rubber stamp it was designed to avoid.** "I attest these 200
  documents are equivalent" is exactly as thin as the method behind it. Requiring the method to be
  named makes the thinness visible; it does not prevent it.
- **Store round-tripping may not be stable at all** for a store whose model is poorer than the
  block model — one that cannot represent nested lists, say. Such a store fails ADR-015's
  eligibility test, but the failure will surface as hash instability rather than as a clean error.
