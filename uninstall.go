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
	outDir := installDir()

	fmt.Println("Uninstalling claude-code-preview...")

	// Remove hook scripts
	for _, name := range []string{"snapshot-file.sh", "track-changes.sh", "diff-popup.sh"} {
		path := filepath.Join(hookDir, name)
		if err := os.Remove(path); err == nil {
			fmt.Printf("  ✓ removed %s\n", path)
		}
	}

	// Remove install dir
	if err := os.RemoveAll(outDir); err == nil {
		fmt.Printf("  ✓ removed %s\n", outDir)
	}

	// Remove hooks from settings.json
	if err := removeHooksFromSettings(configDir, hookDir); err != nil {
		fmt.Printf("  ✗ could not update settings.json: %v\n", err)
	} else {
		fmt.Println("  ✓ updated settings.json")
	}

	fmt.Println("\nRemove this from your tmux.conf:")
	fmt.Printf("\n  bind P run-shell %q\n\n", filepath.Join(outDir, "preview-open.sh"))

	return nil
}

func removeHooksFromSettings(configDir, hookDir string) error {
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

	ourCommands := map[string]bool{
		filepath.Join(hookDir, "snapshot-file.sh"): true,
		filepath.Join(hookDir, "track-changes.sh"): true,
		filepath.Join(hookDir, "diff-popup.sh"):    true,
	}

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
				if ourCommands[h.Command] {
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
