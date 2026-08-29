---
id: ADR-003
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-003 — PostgreSQL First for Memory, Graph, and Event Storage

## Decision
Start with PostgreSQL plus pgvector. Represent graph relationships explicitly in relational tables and store immutable event records alongside materialized projections.

## Rationale
Minimizes operational complexity while supporting relational data, JSON, vector similarity, provenance, analytics, and explicit graph edges. Dedicated graph/vector systems may be added when measured needs justify them.
