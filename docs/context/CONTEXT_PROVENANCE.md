---
id: DES-CONTEXT-PROVENANCE
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

# Context Provenance

Every context item delivered to an agent is recorded.

Example fields:
- context_item_id
- run_id
- source_type/source_id
- authority_level
- retrieval_method
- retrieval_score
- role_relevance_score
- graph_distance
- reason_for_inclusion
- content_hash/version
- whether later accessed/expanded

## Purpose
When an agent fails, the platform can distinguish:
- model/agent failure despite correct context;
- missing context;
- stale/superseded context;
- incorrect retrieval ranking;
- conflicting authoritative sources.

This is essential for regression testing the Context Plane itself.
