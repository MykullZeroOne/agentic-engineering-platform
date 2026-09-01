---
id: ADR-020
type: adr
tier: 1
status: proposed
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-09-01
enforcement: prose
---

# ADR-020 — The Hook Engine

## Context

`.agentic/hooks/hooks.yaml` has carried the same header since it was written: *"NOTHING
EXECUTES THIS FILE YET. There is no hook engine; see the roadmap."* Fourteen bindings are
declared across four hook points. One of them, `validation.docs`, is honoured by CI running
`scripts/validate_docs.py` — but CI runs it because a workflow names the script, not because
anything read the binding. The file has never been executed.

the **Deterministic boundaries** principle says behaviour that must always happen belongs in `hooks.yaml` rather than in
prose asking an agent to remember. A file nothing executes cannot deliver that. It has been
prose in a different syntax.

The cost is measurable rather than theoretical. `work.advance_state` was bound under
`after_merge` in WI-0013 and stayed inert; by WI-0021 the work store had drifted so far that
`devctl work list` showed twelve of twenty rows wrong on its first run. WI-0023 built the
derivation as a command, which is the `check` rung of the enforcement ladder. This ADR is
about the rung above it.

### The constraint that shapes the decision

`after_merge` fires on the trunk, and the obvious first thing to do there — reconcile work
state — writes to `.agentic/work/`. CLAUDE.md rule 1 is unambiguous: *"`main` is protected.
Never commit directly to `main` and never force-push it. All change arrives through a pull
request."*

An engine that runs after merge and writes its result cannot satisfy that rule. This is not
an implementation detail to be worked around; it is the first real design question the hook
engine poses, and the answer determines what the engine is allowed to be.

### Two rounds of review shaped what follows

The first draft was rejected on four counts, all correct: it put the merge record in
authority conflict with ADR-012, it failed open on unimplemented hard hooks, it made a
`soft` binding redden the trunk in contradiction of its own class, and it specified no
runtime contract at all.

The second draft fixed the authority conflict and the soft-hook semantics but was rejected
again on three: `declared` hard bindings still failed open *operationally*, because a skipped
binding did not affect the exit code; keying idempotency by event SHA contradicted a
full-history scan, since the same event yields different results once the trunk grows; and
`work.advance_state` no longer described a hook that does not advance anything.

The corrections below are the substance of this decision, and they narrow it considerably.
**The v1 engine is event-scoped, read-only, and fail-closed.** Everything that needed a wider
scope was moved out of the engine rather than granted to it.

## Decision

**1. The hook engine is a component of `devctl`, invoked as `devctl hook run <point>`.** It
reads `.agentic/hooks/hooks.yaml`, resolves each binding at that point against a registry of
compiled-in Go functions, and executes the ones it has a handler for. It is the first code
that treats `hooks.yaml` as executable configuration rather than documentation.

**2. `hooks.yaml` gains no field naming a command or path to execute.** A binding is a name;
the engine holds the mapping from name to behaviour.

**Two earlier rationales for this clause are dead, and are recorded so nobody revives them.**

The first was security surface — "no mitigation exists". False. Kairo ships
`pkg/security/plugin.go` with `ValidatePluginSpec`, `AuditPluginSpec`, `validatePluginCommand`
and `BuildPluginEnvironment`, and its `pkg/plugin` host runs an audited JSON protocol over
`exec.CommandContext` (ADR-021, findings 4 and 8). Mitigations exist and are portable. Those are
components in a sibling repository, not evidence of production hardening, and the earlier draft
overclaimed by calling them that.

The second was blast radius, and it was **self-defeating**. The argument was that `.agentic/` is
agent-writable, so a command field is a path from "an agent edited configuration" to "an agent
chose what the platform executes". But `platform_config` triggers only on
`.agentic/project.yaml`, `.agentic/registries/**`, `.agentic/hooks/**` and
`.agentic/roles/*.yaml` — **no gate touches Go source at all**. So a `run:` field would need a
human approval record before it could run, while a compiled handler needs only review, which this
repository records as suspended (WI-0016). The chosen design was *less* controlled than the one
it rejected on control grounds.

