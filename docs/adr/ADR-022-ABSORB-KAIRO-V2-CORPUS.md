---
id: ADR-022
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-09-01
approval_record: APR-0018
supersedes: null
superseded_by: null
last_reviewed: 2026-09-01
---

# ADR-022 — Absorb Kairo's V2 Corpus by Translation

## Context

ADR-021 proposes porting Kairo's provenance code and marks `docs/v2/` as reference-only. That is
right for a code port and wrong for the corpus, because the corpus is the part of Kairo that most
directly states AEP's own premise.

**Surveyed revision: `17cdd4565f93dabb3c0cf9285383cbffe9c0c7fa`** (Kairo, branch
`feat/217-projection-drain`, 2026-07-08). Every path, word count, quotation and disposition below
resolves against that commit and nothing else. The revision is pinned because the branch tip
moves and ADR-021's `v1.5.0` reference does not identify what this survey read — 56 commits
separate the two. `REGISTRIES.md` records that external reference has no authority tier and that
its retrieval must reach the provenance manifest; a corpus cited without a revision cannot.

`docs/v2/` is 29 numbered sections plus a `changes/` directory, 63 files, roughly 38,000 words:
vision, philosophy, PRD, SRS, domain model, architecture, hosted, knowledge engine, memory,
events, graph, vector, resume, sync, MCP, API, storage, security, SDK, adapters, GitHub App,
enterprise, deployment, observability, roadmap, 14 ADRs, issues, traceability, admin UI.

Kairo was the first attempt at the problem AEP now solves. Its V2 corpus is the most developed
statement of that problem anywhere, written by the same person, and several of its conclusions
match AEP's reached separately.

### It cannot be copied

**No front matter.** All 63 files lack `id`, `type`, `tier`, `status` and the approval fields
`docs/spec/DOCUMENT_LIFECYCLE.md` requires. Every one would fail `scripts/validate_docs.py`.

**Invalid identifiers and a competing sequence.** `docs/v2/25-adrs/` holds `ADR-0001` through
`ADR-0014`, alongside its own `PRD-01..12` and `FR`/`NFR` numbering. The four-digit form violates
the `ADR-NNN` pattern `DOCUMENT_LIFECYCLE.md` defines and `scripts/validate_docs.py:43` enforces,
so those identifiers are invalid here rather than duplicates of AEP's. The `FR`/`NFR` identifiers
collide with nothing — AEP's requirements use `REQ-<spec>-NNN` — but they are a second
requirement hierarchy with no relationship to the first.

**Conflicts that a copy could not resolve.** Two are named below. The precedence rules make this
worse rather than better: if imported ADRs were somehow made authoritative, `AUTHORITY_MODEL.md`
rule 3 leaves same-tier conflicts **unresolved until one supersedes the other**, which it calls an
error. If they stayed front-matterless or `draft`, rule 2 gives them no authority at all. Neither
outcome is a corpus a reader can use.

## Decision

**Absorb `docs/v2/` by translation. Nothing is copied.**

Each section is triaged into one of three dispositions:

1. **Absorb** — the content becomes or extends an AEP document, rewritten to AEP's schema and
   authority model, citing the Kairo source. The idea carries; the artifact is AEP's.
2. **Reference** — retained as source material and cited where it is useful, most often to explain
   a ported behaviour. Never an AEP document, never in `docs/`, never authoritative.
3. **Out of scope** — describes a hosted multi-tenant product AEP has not decided to be. Recorded
   here so the exclusion is a decision rather than an oversight.

Dispositions are assigned per section. **Four sections carry a stated exception at file
granularity**, marked in the table: `02-prd`, `22-deployment`, `25-adrs` and `27-traceability`. An
earlier draft claimed each section gets exactly one disposition, which the table did not honour.

**Where an APPROVED OR ACCEPTED AEP artifact and the Kairo corpus disagree, the AEP artifact
governs.** Scoped deliberately: `AUTHORITY_MODEL.md` rule 2 gives `draft` and `proposed` artifacts
no authority, so a draft cannot govern anything. Where the AEP side is itself a draft — ADR-021,
PRD-006, PRD-009 — both are proposals, and the absorption is a proposal about proposals until
each is approved on its own.

