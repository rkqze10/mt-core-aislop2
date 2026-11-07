package builtins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rkqze10/mt-tui/pkg/config"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// PluginInspector allows the plugin manager to inspect and reload kernel modules.
type PluginInspector interface {
	CheckAndReload(dir string) (bool, int)
	Inspect() map[string]any
	Search(q string) []plugin.Manifest
}

// PluginManagerCommand provides the /plugin command for managing and scaffolding plugins.
type PluginManagerCommand struct {
	inspector PluginInspector
}

// NewPluginManagerCommand creates a new plugin manager command.
func NewPluginManagerCommand(insp PluginInspector) *PluginManagerCommand {
	return &PluginManagerCommand{inspector: insp}
}

func (pm *PluginManagerCommand) Manifests() []plugin.Manifest {
	return []plugin.Manifest{
		{
			Command:      "plugin",
			Title:        "EKLENTİ YÖNETİCİSİ",
			Category:     "MGMT",
			Description:  "Eklentileri listeler, yeniden yükler veya yeni eklenti oluşturur (/plugin list|reload|new)",
			Capabilities: plugin.CapUI | plugin.CapSuperuser,
		},
		{
			Command:      "plugins",
			Title:        "EKLENTİ YÖNETİCİSİ",
			Category:     "MGMT",
			Description:  "Eklentileri listeler (/plugins list)",
			Capabilities: plugin.CapUI | plugin.CapSuperuser,
		},
	}
}

func (pm *PluginManagerCommand) Init(h superuser.SuperuserHost) error                        { return nil }
func (pm *PluginManagerCommand) Destroy() error                                             { return nil }
func (pm *PluginManagerCommand) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (pm *PluginManagerCommand) Intercept(c string) (bool, string)                           { return true, c }
func (pm *PluginManagerCommand) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}

