---
id: PRD-003
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

# PRD-003 — Agent Communication, Questions, and Escalation

## Objective
Provide a durable, hierarchical communication plane for questions, clarifications, approvals, findings, and notifications.

## Requirements
- Every agent can ask its parent a structured question.
- A parent may answer from canonical knowledge or memory.
- If a parent lacks authority or sufficient evidence, it may escalate upward.
- Human questions must include reason, impact, blocking status, current context, and suggested choices when useful.
- Questions and answers become durable graph entities and may create decisions.
- Agents should not freely bypass reporting relationships except via explicit cross-functional handoff rules.

## Severity classes
- INFO — no response required.
- QUESTION — response required but non-blocking.
- BLOCKING — work cannot proceed.
- APPROVAL — explicit human/authorized gate required.
- CRITICAL — immediate human attention.

## Acceptance criteria
A Compliance sub-agent can ask BA a blocking question; BA can answer from prior approved product knowledge without human interruption. If it cannot, the same question can be escalated to the human and the answer propagated back and stored as a decision.
