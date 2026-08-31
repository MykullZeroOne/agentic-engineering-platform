---
id: DES-COMPONENTS
type: design
tier: 2
status: approved
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
approval_record: APR-0001
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Component Architecture

## Control service
Owns agent/work state machine, routing, hook execution, approvals, and GitHub synchronization.

## Runtime manager
Launches and manages agent runtime adapters. Provides start, message, pause, resume, cancel, restart, handoff, transcript, and status operations.

## Agent registry
Stores durable agent identities, role versions, hierarchy, specialists, skills, tools, policies, memory namespace, and provider preferences.

## Context service
Builds context packets using explicit references, graph traversal, semantic retrieval, role relevance, confidence, and policy.

## Memory service
Stores consolidated memories, graph entities/edges, memory ownership/visibility, confidence, supersession, and evidence.

## Event/evaluation service
Stores immutable events and materialized analytics projections. Creates regression cases and compares configurations.

## GitHub adapter
Maps GitHub issues/projects/PRs/checks/releases into AEP work state and invokes GitHub actions/API/`gh` where appropriate.

## Skills registry
Versioned role-independent and role-specific Agent Skills plus scripts.

## Hook engine
Executes deterministic lifecycle behavior at agent, work-item, PR, CI, review, merge, and memory boundaries.

## UI/API
Modular web application exposing organization view, work graph, projects, agents, live runtimes, memory, context, learning inbox, evaluation and configuration.