func (pm *PluginManagerCommand) View(ctx *plugin.PluginContext) string {
	theme := ctx.Theme
	subcmd := "list"
	if len(ctx.Args) > 0 {
		subcmd = strings.ToLower(ctx.Args[0])
	}

	switch subcmd {
	case "reload":
		changed, count := pm.inspector.CheckAndReload(config.GetPluginsDir())
		ctx.Host.Toast(fmt.Sprintf("🔄 Eklentiler tarandı (%d aktif)", count), false)
		if changed {
			return fmt.Sprintf("✔ Eklentiler başarıyla güncellendi! Toplam: %d eklenti devrede.", count)
		}
		return fmt.Sprintf("✔ Tüm eklentiler güncel. Toplam: %d eklenti devrede.", count)

	case "new", "create":
		if len(ctx.Args) < 2 {
			return "Kullanım: /plugin new <isim> [python|shell|go]\nÖrnek:\n  /plugin new crypto python\n  /plugin new monitor shell"
		}
		name := strings.ToLower(strings.TrimSpace(ctx.Args[1]))
		// Validate plugin name to prevent path traversal
		for _, r := range name {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {
				return "⚠️ Hata: Eklenti ismi sadece küçük harf, rakam ve alt çizgi içerebilir!"
			}
		}
		lang := "python"
		if len(ctx.Args) >= 3 {
			lang = strings.ToLower(ctx.Args[2])
		}

		pluginsDir := config.GetPluginsDir()
		_ = os.MkdirAll(pluginsDir, 0755)

		var filename, template string
		switch lang {
		case "shell", "sh", "bash":
			filename = filepath.Join(pluginsDir, name+".sh")
			template = fmt.Sprintf(`#!/usr/bin/env bash
# @command: %s
# @title: %s
# @category: TOOLS
# @desc: %s shell eklentisi
# @author: root
# @version: 1.0.0
# @icon: ⚡

source "$(dirname "$0")/sdk/mt.sh" 2>/dev/null

mt_toast "⚡ %s eklentisi çalıştı!" false

echo "=================================================="
echo "  %s Eklentisi Başarıyla Çalıştı! "
echo "=================================================="
echo "Girdi Argümanları: $@"
`, name, strings.ToUpper(name), name, name, strings.ToUpper(name))

		case "go", "golang":
			filename = filepath.Join(pluginsDir, name+".go")
			template = fmt.Sprintf(`package main

// @command: %s
// @title: %s
// @category: TOOLS
// @desc: %s Go süper kullanıcı eklentisi
// @author: root
// @version: 1.0.0
// @icon: 🚀

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("==================================================")
	fmt.Printf("  %%s Go Eklentisi Çalıştı!\n", strings.ToUpper("%s"))
	fmt.Println("==================================================")
	fmt.Printf("Parametreler: %%v\n", os.Args[1:])
}
`, name, strings.ToUpper(name), name, name)

		default: // python
			filename = filepath.Join(pluginsDir, name+".py")
			template = fmt.Sprintf(`#!/usr/bin/env python3
# @command: %s
# @title: %s
# @category: TOOLS
# @desc: %s Python süper kullanıcı eklentisi
# @author: root
# @version: 1.0.0
# @icon: 🐍

import sys
import os

try:
    from sdk.mt import MT, Style, Color
    mt = MT()
    mt.toast("⚡ %s eklentisi devrede!", is_error=False)
except ImportError:
    mt = None

print("=" * 52)
print("  %s EKLENTİSİ BAŞARIYLA ÇALIŞTI")
print("=" * 52)
print(f"Argümanlar: {sys.argv[1:]}")
`, name, strings.ToUpper(name), name, name, strings.ToUpper(name))
		}

		if _, err := os.Stat(filename); err == nil {
			return fmt.Sprintf("⚠️ Hata: '%s' dosyası zaten mevcut!", filepath.Base(filename))
		}

		if err := os.WriteFile(filename, []byte(template), 0755); err != nil {
			return fmt.Sprintf("⚠️ Dosya oluşturulamadı: %v", err)
		}

		pm.inspector.CheckAndReload(pluginsDir)
		ctx.Host.Toast(fmt.Sprintf("✔ Yeni eklenti oluşturuldu: /%s", name), false)

		return fmt.Sprintf("✔ Başarılı! '%s' eklenti şablonu oluşturuldu:\nDosya: %s\n\nArtık /%s yazarak çalıştırabilirsiniz!", name, filename, name)

	case "info":
		if len(ctx.Args) < 2 {
			return "Kullanım: /plugin info <komut_adı>\nÖrnek: /plugin info hello"
		}
		target := strings.ToLower(strings.TrimPrefix(ctx.Args[1], "/"))
		matches := pm.inspector.Search(target)
		for _, m := range matches {
			if strings.ToLower(m.Command) == target {
				var b strings.Builder
				header := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true).Render(fmt.Sprintf("● EKLENTİ DETAYI: /%s", m.Command))
				b.WriteString(header + "\n\n")

				addProp := func(key, val string) {
					k := lipgloss.NewStyle().Width(14).Foreground(theme.TextMuted).Render(key + ":")
					v := lipgloss.NewStyle().Foreground(theme.TextMain).Bold(true).Render(val)
					b.WriteString(k + v + "\n")
				}

				addProp("Başlık", m.Title)
				addProp("Kategori", m.Category)
				addProp("Açıklama", m.Description)
				if m.Author != "" {
					addProp("Yazar", m.Author)
				}
				if m.Version != "" {
					addProp("Versiyon", m.Version)
				}
				if m.Icon != "" {
					addProp("İkon", m.Icon)
				}
				if len(m.Aliases) > 0 {
					addProp("Aliaslar", strings.Join(m.Aliases, ", "))
				}
				if len(m.Requires) > 0 {
					addProp("Gereksinimler", strings.Join(m.Requires, ", "))
				}
				if m.Timeout > 0 {
					addProp("Zaman Aşımı", m.Timeout.String())
				}
				return b.String()
			}
		}
		return fmt.Sprintf("'%s' komutuyla eşleşen bir eklenti bulunamadı.", target)

	default: // list
		manifests := pm.inspector.Search("")
		var b strings.Builder

		titleSt := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true).Render("YÜKLÜ SÜPER KULLANICI EKLENTİLERİ & MODÜLLER:")
		b.WriteString(titleSt + "\n\n")

		headCmd := lipgloss.NewStyle().Width(14).Foreground(theme.TextMuted).Render("KOMUT")
		headCat := lipgloss.NewStyle().Width(10).Foreground(theme.TextMuted).Render("KATEGORİ")
		headDesc := lipgloss.NewStyle().Width(32).Foreground(theme.TextMuted).Render("AÇIKLAMA")
		headVer := lipgloss.NewStyle().Width(8).Foreground(theme.TextMuted).Render("VERSİYON")
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, headCmd, headCat, headDesc, headVer) + "\n")
		b.WriteString(strings.Repeat("─", 68) + "\n")

		for _, m := range manifests {
			icon := "•"
			if m.Icon != "" {
				icon = m.Icon
			}
			cmdCell := lipgloss.NewStyle().Width(14).Foreground(theme.Secondary).Bold(true).Render(fmt.Sprintf("%s /%s", icon, m.Command))
			catCell := lipgloss.NewStyle().Width(10).Foreground(theme.TextMuted).Render(fmt.Sprintf("[%s]", m.Category))

			desc := m.Description
			descRunes := []rune(desc)
			if len(descRunes) > 30 {
				desc = string(descRunes[:27]) + "..."
			}
			descCell := lipgloss.NewStyle().Width(32).Foreground(theme.TextMain).Render(desc)

			ver := "v1.0"
			if m.Version != "" {
				ver = m.Version
			}
			verCell := lipgloss.NewStyle().Width(8).Foreground(theme.TextMuted).Render(ver)

			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cmdCell, catCell, descCell, verCell) + "\n")
		}

		b.WriteString("\n" + lipgloss.NewStyle().Foreground(theme.TextMuted).Render(
			fmt.Sprintf("Eklenti Dizini: %s\n", config.GetPluginsDir())+
				"Alt Komutlar:\n"+
				" • /plugin list              : Yüklü eklentileri gösterir.\n"+
				" • /plugin reload            : Disk üzerindeki eklentileri tarar ve yeniler.\n"+
				" • /plugin info <komut>      : Eklenti detaylarını görüntüler.\n"+
				" • /plugin new <ad> <dil>    : Yeni eklenti şablonu oluşturur (python, shell, go).",
		))

		return b.String()
	}
}
