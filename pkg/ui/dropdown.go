package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rkqze10/mt-tui/pkg/plugin"
)

// renderOpenCodeBox builds the interactive command prompt with attached autocomplete dropdown.
func (m *AppModel) renderOpenCodeBox(boxWidth int) string {
	rawInput := m.input.Value()
	isCommandMode := strings.HasPrefix(rawInput, "/")

	var matches []plugin.Manifest
	if isCommandMode {
		matches = m.kernel.Search(rawInput)
	}

	fgPrefix := m.promptCfg.FgPrefix
	if fgPrefix == nil {
		fgPrefix = m.theme.Secondary
	}
	fgArrow := m.promptCfg.FgArrow
	if fgArrow == nil {
		fgArrow = m.theme.Primary
	}

	prompt := lipgloss.NewStyle().Foreground(fgPrefix).Bold(true).Render(m.promptCfg.Prefix)
	arrow := lipgloss.NewStyle().Foreground(fgArrow).Bold(true).Render(m.promptCfg.Arrow)
	inputLine := " " + prompt + arrow + m.input.View()

	if !isCommandMode {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.theme.BorderColor).
			Width(boxWidth).
			Padding(0, 1).
			Render(inputLine)
	}

	dropdownBorder := lipgloss.Border{
		Top: "─", Bottom: "─", Left: "│", Right: "│",
		TopLeft: "╭", TopRight: "╮", BottomLeft: "├", BottomRight: "┤",
	}

	dropdownStyle := lipgloss.NewStyle().
		Border(dropdownBorder).
		BorderForeground(m.theme.BorderColor).
		Width(boxWidth).
		Padding(0, 1)

	inputAttachedBorder := lipgloss.Border{
		Top: "", Bottom: "─", Left: "│", Right: "│",
		TopLeft: "", TopRight: "", BottomLeft: "╰", BottomRight: "╯",
	}

	inputAttachedStyle := lipgloss.NewStyle().
		Border(inputAttachedBorder).
		BorderTop(false).
		BorderForeground(m.theme.BorderColor).
		Width(boxWidth).
		Padding(0, 1)

	var listLines []string
	if len(matches) == 0 {
		listLines = append(listLines, lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("  Eşleşen komut bulunamadı"))
	} else {
		windowSize := 5
		if len(matches) < windowSize {
			windowSize = len(matches)
		}

		// Ensure selectedIndex is clamped
		selIdx := m.selectedIndex
		if selIdx < 0 {
			selIdx = 0
		}
		if selIdx >= len(matches) {
			selIdx = len(matches) - 1
		}

		startIdx := 0
		if selIdx >= windowSize {
			startIdx = selIdx - windowSize + 1
		}
		endIdx := startIdx + windowSize
		if endIdx > len(matches) {
			endIdx = len(matches)
			startIdx = endIdx - windowSize
			if startIdx < 0 {
				startIdx = 0
			}
		}

		innerW := boxWidth - 6
		if innerW < 22 {
			innerW = 22
		}
		cursorW := 2
		cmdW := 14
		badgeW := 6
		if innerW > 50 {
			cmdW = 18
			badgeW = 8
		}
		availDescW := innerW - cursorW - cmdW - badgeW - 3
		if availDescW < 6 {
			availDescW = 6
		}

		for i := startIdx; i < endIdx; i++ {
			item := matches[i]
			isSelected := (i == selIdx)

			cursor := "  "
			if isSelected {
				cursor = "▸ "
			}
			cursorSpan := lipgloss.NewStyle().Width(cursorW).Foreground(m.theme.Primary).Bold(true).Render(cursor)

			cmdSt := lipgloss.NewStyle().Width(cmdW).Foreground(m.theme.Primary)
			if isSelected {
				cmdSt = lipgloss.NewStyle().Width(cmdW).Foreground(m.theme.Secondary).Bold(true)
			}
			cmdSpan := cmdSt.Render("/" + item.Command)

			badgeSpan := lipgloss.NewStyle().Width(badgeW).Foreground(m.theme.TextMuted).Render("[" + item.Category + "]")

			descColor := m.theme.TextMuted
			if isSelected {
				descColor = m.theme.TextMain
			}
			descSpan := lipgloss.NewStyle().
				MaxWidth(availDescW).
				Foreground(descColor).
				Render(item.Description)

			lineContent := lipgloss.JoinHorizontal(lipgloss.Left, cursorSpan, cmdSpan, badgeSpan, descSpan)
			rowStyle := lipgloss.NewStyle().MaxHeight(1)
			if isSelected {
				rowStyle = rowStyle.Background(m.theme.BadgeBg)
			}
			listLines = append(listLines, rowStyle.Render(lineContent))
		}
	}

	topDropdown := dropdownStyle.Render(strings.Join(listLines, "\n"))
	botInput := inputAttachedStyle.Render(inputLine)

	return lipgloss.JoinVertical(lipgloss.Left, topDropdown, botInput)
}
