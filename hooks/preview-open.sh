#!/usr/bin/env bash

if tmux list-panes -s -F "#{pane_title}" | grep -q "^claude-preview$"; then
    tmux list-panes -F "#{window_zoomed_flag}" | grep -q "1" && tmux resize-pane -Z
    exit 0
fi

tmux list-panes -F "#{window_zoomed_flag}" | grep -q "1" && tmux resize-pane -Z

main=$(tmux display-message -p "#{pane_id}")
tmux split-window -h -l "30%" "claude-code-preview"
tmux select-pane -t "$main"
