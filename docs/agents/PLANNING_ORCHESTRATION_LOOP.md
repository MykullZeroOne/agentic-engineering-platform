---
id: DES-PLANNING-ORCHESTRATION-LOOP
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

# Planning and Orchestration

Each role below is a configuration of the universal agent loop (ADR-023,
`UNIVERSAL_AGENT_LOOP.md`).

## Planner objective
Convert approved requirements and architecture into an executable dependency DAG.

Each work unit includes:
- ID and objective;
- owning requirement(s);
- prerequisites;
- affected components;
- required capabilities;
- context references;
- expected outputs;
- verification requirements;
- risk and human-gate requirements.

### Configuration of ADR-023 (Planner)
- **Step-4 specialists**: Dependency Planner, Parallelization Planner,
  Complexity/Risk Estimator, Release Sequencing Analyst.
- **Step-7 gate**: agent-evaluated. Passes when every work unit in the graph
  carries the fields listed above: ID, objective, owning requirement(s),
  prerequisites, affected components, required capabilities, context
  references, expected outputs, verification requirements, and risk/human-gate
  requirements.
- **Step-8 artifacts**: the executable dependency DAG (work graph).
- **Return sources**: none documented.

## Orchestrator objective
Execute an approved DAG without inventing new product/architecture decisions.

Pseudo-loop:
1. Recalculate READY nodes.
2. Filter workers by capability, policy, runtime availability, and project preferences.
3. Select worker using configured routing policy and performance history.
4. Execute pre-run hooks and dispatch.
5. Observe state/events.
6. On completion, run post-run validation.
7. Update GitHub/work graph.
8. Retry, hand off, block, or escalate according to deterministic rules.
9. Repeat until graph reaches completion or human gate.

### Configuration of ADR-023 (Orchestrator)
- **Step-4 specialists**: none. The Orchestrator dispatches work units to
  owning lead roles; that is step 9 Hand off, not step 4 Delegate, which is the
  mismatch ADR-023's Risks section names as the likely counterexample to the
  ten-step loop.
- **Step-7 gate**: agent-evaluated. Passes per the deterministic post-run
  validation and retry/handoff/block/escalate rules in the pseudo-loop above.
- **Step-8 artifacts**: updated GitHub/work graph state.
- **Return sources**: any lead role whose work unit fails its gate.

## Principle
Planner reasons about work structure. Orchestrator executes state transitions.
