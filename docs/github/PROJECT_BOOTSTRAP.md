---
id: DES-PROJECT-BOOTSTRAP
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

# New Project Bootstrap

## Minimal project footprint
- `AGENTS.md`
- `CLAUDE.md` or provider compatibility pointer if desired
- `.agentic/project.yaml`
- `.agentic/skills/` for project-specific skills only
- `docs/prd/`
- `docs/ads/`
- `docs/adr/`
- `docs/architecture/`
- `docs/policies/`
- `docs/standards/`
- thin `.github/workflows/` callers to centrally versioned workflows

## `devctl init`
1. Detect language/framework.
2. Register repository/project.
3. Select baseline organization template.
4. Select enabled agent roles/runtime providers.
5. Configure docs mappings.
6. Install thin GitHub workflow callers.
7. Create/validate GitHub Project fields and issue templates.
8. Run `devctl doctor`.
