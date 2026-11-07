package framebuffer

import (
	"github.com/charmbracelet/lipgloss"
)

// Cell represents a single virtual pixel on the terminal canvas.
type Cell struct {
	Char rune
	Fg   lipgloss.TerminalColor
	Bg   lipgloss.TerminalColor
	Bold bool
}
