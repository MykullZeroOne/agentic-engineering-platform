# CLAUDE.md

Guidance for Claude Code (and other agent runtimes) working in this repository.

## What this repository is

The **Agentic Engineering Platform (AEP)** — a vendor-neutral, subscription-first platform that
models a real engineering organization around GitHub, agent skills, deterministic hooks, durable
per-role memory, and interchangeable execution providers (Claude Code, Codex).

**Current state: specification, plus the tooling that keeps it honest.** There is no application
code and no runtime. Phase 1 of `docs/roadmap/IMPLEMENTATION_ROADMAP.md` has not started. The only
executable code is `scripts/`, which validates the documentation corpus. Do not assume a build
system or product test suite exists — if you need one, it is new work that needs an issue and an
ADR.

AEP is the system being built here, not a system this repository runs on. It still dogfoods its
own model in principle: `.agentic/` is this repository's real configuration and the reference
instance of the schema the first runtime must read unchanged.

## Repository map

| Path | Contents |
| --- | --- |
| `docs/spec/` | Authority model, document lifecycle, registry contracts |
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
| `.agentic/` | This repository's config, hook bindings, and vocabulary registries |
| `.agentic/work/` | The local work store — one YAML file per work item |
| `.agentic/approvals/` | Approval records; the provenance behind every closed gate |
| `scripts/` | Documentation validator and index generator |

Start with `docs/DOCUMENTATION_INDEX.md`. It is generated from document front matter and lists the
intended reading order.

## Non-negotiable principles

These come from `docs/vision/PRINCIPLES.md` and constrain every change:

1. **Human at the top.** The human is CTO/Product Owner. Product, legal, architecture, security,
   and release decisions are theirs. Never self-approve a human gate.
2. **Role != model.** `compliance.primary` is a durable identity; a Claude session is a transient
   runtime. Never write a document that hard-codes a specific model into a role definition.
3. **One store owns work state.** Exactly one store is authoritative, named by `work_store`
   (ADR-012); this repository uses `local`, at `.agentic/work/`. GitHub remains authoritative for
   PRs, checks and releases. Do not create a competing backlog in markdown — no `TODO.md`, no
   checkbox task lists in docs.
4. **Subscription-first.** Prefer supported CLIs over raw APIs. APIs are adapters, not foundations.
5. **Hooks make behavior deterministic.** Behavior that must always happen belongs in
   `.agentic/hooks/hooks.yaml`, not in prose asking an agent to remember.

## Documentation conventions

- Markdown lives under `docs/` in an existing subdirectory, or `examples/`. `README.md` and this
  file are the only permitted root-level markdown documents.
- **Every document carries front matter** declaring `id`, `type`, `tier`, `status`, `version`,
  `owner`, and approval fields, per `docs/spec/DOCUMENT_LIFECYCLE.md`. `README.md`, this file,
  `docs/schemas/*.yaml`, and `examples/project.yaml` are exempt.
- **Only `approved` and `accepted` artifacts carry authority.** A `draft` document may guide your
  work but must never be cited as binding, and a drafted requirement is not approved intent. See
  `docs/spec/AUTHORITY_MODEL.md`.
- **Never close a human gate.** You may *record* a human's approval by writing an approval record
  under `.agentic/approvals/` and pointing `human_approved` / `approved_by` / `approved_on` at it
  (ADR-013). Setting those fields without a record, or inferring approval from anything short of an
  explicit instruction naming what is approved, violates Principle 1.
- **New ADRs need `Context`, `Decision`, `Alternatives considered`, `Consequences`, and `Risks`**
  from ADR-009 onward. Enforced by the validator.
- **PRDs** are `docs/prd/PRD-NNN-SHORT-TITLE.md`. **ADRs** are `docs/adr/ADR-NNN-SHORT-TITLE.md`.
  **ADS** are `docs/ads/ADS-NNN-SHORT-TITLE.yaml` on an *independent* sequence, linked to their
  origin by `source_prd`. Numbers are sequential within their type and never reused, even for
  withdrawn documents.
