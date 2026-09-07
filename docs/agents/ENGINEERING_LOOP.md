---
id: DES-ENGINEERING-LOOP
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

# Engineering Agent Loop

A configuration of the universal agent loop (ADR-023, `UNIVERSAL_AGENT_LOOP.md`).

## Input
One implementation work unit with requirements, architecture constraints, acceptance criteria, verification obligations, and context packet.

## Role-specific steps
- (refines step 2 Hydrate) Verify prerequisites and worktree state.
- (refines step 3 Analyze) Inspect current implementation and related
  organizational patterns.
- (refines step 3 Analyze) Produce a scoped implementation plan.
- (refines step 8 Produce) Implement only authorized scope.
- (refines step 8 Produce) Add/update required tests and documentation.
- (refines step 7 Validate) Execute local checks.
- (refines step 8 Produce) Create evidence summary mapping code/tests to
  requirements.
- (refines step 9 Hand off) Open/update PR with traceability.

## Hard boundaries
- no unapproved architecture deviations;
- no bypass of required human gates;
- no silent requirement changes;
- no self-approval of final implementation.

## Configuration of ADR-023
- **Step-4 specialists**: narrowly scoped sub-agents for implementation,
  refactoring, tests, migrations, documentation, or investigation
  (`ORGANIZATION.md`, Delivery organization).
- **Step-7 gate**: agent-evaluated. Passes when local checks pass, the
  implementation stays within authorized scope, and the evidence summary maps
  code and tests to every requirement/acceptance criterion, per Hard boundaries
  above.
- **Step-8 artifacts**: implementation code, tests, documentation, an evidence
  summary mapping code/tests to requirements, and a PR with traceability.
- **Return sources**: Review, QA.
