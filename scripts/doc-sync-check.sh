#!/bin/bash
# PreToolUse hook: remind about doc sync before git commit
# Reads stdin JSON, checks staged files, never blocks the commit.

input=$(cat | tr -d '\r')

tool_name=$(echo "$input" | jq -r ".tool_name // empty")
tool_cmd=$(echo "$input" | jq -r ".tool_input.command // empty")

# Only act on Bash tool calls that start with "git commit"
if [ "$tool_name" != "Bash" ] || ! echo "$tool_cmd" | grep -qE "^git commit( |$)"; then
  echo '{"continue":true}'
  exit 0
fi

staged=$(git diff --cached --name-only 2>/dev/null)

has_go=false
has_handover=false
has_progress=false

if echo "$staged" | grep -q '\.go$'; then
  has_go=true
fi
if echo "$staged" | grep -qx 'HANDOVER\.md'; then
  has_handover=true
fi
if echo "$staged" | grep -q '^docs/current-progress\.md$'; then
  has_progress=true
fi

if [ "$has_go" = "true" ] && [ "$has_handover" != "true" ] && [ "$has_progress" != "true" ]; then
  jq -n --arg msg "Docs: .go files staged but HANDOVER.md/current-progress.md not updated. Consider updating per wf-doc-006." \
    '{continue: true, systemMessage: $msg}'
elif [ "$has_go" = "true" ] && ( [ "$has_handover" = "true" ] || [ "$has_progress" = "true" ] ); then
  jq -n --arg msg "Docs: Go and docs both staged. Nice work keeping docs in sync." \
    '{continue: true, systemMessage: $msg}'
else
  echo '{"continue":true}'
fi
