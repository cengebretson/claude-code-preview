package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type configFile struct {
	Theme       configTheme `json:"theme"`
	PollMs      int         `json:"poll_ms"`
	PaneWidth   int         `json:"pane_width"`
	PopupEditor bool        `json:"popup_editor"`
}

type configTheme struct {
	Green    string `json:"green"`
	Red      string `json:"red"`
	Mauve    string `json:"mauve"`
	Overlay1 string `json:"overlay1"`
	Surface0 string `json:"surface0"`
	Yellow   string `json:"yellow"`
	Peach    string `json:"peach"`
}

func appConfigDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "claude-code-preview")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude-code-preview")
}

type resolvedConfig struct {
	theme       Theme
	pollRate    time.Duration
	paneWidth   int
	popupEditor bool
}

func loadConfig() resolvedConfig {
	cfg := resolvedConfig{
		theme:       CatppuccinMocha,
		pollRate:    500 * time.Millisecond,
		paneWidth:   40,
		popupEditor: true,
	}

	data, err := os.ReadFile(filepath.Join(appConfigDir(), "config.json"))
	if err != nil {
		return cfg
	}

	var f configFile
	if err := json.Unmarshal(data, &f); err != nil {
		return cfg
	}

	// Apply any explicitly set theme colors over the Mocha defaults.
	t := CatppuccinMocha
	if f.Theme.Green != "" {
		t.Green = lipgloss.Color(f.Theme.Green)
	}
	if f.Theme.Red != "" {
		t.Red = lipgloss.Color(f.Theme.Red)
	}
	if f.Theme.Mauve != "" {
		t.Mauve = lipgloss.Color(f.Theme.Mauve)
	}
	if f.Theme.Overlay1 != "" {
		t.Overlay1 = lipgloss.Color(f.Theme.Overlay1)
	}
	if f.Theme.Surface0 != "" {
		t.Surface0 = lipgloss.Color(f.Theme.Surface0)
	}
	if f.Theme.Yellow != "" {
		t.Yellow = lipgloss.Color(f.Theme.Yellow)
	}
	if f.Theme.Peach != "" {
		t.Peach = lipgloss.Color(f.Theme.Peach)
	}
	cfg.theme = t

	if f.PollMs > 0 {
		cfg.pollRate = time.Duration(f.PollMs) * time.Millisecond
	}
	if f.PaneWidth > 0 {
		cfg.paneWidth = f.PaneWidth
	}
	cfg.popupEditor = f.PopupEditor

	return cfg
}