**The reason that survives is containment of the executable set**, and unlike the other two it is
enforced by the design rather than asserted about it. A compiled handler must exist in this
repository, be reviewed as source, and be built into the binary. A `run:` command may name any
executable on the machine — one not in the repository, never reviewed, never built. Spec
validation checks the *specification*; it cannot check the binary. That distinction holds whatever
the gates say.

**Containment alone is half a control, so the other half is required here.** The change that
introduces the handler package MUST add it as a `platform_config` path trigger, so that changing
what the platform executes needs a human approval record whichever mechanism carries it. Without
that, this clause chooses a bounded set of executables that anyone may change unreviewed, and the
finding above stands. That trigger addition is itself a `platform_config` change.

**The door is named, not sealed.** Revisit when a hook must be written in something other than
Go: the **Open interfaces** principle (`docs/vision/PRINCIPLES.md`) is a real argument against a
Go-only registry, and the ported `security` validation would then be in hand rather than
hypothetical.

**3. Every binding declares its maturity, and the schema requires it.** This is `hooks/v2`.

| `implementation` | Meaning |
| --- | --- |
| `registered` | The engine must find a handler for this ID in its registry. |
| `declared` | Intent only. No handler is expected. |

These tokens, the four result states below, and the `writes` values are **closed sets that a
runtime consumes**, so `REGISTRIES.md` requires them in `vocabularies.yaml` rather than only here.
The `hooks/v2` migration must add `hook_implementation`, `hook_result` and `handler_writes` as
vocabularies; without that, an implementation either hard-codes tokens against an approved tier-0
spec or consumes tokens that do not exist.

There is no default. An earlier draft said "absent means `registered`", which makes the
schema's *interpretation* responsible for a missing field; schema validation is responsible
instead, and a binding without the field is a configuration error. All fourteen existing
bindings must therefore be annotated in the migration, not just the ones anyone remembers. `implemented_by`, which
today names `scripts/validate_docs.py` on the `validation.docs` binding, is removed:
the compiled registry reports the actual handler and its version, and a path in
configuration is the executable-config problem clause 2 rejects.

**4. The engine fails closed.** A `hard` binding that is only `declared` blocks the run.

| Condition | Result | Exit |
| --- | --- | --- |
| `declared`, class `hard` | blocked | 1 |
| `declared`, class `soft` | skipped | 0 |
| `registered`, no handler in registry | configuration error | 2 |
| `registered`, handler declares `writes: external` | configuration error | 2 |
| Handler failed, class `hard` | blocked | 1 |
| Handler **errored**, class `soft` | errored | 3 |
| Handler **reported a finding**, class `soft` | reported | 0 |
| Engine cannot read or validate `hooks.yaml`, or the event | configuration error | 2 |

**Three outcomes, three exit codes, because two of them collapsed in an earlier draft.**

A soft handler erroring is not a soft finding, and it is not a hard block either. An earlier
draft gave it exit 1, the same code as a blocked run, which left a lifecycle caller unable to
distinguish them: it would have to halt on both, contradicting `soft.on_failure: log_and_continue`
in the registry, or continue on both and ignore a hard block. **Exit 3 is that distinction.** A
caller halts on 1 and 2 and continues on 3, and 3 is still not success.

Before that, the same draft gave a soft *error* and a soft *finding* the same exit 0, so a check
that never ran looked like a check that found nothing.
An earlier draft collapsed them, so `work.detect_state_drift` failing to read the supplied
commit looked exactly like it having read the commit and found no drift. The caller saw
success for a check that never ran, which is the same fail-open shape as clause 4 itself,
one level down.

