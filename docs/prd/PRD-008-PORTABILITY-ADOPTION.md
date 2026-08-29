---
id: PRD-008
type: prd
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

# PRD-008 — Project Bootstrap, Portability, and Legacy Adoption

## Objective
Make AEP easy to apply to new projects and incrementally adopt into existing projects.

## New project flow
`devctl init` or UI wizard creates minimal project configuration, documentation structure, GitHub workflow stubs, role bindings, and standards mappings.

## Existing project flow
`devctl adopt` scans repository layout, languages/frameworks, docs, CI, tests, issues, PR history, existing AGENTS/CLAUDE files, and generates a non-destructive adoption report.

## Adoption levels
0. Legacy
1. Agent-aware
2. Knowledge-aware
3. Process-aware
4. Memory-aware
5. Agentic/autonomous

## Requirements
- Old repositories can map existing docs directories instead of moving them.
- Central workflows/skills/hooks are versioned outside project repos.
- Project repos contain only local configuration and domain-specific extensions.
- Optional historical learning can mine prior issues/PRs/reviews for memory candidates.
