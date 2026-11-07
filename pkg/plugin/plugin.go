package plugin

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/superuser"
	"github.com/rkqze10/mt-tui/pkg/theme"
)

// PluginContext provides execution details to the plugin view and update methods.
type PluginContext struct {
	Width      int
	Height     int
	InvokedCmd string
	Args       []string
	RawArgs    string
	PipeInput  string
	Theme      theme.ThemePalette
	Host       superuser.SuperuserHost
}

// LowLevelPlugin defines the contract for both internal built-ins and external modules.
type LowLevelPlugin interface {
	Manifests() []Manifest
	Init(host superuser.SuperuserHost) error
	StartDaemon(ctx context.Context, host superuser.SuperuserHost)
	Intercept(rawCmd string) (proceed bool, modifiedCmd string)
	Update(msg tea.Msg, ctx *PluginContext) (consumed bool, cmd tea.Cmd)
	View(ctx *PluginContext) string
	Destroy() error
}