**Absorption never silently widens scope.** Anything that would change a principle, move
authority, or expand what AEP is gets its own gated change.

**Absorption never edits an accepted ADR.** Accepted ADRs are immutable and reversible only by
supersession (`DOCUMENT_LIFECYCLE.md`). Where the table names an accepted ADR as a *target*, that
means the material lands in a lower-tier design or implementation artifact that the ADR governs,
or it requires a superseding ADR and its `architecture_decision` gate. It never means editing the
ADR in place.

## Two conflicts, named

### 1. Law 2 would move authority

`01-philosophy` states: *"Kairo owns the context around software change: prompts, decisions,
agent sessions, **human approvals**, external snapshots..."*

AEP's approvals live in `.agentic/approvals/` and are closed only by a human, per the
**Human authority** principle and ADR-019. A knowledge layer that "owns human approvals" reads as
the place approvals are recorded, and therefore decided.

**Absorbed narrowed:** the knowledge layer *indexes* approvals and never owns them. It may answer
"which approval closed this gate"; it is never where the answer is written.

This is the most dangerous line in the corpus, precisely because copying it looks harmless.

### 2. The runtime decision conflicts with ADR-011

v2 `ADR-0013` decides a hybrid runtime: a Go core compiled to WASM with TypeScript handlers.
ADR-011 decides Go for the control service, `devctl` and the local worker daemon, with
React/TypeScript confined to the UI. Those are different runtimes for the same layer.

**Out of scope**, and the clearest evidence the v2 corpus answers a different question: it is a
serverless-deployment decision, and AEP has no serverless.

### Not a conflict: hosted-primary

An earlier draft called v2 `ADR-0001` (hosted platform primary) a contradiction of ADR-011. **It
is not.** ADR-011 decides language, storage and a single-deployable modular-monolith *shape*; it
does not forbid hosting that deployable, and `docs/architecture/DEPLOYMENT.md` already specifies
an AEP server with local worker daemons.

Hosted-primary is still **out of scope** — it is new product scope, with its own ADR and gate, and
not something to acquire by absorbing a corpus. But rejecting it as scope is a different and
weaker claim than calling it a contradiction, and the difference matters.

## Corroborations

Places where Kairo reached a conclusion AEP also holds. Absorbed as supporting citation, not as
new decisions:

| Kairo | AEP | Note |
| --- | --- | --- |
| Non-negotiable 4: "must not rely on agents manually remembering to snapshot" | **Deterministic boundaries** (`docs/vision/PRINCIPLES.md`) | Near-identical wording, reached separately |
| Law 5 / v2 ADR-0002: context reconstructed, capture ambient | **PRD-006** (draft) | The distinction PRD-006 rests on |
| Law 4 / v2 ADR-0004: indexes are disposable | **PRD-009** (draft), ADR-003 (accepted) | Already written into PRD-009 |
| Law 3 / v2 ADR-0003: knowledge is event-driven | **ADR-004** (accepted) | |
| v2 ADR-0006: Postgres-first, specialized backends later behind the same interfaces | **ADR-003** (accepted) | Independent agreement |

Two of the AEP artifacts above are `draft` and carry no authority; they are named as proposals.

**This agreement is weaker evidence than it looks.** Both corpora have the same author, so the
same person reaching the same conclusion twice is consistency, not validation. It is recorded
because it is useful context, not because it confirms anything.

## Section triage

