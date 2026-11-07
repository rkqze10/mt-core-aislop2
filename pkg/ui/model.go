package ui

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rkqze10/mt-tui/pkg/builtins"
	"github.com/rkqze10/mt-tui/pkg/config"
	"github.com/rkqze10/mt-tui/pkg/framebuffer"
	"github.com/rkqze10/mt-tui/pkg/kernel"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/state"
	"github.com/rkqze10/mt-tui/pkg/superuser"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

type HotReloadMsg struct {
	Count int
}

type NotifyMsg struct {
	Text    string
	IsError bool
}

type ToastMsg struct {
	ID      int64
	Text    string
	IsError bool
}

type ToastTriggerMsg struct {
	ID      int64
	Text    string
	IsError bool
}

type ToastClearMsg struct {
	ID int64
}

type ExternalTUIFinishedMsg struct {
	Err error
}

type PromptConfig struct {
	Prefix   string
	Arrow    string
	FgPrefix lipgloss.TerminalColor
	FgArrow  lipgloss.TerminalColor
}

// AppModel is the primary Bubble Tea model managing the MT Core hypervisor.
type AppModel struct {
	input         textinput.Model
	kernel        *kernel.KernelEngine
	config        config.PersistentConfig
	program       *tea.Program
	theme         theme.ThemePalette
	hookManager   *superuser.SuperuserHookManager
	store         *state.SharedStateStore
	tabs          []WorkspaceTab
	activeTab     int
	tabCounter    int
	hasWorkspace  bool
	width         int
	height        int
	statusMsg     string
	isError       bool
	selectedIndex int
	activeModal   *superuser.ModalDialog
	activeToast   *ToastMsg
	toastCounter  atomic.Int64
	historyIdx    int
	historyDraft  string
	lastInputVal  string
	watchCancel   context.CancelFunc

	promptCfg             PromptConfig
	keybindings           map[string]string
	keybindMu             sync.RWMutex
	overrideScreenContent string
	pixelMu               sync.RWMutex
	pixelOverrides        map[string]framebuffer.Cell
	inputHijacked         bool

	// Super User x1000 Custom Overrides
	customBanner         string
	customFooter         string
	customFullscreenView func() string

	pendingTUI *superuser.PendingTUIRequest
}

var _ superuser.SuperuserHost = (*AppModel)(nil)

func (m *AppModel) SendTUI(msg tea.Msg) {
	if m.program != nil {
		m.program.Send(msg)
	}
}

func (m *AppModel) ExecOS(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	bytes, err := cmd.CombinedOutput()
	return string(bytes), err
}

func (m *AppModel) Notify(msg string, isError bool) {
	m.SendTUI(NotifyMsg{Text: msg, IsError: isError})
}

func (m *AppModel) Toast(text string, isError bool) {
	id := m.toastCounter.Add(1)
	m.SendTUI(ToastTriggerMsg{
		ID:      id,
		Text:    text,
		IsError: isError,
	})
}

func (m *AppModel) OpenModal(title, content string) {
	m.activeModal = &superuser.ModalDialog{
		Title:   title,
		Content: content,
	}
}

func (m *AppModel) OpenPrompt(title, promptText string, onSubmit func(val string)) {
	ti := textinput.New()
	ti.Placeholder = "metin girin..."
	ti.Focus()
	ti.CharLimit = 80
	ti.Width = 36
	m.activeModal = &superuser.ModalDialog{
		Title:        title,
		Content:      promptText,
		IsPrompt:     true,
		PromptPrompt: promptText,
		InputModel:   &ti,
		Callback:     onSubmit,
	}
}

func (m *AppModel) OpenConfirm(title, message string, onConfirm func(confirmed bool)) {
	m.activeModal = &superuser.ModalDialog{
		Title:     title,
		Content:   message,
		IsConfirm: true,
		OnConfirm: onConfirm,
	}
}

func (m *AppModel) CloseModal() {
	m.activeModal = nil
	m.input.Focus()
}

