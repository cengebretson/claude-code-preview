# claude-code-preview

## Project Overview

A Go TUI application that provides a diff review pane for Claude Code. When Claude edits files, the TUI shows changed files with delta-rendered diffs. Designed to run as a persistent tmux side pane.

## Architecture

Single binary with two modes:
- **No args** — launches the bubbletea TUI (`tui.go`)
- **Subcommands** — `install`, `uninstall`, `status` for setup (`install.go`, `uninstall.go`, `status.go`)

## Key Files

- `main.go` — entry point, subcommand routing
- `tui.go` — bubbletea TUI model, view, and commands
- `install.go` — installs hook scripts and merges Claude Code `settings.json`
- `status.go` — dependency and installation health check
- `uninstall.go` — removes hooks from `settings.json` and deletes scripts
- `hooks/` — shell scripts embedded via Go `embed` and written to disk on install

## Signal Mechanism

The TUI polls `/tmp/claude-preview-signal` every 500ms. The `diff-popup.sh` Stop hook writes changed file paths to this file when it detects the `claude-preview` tmux pane title. Session ID is written to `/tmp/claude-preview-signal.session`.

Snapshots of pre-edit files live at `/tmp/claude-snapshots-{sessionID}/{escaped_path}` — written by `snapshot-file.sh` before each Edit/Write.

## Diff Rendering

Diffs are rendered by piping `git diff` through `delta` with `--file-style omit --hunk-header-style omit --color-always`. This strips file headers and hunk markers, showing only changed code. Delta picks up the user's `~/.gitconfig` theme automatically (Catppuccin Mocha). The `--side-by-side` flag is toggled at runtime via the `s` key.

## Dependencies

- `delta` — diff rendering
- `jq` — JSON parsing in hook scripts
- `tmux` — pane detection and title setting

## Building

```bash
go build -o claude-code-preview .
```

## Catppuccin Mocha Colors

All colors use the Catppuccin Mocha palette to match the user's terminal theme (tmux, starship, delta). Do not introduce colors outside this palette.