| Section | Words | Disposition |
| --- | --- | --- |
| `00-vision` | 409 | **Absorb** into `docs/vision/`. "Git stores the code. It does not store the full engineering state" states AEP's thesis better than AEP states it. Drop the V1→V2 hosted framing. |
| `01-philosophy` | 232 | **Absorb, narrowed.** Laws 1, 3, 4, 5 as supporting citation. Law 2 narrowed per conflict 1. |
| `02-prd` | 1133 | **Out of scope**, *except* its problem statement, absorbable into `00-vision`. A hosted-platform PRD otherwise. |
| `03-srs` | 2083 | **Reference.** A second requirement hierarchy with no mapping to `REQ-<spec>-NNN`. |
| `04-domain` | 1332 | **Absorb** into `docs/spec/entity-model.yaml`, the canonical model per ADR-017. `ENTITY_MODEL.md` is a reading guide and is not the target. Eleven open questions remain there; a second independent model is input for them. |
| `05-architecture` | 504 | **Reference.** ADR-010's seven planes already decide this. |
| `06-hosted` | 277 | **Out of scope.** |
| `07-knowledge-engine` | 254 | **Absorb** into PRD-004 / PRD-009 as the engine's stated contract. |
| `08-memory` | 283 | **Absorbed already** — landed in PRD-004 (`PRD-004:37-39`) via WI-0028. |
| `09-events` | 2017 | **Absorb** the event envelope and fixture contract into ADR-020's event surface. ADR-004 is accepted, so nothing lands in it. |
| `10-graph` | 240 | **Absorbed already** — landed in `docs/memory/KNOWLEDGE_GRAPH.md` and the `graph_node_type` / `graph_edge_type` registries via WI-0030, not in PRD-004 as an earlier draft said. |
| `11-vector` | 851 | **Absorb** into PRD-004. ADR-003 is accepted; pgvector detail lands in design artifacts, not in the ADR. |
| `12-resume` | 218 | **Absorbed already** — landed in PRD-006 (`PRD-006:38-43`) via WI-0028. |
| `13-sync` | 1448 | **Out of scope.** Local↔hosted sync. |
| `14-mcp` | 815 | **Reference**, and it exposes a gap: AEP has no MCP decision. That is a new ADR, not an absorption. |
| `15-api` | 1410 | **Out of scope.** Hosted API surface. |
| `16-storage` | 3445 | **Absorb** the content-addressed object store and GC design alongside the `pkg/storage` port. |
| `17-security` | 1523 | **Split.** `auth-rbac.md` is **absorbed already** — its principal model landed in `docs/spec/CAPABILITIES.md` via WI-0029 and is **approved** under `APR-0014`, so re-translating it risks contradicting an approved document. Only `security-architecture.md` remains, absorbable into `docs/security/GOVERNANCE.md` (`draft`), and it is that half alone that **triggers `security_policy`**. |
| `18-sdk` | 114 | **Reference.** |
| `19-adapters` | 135 | **Reference.** ADR-021 finding 1 establishes these are not AEP's adapters. |
| `20-github-app` | 883 | **Reference.** ADR-018 is accepted and cannot be extended in place; this is input for the design artifact behind it, or for a superseding ADR. |
| `21-enterprise` | 727 | **Out of scope.** Multi-tenant. |
| `22-deployment` | 2688 | **Out of scope**, *except* `production-transport-encryption.md`, absorbable into `GOVERNANCE.md`. |
| `23-observability` | 121 | **Absorbed already** — landed in PRD-007 (`PRD-007:30-33`) via WI-0028, not in ADR-010. |
| `24-roadmap` | 1359 | **Reference.** Two roadmaps is the drift this repository legislates against. |
| `25-adrs` | 5842 | **Triaged individually**, below. |
| `26-issues` | 2463 | **Reference.** Kairo's backlog. |
| `27-traceability` | 3321 | **Partly absorbed already** — the AC convention became `SPEC-EVIDENCE-PACKAGE` via WI-0026. `requirements-traceability.md` remains **absorbable** as the matrix AEP lacks. |
| `28-admin-ui` | 567 | **Out of scope.** |
| `changes/` | 1356 | **Reference.** |

### The 14 v2 ADRs

All fourteen accounted for:

| ADR | Disposition |
| --- | --- |
| `0002` ambient memory, `0003` event-driven, `0004` disposable indexes, `0006` Postgres-first | **Corroborate.** See the table above. |
| `0007` migration tooling | **Absorb** alongside the `pkg/migrate` port. |
| `0005` MCP as agent interface | **Reference.** Exposes the MCP gap; needs its own ADR. |
| `0001` hosted-primary | **Out of scope** as product scope. Not a contradiction — see above. |
| `0013` Go-WASM + TypeScript runtime | **Out of scope.** Conflicts with ADR-011. |
| `0008` compute hosting, `0009` managed data services, `0010` cache and queue, `0011` CI/CD and registry, `0012` embedding provider, `0014` event-driven processing | **Out of scope.** Hosted-infrastructure decisions. |

### What the triage adds up to

