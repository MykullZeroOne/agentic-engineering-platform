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

## Acceptance criteria
A new Codex/Claude/model configuration can be evaluated against an existing role regression suite before becoming the default worker.
