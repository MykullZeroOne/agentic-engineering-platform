---
id: PRD-007
type: prd
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

# PRD-007 — Evaluation, Regression, and Future Training

## Objective
Turn development history into a measurable evaluation corpus for improving models, roles, prompts, skills, retrieval, routing, and orchestration.

## Requirements
- Persist immutable events for agent runs, context retrieval, messages, tool calls, Git events, tests, reviews, human interventions, and outcomes.
- Version role definitions, skills, prompts, policies, context builder, model/runtime, and memory snapshots.
- Allow historical work to be promoted into benchmark/regression cases.
- Compare agent configurations using correctness, requirement coverage, architecture violations, test results, review findings, rework, human intervention, time, and cost/usage where measurable.
- Support regression of model, skill, prompt, retrieval, routing, and planner/orchestrator changes.
- Export structured datasets for future supervised/fine-tuning workflows without making fine-tuning an MVP dependency.

### Retrieval quality telemetry
Absorbed from Kairo's `docs/v2/23-observability` (ADR-022). Everything above measures whether a
*run* was good; none of it measures whether the *context* was, and a bad packet and a bad model
fail identically from the outside.

- Record whether a context packet was accepted or rejected by the agent that received it.
- Record which parts of a packet the agent actually used, against what it was given.
- Record retrieval reported as irrelevant, by an agent or a human.
- Record packet latency and size against its declared budget (PRD-006).

Retrieval is a configuration like a model or a prompt, and the fourth requirement above already
demands regression of retrieval changes. Without these signals there is nothing to regress
against.

## Acceptance criteria
A new Codex/Claude/model configuration can be evaluated against an existing role regression suite before becoming the default worker.

And for retrieval: a change to the context builder can be shown to have improved or degraded
packet acceptance on a fixed corpus of past runs, without re-running the agents.
