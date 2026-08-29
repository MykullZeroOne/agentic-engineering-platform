---
id: SPEC-LIFECYCLE
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
last_reviewed: 2026-08-29
enforcement: check
---

# Document Lifecycle and Front Matter

Every canonical artifact declares its identity, authority, and approval state as machine-readable
front matter. Without it, `authority_level`, supersession, and the "canonical outranks learned"
rule in `docs/memory/MEMORY_ARCHITECTURE.md` cannot be evaluated by anything but a human reading
prose.

Enforced by `scripts/validate_docs.py` (enforcement stage: `check`).

## Format

Markdown artifacts carry YAML front matter delimited by `---`. YAML artifacts (ADS, registries,
schemas) carry the same fields as top-level keys.

```yaml
---

id: PRD-002                 # unique, stable, never reused
type: prd                   # see Types
tier: 2                     # see docs/spec/AUTHORITY_MODEL.md
status: draft               # see states.yaml, document_status
version: 1                  # increments on material change
owner: human.cto            # durable identity accountable for it
human_approved: false       # true only when its gate has been closed
approved_by: null           # identity that closed the gate
approved_on: null           # ISO date
supersedes: null            # id, or null
superseded_by: null         # id, or null
last_reviewed: 2026-08-29   # ISO date
enforcement: prose          # rule-bearing artifacts only
---
```

## Types

| `type` | Tier | Location | Numbering |
| --- | --- | --- | --- |
| `principle` | 0 | `docs/vision/` | none |
| `spec` | 0 | `docs/spec/` | `SPEC-<NAME>` |
| `adr` | 1 | `docs/adr/` | `ADR-NNN`, sequential |
| `prd` | 2 | `docs/prd/` | `PRD-NNN`, sequential |
| `ads` | 2 | `docs/ads/` | `ADS-NNN`, **independent** sequence |
| `design` | 2 | `docs/architecture/`, `docs/agents/`, `docs/memory/`, `docs/context/`, `docs/workflows/`, `docs/github/`, `docs/security/`, `docs/operations/` | `DES-<NAME>` |
| `policy` | 3 | `docs/policies/` | `POL-NNN` |
| `standard` | 3 | `docs/standards/` | `STD-NNN` |
| `skill` | 4 | `.agentic/skills/` | `<name>@<version>` |
| `guide` | `null` | `docs/**` | `GUIDE-<NAME>` |
| `schema` | `null` | `docs/schemas/` | none |
| `example` | `null` | `examples/` | `EX-<NAME>` |

Numbers are sequential within their type and **never reused**, including for withdrawn artifacts.
Non-numbered types derive their `id` from the filename.

### `design` is normative, `guide` is not

`tier: null` means an artifact carries **no authority in conflict resolution**. A guide describes;
it does not require.

`design` documents are normative: they state what the platform must be, sitting downstream of
decisions (tier 1) and upstream of implementation. Most of this corpus is currently `design` and
`draft`, which is the accurate description of a system being specified rather than a finished
one. As the platform is built, design content migrates into its proper home — architectural
constraints into ADS, decisions into ADRs, rules into policies and standards — and the design
document becomes a `guide` or is superseded.

## Exemptions

`README.md` and `CLAUDE.md` carry no front matter. They are repository entry points addressed to
humans and agent runtimes, not canonical artifacts, and neither is retrievable as context.

`docs/schemas/*.yaml` and `examples/project.yaml` carry no front matter. They are illustrative
payload fragments whose own top-level keys (`id: ADS-014`) belong to the example, not to the file
as an artifact — and adding lifecycle keys to a sample `project.yaml` would misrepresent what an
adopting project actually writes. Real YAML artifacts — an ADS instance, a role definition — do
carry the fields as top-level keys.

### ADS numbering is independent of PRD numbering

An ADS links to its origin with `source_prd` rather than borrowing its number. One PRD may yield
several ADS, and an ADS may span several PRDs; coupling the sequences breaks the first time either
happens. `docs/schemas/ads.schema.example.yaml` pairs `ADS-014` with `PRD-014` incidentally, not
as a rule.

## Status transitions

Statuses are defined in `.agentic/registries/states.yaml` under `document_status`.

```
draft ──> proposed ──> approved ──> superseded
   │          │        (accepted      deprecated
   │          │         for ADR)
   └──────────┴──> withdrawn
```

- `draft` and `proposed` carry **no authority**. An agent may read them for direction but must not
  cite them as binding, and must not treat a drafted requirement as approved intent.
- `approved` / `accepted` require the artifact's gate to be closed. `human_approved` and
  `approved_by` record it. An agent must never set these itself — Principle 1.
- ADRs use `accepted` and are **immutable** thereafter. Reverse one by adding a new ADR with
  `supersedes: ADR-NNN` and setting the original's `superseded_by`. Never edit an accepted
  decision.
- `withdrawn` retires an artifact that never reached approval. Its number stays burned.

## Versioning

`version` increments on any material change to an `approved` artifact — a change that could alter
what an agent does. Editorial changes do not increment it. A material change to an approved
artifact re-opens its gate: set `status: proposed` and `human_approved: false`.

## Ownership

`owner` is a durable identity (`docs/agents/ROLE_RUNTIME_CONTRACT.md`), never a model or a
session. During bootstrap every artifact is owned by `human.cto`; ownership moves to role
identities as those roles come online.
