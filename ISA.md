---
task: "Ideal state for the Agentic Engineering Platform"
slug: 20260901-120000_agentic-engineering-platform
project: agentic-engineering-platform
phase: scoping
progress: 0/57
started: 2026-09-01T12:00:00Z
updated: 2026-09-01T12:00:00Z
principal_stated_goal: null
context_sufficient: true
interview_invoked: true
---

# ISA — Agentic Engineering Platform

## Problem

AEP is a specification corpus with a governance layer bolted on and one thin runtime slice
(`devctl doctor`, `devctl work list/show`, the gate check in CI). Everything the product vision
promises — conversational intent capture, a dependency DAG, dispatch across interchangeable
runtimes, durable per-role memory, an evaluation plane — exists only as prose. Twenty-two ADRs
and eight PRDs describe a system nobody can run.

That gap is the real problem, and it has a specific cost. Documentation that no runtime consumes
drifts freely: nothing forces `.agentic/project.yaml` to be readable, nothing forces the seven
planes to hold, nothing forces a role definition to stay model-free. The corpus is currently
self-assessed at adoption level 2 — knowledge-aware — and blocked from level 3 for exactly this
reason: every hook binding is declarative, most rules sit at enforcement stage `prose`, and tier 3
has no artifacts. Meanwhile the humans this is built for keep re-explaining project context to
coding agents that forget it, work inside opaque sessions, and lock them to one vendor's billing.

## Vision

A solo founder describes an idea in conversation, watches a business analyst agent ask the
questions a good BA would ask, approves a specification, and then watches a named engineering
organization plan, build, review, and land the work — with every gate they care about stopping and
asking them, and every gate they don't care about staying out of the way. The surprise is that it
feels like running a team rather than operating a tool: agents have names and histories, they
remember the last argument about the caching layer, and when one is wrong the human can see
exactly which context produced the mistake and correct it once rather than forever.

The second surprise is the bill. It runs on the Claude and Codex subscriptions already being paid
for, and swapping either one out changes nothing about who the agents are.

## Out of Scope

AEP does not replace Git, and does not become an IDE. It does not give legal advice or grant legal
approval. It does not remove human responsibility for consequential decisions, and no agent ever
closes a human gate on a human's behalf. It does not require proprietary models or metered APIs to
function; API-billed routing, if it ever ships, is an option a user turns on, never a dependency.
It does not force adopting projects into one document format.

It is also not a general-purpose agent framework. AEP models a software engineering organization
specifically; the roles, gates, and loops are opinionated about the SDLC and are not intended to
generalize to arbitrary agentic work.

Cloud-hosted multi-tenant SaaS is out of scope for this ideal state. AEP is a system a team runs
itself, against its own repositories and its own subscriptions.

## Language

The corpus glossary at `docs/GLOSSARY.md` is authoritative for AEP's own vocabulary — work item,
work unit, role, agent identity, runtime session, agent run, canonical knowledge, ADS, memory,
event, context packet. That file wins on every term it defines. Only terms whose confusion is
specific to this ISA appear below.

**Claim (ISC)** — an atomic, probe-able statement about AEP's ideal state, checked off when a
named probe passes. _Avoid_: "requirement", "task", "story". A claim is not a work item: work
items in `.agentic/work/` are units of *tracked execution* with their own lifecycle, while a claim
is a unit of *done-ness* for the platform. One claim may take several work items to close, and a
work item may close none.

**Feature (this ISA)** — a vertical slice of AEP that can be verified end to end on its own.
_Avoid_: "plane", "phase", "component". Planes (ADR-010) are a decomposition of responsibility and
deliberately cut *across* features; roadmap phases are a delivery sequence. A feature block here
is neither — it is the smallest cut that produces something a human can watch work.

**Runtime provider** — Claude Code, Codex, or any successor CLI executing an agent identity.
_Avoid_: "model", "the agent", "Claude". The confusion this term exists to kill is the one ADR-002
was written against: a provider is interchangeable plumbing, and it never appears in a role
definition.

## Principles

The eight approved principles in `docs/vision/PRINCIPLES.md` bind this work in full and are not
restated here. Three consequences of them are worth naming because they are what this ISA is
repeatedly tempted to violate:

Human authority is structural, not procedural. It holds because no code path can write an approval
field without an approval record, not because agents are instructed to be polite about it.

Evidence outranks assertion, including AEP's own assertions about itself. The adoption level in
`.agentic/project.yaml`, the enforcement stage of every rule, and the claims in this document are
all subject to the same standard: show the probe or lower the claim.

