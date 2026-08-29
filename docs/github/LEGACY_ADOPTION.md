# Existing Repository Adoption

## `devctl adopt`
Scans:
- language/framework/build system;
- repository layout;
- existing docs/ADRs/specs;
- AGENTS/CLAUDE instructions;
- CI/actions;
- test systems;
- open issues/projects;
- PR/review history;
- branch conventions.

Produces an adoption report and non-destructive patch proposal.

## Directory mappings
Legacy repositories can map arbitrary existing directories into canonical knowledge categories rather than moving files immediately.

## Adoption maturity
0 Legacy
1 Agent-aware — identity/instructions/context discovery
2 Knowledge-aware — docs/ADS indexing
3 Process-aware — hooks/policies/shared skills
4 Memory-aware — session continuity/consolidation
5 Agentic — dispatch, loops, QA/review, automated DAG progression

## Historical learning
Optional `--learn-history` creates candidate lessons/patterns from high-value historical PRs: reverts, bug fixes, architecture changes, high-review threads, and major refactors. Candidates require validation before promotion.
