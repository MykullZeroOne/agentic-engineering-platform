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
