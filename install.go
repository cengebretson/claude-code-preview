package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed hooks/*
var embeddedHooks embed.FS

type settingsHook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type settingsHookEntry struct {
	Matcher string         `json:"matcher,omitempty"`
	Hooks   []settingsHook `json:"hooks"`
}

type settings struct {
	Hooks map[string][]settingsHookEntry `json:"hooks,omitempty"`
	Extra map[string]json.RawMessage     `json:"-"`
}

func claudeConfigDir() string {
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func runInstall() error {
	configDir := claudeConfigDir()
	hookDir := filepath.Join(configDir, "hooks")
	outDir := appConfigDir()

	fmt.Println("Installing claude-code-preview...")

	// Write Claude hook script
	hookScript := filepath.Join(hookDir, "claude-code-preview.sh")
	data, err := embeddedHooks.ReadFile("hooks/claude-code-preview.sh")
	if err != nil {
		return fmt.Errorf("reading embedded claude-code-preview.sh: %w", err)
	}
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(hookScript, data, 0755); err != nil {
		return fmt.Errorf("writing claude-code-preview.sh: %w", err)
	}
	fmt.Printf("  ✓ wrote %s\n", hookScript)

	// Merge settings.json
	if err := mergeSettings(configDir, hookDir); err != nil {
		return fmt.Errorf("updating settings.json: %w", err)
	}
	fmt.Println("  ✓ updated settings.json")

	fmt.Println("\nAdd this to your tmux.conf:")
	fmt.Println("\n  bind P run-shell \"claude-code-preview tmux\"")
	fmt.Println("\nThen reload tmux: prefix+r")

	return nil
}

func mergeSettings(configDir, hookDir string) error {
	settingsPath := filepath.Join(configDir, "settings.json")

	var raw map[string]json.RawMessage
	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing settings.json: %w", err)
		}
	}
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}

	// Parse existing hooks
	var hooks map[string][]json.RawMessage
	if h, ok := raw["hooks"]; ok {
		if err := json.Unmarshal(h, &hooks); err != nil {
			hooks = nil
		}
	}
	if hooks == nil {
		hooks = make(map[string][]json.RawMessage)
	}

	hook := filepath.Join(hookDir, "claude-code-preview.sh")
	wantHooks := map[string][]settingsHookEntry{
		"PreToolUse": {
			{
				Matcher: "Edit|Write|NotebookEdit",
				Hooks:   []settingsHook{{Type: "command", Command: hook}},
			},
		},
		"PostToolUse": {
			{
				Matcher: "Edit|Write|NotebookEdit",
				Hooks:   []settingsHook{{Type: "command", Command: hook}},
			},
		},
		"Stop": {
			{
				Hooks: []settingsHook{{Type: "command", Command: hook}},
			},
		},
	}

	for event, entries := range wantHooks {
		for _, want := range entries {
			if !hookExists(hooks[event], want.Hooks[0].Command) {
				b, _ := json.Marshal(want)
				hooks[event] = append(hooks[event], b)
			}
		}
	}

	hooksJSON, err := json.Marshal(hooks)
	if err != nil {
		return err
	}
	raw["hooks"] = hooksJSON

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}

func hookExists(entries []json.RawMessage, command string) bool {
	for _, raw := range entries {
		var entry settingsHookEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		for _, h := range entry.Hooks {
			if h.Command == command {
				return true
			}
		}
	}
	return false
}
