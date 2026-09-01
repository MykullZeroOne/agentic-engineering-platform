---
id: PRD-009
type: prd
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-31
---

# PRD-009 — Agent Provenance and Checkpoints

## Objective
Make every agent-authored change explainable after the fact: what context the agent held, what
it changed, why, and what state it can be restored to — without the original transcript.

## Artifacts
1. **Context manifest** — the set of references an agent held at a point in time. Already cited
   as `context_manifest` in `docs/schemas/run-record.example.yaml:12` and never defined.
2. **Checkpoint** — a restorable point in an agent run, anchored to a commit.
3. **Session** — the run itself, with its ordered events.
4. **Restore plan** — the derived steps to return a working tree to a checkpoint.

Per ADR-021 these are the `kairo.*.v1` schemas, ported into `docs/schemas/` with their field
shapes unchanged. AEP owns them after the port; they are not a dependency.

## Requirements
- A run record's `context_manifest` resolves to a manifest that exists and validates.
- Every manifest entry carries what it points at and why it was included: `id`, `kind`, `role`,
  and where applicable `repo`, `commit`, `path`, `pointer` (`kairo pkg/manifest/manifest.go:24`).
- A checkpoint names its base commit, so provenance survives independently of the agent runtime.
- Provenance attaches to a commit without rewriting it. Squash merge discards the branch, so a
  trailer written before the merge is the only record a *human* controls, and the convention has
  now failed eight times on this trunk. GitHub writes the pull request number into every squash
  subject itself, which is a second surviving link that depends on nobody remembering anything;
  `devctl work advance` uses it when the trailer is absent (WI-0032). Git notes in a dedicated ref (`kairo pkg/gitref/notes.go:8`) can be
  attached after the fact; commit trailers (`pkg/gitref/trailers.go:8-12`) carry the same
  identifiers when they are written in time. Both are required, because neither alone is
  reliable.
- A manifest entry's `pointer` resolves and validates against git: `git://<repo>@<commit>:<path>`
  (`kairo pkg/gitref/pointer.go:9`). A pointer that cannot be resolved is a finding, not a
  silent gap.
- Provenance is queryable without the transcript: search, related-node, traverse, and explain
  over indexed artifacts (`kairo pkg/knowledge/engine.go:15`).
- Indexes are disposable projections and rebuildable from the canonical artifacts alone.
- The index backend is AEP's, over Postgres + pgvector per ADR-003 and ADR-011. The ported
  `knowledge.Engine` is used as an interface; the ported SQLite backend is a test target, not
  AEP's store.
- Provenance capture never blocks a run. A failed index write is a reported condition.
- Retention and redaction are AEP's: an artifact referencing a secret must be removable without
  invalidating the run record that points at it.

## Non-goals
- **This moves no human gate, approval record, or work state out of AEP.** Gates
  (`.agentic/registries/gates.yaml`), approvals (`.agentic/approvals/`), and the work store named
  by `work_store` (ADR-012) stay where they are and stay authoritative. Kairo has no concept of
  any of them, and acquiring one is not wanted.
- Not the PRD-006 context packet. Provenance records what an agent *held*; PRD-006 decides what
  an agent *should be given*, canonical knowledge first. Kairo's `HydrateContext` returns
  checkpoints, manifests, events and pointers (`kairo pkg/knowledge/types.go:104`) and none of
  the role contract, approved requirements, or applicable ADRs that PRD-006 requires.
- Not AEP's document `content_hash`. Provenance artifacts are hashed by the ported `corehash`;
  AEP documents are hashed per ADR-016. The two are not the same algorithm — see ADR-021.
- Not an execution adapter. Kairo's `pkg/adapters` captures and formats handoffs; it does not
  run a provider, and ADR-021 ports it under a different name. ADR-008 adapters are unaffected.
- Not a hook execution mechanism. Kairo's plugin host runs external binaries; ADR-021 ports it
  unwired, because whether hooks may do that is ADR-020's question.
- No cross-project or shared-tenant provenance in this phase.

## Acceptance criteria
A future Compliance session can take a merged change it did not make, resolve the run record's
`context_manifest`, and state which approved requirements and which prior decisions were in the
agent's context when it made that change — citing the manifest entry for each, and without
loading the original chat transcript.

A second criterion, deliberately separate because it is the one that fails silently: deleting
every index and rebuilding from the canonical artifacts alone reproduces the same answer.
