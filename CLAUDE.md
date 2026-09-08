# CLAUDE.md

Guidance for Claude Code (and other agent runtimes) working in this repository.

## What this repository is

The **Agentic Engineering Platform (AEP)** — a vendor-neutral, subscription-first platform that
models a real engineering organization around GitHub, agent skills, deterministic hooks, durable
per-role memory, and interchangeable execution providers (Claude Code, Codex).

**Current state: a specification corpus, the tooling that keeps it honest, and the first thin
slice of the product.** Be precise about which, because an earlier version of this paragraph said
"there is no application code" long after there was, and told agents not to run a test suite that
already existed.

What exists and runs:

- `scripts/` — the documentation validator, the gate check, and the index generator. All three run
  in CI; `validate` is a required check.
- `cmd/devctl` and `internal/` — roughly 2,700 lines of Go with four test files. `devctl doctor`,
  `devctl work list`, `devctl work show`, and `devctl work advance` are real and used. Those four
  are the whole command surface; `docs/operations/DEVCTL.md` documents a much wider one that is
  still `draft` and unbuilt.

What does not exist: any runtime or long-lived service, the hook engine, a database, an event
store, retrieval, agent execution, and the workspace UI. Every binding in `.agentic/hooks/hooks.yaml`
is declarative — nothing executes them. `.agentic/roles/` and `.agentic/skills/` hold a `.gitkeep`
each. Phase 1 of `docs/roadmap/IMPLEMENTATION_ROADMAP.md` is partially begun, not unstarted: the
`devctl` slices landed under WI-0019, WI-0021 and WI-0023, and the rest is decomposed into WI-0042
through WI-0053.

When you write about this system, distinguish **designed** from **implemented** from **proven**.
Most of the corpus is designed. A little is implemented. Only what CI runs is proven.

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
| `scripts/` | Documentation validator, gate check, and index generator |
| `cmd/devctl/` | The `devctl` CLI — the only product binary |
| `internal/` | Go packages behind `devctl`: `approvals`, `config`, `doctor`, `work` |

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

- Markdown lives under `docs/` in an existing subdirectory, or `examples/`. `README.md`, this
  file, and `ISA.md` are the only permitted root-level markdown documents. `ISA.md` is the
  project's Ideal State Artifact — the articulation of what "done" means for AEP as a whole, and
  the source of the claims the corpus is measured against. It sits at the root because it is
  addressed to whoever opens the repository, not to the corpus: it is not a tier 0–3 artifact, it
  carries no authority over any approved document, and nothing may cite it as binding intent.
  Where it and an approved artifact disagree, the approved artifact wins and the ISA is wrong.
  It is also not a backlog, and principle 3 still holds against it: work state lives only in
  `.agentic/work/`. A checked claim in `ISA.md` records that a probe passed and points at the
  proof; it never records that work was done, and the work store wins on any disagreement.
- **Every document carries front matter** declaring `id`, `type`, `tier`, `status`, `version`,
  `owner`, and approval fields, per `docs/spec/DOCUMENT_LIFECYCLE.md`. `README.md`, this file,
  `docs/schemas/*.yaml`, `examples/project.yaml`, and `ISA.md` are exempt. `ISA.md` carries ISA
  front matter instead, per the ISA format spec; it is not a lifecycle-managed document and has no
  `status`, `tier`, or approval fields to set.
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

   **Two carve-outs, both added after a day that produced twenty pull requests against twelve
   hundred lines of product code.**

   *Approval records batch.* One pull request may carry several approval records closing several
   gates, provided each record names its own artifact and statement. Nothing ever required one
   record per pull request; it was a habit, and it cost four pull requests in a single afternoon
   to record three approvals.

   *An approval-only change needs no work item.* When a pull request changes nothing but approval
   records under `.agentic/approvals/` and the front-matter fields pointing at them, it names the
   approved artifacts instead of a work item. An approval record already states what was approved,
   by whom, on what content hash, and in response to what request. A work item saying "accept X"
   adds a second place to look and no fact. This carve-out is narrow on purpose: any change that
   touches a document's *body* is ordinary work and needs its item.
