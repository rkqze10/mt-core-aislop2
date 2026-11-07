package builtins

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// SysDaemonPlugin provides runtime metrics, memory allocation, and CPU statistics.
type SysDaemonPlugin struct{}

func (s *SysDaemonPlugin) Manifests() []plugin.Manifest {
	return []plugin.Manifest{
		{
			Command:      "sys",
			Title:        "SİSTEM TELEMETRİSİ",
			Category:     "SYS",
			Description:  "Canlı donanım durumu ve bellek ⚡",
			Capabilities: plugin.CapUI,
		},
	}
}

func (s *SysDaemonPlugin) Init(h superuser.SuperuserHost) error                        { return nil }
func (s *SysDaemonPlugin) Destroy() error                                             { return nil }
func (s *SysDaemonPlugin) Intercept(c string) (bool, string)                           { return true, c }
func (s *SysDaemonPlugin) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (s *SysDaemonPlugin) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}

func (s *SysDaemonPlugin) View(ctx *plugin.PluginContext) string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if len(ctx.Args) > 0 {
		switch strings.ToLower(ctx.Args[0]) {
		case "cpu":
			return fmt.Sprintf("İşlemci: %d Threads | Mimari: %s/%s\nDurum: Nominal çalışma frekansı.", runtime.NumCPU(), runtime.GOOS, runtime.GOARCH)
		case "mem":
			return fmt.Sprintf("Tahsis Bellek: %.2f MB\nSistem Rezerv: %.2f MB\nGC Döngüsü   : %d kez tamamlandı", float64(m.Alloc)/1024/1024, float64(m.Sys)/1024/1024, m.NumGC)
		}
	}

	return fmt.Sprintf(
		"İşletim Sistemi  : %s / %s\n"+
			"CPU Thread       : %d Threads 🚀\n"+
			"Aktif Goroutines : %d Adet\n"+
			"Tahsis Bellek    : %.2f MB (Sistem: %.2f MB)\n"+
			"GC Frekansı      : %d döngü tamamlandı\n\n"+
			"● İpucu: /sys cpu veya /sys mem yazabilirsiniz.",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(),
		runtime.NumGoroutine(), float64(m.Alloc)/1024/1024, float64(m.Sys)/1024/1024, m.NumGC,
	)
}
