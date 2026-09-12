#!/usr/bin/env bash
# AEP hook point: pre_commit (partial) — block reading likely secret files into the model.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=aep-common.sh
source "$SCRIPT_DIR/aep-common.sh"

input="$(read_hook_input)"
file_path="$(python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get("file_path",""))' <<<"$input")"
base="$(basename "$file_path")"

case "$base" in
  .env|.env.*|auth.json|credentials.json|*.pem|*.key|id_rsa|id_ed25519)
    deny_read "Reading $base is blocked by AEP secret policy."
    ;;
esac

case "$file_path" in
  */.codex/auth.json|*/.cursor/mcp.json)
    deny_read "Reading credential files is blocked by AEP secret policy."
    ;;
esac

allow_read
