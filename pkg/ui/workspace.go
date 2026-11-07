package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderWorkspace renders the active tab viewport and the tab bar.
func (m *AppModel) renderWorkspace(boxWidth int) string {
	if !m.hasWorkspace || len(m.tabs) == 0 {
		m.hasWorkspace = false
		m.activeTab = 0
		return ""
	}

	if m.activeTab < 0 {
		m.activeTab = 0
	}
	if m.activeTab >= len(m.tabs) {
		m.activeTab = len(m.tabs) - 1
	}
	curIdx := m.activeTab

	maxVisibleTabs := 4
	startTab := 0
	if curIdx >= maxVisibleTabs {
		startTab = curIdx - maxVisibleTabs + 1
	}
	endTab := startTab + maxVisibleTabs
	if endTab > len(m.tabs) {
		endTab = len(m.tabs)
	}

	var tabHeaders []string
	if startTab > 0 {
		tabHeaders = append(tabHeaders, lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(fmt.Sprintf("+%d", startTab)))
	}

	for i := startTab; i < endTab; i++ {
		tab := m.tabs[i]
		title := fmt.Sprintf("[%d: %s]", i+1, tab.Command)
		if i == curIdx {
			tabHeaders = append(tabHeaders, lipgloss.NewStyle().
				Background(m.theme.Primary).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 1).
				Render(title))
		} else {
			tabHeaders = append(tabHeaders, lipgloss.NewStyle().
				Background(m.theme.BadgeBg).
				Foreground(m.theme.TextMuted).
				Padding(0, 1).
				Render(title))
		}
	}
	if len(m.tabs) > endTab {
		tabHeaders = append(tabHeaders, lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(fmt.Sprintf("+%d", len(m.tabs)-endTab)))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Center, tabHeaders...)

	curTab := m.tabs[curIdx]
	scrollInfo := " [Tamamı Görünür] "
	if curTab.LineCount() > curTab.Viewport.Height {
		pct := int(curTab.Viewport.ScrollPercent() * 100)
		scrollInfo = fmt.Sprintf(" [▲ %d%% ▼] ", pct)
	}

	durStr := ""
	if curTab.Duration > 0 {
		durStr = fmt.Sprintf(" [%s] ", formatDuration(curTab.Duration))
	}

	controlsHint := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(
		fmt.Sprintf("%s%s[Ctrl+Y | Ctrl+W]", durStr, scrollInfo),
	)

	topBar := lipgloss.JoinHorizontal(lipgloss.Center, tabBar, "  ", controlsHint)
	vpView := curTab.Viewport.View()
	boxContent := fmt.Sprintf("%s\n\n%s", topBar, vpView)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.BorderColor).
		Padding(1, 2).
		Width(boxWidth).
		Render(boxContent)
}
