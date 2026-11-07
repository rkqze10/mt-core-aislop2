package plugin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/config"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

func (dm *DynamicModule) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {
	// Handled via dm.runDaemon inside Init()
}

func (dm *DynamicModule) runDaemon(ctx context.Context, host superuser.SuperuserHost) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := dm.EnsureExecutable(); err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(3 * time.Second):
					continue
				}
			}

			var cmd *exec.Cmd
			findBin := func(names ...string) string {
				if prefix := os.Getenv("PREFIX"); prefix != "" {
					for _, name := range names {
						cand := filepath.Join(prefix, "bin", name)
						if _, err := os.Stat(cand); err == nil {
							return cand
						}
					}
				}
				for _, name := range names {
					if p, err := exec.LookPath(name); err == nil {
						return p
					}
				}
				if len(names) > 0 {
					return names[0]
				}
				return "sh"
			}

			switch dm.pluginType {
			case PluginTypePython:
				pyExec := findBin("python3", "python")
				cmd = exec.CommandContext(ctx, pyExec, dm.sourcePath, "--daemon")
			case PluginTypeShell:
				shExec := findBin("bash", "sh")
				cmd = exec.CommandContext(ctx, shExec, dm.sourcePath, "--daemon")
			case PluginTypeNode:
				nodeExec := findBin("node", "nodejs")
				cmd = exec.CommandContext(ctx, nodeExec, dm.sourcePath, "--daemon")
			default:
				cmd = exec.CommandContext(ctx, dm.execPath, "--daemon")
			}

			w, h := host.GetTerminalDimensions()
			cmd.Env = append(os.Environ(),
				"MT_DAEMON=1",
				"MT_SUPERUSER=1",
				fmt.Sprintf("MT_WIDTH=%d", w),
				fmt.Sprintf("MT_HEIGHT=%d", h),
				fmt.Sprintf("MT_THEME=%s", host.CurrentTheme().Name),
				fmt.Sprintf("MT_PLUGINS_DIR=%s", config.GetPluginsDir()),
			)
			cmd.Env = append(cmd.Env, host.KV().ExportEnv()...)

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					continue
				}
			}

			if err := cmd.Start(); err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					continue
				}
			}

			scanner := bufio.NewScanner(stdout)
			scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					host.ExecuteIPCAction(line)
				}
			}

			_ = cmd.Wait()

			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func (dm *DynamicModule) Intercept(c string) (bool, string)                 { return true, c }
func (dm *DynamicModule) Update(msg tea.Msg, ctx *PluginContext) (bool, tea.Cmd) {
	return false, nil
}

// View executes the plugin synchronously and renders its output or parses IPC actions.
func (dm *DynamicModule) View(ctx *PluginContext) (outView string) {
	defer func() {
		if r := recover(); r != nil {
			outView = fmt.Sprintf("⚠️ Eklenti yürütme hatası (Panic koruması): %v", r)
		}
	}()

	if err := dm.EnsureExecutable(); err != nil {
		return fmt.Sprintf("⚠️ SİSTEM ÇALIŞTIRMA HATASI:\n%v", err)
	}

	timeout := 15 * time.Second
	if len(dm.manifests) > 0 && dm.manifests[0].Timeout > 0 {
		timeout = dm.manifests[0].Timeout
	}
	execCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	findBin := func(names ...string) string {
		if prefix := os.Getenv("PREFIX"); prefix != "" {
			for _, name := range names {
				cand := filepath.Join(prefix, "bin", name)
				if _, err := os.Stat(cand); err == nil {
					return cand
				}
			}
		}
		for _, name := range names {
			if p, err := exec.LookPath(name); err == nil {
				return p
			}
		}
		if len(names) > 0 {
			return names[0]
		}
		return "sh"
	}

	switch dm.pluginType {
	case PluginTypePython:
		pyExec := findBin("python3", "python")
		cmdArgs := append([]string{dm.sourcePath}, ctx.Args...)
		cmd = exec.CommandContext(execCtx, pyExec, cmdArgs...)
	case PluginTypeShell:
		shExec := findBin("bash", "sh")
		cmdArgs := append([]string{dm.sourcePath}, ctx.Args...)
		cmd = exec.CommandContext(execCtx, shExec, cmdArgs...)
	case PluginTypeNode:
		nodeExec := findBin("node", "nodejs")
		cmdArgs := append([]string{dm.sourcePath}, ctx.Args...)
		cmd = exec.CommandContext(execCtx, nodeExec, cmdArgs...)
	default:
		cmd = exec.CommandContext(execCtx, dm.execPath, ctx.Args...)
	}

	if ctx.PipeInput != "" {
		cmd.Stdin = strings.NewReader(ctx.PipeInput)
	}

	pipeSnippet := ctx.PipeInput
	if len(pipeSnippet) > 1024 {
		pipeSnippet = pipeSnippet[:1024]
	}

	w, h := ctx.Host.GetTerminalDimensions()

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("MT_INVOKED_CMD=%s", ctx.InvokedCmd),
		fmt.Sprintf("MT_PIPE_INPUT=%s", pipeSnippet),
		fmt.Sprintf("MT_WIDTH=%d", w),
		fmt.Sprintf("MT_HEIGHT=%d", h),
		"MT_SUPERUSER=1",
		fmt.Sprintf("MT_THEME=%s", ctx.Theme.Name),
		fmt.Sprintf("MT_PLUGINS_DIR=%s", config.GetPluginsDir()),
	)
	cmd.Env = append(cmd.Env, ctx.Host.KV().ExportEnv()...)

	output, err := cmd.CombinedOutput()
	if execCtx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("⚠️ Hata: Eklenti zaman aşımına uğradı (%s sınırı).", timeout)
	}
	if err != nil {
		return fmt.Sprintf("⚠️ Yürütme Hatası: %v\n\nÇıktı:\n%s", err, string(output))
	}

	rawOut := strings.TrimSpace(string(output))
	var textOutput strings.Builder
	var lastActionFeedback string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			res := ctx.Host.ExecuteIPCAction(line)
			if res != "" {
				lastActionFeedback = res
			}
		} else {
			textOutput.WriteString(line + "\n")
		}
	}

	plainText := strings.TrimSpace(textOutput.String())
	if plainText != "" {
		return plainText
	}
	if lastActionFeedback != "" {
		return lastActionFeedback
	}
	return rawOut
}