The distinction is between *the run* and *the lifecycle*. `hook-points.yaml` defines `soft` as
`log_and_continue`, and that is honoured: a soft error does not stop later bindings and does not
block progression.

**Exit 0 does not mean every binding executed.** A `declared` soft binding is skipped and still
exits 0, so 0 means "nothing blocked and nothing errored" — not "the point is fully implemented".
An audit surface reading 0 as completion would mark a partially implemented point done. The
structured result per binding is what says which ran; the exit code says only whether to proceed.

The second draft got this wrong: it reported `declared` bindings and let the run exit 0, so
`before_agent_run` could succeed while skipping all four of its `hard` controls. A lifecycle
proceeding without its mandatory checks is the failure mode hooks exist to prevent, and an
engine that reports enforcement it did not perform is worse than no engine.

`devctl hook list` reports incomplete bindings without blocking. Blocking belongs to `run`.

**5. The engine never writes to tracked content, and v1 handlers are read-only.**
No change to the working tree, the index, or commit history on any branch, and no push to
`main`. That is what protects rule 1.

**One exception, and it is the reason this clause was rewritten.** `CLAUDE.md` rule 1 reads
*"Never commit directly to `main` and never force-push it."* Its scope is `main`. A **git
note** is not a commit on `main` — it lives in `refs/notes/*`, a namespace branch protection
on `main` does not cover (ADR-021, finding 5).

So a hook **may** attach provenance to a merge commit by writing a note, and doing so violates
no rule. This is not a loophole being exploited; it is the mechanism git provides for exactly
this problem, which is attaching information to a commit you must not rewrite. AEP needs it
more than most: squash merge means the branch tip is never an ancestor of `main`, so the
`Closes WI-NNNN` trailer is the only surviving link a human controls, and it has now failed
eight times on this trunk. GitHub's own pull request number is the second, and it is the more
reliable one (WI-0032).
A trailer must be written before the merge; a note can be written after.

An earlier draft of this ADR concluded the engine could record nothing after a merge, and
built its whole shape around that. That conclusion was wrong, and it was wrong because it read
rule 1 as broader than it is.

**Scope of the exception, deliberately narrow.** Notes may be written only under the dedicated ref
`refs/notes/aep`, only carrying identifiers that resolve to artifacts stored elsewhere, and never
as a substitute for a record belonging in the work store or in `.agentic/approvals/`.

Transport is part of the contract, not an implementation detail: a note written locally is
invisible to everyone until pushed, and `refs/notes/*` is not fetched by a default clone. The
implementing change must specify the push owner and a fetch refspec, or a note written by one
worker disappears with its checkout. A note is
a pointer, not a place to put state. Writing one is still a `writes: external` operation under
clause 5's second paragraph, so it is refused until the atomic claim exists — which means this
clause opens a door that stays shut at v1, on purpose.

Beyond that, **v1 permits no external writes at all**, unconditionally. What is enforced is the
DECLARATION, not the behaviour: a handler declaring `external` is refused, and a handler declaring
`none` that writes to a network service anyway is not detected. Acceptance criteria 8 and 10 do
not close that, because such a handler can write outside the repository while leaving git state
untouched. Calling the restriction "enforced" would be false; it is a checked declaration on top
of an honour system, and it is why v1 registers exactly one handler.

Every handler declares `writes: none` or `writes: external` as
compiled-in registry metadata, and the engine **refuses any handler declaring `external`** — a
configuration error, exit 2, raised in the pre-flight validation of execution-contract step 1.

Unconditionally, because an earlier draft made the refusal conditional on "no durable event
store configured", which is the wrong gate. Having a store is not having a *claim*: a webhook
retried by GitHub, or the same delivery picked up by two replicas, executes the effect twice
whether or not somewhere is recording that it happened. Deduplication has to be atomic to be
deduplication.

Permitting `writes: external` therefore requires an atomic claim-and-replay of
`(project, event_id, binding_id)` — claim before executing, replay the recorded result on a
second arrival — and that is a later decision with a store design behind it, not a
configuration flag. `evaluation.record` (ADR-004) and `memory.consolidate` are blocked on it.

