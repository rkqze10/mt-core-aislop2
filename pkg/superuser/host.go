package superuser

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rkqze10/mt-tui/pkg/framebuffer"
	"github.com/rkqze10/mt-tui/pkg/state"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

// DynamicPixelHook allows programmatic manipulation of the virtual framebuffer during render.
type DynamicPixelHook func(fb *framebuffer.VirtualFramebuffer, width, height int)

// CellMutation represents a single pixel change requested over IPC or API.
type CellMutation struct {
	X    int                    `json:"x"`
	Y    int                    `json:"y"`
	Char rune                   `json:"char"`
	Fg   lipgloss.TerminalColor `json:"fg"`
	Bg   lipgloss.TerminalColor `json:"bg"`
	Bold bool                   `json:"bold"`
}

// PendingTUIRequest stores parameters for full-screen external TUI handoff.
type PendingTUIRequest struct {
	Binary string
	Args   []string
	Name   string
}

// ModalDialog defines an active popup or confirm box.
type ModalDialog struct {
	Title        string
	Content      string
	IsPrompt     bool
	IsConfirm    bool
	PromptPrompt string
	InputModel   *textinput.Model
	Callback     func(result string)
	OnConfirm    func(confirmed bool)
}

// SuperuserHost provides plugins with ring-0 level access to UI, state, terminal, and execution.
type SuperuserHost interface {
	SendTUI(msg tea.Msg)
	ExecOS(name string, args ...string) (string, error)
	Notify(msg string, isError bool)
	Toast(text string, isError bool)
	OpenModal(title, content string)
	OpenPrompt(title, promptText string, onSubmit func(val string))
	OpenConfirm(title, message string, onConfirm func(confirmed bool))
	CloseModal()
	SetTheme(name string) bool
	SetCustomTheme(theme theme.ThemePalette)
	CurrentTheme() theme.ThemePalette
	ClipboardSet(text string)
	KV() *state.SharedStateStore

	MutatePixel(x, y int, ch rune, fg, bg lipgloss.TerminalColor, bold bool)
	MutatePixelBatch(pixels []CellMutation)
	DrawBox(x, y, w, h int, title string, borderFg lipgloss.TerminalColor, bold bool)
	ClearPixels()
	RegisterOverlay(id string, hook DynamicPixelHook)
	RemoveOverlay(id string)
	SetTextHook(target, content string)
	ClearTextHook(target string)
	SetPrompt(prefix, arrow string, fgPrefix, fgArrow lipgloss.TerminalColor)
	ResetPrompt()
	BindKey(key, command string)
	UnbindKey(key string)
	PatchCommand(targetCmd, rewriteCmd string)
	UnpatchCommand(targetCmd string)
	ForceCloseTab(tabIdx int) bool
	InjectTab(title, cmd, output string)
	UpdateTabOutput(tabIdx int, output string) bool
	SetRawScreenContent(rendered string)
	HijackInputFocus(active bool)
	GetTerminalDimensions() (width, height int)
	ExecuteIPCAction(actionJSON string) string

	// External TUI Handover (with user confirmation & guaranteed return)
	RequestExternalTUI(binary string, args []string, name string, confirmMsg string)

	// =========================================================================
	// SUPER USER x1000 - CORE KERNEL & UI HYPERVISOR CONTROL APIS
	// =========================================================================
	RegisterDynamicCommand(cmd, title, category, desc string, handler func(args []string, pipeIn string) string)
	UnregisterDynamicCommand(cmd string)
	TriggerCommand(cmdLine string)
	SetBanner(banner string)
	ResetBanner()
	SetStatus(msg string, isError bool)
	SetFooter(footer string)
	ResetFooter()
	SetInputValue(val string)
	SetInputPlaceholder(ph string)
	SetFullscreenCustomView(viewFn func() string)
	ClearFullscreenCustomView()
	TabCount() int
	ActiveTabIndex() int
	SwitchTab(idx int) bool
	CloseAllTabs()
	SetTabTitle(idx int, title string) bool
	ScrollTab(idx int, delta int)
	SetThemeColor(field string, hexColor string) bool
	SimulateKey(key string)
	InspectKernel() map[string]any

	// Real-Time Streaming, Event Bus & Data Visualization APIs
	StreamAppend(chunk string) bool
	EmitEvent(eventName string, payload any)
	RenderTable(headers []string, rows [][]string) string
	RenderSparkline(values []float64, title string, color string) string
	RenderBarChart(data map[string]float64, title string, color string) string
}
