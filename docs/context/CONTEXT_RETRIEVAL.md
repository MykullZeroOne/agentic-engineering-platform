---
id: DES-CONTEXT-RETRIEVAL
type: design
tier: 2
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

# Context Retrieval Strategy

## Goal
Give agents the smallest sufficient context with explicit provenance.

## Retrieval priority
1. Explicit work-item references.
2. Approved ADS/requirements and parent intent.
3. Applicable policies/standards/ADRs.
4. Direct graph neighbors: components, dependencies, tests, prior PRs.
5. Durable agent/role memories.
6. Semantic similarity across validated memories/history.

## Ranking factors
- authority tier;
- explicit linkage;
- graph distance;
- role relevance;
- semantic relevance;
- project/domain scope;
- confidence;
- recency/validity;
- historical usefulness.

## Context levels
### L0 — identity
Role, permissions, parent, completion criteria.

### L1 — task essentials
Task, requirements, acceptance criteria, dependencies, architecture constraints.

### L2 — related experience
Patterns, lessons, similar prior work, known pitfalls.

### L3 — evidence on demand
Full ADR, prior transcript, PR discussion, test logs, raw events.

The default packet should favor L0-L2 and supply references for L3.
