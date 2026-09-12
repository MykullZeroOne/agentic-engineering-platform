#!/usr/bin/env bash
# Shared helpers for AEP Cursor hook scripts. Authority for what runs when lives in
# .agentic/hooks/hooks.yaml; these scripts are the Cursor-specific adapter (ADR-020).

set -euo pipefail

AEP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$AEP_ROOT"

read_hook_input() {
  cat
}

run_validate_docs() {
  if ! python3 -c "import yaml" 2>/dev/null; then
    if [[ -f requirements-docs.txt ]]; then
      pip install -q -r requirements-docs.txt >/dev/null 2>&1 || true
    fi
  fi
  python3 scripts/validate_docs.py
}

run_check_gates() {
  local base="${1:-origin/main}"
  if git rev-parse --verify "$base" >/dev/null 2>&1; then
    python3 scripts/check_gates.py --strict --base "$base"
  else
    python3 scripts/check_gates.py --strict
  fi
}

deny_shell() {
  python3 -c 'import json,sys; json.dump({"permission":"deny","user_message":sys.argv[1],"agent_message":sys.argv[2]}, sys.stdout); sys.exit(2)' "$1" "$2"
}

allow_shell() {
  python3 -c 'import json,sys; json.dump({"permission":"allow"}, sys.stdout)'
  exit 0
}

deny_read() {
  python3 -c 'import json,sys; json.dump({"permission":"deny","user_message":sys.argv[1]}, sys.stdout); sys.exit(2)' "$1"
}

allow_read() {
  python3 -c 'import json,sys; json.dump({"permission":"allow"}, sys.stdout)'
  exit 0
}

emit_additional_context() {
  python3 -c 'import json,sys; json.dump({"additional_context": sys.stdin.read()}, sys.stdout)' <<<"$1"
}
