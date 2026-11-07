package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

// TabRenderFunc defines a dynamic render callback for reactive tabs like sys telemetry.
type TabRenderFunc func(theme theme.ThemePalette, width, height int) string

// WorkspaceTab represents an open tab in the multi-workspace manager.
type WorkspaceTab struct {
	ID       int
	Command  string
	Title    string
	Output   string
	Duration time.Duration
	Viewport viewport.Model
	RenderFn TabRenderFunc
}

// NewWorkspaceTab constructs a new WorkspaceTab with initialized viewport.
func NewWorkspaceTab(id int, cmd, title, output string, renderFn TabRenderFunc, dur time.Duration, width, height int) WorkspaceTab {
	vpW := width
	vpH := height
	if vpW < 10 {
		vpW = 10
	}
	if vpH < 3 {
		vpH = 3
	}
	vp := viewport.New(vpW, vpH)
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	vp.SetContent(strings.TrimRight(normalized, "\n"))
	vp.GotoTop()
	return WorkspaceTab{
		ID:       id,
		Command:  cmd,
		Title:    title,
		Output:   output,
		Duration: dur,
		Viewport: vp,
		RenderFn: renderFn,
	}
}

// LineCount returns total number of lines in the tab output.
func (tab *WorkspaceTab) LineCount() int {
	if tab.Output == "" {
		return 0
	}
	return strings.Count(tab.Output, "\n") + 1
}
