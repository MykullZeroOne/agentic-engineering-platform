---
id: DES-DATA-MODEL
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

# Core Data Model

## PostgreSQL logical tables

### Organizations / projects
- `organizations`
- `projects`
- `repositories`
- `project_mappings`

### Organization plane
- `roles`
- `role_versions`
- `agent_identities`
- `agent_relationships`
- `agent_capabilities`
- `agent_permissions`
- `agent_runtime_preferences`

### Execution
- `work_items`
- `work_dependencies`
- `agent_runs`
- `runtime_sessions`
- `workspaces`
- `artifacts`

### Communication
- `questions`
- `answers`
- `decisions`
- `findings`
- `approvals`
- `human_interventions`
- `notifications`

### Knowledge
- `knowledge_nodes`
- `knowledge_edges`
- `memories`
- `memory_scopes`
- `memory_evidence`
- `embeddings`
- `context_packets`
- `context_items`

### Process
- `skills`
- `skill_versions`
- `policies`
- `policy_versions`
- `hooks`
- `hook_runs`

### Evaluation
- `events` (append-only)
- `regression_cases`
- `regression_suites`
- `evaluation_runs`
- `evaluation_scores`
- `routing_observations`

## Object storage
Store large transcripts, logs, diff snapshots, test artifacts, generated reports, and exported datasets by immutable URI/hash. Database records retain metadata and provenance.

## Vector data
Use pgvector for embeddings associated with validated knowledge nodes/memories and selectively indexed event/session summaries. Do not blindly embed all raw telemetry.
