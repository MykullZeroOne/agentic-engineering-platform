---
id: DES-ARCHITECTURE-LOOP
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

# Architecture Loop

A configuration of the universal agent loop (ADR-023, `UNIVERSAL_AGENT_LOOP.md`).

## Objective
Determine how groomed requirements fit the approved architecture and whether new architectural decisions are required.

## Role-specific steps
- (refines step 2 Hydrate) Load requirements, architecture map, relevant ADRs,
  patterns, failures, and constraints.
- (refines step 3 Analyze) Identify affected components, interfaces, data,
  security boundaries, operational concerns, and dependencies.
- (refines step 8 Produce) If a new consequential decision is required, propose
  an ADR with alternatives, trade-offs, risk, and recommendation.
- (refines step 7 Validate) Route required human architecture/security
  approvals.
- (refines step 8 Produce) Produce structured architecture constraints for the
  ADS/work graph.

## Readiness gate
- affected architecture known;
- no unresolved contradictions;
- applicable decisions identified;
- new ADRs approved or explicitly deferred;
- security/data/reliability impacts addressed;
- constraints are machine-readable.

## Configuration of ADR-023
- **Step-4 specialists**: Security Architect, Data Architect, Integration
  Architect, Platform/Infrastructure Architect, Performance/Reliability
  Architect.
- **Step-7 gate**: human-owned when a new or superseding ADR is proposed
  (`architecture_decision`) or impact is assessed high-risk
  (`architecture_high_risk`), both in `.agentic/registries/gates.yaml`;
  agent-evaluated against the Readiness gate above otherwise.
- **Step-8 artifacts**: a proposed ADR, when a new decision is required;
  structured architecture constraints for the ADS/work graph.
- **Return sources**: the human, when a proposed ADR or high-risk change is
  declined at the `architecture_decision` or `architecture_high_risk` gate. None
  documented from downstream roles.