Determinism belongs in hooks, CI, policy, and state machines. Language models are for ambiguity
and synthesis. Any behavior that must always happen and is implemented as a paragraph asking an
agent to remember is a defect, regardless of how well it currently works.

## Constraints

The seven planes of ADR-010 are the top-level structural commitment: Experience, Organization,
Communication, Control, Knowledge, Execution, Evaluation. No plane owns another plane's state;
cross-plane access goes through an explicit port; providers attach only to Execution and Control.

AEP ships as a Go modular monolith (ADR-011), packages following plane boundaries, one deployable.
Not services per role, not a service per plane.

Subscription-first (ADR-008). Supported CLIs are the foundation; APIs are adapters. A configuration
that reaches for a metered API where a subscription CLI would do is a design error.

Roles are durable identities and runtimes are transient (ADR-002). No role definition may name a
model or a provider.

Exactly one store owns work state (ADR-012), named by `work_store`. GitHub remains authoritative
for pull requests, checks, and releases regardless.

Approval is recorded, never inferred (ADR-013, ADR-019). Setting `human_approved` without a
matching record under `.agentic/approvals/` is a violation, and merging a pull request closes no
gate.

ADRs are immutable once merged; reversal is a new superseding ADR.

## Goal

AEP reaches its ideal state when a technically capable human can point it at a repository, describe
an idea conversationally, and receive delivered, reviewed, traceable software produced by a durable
agent organization running on subscription CLIs — with every human gate enforced by code rather
than convention, every agent decision traceable to the context that produced it, and the whole
system able to swap its execution provider without touching organizational state.

Concretely: all five roadmap phases land, adoption reaches level 5 on AEP's own ladder with
evidence, and AEP is the runtime that reads this repository's own `.agentic/` configuration
unchanged.

## Not yet specified

- fog: what the evaluation plane actually measures well enough to drive a routing decision — must resolve after Phase 3 produces enough event history to know which signals separate a good run from a lucky one.
- fog: whether learned routing is a model choice, a skill choice, a context-budget choice, or all three — waits on the evaluation plane having any real comparisons at all.
- fog: the plugin/provider SDK's surface — what a third party implements to add a runtime, and whether that is a Go interface, a process contract, or MCP — waits on a second and third runtime adapter existing so the shape is discovered rather than guessed.
- fog: the shape of a tier 3 policy artifact. `docs/policies/` and `docs/standards/` are configured and empty; WI-0006 owns this and no schema exists yet.
- fog: the skill schema and its versioning model — a named gap in `docs/spec/REGISTRIES.md` with no decision behind it.
- fog: what "organization template" means as a distributable artifact, and whether a marketplace is a product surface or a directory of repositories.
- fog: whether distributed workers are a scaling need or a premature one — undecided until a single-node orchestrator has been saturated by real work.
- fog: how AEP is distributed and supported once it is not just this repository — licensing, release cadence, and upgrade path for adopting projects.
- fog: how the independent-review requirement is genuinely satisfied. WI-0016 proposes substituting a reviewing agent; whether an agent reviewing an agent counts as independent is unresolved, and the requirement stays suspended until it is.

## Features

### F0 · Cross-cutting governance and provider neutrality
Why: every other feature can be built correctly and still leave AEP a vendor-locked system with decorative gates. This feature is the part that has to hold when the rest is under delivery pressure.

- [ ] ISC-1: A human gate cannot be closed without a matching approval record under `.agentic/approvals/`; the runtime rejects the write.
- [ ] ISC-2: A role definition that names a model or provider fails validation.
- [ ] ISC-3: Replacing the configured runtime provider for every role changes no row of organizational state.
- [ ] ISC-4: Anti: no code path writes `human_approved: true` as a side effect of a merge.
- [ ] ISC-5: Anti: no agent identity is granted a capability that lets it approve its own work item.
- [ ] ISC-6: Every rule in the corpus carries an enforcement stage from `vocabularies.yaml` (`prose`, `check`, `hook`), and no rule at `check` or `hook` lacks the validator, CI job, or binding it names.
- [ ] ISC-7: A change touching four or more planes surfaces that fact for re-examination before implementation, advisory rather than blocking (ADR-010 calls it a signal, not a requirement, and the accepted ADR is what binds).
- [ ] ISC-8: Anti: the platform runs a full idea-to-merge cycle with no metered API credentials present in the environment.

