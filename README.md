# claude-code-preview

A TUI diff review pane for [Claude Code](https://claude.ai/code). When Claude edits files, a tmux side pane shows the changed files with syntax-highlighted diffs. Navigate files, open them in your editor, or undo Claude's edits without leaving the terminal.

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

## Workflow

1. Start a Claude Code session in tmux as normal
2. Press `prefix+P` to open the preview pane alongside your Claude session
3. Use `prefix+z` to zoom Claude full screen while it works
4. When Claude finishes editing, unzoom (`prefix+z` again) — the preview pane will show the changed files and diffs
5. Navigate with `↑`/`↓`, review diffs, hit `u` to undo a file or `enter` to open it in your editor
6. Press `q` to clear the file list and return to the waiting state

The pane stays open and updates automatically across multiple Claude responses — no need to reopen it.

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

Mouse click selects a file; scroll wheel moves the diff pane. `enter` opens the file in `$VISUAL`, `$EDITOR`, or `nvim` as a fallback.

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

`poll_ms` controls how often the TUI checks for new changes from Claude (default: 500ms). `pane_width` sets the width of the preview pane as a percentage of the terminal width (default: 40). Set `popup_editor` to `false` to open files directly in the preview pane instead of a tmux popup (default: `true`).

## Diff Rendering

Diffs are rendered by [delta](https://github.com/dandavison/delta) using `--file-style omit --hunk-header-style omit` to strip file headers and hunk markers, showing only changed lines. Delta reads your existing `~/.gitconfig` theme automatically, so colors match your current setup.

## How It Works

`claude-code-preview install` adds three hooks to Claude Code's `settings.json`:

1. **PreToolUse** — snapshots each file before Claude edits it
2. **PostToolUse** — records edited file paths
3. **Stop** — signals the TUI with the list of changed files

The TUI diffs the snapshot against the current file, so multiple edits to the same file in one response show as a single net diff.

## Commands

```bash
claude-code-preview            # launch TUI
claude-code-preview install    # install hooks and scripts
claude-code-preview status     # check dependencies and installation health
claude-code-preview uninstall  # remove hooks and scripts
```

Run `claude-code-preview status` if the pane isn't updating — it checks that `delta`, `jq`, and the hook scripts are all installed and wired up correctly.
