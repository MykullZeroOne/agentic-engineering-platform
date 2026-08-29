---
id: ADR-009
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-009 — Trunk-Based Development with Squash Merge

## Context

`CLAUDE.md` establishes trunk-based development and squash-only merging as repository rules, but
never as a recorded decision. It is not a stylistic preference: `after_merge` hooks
(`.agentic/hooks/hooks.yaml`) consolidate memory, refresh the graph, and record an evaluation
datapoint, and every one of those assumes it is reasoning about **one coherent unit of change**.
The branching model is therefore load-bearing for the Knowledge and Evaluation planes, and belongs
at tier 1.

## Decision

Trunk-based development on `main`, which is always releasable.

- One issue, one branch, one PR, one commit on `main`.
- Squash merge only. Merge commits and rebase-merge are disabled.
- Branches are short-lived: target under two days and ~400 changed lines.
- Update a branch by rebasing on `main`, never by merging `main` into it.
- `--force-with-lease` on your own feature branch only.
- Revert first, diagnose second: `git revert <sha>` on `main` is always sufficient.

## Alternatives considered

**A long-lived `develop` branch (Git Flow).** Rejected. It makes "`main` is always releasable"
false by construction, and integration order becomes ambiguous — `after_merge` would fire against
a branch that is not the release trunk, so consolidated memory would describe a state no user ever
receives.

**Preserve merge commits.** Rejected. A merge of N commits gives the consolidation hook N diffs of
mixed intent, including work-in-progress states that were later corrected within the same branch.
Memory candidates derived from those describe attempts, not outcomes — the opposite of what
`docs/memory/CONSOLIDATION.md` wants.

**Rebase-merge (linear history, commits preserved).** The closest alternative, and rejected on a
narrower ground: it keeps history linear but still lands multiple commits per issue. Revert stops
being a single operation, and the evaluation record in ADR-004 has no single SHA to attach an
outcome to.

## Consequences

- The **PR title** is the commit message that lands, so it must be a valid Conventional Commit
  subject. Commits within a branch are disposable.
- `after_merge` hooks may assume one commit equals one issue equals one evidence package.
- Revert is always a single operation, which is what makes "revert first, diagnose second"
  affordable.
- A branch older than five days must be rebased on `main` or closed.
- Agents implement in isolated worktrees so parallel work never shares a working directory.

## Risks

- **Intra-branch history is lost at merge.** Accepted: the PR retains it, and per ADR-004 the
  durable causal record is the event store, not git history.
- **A large change becomes one large commit.** Mitigated by the branch size guidance; a branch that
  outgrows its issue should be split, not widened.
