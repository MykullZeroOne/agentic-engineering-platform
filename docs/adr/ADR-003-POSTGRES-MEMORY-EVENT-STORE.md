# ADR-003 — PostgreSQL First for Memory, Graph, and Event Storage

**Status:** Accepted

## Decision
Start with PostgreSQL plus pgvector. Represent graph relationships explicitly in relational tables and store immutable event records alongside materialized projections.

## Rationale
Minimizes operational complexity while supporting relational data, JSON, vector similarity, provenance, analytics, and explicit graph edges. Dedicated graph/vector systems may be added when measured needs justify them.
