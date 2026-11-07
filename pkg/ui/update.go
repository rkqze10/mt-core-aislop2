package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/config"
	"github.com/rkqze10/mt-tui/pkg/pipeline"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

func (m *AppModel) executeCommandString(rawInput string) (tea.Model, tea.Cmd) {
	inputStr := strings.TrimSpace(rawInput)
	m.input.SetValue("")
	m.selectedIndex = 0
	m.historyIdx = -1
	m.historyDraft = ""
	m.lastInputVal = ""

	if inputStr == "" {
		return m, nil
	}

	if !strings.HasPrefix(inputStr, "/") {
		firstWord := strings.ToLower(strings.Split(inputStr, " ")[0])
		if _, ok := m.kernel.Resolve(firstWord); ok {
			inputStr = "/" + inputStr
		} else {
			inputStr = "/sh " + inputStr
		}
	}

	proceed, filtered := m.kernel.FilterCommand(inputStr)
	if !proceed {
		m.statusMsg = "Komut güvenlik filtresi tarafından durduruldu."
		m.isError = true
		return m, nil
	}

	histLen := len(m.config.History)
	if histLen == 0 || m.config.History[histLen-1] != filtered {
		m.config.History = append(m.config.History, filtered)
		if len(m.config.History) > 100 {
			m.config.History = m.config.History[len(m.config.History)-100:]
		}
		_ = config.SaveConfig(m.config)
	}

	stages := pipeline.ParseExecutionLine(filtered)
	if len(stages) == 0 {
		return m, nil
	}

	var lastOutput string
	boxWidth, vpHeight, _ := m.calculateLayout()
	safeBoxW := boxWidth - 6
	if safeBoxW < 10 {
		safeBoxW = 10
	}
	safeVpH := vpHeight
	if safeVpH < 3 {
		safeVpH = 3
	}

	for _, stage := range stages {
		if stage.CmdName == "" {
			continue
		}

		switch stage.CmdName {
		case "exit", "quit":
			if m.watchCancel != nil {
				m.watchCancel()
			}
			m.kernel.Shutdown()
			return m, tea.Quit
		case "clear":
			m.hasWorkspace = false
			m.tabs = nil
			m.activeTab = 0
			m.statusMsg = "Ekran temizlendi."
			m.isError = false
			return m, nil
		}

		if plug, ok := m.kernel.Resolve(stage.CmdName); ok {
			pipeIn := ""
			if stage.IsPiped {
				pipeIn = lastOutput
			}

			cmdSnapshot := stage.CmdName
			argsSnapshot := stage.Args
			pipeSnapshot := pipeIn

			execFunc := func(th theme.ThemePalette, w, h int) string {
				ctx := &plugin.PluginContext{
					Width:      w,
					Height:     h,
					InvokedCmd: cmdSnapshot,
					Args:       argsSnapshot,
					PipeInput:  pipeSnapshot,
					Theme:      th,
					Host:       m,
				}
				return plug.View(ctx)
			}

			startT := time.Now()
			out := execFunc(m.theme, safeBoxW, safeVpH)
			dur := time.Since(startT)
			lastOutput = out

			cmdLabel := "/" + stage.CmdName
			if len(stage.Args) > 0 {
				cmdLabel += " " + strings.Join(stage.Args, " ")
			}
			if stage.IsPiped {
				cmdLabel = "| " + cmdLabel
			}
			if len(cmdLabel) > 20 {
				cmdLabel = cmdLabel[:17] + "..."
			}

			var tabRenderFn TabRenderFunc
			switch stage.CmdName {
			case "sys", "help":
				tabRenderFn = execFunc
			}

			m.tabCounter++
			newTab := NewWorkspaceTab(m.tabCounter, cmdLabel, stage.CmdName, out, tabRenderFn, dur, safeBoxW, safeVpH)
			m.tabs = append(m.tabs, newTab)
			m.activeTab = len(m.tabs) - 1
			m.hasWorkspace = true
		} else {
			m.statusMsg = fmt.Sprintf("Tanımlanamayan komut: '/%s'. Yardım için: /help", stage.CmdName)
			m.isError = true
			return m, nil
		}
	}

	m.statusMsg = fmt.Sprintf("Sekme #%d açıldı. (Alt+1..%d geçiş | Ctrl+Y kopyala)", len(m.tabs), len(m.tabs))
	m.isError = false
	return m, nil
}

