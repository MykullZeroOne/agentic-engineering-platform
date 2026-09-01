---
id: ADR-021
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-09-01
approval_record: APR-0017
supersedes: null
superseded_by: null
last_reviewed: 2026-09-01
---

# ADR-021 — Port Kairo's Provenance and Knowledge-Index Layer

## Context

AEP has a hole it has already referenced. `docs/schemas/run-record.example.yaml:12` carries
`context_manifest: CTX-99182`, and no AEP document defines what a context manifest is. The
identifier is cited and the artifact does not exist.

Kairo (`github.com/MykullZeroOne/kairo`, MIT, same owner, `v1.5.0`) defines it:
`schemas/kairo.context.v1.json` requires `schema`, `context_id`, `created_at`, `entries`,
with `base_commit` optional; `pkg/manifest/manifest.go:15` is the Go type. It also carries a
knowledge index (`pkg/knowledge/engine.go:15`) whose method set — `IndexCheckpoint`,
`IndexSession`, `IndexEvent`, `Search`, `Related`, `Traverse`, `Explain`, `HydrateContext` —
covers much of roadmap Phase 3, and a canonical content hash (`pkg/corehash/hash.go:15`)
covering much of ADR-016.

296 Go files, 91 test files. Schemas are `v1`; the module is tagged `v1.5.0`.

**Kairo has no roles, human gates, approval records, work store, hook policy binding, or
document authority model.** Those are AEP's and this decision does not move them.

### Three claimed overlaps do not hold

Verified against the code, not the README:

1. **`pkg/adapters` does not cover AEP's execution adapters.** Kairo's `Adapter` interface
   (`pkg/adapters/types.go:19`) is `Name`, `Detect`, `Capabilities`, and its capabilities are
   `handoff` and `capture` only (`types.go:14-16`). `pkg/adapters/claude/adapter.go` and
   `pkg/adapters/codex/` contain no `exec.Command`: they detect a repo, capture session
   metadata, and format handoff text. ADR-008 adapters *run the work*. Two different things
   sharing a word.

2. **`HydrateContext` is not the PRD-006 context packet.** `ContextBundle`
   (`pkg/knowledge/types.go:104`) returns checkpoints, manifests, events and pointers —
   runtime provenance. PRD-006 requires role contract, approved PRD/ADS requirements,
   applicable ADRs and policies, dependency state, and acceptance criteria, and says
   "canonical knowledge first". Kairo supplies none of that half.

3. **`pkg/corehash` is adjacent to ADR-016, not conformant with it.** See the reconciliation
   below.

### Lineage: this is consolidation, not adoption

Kairo was the first attempt at the problem AEP is now solving — giving agents memory across
sessions, rewind snapshots, and a knowledge graph so that what one session learns is available
to the next. AEP is the next iteration of the same intent, with a wider scope: roles, human
gates, an authority model, and an execution plane Kairo never had.

That reframes what this ADR decides. It is not a project taking a dependency on an outside
library; it is one attempt at a problem folding its working parts into its successor. The
schemas, the index and the git-provenance layer are the parts of the first attempt that
survived contact with use, and porting them is how their evidence carries forward.

It also explains the shape of the disposition tables. Where Kairo and AEP overlap, Kairo is
usually the narrower and earlier form — its `Adapter` captures where AEP's executes, its
`HydrateContext` returns provenance where AEP's context packet must return canonical knowledge
first, its `corehash` sorts keys where ADR-016 specifies RFC 8785. Those are not defects in
Kairo. They are the earlier iteration, and where AEP has since decided something more
demanding, **AEP's decision governs and the ported code adapts to it.** That rule resolves
every overlap below without needing to argue each one.

### Ownership changes the calculus

Kairo and AEP have the same owner. That is not a licensing footnote; it removes the usual
reason to prefer an upstream dependency over a port. There is no third party whose release
cadence must be respected, no API contract owed to strangers, and no cost to diverging except
the one AEP chooses to pay.

It also dissolves what looked like a blocker. `go.mod:1` declares `module kairo` and every
internal import is `kairo/pkg/...` (`pkg/knowledge/engine.go:5-8`), so Kairo is not importable
as `github.com/MykullZeroOne/kairo/...` today. Under a dependency model that is a condition
precedent owned by another repository. Under a port it is a find-and-replace.

