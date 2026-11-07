package ui

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rkqze10/mt-tui/pkg/framebuffer"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func (m *AppModel) applyPixelOverlays(rawScreen string, width, height int) string {
	hasOverlays := m.hookManager.HasPixelHooks()

	m.pixelMu.RLock()
	hasOverrides := len(m.pixelOverrides) > 0
	var overridesCopy map[string]framebuffer.Cell
	if hasOverrides {
		overridesCopy = make(map[string]framebuffer.Cell, len(m.pixelOverrides))
		for k, v := range m.pixelOverrides {
			overridesCopy[k] = v
		}
	}
	m.pixelMu.RUnlock()

	if !hasOverlays && !hasOverrides {
		return rawScreen
	}

	fb := framebuffer.New(width, height)
	m.hookManager.ExecutePixelPipelines(fb)

	for k, cell := range overridesCopy {
		parts := strings.Split(k, ":")
		if len(parts) == 2 {
			x, _ := strconv.Atoi(parts[0])
			y, _ := strconv.Atoi(parts[1])
			fb.SetPixel(x, y, cell.Char, cell.Fg, cell.Bg, cell.Bold)
		}
	}

	return fb.CompositeOnANSI(rawScreen)
}

func (m *AppModel) View() (viewOut string) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			_ = os.WriteFile("mt_panic.log", []byte(fmt.Sprintf("PANIC IN VIEW: %v\n\n%s", r, stack)), 0644)
			safeW := m.width
			if safeW < 10 {
				safeW = 80
			}
			safeH := m.height
			if safeH < 10 {
				safeH = 24
			}
			viewOut = lipgloss.Place(safeW, safeH, lipgloss.Center, lipgloss.Center,
				fmt.Sprintf("⚠️ Arayüz Hatası (mt_panic.log): %v\n[ESC] ile temizleyin.", r),
			)
		}
	}()

	if m.width == 0 {
		return "Yükleniyor..."
	}

	// Super User x1000: Custom Fullscreen View Callback
	if m.customFullscreenView != nil {
		return m.customFullscreenView()
	}

	if m.overrideScreenContent != "" {
		return m.overrideScreenContent
	}

	if m.activeModal != nil {
		return m.renderModal()
	}

	boxWidth, _, isCompact := m.calculateLayout()

	var defaultBanner string
	if m.customBanner != "" {
		defaultBanner = m.customBanner
	} else if isCompact {
		defaultBanner = lipgloss.NewStyle().
			Foreground(m.theme.Primary).
			Bold(true).
			Render(fmt.Sprintf("● MT CORE v12.0 // %s [RING -1 HYPERVISOR]", strings.ToUpper(m.theme.Name)))
	} else {
		mPart := lipgloss.NewStyle().Foreground(m.theme.Secondary).Bold(true).Render(theme.DefaultAsciiM)
		tPart := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Render(theme.DefaultAsciiT)
		defaultBanner = lipgloss.JoinHorizontal(lipgloss.Top, mPart, tPart)
	}

	banner := m.hookManager.GetTextHook("banner", defaultBanner)
	if strings.Count(banner, "\n") > 16 {
		banner = defaultBanner
	}

	workspaceView := m.renderWorkspace(boxWidth)
	inputDeck := m.renderOpenCodeBox(boxWidth)

	var toastView string
	if m.activeToast != nil {
		toastColor := m.theme.Secondary
		if m.activeToast.IsError {
			toastColor = lipgloss.Color("#FF5555")
		}
		toastView = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(toastColor).
			Foreground(m.theme.TextMain).
			Padding(0, 1).
			Render("⚡ " + m.activeToast.Text)
	}

	var statusColor lipgloss.Color
	if m.isError {
		statusColor = lipgloss.Color("#EF4444")
	} else {
		statusColor = m.theme.TextMuted
	}

	defaultStatus := lipgloss.NewStyle().Foreground(statusColor).Render(m.statusMsg)
	status := m.hookManager.GetTextHook("status", defaultStatus)

	var footerDeck string
	if m.customFooter != "" {
		footerDeck = m.customFooter
	} else {
		keyPill := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
		keyVal := lipgloss.NewStyle().Foreground(m.theme.TextMain).Bold(true)
		footerDeck = fmt.Sprintf("%s %s   %s %s   %s %s   %s %s   %s %s",
			keyVal.Render("Tab"), keyPill.Render("Tamamla"),
			keyVal.Render("↑/↓"), keyPill.Render("Geçmiş"),
			keyVal.Render("Alt+1..9"), keyPill.Render("Sekme"),
			keyVal.Render("Ctrl+Y"), keyPill.Render("Kopyala"),
			keyVal.Render("Esc"), keyPill.Render("Temizle"),
		)
	}

	var components []string
	components = append(components, "", banner)
	if workspaceView != "" {
		components = append(components, "", workspaceView)
	}
	components = append(components, "", inputDeck)
	if toastView != "" {
		components = append(components, "", toastView)
	}
	components = append(components, "", status, footerDeck)

	console := lipgloss.JoinVertical(lipgloss.Center, components...)

	safeW := m.width
	if safeW < 10 {
		safeW = 80
	}
	safeH := m.height
	if safeH < 10 {
		safeH = 24
	}

	rawScreen := lipgloss.Place(safeW, safeH, lipgloss.Center, lipgloss.Center, console)
	return m.applyPixelOverlays(rawScreen, safeW, safeH)
}
