package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func runUninstall() error {
	configDir := claudeConfigDir()
	hookDir := filepath.Join(configDir, "hooks")

	fmt.Println("Uninstalling claude-code-preview...")

	// Remove hook script
	hookScript := filepath.Join(hookDir, "claude-code-preview.sh")
	if err := os.Remove(hookScript); err == nil {
		fmt.Printf("  ✓ removed %s\n", hookScript)
	}

	// Remove hooks from settings.json
	if err := removeHooksFromSettings(configDir); err != nil {
		fmt.Printf("  ✗ could not update settings.json: %v\n", err)
	} else {
		fmt.Println("  ✓ updated settings.json")
	}

	fmt.Println("\nRemove this from your tmux.conf:")
	fmt.Println("\n  bind P run-shell \"claude-code-preview tmux\"")

	return nil
}

func removeHooksFromSettings(configDir string) error {
	settingsPath := filepath.Join(configDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	hooksRaw, ok := raw["hooks"]
	if !ok {
		return nil
	}

	var hooks map[string][]json.RawMessage
	if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
		return err
	}

	ourScript := expandTilde(hookCommand(configDir))

	for event, entries := range hooks {
		var kept []json.RawMessage
		for _, raw := range entries {
			var entry settingsHookEntry
			if err := json.Unmarshal(raw, &entry); err != nil {
				kept = append(kept, raw)
				continue
			}
			// Keep entries that don't reference our commands
			hasOurs := false
			for _, h := range entry.Hooks {
				if expandTilde(h.Command) == ourScript {
					hasOurs = true
					break
				}
			}
			if !hasOurs {
				kept = append(kept, raw)
			}
		}
		if len(kept) == 0 {
			delete(hooks, event)
		} else {
			hooks[event] = kept
		}
	}

	hooksJSON, err := json.Marshal(hooks)
	if err != nil {
		return err
	}
	if len(hooks) == 0 {
		delete(raw, "hooks")
	} else {
		raw["hooks"] = hooksJSON
	}

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}
