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

	fmt.Println("Dependencies:")
	checkBin("delta")
	checkBin("jq")
	checkBin("tmux")

	fmt.Println("\nHook scripts:")
	check("claude-code-preview.sh", filepath.Join(hookDir, "claude-code-preview.sh"))

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
	hook := filepath.Join(hookDir, "claude-code-preview.sh")
	scripts := map[string]string{
		"PreToolUse":  hook,
		"PostToolUse": hook,
		"Stop":        hook,
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
