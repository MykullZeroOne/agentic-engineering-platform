---
id: SPEC-REGISTRIES
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
- **Evidence package schema.** Required for merge by `CLAUDE.md`, produced by
  `docs/agents/ENGINEERING_LOOP.md`, consumed by QA and review — and undefined. It should become
  the pull request template.
- **Capability namespace.** `agent-role.example.yaml` grants `github.read`, `repository.write`,
  `deployment.*`. The namespace, wildcard semantics, and deny-precedence rules are unspecified.