## Decision

**Port selected Kairo packages into AEP under AEP's own module, adapting them to AEP's
configuration and storage. Do not take a dependency on `github.com/MykullZeroOne/kairo`.**

1. **AEP owns the ported code outright.** It moves under `internal/` on AEP's module path,
   is reformatted to AEP's conventions, and is thereafter AEP's to change without reference
   to Kairo. MIT with the same owner makes this unambiguous; the Kairo copyright notice is
   retained in the ported files.

2. **Porting is selective and stated per package.** "All of Kairo" is not the decision; the
   table below is. A package not listed is not ported.

3. **`kairo.*.v1` schemas are adopted as-is and renamed in place.** The JSON schemas move to
   `docs/schemas/` and keep their field shapes. Changing field names during a port would
   throw away the one thing that is already stable and force a migration on day one.

4. **`.kairo/` configuration is not ported.** `pkg/config` reads `.kairo/config.yaml`; AEP
   reads `.agentic/` per its own registries. The ported packages take configuration through
   AEP types, which is the main adaptation cost.

5. **Postgres is the index backend, per ADR-003 and ADR-011.** `knowledge.Engine` is ported
   as an interface; `pkg/knowledge/sqlite` is ported as a reference implementation and a test
   target, not as AEP's store.

### Package disposition

Surveyed by reading each package, not its README. Line counts exclude tests.

| Package | LOC | Tests | Disposition | Why |
| --- | --- | --- | --- | --- |
| `knowledge` | 4248 | 14 | **Port** | The `Engine` interface and its sqlite/memory/vector/search backends. Covers most of roadmap Phase 3. |
| `handoff` | 1249 | 5 | **Port** | Restore plans, resume, PR summaries. Maps onto AEP's **evidence package**, which `docs/spec/REGISTRIES.md` lists as an open gap with no defined structure. |
| `storage` | 869 | 5 | **Port** | Content-addressed object store, local and S3, with GC. Evidence and manifests need somewhere to put blobs. |
| `integrations` | 858 | 5 | **Port** | GitHub, Jira, Confluence, Bitbucket clients. ADR-015 names Confluence as a document store provider; ADR-018 wants a GitHub adapter. |
| `adapters` | 760 | 5 | **Port, renamed** | Capture and handoff formatting. Must **not** be called `adapters` in AEP: ADR-008 adapters execute work and these do not (see finding 1). Port as `provenance/capture`. |
| `checkpoint` | 709 | 3 | **Port** | Core artifact type. |
| `migrate` | 621 | 2 | **Port** | Artifact and config migration. Needed the moment a ported schema revises. |
| `plugin` | 538 | 1 | **Port, gated** | See the ADR-020 finding below. Not wired to hooks without a separate decision. |
| `gitref` | 490 | 4 | **Port, high value** | Not "worktree helpers" as its doc comment says. Git **notes**, commit **trailers**, and `git://` content **pointers**. See finding 5. |
| `security` | 384 | 3 | **Port** | Secret scanning, redaction, object encryption, plugin-spec validation. Fills a **named AEP gap**: `.agentic/hooks/hooks.yaml` lists `pre_commit` as unbound with the comment "CLAUDE.md forbids committing secrets; no check enforces it". |
| `schema` | 382 | 3 | **Port** | Validates artifacts against the shipped schemas. |
| `session` | 212 | **0** | **Port, test first** | Core artifact type with no test file. Porting untested code that other ported packages depend on is how a port becomes a liability. |
| `manifest` | 132 | 1 | **Port** | The context manifest. The reason this ADR exists. |
| `embedding` | 90 | 2 | **Port** | Pluggable embedding providers; pgvector needs one. |
| `corehash` | 70 | 1 | **Port, not as `content_hash`** | See ADR-016 reconciliation. |
| `core` | 44 | 0 | **Rewrite** | `.kairo` layout constants. AEP's layout is `.agentic/`. |
| `config` | 617 | 2 | **Do not port** | Reads `.kairo/config.yaml`. AEP configuration is `.agentic/` and already has a Go reader. |
| `doctor` | 409 | 1 | **Do not port** | AEP already shipped `internal/doctor` in WI-0019. Two doctors is worse than one. |
| `audit` | 413 | 1 | **Port** | Compliance provenance export. Evaluation plane, ADR-010. |
| `hostedsync` | 1786 | 11 | **Do not port** | Multi-tenant sync. |
| `hostedtenancy` | 449 | 1 | **Do not port** | Tenancy model. |
| `hostedlogin` | 199 | 1 | **Do not port** | Hosted auth. |
| `hostedcredentials`, `hostedmigrate`, `hostedid`, `hostedrbac` | 182 | 4 | **Do not port** | `hostedrbac` is a 16-line stub. |
| `version` | 6 | 0 | **Do not port** | Kairo's version string. |

