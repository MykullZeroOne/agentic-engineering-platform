---
id: DES-CONTEXT-PACKET
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

# Context Packet Contract

A work packet should include:

## Identity
- durable agent ID;
- role/version;
- parent;
- relevant specialists;
- permissions/tools;
- completion criteria.

## Task
- work ID/type;
- objective;
- parent/epic/product intent;
- current GitHub state;
- allowed scope.

## Specifications
- requirement IDs/statements;
- acceptance criteria;
- invariants/preconditions/effects;
- applicable ADR/policy/standard constraints.

## Execution context
- relevant code/components;
- dependencies;
- required verification;
- worktree/branch details.

## Memory
- high-confidence role memories;
- validated patterns;
- relevant episodes/lessons;
- known pitfalls.

## Provenance manifest
For every summarized item, include source identifiers and retrieval reason.
