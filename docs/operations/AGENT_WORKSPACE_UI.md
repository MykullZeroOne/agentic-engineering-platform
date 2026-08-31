---
id: DES-AGENT-WORKSPACE-UI
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

# Agent Workspace UI

## Organization dashboard
Show hierarchical teams and states: Working, Waiting, Blocked, Awaiting Human, Idle, Failed.

## Agent workspace
### Activity
Normalized timeline of important events.

### Live Session
PTY/stream for supported local runtime; human can send direct messages/interventions.

### Context
Exact context packet, source provenance, retrieved memories, requirements, policies, and files.

### Memory
Agent-specific projection: active decisions, lessons, patterns, open questions, recent episodes, confidence/supersession.

### Tools
Granted/denied capabilities, MCP servers, filesystem scope, GitHub permissions, runtime capabilities.

### Work
Issue/work unit, plan, worktree/branch, changed files, tests, PR, review/QA findings, dependencies.

## Controls
Message, Pause, Resume, Cancel, Restart, Handoff, Escalate.

## Human interventions
Every intervention is a durable event and possible memory/training signal.
