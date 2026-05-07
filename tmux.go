package main

import (
	"os/exec"
	"strings"
)

func runTmux() error {
	// If the preview pane is already open, just unzoom and return.
	out, _ := exec.Command("tmux", "list-panes", "-s", "-F", "#{pane_title}").Output()
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == paneTitle {
			exec.Command("tmux", "resize-pane", "-Z").Run()
			return nil
		}
	}

	// Unzoom the current pane if zoomed.
	zoomed, _ := exec.Command("tmux", "display-message", "-p", "#{window_zoomed_flag}").Output()
	if strings.TrimSpace(string(zoomed)) == "1" {
		exec.Command("tmux", "resize-pane", "-Z").Run()
	}

	// Record the current pane, split, then return focus.
	main, _ := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
	exec.Command("tmux", "split-window", "-h", "-l", "30%", "claude-code-preview").Run()
	exec.Command("tmux", "select-pane", "-t", strings.TrimSpace(string(main))).Run()

	return nil
}
