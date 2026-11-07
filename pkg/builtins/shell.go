package builtins

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// ShellPlugin executes native OS commands or filters pipeline data with grep.
type ShellPlugin struct{}

func (sp *ShellPlugin) Manifests() []plugin.Manifest {
	return []plugin.Manifest{
		{
			Command:      "sh",
			Title:        "KABUK SÜRÜCÜSÜ",
			Category:     "OS",
			Description:  "Doğrudan işletim sistemi komutu yürütür (/sh ls)",
			Capabilities: plugin.CapUI | plugin.CapOSProcess,
		},
		{
			Command:      "grep",
			Title:        "METİN FİLTRESİ",
			Category:     "TOOL",
			Description:  "Girdiyi satır satır filtreler (Pipelining)",
			Capabilities: plugin.CapUI,
		},
	}
}

func (sp *ShellPlugin) Init(h superuser.SuperuserHost) error                        { return nil }
func (sp *ShellPlugin) Destroy() error                                             { return nil }
func (sp *ShellPlugin) StartDaemon(ctx context.Context, host superuser.SuperuserHost) {}
func (sp *ShellPlugin) Intercept(c string) (bool, string)                           { return true, c }
func (sp *ShellPlugin) Update(msg tea.Msg, ctx *plugin.PluginContext) (bool, tea.Cmd) {
	return false, nil
}

func (sp *ShellPlugin) View(ctx *plugin.PluginContext) (outStr string) {
	defer func() {
		if r := recover(); r != nil {
			outStr = fmt.Sprintf("⚠️ Kabuk yürütme hatası (Panic koruması): %v", r)
		}
	}()

	if ctx.InvokedCmd == "grep" {
		if len(ctx.Args) == 0 {
			return "Kullanım: /cmd | grep [-i] [-v] <kelime>"
		}
		caseInsensitive := false
		invertMatch := false
		target := ""
		for _, arg := range ctx.Args {
			switch arg {
			case "-i", "--ignore-case":
				caseInsensitive = true
			case "-v", "--invert-match":
				invertMatch = true
			default:
				if target == "" {
					target = arg
				}
			}
		}
		if target == "" {
			return "Kullanım: /cmd | grep [-i] [-v] <kelime>"
		}
		var matched []string
		scanner := bufio.NewScanner(strings.NewReader(ctx.PipeInput))
		scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
		searchTarget := target
		if caseInsensitive {
			searchTarget = strings.ToLower(target)
		}
		for scanner.Scan() {
			line := scanner.Text()
			checkLine := line
			if caseInsensitive {
				checkLine = strings.ToLower(line)
			}
			contains := strings.Contains(checkLine, searchTarget)
			if (contains && !invertMatch) || (!contains && invertMatch) {
				matched = append(matched, line)
			}
		}
		if len(matched) == 0 {
			return fmt.Sprintf("'%s' ile eşleşen satır bulunamadı.", target)
		}
		return strings.Join(matched, "\n")
	}

	if len(ctx.Args) == 0 {
		return "Kullanım: /sh <komut> [parametreler]\nÖrnek: /sh uname -a veya /sh ls -la"
	}

	binary := ctx.Args[0]
	args := ctx.Args[1:]

	// Interactive terminal programs detection
	interactiveTools := map[string]bool{
		"htop": true, "top": true, "vim": true, "vi": true, "nano": true,
		"less": true, "more": true, "fzf": true, "lazygit": true, "tmux": true,
	}
	if interactiveTools[strings.ToLower(binary)] {
		ctx.Host.RequestExternalTUI(binary, args, binary, "")
		return fmt.Sprintf("⏳ İnteraktif araç algılandı: [%s]. Terminal devri için onay bekleniyor...", binary)
	}

	execCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Check if the command contains shell metacharacters (pipes, redirects, wildcards)
	fullCmd := strings.Join(ctx.Args, " ")
	hasShellMetachars := strings.ContainsAny(fullCmd, "|><;&*$`")

	findShell := func() string {
		// 1. Termux PREFIX shell
		if prefix := os.Getenv("PREFIX"); prefix != "" {
			termuxSh := filepath.Join(prefix, "bin", "sh")
			if _, err := os.Stat(termuxSh); err == nil {
				return termuxSh
			}
			termuxBash := filepath.Join(prefix, "bin", "bash")
			if _, err := os.Stat(termuxBash); err == nil {
				return termuxBash
			}
		}
		// 2. Standard LookPath
		if p, err := exec.LookPath("bash"); err == nil {
			return p
		}
		if p, err := exec.LookPath("sh"); err == nil {
			return p
		}
		// 3. Known Termux direct paths
		if _, err := os.Stat("/data/data/com.termux/files/usr/bin/sh"); err == nil {
			return "/data/data/com.termux/files/usr/bin/sh"
		}
		return "sh"
	}

	if hasShellMetachars {
		if runtime.GOOS == "windows" {
			binary = "cmd.exe"
			args = []string{"/c", fullCmd}
		} else {
			binary = findShell()
			args = []string{"-c", fullCmd}
		}
	} else if _, err := exec.LookPath(binary); err != nil {
		if runtime.GOOS == "windows" {
			binary = "cmd.exe"
			args = append([]string{"/c", ctx.Args[0]}, ctx.Args[1:]...)
		} else {
			binary = findShell()
			args = []string{"-c", fullCmd}
		}
	}

	cmd := exec.CommandContext(execCtx, binary, args...)
	if ctx.PipeInput != "" {
		cmd.Stdin = strings.NewReader(ctx.PipeInput)
	}

	out, err := cmd.CombinedOutput()
	if execCtx.Err() == context.DeadlineExceeded {
		return "⚠️ Hata: Komut zaman aşımına uğradı (15 saniye sınırı)."
	}
	if err != nil {
		return fmt.Sprintf("⚠️ Komut Hatası: %v\n\nÇıktı:\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}
