---
id: ADR-005
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

# ADR-005 — Hierarchical Agent Communication

## Decision
Default agent communication follows ownership hierarchy: child asks parent; parent resolves or escalates. Cross-functional communication occurs through explicit handoff/escalation routes.

## Rationale
Prevents uncontrolled agent chatter, preserves accountability, and protects human attention.
