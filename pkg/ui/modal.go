package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderModal generates the ANSI popup overlay for confirmation, alert, or prompt modals.
func (m *AppModel) renderModal() string {
	if m.activeModal == nil {
		return ""
	}

	modalW := 54
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalW < 24 {
		modalW = 24
	}

	titleBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(m.theme.Primary).
		Bold(true).
		Padding(0, 1).
		Render(" " + m.activeModal.Title + " ")

	innerW := modalW - 6
	if innerW < 18 {
		innerW = 18
	}

	var contentB strings.Builder
	contentB.WriteString(titleBar + "\n\n")
	contentB.WriteString(lipgloss.NewStyle().
		Width(innerW).
		Foreground(m.theme.TextMain).
		Render(m.activeModal.Content) + "\n")

	if m.activeModal.IsConfirm {
		btnYes := lipgloss.NewStyle().Background(m.theme.Primary).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 1).Render("[Y] Evet (Arayüze Geç)")
		btnNo := lipgloss.NewStyle().Background(m.theme.BadgeBg).Foreground(m.theme.TextMuted).Padding(0, 1).Render("[N] Hayır (İptal)")
		buttonDeck := lipgloss.JoinHorizontal(lipgloss.Center, btnYes, "   ", btnNo)
		contentB.WriteString("\n" + buttonDeck + "\n")
	} else if m.activeModal.IsPrompt && m.activeModal.InputModel != nil {
		if innerW > 6 {
			m.activeModal.InputModel.Width = innerW - 4
		}
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.theme.Secondary).
			Padding(0, 1).
			Render(m.activeModal.InputModel.View())
		contentB.WriteString("\n" + inputBox + "\n")
		contentB.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("[ENTER] Onayla  |  [ESC] İptal"))
	} else {
		contentB.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("[ESC / ENTER] Kapat"))
	}

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Background(m.theme.BgDark).
		Padding(1, 2).
		Width(modalW).
		Render(contentB.String())

	safeW := m.width
	if safeW < 10 {
		safeW = 80
	}
	safeH := m.height
	if safeH < 10 {
		safeH = 24
	}
	rawModal := lipgloss.Place(safeW, safeH, lipgloss.Center, lipgloss.Center, modalBox)
	return m.applyPixelOverlays(rawModal, safeW, safeH)
}
