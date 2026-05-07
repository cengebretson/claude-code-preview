#!/usr/bin/env bash

json=$(cat)
event=$(printf '%s' "$json" | jq -r '.hook_event_name // ""')
session_id=$(printf '%s' "$json" | jq -r '.session_id // ""')

case "$event" in
  PreToolUse)
    file_path=$(printf '%s' "$json" | jq -r '.tool_input.file_path // ""')
    [[ -z "$file_path" || -z "$session_id" || ! -f "$file_path" ]] && exit 0
    snap_dir="/tmp/claude-snapshots-${session_id}"
    mkdir -p "$snap_dir"
    snap_name="${file_path//\//_}"
    # Only snapshot once per session — preserve the pre-edit original, not the state before the last edit.
    [[ ! -f "$snap_dir/$snap_name" ]] && cp "$file_path" "$snap_dir/$snap_name"
    ;;

  PostToolUse)
    file_path=$(printf '%s' "$json" | jq -r '.tool_input.file_path // ""')
    [[ -z "$file_path" || -z "$session_id" ]] && exit 0
    echo "$file_path" >> "/tmp/claude-changes-${session_id}"
    ;;

  Stop)
    changes_file="/tmp/claude-changes-${session_id}"
    [[ ! -f "$changes_file" ]] && exit 0
    files=$(sort -u "$changes_file")
    rm -f "$changes_file"
    [[ -z "$files" ]] && exit 0

    preview_pane=$(tmux list-panes -s -F "#{pane_title} #{pane_id}" 2>/dev/null \
        | grep "^claude-preview " | awk '{print $2}' | head -1)
    [[ -z "$preview_pane" ]] && exit 0

    signal_file="/tmp/claude-preview-signal"
    echo "$files" > "$signal_file"
    echo "$session_id" > "${signal_file}.session"
    ;;
esac
