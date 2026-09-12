#!/usr/bin/env bash
# AEP hook point: pre_commit (unbound in hooks.yaml; CLAUDE.md forbids secrets, no engine yet).
# Matcher: git commit. Runs validation.docs when the agent commits from the IDE.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=aep-common.sh
source "$SCRIPT_DIR/aep-common.sh"

read_hook_input >/dev/null

if ! run_validate_docs; then
  deny_shell \
    "Documentation validation failed before commit." \
    "scripts/validate_docs.py failed. Fix violations before committing documentation changes."
fi

allow_shell