### F1 · Intent to approved specification
Why: the moment that decides whether the rest of the organization builds the right thing — a human describes an idea and walks away with a specification they actually endorse.

- [ ] ISC-9: A conversational description of an idea produces a BA agent session that asks clarifying questions rather than emitting a specification immediately.
- [ ] ISC-10: The BA loop terminates on the human's approval, not on its own judgment of completeness.
- [ ] ISC-11: Compliance, legal, privacy, and security specialists run preflight analysis on the draft intent and report to the BA lead, not to the human directly (ADR-005).
- [ ] ISC-12: An approved specification emits both a human-readable PRD and a machine-readable ADS from the same underlying intent, neither derived from the other (ADR-006).
- [ ] ISC-13: The `product_spec` gate stops the flow and records an approval record before the specification carries authority.
- [ ] ISC-14: Anti: a `draft` PRD or ADS is never cited by a downstream agent as binding intent.
- [ ] ISC-15: Re-running the BA loop on an unchanged idea produces a specification that differs only in ways the human is shown.

### F2 · Planning and the work graph
Why: turning approved intent into a dependency-ordered plan is where a solo human's throughput ceiling actually is, and it is the part they cannot do at speed themselves.

- [ ] ISC-16: An approved ADS decomposes into a DAG of work units with declared dependencies (after: ISC-12).
- [ ] ISC-17: The orchestrator dispatches a work unit only when its dependencies are satisfied, converting it to a work item at dispatch.
- [ ] ISC-18: Work items live in exactly the store named by `work_store`; a second competing store fails validation (after: ISC-16).
- [ ] ISC-19: Work state advances from the merge record rather than from hand-editing.
- [ ] ISC-20: The planner identifies work units that can run in parallel and the orchestrator actually runs them concurrently.
- [ ] ISC-21: Anti: no planned work unit reaches implementation without tracing to an approved specification.
- [ ] ISC-22: A work item's full trace — intent, specification, plan node, run, evidence, merge — is retrievable in one query.

### F3 · Execution across interchangeable runtimes
Why: this is the claim the whole product rests on. If a Codex-implemented story and a Claude-implemented story differ in anything but the code, "role != model" was a slogan.

- [ ] ISC-23: A Claude Code adapter executes an agent identity against a work item and returns a run record.
- [ ] ISC-24: A Codex adapter does the same, producing a run record of identical shape (after: ISC-23).
- [ ] ISC-25: Switching a role's `runtime_preferences` entry between providers requires no change to the role definition (after: ISC-24).
- [ ] ISC-26: An agent run spanning multiple runtime sessions attaches its evidence to the run, not the session.
- [ ] ISC-27: A run that fails mid-flight is resumable without re-deriving context from the human.
- [ ] ISC-28: The engineering, QA, review, and integration loops each execute as defined in `docs/agents/`, with handoffs recorded.
- [ ] ISC-29: Every merged change carries an evidence package conforming to SPEC-EVIDENCE-PACKAGE.
- [ ] ISC-30: Anti: an agent never merges a pull request that crosses a human gate.

### F4 · Human control surface
Why: the vision's actual promise is watching a team work and stepping in when needed. Without this, AEP is a batch job with good documentation.

- [ ] ISC-31: The workspace UI shows every active agent run with live activity, not a spinner.
- [ ] ISC-32: A human can interrupt a running agent and redirect it mid-run.
- [ ] ISC-33: An open question from an agent surfaces to the human with the escalation chain that produced it, and answering it unblocks the run.
- [ ] ISC-34: Antecedent: the human sees the agent's reasoning and its context provenance side by side, so a wrong answer is diagnosable without reading a transcript.
- [ ] ISC-35: A human gate presents what is being approved, what it affects, and what happens on rejection, before it asks.
- [ ] ISC-36: Anti: no interface element lets a human approve a gate without a record being written.

### F5 · Deterministic policy and hooks
Why: adoption level 3 is blocked on exactly this, and every "the agent forgot to" failure in the system traces back to behavior that lived in prose.

- [ ] ISC-37: A hook engine executes the bindings in `.agentic/hooks/hooks.yaml` at their registered hook points. The execution model is not cited here: ADR-020 is still `proposed` and carries no authority until accepted.
- [ ] ISC-38: A failing `hard` hook blocks the lifecycle transition it guards (after: ISC-37).
- [ ] ISC-38.1: A failing `soft` hook logs and continues, and every later binding at that point still runs.
- [ ] ISC-39: Tier 3 policies and standards exist as artifacts and are enforced by executing checks, not prose (after: ISC-37).
- [ ] ISC-40: Human gates are configurable per project and enforced from `.agentic/registries/gates.yaml` as the sole source.
- [ ] ISC-41: Escalation has a registered hook point and is no longer a documented gap.
- [ ] ISC-42: Anti: no gate, policy, or required check can be weakened by an agent to make a change mergeable.

