---
id: DES-DELIVERY-MODEL
type: design
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Delivery Model

How work flows and when the human engages. Implements ADR-014.

## The constraint

Agent implementation capacity is elastic. Human attention is not. Every design choice below
follows from treating **human supervisory capacity as the binding constraint** and everything else
as subordinate to it.

The corollary is uncomfortable and worth stating plainly: **adding agents can reduce throughput.**
If each unit of agent output costs human review time, more agents means a longer queue at the same
bottleneck, and queue delay grows non-linearly as the bottleneck saturates.

## Why there are no ceremonies

Scrum's four events synchronize humans who cannot read each other's state. Agents share an event
log and explicit context packets, so three of the four are already continuous mechanisms:

| Ceremony | Continuous equivalent |
| --- | --- |
| Standup | Organization dashboard (PRD-005) |
| Planning | Planner producing the work DAG |
| Review / demo | Human release gate with its evidence package |
| Retrospective | **Not** replaced — see below |

**The retrospective has no automated equivalent.** Post-merge consolidation
(`docs/memory/CONSOLIDATION.md`) derives lessons from what the system recorded. It cannot surface
what was never recorded — above all, the questions an agent decided not to ask. That gap is the
subject of the supervisory review.

## Cycles

A cycle is bounded by decisions, not by time.

- **Opens** when a specification closes the `product_spec` gate.
- **Runs** continuously: the orchestrator releases READY DAG nodes as dependencies clear. There is
  no batching by calendar.
- **Closes** at the next human gate, or at completion.

Each cycle declares two bounds before work starts:

- **Appetite** — how much this outcome is worth. Scope flexes to fit; the appetite does not flex to
  fit scope.
- **Blast radius** — the maximum resource, surface area, or irreversibility reachable before a
  human gate. A mis-specified DAG must hit a wall, not a budget.

## Human engagement

Two mechanisms, deliberately different in kind. Neither alone is sufficient.

### Attention budget — decisions

Gates and escalations arrive in a **ranked, batched queue** against a stated daily budget, rather
than as interrupts. Ranking is by cost of delay: what is blocked, how much, and how irreversibly.

Batching is deliberate. Interrupt-driven work is completed faster but at measurably higher stress
and effort, and unbounded interruption is what drives a supervisor past the utilization ceiling.

### Supervisory review — situation awareness

A short, periodic review that **makes no decisions**. Its purpose is the thing event-driven
supervision destroys: knowing what is going on well enough to intervene competently when something
does need a decision.

It covers:

- what the agents are doing and where the graph actually is;
- **the non-escalation audit** — a sample of decisions agents resolved autonomously, read with one
  question: *should this have been escalated?*
- whether the escalation rate is falling, and whether that is competence or silence.

Purely event-driven supervision is rejected for a specific, evidenced reason: it maximizes passive
monitoring, and passive monitoring degrades the situation awareness that makes intervention
effective. In supervisory-control studies, time-to-*notice* dominates time-to-*decide*.

## Sizing the work in progress

WIP is computed from measurement, not chosen by preference.

```
effective fan-out  =  NT / (IT + WT_notice + WT_queue)  +  1

  NT          neglect time    — how long work proceeds unattended before quality decays
  IT          interaction time — human time to service one gate or escalation
  WT_notice   time before the human realises attention is needed
  WT_queue    time an item waits while the human services something else
```

Two consequences:

- **Wait time is in the denominator.** Shrinking time-to-notice — better dashboards, better
  ranking, evidence-backed PRs that are fast to judge — raises capacity more than adding agents.
- **Utilization has a ceiling.** Hold the human's busy-time fraction at or below **~70%**. Above
  it, queue length and latency rise sharply and judgment quality falls. Sustainable gate throughput
  is roughly `0.7 × available_minutes / IT`.

These numbers come from UAV supervision and air traffic control, not software. **Calibrate them
locally.** They are a starting point that must be replaced with measured values, and the
instrumentation to do that does not exist yet.

## Metrics

**Autonomous depth** — how far the DAG advances before requiring a human — is the natural health
signal. It is used as a **diagnostic only**.

It is never a target and never visible to an agent's objective, because optimizing it rewards not
escalating, and under-escalation is invisible by construction. Every publication of depth carries
shadow metrics that would contradict it:

| Signal | Reads as trouble when |
| --- | --- |
| Autonomous depth | rises while any shadow metric worsens |
| Escape-defect rate | rises |
| Rollback rate | rises |
| Human-override rate | rises — the human is undoing agent decisions |
| Non-escalation audit findings | rise — agents are deciding what they should ask |
| Escalation rate | *falls* without a quality improvement to explain it |

The last row is the important one. A falling escalation rate is treated as a possible alarm, not
an achievement.

## Gate tiering

Not every decision deserves the same human. Tiering keeps gate volume low enough that the
approvals which remain are genuine.

| Tier | Character | Handling |
| --- | --- | --- |
| Reversible | Cheap to undo, bounded blast radius | Proceeds automatically; sampled audit |
| Consequential | Costly to undo, or crosses a policy boundary | Enters the attention budget |
| Irreversible | Production, data loss, legal, spend | Named human signs; never batched, never sampled |

Gate IDs live in `.agentic/registries/gates.yaml`. **Tiers are not yet applied there**: ADR-014 is
`proposed`, and adding a tier to every gate is a `platform_config` change that waits on its
acceptance. Until then every gate behaves as `consequential`.

Principle 1 is unchanged at every tier: an agent never closes a gate.

## Known limits of this model

- **The thresholds are borrowed.** Every number here comes from another domain. Until AEP measures
  its own, they are heuristics.
- **No validated prior art exists** for single-human, many-agent software delivery. This model is
  extrapolated from adjacent rigorous work and is built to be instrumented and revised.
- **Complacency is not solvable by discipline.** It appears in experts under load and resists
  training. Sampled deep audits of work that *looks fine* are the mitigation, and they only work if
  the human has slack to perform them — another reason for the utilization ceiling.
- **Bus factor is one.** The standard mitigation is pooled, rotating approval, which requires a
  second qualified human. A solo project cannot do this. Tiering reduces the load; it does not
  remove the single point of judgment. Recorded in ADR-014 as accepted residual risk.
- **Instrumentation does not exist.** Interaction time, time-to-notice, and queue wait are not
  captured today, so no WIP limit here is yet defensible. This is the first thing to build.
