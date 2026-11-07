package builtins

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// SimpleCmd provides a minimal wrapper for simple function-based commands.
type SimpleCmd struct {
	manifest plugin.Manifest
	viewFn   func() string
}

// NewSimpleCmd creates a simple command from a manifest and view callback.
func NewSimpleCmd(manifest plugin.Manifest, viewFn func() string) *SimpleCmd {
	return &SimpleCmd{
		manifest: manifest,
		viewFn:   viewFn,
	}
}

func (sc *SimpleCmd) Manifests() []plugin.Manifest                                  { return []plugin.Manifest{sc.manifest} }
func (sc *SimpleCmd) Init(h superuser.SuperuserHost) error                        { return nil }
func (sc *SimpleCmd) Destroy() error                                             { return nil }
func (sc *SimpleCmd) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (sc *SimpleCmd) Intercept(c string) (bool, string)                           { return true, c }
func (sc *SimpleCmd) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}
func (sc *SimpleCmd) View(ctx *plugin.PluginContext) string {
	if sc.viewFn == nil {
		return ""
	}
	return sc.viewFn()
}

// DynamicCmd supports runtime injected commands that consume CLI arguments and pipeline data.
type DynamicCmd struct {
	manifest plugin.Manifest
	handler  func(args []string, pipeIn string) string
}

// NewDynamicCmd constructs a command that passes arguments and pipe input to the handler.
func NewDynamicCmd(manifest plugin.Manifest, handler func(args []string, pipeIn string) string) *DynamicCmd {
	return &DynamicCmd{
		manifest: manifest,
		handler:  handler,
	}
}

func (dc *DynamicCmd) Manifests() []plugin.Manifest                                  { return []plugin.Manifest{dc.manifest} }
func (dc *DynamicCmd) Init(h superuser.SuperuserHost) error                        { return nil }
func (dc *DynamicCmd) Destroy() error                                             { return nil }
func (dc *DynamicCmd) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (dc *DynamicCmd) Intercept(c string) (bool, string)                           { return true, c }
func (dc *DynamicCmd) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}
func (dc *DynamicCmd) View(ctx *plugin.PluginContext) string {
	if dc.handler == nil {
		return ""
	}
	args := []string{}
	pipeIn := ""
	if ctx != nil {
		args = ctx.Args
		pipeIn = ctx.PipeInput
	}
	return dc.handler(args, pipeIn)
}
