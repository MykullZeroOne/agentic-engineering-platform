---
id: SPEC-CONTENT-HASHING
type: spec
tier: 0
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-30
enforcement: prose
---

# Canonical Content Hashing

ADR-013 binds an approval to a `content_hash` of the document body but never defined how that hash
is computed. Undefined, the hash is a property of *one serialization*, not of the document — so a
reflow, a line-ending change, or a migration between stores (ADR-015) silently invalidates every
approval on a document whose meaning never changed.

This document defines the computation. Implements ADR-016.

## What the hash must do

Two failure modes, in opposite directions:

| Failure | Cause | Cost |
| --- | --- | --- |
| **False invalidation** | Representation changed, meaning did not | Approvals lost for no reason; re-approval fatigue trains the human to rubber-stamp |
| **False stability** | Meaning changed, hash did not | The approval covers content nobody approved |

**False stability is the dangerous one.** Where the rules below are ambiguous, resolve toward
invalidation.

## Text normalization is insufficient

This corpus is hard-wrapped at roughly 100 columns. The same document retrieved from Confluence has
no hard wraps at all — one paragraph is one line. No byte-level or line-level normalization
reconciles those two without *unwrapping* paragraphs.

Unwrapping is only safe if you know which newlines are paragraph soft breaks and which are
structural: a newline inside a fenced code block is content, a newline between table rows is
structure, a newline inside a list item is neither. Distinguishing them requires parsing.

**The canonical form is therefore defined over a document model, not over text.** A regex-based
normalizer cannot implement this correctly and must not be used.

## The document model

Parse the body into a block sequence. Blocks and their significant properties:

| Block | Significant | Ignored |
| --- | --- | --- |
| Heading | level, inline content | underline vs `#` style, trailing `#` |
| Paragraph | inline content | line wrapping, indentation |
| Code block | language, content **verbatim** | fence character, fence length, indented vs fenced |
| List | ordered flag, start number, item sequence, nesting | bullet character, marker style, loose/tight spacing |
| Table | cell grid, inline content per cell | column alignment, padding, separator width |
| Block quote | nested block sequence | marker spacing |
| Thematic break | presence | character used |
| HTML block | content verbatim | — |

Inline content, normalized in turn:

| Inline | Significant | Ignored |
| --- | --- | --- |
| Text | characters after whitespace collapse | — |
| Emphasis / strong | presence, nesting, content | `*` vs `_` |
| Code span | content verbatim | backtick count |
| Link | destination, content | title, angle-bracket form |
| Image | destination, alt content | title |
| Line break | hard break preserved; **soft break becomes a single space** | — |

The soft-break rule is the one that makes hard-wrapped markdown and unwrapped Confluence produce
the same hash.

## Algorithm

1. **Retrieve** the document from its authoritative store.
2. **Strip metadata** — YAML front matter, or the store's native equivalent (Confluence page
   properties, labels). Metadata carries the approval; hashing it would make every approval
   invalidate itself the moment it was recorded.
3. **Strip generated stamps** — mirror banners per ADR-015 are not content.
4. **Parse** into the block model above.
5. **Normalize**: Unicode NFC on all text; collapse whitespace runs to one space within inline
   content; drop leading and trailing whitespace per block; drop empty blocks. Content inside code
   blocks and code spans is exempt from every whitespace rule except line-ending normalization
   to `LF`.
6. **Serialize** the normalized model to canonical JSON — UTF-8, object keys sorted, no
   insignificant whitespace, in the manner of RFC 8785.
7. **Digest**: `sha256` of those bytes, lowercase hex, **full 64 characters**.
8. **Prefix** with the algorithm version: `canonical/v1:sha256:<64 hex>`.

## Two hashes, two jobs

| Field | Over | Answers |
| --- | --- | --- |
| `content_hash` | The canonical form | "Is this the document that was approved?" — survives reformatting and migration |
| `source_hash` | Raw stored bytes, unmodified | "Has the stored artifact been touched?" — detects in-place edits, mirror tampering, encoding drift |

`content_hash` is the approval anchor. `source_hash` is diagnostic and is recorded alongside it,
never used to invalidate an approval on its own.

## Materiality becomes decidable

`docs/spec/DOCUMENT_LIFECYCLE.md` defines a material change as one "that could alter what an agent
does" — a judgment call, and therefore not enforceable.

The canonical hash makes it mechanical: **a change in `content_hash` is a material change.** The
version increments, `status` returns to `proposed`, `human_approved` resets, and the gate re-opens.

The exception is narrow. A change that alters the canonical hash but genuinely carries no meaning —
a typo, a broken link repaired — may be declared **editorial**. That declaration is a human act,
recorded like any other, and an agent may never make it alone. Defaulting to material is the safe
direction, per the failure-mode table above.

## Migration attestation

Canonicalization reduces the risk; it cannot eliminate it. Any two stores may differ in ways no
normal form anticipates, and the algorithm itself will change.

So the process must handle the hash changing legitimately, without either silently carrying the
approval forward or forcing the human to re-read every document.

A **migration attestation** is a record asserting that a set of documents moved representation
without changing meaning:

```yaml
id: ATT-0001
kind: migration
from: {store: local, algorithm: "legacy/truncated-32"}
to:   {store: local, algorithm: "canonical/v1"}
attester: <authenticated identity>
attested_on: <timestamp>
method: >
  How equivalence was established -- rendered diff, round-trip comparison,
  or human reading. Naming the method is required; "trust me" is not a method.
documents:
  - id: ADR-009
    prior_hash: sha256:c83a5f71...
    new_hash: canonical/v1:sha256:...
    approval: APR-0002        # the approval being re-anchored
```

An attestation **re-anchors** existing approvals to new hashes. It does not re-approve content: the
human is attesting that nothing changed, not deciding again that they agree. That distinction keeps
the audit trail honest — a reader can tell which approvals were judgments and which were carries.

One attestation may cover many documents. An attestation an agent writes without a human attester
is void.

## The algorithm is versioned

`canonical/v1` is a prefix, not decoration. Any change to the rules above produces different hashes
for unchanged documents, which is itself a migration and requires an attestation covering every
affected approval.

**The first use of this mechanism is this specification.** Existing records `APR-0001` through
`APR-0003` carry truncated 32-character digests over raw markdown bodies. Adopting `canonical/v1`
changes all of them, so those approvals must be re-anchored by attestation rather than quietly
recomputed.
