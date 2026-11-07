package kernel

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

type mockHost struct {
	superuser.SuperuserHost
}

type mockPlugin struct {
	manifests []plugin.Manifest
}

func (m *mockPlugin) Manifests() []plugin.Manifest { return m.manifests }
func (m *mockPlugin) Init(h superuser.SuperuserHost) error { return nil }
func (m *mockPlugin) Destroy() error { return nil }
func (m *mockPlugin) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (m *mockPlugin) Intercept(c string) (bool, string) { return true, c }
func (m *mockPlugin) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) { return false, nil }
func (m *mockPlugin) View(ctx *plugin.PluginContext) string { return "mock" }

func TestKernelEngine_AliasResolution(t *testing.T) {
	eng := New(&mockHost{})
	plug := &mockPlugin{
		manifests: []plugin.Manifest{
			{
				Command: "status",
				Aliases: []string{"stat", "st"},
			},
		},
	}
	_ = eng.RegisterBuiltin(plug)

	// Primary command
	if _, ok := eng.Resolve("status"); !ok {
		t.Errorf("expected Resolve('status') to succeed")
	}

	// Aliases
	if _, ok := eng.Resolve("stat"); !ok {
		t.Errorf("expected Resolve('stat') via alias to succeed")
	}
	if _, ok := eng.Resolve("st"); !ok {
		t.Errorf("expected Resolve('st') via alias to succeed")
	}
	if _, ok := eng.Resolve("unknown"); ok {
		t.Errorf("expected Resolve('unknown') to fail")
	}
}

func TestKernelEngine_PatchCycleDetection(t *testing.T) {
	eng := New(&mockHost{})
	eng.PatchCommand("a", "b")
	eng.PatchCommand("b", "c")
	eng.PatchCommand("c", "a") // Cycle A -> B -> C -> A

	// Must not hang or stack overflow
	proceed, filtered := eng.FilterCommand("/a hello")
	if !proceed {
		t.Errorf("expected proceed to be true")
	}
	if filtered == "" {
		t.Errorf("expected filtered command to not be empty")
	}
}

func TestKernelEngine_InterceptorDeduplication(t *testing.T) {
	eng := New(&mockHost{})
	plug := &mockPlugin{
		manifests: []plugin.Manifest{
			{Command: "cmd1", Capabilities: plugin.CapInterceptor},
			{Command: "cmd2", Capabilities: plugin.CapInterceptor},
		},
	}
	_ = eng.RegisterBuiltin(plug)

	insp := eng.Inspect()
	if count, ok := insp["interceptors_count"].(int); !ok || count != 1 {
		t.Errorf("expected exactly 1 deduplicated interceptor, got %v", insp["interceptors_count"])
	}
}
