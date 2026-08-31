---
id: DES-PLANNING-ORCHESTRATION-LOOP
type: design
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Planning and Orchestration

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

## Principle
Planner reasons about work structure. Orchestrator executes state transitions.
