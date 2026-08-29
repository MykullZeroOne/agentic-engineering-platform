---
id: ADR-014
type: adr
tier: 1
status: proposed
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-014 — Supervisory-Control Delivery Model

## Context

AEP needs a delivery process. Sprint-based agile does not fit: sprints assume human
implementation capacity is the scarce, hard-to-estimate resource, and in AEP the implementer is an
agent whose capacity is elastic and whose work units take minutes.

The tempting move was to coin an "agentic agile." That would have reinvented an existing field
with worse branding. One human supervising many autonomous agents, where the binding constraint is
intervention capacity, is the subject of **human supervisory control** — Sheridan and Verplank's
levels of automation, Olsen and Wood's fan-out, Crandall's neglect-time model, Cummings' wait-time
correction — and of lean product-flow theory, where Reinertsen derives the same queueing results
for product development.

That literature supplies two things prose could not: a formula for how many agents one person can
oversee, and a calibrated utilization ceiling. It also contradicts two things this project was
about to adopt.

## Decision

**Adopt supervisory control and management-by-exception as the delivery model. Do not coin a new
methodology.**

1. **Vocabulary is borrowed, not invented.** Supervisory control, management by exception,
   human-on-the-loop, human-in-command. The flow substrate is Kanban with WIP limits, cost-of-delay
   sequencing, and small batches.

2. **No sprints — but cadence and batch limits survive.** Sprints die because the time-box is the
   wrong synchronization device when agents share an event log, *not* because bounded batches stop
   mattering. Regular feedback cadence and batch-size limits remain valuable precisely because
   agent output carries high uncertainty.

3. **The unit of delivery is a cycle bounded by a decision.** A cycle opens when a specification
   closes the `product_spec` gate and ends at the next human gate or at completion. Each cycle
   carries an **appetite** (how much this is worth) and a **blast-radius bound**, so a
   mis-specified DAG cannot burn unbounded resource before reaching a gate.

4. **Human engagement is hybrid, never purely event-driven.** A ranked, batched **attention
   budget** handles gates and escalations. Separately, a periodic **supervisory review** makes no
   decisions and exists to maintain situation awareness and to detect escalation gaps.

5. **WIP is computed, not chosen.** Effective fan-out is `NT / (IT + WT) + 1`, and the human's
   busy-time fraction is held at or below roughly 70%. Both are measured, not assumed.

6. **Autonomous depth is a diagnostic, never a target, and never agent-facing.** It is published
   only alongside shadow metrics that would expose depth rising while quality falls.

7. **Gates are tiered by reversibility.** Low-risk reversible actions proceed automatically with
   sampled audit; consequential and irreversible actions require a named human. This reduces gate
   volume so that the approvals which remain are real rather than rubber-stamped.

## Alternatives considered

**Adapt Scrum or SAFe.** Rejected. Their ceremonies exist to synchronize humans who cannot read
each other's state; agents share an event log and explicit context packets. Velocity and story
points forecast human throughput, which is not the constraint.

**Coin "agentic agile" as a new methodology.** Rejected — the substance already exists under
established names with fifty years of empirical work behind it, including quantitative models this
project would otherwise have had to guess at. Inventing a term would have cost us that literature.

**Purely event-driven supervision** — pull the human in only at gates and escalations. Rejected on
evidence. This maximizes passive monitoring, which is the out-of-the-loop condition Endsley and
Kiris showed degrades takeover performance. Cummings' data makes loss of situation awareness the
*dominant* wait-time term: including wait times cut predicted operator capacity by up to 67%, and
still by 36% even in a highly automated management-by-exception system.

**Autonomous depth as the headline metric replacing velocity.** Rejected. It is a Goodhart trap
that rewards agents for not escalating — and what an agent declines to surface is invisible by
construction. The autonomous-vehicle field ran this experiment with disengagements-per-mile and
repudiated it, since the metric ignores difficulty and rewards testing in easy conditions.

**Human WIP as a fixed count of open escalations.** Rejected as too crude. The limit is a function
of per-item interaction time and wait time, so it moves as review ergonomics change. A fixed count
hides the fact that adding agents can *reduce* throughput.

## Consequences

- **Instrumentation precedes limits.** Interaction time, time-to-notice, queue wait, and human
  utilization must be measured before any WIP limit is defensible. Until then, limits are guesses.
- **`memory.consolidate` is not the retrospective.** Post-merge consolidation optimizes the loop
  the system already knows about. A separate supervisory review is required to surface what the
  loop cannot see — chiefly, decisions agents made silently that should have been questions.
- **Escalation must be free.** Any agent objective must not penalize escalating, or the system
  optimizes toward silence. A *falling* escalation rate is treated as a possible alarm.
- **Gates gain a risk tier**, which changes `.agentic/registries/gates.yaml`.
- **Review ergonomics become a first-class concern.** Shrinking interaction time and time-to-notice
  raises capacity more than adding agents does.
- Detailed in `docs/workflows/DELIVERY_MODEL.md`.

## Risks

- **Every quantitative threshold is borrowed from another domain.** Fan-out and the ~70% ceiling
  come from UAV supervision, air traffic control, and process control — not software. They are
  design heuristics to calibrate locally, not laws. Treating them as settled would be a misuse of
  the evidence.
- **No validated prior art exists** for single-human, many-agent software delivery. This model is
  extrapolation from adjacent rigorous work, so it is built to be instrumented and reversed rather
  than assumed correct.
- **Automation complacency cannot be trained away.** Parasuraman and Manzey found it in experts as
  well as novices, under multi-task load. As agents improve, the human inspects less, so the rare
  genuine defect is the one most likely to pass. Sampled deep audits and a utilization ceiling are
  mitigations, not cures.
- **The single-human failure mode is only partly mitigable here.** The standard remedy — pooled,
  rotating approval — requires a second qualified human, which a solo project does not have. Gate
  tiering reduces load, but bus factor remains one, and an overloaded sole approver risks becoming
  a "liability sponge": the appearance of oversight without its substance. This residual risk is
  accepted and named rather than designed away.
- **Skill atrophy.** A human who stops reading implementation loses the judgment the gates depend
  on. Keeping some hands-on review is a deliberate cost, not waste.
