package superuser

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rkqze10/mt-tui/pkg/state"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

type mockHost struct {
	lastToast     string
	lastStream    string
	lastEvent     string
	lastEventData any
	kvStore       *state.SharedStateStore
}

func newMockHost() *mockHost {
	return &mockHost{
		kvStore: state.NewSharedStateStore(""),
	}
}

func (m *mockHost) SendTUI(msg tea.Msg) {}
func (m *mockHost) ExecOS(name string, args ...string) (string, error) { return "", nil }
func (m *mockHost) Notify(msg string, isError bool) {}
func (m *mockHost) Toast(text string, isError bool) { m.lastToast = text }
func (m *mockHost) OpenModal(title, content string) {}
func (m *mockHost) OpenPrompt(title, promptText string, onSubmit func(val string)) {}
func (m *mockHost) OpenConfirm(title, message string, onConfirm func(confirmed bool)) {}
func (m *mockHost) CloseModal() {}
func (m *mockHost) SetTheme(name string) bool { return true }
func (m *mockHost) SetCustomTheme(t theme.ThemePalette) {}
func (m *mockHost) CurrentTheme() theme.ThemePalette { return theme.ThemeAmber }
func (m *mockHost) ClipboardSet(text string) {}
func (m *mockHost) KV() *state.SharedStateStore { return m.kvStore }

func (m *mockHost) MutatePixel(x, y int, ch rune, fg, bg lipgloss.TerminalColor, bold bool) {}
func (m *mockHost) MutatePixelBatch(pixels []CellMutation) {}
func (m *mockHost) DrawBox(x, y, w, h int, title string, borderFg lipgloss.TerminalColor, bold bool) {}
func (m *mockHost) ClearPixels() {}
func (m *mockHost) RegisterOverlay(id string, hook DynamicPixelHook) {}
func (m *mockHost) RemoveOverlay(id string) {}
func (m *mockHost) SetTextHook(target, content string) {}
func (m *mockHost) ClearTextHook(target string) {}
func (m *mockHost) SetPrompt(prefix, arrow string, fgPrefix, fgArrow lipgloss.TerminalColor) {}
func (m *mockHost) ResetPrompt() {}
func (m *mockHost) BindKey(key, command string) {}
func (m *mockHost) UnbindKey(key string) {}
func (m *mockHost) PatchCommand(target, rewrite string) {}
func (m *mockHost) UnpatchCommand(target string) {}
func (m *mockHost) RequestExternalTUI(binary string, args []string, name, prompt string) {}
func (m *mockHost) ForceCloseTab(index int) bool { return true }
func (m *mockHost) InjectTab(title, cmd, output string) {}
func (m *mockHost) UpdateTabOutput(index int, output string) bool { return true }
func (m *mockHost) SetRawScreenContent(rendered string) {}
func (m *mockHost) HijackInputFocus(active bool) {}
func (m *mockHost) ActiveTabIndex() int { return 0 }
func (m *mockHost) TabCount() int { return 1 }

func (m *mockHost) RegisterDynamicCommand(cmd, title, cat, desc string, handler func(args []string, pipeIn string) string) {}
func (m *mockHost) UnregisterDynamicCommand(cmd string) {}
func (m *mockHost) TriggerCommand(cmdLine string) {}
func (m *mockHost) SetBanner(banner string) {}
func (m *mockHost) ResetBanner() {}
func (m *mockHost) SetStatus(msg string, isError bool) {}
func (m *mockHost) SetFooter(deck string) {}
func (m *mockHost) ResetFooter() {}
func (m *mockHost) SetInputValue(val string) {}
func (m *mockHost) SetInputPlaceholder(ph string) {}
func (m *mockHost) SetFullscreenCustomView(fn func() string) {}
func (m *mockHost) ClearFullscreenCustomView() {}
func (m *mockHost) SimulateKey(key string) {}
func (m *mockHost) InspectKernel() map[string]any { return map[string]any{"status": "ok"} }

func (m *mockHost) SwitchTab(index int) bool { return true }
func (m *mockHost) CloseAllTabs() {}
func (m *mockHost) SetTabTitle(index int, title string) bool { return true }
func (m *mockHost) ScrollTab(index, delta int) {}
func (m *mockHost) SetThemeColor(field, hexColor string) bool { return true }
func (m *mockHost) ExecuteIPCAction(rawJSON string) string { return DispatchIPCAction(m, rawJSON) }
func (m *mockHost) GetTerminalDimensions() (width, height int) { return 80, 24 }

func (m *mockHost) StreamAppend(chunk string) bool {
	m.lastStream += chunk
	return true
}

func (m *mockHost) EmitEvent(eventName string, payload any) {
	m.lastEvent = eventName
	m.lastEventData = payload
}

func (m *mockHost) RenderTable(headers []string, rows [][]string) string {
	return strings.Join(headers, "|")
}

func (m *mockHost) RenderSparkline(values []float64, title string, color string) string {
	return title + ":sparkline"
}

func (m *mockHost) RenderBarChart(data map[string]float64, title string, color string) string {
	return title + ":barchart"
}

func TestDispatchIPCAction_AdvancedFeatures(t *testing.T) {
	host := newMockHost()

	// 1. Toast
	res := DispatchIPCAction(host, `{"action":"toast","message":"Hello world"}`)
	if !strings.Contains(res, "Hello world") || host.lastToast != "Hello world" {
		t.Fatalf("Toast failed, got %q", res)
	}

	// 2. Stream append
	DispatchIPCAction(host, `{"action":"stream_append","chunk":"Token1 "}`)
	DispatchIPCAction(host, `{"action":"stream_append","chunk":"Token2"}`)
	if host.lastStream != "Token1 Token2" {
		t.Fatalf("Stream failed, got %q", host.lastStream)
	}

	// 3. Emit event
	resEvt := DispatchIPCAction(host, `{"action":"emit_event","event":"test_evt","data":{"foo":"bar"}}`)
	if !strings.Contains(resEvt, "test_evt") || host.lastEvent != "test_evt" {
		t.Fatalf("Emit event failed, got %q", resEvt)
	}

	// 4. KV Set & Delete
	DispatchIPCAction(host, `{"action":"kv_set","key":"ALPHA","value":"123"}`)
	val, ok := host.KV().Get("ALPHA")
	if !ok || val != "123" {
		t.Fatalf("KV set failed")
	}
	DispatchIPCAction(host, `{"action":"kv_del","key":"ALPHA"}`)
	val, ok = host.KV().Get("ALPHA")
	if ok || val != "" {
		t.Fatalf("KV del failed")
	}

	// 5. Table render
	tableRes := DispatchIPCAction(host, `{"action":"table","headers":["A","B"],"rows":[["1","2"]]}`)
	if tableRes != "A|B" {
		t.Fatalf("Table render failed: %q", tableRes)
	}

	// 6. Sparkline render
	sparkRes := DispatchIPCAction(host, `{"action":"sparkline","values":[1,2,3],"title":"Latency"}`)
	if sparkRes != "Latency:sparkline" {
		t.Fatalf("Sparkline render failed: %q", sparkRes)
	}

	// 7. Barchart render
	barRes := DispatchIPCAction(host, `{"action":"barchart","data":{"a":10},"title":"Traffic"}`)
	if barRes != "Traffic:barchart" {
		t.Fatalf("Barchart render failed: %q", barRes)
	}
}
