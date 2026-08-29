---
id: SPEC-GLOSSARY
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
---

# Glossary

Canonical terms. Where the corpus previously used several words for one concept, the preferred
term is given and the others are marked deprecated. Machine artifacts use the token; prose may
use the term.

## Work

**Work item** — a unit of tracked work. GitHub is authoritative for its state (ADR-001); the
`work_items` table mirrors it. Its lifecycle is `work_state` in
`.agentic/registries/states.yaml`.
*Deprecated synonyms: story, ticket, task.*

**Work unit** — a node in a planner-produced DAG (`docs/agents/PLANNING_ORCHESTRATION_LOOP.md`).
A work unit becomes a work item when the orchestrator dispatches it. The distinction is real: a
work unit is a *plan* element, a work item is an *execution* element.
*Deprecated synonyms: DAG node, plan node.*

**Work graph** — the dependency DAG over work units.

## Agents

**Role** — an organizational responsibility, e.g. `compliance-analyst`. Versioned.

**Agent identity** — a durable instance of a role, e.g. `compliance.primary`. Owns memory,
permissions, history, and hierarchy. Persists across sessions and across runtime changes
(ADR-002).
*Deprecated synonym: agent, used loosely.*

**Runtime session** — a transient process executing an agent identity via a provider CLI.
*Never* the unit of memory.

**Agent run** — one execution of an agent identity against one work item. A run may span several
runtime sessions (e.g. after a handoff). Evidence and evaluation attach to the run.

**Specialist** — a narrowly scoped agent identity reporting to a parent lead (ADR-005).

**Runtime / provider** — the execution backend (Claude Code, Codex). Interchangeable by design.
*Never* part of a role definition.

## Knowledge

**Canonical knowledge** — approved artifacts in tiers 0–3 (`docs/spec/AUTHORITY_MODEL.md`).

**ADS** — Agent Development Specification. The machine-readable projection of approved product
intent. Peer to the PRD, not derived from it: both project the same underlying specification
(ADR-006).

**Memory** — consolidated, evidence-backed observation. Tier 6. Never authoritative on its own.

**Event** — an immutable record of something that happened (ADR-004). Events are the raw
material; memory is derived from them and can be rebuilt by replay.

**Context packet** — the assembled, provenance-tagged context supplied to one agent run
(`docs/context/CONTEXT_PACKET.md`).

**Provenance** — the record of where a context item came from and why it was included.

## Governance

**Standard** — a normative definition of good practice. Informs review. Does not block on its
own. Tier 3.

**Policy** — a mandatory rule with an enforcement stage. Blocks when violated. Tier 3.

**Gate** — a point requiring explicit human approval. Defined only in
`.agentic/registries/gates.yaml`. A gate is not a tier and not a policy.

**Skill** — a reusable procedure describing *how* work is performed. Tier 4.

**Hook** — a binding of an action to a lifecycle boundary, describing *when* it happens
(ADR-007). Hook points are defined only in `.agentic/registries/hook-points.yaml`.

**Enforcement stage** — `prose`, `check`, or `hook`. How a rule is currently enforced.

## Communication

**Question** — a structured request for information routed through the ownership hierarchy.
Carries a `question_severity`.

**Decision** — a durable, authoritative answer, typically from a human. Outranks memory.

**Finding** — an observation from QA, review, or a specialist that requires disposition. Not
itself a decision.

**Escalation** — routing an unresolved question upward when the current agent lacks the authority
or evidence to answer it.

**Evidence** — an artifact proving a requirement is satisfied: a test result, check, review,
diff, or run record.

**Evidence package** — the complete set of evidence attached to a pull request, mapping changes
to the requirements they satisfy. Required for merge by `CLAUDE.md`.