func (m *AppModel) SetTheme(name string) bool {
	t, ok := theme.GetByName(name)
	if !ok {
		return false
	}
	m.theme = t
	m.config.Theme = t.Name
	_ = config.SaveConfig(m.config)

	for i := range m.tabs {
		if m.tabs[i].RenderFn != nil {
			m.tabs[i].Output = m.tabs[i].RenderFn(m.theme, m.tabs[i].Viewport.Width, m.tabs[i].Viewport.Height)
			m.tabs[i].Viewport.SetContent(m.tabs[i].Output)
			m.tabs[i].Viewport.GotoTop()
		}
	}
	return true
}

func (m *AppModel) SetCustomTheme(t theme.ThemePalette) {
	m.theme = t
	for i := range m.tabs {
		if m.tabs[i].RenderFn != nil {
			m.tabs[i].Output = m.tabs[i].RenderFn(m.theme, m.tabs[i].Viewport.Width, m.tabs[i].Viewport.Height)
			m.tabs[i].Viewport.SetContent(m.tabs[i].Output)
			m.tabs[i].Viewport.GotoTop()
		}
	}
}

func (m *AppModel) CurrentTheme() theme.ThemePalette {
	return m.theme
}

func (m *AppModel) KV() *state.SharedStateStore {
	return m.store
}

func (m *AppModel) ClipboardSet(text string) {
	if len(text) < 65536 {
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		fmt.Fprintf(os.Stderr, "\x1b]52;c;%s\x07", encoded)
	}

	go func() {
		defer func() {
			_ = recover()
		}()

		// Termux Android API clipboard check (highest priority on Termux)
		if p, err := exec.LookPath("termux-clipboard-set"); err == nil {
			cmd := exec.Command(p)
			cmd.Stdin = strings.NewReader(text)
			if err := cmd.Run(); err == nil {
				return
			}
		}

		switch runtime.GOOS {
		case "darwin":
			if p, err := exec.LookPath("pbcopy"); err == nil {
				cmd := exec.Command(p)
				cmd.Stdin = strings.NewReader(text)
				_ = cmd.Run()
			}
		case "linux", "android":
			// Wayland first
			if os.Getenv("WAYLAND_DISPLAY") != "" {
				if p, err := exec.LookPath("wl-copy"); err == nil {
					cmd := exec.Command(p)
					cmd.Stdin = strings.NewReader(text)
					if err := cmd.Run(); err == nil {
						return
					}
				}
			}

			// X11 / Xterm / Xwayland: only attempt xclip/xsel if DISPLAY or WAYLAND_DISPLAY is set
			hasX11 := os.Getenv("DISPLAY") != ""
			if hasX11 {
				if p, err := exec.LookPath("xclip"); err == nil {
					cmd := exec.Command(p, "-selection", "clipboard")
					cmd.Stdin = strings.NewReader(text)
					if err := cmd.Run(); err == nil {
						return
					}
				}
				if p, err := exec.LookPath("xsel"); err == nil {
					cmd := exec.Command(p, "-b", "-i")
					cmd.Stdin = strings.NewReader(text)
					if err := cmd.Run(); err == nil {
						return
					}
				}
			}

			// Generic fallbacks without crashing
			if p, err := exec.LookPath("wl-copy"); err == nil {
				cmd := exec.Command(p)
				cmd.Stdin = strings.NewReader(text)
				_ = cmd.Run()
			}
		case "windows":
			if p, err := exec.LookPath("powershell"); err == nil {
				cmd := exec.Command(p, "-NoProfile", "-Command", "$input | Set-Clipboard")
				cmd.Stdin = strings.NewReader(text)
				_ = cmd.Run()
			}
		}
	}()
}

func (m *AppModel) MutatePixel(x, y int, ch rune, fg, bg lipgloss.TerminalColor, bold bool) {
	key := fmt.Sprintf("%d:%d", x, y)
	m.pixelMu.Lock()
	m.pixelOverrides[key] = framebuffer.Cell{
		Char: ch,
		Fg:   fg,
		Bg:   bg,
		Bold: bold,
	}
	m.pixelMu.Unlock()
}

func (m *AppModel) MutatePixelBatch(pixels []superuser.CellMutation) {
	m.pixelMu.Lock()
	for _, p := range pixels {
		key := fmt.Sprintf("%d:%d", p.X, p.Y)
		m.pixelOverrides[key] = framebuffer.Cell{
			Char: p.Char,
			Fg:   p.Fg,
			Bg:   p.Bg,
			Bold: p.Bold,
		}
	}
	m.pixelMu.Unlock()
}