By word count across the 30 rows: about **15,100 words Absorb**, **8,800 Reference**, **8,250 Out
of scope**, and **5,842 in `25-adrs`**, split further in the table above. Mixed rows put a few
hundred words on both sides of a line.

An earlier draft asserted 40 / 25 / 35 percent and claimed every line of the out-of-scope share
was hosted. Neither survives the arithmetic. What is true: **the majority of the out-of-scope
material is hosted, multi-tenant or serverless**, and the single largest absorbable block is
`16-storage`.

## Alternatives considered

**(a) Copy `docs/v2/` into `docs/`.** Rejected on all three obstacles: 63 validator failures,
invalid identifiers from a competing sequence, and imported conflicts that the precedence rules
leave unresolved rather than settle. It would also make a large share of AEP's corpus describe a
product AEP has not decided to be.

**(b) Reference only, as ADR-021 proposes.** The status quo. Rejected because it leaves the
strongest statement of AEP's own premise outside AEP, and because two absorptions unblock things
AEP is stuck on: `04-domain` feeds the eleven open questions in `entity-model.yaml`, and
`17-security/security-architecture.md` feeds a `GOVERNANCE.md` still in draft.

**(c) Absorb everything including the hosted sections.** Rejected as a product decision smuggled
in as a documentation task. If AEP should be a hosted platform, that deserves an ADR arguing it.

**(d) Translate lazily, when each section is needed.** Genuinely attractive and nearly chosen: it
avoids work with no consumer. Rejected because the triage itself is the valuable artifact and it
decays fast — four rows in the table above went stale within a day of being written. This ADR
fixes the triage; the translations still happen lazily.

## Consequences

- Roughly a dozen absorptions become work. Several targets are `draft` and cross no gate; the
  ones that do are enumerated below rather than assumed.
- **Gates the remaining translations trigger**, per `.agentic/registries/gates.yaml`:
  `constitutional_change` for `docs/spec/**`; `product_spec` for approving a PRD or ADS;
  `security_policy` for `17-security/security-architecture.md` and `22-deployment` landing in
  `docs/security/**`; and
  `architecture_decision` for any superseding ADR required where an accepted one is involved.
- `docs/vision/` and `docs/security/GOVERNANCE.md` get materially better inputs.
- The eleven open questions in `entity-model.yaml` get a second independent model to answer
  against.
- AEP acquires no hosted, multi-tenant or serverless commitment, and this is where that refusal is
  recorded.
- Nothing here changes `docs/vision/PRINCIPLES.md`. Absorbing Kairo's non-negotiables as
  *citation* is not amending AEP's principles, which is tier 0 and gated.

## Risks

**Translation is where meaning gets lost, and there is no test for it.** A ported function either
compiles or does not. An absorbed idea can arrive subtly wrong and nothing catches it. Conflict 1
is the demonstration: one word, "owns", would have moved authority, and it was caught by reading.

**The triage goes stale faster than it can be executed.** Six rows were already stale within a
day of the first draft, because the absorptions they described had landed — and one of those,
`17-security/auth-rbac.md`, had landed in a document that is now *approved*. A row that sends
later work to re-translate approved content is worse than a merely out-of-date one: it invites a
contradiction against an artifact carrying authority.

A triage table is a snapshot of remaining work and this document is not the work tracker. If the
absorptions run over weeks rather than days, the table should be replaced by work items and this
ADR should keep only the dispositions, which do not decay.

**Absorption grows AEP's scope at the moment it should be shrinking.** The out-of-scope share is
the mitigation, and it will be under pressure the first time a hosted section looks convenient.

**This ADR triages 63 files from reading twelve of them in full.** `00-vision`, `01-philosophy`,
`02-prd`, all five of `27-traceability`, `04-domain`, `17-security`, two v2 ADRs and the full ADR
title list were read; the rest was triaged from titles, word counts and section names. The
**Absorb** rows are the likeliest to be wrong, and each should be confirmed when its translation
is written.

**A first draft of this document contained sixteen defects**, four of them authority-model errors
including instructions to edit immutable accepted ADRs, a misstatement of precedence rule 3, and
an incomplete gate enumeration. They were found by an independent review, not by writing more
carefully. That is the strongest available argument for the review requirement this repository
currently has suspended (WI-0016).
