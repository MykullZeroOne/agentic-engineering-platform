---
id: ADR-004
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-004 — Preserve Immutable Agent/Event History

## Decision
Capture important platform events immutably and derive memory, analytics, graph projections, and regression datasets from those events.

## Rationale
Derived memory algorithms will evolve. Raw evidence must remain replayable so knowledge can be rebuilt and agent regressions can be reproduced.
