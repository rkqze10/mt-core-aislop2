package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ThemePalette defines the color scheme across all UI components.
type ThemePalette struct {
	Name        string         `json:"name"`
	Primary     lipgloss.Color `json:"primary"`
	Secondary   lipgloss.Color `json:"secondary"`
	BgDark      lipgloss.Color `json:"bg_dark"`
	BorderColor lipgloss.Color `json:"border_color"`
	TextMain    lipgloss.Color `json:"text_main"`
	TextMuted   lipgloss.Color `json:"text_muted"`
	BadgeBg     lipgloss.Color `json:"badge_bg"`
}

var (
	ThemeAmber = ThemePalette{
		Name:        "amber",
		Primary:     lipgloss.Color("#FF9E3B"),
		Secondary:   lipgloss.Color("#00FF88"),
		BgDark:      lipgloss.Color("#18181B"),
		BorderColor: lipgloss.Color("#FF9E3B"),
		TextMain:    lipgloss.Color("#F4F4F5"),
		TextMuted:   lipgloss.Color("#71717A"),
		BadgeBg:     lipgloss.Color("#27272A"),
	}
	ThemeEmerald = ThemePalette{
		Name:        "emerald",
		Primary:     lipgloss.Color("#00FF88"),
		Secondary:   lipgloss.Color("#05FFA1"),
		BgDark:      lipgloss.Color("#0A160F"),
		BorderColor: lipgloss.Color("#00FF88"),
		TextMain:    lipgloss.Color("#E8FFE8"),
		TextMuted:   lipgloss.Color("#3A6B4E"),
		BadgeBg:     lipgloss.Color("#12261A"),
	}
	ThemeCyber = ThemePalette{
		Name:        "cyber",
		Primary:     lipgloss.Color("#00F0FF"),
		Secondary:   lipgloss.Color("#FF007F"),
		BgDark:      lipgloss.Color("#0E0B16"),
		BorderColor: lipgloss.Color("#00F0FF"),
		TextMain:    lipgloss.Color("#FFFFFF"),
		TextMuted:   lipgloss.Color("#7B7291"),
		BadgeBg:     lipgloss.Color("#1F1A2E"),
	}
	ThemeMono = ThemePalette{
		Name:        "mono",
		Primary:     lipgloss.Color("#FFFFFF"),
		Secondary:   lipgloss.Color("#A1A1AA"),
		BgDark:      lipgloss.Color("#09090B"),
		BorderColor: lipgloss.Color("#71717A"),
		TextMain:    lipgloss.Color("#FAFAFA"),
		TextMuted:   lipgloss.Color("#52525B"),
		BadgeBg:     lipgloss.Color("#27272A"),
	}
)

// GetByName returns a ThemePalette corresponding to the given string name.
func GetByName(name string) (ThemePalette, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "amber":
		return ThemeAmber, true
	case "emerald":
		return ThemeEmerald, true
	case "cyber":
		return ThemeCyber, true
	case "mono":
		return ThemeMono, true
	default:
		return ThemeAmber, false
	}
}

// AvailableThemes returns the list of default theme names.
func AvailableThemes() []string {
	return []string{"amber", "emerald", "cyber", "mono"}
}

const DefaultAsciiM = ` ███╗   ███╗
 ████╗ ████║
 ██╔████╔██║
 ██║╚██╔╝██║
 ██║ ╚═╝ ██║
 ╚═╝     ╚═╝`

const DefaultAsciiT = ` ████████╗
 ╚══██╔══╝
    ██║   
    ██║   
    ██║   
    ╚═╝   `