func (m *AppModel) DrawBox(x, y, w, h int, title string, borderFg lipgloss.TerminalColor, bold bool) {
	fb := framebuffer.New(w, h)
	fb.DrawBox(0, 0, w, h, title, borderFg, bold)
	m.pixelMu.Lock()
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			c, ok := fb.GetPixel(col, row)
			if ok && c.Char != 0 && c.Char != ' ' {
				key := fmt.Sprintf("%d:%d", x+col, y+row)
				m.pixelOverrides[key] = c
			}
		}
	}
	m.pixelMu.Unlock()
}

func (m *AppModel) ClearPixels() {
	m.pixelMu.Lock()
	m.pixelOverrides = make(map[string]framebuffer.Cell)
	m.pixelMu.Unlock()
}

func (m *AppModel) RegisterOverlay(id string, hook superuser.DynamicPixelHook) {
	m.hookManager.RegisterPixelHook(id, hook)
}

func (m *AppModel) RemoveOverlay(id string) {
	m.hookManager.RemovePixelHook(id)
}

func (m *AppModel) SetTextHook(target, content string) {
	m.hookManager.SetTextHook(target, content)
}

func (m *AppModel) ClearTextHook(target string) {
	m.hookManager.ClearTextHook(target)
}

func (m *AppModel) SetPrompt(prefix, arrow string, fgPrefix, fgArrow lipgloss.TerminalColor) {
	m.promptCfg = PromptConfig{
		Prefix:   prefix,
		Arrow:    arrow,
		FgPrefix: fgPrefix,
		FgArrow:  fgArrow,
	}
}

func (m *AppModel) ResetPrompt() {
	m.promptCfg = PromptConfig{
		Prefix:   "ROOT ",
		Arrow:    "❯ ",
		FgPrefix: m.theme.Secondary,
		FgArrow:  m.theme.Primary,
	}
}

func (m *AppModel) BindKey(key, command string) {
	m.keybindMu.Lock()
	defer m.keybindMu.Unlock()
	m.keybindings[strings.ToLower(strings.TrimSpace(key))] = command
}

func (m *AppModel) UnbindKey(key string) {
	m.keybindMu.Lock()
	defer m.keybindMu.Unlock()
	delete(m.keybindings, strings.ToLower(strings.TrimSpace(key)))
}

func (m *AppModel) PatchCommand(targetCmd, rewriteCmd string) {
	m.kernel.PatchCommand(targetCmd, rewriteCmd)
}

func (m *AppModel) UnpatchCommand(targetCmd string) {
	m.kernel.UnpatchCommand(targetCmd)
}

func (m *AppModel) ForceCloseTab(tabIdx int) bool {
	if tabIdx >= 0 && tabIdx < len(m.tabs) {
		m.tabs = append(m.tabs[:tabIdx], m.tabs[tabIdx+1:]...)
		if len(m.tabs) == 0 {
			m.hasWorkspace = false
			m.activeTab = 0
		} else if m.activeTab >= len(m.tabs) {
			m.activeTab = len(m.tabs) - 1
		}
		return true
	}
	return false
}

func (m *AppModel) InjectTab(title, cmd, output string) {
	boxWidth, vpHeight, _ := m.calculateLayout()
	safeBoxW := boxWidth - 6
	if safeBoxW < 10 {
		safeBoxW = 10
	}
	safeVpH := vpHeight
	if safeVpH < 3 {
		safeVpH = 3
	}

	m.tabCounter++
	newTab := NewWorkspaceTab(m.tabCounter, cmd, title, output, nil, 0, safeBoxW, safeVpH)
	m.tabs = append(m.tabs, newTab)
	m.activeTab = len(m.tabs) - 1
	m.hasWorkspace = true
}

func (m *AppModel) UpdateTabOutput(tabIdx int, output string) bool {
	if tabIdx >= 0 && tabIdx < len(m.tabs) {
		m.tabs[tabIdx].Output = output
		m.tabs[tabIdx].Viewport.SetContent(strings.TrimRight(output, "\r\n"))
		m.tabs[tabIdx].Viewport.GotoTop()
		return true
	}
	return false
}

