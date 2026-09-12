#!/usr/bin/env bash
# AEP hook point: before_pr — validation.docs (hard binding in hooks.yaml).
# Matcher: git push, gh pr create/ready. CI runs the same checks; this catches them when
# the agent drives git from the IDE.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=aep-common.sh
source "$SCRIPT_DIR/aep-common.sh"

read_hook_input >/dev/null

if ! run_validate_docs; then
  deny_shell \
    "Documentation validation failed (validation.docs)." \
    "scripts/validate_docs.py failed. Fix corpus violations before push or PR. Run: python3 scripts/validate_docs.py"
fi

if ! run_check_gates; then
  deny_shell \
    "An open human gate blocks this change (check_gates.py --strict)." \
    "scripts/check_gates.py reported an OPEN gate. Record approval under .agentic/approvals/ or narrow the change. Agents never close human gates (ADR-019, ISC-30)."
fi

allow_shell
