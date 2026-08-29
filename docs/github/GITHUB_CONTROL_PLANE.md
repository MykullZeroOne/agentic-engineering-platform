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
