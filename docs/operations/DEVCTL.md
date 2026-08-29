# `devctl` CLI

A portable client/runtime abstraction shared by humans, hooks, agents, CI, and MCP tools.

## Project commands
- `devctl init`
- `devctl adopt`
- `devctl doctor`
- `devctl status`

## Context commands
- `devctl context <work-id>`
- `devctl context explain <run-id>`

## Agent commands
- `devctl agent list`
- `devctl agent run <agent-id> <work-id>`
- `devctl agent message <run-id>`
- `devctl agent pause|resume|cancel|handoff`

## Memory commands
- `devctl memory search`
- `devctl memory related`
- `devctl memory consolidate`
- `devctl memory challenge`
- `devctl memory promote`

## Evaluation commands
- `devctl eval run <suite>`
- `devctl eval compare <baseline> <candidate>`

## Design goal
Agents should not need to know internal GitHub/DB mechanics. The CLI and MCP expose stable business-level operations.