- ADRs are immutable once merged. To reverse a decision, add a new ADR that supersedes it and add a
  "Superseded by ADR-NNN" line to the original — never edit the original's decision.
- `docs/DOCUMENTATION_INDEX.md` is **generated**. Do not edit it by hand — run
  `python3 scripts/build_index.py` and commit the result in the same commit as the new document.
- **Gate IDs, hook points, and state tokens come from `.agentic/registries/`.** Those registries
  are the sole source for that vocabulary; never coin a new gate name, hook point, or status in
  prose. See `docs/spec/REGISTRIES.md`.
- Keep the existing terse, heading-driven style. These documents are read by agents as context;
  favor short declarative statements over narrative prose.

### Open gaps

Recorded in `docs/spec/REGISTRIES.md` under "Known gaps": tier 3 (`docs/policies/`,
`docs/standards/`) has no artifacts, skills have no schema, the evidence package has no defined
structure, and escalation has no hook point. Do not invent any of these in passing — each needs a
decision.

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
<type>/<work-item-number>-<kebab-summary>
```

The number is the work item's, from `.agentic/work/` — `WI-0007` gives `7`. `<type>` is the work
item's own `type` field, so the branch prefix and the item's type are one token, not two that
drift.

Types: `feat`, `fix`, `docs`, `spec`, `adr`, `chore`, `refactor`, `test`, `ci`.

```
docs/42-context-packet-provenance
adr/57-supersede-postgres-first
feat/103-devctl-doctor
```

Work items are structured data, not prose: see `docs/spec/LOCAL_WORK_STORE.md` for the format and
`.agentic/registries/states.yaml` for the `work_state` axis.

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

1. It names its work item and closes it. Work items live in the store named by `work_store`
   (`.agentic/work/` here); the squash commit body carries `Closes WI-NNNN`.
2. All required checks pass. Never merge with a red or skipped required check, and never weaken or
   disable a check to make a PR mergeable.
3. It carries its **evidence package** — what changed, why, and how it was verified. A PR is an
   evidence artifact, not just a diff (`docs/workflows/END_TO_END_SDLC.md`, Phase 5–6).
4. It has an independent review — **suspended, not satisfied.** This repository has one human
   identity, and agents act as that identity, so GitHub cannot tell author from reviewer and no
   branch protection setting enforces this. Never cite this requirement as met, and never treat a
   self-approval as independent. It is an open risk, alongside the bus factor ADR-014 already
   accepts, until `WI-0016` substitutes a reviewing agent or a second identity exists.
5. Every **human gate** the change triggers has explicit human approval. The gates in force are
   `human_gates` in `.agentic/project.yaml`, defined in `.agentic/registries/gates.yaml`; read
   them there rather than from any list restated in prose. **An agent must never merge a PR that
   crosses a human gate**, and merging closes no gate (ADR-019) — only an approval record does.
6. It is up to date with `main` (rebased, not merged).

### Change classes that always require human approval

Gate IDs below resolve against `.agentic/registries/gates.yaml`, which holds the triggers,
approvers, and prior aliases for each.

| Change | Gate |
| --- | --- |
| Any new or superseding ADR (`docs/adr/**`) | `architecture_decision` |
| `.agentic/project.yaml`, `registries/**`, `hooks/**`, `roles/**` | `platform_config` |
| Anything under `docs/security/**` | `security_policy` |
| Approving a PRD or ADS (`docs/prd/**`, `docs/ads/**`) | `product_spec` |
| Anything altering a gate, policy, or required check | `platform_config` |
| Future: `migrations/**`, per `docs/workflows/POLICY_MODEL.md` | `destructive_data_change` |

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
- **Validate before you push.** Both commands must be clean; `docs` is a required check.

  ```
  pip install -r requirements-docs.txt
  python3 scripts/build_index.py     # regenerate the index
  python3 scripts/validate_docs.py   # front matter, tiers, registries, index freshness
  ```

- No product code exists yet. When it lands, this section gets its build, test, and lint commands
  — add them in the same PR that introduces them.
