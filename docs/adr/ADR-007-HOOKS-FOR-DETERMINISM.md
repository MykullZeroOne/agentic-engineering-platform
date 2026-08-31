---
id: ADR-007
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
approval_record: APR-0013
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-007 — Hooks and Policies Enforce Lifecycle Behavior

## Decision
Use hooks/policies/checks for mandatory lifecycle actions. Skills guide how agents perform work; hooks determine when required actions happen.

## Examples
Before run: context hydration, dependency validation, worktree creation.
After implementation: tests, diff inspection, evidence capture.
After merge: memory consolidation, dependency recalculation, evaluation record.
