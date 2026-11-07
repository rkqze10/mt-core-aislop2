package builtins

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// ThemeSwitchPlugin allows hot-swapping color palettes dynamically.
type ThemeSwitchPlugin struct{}

func (tp *ThemeSwitchPlugin) Manifests() []plugin.Manifest {
	return []plugin.Manifest{
		{
			Command:      "theme",
			Title:        "DİNAMİK RENK TEMASI",
			Category:     "GUI",
			Description:  "Temayı değiştirir (/theme cyber|emerald|amber|mono)",
			Capabilities: plugin.CapUI,
		},
	}
}

func (tp *ThemeSwitchPlugin) Init(h superuser.SuperuserHost) error                        { return nil }
func (tp *ThemeSwitchPlugin) Destroy() error                                             { return nil }
func (tp *ThemeSwitchPlugin) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (tp *ThemeSwitchPlugin) Intercept(c string) (bool, string)                           { return true, c }
func (tp *ThemeSwitchPlugin) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}

func (tp *ThemeSwitchPlugin) View(ctx *plugin.PluginContext) string {
	if len(ctx.Args) == 0 {
		return "Kullanım: /theme <amber | emerald | cyber | mono>\nÖrnek: /theme cyber"
	}
	target := strings.ToLower(ctx.Args[0])
	if ctx.Host.SetTheme(target) {
		ctx.Host.Toast(fmt.Sprintf("Tema '%s' yapıldı ve diske kaydedildi!", target), false)
		return fmt.Sprintf("Tema başarıyla '%s' olarak güncellendi.", target)
	}
	return fmt.Sprintf("Geçersiz tema: '%s'. Seçenekler: amber, emerald, cyber, mono", target)
}
