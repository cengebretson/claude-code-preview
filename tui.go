package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	signalFile = "/tmp/claude-preview-signal"
	sessionFile = "/tmp/claude-preview-signal.session"
	paneTitle  = "claude-preview"
)

type Theme struct {
	Green    lipgloss.Color
	Red      lipgloss.Color
	Mauve    lipgloss.Color
	Overlay1 lipgloss.Color
	Surface0 lipgloss.Color
	Yellow   lipgloss.Color
	Peach    lipgloss.Color
}

type styles struct {
	header   lipgloss.Style
	selected lipgloss.Style
	dim      lipgloss.Style
	sep      lipgloss.Style
	wait     lipgloss.Style
	add      lipgloss.Style
	del      lipgloss.Style
	status   lipgloss.Style
	undo     lipgloss.Style
	helpKey  lipgloss.Style
}

func newStyles(t Theme) styles {
	return styles{
		header:   lipgloss.NewStyle().Foreground(t.Mauve),
		selected: lipgloss.NewStyle().Foreground(t.Green).Background(t.Surface0),
		dim:      lipgloss.NewStyle().Foreground(t.Overlay1),
		sep:      lipgloss.NewStyle().Foreground(t.Overlay1),
		wait:     lipgloss.NewStyle().Foreground(t.Overlay1).Padding(1, 2),
		add:      lipgloss.NewStyle().Foreground(t.Green),
		del:      lipgloss.NewStyle().Foreground(t.Red),
		status:   lipgloss.NewStyle().Foreground(t.Yellow),
		undo:     lipgloss.NewStyle().Foreground(t.Peach),
		helpKey:  lipgloss.NewStyle().Foreground(t.Mauve).Width(20),
	}
}

// Catppuccin Mocha
var CatppuccinMocha = Theme{
	Green:    "#a6e3a1",
	Red:      "#f38ba8",
	Mauve:    "#cba6f7",
	Overlay1: "#7f849c",
	Surface0: "#313244",
	Yellow:   "#f9e2af",
	Peach:    "#fab387",
}

var defaultStyles styles
var pollRate = 500 * time.Millisecond

var fileIcons = map[string]string{
	".go":    "",
	".js":    "",
	".ts":    "",
	".tsx":   "",
	".jsx":   "",
	".py":    "",
	".rs":    "",
	".sh":    "",
	".bash":  "",
	".fish":  "",
	".zsh":   "",
	".md":    "",
	".json":  "",
	".yaml":  "",
	".yml":   "",
	".toml":  "",
	".css":   "",
	".scss":  "",
	".html":  "",
	".lua":   "",
	".vim":   "",
	".rb":    "",
	".java":  "",
	".swift": "",
	".c":     "",
	".cpp":   "",
	".cs":    "",
	".sql":   "",
}

func fileIcon(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if icon, ok := fileIcons[ext]; ok {
		return icon + " "
	}
	return " "
}

type tuiModel struct {
	files      []string
	fileStats  map[string][2]int // [added, removed] per file
	sessionID  string
	selected   int
	listOffset int
	waiting    bool
	sideBySide bool
	showHelp   bool
	statusMsg  string
	width      int
	height     int
	viewport   viewport.Model
	ready      bool
}

type tickMsg time.Time
type filesLoadedMsg struct {
	files     []string
	sessionID string
}
type diffOutputMsg string
type statsLoadedMsg struct {
	file    string
	added   int
	removed int
}
type nvimDoneMsg struct {
	file      string
	sessionID string
}

func tickCmd() tea.Cmd {
	return tea.Tick(pollRate, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func checkSignalCmd() tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(signalFile)
		if err != nil {
			return nil
		}
		sessionData, _ := os.ReadFile(sessionFile)
		os.Remove(signalFile)
		os.Remove(sessionFile)

		var files []string
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				files = append(files, line)
			}
		}
		return filesLoadedMsg{
			files:     files,
			sessionID: strings.TrimSpace(string(sessionData)),
		}
	}
}

