package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTempSettings(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readSettings(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return raw
}

func TestMergeSettings_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	hookDir := filepath.Join(dir, "hooks")
	writeTempSettings(t, dir, `{}`)

	if err := mergeSettings(dir, hookDir); err != nil {
		t.Fatal(err)
	}

	raw := readSettings(t, filepath.Join(dir, "settings.json"))
	if _, ok := raw["hooks"]; !ok {
		t.Fatal("expected hooks key in settings.json")
	}

	var hooks map[string][]json.RawMessage
	json.Unmarshal(raw["hooks"], &hooks)

	for _, event := range []string{"PreToolUse", "PostToolUse", "Stop"} {
		if len(hooks[event]) == 0 {
			t.Errorf("expected %s hook to be added", event)
		}
	}
}

func TestMergeSettings_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	hookDir := filepath.Join(dir, "hooks")
	writeTempSettings(t, dir, `{}`)

	// Run install twice
	mergeSettings(dir, hookDir)
	mergeSettings(dir, hookDir)

	raw := readSettings(t, filepath.Join(dir, "settings.json"))
	var hooks map[string][]json.RawMessage
	json.Unmarshal(raw["hooks"], &hooks)

	for _, event := range []string{"PreToolUse", "PostToolUse", "Stop"} {
		if len(hooks[event]) != 1 {
			t.Errorf("%s: expected 1 hook entry, got %d", event, len(hooks[event]))
		}
	}
}

func TestMergeSettings_PreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	hookDir := filepath.Join(dir, "hooks")
	writeTempSettings(t, dir, `{"theme": "dark", "editorMode": "vim"}`)

	if err := mergeSettings(dir, hookDir); err != nil {
		t.Fatal(err)
	}

	raw := readSettings(t, filepath.Join(dir, "settings.json"))
	if _, ok := raw["theme"]; !ok {
		t.Error("expected theme to be preserved")
	}
	if _, ok := raw["editorMode"]; !ok {
		t.Error("expected editorMode to be preserved")
	}
}

func TestMergeSettings_PreservesExistingHooks(t *testing.T) {
	dir := t.TempDir()
	hookDir := filepath.Join(dir, "hooks")
	writeTempSettings(t, dir, `{
		"hooks": {
			"Stop": [{"hooks": [{"type": "command", "command": "/usr/local/bin/my-hook.sh"}]}]
		}
	}`)

	if err := mergeSettings(dir, hookDir); err != nil {
		t.Fatal(err)
	}

	raw := readSettings(t, filepath.Join(dir, "settings.json"))
	var hooks map[string][]json.RawMessage
	json.Unmarshal(raw["hooks"], &hooks)

	if len(hooks["Stop"]) < 2 {
		t.Errorf("expected existing Stop hook to be preserved, got %d entries", len(hooks["Stop"]))
	}
}

func TestHookExists(t *testing.T) {
	entry := settingsHookEntry{
		Hooks: []settingsHook{{Type: "command", Command: "/path/to/hook.sh"}},
	}
	b, _ := json.Marshal(entry)
	entries := []json.RawMessage{b}

	if !hookExists(entries, "/path/to/hook.sh") {
		t.Error("expected hookExists to return true for existing command")
	}
	if hookExists(entries, "/other/hook.sh") {
		t.Error("expected hookExists to return false for missing command")
	}
}

func TestRemoveHooksFromSettings(t *testing.T) {
	dir := t.TempDir()
	hookDir := filepath.Join(dir, "hooks")
	writeTempSettings(t, dir, `{}`)

	// Install then uninstall
	mergeSettings(dir, hookDir)
	if err := removeHooksFromSettings(dir, hookDir); err != nil {
		t.Fatal(err)
	}

	raw := readSettings(t, filepath.Join(dir, "settings.json"))
	if _, ok := raw["hooks"]; ok {
		var hooks map[string][]json.RawMessage
		json.Unmarshal(raw["hooks"], &hooks)
		for _, event := range []string{"PreToolUse", "PostToolUse", "Stop"} {
			if len(hooks[event]) > 0 {
				t.Errorf("%s: expected hooks to be removed", event)
			}
		}
	}
}

func TestRemoveHooksFromSettings_PreservesOtherHooks(t *testing.T) {
	dir := t.TempDir()
	hookDir := filepath.Join(dir, "hooks")
	writeTempSettings(t, dir, `{
		"hooks": {
			"Stop": [{"hooks": [{"type": "command", "command": "/usr/local/bin/my-hook.sh"}]}]
		}
	}`)

	mergeSettings(dir, hookDir)
	removeHooksFromSettings(dir, hookDir)

	raw := readSettings(t, filepath.Join(dir, "settings.json"))
	var hooks map[string][]json.RawMessage
	json.Unmarshal(raw["hooks"], &hooks)

	if !hookExists(hooks["Stop"], "/usr/local/bin/my-hook.sh") {
		t.Error("expected existing Stop hook to be preserved after uninstall")
	}
}
