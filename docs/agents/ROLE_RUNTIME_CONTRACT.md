---
id: DES-ROLE-RUNTIME-CONTRACT
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

# Role / Agent / Runtime Contract

## Role
Defines organizational responsibility: e.g. `ios-engineer`, `compliance-analyst`.

## Agent identity
A durable instance of a role: e.g. `ios.engineer.primary`, `compliance.primary`.

Contains:
- parent identity;
- specialist children;
- role version;
- permissions;
- skills;
- tools;
- memory namespace;
- routing preferences;
- completion criteria;
- active work/history.

## Runtime session
Transient process/session that executes the agent identity using Claude, Codex, future model, or local agent.

## Runtime adapter interface
- `start(workPacket)`
- `send(message)`
- `pause()`
- `resume()`
- `cancel()`
- `restart()`
- `handoff(targetRuntime)`
- `status()`
- `events()`
- `transcript()`

## Handoff
A handoff supplies durable identity + current work + concise session summary + relevant memory + worktree state + failed attempts + test state. Full transcript is available as evidence but is not blindly injected.
