---
id: POL-001
type: policy
tier: 3
status: approved
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-09-08
approval_record: APR-0022
supersedes: null
superseded_by: null
last_reviewed: 2026-09-08
enforcement: prose
kind: policy
scope: project
applies_to:
  work_item_type: [feat, fix, docs, spec, adr, chore, refactor, test, ci]
---

# POL-001 — Branching and Merging

The rules under which change reaches `main`. Extracted from `CLAUDE.md` (WI-0006, WI-0020),
which had carried them since the repository's first commit: tier-3 policy with real force living
in an agent guide, ungated. The guide keeps the reasoning and the history; this artifact holds the
rules. Where the two disagree once this is approved, this wins.

`enforcement: prose` is the honest stage for the artifact as a whole. Each rule below names its
own stage, because the mix is the point: two are `check`, the rest rest on agents reading them.

## Branching

**Trunk-based development on `main`.** `main` is always releasable. There is no `develop` branch,
no long-lived release branch, and no environment branch.

| # | Rule | Stage |
| --- | --- | --- |
| B1 | `main` is protected. No direct commits, no force-push. All change arrives by pull request. | `check` (branch protection) |
| B2 | One work item, one branch, one pull request. A branch that cannot name the item it serves must not exist; work that outgrows its item is split, not widened. | `prose` |
| B3 | Branch from current `main`, never from another feature branch. A dependency on unmerged work is said on the item and waited for, or landed first. | `prose` |
| B4 | Short-lived: under two days and under ~400 changed lines. A branch older than five days is rebased or closed. | `prose` |
| B5 | Agents implement in a dedicated worktree per branch, removed when the pull request merges. | `prose` |
| B6 | A branch is brought up to date by rebase or by merging `main` in, either. Prefer rebase on a branch only you hold; merge `main` in on one someone else has checked out. | `prose` |
| B7 | Force-push is permitted on your own feature branch only, and only with `--force-with-lease`. | `prose` |

**Carve-outs to B2**, both narrow:

- *Approval records batch.* One pull request may carry several approval records closing several
  gates, provided each record names its own artifact and statement.
- *An approval-only change needs no work item.* A pull request changing nothing but approval
  records and the front-matter fields that point at them names the approved artifacts instead. Any
  change to a document's body is ordinary work and needs its item.

### Naming

```
<type>/<work-item-number>-<kebab-summary>
```

`<type>` is the work item's own `type` field, so branch prefix and item type are one token. The
number is the item's. Agent-driven branches may add the role as a suffix: `--backend.engineer`.

## Merging

**Squash merge only.** One work item becomes one commit on `main`. Merge commits and rebase-merge
are disabled. The branch is deleted on merge.

| # | Rule | Stage |
| --- | --- | --- |
| M1 | The pull request names its work item and closes it. `Closes WI-NNNN` lives in the branch's commit message, not only the pull request body, so the squash body carries it to `main`. | `prose` |
| M2 | All required checks pass. A required check is never weakened, skipped, or disabled to make a pull request mergeable. | `check` (required status checks) |
| M3 | The pull request carries its evidence package: what changed, why, and how it was verified, per SPEC-EVIDENCE-PACKAGE. The template is a projection of it. | `prose` |
| M4 | The pull request has an independent review. **Suspended repository-wide** (WI-0015): one human identity means author and reviewer cannot be told apart. The suspension is a standing fact, never a per-pull-request checkbox, never cited as met, and never satisfied by self-approval. Open until WI-0016. | `prose` (suspended) |
| M5 | Every human gate the change triggers has an approval record. An agent never merges a pull request that crosses a human gate. Merging closes no gate (ADR-019); only a record does. | `check` (`check_gates.py --strict`) |
| M6 | The branch is up to date with `main`. | `check` (branch protection) |

The gates in force are `human_gates` in `.agentic/project.yaml`, defined in
`.agentic/registries/gates.yaml`. This policy does not restate them: a list restated in prose is
a second source that drifts, and the registry is the only one that resolves.

The squash commit subject is the pull request title, so the title is a Conventional Commit
subject: `type(scope): summary`, imperative, under 72 characters.

## Reverting

Revert first, diagnose second. Because merges are squashed, `git revert <sha>` on `main` is always
sufficient. The root cause becomes a new work item; `main` is never hot-fixed forward.

## Known gaps

Recorded here so the policy does not claim more than it enforces.

- **M1 has no `check`.** Nothing verifies the trailer before merge, which is why PRs #8 and #9
  landed without one and their items can never derive state.
- **A pull request that files a work item has nothing to close**, and M1 assumes every pull
  request closes one. Filing an item on a branch named for it is self-closing under the merge
  fallback (WI-0039 reopened). Needs a convention decision; none is made here.
- **B4 and B5 are unobserved.** No check reads branch age or worktree use.
