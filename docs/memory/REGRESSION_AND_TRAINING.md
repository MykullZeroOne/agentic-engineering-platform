---
id: DES-REGRESSION-AND-TRAINING
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

# Regression and Future Training

## Why preserve this data
AEP should be able to answer whether a new model, skill, prompt, retrieval algorithm, role definition, or orchestration version actually improves engineering outcomes.

## Regression dimensions
- Model/runtime regression
- Role prompt regression
- Skill regression
- Context retrieval regression
- Memory consolidation regression
- Planner regression
- Routing regression
- Orchestration regression

## Historical benchmark case
A regression case contains:
- original task and approved specification;
- canonical context snapshot/reference set;
- expected behavioral outcomes;
- required tests/checks;
- architecture/policy criteria;
- scoring rubric;
- optional historical baseline result.

## Metrics
- functional correctness;
- requirement coverage;
- architecture/policy violations;
- test pass rate;
- QA/review findings;
- rework cycles;
- human interventions;
- elapsed time;
- usage/cost where available;
- context recall/precision;
- successful handoff rate.

## High-value learning triples
- bad implementation -> reviewer critique -> corrected implementation;
- vague product idea -> BA questions -> approved PRD/ADS;
- architecture proposal -> human approval/rejection -> final ADR;
- missed context -> failure -> corrected retrieval rule;
- human intervention -> changed behavior -> successful result.

## Training ladder
1. Improve retrieval.
2. Improve skills/prompts.
3. Retrieve few-shot examples from successful history.
4. Learn routing/role assignment.
5. Export curated datasets for fine-tuning or custom models if justified.
