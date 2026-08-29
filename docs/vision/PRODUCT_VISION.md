---
id: GUIDE-PRODUCT-VISION
type: guide
tier: null
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

# Product Vision

## Vision

Build an agentic engineering platform that allows one or a small number of technically capable humans to operate like the leadership layer of a disciplined software organization while specialized AI agents perform the product, analysis, architecture, planning, engineering, QA, review, and release work beneath them.

The platform should increase throughput without sacrificing engineering standards, traceability, human control, or institutional learning.

## Problem

Current coding agents are powerful but fragmented. They often:

- operate inside isolated sessions with poor continuity;
- hide ongoing work behind opaque progress indicators;
- require humans to repeatedly re-explain project context;
- lack durable organizational roles and reporting relationships;
- mix requirements, planning, architecture, implementation, and review into one agent;
- rely on prose prompts instead of enforceable policies and lifecycle hooks;
- do not preserve a high-quality causal record of why a decision was made;
- make cross-agent learning difficult;
- tie orchestration to one model provider or usage-billed platform.

Small teams already paying for Claude/Codex subscriptions should be able to compose those capabilities into a disciplined multi-agent engineering organization without immediately taking on unpredictable API spend.

## Target user

Primary: technically sophisticated solo founders and 2–10 person engineering teams that already use GitHub and one or more coding-agent subscriptions.

Secondary: teams wanting a vendor-neutral control and knowledge plane above GitHub/GitLab and multiple agent runtimes.

## Product promise

> Your engineering process belongs to you. Models are replaceable workers.

## Outcomes

A user should be able to:

- describe a product idea conversationally;
- interact with a BA agent that asks intelligent follow-up questions;
- automatically receive compliance, legal, privacy, security, and domain preflight analysis;
- approve a stable product/requirements specification;
- have agents groom, architect, decompose, and plan work into a dependency DAG;
- dispatch implementation work to the best available model/agent;
- observe every active agent and intervene directly when necessary;
- maintain rigorous policies and human approval gates;
- preserve per-role memory across ephemeral sessions;
- learn from failures, reviews, human corrections, and successful patterns;
- regression-test models, prompts, skills, context retrieval, and orchestration changes;
- migrate existing repositories incrementally.

## Non-goals

- Replace Git as source control.
- Build another IDE.
- Provide autonomous legal advice or legal approval.
- Remove human responsibility for consequential decisions.
- Force all projects into a single document format.
- Require proprietary models or APIs.
