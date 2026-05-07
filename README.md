# claude-code-preview

A TUI diff review pane for [Claude Code](https://claude.ai/code). When Claude edits files, a tmux side pane shows the changed files alongside a syntax-highlighted delta diff. Navigate files with arrow keys, open in your editor, or undo Claude's edits directly from the pane.

## Features

- File list with `+/-` change counts and file type icons
- Scrollable delta diff preview
- Mouse click to select files, scroll wheel to navigate diff
- `u` to restore a file to its pre-edit state
- `U` to restore all edited files
- `s` to toggle side-by-side diff
- `y` to copy file path to clipboard
- Polls for new changes automatically — stays live across multiple Claude responses
- Opens files in `$VISUAL` / `$EDITOR` (default: nvim)
- Themeable via `~/.config/claude-code-preview/config.json`

## Requirements

- [tmux](https://github.com/tmux/tmux)
- [delta](https://github.com/dandavison/delta)
- `jq`

## Install

```bash
go install github.com/cengebretson/claude-code-preview@latest
claude-code-preview install
```

Then add the tmux binding printed by `install` to your `tmux.conf` and reload.

## Usage

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous file |
| `↓` / `j` | Next file |
| `enter` | Open in `$VISUAL` / `$EDITOR` |
| `u` | Restore current file from snapshot |
| `U` | Restore all files from snapshots |
| `s` | Toggle side-by-side diff |
| `y` | Copy file path to clipboard |
| `r` | Refresh diff for current file |
| `q` | Clear / quit |
| `?` | Show keybindings |

Mouse click selects a file; scroll wheel moves the diff pane.

## Editor

The file opener respects `$VISUAL` first, then `$EDITOR`, then falls back to `nvim`. Set either variable in your shell config to use a different editor.

## Configuration

Create `~/.config/claude-code-preview/config.json` to customize behavior. All fields are optional and fall back to defaults.

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
  "pane_width": 40
}
```

`poll_ms` controls how often the TUI checks for new changes from Claude (default: 500ms).

## Diff Rendering

Diffs are rendered by [delta](https://github.com/dandavison/delta) using `--file-style omit --hunk-header-style omit` to strip file headers and hunk markers, showing only changed lines. Delta reads your `~/.gitconfig` theme automatically. Delta reads your existing `~/.gitconfig` theme automatically, so colors match your current setup.

## How It Works

`claude-code-preview install` adds three hooks to Claude Code's `settings.json`:

1. **PreToolUse** — snapshots each file before Claude edits it
2. **PostToolUse** — records edited file paths
3. **Stop** — signals the TUI with the list of changed files

Open the side pane with your tmux binding (`prefix+P` by default). Use `prefix+z` to zoom your main Claude pane full screen and unzoom to review changes.

## Commands

```bash
claude-code-preview            # launch TUI
claude-code-preview install    # install hooks and scripts
claude-code-preview status     # check dependencies and installation
claude-code-preview uninstall  # remove hooks and scripts
```
