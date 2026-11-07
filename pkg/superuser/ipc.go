package superuser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rkqze10/mt-tui/pkg/framebuffer"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

// DispatchIPCAction parses a raw JSON string and executes the corresponding action on the host.
func DispatchIPCAction(host SuperuserHost, rawJSON string) string {
	rawTrimmed := strings.TrimSpace(rawJSON)
	if !strings.HasPrefix(rawTrimmed, "{") || !strings.HasSuffix(rawTrimmed, "}") {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(rawTrimmed), &payload); err != nil {
		return ""
	}

	getString := func(key string) string {
		val, ok := payload[key]
		if !ok || val == nil {
			return ""
		}
		return fmt.Sprintf("%v", val)
	}

	getInt := func(key string, defaultVal int) int {
		val, ok := payload[key]
		if !ok || val == nil {
			return defaultVal
		}
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
		return defaultVal
	}

	action := strings.ToLower(getString("action"))
	if action == "" {
		action = strings.ToLower(getString("type"))
	}

	switch action {
	// =========================================================================
	// SUPER USER x1000: DYNAMIC RUNTIME COMMAND INJECTION
	// =========================================================================
	case "register_cmd", "add_command":
		cmd := getString("command")
		title := getString("title")
		if title == "" {
			title = strings.ToUpper(cmd)
		}
		cat := strings.ToUpper(getString("category"))
		if cat == "" {
			cat = "ROOT-1000"
		}
		desc := getString("desc")
		if desc == "" {
			desc = "Dinamik süper kullanıcı komutu"
		}
		execScript := getString("exec")
		shExec := getString("sh")
		if shExec != "" {
			execScript = shExec
		}

		host.RegisterDynamicCommand(cmd, title, cat, desc, func(args []string, pipeIn string) string {
			if execScript != "" {
				fullCmd := execScript
				if len(args) > 0 {
					fullCmd += " " + strings.Join(args, " ")
				}
				shBin := "sh"
				if prefix := os.Getenv("PREFIX"); prefix != "" {
					candidate := prefix + "/bin/sh"
					if _, err := os.Stat(candidate); err == nil {
						shBin = candidate
					}
				} else if p, err := exec.LookPath("bash"); err == nil {
					shBin = p
				} else if p, err := exec.LookPath("sh"); err == nil {
					shBin = p
				}
				out, err := exec.Command(shBin, "-c", fullCmd).CombinedOutput()
				if err != nil {
					return fmt.Sprintf("⚠️ Komut hatası: %v\n%s", err, string(out))
				}
				return strings.TrimSpace(string(out))
			}
			return fmt.Sprintf("⚡ [%s] dinamik komutu çalıştırıldı. Argümanlar: %v", cmd, args)
		})
		return fmt.Sprintf("⚡ [SUPER USER x1000] Komut çekirdeğe enjekte edildi: /%s", cmd)

	case "unregister_cmd", "remove_command":
		cmd := getString("command")
		if cmd != "" {
			host.UnregisterDynamicCommand(cmd)
			return fmt.Sprintf("⚡ [SUPER USER x1000] Komut çekirdekten kaldırıldı: /%s", cmd)
		}

	case "trigger_cmd", "run_cmd", "exec_cmd":
		cmdLine := getString("command")
		if cmdLine == "" {
			cmdLine = getString("cmd")
		}
		if cmdLine != "" {
			host.TriggerCommand(cmdLine)
			return fmt.Sprintf("⚡ [SUPER USER x1000] Komut tetiklendi: %s", cmdLine)
		}

	// =========================================================================
	// SUPER USER x1000: LIVE UI & MODEL MUTATION
	// =========================================================================
	case "set_banner":
		banner := getString("banner")
		if banner != "" {
			host.SetBanner(banner)
			return "⚡ [SUPER USER x1000] Üst ASCII banner değiştirildi."
		}

	case "reset_banner":
		host.ResetBanner()
		return "✔ Banner varsayılana döndürüldü."

	case "reset_gui", "reset_all":
		host.CloseAllTabs()
		host.ClearPixels()
		host.ResetBanner()
		host.ResetFooter()
		host.ResetPrompt()
		host.ClearFullscreenCustomView()
		return "✔ [GOD MODE] GUI ve arayüz tamamen sıfırlandı."

	case "set_status":
		msg := getString("message")
		isErr := false
		if b, ok := payload["is_error"].(bool); ok {
			isErr = b
		}
		host.SetStatus(msg, isErr)
		return fmt.Sprintf("⚡ [SUPER USER x1000] Durum metni güncellendi: %s", msg)

	case "set_footer":
		deck := getString("footer")
		if deck != "" {
			host.SetFooter(deck)
			return "⚡ [SUPER USER x1000] Altbilgi kısayolları değiştirildi."
		}

	case "reset_footer":
		host.ResetFooter()
		return "✔ Altbilgi varsayılana döndürüldü."

	case "set_input_value":
		val := getString("value")
		host.SetInputValue(val)
		return fmt.Sprintf("⚡ [SUPER USER x1000] İmleç girdisi ayarlandı: %s", val)

	case "set_input_placeholder":
		ph := getString("placeholder")
		host.SetInputPlaceholder(ph)
		return fmt.Sprintf("⚡ [SUPER USER x1000] Girdi placeholder'ı ayarlandı: %s", ph)

	case "hijack_screen":
		content := getString("content")
		host.SetFullscreenCustomView(func() string {
			return content
		})
		return "⚡ [SUPER USER x1000] Ekran tamamen eklenti tarafından devralındı."

	case "release_screen":
		host.ClearFullscreenCustomView()
		return "✔ Ekran devri sonlandırıldı."

	case "simulate_key":
		k := getString("key")
		if k != "" {
			host.SimulateKey(k)
			return fmt.Sprintf("⚡ [SUPER USER x1000] Tuş simüle edildi: [%s]", k)
		}

	case "inspect":
		info := host.InspectKernel()
		if bytes, err := json.MarshalIndent(info, "", "  "); err == nil {
			return string(bytes)
		}

	// =========================================================================
	// SUPER USER x1000: TAB & WORKSPACE CONTROL
	// =========================================================================
	case "switch_tab":
		idx := getInt("index", 0)
		if host.SwitchTab(idx) {
			return fmt.Sprintf("⚡ [SUPER USER x1000] Sekme #%d aktif yapıldı.", idx+1)
		}
		return fmt.Sprintf("⚠️ Sekme #%d bulunamadı.", idx+1)

	case "close_all_tabs":
		host.CloseAllTabs()
		return "⚡ [SUPER USER x1000] Tüm çalışma sekmeleri kapatıldı."

	case "set_tab_title":
		idx := getInt("index", 0)
		title := getString("title")
		if host.SetTabTitle(idx, title) {
			return fmt.Sprintf("⚡ [SUPER USER x1000] Sekme #%d başlığı güncellendi: %s", idx+1, title)
		}
		return fmt.Sprintf("⚠️ Sekme #%d başlığı güncellenemedi.", idx+1)

	case "scroll_tab":
		idx := getInt("index", -1)
		delta := getInt("delta", 3)
		host.ScrollTab(idx, delta)
		return fmt.Sprintf("⚡ [SUPER USER x1000] Sekme kaydırıldı (delta: %d)", delta)

	case "set_theme_color":
		field := getString("field")
		col := getString("color")
		if field != "" && col != "" {
			if host.SetThemeColor(field, col) {
				return fmt.Sprintf("⚡ [SUPER USER x1000] Tema rengi güncellendi: %s=%s", field, col)
			}
			return fmt.Sprintf("⚠️ Geçersiz tema alanı: '%s'", field)
		}
		return "⚠️ set_theme_color için 'field' ve 'color' gereklidir."

	// =========================================================================
	// STANDARD / INTERMEDIATE IPC ACTIONS
	// =========================================================================
	case "launch_tui", "handover_tui", "external_tui":
		bin := getString("command")
		if bin == "" {
			bin = getString("binary")
		}
		name := getString("name")
		prompt := getString("prompt")
		var args []string
		if rawArgs, ok := payload["args"].([]any); ok {
			for _, a := range rawArgs {
				args = append(args, fmt.Sprintf("%v", a))
			}
		}
		if bin != "" {
			host.RequestExternalTUI(bin, args, name, prompt)
			return fmt.Sprintf("⏳ Harici TUI devir isteği onay bekliyor: [%s]", bin)
		}

	case "pixel", "set_pixel":
		x := getInt("x", 0)
		y := getInt("y", 0)
		chStr := getString("char")
		r := '█'
		if len(chStr) > 0 {
			r = []rune(chStr)[0]
		}
		fgHex := getString("fg")
		bgHex := getString("bg")
		var fg, bg lipgloss.TerminalColor
		if fgHex != "" {
			fg = lipgloss.Color(fgHex)
		}
		if bgHex != "" {
			bg = lipgloss.Color(bgHex)
		}
		host.MutatePixel(x, y, r, fg, bg, true)
		return fmt.Sprintf("⚡ [ROOT] Tekil piksel manipüle edildi: (%d, %d)", x, y)

	case "pixels", "draw_batch":
		rawList, ok := payload["pixels"].([]any)
		if ok {
			var batch []CellMutation
			for _, item := range rawList {
				pMap, isMap := item.(map[string]any)
				if !isMap {
					continue
				}
				px := 0
				py := 0
				if val, ok := pMap["x"]; ok && val != nil {
					switch v := val.(type) {
					case float64:
						px = int(v)
					case int:
						px = v
					case string:
						px, _ = strconv.Atoi(v)
					}
				}
				if val, ok := pMap["y"]; ok && val != nil {
					switch v := val.(type) {
					case float64:
						py = int(v)
					case int:
						py = v
					case string:
						py, _ = strconv.Atoi(v)
					}
				}
				pr := '█'
				if cVal, ok := pMap["char"].(string); ok && len(cVal) > 0 {
					pr = []rune(cVal)[0]
				}
				var pFg, pBg lipgloss.TerminalColor
				if fVal, ok := pMap["fg"].(string); ok && fVal != "" {
					pFg = lipgloss.Color(fVal)
				}
				if bVal, ok := pMap["bg"].(string); ok && bVal != "" {
					pBg = lipgloss.Color(bVal)
				}
				batch = append(batch, CellMutation{
					X:    px,
					Y:    py,
					Char: pr,
					Fg:   pFg,
					Bg:   pBg,
					Bold: true,
				})
			}
			host.MutatePixelBatch(batch)
			return fmt.Sprintf("⚡ [ROOT] Toplu piksel matrisi işlendi: %d hücre", len(batch))
		}

	case "box", "rect":
		bx := getInt("x", 0)
		by := getInt("y", 0)
		bw := getInt("w", 20)
		bh := getInt("h", 6)
		if bw < 2 {
			bw = 2
		}
		if bh < 2 {
			bh = 2
		}
		title := getString("title")
		fgHex := getString("fg")
		var fg lipgloss.TerminalColor
		if fgHex != "" {
			fg = lipgloss.Color(fgHex)
		} else {
			fg = host.CurrentTheme().Primary
		}
		host.DrawBox(bx, by, bw, bh, title, fg, true)
		return fmt.Sprintf("⚡ [ROOT] Geometrik kutu çizildi: (%d, %d) [%dx%d]", bx, by, bw, bh)

	case "clear_pixels", "wipe_screen":
		host.ClearPixels()
		return "⚡ [ROOT] Sanal piksel katmanı temizlendi."

	case "set_prompt":
		pfx := getString("prefix")
		arr := getString("arrow")
		fgP := lipgloss.Color(getString("fg_prefix"))
		fgA := lipgloss.Color(getString("fg_arrow"))
		host.SetPrompt(pfx, arr, fgP, fgA)
		return fmt.Sprintf("⚡ [ROOT] Kabuk prompt'u canlıda güncellendi: %s%s", pfx, arr)

	case "reset_prompt":
		host.ResetPrompt()
		return "✔ Prompt varsayılana döndü."

	case "bind_key":
		k := getString("key")
		cmd := getString("cmd")
		if k != "" && cmd != "" {
			host.BindKey(k, cmd)
			return fmt.Sprintf("⚡ [ROOT] Tuş kancalandı: [%s] -> %s", strings.ToUpper(k), cmd)
		}

	case "unbind_key":
		k := getString("key")
		host.UnbindKey(k)
		return fmt.Sprintf("✔ Tuş bağı kaldırıldı: [%s]", strings.ToUpper(k))

	case "patch_cmd":
		target := getString("target")
		rewrite := getString("rewrite")
		if target != "" && rewrite != "" {
			host.PatchCommand(target, rewrite)
			return fmt.Sprintf("⚡ [ROOT] Komut yamalandı: /%s -> /%s", target, rewrite)
		}

	case "unpatch_cmd":
		target := getString("target")
		host.UnpatchCommand(target)
		return fmt.Sprintf("✔ Komut yaması kaldırıldı: /%s", target)

	case "overlay", "matrix":
		id := getString("id")
		if id == "" {
			id = "plugin_overlay"
		}
		text := getString("text")
		posX := getInt("x", 0)
		posY := getInt("y", 0)
		fgHex := getString("fg")
		var fg lipgloss.TerminalColor
		if fgHex != "" {
			fg = lipgloss.Color(fgHex)
		}
		host.RegisterOverlay(id, func(fb *framebuffer.VirtualFramebuffer, width, height int) {
			fb.DrawText(posX, posY, text, fg, nil, true)
		})
		return fmt.Sprintf("⚡ [ROOT] Sanal kanca overlay enjekte edildi: [%s]", id)

	case "remove_overlay":
		id := getString("id")
		host.RemoveOverlay(id)
		return fmt.Sprintf("✔ Overlay kaldırıldı: [%s]", id)

	case "hook":
		target := getString("target")
		if target == "" {
			target = "banner"
		}
		content := getString("content")
		if content == "" {
			host.ClearTextHook(target)
			return fmt.Sprintf("✔ Kanca sıfırlandı: [%s]", target)
		}
		host.SetTextHook(target, content)
		return fmt.Sprintf("✔ Metin kancası güncellendi: [%s]", target)

	case "toast":
		msg := getString("message")
		isErr := false
		if b, ok := payload["is_error"].(bool); ok {
			isErr = b
		}
		host.Toast(msg, isErr)
		return fmt.Sprintf("✔ Bildirim gönderildi: %s", msg)

	case "theme_custom":
		palette := theme.ThemePalette{
			Name:        getString("name"),
			Primary:     lipgloss.Color(getString("primary")),
			Secondary:   lipgloss.Color(getString("secondary")),
			BgDark:      lipgloss.Color(getString("bg_dark")),
			BorderColor: lipgloss.Color(getString("border_color")),
			TextMain:    lipgloss.Color(getString("text_main")),
			TextMuted:   lipgloss.Color(getString("text_muted")),
			BadgeBg:     lipgloss.Color(getString("badge_bg")),
		}
		if palette.Name == "" {
			palette.Name = "custom_root"
		}
		host.SetCustomTheme(palette)
		return fmt.Sprintf("⚡ [ROOT] Özel tema enjekte edildi: %s", palette.Name)

	case "kv_set":
		k := getString("key")
		v := getString("value")
		if k != "" {
			host.KV().Set(k, v)
			return fmt.Sprintf("✔ Shared State kaydedildi: %s=%s", k, v)
		}

	case "force_close_tab":
		idx := getInt("index", -1)
		if idx == -1 {
			idx = host.ActiveTabIndex()
		}
		if idx >= 0 && host.ForceCloseTab(idx) {
			return fmt.Sprintf("⚡ Sekme #%d çekirdek tarafından kapatıldı.", idx+1)
		}
		return fmt.Sprintf("⚠️ Kapatılacak sekme (#%d) bulunamadı.", idx+1)

	case "inject_tab":
		title := getString("title")
		cmd := getString("cmd")
		out := getString("output")
		host.InjectTab(title, cmd, out)
		return fmt.Sprintf("⚡ Doğrudan sekme oluşturuldu: [%s]", title)

	case "update_tab":
		idx := getInt("index", -1)
		if idx == -1 {
			idx = host.ActiveTabIndex()
		}
		out := getString("output")
		if idx >= 0 && host.UpdateTabOutput(idx, out) {
			return fmt.Sprintf("⚡ Sekme #%d güncellendi.", idx+1)
		}
		return fmt.Sprintf("⚠️ Güncellenecek sekme (#%d) bulunamadı.", idx+1)

	case "theme":
		name := getString("name")
		if host.SetTheme(name) {
			return fmt.Sprintf("✔ Tema güncellendi: %s", name)
		}
		return fmt.Sprintf("⚠️ Geçersiz tema adı: '%s'. Seçenekler: amber, emerald, cyber, mono", name)

	// =========================================================================
	// ADVANCED ENGINE: STREAMING & EVENT BUS
	// =========================================================================
	case "stream_append", "stream":
		chunk := getString("chunk")
		if chunk == "" {
			chunk = getString("text")
		}
		if chunk == "" {
			chunk = getString("content")
		}
		host.StreamAppend(chunk)
		return ""

	case "emit_event", "event":
		evt := getString("event")
		if evt == "" {
			evt = getString("name")
		}
		payloadData := payload["data"]
		if payloadData == nil {
			payloadData = payload["payload"]
		}
		host.EmitEvent(evt, payloadData)
		return fmt.Sprintf("⚡ [EVENT BUS] Olay yayınlandı: '%s'", evt)

	// =========================================================================
	// ADVANCED ENGINE: RICH GRAPHICS & VISUAL WIDGETS
	// =========================================================================
	case "table", "render_table":
		var headers []string
		if hRaw, ok := payload["headers"].([]any); ok {
			for _, h := range hRaw {
				headers = append(headers, fmt.Sprintf("%v", h))
			}
		}
		var rows [][]string
		if rRaw, ok := payload["rows"].([]any); ok {
			for _, rowItem := range rRaw {
				if rList, ok := rowItem.([]any); ok {
					var row []string
					for _, c := range rList {
						row = append(row, fmt.Sprintf("%v", c))
					}
					rows = append(rows, row)
				}
			}
		}
		return host.RenderTable(headers, rows)

	case "sparkline", "render_sparkline":
		var vals []float64
		if vRaw, ok := payload["values"].([]any); ok {
			for _, v := range vRaw {
				switch num := v.(type) {
				case float64:
					vals = append(vals, num)
				case int:
					vals = append(vals, float64(num))
				case string:
					if f, err := strconv.ParseFloat(num, 64); err == nil {
						vals = append(vals, f)
					}
				}
			}
		}
		title := getString("title")
		color := getString("color")
		return host.RenderSparkline(vals, title, color)

	case "barchart", "bar_chart", "render_barchart":
		data := make(map[string]float64)
		if dRaw, ok := payload["data"].(map[string]any); ok {
			for k, v := range dRaw {
				switch num := v.(type) {
				case float64:
					data[k] = num
				case int:
					data[k] = float64(num)
				case string:
					if f, err := strconv.ParseFloat(num, 64); err == nil {
						data[k] = f
					}
				}
			}
		}
		title := getString("title")
		color := getString("color")
		return host.RenderBarChart(data, title, color)

	case "kv_del", "kv_delete":
		k := getString("key")
		if k != "" {
			host.KV().Delete(k)
			return fmt.Sprintf("✔ Shared State silindi: %s", k)
		}

	case "clear_tab":
		idx := getInt("index", -1)
		if idx == -1 {
			idx = host.ActiveTabIndex()
		}
		if idx >= 0 && host.UpdateTabOutput(idx, "") {
			return "✔ Sekme içeriği temizlendi."
		}

	case "modal":
		host.OpenModal(getString("title"), getString("content"))
		return "✔ Modal açıldı."
	}

	return ""
}