func loadDiffCmd(file, sessionID string, sideBySide bool) tea.Cmd {
	return func() tea.Msg {
		snapName := strings.ReplaceAll(file, "/", "_")
		snapshot := fmt.Sprintf("/tmp/claude-snapshots-%s/%s", sessionID, snapName)

		flags := "--file-style omit --hunk-header-style omit"
		if sideBySide {
			flags += " --side-by-side"
		}

		var cmd *exec.Cmd
		if _, err := os.Stat(snapshot); err == nil {
			cmd = exec.Command("sh", "-c", fmt.Sprintf(
				"git diff --no-index %q %q 2>/dev/null | delta %s",
				snapshot, file, flags,
			))
		} else {
			cmd = exec.Command("sh", "-c", fmt.Sprintf(
				"git diff HEAD -- %q | delta %s",
				file, flags,
			))
		}
		out, _ := cmd.Output()
		if len(strings.TrimSpace(string(out))) == 0 {
			return diffOutputMsg("  no diff available")
		}
		return diffOutputMsg(string(out))
	}
}

func loadStatsCmd(file, sessionID string) tea.Cmd {
	return func() tea.Msg {
		snapName := strings.ReplaceAll(file, "/", "_")
		snapshot := fmt.Sprintf("/tmp/claude-snapshots-%s/%s", sessionID, snapName)

		var script string
		if _, err := os.Stat(snapshot); err == nil {
			script = fmt.Sprintf("git diff --numstat --no-index %q %q 2>/dev/null | awk '{print $1, $2}'", snapshot, file)
		} else {
			script = fmt.Sprintf("git diff --numstat HEAD -- %q | awk '{print $1, $2}'", file)
		}
		out, _ := exec.Command("sh", "-c", script).Output()
		parts := strings.Fields(strings.TrimSpace(string(out)))
		added, removed := 0, 0
		if len(parts) == 2 {
			added, _ = strconv.Atoi(parts[0])
			removed, _ = strconv.Atoi(parts[1])
		}
		return statsLoadedMsg{file: file, added: added, removed: removed}
	}
}

func snapshotPath(file, sessionID string) string {
	snapName := strings.ReplaceAll(file, "/", "_")
	return fmt.Sprintf("/tmp/claude-snapshots-%s/%s", sessionID, snapName)
}

func preferredEditor() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	return "nvim"
}

func undoFile(file, sessionID string) error {
	snap := snapshotPath(file, sessionID)
	data, err := os.ReadFile(snap)
	if err != nil {
		return fmt.Errorf("no snapshot for %s", filepath.Base(file))
	}
	return os.WriteFile(file, data, 0644)
}

func (m tuiModel) listAreaHeight() int {
	n := len(m.files) + 3 // header + files + hint
	max := m.height / 3
	if n > max {
		n = max
	}
	if n < 3 {
		n = 3
	}
	return n
}

func (m tuiModel) visibleFileCount() int {
	n := m.listAreaHeight() - 2 // minus header and hint lines
	if n < 1 {
		n = 1
	}
	return n
}