**6. Derivation produces `effective_state`, a projection. Stored `work_state` remains the
sole authority.** ADR-012 clause 2 says `work_store` names the authoritative store and never
two. `effective_state` is a read-time projection, and the mapping is fixed here rather than left to an
implementation: for an item with a merge record, it is `done` when every gate in its
`required_gates` is covered by a current approval record and `awaiting_human` when any is not;
for an item with no merge record, it is the stored value. That is `work.detect_state_drift`'s own
rule. Without freezing it, an implementation assigning every merged item `done` would satisfy
criterion 11, which only checks that two values are displayed. It is never persisted and never
consulted for dispatch. Where the two disagree, that is **drift** — a reportable condition,
not a contest over which is true. Any surface showing work state must show which of the two
it is showing. ADR-012 is neither superseded nor amended, because nothing moves authority.

**7. The binding is renamed `work.detect_state_drift`.** `work.advance_state` promised to
advance stored state, and clause 5 forbids exactly that. A binding whose name describes
behaviour the engine is prohibited from performing will mislead every reader of the file.
`devctl work advance` — the human-run command that does mutate, on a branch, through a pull
request — keeps its name and its job.

**8. `after_merge` inspects only the commit its event supplies.** Repository-wide
reconciliation moves to `devctl work reconcile --ref main`, a separate command. This is what
makes the event contract and the idempotency contract consistent: a full-history scan keyed
by one event returns different answers as the trunk grows, so the same delivery replayed
later would not be a replay.

A consequence worth stating plainly: **a shallow checkout is sufficient for `after_merge`**,
provided the event commit is present. Full history is required only for `reconcile`.

**9. The event source is the GitHub App webhook, per ADR-018 — which means `after_merge` has
no automatic trigger yet, and this decision does not give it one.**

An earlier draft specified a GitHub Actions workflow on push to `main`. ADR-018 is accepted
and says the opposite twice: clause 3, *"Webhooks are the event source"*, and it rejects
"GitHub Actions as the integration surface" as the primary mechanism while allowing Actions
"for the checks themselves". A hook run is not a check; it is event delivery, which is
exactly what ADR-018 reserves for the App.

An earlier draft added a second reason that does not hold: that Actions cannot construct a stable
`event_id` because it never sees `X-GitHub-Delivery`. Actions exposes `github.run_id` and
`github.run_attempt`, which is exactly what the `delivery` block models, so a retry-stable identity
is constructible there. **ADR-018 alone is the reason, and it is sufficient.** An immutable ADR
carrying a false rationale invites someone to reopen the decision on that ground.

So the trigger is blocked on the App adapter, which is a decision without an implementation.
Until it exists, the engine is invoked explicitly — `devctl hook run <point> --event <file>` —
and nothing fires it automatically. **Naming the dependency is the decision.** Wiring Actions
as a stopgap would contradict an accepted ADR to buy a trigger for one soft, read-only hook,
and would bake an unconstructable key into the contract.

## The event envelope

Every invocation receives one envelope. Hook points do not share a payload shape, and several
have no commit at all — `before_agent_run` needs a work item, an agent run and a workspace,
not a SHA — so the context is generic with a hook-specific payload.

```yaml
event_id: github:push:<delivery-id>     # stable across re-runs of the same delivery
point: after_merge
project: PRJ-mykullzeroone/agentic-engineering-platform
occurred_at: 2026-08-31T18:30:00Z
source: github
subjects:
  commit_sha: abc123
  work_item: null
  agent_run: null
  workspace: null
payload: {}
actor:
  principal: token                  # user | team | token, per SPEC-CAPABILITIES
  id: tok_ci
  on_behalf_of: usr_mykullzeroone   # audit only, never authorization
delivery:
  run_id: "12345"
  attempt: 2
```

