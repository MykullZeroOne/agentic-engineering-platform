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
