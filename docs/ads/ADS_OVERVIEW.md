---
id: DES-ADS-OVERVIEW
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

# Agent Development Specification (ADS)

## Purpose

The ADS is the machine-oriented representation of approved product intent. PRDs and ADRs remain human-readable views, but agents should consume a normalized specification graph that reduces ambiguity and supports automated decomposition, validation, traceability, and evidence.

## Core entities
- Intent
- Actor
- Capability
- Requirement
- Invariant
- Precondition
- Effect
- Acceptance Criterion
- Verification Requirement
- Risk
- Constraint
- Architecture Rule
- Work Unit
- Evidence

## Relationship model

`Intent -> Capability -> Requirement -> Acceptance Criterion -> Verification -> Evidence`

Architecture overlays the graph:

`Requirement -> governed_by -> ArchitectureRule/ADR`

Execution overlays the graph:

`Requirement -> implemented_by -> WorkUnit -> PR/Commit`

## Human and agent projections
- PRD: narrative product view.
- ADR: narrative architectural decision.
- ADS: structured machine view.
- GitHub issues: execution view.
- Evidence matrix: verification view.

## Readiness
An ADS is implementation-ready when all material requirements are unambiguous, testable, scoped, linked to applicable constraints, and either have no unresolved blockers or explicitly approved exceptions.
