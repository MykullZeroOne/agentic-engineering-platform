---
id: DES-TECHNOLOGY-STACK
type: design
tier: 2
status: approved
version: 1
owner: human.cto
human_approved: true
approved_by: human.cto
approved_on: 2026-08-29
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Recommended Technology Stack

## Backend/control service
**Go** is recommended for the first implementation because it fits `devctl`, worker daemons, GitHub integrations, long-running orchestration, PTY/process management, and single-binary distribution well.

Alternative: TypeScript if UI/backend velocity and shared types are prioritized over native process/runtime ergonomics.

## UI
React + TypeScript with a component-driven modular UI. Use WebSocket/SSE for agent activity and PTY streaming.

## Persistence
PostgreSQL + pgvector. Add S3-compatible artifact storage only when transcripts/logs outgrow local/simple storage.

## Integration
- GitHub App + GitHub API/webhooks
- `gh` for local/agent-friendly operations
- MCP server for knowledge/context/communication/runtime operations
- official Claude Code / Codex CLIs through adapters
- GitHub Actions for deterministic CI/control events

## Worker
Local daemon on macOS/Linux that registers runtime capabilities and launches authenticated provider CLIs in isolated worktrees.

## Architecture style
Start as a modular monolith with clearly separated packages/ports. Avoid distributed-service complexity until workload measurements require it.