func (m *AppModel) SetRawScreenContent(rendered string) {
	m.overrideScreenContent = rendered
}

func (m *AppModel) HijackInputFocus(active bool) {
	m.inputHijacked = active
}

func (m *AppModel) GetTerminalDimensions() (width, height int) {
	return m.width, m.height
}

func (m *AppModel) RequestExternalTUI(binary string, args []string, name string, confirmMsg string) {
	if name == "" {
		name = binary
	}
	msg := confirmMsg
	if msg == "" {
		msg = fmt.Sprintf("⚠️ '%s' eklentisi terminali devralarak kendi tam ekran arayüzünü (TUI) başlatmak istiyor.\n\nKomut: %s %s\n\nAna sisteme dönmek için eklenti arayüzünden çıkmanız ('q' veya Ctrl+C) yeterlidir.\n\nDevam etmek istiyor musunuz?", name, binary, strings.Join(args, " "))
	}

	m.pendingTUI = &superuser.PendingTUIRequest{
		Binary: binary,
		Args:   args,
		Name:   name,
	}

	m.OpenConfirm("HARİCİ TUI YÖNLENDİRME ONAYI", msg, func(confirmed bool) {
		if !confirmed {
			m.pendingTUI = nil
			m.Toast("Harici TUI yönlendirmesi iptal edildi.", false)
		}
	})
}

func (m *AppModel) ExecuteIPCAction(actionJSON string) string {
	return superuser.DispatchIPCAction(m, actionJSON)
}

// =========================================================================
// SUPER USER x1000 - APIS IMPLEMENTATION
// =========================================================================

func (m *AppModel) RegisterDynamicCommand(cmd, title, category, desc string, handler func(args []string, pipeIn string) string) {
	mf := plugin.Manifest{
		Command:      strings.ToLower(strings.TrimSpace(cmd)),
		Title:        title,
		Category:     category,
		Description:  desc,
		Capabilities: plugin.CapUI | plugin.CapSuperuser,
	}
	cmdPlugin := builtins.NewDynamicCmd(mf, handler)
	m.kernel.RegisterRuntimeCommand(cmd, cmdPlugin)
	m.Toast(fmt.Sprintf("⚡ [SUPER USER x1000] Komut çekirdeğe bağlandı: /%s", cmd), false)
}

func (m *AppModel) UnregisterDynamicCommand(cmd string) {
	m.kernel.UnregisterRuntimeCommand(cmd)
	m.Toast(fmt.Sprintf("⚡ [SUPER USER x1000] Komut kaldırıldı: /%s", cmd), false)
}

func (m *AppModel) TriggerCommand(cmdLine string) {
	m.SendTUI(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(cmdLine)})
	_, _ = m.executeCommandString(cmdLine)
}

func (m *AppModel) SetBanner(banner string) {
	m.customBanner = banner
}

func (m *AppModel) ResetBanner() {
	m.customBanner = ""
}

func (m *AppModel) SetStatus(msg string, isError bool) {
	m.statusMsg = msg
	m.isError = isError
}

func (m *AppModel) SetFooter(footer string) {
	m.customFooter = footer
}

func (m *AppModel) ResetFooter() {
	m.customFooter = ""
}

func (m *AppModel) SetInputValue(val string) {
	m.input.SetValue(val)
	m.input.SetCursor(len(val))
}

func (m *AppModel) SetInputPlaceholder(ph string) {
	m.input.Placeholder = ph
}

func (m *AppModel) SetFullscreenCustomView(viewFn func() string) {
	m.customFullscreenView = viewFn
}

func (m *AppModel) ClearFullscreenCustomView() {
	m.customFullscreenView = nil
	m.overrideScreenContent = ""
}

func (m *AppModel) TabCount() int {
	return len(m.tabs)
}

func (m *AppModel) ActiveTabIndex() int {
	return m.activeTab
}

func (m *AppModel) SwitchTab(idx int) bool {
	if idx >= 0 && idx < len(m.tabs) {
		m.activeTab = idx
		m.hasWorkspace = true
		return true
	}
	return false
}

