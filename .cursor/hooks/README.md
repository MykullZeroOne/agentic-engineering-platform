# Cursor hooks — AEP adapter

Authority for **what** must run and **when** lives in `.agentic/hooks/hooks.yaml` and
`.agentic/registries/hook-points.yaml`. This directory is the **Cursor IDE adapter**: shell
hooks wired in `.cursor/hooks.json` that invoke the same scripts CI uses where possible.

The AEP hook engine (WI-0052, ADR-020) does not exist yet. Nothing here replaces it; it makes
Cursor sessions respect the bindings that already have implementations.

## Mapping

| AEP hook point | Binding | Cursor hook | Script |
| --- | --- | --- | --- |
| `before_agent_run` | `context.build` (partial) | `sessionStart` | `session-start.sh` |
| `before_pr` | `validation.docs` (hard) | `beforeShellExecution` | `before-pr.sh` |
| `pre_commit` | *(unbound; gap in hooks.yaml)* | `beforeShellExecution` | `pre-commit.sh` |
| `before_pr` | gate coverage | `beforeShellExecution` | `before-pr.sh` (`check_gates.py`) |
| `tools.deny` | `github.merge` on `engineer.primary` | `beforeShellExecution` | `block-merge.sh` |
| `pre_commit` | secret policy (partial) | `beforeReadFile` | `block-secrets.sh` |

Not mapped in Cursor (platform/CI/`devctl` only):

- `after_merge` (`work.advance_state`, `memory.consolidate`, …)
- `before_agent_run` hard checks (`work.validate_state`, `dependencies.validate`, `workspace.prepare`)
- `after_agent_run` (`validation.required_checks`, `evidence.capture`)
- PR lifecycle (`pr_opened`, `review_requested`, `changes_requested`)

## Enforcement

| Script | Implements | Exit on failure |
| --- | --- | --- |
| `before-pr.sh` | `scripts/validate_docs.py`, `scripts/check_gates.py --strict` | deny (`exit 2`) |
| `pre-commit.sh` | `scripts/validate_docs.py` | deny |
| `block-merge.sh` | `engineer.primary` `tools.deny: github.merge` | deny |
| `block-secrets.sh` | `.env`, keys, `auth.json` | deny (`failClosed: true`) |
| `session-start.sh` | Injects AEP context | always allow |

CI remains authoritative for merges; these hooks apply when the **agent** drives shell commands
from Cursor or Cloud Agents (which load project `.cursor/hooks.json`).
