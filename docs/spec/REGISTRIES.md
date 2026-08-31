---
id: SPEC-REGISTRIES
type: spec
tier: 0
status: proposed
version: 2
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
enforcement: check
---

# Registries

Registries hold vocabulary that the platform will consume at runtime: gate IDs, hook points, and
state tokens. They live in `.agentic/registries/` as YAML rather than in `docs/` as prose,
because they are configuration, not narrative.

**This document does not restate their contents.** With no generator to keep two copies in sync,
duplicating a registry in markdown guarantees drift. Read the YAML; it is commented.

| Registry | Defines | Replaces |
| --- | --- | --- |
| `gates.yaml` | Human-approval gate IDs, triggers, approvers | 4 divergent vocabularies across `project.yaml`, `GOVERNANCE.md`, `CLAUDE.md`, `agent-role.example.yaml` |
| `hook-points.yaml` | The closed set of lifecycle boundaries and the hook ID namespace | 3 divergent sets across `hooks.yaml`, `HOOKS.md`, ADR-007 |
| `states.yaml` | Six state axes and their canonical tokens | 5 overlapping vocabularies with colliding tokens |
| `vocabularies.yaml` | Closed enumerations that are not state axes | Hardcoded constants in `validate_docs.py` and enumerations inline in specifications |

`states.yaml` and `vocabularies.yaml` are separate for a structural reason, not a stylistic one:
a state axis has transitions, terminal values and a normal flow; a vocabulary is a flat closed
set. Folding one into the other would give every vocabulary fields it has no meaning for.

Approval records live alongside them in `.agentic/approvals/`. They are not a registry — each is a
single immutable event — but they are the provenance the registries' gates depend on. See ADR-013.

## Rules

1. **A registry is the sole source for its vocabulary.** A document that names a gate, hook point,
   or state must use a token defined here. `scripts/validate_docs.py` enforces this.
2. **Registries are gated.** They live under `.agentic/`, so changes trigger the `platform_config`
   gate.
3. **Reconciliation is recorded, not discarded.** Every entry carries `aliases` listing prior
   spellings and `sources` listing the documents it reconciles. `legal_approval` did not
   disappear — it is an alias of `legal`.
4. **Closed sets stay closed.** Adding a hook point or a state value is a deliberate change with
   an approval, not an incidental one.

## Known gaps

Recorded rather than invented. Each needs a decision before it can be added:

- **`question_escalated` hook point.** `docs/workflows/QUESTION_ESCALATION.md` describes escalation
  as a lifecycle process, but no accepted document specifies a hook contract for it. Escalation
  currently has no deterministic boundary, which means the notification-quality requirements in
  PRD-003 rest entirely on agent compliance.
- **Policy and standard registries.** Tier 3 has no artifacts yet. `docs/policies/` and
  `docs/standards/` do not exist, though five documents depend on the tier. Until they exist, the
  promotion ladder in `docs/memory/CONSOLIDATION.md` has no terminus.
- **Skill schema.** Skills are referenced with a `name@version` syntax throughout
  (`compliance-review@1`, `implement-story@4`) but have no schema, and `.agentic/skills/` is
  empty.
- **Evidence package schema.** The package now has a *shape* —
  `.github/pull_request_template.md` — covering what changed, why, traceability,
  verification, and gates crossed. It does not yet have a *schema*. Under ADR-017 an
  evidence package is structured data of which the pull request body is one projection, so
  the template is an interim human-facing form, not the definition. QA and review still
  cannot consume it mechanically.
- **External reference has no authority tier.** `docs/spec/AUTHORITY_MODEL.md` covers project
  intent (tiers 0-4), execution state (5), and learned memory (6). Facts about the outside world —
  a library's API, a vendor's documentation — are none of these. They are not project knowledge and
  must not be conflated with it, but agents need them and their retrieval must reach the context
  provenance manifest, since stale library knowledge is a common cause of the "stale context"
  failure `CONTEXT_PROVENANCE.md` exists to distinguish. Distinct from ADR-015, which governs where
  *our* documents live.
- **Capability namespace.** `agent-role.example.yaml` grants `github.read`, `repository.write`,
  `deployment.*`. The namespace, wildcard semantics, and deny-precedence rules are unspecified.
