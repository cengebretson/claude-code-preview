# claude-code-preview

## Project Overview

A Go TUI application that provides a diff review pane for Claude Code. When Claude edits files, the TUI shows changed files with delta-rendered diffs. Designed to run as a persistent tmux side pane.

## Architecture

Single binary with two modes:
- **No args** — launches the bubbletea TUI (`tui.go`)
- **Subcommands** — `install`, `uninstall`, `status` for setup (`install.go`, `uninstall.go`, `status.go`)

## Key Files

- `main.go` — entry point, subcommand routing
- `tui.go` — bubbletea TUI model, view, update, and commands
- `config.go` — config file loading, `loadTheme()`, `appConfigDir()`
- `install.go` — installs hook scripts and merges Claude Code `settings.json`
- `status.go` — dependency and installation health check
- `uninstall.go` — removes hooks from `settings.json` and deletes scripts
- `hooks/` — shell scripts embedded via Go `embed` and written to disk on install

## Signal Mechanism

The TUI polls `/tmp/claude-preview-signal` every 500ms. The `diff-popup.sh` Stop hook writes changed file paths to this file when it detects the `claude-preview` tmux pane title. Session ID is written to `/tmp/claude-preview-signal.session`.

Snapshots of pre-edit files live at `/tmp/claude-snapshots-{sessionID}/{escaped_path}` — written by `snapshot-file.sh` before each Edit/Write.

## Diff Rendering

Diffs are rendered by piping `git diff` through `delta` with `--file-style omit --hunk-header-style omit`. This strips file headers and hunk markers, showing only changed code. Delta picks up the user's `~/.gitconfig` theme automatically. The `--side-by-side` flag is toggled at runtime via the `s` key.

## Theme System

Colors are defined in a `Theme` struct in `tui.go`. The default is `CatppuccinMocha`. At startup, `runTUI()` calls `loadTheme()` (in `config.go`) which reads `~/.config/claude-code-preview/config.json` if it exists. Missing fields fall back to `CatppuccinMocha` — partial overrides are supported.

`newStyles(t Theme)` builds a `styles` struct of lipgloss styles from the theme. `defaultStyles` is a package-level var set in `runTUI()` before the program starts.

To add a new built-in theme: define a new `var` of type `Theme` alongside `CatppuccinMocha` in `tui.go`.

## Editor

`preferredEditor()` checks `$VISUAL`, then `$EDITOR`, then falls back to `nvim`. Used when the user hits `enter` on a file.

## Config Directory

Everything lives in `~/.config/claude-code-preview/` (or `$XDG_CONFIG_HOME/claude-code-preview/`):

- `config.json` — theme colors and poll interval (user-created, optional)
- `preview-open.sh` — written by `install`, referenced in `tmux.conf`

`installDir()` is gone — `appConfigDir()` is used everywhere.

## Config File

```json
{
  "theme": {
    "green":    "#a6e3a1",
    "red":      "#f38ba8",
    "mauve":    "#cba6f7",
    "overlay1": "#7f849c",
    "surface0": "#313244",
    "yellow":   "#f9e2af",
    "peach":    "#fab387"
  },
  "poll_ms": 500,
  "pane_width": 40,
  "popup_editor": true
}
```

All fields are optional — missing fields fall back to defaults. `poll_ms` controls how often the TUI checks for new Claude changes (default 500ms).

## Dependencies

- `delta` — diff rendering
- `jq` — JSON parsing in hook scripts
- `tmux` — pane detection and title setting

## Catppuccin Mocha Colors

All default colors use the Catppuccin Mocha palette. Do not introduce colors outside this palette unless adding a new theme.

| Role     | Hex       |
|----------|-----------|
| green    | `#a6e3a1` |
| red      | `#f38ba8` |
| mauve    | `#cba6f7` |
| overlay1 | `#7f849c` |
| surface0 | `#313244` |
| yellow   | `#f9e2af` |
| peach    | `#fab387` |

## Building

```bash
go build -o claude-code-preview .
```

## Testing

```bash
go test ./...
```

Tests cover `mergeSettings`, `removeHooksFromSettings`, and `hookExists` in `install_test.go`.

## Future Enhancements

### Signal / Architecture
- **Named pipe (FIFO) signal** — replace the polling + signal file approach with a FIFO at `/tmp/claude-preview-signal.fifo`. The TUI creates it on startup and blocks on a read goroutine; the hook writes to it for instant delivery with no polling overhead. Main tradeoff is complexity around FIFO lifecycle (create on start, recreate on error). Current file polling is fine for single-user local use.
- **Multiple session support** — if two Claude sessions finish simultaneously the signal file gets clobbered; a queue or append-based approach would handle concurrent sessions cleanly
- **`fsnotify` instead of polling** — watch the signal file path with `github.com/fsnotify/fsnotify` rather than ticking every 500ms; pairs well with the FIFO idea

### UX
- **Persistent side-by-side preference** — `s` toggle resets each session; save to `config.json` so it survives restarts
- **Remember scroll position per file** — navigating away and back resets the diff viewport to top; store per-file offsets in the model
- **Dismiss a file** — `d` to remove a file from the list without undoing it
- **Total diff summary** — show aggregate `+N -N` across all files in the header
- **tmux notification on new changes** — `tmux display-message` or visual bell when a signal arrives while the TUI is in waiting state, so you know to unzoom
- **Auto-unzoom on signal** — when new changes arrive, check `#{window_zoomed_flag}` and call `tmux resize-pane -Z` to unzoom the window automatically; note that `-Z` is a window-level toggle so it should unzoom whichever pane is currently zoomed

### Git Integration
- **Stage changes** — `a` to `git add` the current file directly from the TUI
- **Jump to hunk** — `n`/`p` to navigate between diff hunks within a file