func (m *AppModel) Update(msg tea.Msg) (retModel tea.Model, cmd tea.Cmd) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			_ = os.WriteFile("mt_panic.log", []byte(fmt.Sprintf("PANIC IN UPDATE: %v\n\n%s", r, stack)), 0644)
			m.statusMsg = fmt.Sprintf("⚠️ Çekirdek Hatası (mt_panic.log): %v", r)
			m.isError = true
			m.input.Focus()
			retModel = m
			cmd = nil
		}
	}()

	for _, it := range m.kernel.Builtins() {
		ctx := &plugin.PluginContext{Host: m, Width: m.width, Height: m.height, Theme: m.theme}
		if consumed, hijackedCmd := it.Update(msg, ctx); consumed {
			return m, hijackedCmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		boxWidth, vpHeight, _ := m.calculateLayout()
		inputW := boxWidth - 10
		if inputW < 10 {
			inputW = 10
		}
		m.input.Width = inputW

		safeBoxW := boxWidth - 6
		if safeBoxW < 10 {
			safeBoxW = 10
		}
		safeVpH := vpHeight
		if safeVpH < 2 {
			safeVpH = 2
		}

		for i := range m.tabs {
			m.tabs[i].Viewport.Width = safeBoxW
			m.tabs[i].Viewport.Height = safeVpH
			if m.tabs[i].RenderFn != nil {
				m.tabs[i].Output = m.tabs[i].RenderFn(m.theme, safeBoxW, safeVpH)
			}
			m.tabs[i].Viewport.SetContent(m.tabs[i].Output)
			m.tabs[i].Viewport.GotoTop()
		}
		return m, nil

	case ExternalTUIFinishedMsg:
		m.statusMsg = "✔ Harici TUI oturumundan ana sisteme dönüldü."
		m.isError = false
		if msg.Err != nil {
			m.Toast(fmt.Sprintf("Harici TUI sonlandı: %v", msg.Err), false)
		} else {
			m.Toast("✔ Harici TUI oturumu başarıyla tamamlandı.", false)
		}
		return m, nil

	case NotifyMsg:
		m.statusMsg = msg.Text
		m.isError = msg.IsError
		return m, nil

	case HotReloadMsg:
		allCmds := m.kernel.Search("")
		m.statusMsg = fmt.Sprintf("Modüller güncellendi. (%d komut aktif [%d eklenti])", len(allCmds), msg.Count)
		return m, nil

	case ToastTriggerMsg:
		m.activeToast = &ToastMsg{ID: msg.ID, Text: msg.Text, IsError: msg.IsError}
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ToastClearMsg{ID: msg.ID}
		})

	case ToastClearMsg:
		if m.activeToast != nil && m.activeToast.ID == msg.ID {
			m.activeToast = nil
		}
		return m, nil

	case tea.MouseMsg:
		if m.hasWorkspace && len(m.tabs) > 0 && m.activeTab >= 0 && m.activeTab < len(m.tabs) {
			switch msg.Type {
			case tea.MouseWheelUp:
				m.tabs[m.activeTab].Viewport.LineUp(3)
				return m, nil
			case tea.MouseWheelDown:
				m.tabs[m.activeTab].Viewport.LineDown(3)
				return m, nil
			}
		}

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			if m.activeModal != nil {
				m.CloseModal()
				return m, nil
			}
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.selectedIndex = 0
				m.historyIdx = -1
				m.historyDraft = ""
				m.lastInputVal = ""
				return m, nil
			}
			if m.watchCancel != nil {
				m.watchCancel()
			}
			m.kernel.Shutdown()
			return m, tea.Quit
		}

		m.keybindMu.RLock()
		boundCmd, hasKeybind := m.keybindings[strings.ToLower(msg.String())]
		m.keybindMu.RUnlock()
		if hasKeybind && boundCmd != "" {
			return m.executeCommandString(boundCmd)
		}

		if m.activeModal != nil {
			if m.activeModal.IsConfirm {
				switch strings.ToLower(msg.String()) {
				case "y", "enter":
					cb := m.activeModal.OnConfirm
					m.CloseModal()
					if cb != nil {
						cb(true)
					}
					if m.pendingTUI != nil {
						req := m.pendingTUI
						m.pendingTUI = nil

						if _, err := exec.LookPath(req.Binary); err != nil {
							m.Toast(fmt.Sprintf("Program bulunamadı: %s", req.Binary), true)
							return m, nil
						}

						cmdExec := exec.Command(req.Binary, req.Args...)
						return m, tea.ExecProcess(cmdExec, func(err error) tea.Msg {
							return ExternalTUIFinishedMsg{Err: err}
						})
					}
					return m, nil

				case "n", "esc", "q":
					cb := m.activeModal.OnConfirm
					m.CloseModal()
					if cb != nil {
						cb(false)
					}
					m.pendingTUI = nil
					return m, nil
				}
				return m, nil
			}

			if msg.String() == "esc" || (!m.activeModal.IsPrompt && msg.String() == "q") {
				m.CloseModal()
				return m, nil
			}

			if m.activeModal.IsPrompt && m.activeModal.InputModel != nil {
				if msg.String() == "enter" {
					val := m.activeModal.InputModel.Value()
					cb := m.activeModal.Callback
					m.CloseModal()
					if cb != nil {
						cb(val)
					}
					return m, nil
				}
				updatedTi, tiCmd := m.activeModal.InputModel.Update(msg)
				*m.activeModal.InputModel = updatedTi
				return m, tiCmd
			}

			if msg.String() == "enter" {
				m.CloseModal()
				return m, nil
			}
			return m, nil
		}

		if msg.Type == tea.KeyCtrlY {
			if m.hasWorkspace && len(m.tabs) > 0 && m.activeTab >= 0 && m.activeTab < len(m.tabs) {
				m.ClipboardSet(m.tabs[m.activeTab].Output)
				m.Toast("📋 Sekme çıktısı panoya kopyalandı!", false)
				return m, nil
			}
		}

		if msg.Type == tea.KeyCtrlW {
			if m.input.Value() != "" {
				// Delete previous word in input line
				val := m.input.Value()
				trimmed := strings.TrimRight(val, " ")
				lastSpace := strings.LastIndex(trimmed, " ")
				if lastSpace >= 0 {
					newVal := trimmed[:lastSpace+1]
					m.input.SetValue(newVal)
					m.input.SetCursor(len(newVal))
				} else {
					m.input.SetValue("")
					m.input.SetCursor(0)
				}
				return m, nil
			}
			if m.hasWorkspace && len(m.tabs) > 0 {
				m.ForceCloseTab(m.activeTab)
				m.Toast("Sekme kapatıldı.", false)
				return m, nil
			}
		}

		switch msg.String() {
		case "alt+right", "alt+l":
			if len(m.tabs) > 0 {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				m.hasWorkspace = true
				return m, nil
			}
		case "alt+left", "alt+h":
			if len(m.tabs) > 0 {
				m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
				m.hasWorkspace = true
				return m, nil
			}
		}

		if strings.HasPrefix(msg.String(), "alt+") {
			keyNum := strings.TrimPrefix(msg.String(), "alt+")
			if idx, err := strconv.Atoi(keyNum); err == nil && idx >= 1 && idx <= len(m.tabs) {
				m.activeTab = idx - 1
				m.hasWorkspace = true
				return m, nil
			}
		}

		if m.hasWorkspace && len(m.tabs) > 0 && m.activeTab >= 0 && m.activeTab < len(m.tabs) {
			switch msg.Type {
			case tea.KeyPgUp:
				m.tabs[m.activeTab].Viewport.HalfViewUp()
				return m, nil
			case tea.KeyPgDown:
				m.tabs[m.activeTab].Viewport.HalfViewDown()
				return m, nil
			}
		}

		raw := m.input.Value()
		isSearching := strings.HasPrefix(raw, "/")
		matches := m.kernel.Search(raw)

		switch msg.String() {
		case "esc":
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.selectedIndex = 0
				m.historyIdx = -1
				m.historyDraft = ""
				m.lastInputVal = ""
				return m, nil
			}
			if m.hasWorkspace {
				m.hasWorkspace = false
				m.tabs = nil
				m.activeTab = 0
				m.statusMsg = "Çalışma alanı kapatıldı."
				return m, nil
			}
			return m, nil

		case "up":
			if m.historyIdx == -1 && isSearching && len(matches) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(matches) - 1
				}
				return m, nil
			}
			if len(m.config.History) > 0 {
				if m.historyIdx == -1 {
					m.historyDraft = m.input.Value()
				}
				if m.historyIdx < len(m.config.History)-1 {
					m.historyIdx++
					target := m.config.History[len(m.config.History)-1-m.historyIdx]
					m.input.SetValue(target)
					m.input.SetCursor(len(target))
					m.lastInputVal = target
				}
				return m, nil
			}

		case "down":
			if m.historyIdx == -1 && isSearching && len(matches) > 0 {
				m.selectedIndex++
				if m.selectedIndex >= len(matches) {
					m.selectedIndex = 0
				}
				return m, nil
			}
			if m.historyIdx >= 0 {
				m.historyIdx--
				if m.historyIdx < 0 {
					m.input.SetValue(m.historyDraft)
					m.input.SetCursor(len(m.historyDraft))
					m.lastInputVal = m.historyDraft
					m.historyDraft = ""
				} else {
					target := m.config.History[len(m.config.History)-1-m.historyIdx]
					m.input.SetValue(target)
					m.input.SetCursor(len(target))
					m.lastInputVal = target
				}
				return m, nil
			}

		case "tab":
			if isSearching && len(matches) > 0 {
				targetIdx := m.selectedIndex
				if targetIdx >= len(matches) || targetIdx < 0 {
					targetIdx = 0
				}
				selectedCmd := "/" + matches[targetIdx].Command + " "
				m.input.SetValue(selectedCmd)
				m.input.SetCursor(len(selectedCmd))
				m.lastInputVal = selectedCmd
				return m, nil
			}

		case "enter":
			return m.executeCommandString(m.input.Value())
		}
	}

	m.input, cmd = m.input.Update(msg)

	currVal := m.input.Value()
	if currVal != m.lastInputVal {
		m.selectedIndex = 0
		m.lastInputVal = currVal
	}

	return m, cmd
}
