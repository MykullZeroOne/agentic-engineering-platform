---
id: DES-KNOWLEDGE-GRAPH
type: design
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
last_reviewed: 2026-08-29
---

# Knowledge Graph

## Purpose
Represent the relationships between intent, decisions, people/agents, code, tests, evidence, failures, lessons, and organizational behavior.

## Node types
- ProductIntent
- PRD
- ADS
- Capability
- Requirement
- AcceptanceCriterion
- ADR
- Policy
- Standard
- Component
- API
- DataEntity
- Issue
- WorkUnit
- Branch
- Commit
- PullRequest
- Test
- Check
- Release
- AgentIdentity
- AgentRun
- Session
- Question
- Answer
- Decision
- Finding
- HumanIntervention
- Memory
- Pattern
- Procedure
- AntiPattern
- RegressionCase

### Provenance nodes

Absorbed from Kairo's `docs/v2/10-graph` per ADR-022. The list above is intent-heavy and
provenance-light: it can express what was decided and required, and not what an agent actually
held, read, or produced. These are the missing half.

- Repository
- File
- GitBlob
- ContextManifest
- Checkpoint
- Prompt
- ToolCall
- TestRun
- ExternalSnapshot
- Model

Kairo's list is not adopted wholesale. Its `Organization` is dropped, because AEP has no
tenancy, and where its names collide with AEP's the AEP name wins per ADR-022's translation
rule — its `Session` and `Decision` are already here, and `Agent` is this document's
`AgentIdentity`.

## Edge types
- DEFINES
- IMPLEMENTS
- GOVERNED_BY
- DEPENDS_ON
- BLOCKS
- VERIFIED_BY
- PRODUCED_BY
- REVIEWED_BY
- FAILED_BECAUSE
- FIXED_BY
- LEARNED_FROM
- USES_PATTERN
- VIOLATES
- SUPERSEDES
- ANSWERS
- ESCALATED_TO
- RETRIEVED_FOR
- RELATED_TO
- AFFECTS

### Provenance edges

Same source, same reason. The set above says how artifacts relate by *intent*; these say what
happened.

- REFERENCES — a manifest or checkpoint points at an artifact
- READ — a run consumed this as context
- MODIFIED — a run changed this
- MOTIVATED — this is why the work started
- EXPLAINS — this accounts for why a change exists
- RESTORES — this returns state to that

`READ` and `MODIFIED` are the pair that matter most: together they make "what context did this
agent hold when it made this change" a graph query rather than an archaeology exercise, which
is PRD-009's entire acceptance criterion.

Kairo's `PRODUCED`, `VALIDATED` and `FAILED` are deliberately **not** added: they are this
document's `PRODUCED_BY`, `VERIFIED_BY` and `FAILED_BECAUSE` under different names, and two
names for one edge is the drift a registry exists to prevent. Its `HAS`, `CREATED` and
`GENERATED` are dropped as too general to constrain anything.

## Agent-specific projections
The graph is shared, but every durable agent has a relevance projection. Example: a Compliance agent highly weights compliance decisions, product risks, privacy policies, prior compliance questions, and related human decisions. An iOS agent weights UI architecture, Swift patterns, client-side pitfalls, relevant requirements, and code/test relationships.

This prevents silos while preserving role continuity.