**`actor` is required, and an approved artifact requires it.** `docs/spec/CAPABILITIES.md` is
tier 0 and approved (`APR-0014`): a token may act `on_behalf_of` a user, and that attribution is
*carried on every event and checked for audit, never for authorization*. An envelope with no
principal cannot satisfy that, and an earlier draft had none, which would have contradicted a
tier-0 spec approved the same week. The engine validates that `actor` is present and well-formed;
it never consults `on_behalf_of` to decide what a handler may do.

Each point validates the subjects it requires, and a missing subject is a configuration error
(exit 2) rather than a hook that improvises:

| Point | Required subjects |
| --- | --- |
| `before_agent_run` | work item, agent run, workspace |
| `after_agent_run` | work item, completed agent run |
| `before_pr` | base SHA, head SHA |
| `after_merge` | the exact merged commit SHA |

**Idempotency key: `(project, event_id, binding_id)`.** Handler version and configuration
hash are recorded as result metadata, deliberately *not* in the key — including them would
make a handler upgrade look like a new event and defeat replay protection.

Idempotency is **required of handlers, not provided by the engine**. GitHub's re-run button
is the only retry mechanism; the engine has no persistence to deduplicate against and never
retries on its own. Clause 5's read-only restriction is what makes that safe at v1.

## Deterministic execution contract

1. Validate the event, the configuration, and the completeness of the handler registry
   **before executing anything**. A run that would fail on its fourth binding fails before
   its first.
2. Execute `hard` bindings sequentially, in the order they appear in `hooks.yaml`. The file
   is the declaration, so the file is the order: stable under re-run, reviewable in a diff.
3. Stop at the first `blocked` or failed `hard` binding.
4. Execute `soft` bindings only after every `hard` binding has passed.
5. No concurrency and no automatic retry in v1.
6. Emit a structured result per binding: event id, binding id, status, finding, handler version,
   configuration hash. **Handler version is omitted for a `skipped` binding**, which by definition
   has no handler; requiring it there would leave implementations to choose between null, omission
   and a fabricated value.

## Acceptance criteria

**These accept the v1 engine component, not "the hook engine".** `AUTHORITY_MODEL.md` makes the
full set of prose-enforced rules the hook engine's own acceptance criteria, and every one of these
can pass with a single read-only handler, no lifecycle invoking the engine, and most rules still
at `prose`. Passing them must never be reported as completing the hook engine.

This decision's v1 component is not implemented until each of these is demonstrated by a test:

1. A `declared` `hard` binding exits 1.
2. A `declared` `soft` binding is skipped and the run exits 0.
3. A `registered` binding with no handler in the registry exits 2.
4. A failed `hard` binding prevents every later binding from running.
5. A `soft` finding leaves the aggregate exit at 0.
6. `after_merge` reads only the commit its event supplies, and produces the same findings
   whether or not the trunk has advanced since.
7. Replaying the same envelope **at the same handler version and configuration hash**
   produces identical findings. Those two values are in the result envelope precisely so a
   divergence is attributable rather than mysterious: v1 persists no results, so a replay
   after `hooks.yaml` or a handler changes necessarily runs the current implementation, and
   requiring byte-identical output across that would be requiring something the design
   cannot deliver. Determinism is the claim; deduplication is not.
8. A handler declaring `writes: external` is refused with exit 2, unconditionally. An
   earlier draft asserted "re-running one event does not duplicate an external effect", which
   passes vacuously: with external writes forbidden there is nothing to duplicate, so it
   would have stayed green right up until the first writable handler double-applied an event.
9. A `soft` handler that errors exits 3, a `soft` handler that reports a finding exits 0, and a
   blocked `hard` binding exits 1 — three distinguishable outcomes. The soft-error case must also
   assert that a **later binding still ran**, since an exit code alone cannot tell
   `log_and_continue` from stopping.
