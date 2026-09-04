---
id: ADR-023
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
last_reviewed: 2026-09-04
enforcement: prose
---

# ADR-023 — The Universal Agent Loop

## Context

AEP's central structural claim is that it models a software engineering organization in which
every tier runs the same recursive primitive, and only the role configuration changes between
tiers. Nothing in tiers 0 or 1 said so.

A corpus-wide audit on 2026-09-04 established the scale of the omission. Across all 22 accepted
architecture decisions, the word `specialist` appears zero times, `recursive` zero times, and
`definition of done` zero times. The only *approved* statement of AEP's organizational flow is
`docs/architecture/SYSTEM_ARCHITECTURE.md:46`, a single arrow chain from Human to Merge with no
recursion, no delegation and no escalation — the linear pipeline this platform exists not to be.
The loop itself lived in `docs/agents/UNIVERSAL_AGENT_LOOP.md`, a `draft` design document in the
one directory tree no gate protects, cited by none of the eight per-role loop documents that are
supposed to be its configurations.

That is why the loop needs a decision rather than another document. A thesis with no tier-1
authority does not hold, and the audit's other headline finding is the consequence: 38 of the
first 41 work items went to the repository's own governance machinery, because governance was the
part that was ratified and gated.

Three incompatible versions of the loop were in circulation when this ADR was written: eight
steps in the principal's own architecture infographic, eleven in `UNIVERSAL_AGENT_LOOP.md`, and
thirteen in a restatement of product intent. They agree in spirit and disagree in count, and no
role configuration, hook point, or test can be written against three shapes.

## Decision

**Every organizational tier in AEP runs the same ten-step loop. Only the role configuration
changes between tiers.**

1. **Receive** — accept a work assignment with an explicit objective and authority boundary.
2. **Hydrate** — assemble the context packet: role identity, canonical knowledge, role and
   organizational memory, prior experience, procedures, pitfalls, open questions.
3. **Analyze** — determine knowns, unknowns, risks, missing evidence, and which specialists
   materially improve confidence.
4. **Delegate** — invoke specialist sub-agents.
5. **Collect** — gather specialist findings and synthesize them, resolving contradictions.
6. **Resolve** — answer questions from canonical knowledge and durable memory. Escalate only what
   cannot be answered that way.
7. **Validate** — evaluate this role's completion gate, its definition of done.
8. **Produce** — emit artifacts and the evidence for them.
9. **Hand off** — pass output and approved exceptions to the next owning role.
10. **Consolidate** — extract memory candidates from the run.

**Escalation is a branch, not a step.** It hangs off step 6 and travels one hop at a time up the
hierarchy: sub-agent, parent, parent, human. It is deliberately absent from the numbering, because
a numbered step implies every pass performs it, and the human is the most expensive reasoning
resource in the organization. The ladder itself is a separate decision and is not settled here.

**One re-entry edge makes the loop recursive.** From step 6 or step 7, if unresolved unknowns
remain, control returns to step 3. This edge is the recursion. Without it the ten steps are a
pipeline with more boxes.

**A tier configures the loop; it does not redefine it.** A role definition supplies the specialist
set for step 4, the completion criteria for step 7, and the artifact and evidence shapes for step
8. A tier that needs a different step sequence is evidence the decomposition is wrong, and is
grounds to supersede this ADR rather than to special-case a role.

## Alternatives considered

**The eight-step form from the architecture infographic.** Rejected on two specifics, and adopted
on a third. It omits Receive, leaving the loop with no entry contract and nowhere to bind the
authority boundary. It folds Consolidate into Hand Off, which leaves the memory plane with no step
to attach to even though `memory.consolidate` is already a declared `after_merge` binding in
`.agentic/hooks/hooks.yaml`. Its treatment of escalation as a side path rather than a numbered
step is correct, and is adopted here.

**The thirteen-step form.** Rejected as over-decomposed. "Identify unknowns" is the output of
Analyze rather than a phase after it, and "Refine" is the re-entry edge wearing a step number.
Neither admits a probe distinct from its neighbour's, and a step that cannot be verified
independently is decoration.

**The existing eleven-step form.** The base for this decision, corrected twice. `Escalate` becomes
a branch off Resolve for the reason given above. `Re-evaluate` becomes an explicit edge rather
than a step, because as prose it says only "determine whether more is needed", which is a branch
condition and not a state transition.

**Per-tier bespoke loops.** Rejected, and this is the alternative the corpus had drifted into
without deciding it: of the eight per-role loop documents, only two visibly derive from the
universal one, `ENGINEERING_LOOP.md` is a nine-step build pipeline with no delegate, escalate or
consolidate step at all, and `QA_REVIEW_INTEGRATION_LOOPS.md` has no numbered loop. Bespoke loops
per tier are simpler to write and destroy the claim that AEP is one organizational primitive
rather than a set of hardcoded stages.

## Consequences

- `docs/agents/UNIVERSAL_AGENT_LOOP.md` becomes the design projection of this decision and is
  updated from eleven steps to ten, citing this ADR.
- The eight per-role loop documents must each declare themselves a configuration of this loop and
  name their step-4 specialists, step-7 completion criteria, and step-8 artifacts. Three of them
  currently have no completion-criteria section at all.
- `docs/agents/RELEASE_LOOP.md` must be written. `ORGANIZATION.md` defines a Release tier with
  five specialists and it is the only tier with no loop document.
- `SYSTEM_ARCHITECTURE.md:46` is wrong as written and needs the return paths the principal's own
  infographic draws as "feedback loops and change requests, can originate at any stage". That is a
  gated edit under `architecture_decision` and is not made here.
- Step 7 gives per-role definition of done a home. It is currently present for two of eight tiers.
- Step 10 gives session consolidation a defined attachment point, and step 2 gives context
  hydration one.
- Nothing here is enforced. There is no runtime, no hook engine, and no agent execution. This ADR
  is `designed`, not `implemented`, and certainly not `proven`.

## Risks

- **Ten steps may still be wrong at a tier nobody has built.** Orchestration is the likely
  counterexample: `ORGANIZATION.md:69` already says "the orchestrator should be less
  improvisational than upstream analytical agents", which is an unreconciled admission that one
  tier may not delegate or analyze the way the others do. If that holds under implementation, this
  ADR is superseded rather than special-cased.
- **Ratifying a loop no code executes risks ratifying a fiction.** The mitigation is the Simple
  First Mile: prove the loop end to end on one story before building the rest of the organization
  around it.
- **The re-entry edge is the least specified part.** "Unresolved unknowns remain" is a judgement
  today with no threshold and no probe. Step 7's completion criteria are where that becomes
  testable, and they do not exist for most tiers yet.
- **Adopting the infographic's shape and the document's content pleases neither source exactly.**
  The infographic gains two boxes and the design document loses one step. Both must be updated or
  the three-version problem returns as a two-version problem.
