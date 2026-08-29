---
id: DES-SYSTEM-ARCHITECTURE
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

# System Architecture

## Architectural model

AEP is divided into seven planes.

### 1. Experience Plane
UI, CLI, dashboards, agent workspace, human approvals, learning inbox, graph browser, project setup/adoption.

### 2. Organization Plane
Durable agent identities, roles, hierarchy, specialist sub-agents, ownership, role versions, permissions, completion criteria.

### 3. Communication Plane
Questions, answers, escalations, findings, approvals, notifications, human interventions, cross-functional handoffs.

### 4. Control Plane
GitHub issues/projects/PRs/checks/releases, work DAG, policies, hooks, state machine, routing and deterministic orchestration.

### 5. Knowledge Plane
Canonical specs, project docs, memories, graph relationships, evidence, provenance, context retrieval.

### 6. Execution Plane
Claude Code, Codex, future CLIs/APIs/local agents, worktrees, shell, MCP tools, CI/test runners.

### 7. Evaluation Plane
Immutable telemetry/event log, regression suites, model/skill/prompt/retrieval comparison, quality metrics, future training exports.

## High-level flow

Human -> BA organization -> Product approval -> Grooming -> Architecture -> Planning DAG -> Orchestrator -> Specialized engineering -> QA -> Review -> Integration -> Human release gate -> Merge/release -> Memory consolidation/evaluation.

## Foundational rule

No model/provider owns organizational state. Durable state is stored in GitHub, project knowledge, AEP persistence, and event history.
