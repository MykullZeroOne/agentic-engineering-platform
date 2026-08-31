---
id: GUIDE-IMPLEMENTATION-ROADMAP
type: guide
tier: null
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

# Implementation Roadmap

## Phase 1 — Foundation / usable vertical slice
- Modular monolith API/UI
- PostgreSQL schema
- GitHub app/adapter + `gh`
- Project/agent registry
- Claude/Codex local runtime adapters
- `devctl init/adopt/doctor`
- Agent workspace live activity/session
- Universal agent loop
- BA + Compliance/Legal specialist loop
- file-backed canonical docs + basic ADS
- deterministic hooks

**Exit:** Human idea -> BA questions -> approved PRD/ADS -> one planned story -> Codex implementation -> Claude review -> human merge.

## Phase 2 — Full virtual organization
- Grooming, Architecture, Planning, QA, Review, Integration roles
- hierarchical questions/escalation
- dependency DAG/orchestrator
- GitHub Project state mapping
- skill registry/versioning
- policies and configurable human gates

## Phase 3 — Durable memory/context plane
- event store
- session consolidation
- per-agent memory projections
- pgvector retrieval
- explicit graph entities/edges
- context provenance
- learning inbox

## Phase 4 — Evaluation plane
- regression case creation
- model/skill/prompt/retrieval comparisons
- telemetry dashboards
- learned routing inputs
- historical repository learning

## Phase 5 — Ecosystem/platform
- MCP server/client surfaces
- plugin/provider SDK
- GitLab/Bitbucket adapters
- additional runtime providers
- distributed workers
- optional API model routing
- organization templates/marketplace
