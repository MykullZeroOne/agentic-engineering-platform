---
id: DES-GITHUB-CONTROL-PLANE
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

# GitHub Control Plane

## Authoritative execution objects
- GitHub Issue / sub-issue — work intent/execution unit.
- GitHub Project — portfolio/work-state view.
- Branch/worktree — isolated implementation.
- Pull Request — implementation/evidence package.
- Check/Action — deterministic validation.
- Review — independent evaluation.
- Release — deployable version record.

## Suggested project fields
- Type
- Status
- Priority
- Iteration
- Risk
- Architecture Impact
- Human Gate
- Agent Role
- Runtime Preference
- Agent State
- ADS/Requirement IDs
- Release

## Suggested status machine
Intake -> Refinement -> Ready -> In Progress -> Review -> Validation -> Integration -> Ready to Merge -> Done

Alternative/error states: Blocked, Awaiting Human, Failed, Cancelled.

## Core GitHub principle
AEP may enrich state but does not keep a hidden competing backlog. If work execution state differs, GitHub is authoritative unless explicitly configured otherwise.