3. **Branch from current `main`.** Always `git fetch origin && git switch -c <branch> origin/main`.
   Never branch from another feature branch; if you genuinely depend on unmerged work, say so on
   the issue and wait, or land the dependency first.
4. **Short-lived.** Target under two days of work and under ~400 changed lines. A branch older than
   five days must be rebased on `main` or closed.
5. **Isolated worktrees.** Agents implement in a dedicated git worktree per branch
   (`git worktree add ../aep-<issue> -b <branch> origin/main`), so parallel agents never share a
   working directory. Remove the worktree when the PR merges.
6. **Update a branch either way — rebase, or merge `main` into it.** Squash collapses the branch
   to one commit, so neither choice reaches `main`: PR #9 merged carrying a `main` merge commit
   and landed as a single-parent squash anyway. The old rule was rebase-only "to keep history
   linear", which is an argument for a merge-commit workflow this repository does not use. Prefer
   rebase on a branch only you hold, because it keeps the PR diff easier to read; merge `main` in
   on a branch someone else has checked out, because rewriting theirs is the greater harm.
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
- The squash commit subject is the PR title. **Put `Closes WI-NNNN` in the branch's commit
  message, not only in the PR body.** GitHub's squash body defaults to the branch's commit
  messages, so a line living only in the pull request description never reaches `main` — PRs #8
  and #9 both merged without one. Setting the repository's squash default to "pull request title
  and description" also works, but a convention that survives a settings toggle is the sturdier
  half, and this repository holds only one of the two.
- Delete the branch on merge. No local git test identifies a merged branch here: squash means the
  branch tip is never an ancestor of `main`, and its tree diverges as soon as `main` moves on.
  GitHub's merge record is the only authority — which is also what `work.advance_state` will have
  to read (`WI-0013`).

### Merge requirements

A PR may merge only when all of these hold:

1. It names its work item and closes it. Work items live in the store named by `work_store`
   (`.agentic/work/` here), and the branch's commit message carries `Closes WI-NNNN` so that the
   squash body carries it onto `main`.
2. All required checks pass. Never merge with a red or skipped required check, and never weaken or
   disable a check to make a PR mergeable.
3. It carries its **evidence package** — what changed, why, and how it was verified. A PR is an
   evidence artifact, not just a diff (`docs/workflows/END_TO_END_SDLC.md`, Phase 5–6).
4. It has an independent review — **suspended repository-wide, not satisfied per pull request.**
   This repository has one human identity and agents act as that identity, so GitHub cannot tell
   author from reviewer. The suspension is a standing fact recorded here; **it is not a checkbox
   on every pull request**, because a box that is false on every change teaches people to skip the
   list rather than read it. Never cite the requirement as met and never treat a self-approval as
   independent. Open risk until `WI-0016` substitutes a reviewing agent or a second identity
   exists.

   An automated adversarial review is not this, and is worth running anyway: two runs on
   2026-09-01 found 38 findings across two documents, 8 of them severity 1, including a
   contradiction with a tier-0 spec approved the same week.
5. Every **human gate** the change triggers has explicit human approval. The gates in force are
   `human_gates` in `.agentic/project.yaml`, defined in `.agentic/registries/gates.yaml`; read
   them there rather than from any list restated in prose. **An agent must never merge a PR that
   crosses a human gate**, and merging closes no gate (ADR-019) — only an approval record does.
6. It is up to date with `main`. How it got there does not matter — see branching rule 6.

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
- **Validate before you push.** Every command must be clean. `validate` and `build` are both
  required checks, so a red one blocks the merge.

  ```
  pip install -r requirements-docs.txt
  python3 scripts/build_index.py                      # regenerate the index
  python3 scripts/validate_docs.py                    # front matter, tiers, registries, index freshness
  python3 scripts/check_gates.py --base origin/main   # which human gates the change crosses
  go build ./... && go test ./...                     # required whenever cmd/ or internal/ changes
  ```

  Run the Go commands on any change under `cmd/` or `internal/`. The previous version of this
  section said no product code existed and listed only the Python commands, which meant an agent
  following it literally skipped a test suite that was already there.