10. A hook run leaves the working tree, the index, and commit history unchanged. At v1 this
    extends to `refs/notes/*`: the note exception in clause 5 is a `writes: external`
    operation, so it is refused until the atomic claim exists, and a v1 run must write no ref
    at all.
11. Stored `work_state` and projected `effective_state` are displayed as separate values.
12. No command or path taken from `hooks.yaml` is ever executed.

## Alternatives considered

**Bindings name a shell command (`run:` field in `hooks.yaml`).** Maximally general,
language-neutral, and a better fit for the **Open interfaces** principle: a hook would not have to
be written in Go. **Rejected on blast radius, not on security surface**, per clause 2: the
mitigation exists and is portable, but `.agentic/` is agent-writable and a validated arbitrary
command is still an arbitrary command. Acceptance criterion 12 keeps the rejection testable.
The registry is an internal seam, not a published contract, so this stays cheap to supersede
the moment a hook must be written in something other than Go.

**A bot identity pushes bookkeeping commits straight to `main`.** Simplest possible engine
and the only option where the store is never stale. Rejected because it requires amending
CLAUDE.md rule 1, and the carve-out is not small: "bookkeeping only" is a judgement the engine
would make about its own writes, and an engine deciding which of its writes are exempt from
review is the shape of problem the **Human authority** principle exists to prevent.

**The hook opens a pull request with the bookkeeping diff.** Keeps rule 1 literally true and
keeps the store current. Rejected because it automates the bookkeeping crank rather than
removing it: one extra pull request per merge, each itself work needing a work item, which is
the recursion WI-0013 already described.

**GitHub Actions as the event source.** What an earlier draft specified, and rejected on two
count: it contradicts accepted ADR-018 clause 3, which makes webhooks the event source. Actions
remains the right host for *checks*, which ADR-018 permits and which `docs` and `build` already
are.

**`after_merge` reconciles the whole history.** What the second draft specified. Rejected
because it cannot coexist with event-scoped idempotency: replaying a delivery after the trunk
has moved would produce a different result, so the replay is not a replay. Moved to
`devctl work reconcile`, where being history-wide is the point rather than a contradiction.

**Do nothing; leave `hooks.yaml` declarative.** The honest option, and the one in force until
now. Rejected because the ladder's top rung has never been reached and the **Deterministic boundaries** principle is unbacked
without it: a platform whose central determinism mechanism has no implementation is specifying
a claim it cannot make.

## Consequences

- **Three of the four hook points cannot pass at v1**, not one. `before_agent_run` has four
  `declared` hard bindings, `after_agent_run` has two, and `before_pr` has one: seven in total.
  Clause 4 blocks on every one of them. That is the decision working rather than a defect: the engine refuses
  to certify a lifecycle boundary it cannot enforce. `after_merge` is the only point that runs
  clean, and the only one with a registered handler.

- `hooks.yaml` moves to `hooks/v2`, gains `implementation` on every binding, drops
  `implemented_by`, and renames one binding. Its header stops being true. All of that is a
  `platform_config` change needing its own approval.

- `hooks.yaml` holds **fourteen** bindings today (5 at `before_agent_run`, 3 at
  `after_agent_run`, 1 at `before_pr`, 5 at `after_merge`). Exactly one —
  `work.detect_state_drift` — is `registered`, so **thirteen must be marked `declared`**,
  seven of them `hard`. Under clause 3 a binding with no `implementation` field is a
  configuration error, so miscounting here is not cosmetic: it exits 2. The file stops
  implying enforcement it does not have, and that list becomes the specification for what the
  engine must grow.

- Drift becomes reportable rather than discovered by reading twenty files. Reportable, not
  failing: `work.detect_state_drift` is `soft`, so it leaves the exit code at 0 — and not yet
  automatic, since clause 9 leaves the trigger blocked on the App adapter.

- Hooks are testable for the first time. A registry of Go functions is exercisable without a
  merge, a worktree, or a network, which is what makes the acceptance criteria above
  achievable rather than aspirational.

