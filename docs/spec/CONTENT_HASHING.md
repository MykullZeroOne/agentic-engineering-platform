---
id: SPEC-CONTENT-HASHING
type: spec
tier: 0
status: draft
version: 2
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

ADR-013 binds an approval to a `content_hash` and did not define the computation. This document
defines it. Implements ADR-016, and depends on ADR-017 for what is being hashed.

## What the hash must do

| Failure | Cause | Cost |
| --- | --- | --- |
| **False invalidation** | Representation changed, meaning did not | Approvals lost for no reason; re-approval fatigue trains the human to rubber-stamp |
| **False stability** | Meaning changed, hash did not | The approval covers content nobody approved |

**False stability is the dangerous one.** Where a rule is ambiguous, resolve toward invalidation.

## Hash the artifact, never a rendering

Per ADR-017 the canonical artifact is structured data; markdown is a projection. So the hash is
taken over the structure, and no markdown is parsed at any point.

This is what removes the parser problem. An earlier draft canonicalized a CommonMark block model so
that hard-wrapped markdown and the same content from another store would agree. It worked, but the
hash then depended on a parser library rather than on a specification — two conformant
implementations can differ on edge cases, and a parser upgrade becomes a silent algorithm change.
Structured data has a standard canonical serialization and needs none of it.

## Algorithm

1. **Load** the artifact from its authoritative store as structured data.
2. **Exclude approval metadata** — `status`, `version`, `human_approved`, `approved_by`,
   `approved_on`, `last_reviewed`, and the hashes themselves. These carry the approval; hashing
   them would make every approval invalidate itself the moment it was recorded.
3. **Normalize prose leaves** (below).
4. **Serialize** per RFC 8785 — UTF-8, object keys sorted by code point, no insignificant
   whitespace, canonical number form.
5. **Digest**: `sha256` over those bytes, lowercase hex, full 64 characters.
6. **Prefix**: `canonical/v2:sha256:<64 hex>`.

Field order, indentation, and whether the store holds YAML or JSON are all invisible to the hash.
None of them are the artifact.

## Prose leaves

A prose leaf is a field whose value is human argument rather than data — an ADR's `decision`, a
requirement's `statement`. Normalize before serializing:

- Unicode NFC
- line endings to `LF`
- whitespace runs collapsed to a single space
- leading and trailing whitespace removed
- soft line breaks treated as spaces, so hard-wrapping is invisible
- **fenced content exempt** from whitespace rules, line-ending normalization aside

Plain-text normalization is safe *here* and unsafe over a whole document. Applied to markdown it
erases headings, list nesting and link destinations — meaning disguised as formatting, and the
last of those is severe in a corpus built on cross-references. Inside a structured artifact those
are already fields, so nothing meaningful survives in the leaf for normalization to destroy.

## Two hashes, two jobs

| Field | Over | Answers |
| --- | --- | --- |
| `content_hash` | The canonical form | "Is this the artifact that was approved?" — survives reformatting, re-serialization, and migration |
| `source_hash` | Raw stored bytes | "Has the stored artifact been touched?" — detects in-place edits and encoding drift |

`content_hash` is the approval anchor. `source_hash` is diagnostic and never invalidates an
approval on its own.

## Materiality becomes decidable

`docs/spec/DOCUMENT_LIFECYCLE.md` defines a material change as one "that could alter what an agent
does" — a judgment call, and so unenforceable.

Mechanically: **a change in `content_hash` is a material change.** The version increments, `status`
returns to `proposed`, `human_approved` resets, the gate re-opens.

The exception is narrow. A change that alters the hash but carries no meaning — a typo, a repaired
link — may be declared **editorial**. That declaration is a human act, recorded like any other, and
an agent may never make it alone.

## Migration attestation

Canonicalization reduces risk; it cannot eliminate it. Schemas evolve, stores differ, and the
algorithm itself will change. So the process must handle a hash changing legitimately without
either silently carrying the approval forward or forcing the human to re-read everything.

```yaml
id: ATT-0001
kind: migration
from: {store: local, algorithm: "legacy/truncated-32"}
to:   {store: local, algorithm: "canonical/v2"}
attester: <authenticated identity>
attested_on: <timestamp>
method: >
  How equivalence was established -- rendered diff, round-trip comparison, plain-text
  equivalence check, or human reading. Naming the method is required; "trust me" is not
  a method.
documents:
  - id: ADR-009
    prior_hash: sha256:c83a5f71...
    new_hash: canonical/v2:sha256:...
    approval: APR-0002        # the approval being re-anchored
```

An attestation **re-anchors** approvals to new hashes. It does not re-approve content: the human
attests that nothing changed, not that they agree again. That distinction keeps the audit trail
honest — a reader can tell which approvals were judgments and which were carries.

One attestation may cover many artifacts. An attestation without a human attester is void.

## The algorithm is versioned

`canonical/v2` is a prefix, not decoration. Any change to these rules produces different hashes for
unchanged artifacts, which is a migration and requires an attestation covering every affected
approval.

**The first use of this mechanism is its own introduction.** `APR-0001` through `APR-0003` carry
truncated 32-character digests over raw markdown bodies. Adopting `canonical/v2` changes all of
them, so those approvals must be re-anchored by attestation rather than quietly recomputed.

## Not yet applicable

Nothing in this repository is structured base data today, so there is nothing to hash by these
rules. The existing truncated digests remain in force until the base-data schema exists and
artifacts are migrated. That interim is a known inconsistency, not a design.
