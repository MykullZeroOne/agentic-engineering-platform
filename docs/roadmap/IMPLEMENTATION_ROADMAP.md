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

Intent, not state. Phases, sequencing and exit criteria live here; status lives only in the work
store named by `work_store` (ADR-012 clause 5). Each Phase 1 bullet names the work item that
carries it, and each later phase is one item until it becomes current and decomposes. This file
never says whether anything is done; `devctl work list` does.

## Phase 1 — Foundation / usable vertical slice
- Modular monolith API/UI — WI-0042
- PostgreSQL schema — WI-0043
- GitHub app/adapter + `gh` — WI-0044
- Project/agent registry — WI-0045
- Claude/Codex local runtime adapters — WI-0046
- `devctl init/adopt/doctor` — WI-0047 (doctor: WI-0019)
- Agent workspace live activity/session — WI-0048
- Universal agent loop — WI-0049
- BA + Compliance/Legal specialist loop — WI-0050
- file-backed canonical docs + basic ADS — WI-0051
- deterministic hooks — WI-0052

**Exit:** Human idea -> BA questions -> approved PRD/ADS -> one planned story -> Codex implementation -> Claude review -> human merge. — WI-0053

## Phase 2 — Full virtual organization

Held as WI-0054 until current.
- Grooming, Architecture, Planning, QA, Review, Integration roles
- hierarchical questions/escalation
- dependency DAG/orchestrator
- GitHub Project state mapping
- skill registry/versioning
- policies and configurable human gates

## Phase 3 — Durable memory/context plane

Held as WI-0055 until current.
- event store
- session consolidation
- per-agent memory projections
- pgvector retrieval
- explicit graph entities/edges
- context provenance
- learning inbox

## Phase 4 — Evaluation plane

Held as WI-0056 until current.
- regression case creation
- model/skill/prompt/retrieval comparisons
- telemetry dashboards
- learned routing inputs
- historical repository learning

## Phase 5 — Ecosystem/platform

Held as WI-0057 until current.
- MCP server/client surfaces
- plugin/provider SDK
- GitLab/Bitbucket adapters
- additional runtime providers
- distributed workers
- optional API model routing
- organization templates/marketplace