- Writing a hook now requires writing Go and shipping a `devctl` release. This is the main
  cost the rejected shell-command alternative would have avoided.

- `evaluation.record` and `memory.consolidate` are blocked on an **atomic claim-and-replay schema
  with lease and recovery semantics**, which no document specifies. The *store* is not the gap:
  ADR-003 already chooses PostgreSQL for the event log and ADR-004 requires immutable history.
  Naming a missing event store would reopen a settled decision. Recovery is not optional either:
  a worker that crashes after claiming a key but before recording a result leaves a redelivery
  holding a claim with no result, unable to tell whether the effect happened.

- **Git notes become the intended mechanism for post-merge provenance**, and the atomic claim
  becomes the thing standing between AEP and having it. That reorders what matters next: the
  event store is no longer only an `evaluation.record` prerequisite, it is what unblocks the
  one write the engine actually wants to make.

## Risks

**Fail-closed makes the engine mostly a blocker at v1.** One point runs clean and one is
permanently red until four hard hooks are written. If that proves intolerable, the pressure
will be to mark a `hard` binding `soft` to get past it, which would quietly convert a gate
into a log line. Any such reclassification should require the same scrutiny as removing the
gate, because that is what it is.

**The engine reports drift and nothing forces anyone to fix it.** Honouring the `soft` class
means drift never fails a build, so it lands in a job summary that can be ignored
indefinitely. Accepted because `effective_state` makes a stale stored field visible rather
than misleading. If the report is ignored in practice, the fix is an explicit policy
promotion of that one binding to blocking, which `hook-points.yaml` already sanctions, and
not a quiet change to soft semantics.

**`effective_state` adds a second thing "the state" can mean.** Only one is authoritative and
the other is never persisted, but every surface showing work state must now say which it is
showing, and any that forgets will mislead in a way the single-field design could not.

**No external effect is safe under duplicate delivery, and v1 does not make one safe — it forbids
it.** That is a real capability ceiling: **four** `after_merge` bindings that would write stay
`declared` until an atomic claim exists — `memory.consolidate`, `graph.refresh`,
`evaluation.record`, and `work.recalculate_ready`, whose description in `hooks.yaml` recomputes
which items carry the `ready` token and therefore writes the authoritative work store exactly as
the renamed `work.advance_state` would have. The engine's useful surface at v1 is one read-only handler.

**A handler can still lie about `writes`.** The declaration is checked, so a handler that
*declares* `external` is refused; one that declares `none` and writes anyway violates the
contract silently. This is stronger than the honour system an earlier draft relied on and
weaker than sandboxing, and the gap is the reason v1 registers exactly one handler.

**The engine ships with no automatic trigger.** Blocked on the ADR-018 App adapter, so until
that exists the only thing running `after_merge` is a person typing the command — which is
the position the whole exercise was meant to escape. The alternative was contradicting an
accepted ADR, and the dependency is named rather than routed around.

**Reading a rule as broader than it is, is itself a failure mode.** Two drafts of this ADR were
shaped by the belief that rule 1 forbade all post-merge writing. It forbids commits to `main`.
The cost of that misreading was a decision built around a constraint that did not exist, and
it took a survey of an unrelated repository to notice. Every other constraint cited here
deserves the same literal check, and this ADR does not claim to have given it.

**The note exception is a door someone will want to widen.** It permits a pointer under a
dedicated ref and nothing else, and the pressure will be to put state there — a work item's
status, a gate's closure — because notes are convenient and unreviewed. A note is not a
review surface. Anything that must be approved belongs in `.agentic/`, where a human sees it
in a diff.

**`.agentic/` gets less self-describing.** CLAUDE.md calls it "the reference instance of the
schema the first runtime must read unchanged". After this, reading the schema tells you a
hook is bound and whether it is expected to have a handler, but not what it does. That is
discoverable from the tool, not from the file.