### F6 · Knowledge, memory, and context
Why: the difference between an agent that has worked here for six months and one that started this morning. It is also what stops the human re-explaining the project forever.

- [ ] ISC-43: Every agent run is recorded as immutable events, and derived memory is rebuildable by replay (ADR-004).
- [ ] ISC-44: An agent identity's memory persists across runtime sessions and across a provider change.
- [ ] ISC-45: Retrieval assembles a context packet with provenance on every element, within a declared context budget.
- [ ] ISC-46: Canonical approved knowledge outranks learned memory at retrieval time, and a conflict is visible rather than silently resolved.
- [ ] ISC-47: Learned memory carries source, confidence, scope, and supersession.
- [ ] ISC-48: Provenance nodes and edges record why a change was made, not only that it was (ADR-021).
- [ ] ISC-49: Anti: memory is never cited as authority for a decision that contradicts an approved artifact.

### F7 · Evaluation
Why: without it, every claim about whether a model, prompt, skill, or retrieval change helped is an anecdote — and AEP's whole thesis is evidence over assertion.

- [ ] ISC-50: A failed or corrected run can be converted into a regression case and replayed.
- [ ] ISC-51: Two configurations differing in one variable can be compared over the same case set with a reported result.

### F8 · Adoption and ecosystem
Why: AEP is worth nothing if adopting it requires a greenfield repository, and it is worth little if it only ever runs against GitHub and two CLIs.

- [ ] ISC-52: `devctl init` and `devctl adopt` bring an existing repository to a working AEP configuration, and `devctl doctor` reports what is missing.
- [ ] ISC-53: AEP reads this repository's own `.agentic/` configuration unchanged and operates on it.
- [ ] ISC-54: An MCP surface exposes knowledge, context, communication, and runtime operations to external clients.
- [ ] ISC-55: A second SCM adapter exists, proving the Control plane's port is real (after: ISC-53).
- [ ] ISC-56: Anti: adopting AEP never requires restructuring an existing repository's documentation format.

## Test Strategy

