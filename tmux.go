package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runTmux() error {
	// If the preview pane is already open, just unzoom and return.
	out, err := exec.Command("tmux", "list-panes", "-s", "-F", "#{pane_title} #{pane_id}").Output()
	if err != nil {
		return fmt.Errorf("tmux list-panes: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == paneTitle {
			paneID := parts[1]
			// Verify the pane is still alive — a force-closed TUI may leave a stale title.
			if exec.Command("tmux", "display-message", "-t", paneID, "-p", "#{pane_id}").Run() == nil {
				exec.Command("tmux", "resize-pane", "-Z").Run()
				return nil
			}
			// Pane is dead — fall through and open a new one.
		}
	}

	// Unzoom the current pane if zoomed.
	zoomed, _ := exec.Command("tmux", "display-message", "-p", "#{window_zoomed_flag}").Output()
	if strings.TrimSpace(string(zoomed)) == "1" {
		exec.Command("tmux", "resize-pane", "-Z").Run()
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	main, err := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
	if err != nil {
		return fmt.Errorf("tmux display-message: %w", err)
	}

	cfg := loadConfig()
	width := fmt.Sprintf("%d%%", cfg.paneWidth)
	if err := exec.Command("tmux", "split-window", "-h", "-l", width, self).Run(); err != nil {
		return fmt.Errorf("tmux split-window: %w", err)
	}

	exec.Command("tmux", "select-pane", "-t", strings.TrimSpace(string(main))).Run()
	return nil
}
