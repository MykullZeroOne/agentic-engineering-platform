---
id: DES-END-TO-END-SDLC
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

# End-to-End Agentic SDLC

## Phase 0 — Human Intent
Human submits an idea/problem/objective.

## Phase 1 — BA/Product Loop
BA clarifies intent, delegates preflight specialists, resolves/escalates questions, and produces PRD + ADS. Human approves product specification.

## Phase 2 — Grooming
Requirements Lead checks ambiguity, size, testability, edge cases, dependencies, and measurable acceptance criteria. Gaps return to BA.

## Phase 3 — Architecture
Architecture Lead and specialists map impact, apply existing ADRs, propose new ADRs where required, and obtain configured human/security approvals.

## Phase 4 — Planning
Planner creates work-unit DAG with requirements, capabilities, dependencies, verification, risk, and context references. Human approval is optional/configurable by risk.

## Phase 5 — Orchestration and Engineering
Orchestrator releases ready DAG nodes and dispatches specialized workers. Workers implement in isolated worktrees/branches and create PR evidence packages.

## Phase 6 — QA and Review
QA verifies requirements via evidence. Review independently checks architecture, standards, security, scope, maintainability, and tests. Findings route back to the responsible agent.

## Phase 7 — Integration
Integration Lead validates combined work and release-candidate behavior.

## Phase 8 — Human Release Gate
Human reviews material evidence and approves merge/release based on configured policy.

## Phase 9 — Consolidation/Evaluation
Post-merge hooks consolidate memory candidates, update knowledge relationships, recalculate blocked work, record evaluation metrics, and optionally promote regression cases.
