#!/usr/bin/env bash
# AEP hook point: before_agent_run (partial) — inject project context at session start.
# Maps: context.build (soft), work.validate_state reminder. Full dispatch checks remain
# with devctl run; see .agentic/hooks/hooks.yaml.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=aep-common.sh
source "$SCRIPT_DIR/aep-common.sh"

read_hook_input >/dev/null

context=$(cat <<'EOF'
AEP (Agentic Engineering Platform) session — follow the repository's configured organization:

- Agent contract: CLAUDE.md (principles, validate-before-push, work store, gates, branching per POL-001)
- Project config: .agentic/project.yaml (runtime_preferences, human_gates, work_store)
- Hook bindings (authority): .agentic/hooks/hooks.yaml — Cursor hooks here are an IDE adapter only
- Work state: .agentic/work/ only; never invent a parallel backlog in markdown
- Reading order: docs/DOCUMENTATION_INDEX.md
- Role for implementation: .agentic/roles/engineer.primary.yaml (tools.deny includes github.merge — agents never merge)
- Role for product intent: .agentic/roles/ba.primary.yaml — use devctl intent run IDEA and devctl intent resume RUN-ID [--answer TEXT]; see .cursor/rules/ba-loop.mdc
- Before pushing or opening a PR: python3 scripts/validate_docs.py and python3 scripts/check_gates.py --strict --base origin/main
- Go changes: go build ./... && go test ./...
EOF
)

emit_additional_context "$context"
exit 0
