#!/usr/bin/env bash

json=$(cat)
session_id=$(printf '%s' "$json" | jq -r '.session_id // ""')
file_path=$(printf '%s' "$json" | jq -r '.tool_input.file_path // ""')

[[ -z "$file_path" || -z "$session_id" ]] && exit 0
[[ ! -f "$file_path" ]] && exit 0

snap_dir="/tmp/claude-snapshots-${session_id}"
mkdir -p "$snap_dir"
snap_name="${file_path//\//_}"
cp "$file_path" "$snap_dir/$snap_name"
