# CLAUDE.md

Guidance for Claude Code (and other agent runtimes) working in this repository.

## What this repository is

The **Agentic Engineering Platform (AEP)** — a vendor-neutral, subscription-first platform that
models a real engineering organization around GitHub, agent skills, deterministic hooks, durable
per-role memory, and interchangeable execution providers (Claude Code, Codex).

**Current state: specification only.** There is no application code yet. Every file here is
documentation, schema, or configuration. Phase 1 of `docs/roadmap/IMPLEMENTATION_ROADMAP.md` has
not started. Do not assume a build system, test suite, or runtime exists — if you need one, it is
new work that needs an issue and an ADR.

## Repository map

| Path | Contents |
| --- | --- |
| `docs/vision/` | Product vision and the ten core principles |
| `docs/prd/` | PRD-001..008, human-readable requirements |
| `docs/ads/` | Agent Development Specification — the machine-readable spec layer |
| `docs/adr/` | ADR-001..008, architecture decisions |
| `docs/architecture/` | System architecture, components, data model, deployment, stack |
| `docs/agents/` | Role organization, the universal agent loop, per-role loops |
| `docs/memory/` | Memory architecture, event model, consolidation, knowledge graph |
| `docs/context/` | Context packet, retrieval, provenance |
| `docs/workflows/` | End-to-end SDLC, policy model, hooks, question escalation |
| `docs/github/` | GitHub control plane, project bootstrap, legacy adoption |
| `docs/security/` | Governance |
| `docs/operations/` | `devctl` and agent workspace UI |
| `docs/schemas/` | Example YAML for ADS, roles, run records, memory, questions |
| `docs/roadmap/` | Phased implementation roadmap |
| `examples/` | Worked idea-to-delivery walkthrough |
| `.agentic/` | Project config (`project.yaml`), lifecycle hooks, role and skill registries |

Start with `docs/DOCUMENTATION_INDEX.md`. It lists the intended reading order.

## Non-negotiable principles

These come from `docs/vision/PRINCIPLES.md` and constrain every change:

1. **Human at the top.** The human is CTO/Product Owner. Product, legal, architecture, security,
   and release decisions are theirs. Never self-approve a human gate.
2. **Role != model.** `compliance.primary` is a durable identity; a Claude session is a transient
   runtime. Never write a document that hard-codes a specific model into a role definition.
3. **GitHub owns work state.** Issues, PRs, checks, and releases are authoritative. Do not create a
   competing backlog in markdown — no `TODO.md`, no checkbox task lists in docs.
4. **Subscription-first.** Prefer supported CLIs over raw APIs. APIs are adapters, not foundations.
5. **Hooks make behavior deterministic.** Behavior that must always happen belongs in
   `.agentic/hooks/hooks.yaml`, not in prose asking an agent to remember.

## Documentation conventions

- Markdown lives under `docs/` in an existing subdirectory, or `examples/`. `README.md` and this
  file are the only permitted root-level markdown documents.
- **PRDs** are `docs/prd/PRD-NNN-SHORT-TITLE.md`. **ADRs** are `docs/adr/ADR-NNN-SHORT-TITLE.md`.
  Numbers are sequential and never reused, even for withdrawn documents.
- ADRs are immutable once merged. To reverse a decision, add a new ADR that supersedes it and add a
  "Superseded by ADR-NNN" line to the original — never edit the original's decision.
- Any new document must be added to `docs/DOCUMENTATION_INDEX.md` in the same commit.
- Keep the existing terse, heading-driven style. These documents are read by agents as context;
  favor short declarative statements over narrative prose.

### Known inconsistency

`.agentic/project.yaml` points at `docs/prds`, `docs/adrs`, and `docs/standards`. The real
directories are `docs/prd/` and `docs/adr/`, and `docs/standards/` does not exist. Do not silently
"fix" one side — decide deliberately which is canonical, and change it in a dedicated PR.

## Branching strategy

**Trunk-based development on `main`.** `main` is always releasable. There is no `develop` branch,
no long-lived release branches, and no environment branches.

### Rules

1. **`main` is protected.** Never commit directly to `main` and never force-push it. All change
   arrives through a pull request.