func (m *AppModel) CloseAllTabs() {
	m.tabs = nil
	m.activeTab = 0
	m.hasWorkspace = false
	m.statusMsg = "Tüm çalışma sekmeleri kapatıldı."
}

func (m *AppModel) SetTabTitle(idx int, title string) bool {
	if idx >= 0 && idx < len(m.tabs) {
		m.tabs[idx].Title = title
		m.tabs[idx].Command = title
		return true
	}
	return false
}

func (m *AppModel) ScrollTab(idx int, delta int) {
	targetIdx := idx
	if targetIdx < 0 || targetIdx >= len(m.tabs) {
		targetIdx = m.activeTab
	}
	if targetIdx >= 0 && targetIdx < len(m.tabs) {
		if delta > 0 {
			m.tabs[targetIdx].Viewport.LineDown(delta)
		} else if delta < 0 {
			m.tabs[targetIdx].Viewport.LineUp(-delta)
		}
	}
}

func (m *AppModel) SetThemeColor(field string, hexColor string) bool {
	c := lipgloss.Color(hexColor)
	switch strings.ToLower(field) {
	case "primary":
		m.theme.Primary = c
	case "secondary":
		m.theme.Secondary = c
	case "bg_dark":
		m.theme.BgDark = c
	case "border_color":
		m.theme.BorderColor = c
	case "text_main":
		m.theme.TextMain = c
	case "text_muted":
		m.theme.TextMuted = c
	case "badge_bg":
		m.theme.BadgeBg = c
	default:
		return false
	}
	return true
}

func (m *AppModel) SimulateKey(key string) {
	switch strings.ToLower(key) {
	case "enter":
		m.SendTUI(tea.KeyMsg{Type: tea.KeyEnter})
	case "tab":
		m.SendTUI(tea.KeyMsg{Type: tea.KeyTab})
	case "esc", "escape":
		m.SendTUI(tea.KeyMsg{Type: tea.KeyEsc})
	case "up":
		m.SendTUI(tea.KeyMsg{Type: tea.KeyUp})
	case "down":
		m.SendTUI(tea.KeyMsg{Type: tea.KeyDown})
	default:
		m.SendTUI(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	}
}

func (m *AppModel) InspectKernel() map[string]any {
	var tabInfo []map[string]any
	for i, t := range m.tabs {
		tabInfo = append(tabInfo, map[string]any{
			"index":    i,
			"command":  t.Command,
			"title":    t.Title,
			"duration": t.Duration.String(),
			"lines":    t.LineCount(),
		})
	}

	stateDump := m.store.GetAll()
	kernelData := m.kernel.Inspect()

	return map[string]any{
		"theme":          m.theme.Name,
		"tabs_count":     len(m.tabs),
		"active_tab_idx": m.activeTab,
		"tabs":           tabInfo,
		"kv_state":       stateDump,
		"kernel":         kernelData,
		"terminal_w":     m.width,
		"terminal_h":     m.height,
	}
}

func (m *AppModel) StreamAppend(chunk string) bool {
	if len(m.tabs) > 0 && m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		m.tabs[m.activeTab].Output += chunk
		normalized := strings.ReplaceAll(m.tabs[m.activeTab].Output, "\r\n", "\n")
		normalized = strings.ReplaceAll(normalized, "\r", "\n")
		m.tabs[m.activeTab].Viewport.SetContent(strings.TrimRight(normalized, "\n"))
		m.tabs[m.activeTab].Viewport.GotoBottom()
		return true
	}
	return false
}

func (m *AppModel) EmitEvent(eventName string, payload any) {
	evtJSON, _ := json.Marshal(payload)
	m.store.Set("EVENT_"+strings.ToUpper(eventName), string(evtJSON))
	m.Toast(fmt.Sprintf("📡 [EVENT] %s yayınlandı", eventName), false)
}

