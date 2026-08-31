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

## Agent-specific projections
The graph is shared, but every durable agent has a relevance projection. Example: a Compliance agent highly weights compliance decisions, product risks, privacy policies, prior compliance questions, and related human decisions. An iOS agent weights UI architecture, Swift patterns, client-side pitfalls, relevant requirements, and code/test relationships.

This prevents silos while preserving role continuity.
