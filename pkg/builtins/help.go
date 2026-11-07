package builtins

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rkqze10/mt-tui/pkg/config"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// Searcher represents an interface to search available manifests.
type Searcher interface {
	Search(query string) []plugin.Manifest
}

// HelpCommand lists all registered commands and system shortcuts.
type HelpCommand struct {
	searcher Searcher
}

// NewHelpCommand creates a new HelpCommand plugin.
func NewHelpCommand(s Searcher) *HelpCommand {
	return &HelpCommand{searcher: s}
}

func (hc *HelpCommand) Manifests() []plugin.Manifest {
	return []plugin.Manifest{
		{
			Command:      "help",
			Title:        "SİSTEM KILAVUZU & KOMUT REHBERİ",
			Category:     "CORE",
			Description:  "Tüm sistem komutlarını ve modülleri listeler.",
			Capabilities: plugin.CapUI,
		},
	}
}

func (hc *HelpCommand) Init(h superuser.SuperuserHost) error                        { return nil }
func (hc *HelpCommand) Destroy() error                                             { return nil }
func (hc *HelpCommand) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (hc *HelpCommand) Intercept(c string) (bool, string)                           { return true, c }
func (hc *HelpCommand) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}

func (hc *HelpCommand) View(ctx *plugin.PluginContext) string {
	theme := ctx.Theme
	query := ""
	if ctx != nil && len(ctx.Args) > 0 {
		query = strings.Join(ctx.Args, " ")
	}

	var b strings.Builder
	title := "GELİŞMİŞ SİSTEM KOMUTLARI (ROOT / RING -1 AKTİF):"
	if query != "" {
		title = fmt.Sprintf("ARAMA SONUÇLARI ('%s'):", query)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(theme.Primary).Bold(true).Render(title + "\n\n"))

	matches := hc.searcher.Search(query)
	if len(matches) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.TextMuted).Render(fmt.Sprintf("'%s' ile eşleşen bir komut bulunamadı.\n", query)))
	} else {
		for _, item := range matches {
			catSpan := lipgloss.NewStyle().
				Width(12).
				Foreground(theme.TextMuted).
				Background(theme.BadgeBg).
				Padding(0, 1).
				Render(item.Category)

			nameSpan := lipgloss.NewStyle().
				Width(20).
				Foreground(theme.Secondary).
				Bold(true).
				Render("/" + item.Command)

			descSpan := lipgloss.NewStyle().
				Foreground(theme.TextMuted).
				Render(item.Description)

			row := lipgloss.JoinHorizontal(lipgloss.Left, catSpan, "  ", nameSpan, "  ", descSpan)
			b.WriteString(row + "\n")
		}
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(theme.TextMuted).Render(
		fmt.Sprintf("Eklenti Dizini: %s\n", config.GetPluginsDir())+
			"Klavye & Fare Yetenekleri:\n"+
			" • Tab               : Otomatik komut tamamla.\n"+
			" • Alt+1 .. Alt+9    : Açık sekmeler arasında geçiş yapar.\n"+
			" • Alt+← / Alt+→     : Sekmeler arasında ileri/geri döner.\n"+
			" • Ctrl+W            : Aktif sekmeyi kapatır (boş satırda).\n"+
			" • Ctrl+Y            : Aktif sekme çıktısını panoya kopyalar.\n"+
			" • PgUp / PgDn / Çark: Çıktı alanını yukarı/aşağı kaydırır.\n"+
			" • /tui <cmd>        : Onaylı Harici TUI başlatıcı (Örn: /tui htop).",
	))
	return b.String()
}
