---
id: SPEC-ENTITY-MODEL-GUIDE
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
---

# Entity Model — How to Read It

The model itself is `docs/spec/entity-model.yaml`. It is data, not prose, because ADR-017 makes
structured data the canonical artifact. **This document does not restate the entities**, so there
is nothing here to drift against — the same discipline `REGISTRIES.md` applies to the registries.

## What it is

41 entities, 57 relationships, across the seven planes of ADR-010. It reconciles six partial views
that disagreed with each other:

| Source | Contributed |
| --- | --- |
| `DOCUMENT_LIFECYCLE.md` | Front matter, status lifecycle, artifact types |
| `ADS_OVERVIEW.md`, `ads.schema.example.yaml`, `ADS-001` | Intent, Capability, Requirement, AcceptanceCriterion, Evidence |
| `KNOWLEDGE_GRAPH.md` | ~30 node types, ~19 edge types, most overlapping the above |
| `DATA_MODEL.md` | ~50 logical tables |
| `states.yaml` | Six state axes |
| Role, memory, question, run-record examples | Organization and Execution shapes |

75 raw entities and 67 conflicts went in; 41 entities and 48 recorded resolutions came out.

## The three rules it applies

**One concept, one name.** The same thing appears under different names across ADS, the knowledge
graph, and the data model. Each entity records `supersedes_existing` naming what it absorbed, so a
reader tracing an old term can find where it went.

**Structure what is queried, prose what is argued** (ADR-017). Each prose field carries a
`rationale`. An unjustified prose field is usually a structured field someone did not want to
model — the rationale exists to make that visible rather than to decorate.

**No identity depends on a path.** Under ADR-015 a file location is a `local` provider detail, so
every `identity_rule` is store-independent and store-minted.

## Read `open_questions` first

Seventeen questions could not be settled from the corpus and need a human decision. They are not
polish. Several determine whether parts of the model are correct at all:

- Whether Policy and Standard are one entity or two
- Which store is authoritative for each plane
- Whether a narrative PRD with no structured requirements may close the `product_spec` gate —
  seven of the eight current PRDs are narrative
- Whether invariants get a predicate language or are agent-judged
- Whether skills are agent-authorable, which `AUTHORITY_MODEL` and `gates.yaml` currently answer
  differently

The model is `draft` and carries no authority. Treating it as settled would be exactly the
false confidence its own critique pass was built to prevent.

## What the critique found

Three adversarial lenses ran against the draft — completeness and conflict-hiding, the
fields-versus-prose boundary, and traceability with identity and hashing. They returned 70
findings, 17 of them blocking, and the model grew from 31 entities to 41 in response.

Several findings are defects in the **repository**, not in the model, and outlive it:

- **ADR-013 and ADR-016 define `content_hash` incompatibly**, and both are accepted. ADR-016 was
  written as a refinement; it is also a contradiction.
- **`scripts/validate_docs.py` rejects the hash format its own spec mandates** — it requires a
  `sha256:` prefix, and `canonical/v2:sha256:…` does not match.
- **`approved_by` is specified as a pointer to an approval record and implemented as a name**, so
  the provenance link ADR-013 claims does not exist in the data.
- **Every tier-0 spec is `draft`**, including `AUTHORITY_MODEL` itself, so by its own precedence
  rule the document defining authority carries none.
- **`GLOSSARY.md` and `states.yaml` still assert GitHub owns work state**, which ADR-012 v2
  superseded.

These are tracked here rather than silently fixed, because each changes an accepted decision or a
gated artifact.
