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

Recorded in `docs/spec/REGISTRIES.md` under "Known gaps", though that entry is now partly stale.
Tier 3 has one artifact, POL-001, in `docs/policies/`; `docs/standards/` is still empty and
materializes with STD-001. Skills still have no schema and escalation still has no hook point. Do
not invent either in passing — each needs a decision.

## Branching and merging

`docs/policies/POL-001-BRANCHING-AND-MERGING.md` is approved (APR-0022) and holds the rules. This
guide does not restate them; where the two disagree, POL-001 wins. Rule ids below (B1-B7, M1-M6)
resolve against POL-001. What follows is the reasoning and incident history POL-001 deliberately
leaves out.

**Trunk-based development on `main`.** `main` is always releasable, protected (B1), and updated
only by pull request.

- **B2's two carve-outs** (approval records batching, and an approval-only change needing no work
  item) exist because a single day once produced twenty pull requests against twelve hundred lines
  of product code, four of them just to record three approvals under a one-work-item-per-PR rule
  with no exception for records.
- **B6** (rebase or merge `main` in, either is fine) replaced a rebase-only rule after PR #9
  merged carrying a `main` merge commit and still landed as a single-parent squash: proof that
  "rebase to keep history linear" was an argument for a merge-commit workflow this repository does
  not run.
- No local git test identifies a merged branch. Squash means the branch tip is never an ancestor
  of `main`; GitHub's merge record is the only authority for that fact, the one
  `work.advance_state` will have to read (WI-0013).

### Naming

B2's naming convention, from POL-001:

```
<type>/<work-item-number>-<kebab-summary>
```

The number is the work item's, from `.agentic/work/` (`WI-0007` gives `7`). `<type>` is the work
item's own `type` field, so the branch prefix and the item's type are one token, not two that
drift.

```
docs/42-context-packet-provenance
adr/57-supersede-postgres-first
feat/103-devctl-doctor
```

Agent-driven branches may carry the role as a suffix when several roles touch one issue:
`feat/103-devctl-doctor--backend.engineer`.

**Squash merge only.** One work item becomes one commit on `main`; merge commits and rebase-merge
are both disabled.

- **M1** (the `Closes WI-NNNN` trailer belongs in the branch's commit message, not only the PR
  body) exists because GitHub's squash body defaults to the branch's commit messages. PRs #8 and
  #9 both merged without the trailer and reached `main` unlinked to their work item.
- **M4** (independent review) is suspended repository-wide, not satisfied per pull request: one
  human identity means GitHub cannot tell author from reviewer. Never cite it as met, never treat
  self-approval as independent; open until WI-0016. An automated adversarial review is not this
  requirement but is worth running anyway: two runs on 2026-09-01 found 38 findings across two
  documents, 8 of them severity 1, including a contradiction with a tier-0 spec approved the same
  week.

Gates in force are `human_gates` in `.agentic/project.yaml`, defined in
`.agentic/registries/gates.yaml`; read them there, not from any list restated in prose.

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

**Revert first, diagnose second.** Because merges are squashed, `git revert <sha>` on `main` is
always sufficient; open a follow-up issue for the root cause rather than hot-fixing forward on
`main`.

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