func (m *AppModel) RenderTable(headers []string, rows [][]string) string {
	if len(headers) == 0 && len(rows) == 0 {
		return ""
	}
	colCount := len(headers)
	for _, r := range rows {
		if len(r) > colCount {
			colCount = len(r)
		}
	}
	colWidths := make([]int, colCount)
	for i, h := range headers {
		if len(h) > colWidths[i] {
			colWidths[i] = len(h)
		}
	}
	for _, r := range rows {
		for i, cell := range r {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	for i := range colWidths {
		colWidths[i] = colWidths[i] + 2 // padding
		if colWidths[i] < 6 {
			colWidths[i] = 6
		}
	}

	var sb strings.Builder
	// Top border
	sb.WriteString("╭")
	for i, w := range colWidths {
		sb.WriteString(strings.Repeat("─", w))
		if i < len(colWidths)-1 {
			sb.WriteString("┬")
		}
	}
	sb.WriteString("╮\n")

	// Header row
	if len(headers) > 0 {
		sb.WriteString("│")
		for i, w := range colWidths {
			h := ""
			if i < len(headers) {
				h = headers[i]
			}
			cellStr := lipgloss.NewStyle().Width(w).Foreground(m.theme.Primary).Bold(true).Render(" " + h)
			sb.WriteString(cellStr + "│")
		}
		sb.WriteString("\n├")
		for i, w := range colWidths {
			sb.WriteString(strings.Repeat("─", w))
			if i < len(colWidths)-1 {
				sb.WriteString("┼")
			}
		}
		sb.WriteString("┤\n")
	}

	// Data rows
	for _, r := range rows {
		sb.WriteString("│")
		for i, w := range colWidths {
			val := ""
			if i < len(r) {
				val = r[i]
			}
			cellStr := lipgloss.NewStyle().Width(w).Foreground(m.theme.TextMain).Render(" " + val)
			sb.WriteString(cellStr + "│")
		}
		sb.WriteString("\n")
	}

	// Bottom border
	sb.WriteString("╰")
	for i, w := range colWidths {
		sb.WriteString(strings.Repeat("─", w))
		if i < len(colWidths)-1 {
			sb.WriteString("┴")
		}
	}
	sb.WriteString("╯\n")

	return sb.String()
}

func (m *AppModel) RenderSparkline(values []float64, title string, color string) string {
	if len(values) == 0 {
		return ""
	}
	minV := values[0]
	maxV := values[0]
	for _, v := range values {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	blocks := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var sb strings.Builder
	fgCol := m.theme.Secondary
	if color != "" {
		fgCol = lipgloss.Color(color)
	}

	for _, v := range values {
		idx := 0
		if maxV > minV {
			ratio := (v - minV) / (maxV - minV)
			idx = int(ratio * float64(len(blocks)-1))
		}
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		sb.WriteRune(blocks[idx])
	}
	chartStr := lipgloss.NewStyle().Foreground(fgCol).Bold(true).Render(sb.String())
	titleStr := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Render(title)
	statsStr := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(fmt.Sprintf(" [min: %.1f | max: %.1f | curr: %.1f]", minV, maxV, values[len(values)-1]))

	return fmt.Sprintf("%s: %s %s\n", titleStr, chartStr, statsStr)
}

func (m *AppModel) RenderBarChart(data map[string]float64, title string, color string) string {
	if len(data) == 0 {
		return ""
	}
	var maxKeyLen int
	var maxVal float64
	for k, v := range data {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	fgCol := m.theme.Primary
	if color != "" {
		fgCol = lipgloss.Color(color)
	}

	barWidth := 24
	var lines []string
	if title != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(fgCol).Bold(true).Render(title), strings.Repeat("─", 44))
	}

	for k, v := range data {
		kPadded := lipgloss.NewStyle().Width(maxKeyLen + 2).Foreground(m.theme.TextMuted).Render(k + ":")
		ratio := v / maxVal
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		filled := int(ratio * float64(barWidth))
		empty := barWidth - filled
		bar := lipgloss.NewStyle().Foreground(fgCol).Bold(true).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(m.theme.BadgeBg).Render(strings.Repeat("░", empty))
		valStr := lipgloss.NewStyle().Foreground(m.theme.TextMain).Bold(true).Render(fmt.Sprintf(" %.1f", v))
		lines = append(lines, fmt.Sprintf("%s %s%s", kPadded, bar, valStr))
	}

	return strings.Join(lines, "\n")
}

func (m *AppModel) StartWatchdog(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(3000 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if changed, count := m.kernel.CheckAndReload(config.GetPluginsDir()); changed {
					m.SendTUI(HotReloadMsg{Count: count})
				}
			}
		}
	}()
}

