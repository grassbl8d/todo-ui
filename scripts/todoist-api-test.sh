#!/usr/bin/env bash
#
# Run the live Todoist API integration guard. These tests hit the real Todoist
# server to verify the endpoints todo-ui depends on still work — the integration
# most likely to break. They are behind the `integration` build tag, so the
# normal `go test ./...` never runs them.
#
# Usage:
#   scripts/todoist-api-test.sh                 # read-only checks
#   TODOUI_INTEGRATION_WRITE=1 scripts/todoist-api-test.sh   # + create/delete round-trip
#
# Token resolution (so you never have to `export` it by hand): if
# $TODOIST_API_TOKEN is unset, it is read from your app config and exported for
# the test run, in the same order the app itself looks:
#   1. ~/.config/todo-ui/config.json   (todo-ui's own config)
#   2. ~/.config/todoui/config.json    (legacy, pre-rename)
#   3. ~/.config/todoist/config.json   (shared sachaos CLI config)
# The test suite pins HOME to a temp dir for isolation, so it can't read these
# files itself — exporting the token here is what bridges that gap.
#
set -euo pipefail

cd "$(dirname "$0")/.."

# token_from <config.json> — print the "token" field, or nothing.
token_from() {
  local f="$1"
  [ -f "$f" ] || return 0
  if command -v jq >/dev/null 2>&1; then
    jq -er '.token // empty' "$f" 2>/dev/null || true
  elif command -v python3 >/dev/null 2>&1; then
    python3 -c 'import json,sys
try:
    print(json.load(open(sys.argv[1])).get("token","") or "")
except Exception:
    pass' "$f" 2>/dev/null || true
  else
    # Fallback: pull the first "token":"…" value with sed.
    sed -n 's/.*"token"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$f" | head -n1
  fi
}

if [ -z "${TODOIST_API_TOKEN:-}" ]; then
  for cfg in \
    "$HOME/.config/todo-ui/config.json" \
    "$HOME/.config/todoui/config.json" \
    "$HOME/.config/todoist/config.json"; do
    tok="$(token_from "$cfg")"
    if [ -n "$tok" ]; then
      export TODOIST_API_TOKEN="$tok"
      echo "==> using token from $cfg"
      break
    fi
  done
fi

if [ -z "${TODOIST_API_TOKEN:-}" ]; then
  echo "integration: no Todoist token found." >&2
  echo "  set TODOIST_API_TOKEN=… or log in via the app first." >&2
  exit 1
fi

echo "==> go test -tags integration -run Integration"
go test -tags integration -run Integration -v -count=1 ./internal/todoui
