---
id: SPEC-LOCAL-WORK-STORE
type: spec
tier: 0
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-30
enforcement: check
---

# Local Work Store

ADR-012 makes the work-item schema the contract and the store an adapter, and calls the local
store's on-disk format "the next deliverable" — noting that until it exists, *neither* store is
fully specified. `.agentic/project.yaml` has said `work_store: local` since ADR-012 was accepted,
naming a format that did not exist.

This is that format. Enforced by `scripts/validate_docs.py`.

## Why it matters more than it looks

Three things have been blocked on it:

- **`CLAUDE.md` merge requirement #1** — "links its issue and closes it" — has been unsatisfiable
  since the first commit. Every pull request so far has had to say so.
- **The branch naming rule** `<type>/<issue-number>-<kebab-summary>` cannot be followed without
  issue numbers.
- **ADR-012's consequence** that `IMPLEMENTATION_ROADMAP.md` be converted out of documentation and
  into work items.

## Layout

```
.agentic/work/
  WI-0001.yaml
  WI-0002.yaml
```

One file per work item. Not one file with many items: unrelated items must not conflict in git,
and each item has to be independently addressable.

**The path is not the identity.** `WI-0007` is the identity; `.agentic/work/WI-0007.yaml` is where
the `local` provider happens to keep it. An agent addresses the id and never derives one from a
filename.

## Identity and allocation

Ids are `WI-NNNN`, sequential, never reused — including for cancelled items.

Allocation under `local` is *the store scanning its own index*: the next id is the highest existing
plus one. This is not a client deriving identity from a directory, which the entity model forbids;
under `local` the directory **is** the store, and scanning it is the store's own allocation
mechanism.

Concurrency needs no counter file. Two branches allocating at once both create `WI-0007.yaml`, and
git reports the conflict on the file itself. The filenames are the counter, and the collision is
the detection.

## Format

```yaml
id: WI-0007
store: local
project: PRJ-mykullzeroone/agentic-engineering-platform

type: spec                    # see Types
work_state: ready             # states.yaml, work_state axis
priority: normal              # low | normal | high | urgent

title: Short imperative summary
description: >
  Prose. What needs doing and why. The one prose field: everything an agent
  filters, sorts or routes on is structured above.

serves: [ISC-23]              # what this advances -- see Serves

# Optional linkage. Present when known, omitted when not -- never null-padded.
requirement_refs: [REQ-001-006]
specification_ref: ADS-001
work_unit: WU-001-1-003
parent: WI-0004
dependencies: [WI-0005, WI-0006]
required_gates: [architecture_decision]
origin_finding: FND-0012

# Execution state, written as it happens.
branch: spec/7-local-work-store
pull_request: 4
assignee: null
agent_role: null
```

### Serves

`serves` names what a work item advances. It is a list, and every entry must resolve:

| Entry | Resolves against | Means |
| --- | --- | --- |
| `ISC-<n>` | a claim in `ISA.md` | this item advances a stated criterion for done |
| `PRD-<nnn>` | a document in `docs/prd/` | this item advances an approved product requirement |
| `governance` | nothing; it is a literal | this item improves how the repository governs itself, not the product |

`governance` is deliberately a first-class value rather than an escape hatch. Work on the
validator, the gate check, the approval machinery and the work store is real and worth doing. The
point is not to discourage it. The point is to make its share **visible**, because it was not.

An audit on 2026-09-04 found that 38 of the first 41 work items were governance and three were
product. Nobody chose that ratio; nothing displayed it. A field that every item must fill turns
the question "where is the effort going" from an investigation into a query.

An item that cannot name anything it serves is the finding. Either the target is missing from
`ISA.md`, or the work does not advance the product and should say `governance`, or it should not
be done.

**Enforcement is advisory first.** A missing `serves` is a warning; an entry that does not resolve
is an error. This follows WI-0012 and WI-0013, where the gate check was introduced advisory and
made blocking one change later, once the existing debt was cleared. Flipping the warning to an
error is a separate change, after the backfill.

### Types

Lifted from `CLAUDE.md`'s branch types, so a work item's type and its branch prefix are the same
token rather than two vocabularies that drift:

`feat` · `fix` · `docs` · `spec` · `adr` · `chore` · `refactor` · `test` · `ci`

### States

`work_state` uses the axis in `.agentic/registries/states.yaml`. No state is defined here; the
registry is the sole source.

## Deviations from the entity model, and why

`docs/spec/entity-model.yaml` is `draft` and carries no authority, so this format implements it
where it can and records where it does not:

| Model says | Here | Why |
| --- | --- | --- |
| `cycle` required | Omitted | ADR-014's Cycle has no instances and no allocator. Requiring it would make every work item unwritable |
| `iteration` required | Omitted | A GitHub Projects field with no local meaning |
| `external_key` required | Omitted under `local` | The store mints the id itself; there is no external system to key against |
| `type` is a `VocabularyTerm` ref | Enumeration here | No controlled-vocabulary registry exists yet. `states.yaml` holds state axes, and a type is not a state |

The last one is a genuine gap: `ControlledVocabulary` is an entity in the model with nowhere to
live. Recorded rather than solved by inventing a registry in passing.

## What this store does not hold

ADR-012's intent/state boundary applies unchanged. Phases, sequencing, exit criteria and the
argument for an ordering are **intent** and stay in documentation. Anything carrying a status, an
assignee or a checkbox is **state** and lives here.

A work item is not a place to argue. If a decision is needed, that is an ADR; if a requirement is
needed, that is a specification. The `description` field says what to do, not why it was chosen.

## Migration

Moving to `github` is one-way and explicit, per ADR-012: each item is created in the target store,
`external_key` records the issue number, and `work_store` flips. There is no sync back, and the
local files stop being authoritative the moment the flip lands.
