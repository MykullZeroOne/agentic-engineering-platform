---
id: DES-ENGINEERING-LOOP
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
last_reviewed: 2026-08-29
---

# Engineering Agent Loop

## Input
One implementation work unit with requirements, architecture constraints, acceptance criteria, verification obligations, and context packet.

## Steps
1. Verify prerequisites and worktree state.
2. Inspect current implementation and related organizational patterns.
3. Produce a scoped implementation plan.
4. Implement only authorized scope.
5. Add/update required tests and documentation.
6. Execute local checks.
7. Create evidence summary mapping code/tests to requirements.
8. Open/update PR with traceability.
9. Respond to review/QA findings via repair loops.

## Hard boundaries
- no unapproved architecture deviations;
- no bypass of required human gates;
- no silent requirement changes;
- no self-approval of final implementation.
