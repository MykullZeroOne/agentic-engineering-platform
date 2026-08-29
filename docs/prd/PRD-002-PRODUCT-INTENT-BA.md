---
id: PRD-002
type: prd
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

# PRD-002 — Product Intent and BA Loop

## Objective
Turn a human's product idea into a high-quality, approved product specification through an interactive BA loop rather than one-shot generation.

## BA behavior
The BA Agent shall:
- extract goals, actors, outcomes, assumptions, constraints, and known unknowns;
- ask focused follow-up questions;
- avoid asking questions already resolved by approved knowledge;
- delegate domain-specific analysis to Compliance, Legal, Privacy, Domain, and UX specialists when relevant;
- resolve specialist questions itself when authoritative context exists;
- escalate unresolved material questions to the human;
- repeat until the readiness gate passes;
- generate a human-readable PRD and machine-readable ADS from the same underlying specification model.

## Readiness gate
- critical unknowns = 0;
- target users/actors identified;
- major user journeys defined;
- success criteria defined;
- out-of-scope boundaries recorded;
- compliance/privacy/legal preflight complete where applicable;
- risks and assumptions documented;
- requirements are decomposable.

## Human gate
Human approves the product specification before grooming/architecture execution begins.
