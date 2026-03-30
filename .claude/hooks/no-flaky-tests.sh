#!/usr/bin/env bash
# Hook: Block flaky test patterns in Go test files.
# Runs as a PreToolUse hook on Write|Edit.
#
# Blocked patterns:
#   time.Sleep     — explicit sleep, always flaky
#   <-time.After   — inline sleep disguised as channel read
#   time.Tick      — leaks timers, often used for polling loops
#   time.NewTimer  — usually a sleep in disguise in test code
#
# Use instead:
#   teatest.WaitFor       — poll TUI output for conditions
#   channels/sync         — signal when ready
#   direct model.Update   — synchronous state transitions
#   t.Deadline()          — test-aware timeouts

set -euo pipefail

INPUT=$(cat)

FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // ""')
CONTENT=$(echo "$INPUT" | jq -r '.tool_input.new_string // .tool_input.content // ""')

# Only check Go test files.
if [[ ! "$FILE_PATH" =~ _test\.go$ ]]; then
  echo '{}'
  exit 0
fi

# No content to check (e.g., a read operation).
if [[ -z "$CONTENT" ]]; then
  echo '{}'
  exit 0
fi

# Check for flaky patterns.
VIOLATIONS=()

if echo "$CONTENT" | rg -q 'time\.Sleep'; then
  VIOLATIONS+=("time.Sleep")
fi

if echo "$CONTENT" | rg -q '<-time\.After'; then
  VIOLATIONS+=("<-time.After (inline sleep)")
fi

if echo "$CONTENT" | rg -q 'time\.Tick[^e]'; then
  VIOLATIONS+=("time.Tick (leaks timers)")
fi

if echo "$CONTENT" | rg -q 'time\.NewTimer'; then
  VIOLATIONS+=("time.NewTimer")
fi

if [[ ${#VIOLATIONS[@]} -gt 0 ]]; then
  FOUND=$(IFS=', '; echo "${VIOLATIONS[*]}")
  cat <<EOF
{"decision":"block","reason":"Flaky test pattern detected: ${FOUND}. Use condition-based waiting instead: teatest.WaitFor, channels, sync primitives, or direct model.Update calls. Never rely on timing in tests."}
EOF
  exit 0
fi

echo '{}'