### Outside `pkg/`

The first survey covered `pkg/` only. The rest of the repository:

| Path | Size | Disposition | Why |
| --- | --- | --- | --- |
| `schemas/*.json` | 8 files | **Port** | The artifact contracts. Clause 3. |
| `pkg/knowledge/hook` | — | **Port** | `OnCheckpoint`/`OnSession`/`OnEvent` compiled-in callbacks (`internal/cli/hooks.go:14-36`). This is the same registry model ADR-020 chose, which is corroboration rather than conflict. |
| `testdata/golden` | 39 files | **Port with the code** | Golden files. Porting implementations without their fixtures discards the tests' evidence. |
| `internal/cli` | 2690 | **Reference, do not port** | 40 subcommands. Useful as a map of what the layer needs to expose; AEP has its own `devctl` shape and grafting Kairo's on would fight it. |
| `migrations/` | 17 files | **Do not port** | SQL for hosted identity, projects and canonical records. Belongs to the `hosted*` family. |
| `apps/hosted` | 5928 files | **Do not port** | The hosted web application. |
| `cmd/`, `internal/app` | ~700 | **Do not port** | Kairo's binary wiring. |
| `.goreleaser.yaml`, `Makefile`, `.github/workflows` | — | **Do not port** | Release tooling for a different product. AEP may want goreleaser eventually; that is a separate decision, not a port. |
| `skills/`, `adapters/*.md`, `templates/` | — | **Do not port** | Kairo's own agent-skill and adapter docs. AEP has its own model. |
| `docs/v1`, `docs/v2` | 161 files | **Reference, do not port** | Kairo's specs. Cited where they explain a ported behaviour; they are not AEP documents and would not survive `validate_docs.py`. |

The `hosted*` family is ~2.6k lines of multi-tenant SaaS. AEP has no tenancy concept: ADR-010
decomposes responsibility into seven planes with no tenant boundary, and `.agentic/project.yaml`
describes exactly one project. Porting it would import an unowned abstraction.

## Reconciliation with accepted decisions

