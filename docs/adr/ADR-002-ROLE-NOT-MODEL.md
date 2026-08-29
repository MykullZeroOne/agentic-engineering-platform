# ADR-002 — Durable Role Identity Is Separate From Runtime Model

**Status:** Accepted

## Decision
Model an agent as a durable organizational identity with role, memory, permissions, history, and responsibility. Model sessions are ephemeral executions of that identity.

## Consequences
- `compliance.primary` can move from Claude to Codex without losing continuity.
- Session transcripts are not the primary memory store.
- Session start/end hooks hydrate/consolidate context.
- Evaluation can compare model runtimes for the same organizational role.
