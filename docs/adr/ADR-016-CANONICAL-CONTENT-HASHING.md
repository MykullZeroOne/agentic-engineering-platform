---
id: ADR-016
type: adr
tier: 1
status: proposed
version: 2
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-30
---

# ADR-016 — Canonical Content Hashing and Migration Attestation

This ADR **refines** ADR-013 and depends on ADR-017. It does not supersede either.

## Context

ADR-013 binds every approval to a `content_hash` of an artifact and never defined the computation.
Today it is `sha256` of raw markdown after front matter, truncated to 32 characters — a property
of one serialization, not of the artifact.

The first draft of this ADR answered that by defining a canonical form over a parsed CommonMark
block model, so that a hard-wrapped markdown document and the same content from Confluence would
hash alike. That was solving the problem as posed. ADR-017 improves the posing: **canonical
artifacts are structured data and documents are projections of it**, so there is no markdown to
canonicalize. Hashing a rendering was the mistake; the block model was an elaborate way of
committing it carefully.

## Decision

**Hash the structured artifact with canonical JSON. Normalize prose leaves as text. Re-anchor by
human attestation when a hash legitimately changes.**

1. **Canonical JSON over the base data.** Serialize the artifact per RFC 8785 — UTF-8, sorted keys,
   no insignificant whitespace — and take `sha256` over those bytes. Field order, indentation and
   YAML-versus-JSON representation are all invisible to the hash, because none of them are the
   artifact.

2. **Prose leaves are normalized text.** A prose field — an ADR's `Decision`, a requirement's
   `statement` — is normalized before serialization: Unicode NFC, `LF` line endings, whitespace
   runs collapsed to one space, leading and trailing whitespace dropped, soft line breaks treated
   as spaces. Fenced content inside a prose leaf is exempt from whitespace normalization.

   This is the plain-text approach, and it is safe *here* precisely because it is confined to a
   leaf. Applied to a whole markdown document it would erase headings, list nesting and link
   destinations — meaning disguised as formatting. Inside a structured artifact those are already
   fields, so nothing meaningful is left for normalization to destroy.

3. **Two hashes.** `content_hash` over the canonical form anchors approvals. `source_hash` over the
   raw stored bytes detects in-place edits and is diagnostic only, never grounds for invalidating
   an approval by itself.

4. **The algorithm is versioned**, `canonical/v2`, and the version prefixes every hash. Changing
   the rules is itself a migration.

5. **A change in `content_hash` is a material change** — the version increments, status returns to
   `proposed`, the gate re-opens. An editorial exemption is a human act, never an agent's.

6. **Migration attestation.** When representation changes legitimately, a human attests that
   meaning did not, naming the method used to establish it. The attestation re-anchors existing
   approvals to new hashes. It is not a re-approval, and the record keeps the two distinguishable.

7. **Full 64-character digests**, replacing the current truncation.

## Alternatives considered

**A canonical form over a parsed markdown block model.** This ADR's own first draft. Rejected once
ADR-017 landed: it requires a CommonMark parser, which makes the hash depend on a library rather
than only on a specification — two conformant implementations can disagree on edge cases, and a
parser upgrade silently becomes an algorithm change. All of that complexity exists only to hash a
rendering.

**Strip all formatting, reduce a whole document to plain text, hash that.** Rejected as the anchor,
adopted for leaves. Over a whole markdown document it fails in the dangerous direction: a heading
demoted to body text hashes identically, list nesting vanishes, and — worst for a corpus built on
cross-references — link destinations disappear, so "see ADR-002" is indistinguishable from "see
ADR-003". It remains an excellent *attestation method*, which is where it now lives.

**Re-approve every artifact on every migration.** Rejected. Safe against false stability, but mass
re-approval is how a human learns to click through without reading — the automation complacency
ADR-014 identifies as un-trainable-away. A mechanism that manufactures rubber-stamping to protect
integrity has traded one failure for a worse one.

**Carry approvals forward automatically when a diff looks equivalent.** Rejected: it puts an agent
in the position of deciding a human's approval still holds, which Principle 1 forbids.

**Hash a semantic summary or embedding.** Rejected outright. Approximate matching cannot separate
"reworded" from "reversed," which is the entire distinction an approval rests on.

## Consequences

- **No parser dependency.** Canonical JSON is a specification, implementable identically anywhere.
  This removes the sharpest risk in the previous draft.
- Approval records carry `source_hash` and an algorithm-qualified `content_hash`.
- Attestation records `ATT-NNNN` live under `.agentic/attestations/`.
- **`APR-0001` through `APR-0003` must be re-anchored by the first attestation.** They carry
  truncated digests over raw markdown bodies. The mechanism's first use remains its own
  introduction.
- Materiality becomes mechanical, so `DOCUMENT_LIFECYCLE.md`'s prose definition applies only to
  the editorial exemption.
- Hashing cannot be implemented before the base-data schema exists, so this ADR is gated on
  ADR-017's blocking deliverable.

## Risks

- **The schema decides what is hashable.** A field the schema omits is invisible to the hash. If
  something meaningful lives outside the structure — a comment, an unmodelled attribute — it can
  change without invalidating an approval. This moves the false-stability risk from the
  canonicalizer into the schema, where it is easier to reason about but no less real.
- **Prose-leaf normalization can still over-normalize.** Whitespace is not always insignificant
  inside prose; ASCII tables or aligned examples written into a prose field would be flattened.
  Fenced content is exempt, which covers the common case and not every case.
- **The interim is unhashable.** Until artifacts are structured, there is nothing to apply this to,
  and the existing truncated hashes stay in force. A long interim means the approval chain is
  anchored to a scheme this ADR has already declared wrong.
- **Attestation can become the rubber stamp it was designed to avoid.** "I attest these 200
  artifacts are equivalent" is exactly as thin as the method behind it. Requiring the method to be
  named makes the thinness visible; it does not prevent it.
