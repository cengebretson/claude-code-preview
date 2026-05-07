#!/usr/bin/env bash

json=$(cat)
session_id=$(printf '%s' "$json" | jq -r '.session_id // ""')
file_path=$(printf '%s' "$json" | jq -r '.tool_input.file_path // ""')

[[ -z "$file_path" || -z "$session_id" ]] && exit 0

echo "$file_path" >> "/tmp/claude-changes-${session_id}"
