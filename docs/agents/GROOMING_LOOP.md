---
id: DES-GROOMING-LOOP
type: design
tier: 2
status: draft
version: 2
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-09-06
---

# Requirements / Grooming Loop

A configuration of the universal agent loop (ADR-023, `UNIVERSAL_AGENT_LOOP.md`).

## Objective
Turn approved product intent into implementation-ready, testable requirements without prematurely designing the solution.

## Checks
- ambiguity;
- requirement size/cohesion;
- testability;
- measurable acceptance criteria;
- missing edge cases;
- dependency completeness;
- actor/authorization implications;
- data lifecycle implications;
- accessibility/UX requirements;
- conflicting or duplicate requirements.

## Outcomes
- READY — requirement may proceed to architecture/planning.
- NEEDS_PRODUCT_CLARIFICATION — return structured questions to BA.
- NEEDS_SPECIALIST — delegate to specialist.
- BLOCKED — dependency/decision prevents readiness.

## Configuration of ADR-023
- **Step-4 specialists**: Testability Analyst, Dependency Analyst, Edge-case
  Analyst, UX/Accessibility Requirements Analyst.
- **Step-7 gate**: agent-evaluated. Passes on the `READY` outcome against the
  Checks above; `NEEDS_PRODUCT_CLARIFICATION`, `NEEDS_SPECIALIST`, and `BLOCKED`
  are non-passing outcomes.
- **Step-8 artifacts**: implementation-ready, testable requirements, groomed
  against the Checks above.
- **Return sources**: none documented.