func (m *tuiModel) clampListOffset() {
	vis := m.visibleFileCount()
	if m.selected < m.listOffset {
		m.listOffset = m.selected
	}
	if m.selected >= m.listOffset+vis {
		m.listOffset = m.selected - vis + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

func (m tuiModel) Init() tea.Cmd {
	exec.Command("tmux", "select-pane", "-t", os.Getenv("TMUX_PANE"), "-T", paneTitle).Run()
	return tea.Batch(tickCmd(), checkSignalCmd())
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vpHeight := m.height - m.listAreaHeight() - 1
		if vpHeight < 1 {
			vpHeight = 1
		}
		if !m.ready {
			m.viewport = viewport.New(m.width, vpHeight)
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = vpHeight
		}

	case tickMsg:
		cmds = append(cmds, tickCmd(), checkSignalCmd())

	case filesLoadedMsg:
		if len(msg.files) > 0 {
			m.files = msg.files
			m.fileStats = make(map[string][2]int)
			m.sessionID = msg.sessionID
			m.selected = 0
			m.listOffset = 0
			m.waiting = false
			m.statusMsg = ""
			vpHeight := m.height - m.listAreaHeight() - 1
			if vpHeight < 1 {
				vpHeight = 1
			}
			m.viewport.Height = vpHeight
			m.viewport.GotoTop()
			cmds = append(cmds, loadDiffCmd(m.files[0], m.sessionID, m.sideBySide))
			for _, f := range m.files {
				cmds = append(cmds, loadStatsCmd(f, m.sessionID))
			}
		}

	case statsLoadedMsg:
		if m.fileStats == nil {
			m.fileStats = make(map[string][2]int)
		}
		m.fileStats[msg.file] = [2]int{msg.added, msg.removed}

	case diffOutputMsg:
		m.viewport.SetContent(string(msg))
		m.viewport.GotoTop()

	case nvimDoneMsg:
		cmds = append(cmds, loadDiffCmd(msg.file, msg.sessionID, m.sideBySide))
		cmds = append(cmds, loadStatsCmd(msg.file, msg.sessionID))

	case tea.KeyMsg:
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "Q":
			if !m.waiting {
				m.waiting = true
				m.files = nil
				m.fileStats = nil
				m.statusMsg = ""
				m.viewport.SetContent("")
				return m, nil
			}
			return m, tea.Quit

		case "ctrl+c":
			return m, tea.Quit

		case "r":
			if !m.waiting && len(m.files) > 0 {
				m.statusMsg = ""
				cmds = append(cmds, loadDiffCmd(m.files[m.selected], m.sessionID, m.sideBySide))
				cmds = append(cmds, loadStatsCmd(m.files[m.selected], m.sessionID))
			}

		case "?":
			m.showHelp = true

		case "up", "k":
			if !m.waiting && m.selected > 0 {
				m.selected--
				m.clampListOffset()
				m.viewport.GotoTop()
				m.statusMsg = ""
				cmds = append(cmds, loadDiffCmd(m.files[m.selected], m.sessionID, m.sideBySide))
			}

		case "down", "j":
			if !m.waiting && m.selected < len(m.files)-1 {
				m.selected++
				m.clampListOffset()
				m.viewport.GotoTop()
				m.statusMsg = ""
				cmds = append(cmds, loadDiffCmd(m.files[m.selected], m.sessionID, m.sideBySide))
			}

		case "enter":
			if !m.waiting && len(m.files) > 0 {
				file := m.files[m.selected]
				sessionID := m.sessionID
				c := exec.Command(preferredEditor(), file)
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return nvimDoneMsg{file: file, sessionID: sessionID}
				})
			}

		case "u":
			if !m.waiting && len(m.files) > 0 {
				file := m.files[m.selected]
				if err := undoFile(file, m.sessionID); err != nil {
					m.statusMsg = "✗ " + err.Error()
				} else {
					m.statusMsg = "✓ restored " + filepath.Base(file)
					cmds = append(cmds, loadDiffCmd(file, m.sessionID, m.sideBySide))
					cmds = append(cmds, loadStatsCmd(file, m.sessionID))
				}
			}

		case "U":
			if !m.waiting && len(m.files) > 0 {
				count := 0
				for _, f := range m.files {
					if undoFile(f, m.sessionID) == nil {
						count++
						cmds = append(cmds, loadStatsCmd(f, m.sessionID))
					}
				}
				m.statusMsg = fmt.Sprintf("✓ restored %d file(s)", count)
				cmds = append(cmds, loadDiffCmd(m.files[m.selected], m.sessionID, m.sideBySide))
			}

		case "s":
			if !m.waiting {
				m.sideBySide = !m.sideBySide
				if len(m.files) > 0 {
					cmds = append(cmds, loadDiffCmd(m.files[m.selected], m.sessionID, m.sideBySide))
				}
			}

		case "y":
			if !m.waiting && len(m.files) > 0 {
				file := m.files[m.selected]
				c := exec.Command("pbcopy")
				c.Stdin = strings.NewReader(file)
				if c.Run() == nil {
					m.statusMsg = "✓ copied path"
				}
			}

		default:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.viewport.LineUp(3)
		case tea.MouseButtonWheelDown:
			m.viewport.LineDown(3)
		case tea.MouseButtonLeft:
			// Click on file list (row 1 = first file, row 0 = header)
			if msg.Y >= 1 && msg.Y <= m.visibleFileCount() && !m.waiting {
				idx := msg.Y - 1 + m.listOffset
				if idx >= 0 && idx < len(m.files) && idx != m.selected {
					m.selected = idx
					m.clampListOffset()
					m.viewport.GotoTop()
					m.statusMsg = ""
					cmds = append(cmds, loadDiffCmd(m.files[m.selected], m.sessionID, m.sideBySide))
				}
			}
		}
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	if vpCmd != nil {
		cmds = append(cmds, vpCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m tuiModel) View() string {
	if !m.ready {
		return ""
	}

	if m.showHelp {
		return m.helpView()
	}

	if m.waiting {
		return defaultStyles.wait.Render("󱙺  waiting for claude changes...")
	}

	// File list
	vis := m.visibleFileCount()
	end := m.listOffset + vis
	if end > len(m.files) {
		end = len(m.files)
	}

	lines := []string{defaultStyles.header.Render("󱙺 changed files")}
	for i := m.listOffset; i < end; i++ {
		f := m.files[i]
		display := strings.Replace(f, os.Getenv("HOME"), "~", 1)
		icon := fileIcon(f)

		statsStr := ""
		if stats, ok := m.fileStats[f]; ok {
			a := defaultStyles.add.Render(fmt.Sprintf("+%d", stats[0]))
			d := defaultStyles.del.Render(fmt.Sprintf("-%d", stats[1]))
			statsStr = " " + a + " " + d
		}

		if i == m.selected {
			lines = append(lines, defaultStyles.selected.Render(fmt.Sprintf("  ▶ %s%s", icon, display))+statsStr)
		} else {
			lines = append(lines, defaultStyles.dim.Render(fmt.Sprintf("    %s%s", icon, display))+statsStr)
		}
	}

	hint := "  ↑↓ · enter: open · q: clear · ?: help"
	if m.statusMsg != "" {
		hint = "  " + defaultStyles.status.Render(m.statusMsg)
	}
	lines = append(lines, defaultStyles.dim.Render(hint))

	listView := lipgloss.NewStyle().
		Height(m.listAreaHeight()).
		Width(m.width).
		Render(strings.Join(lines, "\n"))

	sep := defaultStyles.sep.Render(strings.Repeat("─", m.width))

	return lipgloss.JoinVertical(lipgloss.Left, listView, sep, m.viewport.View())
}

func (m tuiModel) helpView() string {
	bindings := [][]string{
		{"↑ / k", "previous file"},
		{"↓ / j", "next file"},
		{"enter", "open in $VISUAL / $EDITOR (default: nvim)"},
		{"u", "undo current file (restore snapshot)"},
		{"U", "undo all files"},
		{"s", "toggle side-by-side diff"},
		{"y", "copy file path to clipboard"},
		{"ctrl+u / ctrl+d", "scroll diff half page"},
		{"ctrl+f / ctrl+b", "scroll diff full page"},
		{"q", "clear / quit"},
		{"?", "toggle this help"},
	}

	var sb strings.Builder
	sb.WriteString(defaultStyles.header.Render("󱙺 keybindings") + "\n\n")
	for _, b := range bindings {
		key := defaultStyles.helpKey.Render(b[0])
		sb.WriteString(fmt.Sprintf("  %s %s\n", key, defaultStyles.dim.Render(b[1])))
	}
	sb.WriteString("\n" + defaultStyles.dim.Render("  press any key to close"))
	return sb.String()
}

func runTUI() error {
	missing := []string{}
	for _, bin := range []string{"delta", "jq", "tmux"} {
		if _, err := exec.LookPath(bin); err != nil {
			missing = append(missing, bin)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required dependencies: %s\nInstall with: brew install %s",
			strings.Join(missing, ", "), strings.Join(missing, " "))
	}

	cfg := loadConfig()
	defaultStyles = newStyles(cfg.theme)
	pollRate = cfg.pollRate

	p := tea.NewProgram(
		tuiModel{waiting: true},
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
