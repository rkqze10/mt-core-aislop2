package builtins

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// TuiHandoverPlugin delegates terminal control to an external interactive TUI program.
type TuiHandoverPlugin struct{}

func (t *TuiHandoverPlugin) Manifests() []plugin.Manifest {
	return []plugin.Manifest{
		{
			Command:      "tui",
			Title:        "HARİCİ TUI ÇALIŞTIRICI",
			Category:     "RING-0",
			Description:  "Harici interaktif TUI programını onay sorarak güvenle başlatır (/tui htop)",
			Capabilities: plugin.CapUI | plugin.CapSuperuser,
		},
	}
}

func (t *TuiHandoverPlugin) Init(h superuser.SuperuserHost) error                        { return nil }
func (t *TuiHandoverPlugin) Destroy() error                                             { return nil }
func (t *TuiHandoverPlugin) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (t *TuiHandoverPlugin) Intercept(c string) (bool, string)                           { return true, c }
func (t *TuiHandoverPlugin) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}

func (t *TuiHandoverPlugin) View(ctx *plugin.PluginContext) string {
	if len(ctx.Args) == 0 {
		return "Kullanım: /tui <program_veya_eklenti> [parametreler]\nÖrnek:\n  /tui htop\n  /tui nano dosya.txt\n  /tui python3 my_curses_app.py"
	}
	binary := ctx.Args[0]
	args := ctx.Args[1:]
	ctx.Host.RequestExternalTUI(binary, args, binary, "")
	return fmt.Sprintf("⏳ Harici TUI başlatma isteği gönderildi: [%s] (Onay bekleniyor)", binary)
}
