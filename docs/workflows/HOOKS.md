---
id: DES-HOOKS
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

# Lifecycle Hooks

## Hook philosophy
Skills say **how**. Policies say **what is required**. Hooks say **when it must happen**.

## Agent/runtime hooks
### after_workspace_create
- create/verify isolated worktree;
- install/validate environment where safe;
- record workspace identity.

### before_agent_run
- validate work state and dependencies;
- load durable agent identity;
- build context packet;
- attach active questions/decisions;
- select required skills;
- mark run active.

### after_agent_run
- capture output/events;
- execute configured local validation;
- update evidence;
- determine next state;
- create session-consolidation event.

### before_workspace_remove
- preserve uncommitted evidence if needed;
- finalize telemetry;
- clean workspace.

## Git/PR hooks
- pre-commit: formatting/lint/secret checks where appropriate.
- before PR: traceability, tests, required artifact checks.
- PR opened/updated: CI and policy validation.
- review requested: independent reviewer dispatch.
- changes requested: repair routing.
- PR merged: memory consolidation, DAG recalculation, evaluation capture.

## Hard vs soft
Hard hooks block lifecycle progression when they fail. Soft hooks enrich memory/telemetry and should not prevent emergency work unless policy says otherwise.
