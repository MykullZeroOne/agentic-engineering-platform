---
id: DES-UNIVERSAL-AGENT-LOOP
type: design
tier: 2
status: draft
version: 2
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-09-06
---

# Universal Agent Loop

Every organizational tier in AEP runs the same ten-step loop; only the role
configuration changes between tiers. This document is the design projection of
ADR-023, which is the decision's source of truth — where the two disagree,
ADR-023 wins.

1. **Receive** — accept a work assignment with an explicit objective and authority
   boundary.
2. **Hydrate** — assemble the context packet: role identity, canonical knowledge,
   role and organizational memory, prior experience, procedures, pitfalls, open
   questions.
3. **Analyze** — determine knowns, unknowns, risks, missing evidence, and which
   specialists materially improve confidence.
4. **Delegate** — invoke specialist sub-agents.
5. **Collect** — gather specialist findings and synthesize them, resolving
   contradictions.
6. **Resolve** — answer questions from canonical knowledge and durable memory.
   Escalate only what cannot be answered that way.
7. **Validate** — evaluate this role's completion gate, its definition of done.
8. **Produce** — emit artifacts and the evidence for them.
9. **Hand off** — pass output and approved exceptions to the next owning role.
10. **Consolidate** — extract memory candidates from the run.

## Escalation

Escalation is a branch off step 6, not a numbered step — a numbered step would
imply every pass performs it, and the human is the most expensive reasoning
resource in the organization. It travels one hop at a time up the hierarchy:
sub-agent, parent, parent, human, per ADR-005's hierarchical-communication ladder.

## Re-entry edge

From step 6 or step 7, if unresolved unknowns remain, control returns to step 3.
This edge is the recursion; without it the ten steps are a pipeline with more
boxes.

## Return edge

A downstream role may return a completed hand-off to the role that produced it —
Review returns findings to Engineering, QA returns a failed acceptance to
Engineering, the human returns a rejected specification to the BA. A return is a
new Receive on the same run: it enters at step 1 with the returned findings as its
objective, hydrates the original run's context at step 2, and attaches its
evidence to the run that produced the hand-off, not to a fresh one. Step 10 is
terminal for a pass, never for a run; a run ends when its hand-off is accepted
downstream or the work item is cancelled.

## The human as specialist and gate owner

The human is a specialist and a gate owner, not only an escalation target. A
question whose answer exists only in the human's head — the BA's clarifying
questions, an architect's request for a product trade-off, a release lead's
go/no-go — is not one escalation resolves; a role may name the human in its
step-4 specialist set, and step 5 collects the answer like any other finding. A
role may also declare its step-7 completion gate human-owned: the gate passes on
the human's explicit approval, and on nothing the agent concludes about
completeness.

## Configuration

A tier configures this loop; it does not redefine it. A role definition supplies
the specialist set for step 4 (which may include the human), the completion
criteria for step 7 and whether the gate is human-owned, the artifact and evidence
shapes for step 8, and which roles may return work to it. Each per-role loop
document in this directory states its configuration under a heading of
`## Configuration of ADR-023`.

## Loop invariants
- Do not finalize with unresolved critical unknowns.
- Do not ask the human a question already answered by authoritative context.
- Do not silently infer a consequential product/legal/security decision.
- Specialist findings return to their owning parent.
- Every material conclusion has evidence/provenance.
