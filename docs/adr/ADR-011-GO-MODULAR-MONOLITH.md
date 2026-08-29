---
id: ADR-011
type: adr
tier: 1
status: proposed
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-011 — Go and a Modular Monolith for the First Implementation

## Context

`docs/architecture/TECHNOLOGY_STACK.md` *recommends* Go with TypeScript as a live alternative, and
recommends starting as a modular monolith. Roadmap Phase 1 assumes both. That document is now
`approved` at tier 2, which makes the ambiguity worse rather than better: an approved recommendation
still is not a decision, and agents cannot tell whether they may choose otherwise. This ADR makes
it a tier-1 decision that outranks the description.

## Decision

- **Go** for the control service, `devctl`, and the local worker daemon.
- **React + TypeScript** for the UI, over WebSocket/SSE for activity and PTY streaming.
- **A single deployable modular monolith**, with one package per plane (ADR-010) and explicit ports
  between them.
- **PostgreSQL + pgvector** as the single store, per ADR-003.

Split into separate services only when measured load or operational need requires it, per
`docs/architecture/DEPLOYMENT.md`.

## Alternatives considered

**TypeScript end to end.** The strongest alternative, and a real trade-off: shared types across UI
and API, one toolchain, and faster UI-adjacent velocity. Rejected because the load-bearing work is
not web serving — it is PTY and process supervision of provider CLIs, long-running orchestration,
and **single-binary distribution** of `devctl` and the worker daemon onto a user's own machine
(ADR-008). Node can do the first two; the third is where it hurts most and matters most.

**Python.** Best agent-ecosystem fit. Rejected on the same distribution ground — `devctl` and the
worker must ship as a single binary — and on long-running process supervision ergonomics.

**Rust.** Rejected: no clear benefit over Go for this workload, at a meaningful cost in iteration
speed while the design is still moving.

**Microservices from the start.** Rejected. `DEPLOYMENT.md` already says split only when load
justifies it, and splitting before the Control/Knowledge seam is understood would force distributed
transactions across exactly the boundary we are least sure about.

## Consequences

- `devctl`, the worker, and the control service share Go packages; the CLI is not a separate
  reimplementation of the API.
- UI and backend do **not** share types, so the API contract must be explicit. Generate the client
  from an OpenAPI spec rather than hand-maintaining two definitions.
- Plane boundaries become Go package boundaries with ports, so a later split is a deployment change
  rather than a rewrite.
- Two toolchains to maintain, and contributors need both.

## Risks

- **Go's LLM and agent library ecosystem is thinner than Python's.** This is the obvious objection
  and it is largely neutralized by ADR-008: subscription-first means we drive official CLIs as
  subprocesses rather than depending on provider SDKs. If that decision is ever superseded, this
  one should be revisited with it.
- **A modular monolith erodes into a ball of mud.** Mitigated by ports at plane boundaries and by
  the ADR-010 rule that a change touching four planes is a design smell.