| isc | type | check | threshold | tool | anchors_to |
| --- | --- | --- | --- | --- | --- |
| ISC-1 | integration | write a gate-close with no record | rejected | go test | ADR-019 |
| ISC-2 | unit | validate a role naming a provider | fails | go test | ADR-002 |
| ISC-3 | integration | diff org state across provider swap | zero rows | go test | ADR-002 |
| ISC-4 | integration | merge a PR, inspect approval fields | unchanged | go test | ADR-019 |
| ISC-5 | unit | capability grant self-approval attempt | denied | go test | PRINCIPLES#human-authority |
| ISC-6 | static | audit rules at stage check/hook for a backing artifact | zero unbacked | validate_docs.py | vocabularies.yaml#enforcement_stage |
| ISC-7 | ci | plane-count check on a diff | warns at 4, never blocks | GitHub Actions | ADR-010#consequences |
| ISC-8 | integration | full cycle, no API keys in env | completes | shell harness | ADR-008 |
| ISC-9 | scenario | submit an idea, count questions asked | >=1 before spec | scripted run | PRODUCT_VISION#outcomes |
| ISC-10 | scenario | withhold approval, observe loop | does not exit | scripted run | PRINCIPLES#human-authority |
| ISC-11 | integration | inspect specialist report routing | to BA lead | go test | ADR-005 |
| ISC-12 | integration | approve intent, inspect outputs | PRD + ADS emitted | go test | ADR-006 |
| ISC-13 | integration | advance past spec without record | blocked | go test | gates.yaml#product_spec |
| ISC-14 | static | retrieval over a draft artifact | not returned as binding | go test | AUTHORITY_MODEL |
| ISC-15 | scenario | re-run BA on same idea | diff surfaced | scripted run | PRODUCT_VISION#outcomes |
| ISC-16 | integration | decompose an ADS | DAG with edges | go test | roadmap#phase-2 |
| ISC-17 | unit | dispatch with unmet dependency | not dispatched | go test | GLOSSARY#work-unit |
| ISC-18 | unit | configure two stores | validation error | go test | ADR-012 |
| ISC-19 | integration | merge, then read work state | derived, not edited | go test | WI-0023 |
| ISC-20 | integration | plan with independent branches | concurrent runs | go test | roadmap#phase-2 |
| ISC-21 | integration | dispatch unspecified work unit | rejected | go test | PRINCIPLES#evidence |
| ISC-22 | integration | query one work item's trace | full chain returned | go test | ADR-021 |
| ISC-23 | integration | run a work item via Claude adapter | run record produced | go test | ADR-008 |
| ISC-24 | integration | same via Codex adapter | identical shape | go test | ADR-008 |
| ISC-25 | unit | swap runtime_preferences | role file unchanged | go test | ADR-002 |
| ISC-26 | integration | two-session run, inspect evidence | attached to run | go test | GLOSSARY#agent-run |
| ISC-27 | scenario | kill mid-run, resume | no human re-input | scripted run | PRODUCT_VISION#problem |
| ISC-28 | scenario | idea to merge, inspect handoffs | all recorded | scripted run | docs/agents |
| ISC-29 | ci | evidence package on merged PR | schema-valid | GitHub Actions | SPEC-EVIDENCE-PACKAGE |
| ISC-30 | integration | agent merge across a gate | blocked | go test | ADR-019 |
| ISC-31 | ui | observe a live run in workspace | activity within 2s | Interceptor | roadmap#phase-1 |
| ISC-32 | ui | interrupt and redirect a run | run changes course | Interceptor | PRODUCT_VISION#outcomes |
| ISC-33 | ui | agent raises a question | chain shown, answer unblocks | Interceptor | QUESTION_ESCALATION |
| ISC-34 | ui | inspect a wrong answer | context provenance visible | Interceptor | docs/context/PROVENANCE |
| ISC-35 | ui | trigger a gate | scope and consequence shown | Interceptor | ADR-019 |
| ISC-36 | integration | approve via UI, inspect store | record written | go test | ADR-019 |
| ISC-37 | integration | fire a registered hook point | binding executes | go test | ADR-007 |
| ISC-38 | integration | failing hard hook on a transition | transition blocked | go test | hooks.yaml#class |
| ISC-38.1 | integration | failing soft hook, inspect later bindings | logged, all still run | go test | hooks.yaml#class |
| ISC-39 | static | tier 3 artifact count and stage | >0, enforced | validate_docs.py | REGISTRIES#known-gaps |
| ISC-40 | unit | gate list read from registry | single source | go test | gates.yaml |
| ISC-41 | static | escalation hook point registered | present | validate_docs.py | hook-points.yaml |
| ISC-42 | ci | agent-authored check weakening | blocked | GitHub Actions | PRINCIPLES#deterministic |
| ISC-43 | integration | replay events, rebuild memory | byte-identical | go test | ADR-004 |
| ISC-44 | integration | memory across session + provider change | intact | go test | ADR-002 |
| ISC-45 | integration | assemble a context packet | provenance on all, within budget | go test | docs/context |
| ISC-46 | integration | conflicting memory vs approved doc | canonical wins, conflict logged | go test | PRINCIPLES#memory |
| ISC-47 | unit | memory record schema | four fields present | go test | PRINCIPLES#memory |
| ISC-48 | integration | inspect a change's provenance | why recorded | go test | ADR-021 |
| ISC-49 | integration | memory contradicting an ADR | not cited as authority | go test | AUTHORITY_MODEL |
| ISC-50 | integration | convert a failed run, replay | reproduces | go test | ADR-004 |
| ISC-51 | integration | A/B one variable over a case set | result reported | go test | roadmap#phase-4 |
| ISC-52 | cli | init/adopt/doctor on a bare repo | working config | go test | roadmap#phase-1 |
| ISC-53 | integration | run AEP against this repo's .agentic | reads unchanged | go test | project.yaml |
| ISC-54 | integration | MCP client against the surface | four operation classes | mcp client | TECHNOLOGY_STACK |
| ISC-55 | integration | second SCM adapter smoke test | passes | go test | ADR-010 |
| ISC-56 | scenario | adopt a repo with foreign doc format | no restructure required | scripted run | PRODUCT_VISION#non-goals |

## Decisions

