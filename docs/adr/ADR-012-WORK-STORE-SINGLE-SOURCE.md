---
id: ADR-012
type: adr
tier: 1
status: proposed
version: 2
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-012 — One Authoritative Work Store, Configurable Per Project

## Context

Principle 3 states that GitHub owns work state, and ADR-001 made GitHub the control plane. The
first draft of this ADR read that as "GitHub issues are the only backlog."

That conflates two separate things:

- **There must be exactly one authoritative store.** This is the real requirement. Two stores
  drift, and a drifting tier-5 store puts execution-state authority in question at exactly the
  moments it matters.
- **Which store it is.** This is a project-level choice, and ADR-001 already anticipated it by
  wrapping GitHub behind an SCM/control-provider interface "to avoid architectural lock-in."

For a bootstrapping project or a one- or two-person team, GitHub issues can cost more ceremony
than they return. This repository is the working example: it has zero issues, and
`IMPLEMENTATION_ROADMAP.md` has been doing the backlog's job. The original rule would have made
that a permanent violation rather than a reasonable choice at this scale.

What has to be constant is not the store. It is the **format**.

## Decision

**Work items have one canonical schema. Exactly one store is authoritative per project. The store
is configurable; the schema is not.**

1. **The schema is the contract.** A work item has the same fields, states, and identity rules
   wherever it lives. `states.yaml` `work_state` governs its lifecycle in every store.
2. **One store, declared.** `work_store` in `.agentic/project.yaml` names the authoritative store.
   Never two.
3. **Two stores supported initially:** `local` (structured files in the repository) and `github`
   (issues). Both implement the same schema.
4. **Migration is one-way and explicit.** Moving from `local` to `github` is an operation with a
   before and an after. There is no bidirectional sync, ever.
5. **The intent/state boundary holds regardless of store.** Documents carry phases, sequencing, and
   exit criteria. The work store carries status, assignee, and progress. A markdown checkbox list
   remains forbidden — not because it is local, but because it is unstructured.

## Alternatives considered

**GitHub issues as the only backlog** (this ADR's first draft). Rejected. It imposes GitHub
ceremony on solo and bootstrap projects, contradicts ADR-001's own control-provider abstraction,
and is currently unsatisfiable in this repository — which is evidence about the rule, not about the
repository.

**Markdown checklists as the local option.** Rejected. A checklist is not a schema. Local storage
is fine; *unstructured* local storage reintroduces drift under another name, since nothing can
validate it and every agent parses it differently.

**Bidirectional sync between local and GitHub.** Rejected outright. Two authoritative stores kept
in agreement by machinery is the precise failure this decision exists to prevent, and it fails
worst under conflict — which is when work state matters most.

**A database as the only store.** Rejected for now: it makes work state invisible to anyone
reading the repository, and unavailable offline. `work_items` remains a projection for querying,
not the source.

## Consequences

- `.agentic/project.yaml` gains a `work_store` key. This repository starts on `local`.
- The local store needs a defined on-disk format and its own validation — **this is the next
  deliverable**, and until it exists neither store is fully specified.
- `devctl` and every agent operate on the schema, never on the store. No agent branches on which
  store is configured; that is the whole point of the abstraction.
- ADR-001 is unaffected. GitHub remains authoritative for code, pull requests, checks, and
  releases. This decision concerns work items only.
- `CLAUDE.md`'s rule that a branch must name its issue becomes satisfiable under either store,
  since both mint stable work-item IDs.

## Risks

- **Two implementations to maintain.** Mitigated by keeping the adapter surface small: the schema
  carries the complexity, the adapter carries only persistence.
- **A local store forfeits GitHub's ecosystem** — notifications, mobile, cross-repo views. Accepted:
  that is the trade a solo project makes knowingly, and migration is available when it stops being
  worth it.
- **Drift moves rather than disappears**, from between-stores to between-schema-versions. Mitigated
  by versioning the work-item schema explicitly, as with every other artifact.
- **A project could switch stores mid-flight and lose history.** Mitigated by migration being an
  explicit operation that must carry existing items across, not a configuration toggle.
