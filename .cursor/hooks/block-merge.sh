#!/usr/bin/env bash
# AEP tools.deny: github.merge on engineer.primary — agents never merge (ISC-30, POL-001 M5).

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=aep-common.sh
source "$SCRIPT_DIR/aep-common.sh"

input="$(read_hook_input)"
command="$(python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get("command",""))' <<<"$input")"

if echo "$command" | grep -qE '(^|[[:space:]])gh[[:space:]]+pr[[:space:]]+merge|gh[[:space:]]+api[[:space:]].*merges'; then
  deny_shell \
    "Agent merge is forbidden (github.merge deny, ISC-30)." \
    "engineer.primary tools.deny includes github.merge. Open a pull request and let the human merge. Merging does not close gates (ADR-019)."
fi

allow_shell
