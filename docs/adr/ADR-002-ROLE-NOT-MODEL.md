---
id: ADR-002
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: human.cto
approved_on: 2026-08-29
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-002 — Durable Role Identity Is Separate From Runtime Model

## Decision
Model an agent as a durable organizational identity with role, memory, permissions, history, and responsibility. Model sessions are ephemeral executions of that identity.

## Consequences
- `compliance.primary` can move from Claude to Codex without losing continuity.
- Session transcripts are not the primary memory store.
- Session start/end hooks hydrate/consolidate context.
- Evaluation can compare model runtimes for the same organizational role.
