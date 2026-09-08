---
id: DES-RELEASE-LOOP
type: design
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-09-06
---

# Release Loop

A configuration of the universal agent loop (ADR-023, `UNIVERSAL_AGENT_LOOP.md`).

## Input
A release candidate that has passed Integration Lead's system-level validation, with its evidence package.

## Objective
Prepare release evidence. The Release Lead does not itself approve production release; the human owns that gate.

## Role-specific steps
- (refines step 3 Analyze) Assess deployment, migration, rollback, and
  observability readiness for the candidate, and what the release notes must
  cover.
- (refines step 4 Delegate) Delegate to the Deployment, Migration, Rollback,
  Observability, and Release Notes specialists as relevant.
- (refines step 8 Produce) Assemble the release evidence package: deployment
  plan, migration plan, rollback plan, observability plan, and release notes.
- (refines step 9 Hand off) Route the assembled evidence to the human for the
  production release gate.

## Release readiness gate
- deployment plan defined;
- migration plan defined or explicitly not applicable;
- rollback plan defined;
- observability plan defined;
- release notes drafted;
- no unresolved specialist finding.

Passing this gate makes the candidate presentable to the human; it does not
authorize release. `production_release` (`.agentic/registries/gates.yaml`) is the
authority for that: "Approval to deploy or roll back a release in production,"
approver `human.cto`, `risk_tier: irreversible`, triggered when a "release
candidate reaches the human release gate (SDLC phase 8)."

## Configuration of ADR-023
- **Step-4 specialists**: Deployment, Migration, Rollback, Observability,
  Release Notes.
- **Step-7 gate**: human-owned. `production_release`
  (`.agentic/registries/gates.yaml`) is a human gate; the Release Lead's own
  evaluation of the Release readiness gate above never substitutes for the
  human's explicit approval.
- **Step-8 artifacts**: deployment plan, migration plan, rollback plan,
  observability plan, and release notes — the release evidence package.
- **Return sources**: the human, when the `production_release` gate is declined.
