package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTypeSlashAndCommands(t *testing.T) {
	model := NewAppModel()
	model.width = 60
	model.height = 20

	// 1. Send WindowSizeMsg
	m, _ := model.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	appM := m.(*AppModel)

	// 2. Type "/"
	m, _ = appM.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	appM = m.(*AppModel)

	// Render view - this is where typing "/" happens!
	v := appM.View()
	if v == "" {
		t.Fatal("View returned empty string")
	}

	// 3. Type "h" -> "/h"
	m, _ = appM.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	appM = m.(*AppModel)
	v = appM.View()
	if v == "" {
		t.Fatal("View returned empty string")
	}

	// 4. Type "elp" and enter
	for _, ch := range "elp" {
		m, _ = appM.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		appM = m.(*AppModel)
		_ = appM.View()
	}

	// Execute command on enter
	m, _ = appM.Update(tea.KeyMsg{Type: tea.KeyEnter})
	appM = m.(*AppModel)
	v = appM.View()
	if v == "" {
		t.Fatal("View returned empty string after enter")
	}

	// 5. Very small mobile terminal simulation (Termux with keyboard up: width 40, height 12)
	m, _ = appM.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	appM = m.(*AppModel)
	m, _ = appM.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	appM = m.(*AppModel)
	_ = appM.View()
}