| Decision | Relationship | Contradicted? |
| --- | --- | --- |
| **ADR-003** Postgres + pgvector | Kairo's default index is SQLite. Clause 5 ports `Engine` as an interface and supplies a Postgres backend. | **No — only under clause 5.** Shipping AEP on the ported SQLite engine would contradict it and require superseding ADR-003. |
| **ADR-008** subscription-first adapters | No overlap. Kairo's adapters capture and format; they do not execute. Ported under a different name so the collision does not enter the codebase. | No. |
| **ADR-010** seven planes | Ported code lands in **Knowledge**, with `audit` in Evaluation and `integrations` in Control. Rule 1 ("no plane owns another plane's state") means the port cannot be one package: it splits along plane lines. | No, but it constrains the port's shape. |
| **ADR-011** Go modular monolith | A port is *more* compatible than a dependency: one module, one build, one package per plane. ADR-011 also restates Postgres as the single store — same condition as ADR-003. | **No — only under clause 5.** |
| **ADR-012** one work store | Untouched. Kairo has no work item concept. | No. |
| **ADR-015** pluggable knowledge stores | **Kairo is not a store provider and is not eligible to be one.** ADR-015 governs *documents* carrying the `DOCUMENT_LIFECYCLE` schema — `tier`, `status`, approval fields. Checkpoints and sessions have none and are a different artifact class. Separately, `pkg/integrations/confluence` is a plausible *implementation* of an ADR-015 provider. | No. |
| **ADR-016** canonical content hashing | **Partial. `corehash` is not `canonical/v2`.** See below. | **Narrowed** — provided AEP does not claim `corehash` satisfies ADR-016. |
| **ADR-018** GitHub App adapter | `pkg/integrations/github` is a REST client (`client.go`, `remote.go`, `origin.go`), not an App with installation-scoped permissions. It does not deliver ADR-018; it is a starting point below it. | No. |
| **ADR-020** hook engine *(proposed, PR #19)* | **Direct tension — see finding 4.** | Not yet, because ADR-020 is not accepted. |
| **PRD-004** knowledge and memory | Supplies episodic material and the graph. Supplies nothing for canonical knowledge, semantic memory, or "approved decisions outrank learned memories". | No. |
| **PRD-006** context retrieval | Partial, and the smaller half. Finding 2. | No. |
| **PRD-008** portability and adoption | A port adds no external dependency, so `init`/`adopt` are unaffected. | No. |

### Finding 5: git notes, commit trailers, and content pointers

`pkg/gitref`'s package comment says "Git worktree helpers by shelling out to the git CLI",
which undersells it by a wide margin. Four capabilities AEP has needed and not had:

**Git notes** (`pkg/gitref/notes.go`). `NotesRef = "kairo"`, so notes are written to
`refs/notes/kairo` via `git notes --ref=<ref> add -f`. `NotePayload` carries
`CheckpointID`, `SessionID`, `ContextID`, with `AddNote`, `ShowNote`, `NoteExists`,
`ParseNote`, `ListNoteCommits`, `LoadNotePayload`.

This matters to AEP more than it matters to Kairo. **Squash merge destroys the link between
branch work and the trunk commit** — `CLAUDE.md` already records that "no local git test
identifies a merged branch here: squash means the branch tip is never an ancestor of `main`."
AEP's answer so far is a `Closes WI-NNNN` trailer, which works only because someone remembers
to write it: PRs #8 and #9 merged without one. A note can be attached to the merge commit
*after the fact*, without rewriting it, which is the one thing a trailer cannot do.

**Commit trailers** (`pkg/gitref/trailers.go`). `Agent-Session`, `Agent-Checkpoint`,
`Agent-Context`, with `ParseCommitTrailers`, `LoadCommitTrailers`, and a typed
`ErrTrailersNotFound`. AEP already parses one trailer — `internal/work/advance.go` holds a
hand-written `Closes WI-\d{4}` regex from WI-0023. This generalizes it and adds the loading
and error handling that regex does not have.

**Content pointers** (`pkg/gitref/pointer.go`). `git://<repo>@<commit>:<path>`, with
`ParsePointer`, `ResolvePointer`, `ValidatePointer`. This is what a context manifest entry's
`pointer` field (`pkg/manifest/manifest.go:31`) actually resolves against: a reference into
git that names a repository, a commit and a path, and can be checked for existence. AEP's
`run_record` has nothing equivalent.

**Revision helpers** (`pkg/gitref/revisions.go`). `MergeBase`, `RevList`,
`DefaultBaseBranch` — the primitives `check_gates.py` currently shells out for in Python.

**No `.gitattributes` use.** Searched: Kairo touches notes and trailers, not attributes.

#### This bears on ADR-020's central constraint

ADR-020 (proposed) is built around the premise that a hook cannot record anything after a
merge, because `CLAUDE.md` rule 1 forbids committing to `main`. Rule 1's text is scoped
precisely: *"Never commit directly to `main` and never force-push it."*

**A git note is not a commit on `main`.** It lives in `refs/notes/*`, a separate namespace
that branch protection on `main` does not cover. So an `after_merge` hook could attach
provenance to the merge commit without touching `main` and without violating rule 1 as
written.

This ADR does **not** act on that. It is a finding about ADR-020, which is proposed and not
accepted, and quietly widening what a hook may write by porting a library is the same
reinterpretation finding 4 refuses. It is recorded here so the option is visible when ADR-020
is decided.

### Finding 4: `pkg/plugin` versus ADR-020

`pkg/plugin/invoke.go:42` runs `exec.CommandContext(ctx, spec.Command, spec.Args...)` — a
plugin host that executes external binaries named in configuration, over a JSON protocol
(`pkg/plugin/protocol.go:9-14`).

ADR-020 (proposed) clause 2 says `hooks.yaml` gains no field naming a command to execute, and
its acceptance criterion 12 requires that no command or path from `hooks.yaml` is ever
executed. Wiring `pkg/plugin` into the hook engine would contradict that directly.

**But ADR-020 rejected the shell-command alternative on security surface, and Kairo already
built the mitigation ADR-020 assumed did not exist.** `pkg/security/plugin.go` provides
`ValidatePluginSpec`, `AuditPluginSpec`, `validatePluginCommand` and `BuildPluginEnvironment` —
spec validation, an audit finding type, and a controlled environment.

This ADR does not resolve that. It ports `pkg/plugin` and `pkg/security` and **leaves the
plugin host unwired**, because whether hooks may execute external commands is ADR-020's
question and reversing it silently through a port is exactly the "quietly reinterpret an ADR"
move this repository forbids. If the mitigation changes the answer, that is an amendment to
ADR-020 before it is accepted, not a side effect of this one.

### ADR-016 in detail

`pkg/corehash/hash.go:15` computes `sha256:<hex>` over canonical JSON with `content_hash`
excluded. ADR-016 clause 1 specifies RFC 8785. Four differences, all real:

- **HTML escaping.** `MarshalCanonical` (`hash.go:32`) sorts keys and calls `json.Marshal`,
  which escapes `<`, `>` and `&` to their `\u003c` / `\u003e` / `\u0026` forms by default.
  RFC 8785 does not. Any artifact containing those bytes hashes differently under the two rules.
- **No number canonicalization.** RFC 8785 specifies ECMAScript number serialization;
  `normalize` (`hash.go:40`) passes numbers through untouched.
- **No prose-leaf normalization.** ADR-016 clause 2 requires NFC, LF and collapsed whitespace
  on prose fields. `corehash` has none.
- **Version prefix.** Kairo emits `sha256:`; ADR-016 clause 4 requires `canonical/v2` to prefix
  every hash, because the algorithm is versioned.

So `corehash` is the right shape and the wrong specification. It is ported as the hash for
*provenance artifacts* and is **not** adopted as AEP's document `content_hash`. Because AEP now
owns the code, closing those four gaps to make it RFC 8785 conformant is available and would
resolve ADR-016's dependency on ADR-017 structured base data (WI-0008). That is a separate
change with its own gate.

### Finding 6: the branches, and where Kairo's momentum actually is

222 branches. The checked-out head is `feat/217-projection-drain`, one commit ahead of
`master` and **56 commits ahead of the `v1.5.0` tag**. The unmerged work is
`feat/215-event-spec` and `feat/216-event-ingestion`.

Every one of those recent commits is hosted-platform work: outbox projection, presigned
object upload, sync push, event ingestion, `apps/hosted/openapi`, `migrations/0009_events_*`.

**Kairo's active development is entirely in the half this ADR excludes.** The packages AEP
wants — `knowledge`, `gitref`, `manifest`, `checkpoint`, `session` — sit in the stable half
that has barely moved since `v1.5.0`.

That cuts both ways and both ways favour the port. Porting stable code is the lower-risk
version of porting. And the divergence risk in the Risks section is smaller than it first
looks, because staying in sync with upstream would mostly mean syncing changes to a hosted
SaaS AEP is not building.

### Finding 7: `docs/v2/` is the most valuable thing in the repository, and it is not code

63 markdown files across 28 numbered sections: vision, philosophy, PRD, SRS, domain,
architecture, knowledge engine, memory, events, graph, vector, resume, sync, MCP, API,
storage, security, SDK, adapters, GitHub App, enterprise, deployment, observability, roadmap,
ADRs, issues, traceability, admin UI.

AEP is a specification repository. This is a specification corpus for an adjacent problem,
written by the same person. Two pieces are directly relevant:

**`27-traceability/requirements-traceability.md`** is a working PRD → FR/NFR → docs →
milestone matrix with an "exit evidence" column per milestone. AEP has PRDs and ADS and no
matrix connecting them to anything.

**`27-traceability/acceptance-criteria-convention.md`** answers a gap AEP has named and not
filled. `docs/spec/REGISTRIES.md` lists "the evidence package has no defined structure" under
Known gaps, and CLAUDE.md merge requirement 3 demands an evidence package with no schema
behind it. Kairo's convention is four rules: automate first; manual only when necessary, as
Given/When/Then; preconditions explicit; **one proof per checkbox**, citing exactly one make
target, test path, or scenario id.

These are **reference, not port** — they are Kairo's documents, would not pass
`validate_docs.py`, and describe a hosted product AEP is not building. But the AC convention
is the strongest candidate in the whole repository for adoption as an AEP standard, and it is
cheaper to adopt than anything in `pkg/`.

**`25-adrs/0006-postgres-first.md` independently reaches AEP's ADR-003.** Postgres as
metadata store, event log, graph edge store and vector index via `pgvector`, with specialized
backends introduced later "behind the interfaces the code already depends on". That
corroborates clause 5 rather than fighting it: porting toward Postgres moves with Kairo's own
direction, not against it. `25-adrs/0004-disposable-indexes.md` likewise matches the rebuild
requirement in PRD-009.

### Finding 8: Kairo has working agent hooks; AEP has specified ones

`.claude/settings.json` binds four Claude Code events to shell scripts in `.claude/hooks/`:
`SessionStart`, `UserPromptSubmit`, `PostToolUse` on `Edit|Write`, and `Stop`. They start a
Kairo session on the first prompt, mark file edits, checkpoint when a turn ends, and inject
the previous handoff excerpt as context on resume.

AEP's `hooks.yaml` has fourteen bindings and has never executed one. Kairo's equivalent runs
today. The mechanisms are not interchangeable — Claude Code fires at harness events, AEP's
points are `before_agent_run` and `after_merge` — but this is a working implementation of the
capture behaviour AEP specified for `before_agent_run`/`after_agent_run` and never built.

It is also a **third data point on the ADR-020 tension**: these hooks are shell commands named
in configuration, the model ADR-020 clause 2 rejects. Finding 4 found the executable plugin
host, finding 5 found that notes escape rule 1, and this finds the pattern already working in
production in a sibling repository. None of the three is acted on here, for the same reason
each time.

### Also present, smaller

- `test/integration/` — 10 integration suites including `golden_test.go`, `sdk_test.go`,
  `enterprise_test.go`. Port with the code they exercise, or the port ships untested.
- `examples/embed/` — a separate module consuming Kairo as a library. The clearest statement
  of the seam AEP would be porting across.
- `examples/plugin-echo/` — a plugin implementation against the protocol in finding 4.

## Alternatives considered

**(a) Build it in AEP from scratch.** Define the manifest, checkpoint, session and index
schemas and implementations in Phase 1 and Phase 3. Rejected on duplication: it re-derives
roughly 16k lines of working, tested code the same person already wrote, and the result would
be a second dialect of the same idea. What it buys is a clean-room design fitted to AEP rather
than adapted to it, which is a real benefit this decision gives up.

**(b) Port selected packages into AEP.** Chosen. One module, one build, no cross-repo release
coupling, and immediate freedom to adapt to `.agentic/` and Postgres. The cost is a permanent
fork: see Risks.

**(c) Depend on `github.com/MykullZeroOne/kairo` upstream.** Keeps one implementation and one
changelog, and would require only a module-path change and a retag in a repository the same
person owns. Rejected because AEP needs adaptations Kairo should not carry — `.agentic/`
configuration, a Postgres `Engine`, AEP-specific manifest entry kinds — and pushing those
upstream would make Kairo worse as a standalone product. Note the coupling this avoids is
real: two APIs AEP would depend on are marked unstable in the source. `IndexManifest`'s
signature "may narrow to `manifest.Manifest` at G-KG1" (`pkg/knowledge/engine.go:18-20`) and
`ContextBundle` is "Experimental (v1.3)" (`pkg/knowledge/types.go:103`). The `v1` schema tags
do not cover the Go API.

**(d) Do nothing until Phase 3.** Rejected because `context_manifest` is *already* referenced
by an AEP schema example, so the artifact is already assumed and already undefined. Deferring
leaves a dangling identifier in the run record and lets Phase 1 build against an assumption
nothing checks.

## Consequences

- `context_manifest` acquires a definition, and `run-record.example.yaml:12` stops citing an
  identifier for a thing that does not exist.
- **AEP becomes mostly ported code.** Its entire product surface today is 1229 lines of Go
  plus 930 lines of test. Porting ~11k lines makes roughly 90% of the codebase Kairo-derived,
  and inverts what AEP is: a specification repository with a small tool becomes a large
  provenance library with a specification attached. That is the detriment worth weighing
  against everything the port buys, and it is an argument for porting in sequenced slices
  driven by a need, rather than all at once because the code exists.
- The port must split along plane lines per ADR-010 rule 1, so it cannot land as one package
  or one pull request. Sequencing is its own work item.
- AEP inherits Kairo's transitive dependencies — `modernc.org/sqlite`, `aws-sdk-go-v2`,
  `cobra`, `viper`, `golang-migrate`, `lib/pq`. `devctl` currently depends on `yaml.v3` alone.
- AEP must write a Postgres `knowledge.Engine` to keep ADR-003 and ADR-011. Work this decision
  creates, not work it avoids.
- `pkg/session` must gain tests before or during the port; it has none and other ported
  packages depend on it.
- Kairo and AEP begin diverging on the day of the port, permanently.

## Risks

**Single-maintainer bus factor, now doubled.** Kairo and AEP have the same owner, so the port
adds no diversity — it adds ~11k lines of surface area to the same single point of failure that
ADR-014 already accepts for AEP alone. Porting concentrates the risk where depending would have
at least kept the code independently useful.

**Divergence is now certain rather than possible.** Under a dependency, drift is a version
mismatch that a build catches. Under a port there is no version to compare and no build to
fail: a fix made in Kairo simply never arrives in AEP, and nothing anywhere reports that. This
is the cost the decision buys freedom with, and it is the risk most likely to be underestimated
because it produces no signal. Finding 6 softens it materially — the ported half has barely
moved since `v1.5.0` while 56 commits of momentum went into the hosted half AEP excludes — but
softening a risk by observing that upstream is busy elsewhere is not the same as removing it.

**SQLite-versus-Postgres divergence.** Clause 5 keeps AEP on Postgres, so AEP runs a backend
Kairo's 14 knowledge tests do not exercise. Two implementations of one interface, one of them
unexercised, is where behavioural drift appears — a query the ported SQLite engine answers one
way and AEP's Postgres engine answers another, with nothing comparing them. A conformance suite
run against both would mitigate it and does not exist.

**Schema drift between `kairo.*.v1` and AEP's schemas.** Clause 3 adopts the schemas as-is, but
AEP's run record and document schema reference them by identifier. Once ported, AEP will revise
them for its own needs while Kairo revises them for its, and both will still be called
`kairo.context.v1`. The version string stops being a shared meaning the moment the fork happens,
and `pkg/migrate` migrates artifacts within one lineage, not between two.

**Porting untested code.** `pkg/session` has 0 test files, `pkg/core` and `pkg/version` have
none, and `pkg/plugin` — the package that executes external binaries — has 1. Test coverage in
the source repository is uneven, and a port inherits that unevenness silently unless it is
measured before landing.

**Porting more than is needed.** The instruction was "all of it that has value", and value is
easy to assert for code that already exists and passes its own tests. The disposition tables
say what to take, and they should be read as an upper bound rather than a plan: a package with
no AEP caller is a maintenance obligation with no benefit. Sequencing by demonstrated need is
the mitigation, and this ADR does not specify that sequence.

**Scope judgement.** The tables above are an attempt to make "value" concrete. It excludes ~2.6k lines of `hosted*` and two packages AEP has
already solved differently. That judgement is the most reversible part of this ADR and the most
likely to be wrong in either direction.
