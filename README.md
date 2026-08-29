# Agentic Engineering Platform (AEP)

A vendor-neutral, subscription-first agentic software development platform that models a real engineering organization around GitHub, open tooling, reusable agent skills, deterministic hooks, durable per-role memory, and interchangeable execution providers such as Claude Code and Codex.

## Product thesis

Human leadership should operate at the level of intent, governance, architecture approval, and release confidence—not at the level of writing production code. Specialized agent roles perform the work of a normal product and engineering organization, using explicit handoffs, recursive clarification loops, specialist sub-agents, escalation paths, quality gates, and a shared institutional knowledge system.

GitHub is the default control plane. AEP owns the process model, knowledge model, role model, orchestration, observability, and memory. Agents and models remain replaceable.

## Core principles

1. **Human at the top** — human acts as CTO/Product Owner and approves material product, legal, architecture, security, and release decisions.
2. **Role != model** — `compliance.primary` or `ios.engineer` is a durable organizational identity; Claude/Codex sessions are transient runtimes.
3. **GitHub owns work state** — issues, PRs, checks, releases, and work graph are authoritative execution state.
4. **Knowledge is structured** — PRDs/ADRs remain human-friendly views over richer machine-readable Agent Development Specifications (ADS).
5. **Every tier loops** — each lead agent can delegate, resolve, ask its parent, escalate to human, re-evaluate, and only then hand off.
6. **Hooks make behavior deterministic** — context loading, validation, CI, traceability, and memory consolidation happen at lifecycle boundaries.
7. **Memory is evidence-backed** — memories have provenance, confidence, scope, ownership, and supersession semantics.
8. **Everything is observable** — humans can see what each agent is doing, inspect context/memory/tools, intervene, pause, redirect, or hand off.
9. **Learning is measurable** — immutable run/event history supports agent regression, skill regression, model comparison, routing improvement, and future training.
10. **Subscription-first, API-optional** — use supported CLIs and existing Claude/Codex subscriptions where practical; APIs are adapters, not foundations.

## Repository map

- `docs/spec/` — authority model, document lifecycle, registry contracts
- `docs/vision/` — vision, principles, personas, success criteria
- `docs/prd/` — human-readable product requirements
- `docs/ads/` — machine-oriented Agent Development Specification design
- `docs/adr/` — architecture decisions
- `docs/architecture/` — system, control, knowledge, execution, communication, evaluation planes
- `docs/agents/` — role contracts, hierarchy, loops, specialist sub-agents, permissions
- `docs/memory/` — memory model, graph, event sourcing, consolidation, regression/training
- `docs/context/` — retrieval, context packets, provenance, role-specific projections
- `docs/workflows/` — end-to-end SDLC, hooks, escalation, quality gates
- `docs/github/` — GitHub control-plane integration and repository adoption
- `docs/security/` — governance, human approvals, permissions, legal/compliance boundaries
- `docs/operations/` — observability, agent workspaces, runtime adapters
- `docs/roadmap/` — implementation phases and milestones
- `docs/schemas/` — canonical YAML/JSON schemas
- `.agentic/` — this repository's configuration, hook bindings, and vocabulary registries
- `scripts/` — documentation validator and index generator
- `examples/` — example product request and resulting artifacts

Start at [`docs/DOCUMENTATION_INDEX.md`](docs/DOCUMENTATION_INDEX.md), which is generated from
document front matter and gives the intended reading order.

## Status

Specification, plus the tooling that keeps it consistent. There is no application code or runtime
yet. Every canonical document declares its authority tier and approval state
([`docs/spec/AUTHORITY_MODEL.md`](docs/spec/AUTHORITY_MODEL.md)); `scripts/validate_docs.py`
enforces that in CI. Most of the corpus is `draft` — the specification is being laid out, not
finished.

## Recommended MVP

The MVP should prove one complete loop:

`Human idea -> BA clarification -> Compliance/Legal preflight -> Human approval -> ADS -> Grooming -> Architecture -> Plan DAG -> Codex implementation -> QA/Review -> Human merge -> Memory consolidation -> regression record`

Start with GitHub + local subscription-backed Claude/Codex workers + file-backed knowledge. Add PostgreSQL/pgvector and richer graph projections only after the workflow is useful.