2. **One issue, one branch, one PR.** A branch that cannot name the issue it serves should not
   exist. If work grows beyond its issue, split it — do not widen the branch.
3. **Branch from current `main`.** Always `git fetch origin && git switch -c <branch> origin/main`.
   Never branch from another feature branch; if you genuinely depend on unmerged work, say so on
   the issue and wait, or land the dependency first.
4. **Short-lived.** Target under two days of work and under ~400 changed lines. A branch older than
   five days must be rebased on `main` or closed.
5. **Isolated worktrees.** Agents implement in a dedicated git worktree per branch
   (`git worktree add ../aep-<issue> -b <branch> origin/main`), so parallel agents never share a
   working directory. Remove the worktree when the PR merges.
6. **Rebase to update, never merge `main` into a branch.** `git fetch origin && git rebase
   origin/main`. This keeps history linear and keeps the PR diff honest.
7. **Force-push is allowed on your own feature branch only**, and only with `--force-with-lease`.

### Naming

```
<type>/<issue-number>-<kebab-summary>
```

Types: `feat`, `fix`, `docs`, `spec`, `adr`, `chore`, `refactor`, `test`, `ci`.

```
docs/42-context-packet-provenance
adr/57-supersede-postgres-first
feat/103-devctl-doctor
```

Agent-driven branches may carry the role as a suffix when several roles touch one issue:
`feat/103-devctl-doctor--backend.engineer`.

## Merging strategy

**Squash merge only.** One issue becomes one commit on `main`. This keeps `main` linear, makes
revert a single operation, and makes the post-merge consolidation hooks
(`.agentic/hooks/hooks.yaml`) reason about one coherent unit of change.

- Merge commits: disabled. Rebase-merge: disabled.
- The squash commit subject is the PR title; the body must contain `Closes #<issue>`.
- Delete the branch on merge.

### Merge requirements

A PR may merge only when all of these hold:

1. It links its issue and closes it.
2. All required checks pass. Never merge with a red or skipped required check, and never weaken or
   disable a check to make a PR mergeable.
3. It carries its **evidence package** — what changed, why, and how it was verified. A PR is an
   evidence artifact, not just a diff (`docs/workflows/END_TO_END_SDLC.md`, Phase 5–6).
4. It has an independent review. The agent that wrote the change never approves it.
5. Every **human gate** the change triggers has explicit human approval. Per
   `.agentic/project.yaml`: product specification, high-risk architecture, legal, and production
   release. **An agent must never merge a PR that crosses a human gate.**
6. It is up to date with `main` (rebased, not merged).

### Change classes that always require human approval

- Any new or superseding ADR (`docs/adr/**`)
- Any change to `.agentic/project.yaml`, `.agentic/hooks/**`, or role definitions
- Anything under `docs/security/**`
- Anything altering a human gate, policy, or required check
- Future: `migrations/**`, per `docs/workflows/POLICY_MODEL.md`

### Reverting

Revert first, diagnose second. Because merges are squashed, `git revert <sha>` on `main` is always
sufficient. Open a follow-up issue for the root cause; do not hot-fix forward on `main`.

## Commits

Conventional Commits: `type(scope): summary`, imperative mood, under 72 characters.

```
docs(context): add provenance fields to context packet
adr(memory): supersede ADR-003 with dual-store decision
```

Individual commits within a branch may be messy — they are squashed. The **PR title** is the
message that lands, so it must be a clean Conventional Commit subject.

Never commit: secrets, tokens, `.env` files, `.DS_Store`, agent scratch output, or run transcripts.

## Working agreement for agents

- **Read before writing.** Load `docs/DOCUMENTATION_INDEX.md` and the documents relevant to your
  change. Contradicting an existing PRD or ADR is a finding to raise, not a thing to quietly do.
- **Ask rather than assume.** Ambiguity in intent escalates per
  `docs/workflows/QUESTION_ESCALATION.md`. Do not invent product requirements.
- **Stay in scope.** Fix what the issue asks. Unrelated problems you notice become new issues.
- **Cite your sources.** When a change follows from a document, name it (`per ADR-002`).
- No code exists yet, so there is nothing to build or test. When code lands, this section gets
  build, test, and lint commands — add them in the same PR that introduces them.
