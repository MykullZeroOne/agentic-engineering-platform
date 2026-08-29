---
id: DES-GROOMING-LOOP
type: design
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Requirements / Grooming Loop

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
