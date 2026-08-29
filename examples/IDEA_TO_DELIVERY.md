---
id: EX-IDEA-TO-DELIVERY
type: example
tier: null
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

# Example — From Human Idea to Delivery

## Human
"I want parents to assign recurring chores to children and optionally associate an allowance."

## BA loop
BA asks about recurrence behavior, approval, missed chores, reward rules, multiple children, time zones, offline behavior, and parent override.

Compliance/Privacy specialists inspect child data, notification, and financial/reward implications. Legal preflight flags any material terms/consent questions for human review.

Human approves PRD/ADS.

## Grooming
Requirement examples:
- REQ-001 Parent creates household chore.
- REQ-002 Parent assigns chore to child.
- REQ-003 Chore recurrence generates scheduled instances.
- REQ-004 Child marks instance complete.
- REQ-005 Parent approves/rejects completion.
- REQ-006 Approved completion contributes to configured reward.

## Architecture
Architect applies household authorization patterns and proposes ADR for recurrence scheduling if no approved mechanism exists.

## Plan DAG
Domain model -> persistence/API -> client models -> parent UI/child UI -> notifications -> integration/QA.

## Execution
Codex workers implement independent ready nodes. Claude reviewer evaluates PRs. QA agents map evidence to acceptance criteria. Integration Lead validates combined behavior.

## Human release gate
Human sees requirement coverage, CI/QA/review status, outstanding risk, and release evidence—not raw implementation details unless desired.

## Learning
After merge, a recurrence/DST test failure may become a QA memory. A successful recurrence pattern may become a candidate procedure/skill after repeated use.
