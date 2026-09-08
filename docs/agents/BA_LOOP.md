---
id: DES-BA-LOOP
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

# BA / Product Analyst Loop

A configuration of the universal agent loop (ADR-023, `UNIVERSAL_AGENT_LOOP.md`).

## Input
A human idea, problem statement, feature request, or product objective.

## Role-specific steps
- (refines step 3 Analyze) Parse initial intent into goals, actors, value,
  workflows, assumptions, constraints, out-of-scope items, and unknowns.
- (refines step 3 Analyze) Detect domains requiring specialist review.
- (refines step 4 Delegate) Ask the smallest set of high-value clarifying
  questions of the human specialist.
- (refines step 5 Collect) Update the structured intent model after each human
  response.
- (refines step 7 Validate) Present material decisions, assumptions, risks, and
  specialist findings to the human for product-spec approval.
- (refines step 8 Produce) Generate synchronized PRD and ADS projections.

## BA readiness gate
- objective and value clear;
- actors/personas clear;
- primary workflows clear;
- critical unknowns = 0;
- assumptions explicitly recorded;
- out-of-scope boundaries clear;
- success criteria measurable;
- specialist preflight complete;
- material legal/compliance questions either resolved or explicitly gated;
- ADS requirements can be groomed/tested.

## Specialist behavior
Compliance/legal/privacy specialists are advisory unless a policy explicitly grants decision authority. Legal findings can require a human/legal gate but must never claim legal approval.

## Configuration of ADR-023
- **Step-4 specialists**: Domain Analyst, Compliance Analyst, Legal Preflight
  Analyst, Privacy Analyst, UX/Product Flow Analyst, and the human.
- **Step-7 gate**: human-owned (ISC-10). The BA readiness gate above is necessary
  but not sufficient — the gate passes only on the human's explicit approval of
  the presented material decisions, assumptions, risks, and specialist findings,
  never on the BA's own judgment that the readiness criteria are satisfied.
- **Step-8 artifacts**: synchronized PRD and ADS projections.
- **Return sources**: the human, returning a rejected product specification (per
  ADR-023's return edge); Grooming, via `NEEDS_PRODUCT_CLARIFICATION`
  (`GROOMING_LOOP.md`).
