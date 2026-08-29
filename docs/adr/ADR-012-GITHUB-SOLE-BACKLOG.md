---
id: ADR-012
type: adr
tier: 1
status: proposed
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-012 — GitHub Issues Are the Only Backlog

## Context

Principle 3 states that GitHub owns work state, and `CLAUDE.md` forbids `TODO.md` and checkbox task
lists in documentation. ADR-001 made GitHub the authoritative control plane. Despite all three,
this repository currently has **zero issues**, and `docs/roadmap/IMPLEMENTATION_ROADMAP.md` is
functioning as the de facto backlog — so the rule is already being violated by the only planning
artifact that exists. The rule also has an unstated boundary: it is not obvious which parts of a
roadmap are forbidden state and which are legitimate intent.

## Decision

GitHub issues are the only backlog. There is no parallel tracker, no `TODO.md`, no checkbox task
list in documentation, and no markdown file that carries work status.

The boundary between a document and the backlog is **intent versus state**:

| Belongs in documentation | Belongs in GitHub |
| --- | --- |
| Phases, sequencing, exit criteria | Status, assignee, progress |
| Why work is ordered this way | Whether work has started or finished |
| What "done" means for a phase | Whether a given item is done |

Anything carrying a status, an assignee, or a checkbox is state and belongs in GitHub. AEP may
enrich and project work state, but never hold a competing one — where they differ, GitHub wins.

## Alternatives considered

**A markdown backlog in the repository.** Rejected. It diverges from GitHub the moment either
changes, which creates a second source of truth for tier 5 — precisely the authority ADR-001
assigned to GitHub and `docs/spec/AUTHORITY_MODEL.md` scopes to execution state.

**An external tracker (Jira, Linear).** Rejected for now, not on principle. ADR-001's rationale —
open semantics, issue/PR/check traceability in one place, broad agent compatibility — still holds,
and the SCM/control-provider interface keeps this reversible. Revisit only by superseding.

**Both, kept in sync.** Rejected. Bidirectional sync between two authoritative stores is a
well-known failure mode, and it would put tier-5 authority in question at exactly the moments it
matters most: a conflict during a release gate.

## Consequences

- `IMPLEMENTATION_ROADMAP.md` must be converted into issues. Its phase structure and exit criteria
  stay as prose; anything resembling task tracking does not.
- `CLAUDE.md`'s rule that a branch must name its issue becomes satisfiable — today it is not, since
  no issues exist.
- `work.recalculate_ready` (`after_merge`) reads GitHub, not a file.
- Agents obtain work state exclusively through the GitHub adapter.

## Risks

- **GitHub issues express dependency graphs poorly.** Mitigated by sub-issues plus the
  `work_dependencies` table as a *projection*, never as an alternative source of truth. If that
  proves insufficient, the fix is a richer projection, not a second backlog.
- **A GitHub outage blocks work-state transitions.** Accepted; ADR-001 already took this risk, and
  this decision does not add to it.