- 2026-09-01: Goal-signal detection fired on signal 4 (structural directive), but the literal — "start an new ISA based on the current project agentic-engineering-platform" — directs the creation of this artifact rather than stating AEP's target state. Anchoring derivation to it would be useless, so `principal_stated_goal` is null per the fail-closed minimum-content rule and the candidate is logged here instead.
- 2026-09-01: Ambiguity check fired: "an ISA for this project" supported a Phase-1-slice reading and a full-platform reading, which produce materially different artifacts. Principal chose the full platform through roadmap Phase 5. Phases 1–5 are therefore all in scope and much of Phases 4–5 is legitimately fog rather than claims.
- 2026-09-01: Principal chose the repo root for this file. `CLAUDE.md` permits only `README.md` and `CLAUDE.md` as root-level markdown, so this placement conflicts with a stated convention. It breaks no check — `scripts/validate_docs.py` collects only `docs/**` and `examples/**` — but the convention needs an explicit carve-out rather than a silent exception. Tracked in Remaining Work.
- 2026-09-01: Resolved the root-placement conflict in `CLAUDE.md` rather than moving the file. The carve-out is explicit about what `ISA.md` is not: not a tier 0-3 artifact, never citable as binding intent, and losing to any approved document it contradicts. That bound is what makes a root-level exception safe.
- 2026-09-01: Review found four claims that overstated their sources, all corrected rather than defended. ISC-6 cited an enforcement stage `enforced` that `vocabularies.yaml` does not define (the tokens are `prose`, `check`, `hook`). ISC-7 made ADR-010's four-plane rule a blocking CI gate demanding an ADR; the accepted ADR calls it a signal to re-examine, and the accepted artifact binds. ISC-38 required every failing hook to block, contradicting `hooks.yaml`'s soft class, and now splits into ISC-38 (hard blocks) and ISC-38.1 (soft logs and continues, later bindings still run).
- 2026-09-01: The fourth finding was the one worth the carve-out changing. A hand-checked box beside `.agentic/work/` is the competing markdown backlog principle 3 forbids, so `CLAUDE.md` now states that `ISA.md` holds no work state, that a checked claim records a passing probe rather than completed work, and that the work store wins on disagreement. The Remaining Work item changed from reconciling two views to deriving the checked state from evidence.
- 2026-09-01: Second review pass corrected three more anchors. ISC-1 and ISC-35 pointed at ADR-013, which ADR-019 superseded precisely because its hash-scope and approval-pointer rules were wrong; both now cite ADR-019. ISC-37 cited ADR-020 for the hook execution model, but ADR-020 is `proposed` and a proposed artifact carries no authority, so the claim now names the binding registry and says explicitly why the ADR is not cited.
- 2026-09-01: `## Remaining Work` drops its checkboxes. The ISA format spec writes those lines as `- [ ]`, but principle 3 forbids a checkbox task list in markdown beside `.agentic/work/`, and the first box had already gone `[x]` while WI-0038 was still `in_progress` -- the exact silent disagreement the principle exists to prevent. The repository's constitution wins over the artifact format inside this repository.
- 2026-09-01: WI-0038 now declares `required_gates: [platform_config]`. Whether amending a documentation convention is a policy change was left open earlier; leaving it open was the wrong call, because `work.advance_state` treats an item with no required gates as `done` at merge, so silence would have auto-closed the question in the direction the agent preferred. Declaring the gate makes the item land `awaiting_human` instead, which is where a question for the principal belongs.
- 2026-09-01: `## Dependencies` and `## Bridge Criteria` omitted — AEP has no sibling ISAs and no cross-ISA contracts.
- 2026-09-01: `## Learning` omitted at scaffold. Nothing has been conjectured and refuted yet; the section appears when it has a four-piece entry to hold.
- 2026-09-01: Independent review is recorded as suspended, not satisfied (WI-0015), and this ISA does not claim it. Whether an agent reviewer satisfies it is held as fog, not asserted as ISC-closable.

## Verification

No claims closed. Provenance stubs land here as claims go `[x]`.

## Remaining Work

Not a backlog and not work state: `.agentic/work/` owns both, and these carry no checkboxes so
the two can never disagree. Each line names something this ISA still owes and where it is tracked.

- Derive each claim's checked state from evidence rather than hand-maintaining it. A claim goes
  `[x]` because its probe passed and a provenance stub exists, never because someone ticked it.
  Until a tool derives it, the work store wins on any disagreement.
- Give each open claim the work items that would close it, so the ISA reads as a view over
  `.agentic/work/` rather than a parallel list. Needs a work item; none exists yet.
- Decide whether `progress:` in this front matter should be tool-derived. Hand-maintained counts
  drift, and this one already did once.
- Root placement is settled: `CLAUDE.md` now admits `ISA.md` and bounds it. Tracked by WI-0038,
  which is where its state lives.
