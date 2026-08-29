# New Project Bootstrap

## Minimal project footprint
- `AGENTS.md`
- `CLAUDE.md` or provider compatibility pointer if desired
- `.agentic/project.yaml`
- `.agentic/skills/` for project-specific skills only
- `docs/prds/`
- `docs/ads/`
- `docs/adrs/`
- `docs/architecture/`
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
