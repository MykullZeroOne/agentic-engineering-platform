---
id: PRD-006
type: prd
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

# PRD-006 — Context Retrieval and Context Packets

## Objective
Deliver concise, role-specific, evidence-backed context to agents automatically before work begins.

## Context packet composition
- agent identity and role contract;
- current work item and parent chain;
- approved PRD/ADS requirements;
- applicable ADRs and policies;
- relevant architecture/components;
- dependency state;
- acceptance criteria and verification requirements;
- role-specific memories and lessons;
- similar prior implementations;
- known pitfalls;
- provenance for each included item.

## Retrieval principles
- canonical knowledge first;
- explicit links before semantic similarity;
- graph proximity + semantic relevance + role relevance + recency + confidence;
- concise conclusions first, deeper evidence on demand;
- do not flood the context window;
- record what was retrieved and why.

## Acceptance criteria
Every substantial agent run has a reproducible context manifest showing exactly which knowledge items were supplied and why.
