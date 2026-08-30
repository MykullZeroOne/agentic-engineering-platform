---
id: ADR-017
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-30
supersedes: null
superseded_by: null
last_reviewed: 2026-08-30
---

# ADR-017 — Structured Base Data, Documents as Projections

## Context

ADR-006 already decided this and we have not been following it. Its consequence reads: "PRDs,
issues, and evidence matrices should be projections of a shared underlying specification whenever
possible rather than independently maintained sources of truth." `ADS_OVERVIEW.md` even enumerates
the projections, including "GitHub issues: execution view."

In practice we have treated markdown files as the base data. Every mechanism built so far inherits
that mistake: ADR-016 needs a CommonMark parser to hash a document, ADR-015 needs a canonical form
per store, and the validator reads paths. All of that is complexity created by hashing and
diffing *a rendering* rather than the thing being rendered.

Two further questions were open and are settled here: whether several interfaces writing the same
artifact constitutes sync, and whether AEP should own storage or defer to git indefinitely.

## Decision

**Canonical artifacts are structured data. Documents are projections of it. Exactly one store is
authoritative, and AEP-native storage is the target with git as the stepping stone.**

1. **Base data is structured** — JSON or YAML with an explicit schema — for everything that
   reduces to fields: ADS entities, requirements, acceptance criteria, work items, approvals,
   attestations, registries.

2. **Prose stays prose, inside structure.** An ADR's value is its argument. `Context`, `Decision`,
   `Alternatives considered`, `Consequences` and `Risks` become named fields whose *contents* are
   prose. The skeleton is data; the reasoning is text. Partial structuring is the honest answer,
   not a compromise.

3. **Markdown becomes a generated view.** The `docs/**` tree becomes a rendering, not the source.
   This is ADR-015's mirror rule turned inward: generated, stamped, never edited in place.

4. **Multiple write interfaces, one store.** The app, chat, the CLI, and a pull request review are
   all clients of the same store. This is *not* sync and does not engage ADR-012's prohibition:
   ADR-013 already established chat, `github_review` and `ui` as approval surfaces writing one
   record. Two authoritative copies reconciled by machinery remains forbidden.

5. **AEP-native storage is the target; git is the stepping stone.** Today the store is the
   repository, because AEP does not exist and specifications cannot wait on the system they
   specify. As AEP gains a store, projects migrate under ADR-015's provider model. Git remains a
   supported provider permanently — it is the right answer for teams that want specifications
   versioned with code.

6. **Derived indexes are not sources of truth.** PostgreSQL holds projections for query and UI
   performance, rebuildable by replay, exactly as ADR-004 requires of everything derived.

## Alternatives considered

**Keep markdown as the base data.** Rejected. It forces every mechanism to reason about a
rendering: canonical forms per store, a parser dependency for hashing, diffing prose to detect a
status change. It also makes traceability lexical — you cannot ask "which requirements does this
ADR govern" without parsing English.

**Full structuring, including prose.** Rejected. Decomposing an argument into fields destroys the
argument. "Why we rejected TypeScript" is three paragraphs of reasoning, not an enum, and an ADR
whose alternatives are checkboxes is worthless to the person who has to understand the decision in
two years.

**AEP-native storage immediately.** Rejected on circularity. AEP does not exist; specifications
written now would have nowhere to live, and we would be unable to specify the system until we had
built it. Git dissolves this at the cost of nothing that matters.

**Git as the permanent and only store.** Rejected as a target, retained as a provider. Writing
through commits is workable but clunky for interactive editing, offers no query surface, and ties
the knowledge plane to one tool's semantics — which ADR-001 explicitly set out to avoid.

**Bidirectional sync between AEP and git.** Rejected, as in ADR-012 and ADR-015. Two authoritative
stores reconciled by machinery fail worst under conflict.

## Consequences

- **ADR-016 changes shape.** Hashing becomes canonical JSON (RFC 8785) over the structured base,
  plus normalized text for prose leaves. The CommonMark parser dependency disappears, and with it
  the block model and the risk that two parsers disagree.
- The `docs/**` markdown tree must eventually be generated. Until the generator exists it stays
  hand-authored, and that is a known inconsistency rather than a design.
- A base-data schema is now the blocking deliverable. `DOCUMENT_LIFECYCLE.md` front matter,
  `ads.schema.example.yaml`, and `states.yaml` are three partial views of it.
- Traceability becomes queryable rather than lexical: requirement to work unit to evidence is a
  graph walk, not a grep.
- `scripts/validate_docs.py` becomes schema validation over data instead of front-matter parsing
  over files.
- Migration from git to an AEP-native store is a store migration under ADR-015, requiring the
  attestation ADR-016 defines.

## Risks

- **Structuring the wrong things.** Over-structure and documents become forms nobody wants to
  write; under-structure and we are back to parsing prose. The boundary — fields for what is
  queried, prose for what is argued — is a judgment that will need revisiting with evidence.
- **The generated-markdown step is where this most likely stalls.** Hand-authoring markdown while
  claiming data is canonical is exactly the drift this repository keeps legislating against, and
  it will be true of us until the generator ships.
- **AEP-native storage weakens atomicity.** A specification in AEP's store no longer changes in the
  same commit as the code implementing it, so "which version was in force at this merge" becomes a
  timestamp question. Git-provider projects keep the stronger guarantee, which is a real reason the
  provider stays supported rather than deprecated.
- **The interim has two half-truths**: structured data that is not yet the source, and markdown
  that is not yet generated. That state must be short, or it becomes the architecture.