// NewAppModel initializes all sub-engines and built-in commands.
func NewAppModel() *AppModel {
	cfg := config.LoadConfig()
	initialTheme, _ := theme.GetByName(cfg.Theme)

	store := state.NewSharedStateStore(config.GetStatePath())

	ti := textinput.New()
	ti.Placeholder = "komut yazın (/help, /sys, /tui htop, /sh ls | grep main)..."
	ti.Focus()
	ti.CharLimit = 160
	ti.Prompt = ""
	ti.Width = 60

	m := &AppModel{
		input:          ti,
		config:         cfg,
		theme:          initialTheme,
		hookManager:    superuser.NewSuperuserHookManager(),
		store:          store,
		statusMsg:      "MT Core v12.0 (Super User x1000) devrede. Komut için '/' yazın.",
		historyIdx:     -1,
		lastInputVal:   "",
		pixelOverrides: make(map[string]framebuffer.Cell),
		keybindings:    make(map[string]string),
	}

	m.ResetPrompt()
	m.kernel = kernel.New(m)

	_ = m.kernel.RegisterBuiltin(builtins.NewHelpCommand(m.kernel))
	_ = m.kernel.RegisterBuiltin(builtins.NewPluginManagerCommand(m.kernel))
	_ = m.kernel.RegisterBuiltin(&builtins.TuiHandoverPlugin{})
	_ = m.kernel.RegisterBuiltin(&builtins.SysDaemonPlugin{})
	_ = m.kernel.RegisterBuiltin(&builtins.ShellPlugin{})
	_ = m.kernel.RegisterBuiltin(&builtins.ThemeSwitchPlugin{})

	_ = m.kernel.RegisterBuiltin(builtins.NewSimpleCmd(
		plugin.Manifest{Command: "clear", Title: "EKRANI TEMİZLE", Category: "CORE", Description: "Tüm çalışma sekmelerini kapatır", Capabilities: plugin.CapUI},
		func() string { return "" },
	))
	_ = m.kernel.RegisterBuiltin(builtins.NewSimpleCmd(
		plugin.Manifest{Command: "exit", Title: "GÜVENLİ ÇIKIŞ", Category: "CORE", Description: "Terminal oturumunu sonlandırır", Capabilities: plugin.CapUI},
		func() string { return "" },
	))

	_, modCount := m.kernel.CheckAndReload(config.GetPluginsDir())
	allCmds := m.kernel.Search("")
	m.statusMsg = fmt.Sprintf("MT Core v12.0 hazır. (%d komut [%d eklenti] | Tema: %s)", len(allCmds), modCount, m.theme.Name)

	return m
}

func (m *AppModel) Init() tea.Cmd {
	return textinput.Blink
}

// Run starts the Bubbletea terminal program with alt screen and mouse motion.
func Run() error {
	model := NewAppModel()
	ctx, cancel := context.WithCancel(context.Background())
	model.watchCancel = cancel
	defer cancel()

	// Termux or minimal terminals may not support xterm mouse cell motion or alternate screens cleanly.
	// We dynamically configure program options according to environment.
	isTermux := os.Getenv("PREFIX") != "" && strings.Contains(os.Getenv("PREFIX"), "com.termux")
	termEnv := strings.ToLower(os.Getenv("TERM"))
	isDumbTerm := termEnv == "dumb" || termEnv == ""

	var opts []tea.ProgramOption
	if !isDumbTerm {
		opts = append(opts, tea.WithAltScreen())
		// Only enable mouse cell motion if not in dumb terminal and not explicitly disabled
		if os.Getenv("MT_NO_MOUSE") != "1" && !isTermux {
			opts = append(opts, tea.WithMouseCellMotion())
		}
	}

	p := tea.NewProgram(model, opts...)
	model.program = p

	model.StartWatchdog(ctx)

	_, err := p.Run()
	return err
}
