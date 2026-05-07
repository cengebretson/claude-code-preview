package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func check(label, path string) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  ✓ %s\n", label)
	} else {
		fmt.Printf("  ✗ %s — not found at %s\n", label, path)
	}
}

func checkBin(name string) {
	if _, err := exec.LookPath(name); err == nil {
		fmt.Printf("  ✓ %s\n", name)
	} else {
		fmt.Printf("  ✗ %s — not in PATH\n", name)
	}
}

func runStatus() {
	configDir := claudeConfigDir()
	hookDir := filepath.Join(configDir, "hooks")
	outDir := installDir()

	fmt.Println("Dependencies:")
	checkBin("delta")
	checkBin("jq")
	checkBin("tmux")

	fmt.Println("\nHook scripts:")
	check("snapshot-file.sh", filepath.Join(hookDir, "snapshot-file.sh"))
	check("track-changes.sh", filepath.Join(hookDir, "track-changes.sh"))
	check("diff-popup.sh", filepath.Join(hookDir, "diff-popup.sh"))
	check("preview-open.sh", filepath.Join(outDir, "preview-open.sh"))

	fmt.Println("\nSettings.json hooks:")
	settingsPath := filepath.Join(configDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Printf("  ✗ could not read %s\n", settingsPath)
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Printf("  ✗ could not parse %s\n", settingsPath)
		return
	}
	events := []string{"PreToolUse", "PostToolUse", "Stop"}
	if _, ok := raw["hooks"]; !ok {
		for _, e := range events {
			fmt.Printf("  ✗ %s — no hooks configured\n", e)
		}
		return
	}
	var hooks map[string][]json.RawMessage
	json.Unmarshal(raw["hooks"], &hooks)
	scripts := map[string]string{
		"PreToolUse":  filepath.Join(hookDir, "snapshot-file.sh"),
		"PostToolUse": filepath.Join(hookDir, "track-changes.sh"),
		"Stop":        filepath.Join(hookDir, "diff-popup.sh"),
	}
	for _, event := range events {
		cmd := scripts[event]
		if hookExists(hooks[event], cmd) {
			fmt.Printf("  ✓ %s\n", event)
		} else {
			fmt.Printf("  ✗ %s — hook not found in settings.json\n", event)
		}
	}
}
