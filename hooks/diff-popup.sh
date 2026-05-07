#!/usr/bin/env bash

json=$(cat)
session_id=$(printf '%s' "$json" | jq -r '.session_id // ""')
changes_file="/tmp/claude-changes-${session_id}"

[[ ! -f "$changes_file" ]] && exit 0

files=$(sort -u "$changes_file")
rm -f "$changes_file"
[[ -z "$files" ]] && exit 0

files_list=$(mktemp)
echo "$files" > "$files_list"

preview_pane=$(tmux list-panes -s -F "#{pane_title} #{pane_id}" 2>/dev/null \
    | grep "^claude-preview " | awk '{print $2}' | head -1)

if [[ -n "$preview_pane" ]]; then
    signal_file="/tmp/claude-preview-signal"
    cp "$files_list" "$signal_file"
    echo "$session_id" > "${signal_file}.session"
    rm -f "$files_list"
else
    rm -f "$files_list"
fi
