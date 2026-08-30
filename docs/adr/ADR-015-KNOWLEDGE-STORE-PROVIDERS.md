---
id: ADR-015
type: adr
tier: 1
status: accepted
version: 2
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-30
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-015 — Pluggable Knowledge Stores, One Authoritative Home Per Document

## Context

ADR-012 settled where work items live: one canonical schema, exactly one authoritative store per
project, configurable, never synced. Canonical *documentation* has the same problem and has not
been decided.

Today the corpus assumes a filesystem. `.agentic/project.yaml` maps each knowledge category to a
path, and `scripts/validate_docs.py` reads those paths directly. But teams keep PRDs in Confluence,
specs in Notion, and ADRs in the repository, and PRD-008 promises adoption "without destructive
changes" — mapping existing directories rather than moving files. That promise currently only
covers the case where the existing documents are already files.

Without a decision, the predictable outcome is the one ADR-012 exists to prevent: a PRD in
Confluence, a stale copy in `docs/prd/`, and no answer to which one an agent should believe.

## Decision

**The document schema is the contract. Each document type has exactly one authoritative store.
Stores are pluggable.**

1. **The schema is `docs/spec/DOCUMENT_LIFECYCLE.md`** — `id`, `type`, `tier`, `status`, `version`,
   `owner`, approval fields, and supersession links. A store may represent these however it likes;
   it may not omit them. Per ADR-017 a store holds structured base data, and markdown is a
   generated projection of it rather than the thing stored.

2. **One authoritative store per document type, not per project.** This deliberately differs from
   ADR-012. Work items are one homogeneous collection, so one store is right. Documents are typed
   categories with different audiences — PRDs face stakeholders, ADRs face engineers — and teams
   legitimately keep them in different places. The invariant that prevents drift is *one home per
   document*, which per-type stores satisfy. `.agentic/project.yaml` already maps knowledge per
   category, so the configuration shape already anticipates this.

3. **Configuration extends the existing `knowledge:` block**, with a bare string still meaning a
   local path:

   ```yaml
   knowledge:
     adrs: docs/adr                 # shorthand: local provider at this path
     prds:
       store: confluence
       space: PROD
       parent: Product Requirements
   ```

4. **A store that cannot represent the schema cannot be authoritative.** If a tool has nowhere to
   put `tier`, `status`, or the approval fields without loss, it is not eligible. Provider
   conformance is demonstrable, not asserted.

   This test disqualifies most of GitHub for documents, which is worth stating because the
   assumption is natural. Issues are work items with a state machine, not documents. The wiki is a
   second repository with a weak API and no review flow. Discussions are conversational. **The only
   good document home GitHub offers is repository files** — which is the `local` provider, since
   `local` means "files in the repository," not "on this machine." GitHub-the-work-tracker and
   GitHub-as-a-git-host are different things, and only the second is a knowledge store.

5. **Mirrors are permitted and are never authoritative.** A read-only copy published into another
   system for visibility is fine, must be stamped as generated, and must never be edited in place.
   This is the escape valve that keeps the rule realistic: stakeholders get documents where they
   read them, without a second source of truth.

6. **Migration is one-way and explicit. There is no bidirectional sync.** Same reasoning as
   ADR-012, and the same absolute.

## Alternatives considered

**Repository files always authoritative.** Simplest, and what the corpus assumes today. Rejected
because it forces migration on any team with an existing Confluence or Notion estate, which breaks
PRD-008's non-destructive adoption promise and makes AEP unadoptable exactly where adoption matters
most — projects that already have documentation.

**One store per project, mirroring ADR-012 exactly.** Rejected as needlessly strict. It would force
a team that keeps PRDs in Confluence to also move their ADRs there, or vice versa. Drift happens
when *one document* has two homes; it does not happen because two different documents live in two
different systems.

**External system always authoritative.** Rejected. Documents stop being versioned alongside the
code that implements them, which breaks tier-2 to tier-5 traceability, and local or offline agents
lose access to canonical knowledge entirely.

**Bidirectional sync.** Rejected outright, as in ADR-012. Two authoritative copies reconciled by
machinery fail worst under conflict, which is when the answer matters most.

**Free-form per-type stores with no schema requirement.** Rejected. Without a representable schema,
`authority_level`, supersession, and approval cannot be evaluated, and `docs/spec/AUTHORITY_MODEL.md`
stops functioning for any document outside the repository.

## Consequences

- `.agentic/project.yaml` gains provider syntax under `knowledge:`, backwards compatible with the
  current path shorthand.
- **`scripts/validate_docs.py` must become provider-backed.** It reads the filesystem directly
  today, so it can only validate the `local` provider. Until that changes, no non-local store can
  be adopted without losing validation — which is the practical blocker on this decision.
- Each provider must publish a field mapping showing how it represents every required schema field.
- `devctl` and agents address documents by `id`, never by path. A path is a `local` implementation
  detail.
- Generated mirrors need a stamp and a generator; hand-copying a document into another system is
  the drift this decision forbids.
- **A first-party view over the knowledge plane becomes necessary, not optional.** Pluggable stores
  make the store an implementation detail for agents; without a single place a human reads and
  approves documents, this decision fragments *their* experience across Confluence, GitHub and the
  repository instead. The abstraction has to reach the Experience Plane or it only half works.
  `PRD-005` covers agent observability and has no knowledge surface; that gap is now load-bearing.
- Content hashing must be defined before any non-local provider ships. Specified by ADR-016.

## Risks

- **Content hashing may break across representations.** ADR-013 binds an approval to a
  `content_hash` of the document body. A document in Confluence has a different serialization than
  the same document as markdown, so a naive hash would change on migration and silently invalidate
  every approval. This needs a canonical serialization for hashing, defined before any non-local
  provider ships. **This is the sharpest risk in this decision** and is not yet solved.
- **Loss of atomic history.** A document in an external store no longer changes in the same commit
  as the code it governs, so "which version of the spec did this PR implement" becomes a
  timestamp question rather than a commit question.
- **Offline and local-agent access.** An agent working without network access can read a `local`
  store and cannot read Confluence. Providers may need a cache, which reintroduces a copy — and
  the copy must be explicitly non-authoritative and revalidated, or it becomes drift by another
  name.
- **Provider surface grows.** Each store is code to maintain and a conformance suite to keep
  passing. Mitigated by keeping the adapter thin: the schema carries the complexity.
- **Per-type stores could still fragment.** Nothing stops a team from putting all eight categories
  in eight systems. That is legal under this decision and probably unwise; the constraint is
  correctness, not taste.
