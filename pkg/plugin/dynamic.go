package plugin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rkqze10/mt-tui/pkg/config"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// PluginType classifies how a script or executable should be invoked.
type PluginType int

const (
	PluginTypeGoSource PluginType = iota
	PluginTypePython
	PluginTypeShell
	PluginTypeNode
	PluginTypeBinary
)

// DynamicModule represents an on-disk script or executable loaded dynamically.
type DynamicModule struct {
	mu           sync.Mutex
	pluginType   PluginType
	sourcePath   string
	execPath     string
	lastMod      time.Time
	manifests    []Manifest
	host         superuser.SuperuserHost
	isDaemon     bool
	daemonCtx    context.Context
	cancelDaemon context.CancelFunc
}

// SourcePath returns the source file path.
func (dm *DynamicModule) SourcePath() string { return dm.sourcePath }

// LastMod returns the last modification time.
func (dm *DynamicModule) LastMod() time.Time { return dm.lastMod }

// Manifests returns the manifests declared by this module.
func (dm *DynamicModule) Manifests() []Manifest { return dm.manifests }

// ParseDynamicModule scans a script or binary and extracts commands and configuration.
func ParseDynamicModule(filePath string) (*DynamicModule, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	baseName := filepath.Base(absPath)
	var pType PluginType
	var cmdBaseName string

	switch {
	case strings.HasSuffix(baseName, ".go"):
		pType = PluginTypeGoSource
		cmdBaseName = strings.TrimSuffix(baseName, ".go")
	case strings.HasSuffix(baseName, ".py"):
		pType = PluginTypePython
		cmdBaseName = strings.TrimSuffix(baseName, ".py")
	case strings.HasSuffix(baseName, ".sh"):
		pType = PluginTypeShell
		cmdBaseName = strings.TrimSuffix(baseName, ".sh")
	case strings.HasSuffix(baseName, ".js") || strings.HasSuffix(baseName, ".mjs") || strings.HasSuffix(baseName, ".cjs"):
		pType = PluginTypeNode
		cmdBaseName = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(baseName, ".js"), ".mjs"), ".cjs")
	default:
		pType = PluginTypeBinary
		cmdBaseName = strings.TrimSuffix(baseName, ".exe")
	}

	var manifests []Manifest
	manifestMap := make(map[string]bool)

	addManifest := func(m Manifest) {
		cmdClean := strings.ToLower(strings.TrimSpace(m.Command))
		if cmdClean == "" || manifestMap[cmdClean] {
			return
		}
		manifestMap[cmdClean] = true
		m.Command = cmdClean
		manifests = append(manifests, m)
	}

	var singleCmd, defaultTitle, defaultCategory, defaultDesc string
	var defaultAuthor, defaultVersion, defaultIcon string
	var defaultAliases, defaultRequires, defaultPermissions, defaultHooks, defaultTags []string
	defaultTimeout := 15 * time.Second
	defaultCaps := CapUI | CapSuperuser
	isDaemon := false

	if pType != PluginTypeBinary {
		if file, err := os.Open(absPath); err == nil {
			scanner := bufio.NewScanner(file)
			lineCount := 0
			for scanner.Scan() && lineCount < 100 {
				line := strings.TrimSpace(scanner.Text())
				lineCount++

				cleanLine := ""
				if strings.HasPrefix(line, "//") {
					cleanLine = strings.TrimSpace(strings.TrimPrefix(line, "//"))
				} else if strings.HasPrefix(line, "#") {
					cleanLine = strings.TrimSpace(strings.TrimPrefix(line, "#"))
				}

				if strings.HasPrefix(cleanLine, "@cmd:") {
					raw := strings.TrimPrefix(cleanLine, "@cmd:")
					parts := strings.SplitN(raw, ":", 4)
					if len(parts) >= 4 {
						addManifest(Manifest{
							Command:      parts[0],
							Title:        strings.TrimSpace(parts[1]),
							Category:     strings.ToUpper(strings.TrimSpace(parts[2])),
							Description:  strings.TrimSpace(parts[3]),
							Permissions:  defaultPermissions,
							Hooks:        defaultHooks,
							Tags:         defaultTags,
							Capabilities: defaultCaps,
							Timeout:      defaultTimeout,
						})
					}
				} else if strings.HasPrefix(cleanLine, "@command:") || strings.HasPrefix(cleanLine, "@name:") {
					singleCmd = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(cleanLine, "@command:"), "@name:"))
				} else if strings.HasPrefix(cleanLine, "@title:") {
					defaultTitle = strings.TrimSpace(strings.TrimPrefix(cleanLine, "@title:"))
				} else if strings.HasPrefix(cleanLine, "@category:") {
					defaultCategory = strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(cleanLine, "@category:")))
				} else if strings.HasPrefix(cleanLine, "@desc:") {
					defaultDesc = strings.TrimSpace(strings.TrimPrefix(cleanLine, "@desc:"))
				} else if strings.HasPrefix(cleanLine, "@author:") {
					defaultAuthor = strings.TrimSpace(strings.TrimPrefix(cleanLine, "@author:"))
				} else if strings.HasPrefix(cleanLine, "@version:") {
					defaultVersion = strings.TrimSpace(strings.TrimPrefix(cleanLine, "@version:"))
				} else if strings.HasPrefix(cleanLine, "@icon:") {
					defaultIcon = strings.TrimSpace(strings.TrimPrefix(cleanLine, "@icon:"))
				} else if strings.HasPrefix(cleanLine, "@alias:") || strings.HasPrefix(cleanLine, "@aliases:") {
					rawAliases := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(cleanLine, "@aliases:"), "@alias:"))
					for _, a := range strings.Split(rawAliases, ",") {
						if cleanA := strings.TrimSpace(a); cleanA != "" {
							defaultAliases = append(defaultAliases, cleanA)
						}
					}
				} else if strings.HasPrefix(cleanLine, "@requires:") || strings.HasPrefix(cleanLine, "@require:") {
					rawReqs := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(cleanLine, "@requires:"), "@require:"))
					for _, r := range strings.Split(rawReqs, ",") {
						if cleanR := strings.TrimSpace(r); cleanR != "" {
							defaultRequires = append(defaultRequires, cleanR)
						}
					}
				} else if strings.HasPrefix(cleanLine, "@timeout:") {
					rawDur := strings.TrimSpace(strings.TrimPrefix(cleanLine, "@timeout:"))
					if d, err := time.ParseDuration(rawDur); err == nil && d > 0 {
						defaultTimeout = d
					}
				} else if strings.HasPrefix(cleanLine, "@daemon:") {
					val := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(cleanLine, "@daemon:")))
					isDaemon = (val == "true" || val == "1")
					if isDaemon {
						defaultCaps |= CapDaemon
					}
				} else if strings.HasPrefix(cleanLine, "@perm:") || strings.HasPrefix(cleanLine, "@permission:") || strings.HasPrefix(cleanLine, "@permissions:") {
					rawPerms := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(cleanLine, "@permissions:"), "@permission:"), "@perm:"))
					for _, p := range strings.Split(rawPerms, ",") {
						cleanP := strings.ToLower(strings.TrimSpace(p))
						if cleanP != "" {
							defaultPermissions = append(defaultPermissions, cleanP)
							switch cleanP {
							case "net", "network":
								defaultCaps |= CapNetwork
							case "fs", "filesystem":
								defaultCaps |= CapFileSystem
							case "stream", "streaming":
								defaultCaps |= CapStream
							case "event", "events", "eventbus":
								defaultCaps |= CapEventBus
							case "daemon":
								defaultCaps |= CapDaemon
							}
						}
					}
				} else if strings.HasPrefix(cleanLine, "@hook:") || strings.HasPrefix(cleanLine, "@hooks:") {
					rawHooks := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(cleanLine, "@hooks:"), "@hook:"))
					for _, h := range strings.Split(rawHooks, ",") {
						if cleanH := strings.TrimSpace(h); cleanH != "" {
							defaultHooks = append(defaultHooks, cleanH)
						}
					}
				} else if strings.HasPrefix(cleanLine, "@tag:") || strings.HasPrefix(cleanLine, "@tags:") {
					rawTags := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(cleanLine, "@tags:"), "@tag:"))
					for _, t := range strings.Split(rawTags, ",") {
						if cleanT := strings.TrimSpace(t); cleanT != "" {
							defaultTags = append(defaultTags, cleanT)
						}
					}
				}
			}
			file.Close()
		}
	}

	if singleCmd != "" {
		if defaultTitle == "" {
			defaultTitle = strings.ToUpper(singleCmd)
		}
		if defaultCategory == "" {
			defaultCategory = "RING-0"
		}
		addManifest(Manifest{
			Command:      singleCmd,
			Title:        defaultTitle,
			Category:     defaultCategory,
			Description:  defaultDesc,
			Author:       defaultAuthor,
			Version:      defaultVersion,
			Icon:         defaultIcon,
			Aliases:      defaultAliases,
			Requires:     defaultRequires,
			Permissions:  defaultPermissions,
			Hooks:        defaultHooks,
			Tags:         defaultTags,
			Timeout:      defaultTimeout,
			Capabilities: defaultCaps,
		})
	}

	if len(manifests) == 0 {
		addManifest(Manifest{
			Command:      cmdBaseName,
			Title:        strings.ToUpper(cmdBaseName),
			Category:     "RING-0",
			Description:  fmt.Sprintf("Süper kullanıcı dinamik modülü (%s)", cmdBaseName),
			Author:       defaultAuthor,
			Version:      defaultVersion,
			Icon:         defaultIcon,
			Aliases:      defaultAliases,
			Requires:     defaultRequires,
			Permissions:  defaultPermissions,
			Hooks:        defaultHooks,
			Tags:         defaultTags,
			Timeout:      defaultTimeout,
			Capabilities: defaultCaps,
		})
	}

	execPath := absPath
	if pType == PluginTypeGoSource {
		binName := cmdBaseName
		if runtime.GOOS == "windows" {
			binName += ".exe"
		}
		execPath = filepath.Join(config.GetBinDir(), binName)
	}

	return &DynamicModule{
		pluginType: pType,
		sourcePath: absPath,
		execPath:   execPath,
		lastMod:    info.ModTime(),
		manifests:  manifests,
		isDaemon:   isDaemon,
	}, nil
}

// PluginType returns the classified runtime type of the module.
func (dm *DynamicModule) PluginType() PluginType { return dm.pluginType }

// PluginTypeString returns human readable string for the plugin type.
func (dm *DynamicModule) PluginTypeString() string {
	switch dm.pluginType {
	case PluginTypeGoSource:
		return "Go"
	case PluginTypePython:
		return "Python"
	case PluginTypeShell:
		return "Shell"
	case PluginTypeNode:
		return "Node.js"
	default:
		return "Binary"
	}
}

// IsDaemon reports whether this module runs as a background daemon.
func (dm *DynamicModule) IsDaemon() bool { return dm.isDaemon }

// PrimaryManifest returns the first manifest associated with the module.
func (dm *DynamicModule) PrimaryManifest() Manifest {
	if len(dm.manifests) > 0 {
		return dm.manifests[0]
	}
	return Manifest{}
}

// Init sets up the module and starts its daemon if configured.
func (dm *DynamicModule) Init(h superuser.SuperuserHost) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.host = h
	if dm.isDaemon {
		if dm.cancelDaemon != nil {
			dm.cancelDaemon()
		}
		ctx, cancel := context.WithCancel(context.Background())
		dm.daemonCtx = ctx
		dm.cancelDaemon = cancel
		go dm.runDaemon(ctx, h)
	}
	return nil
}

// Destroy cleans up background resources.
func (dm *DynamicModule) Destroy() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if dm.cancelDaemon != nil {
		dm.cancelDaemon()
		dm.cancelDaemon = nil
	}
	return nil
}

func (dm *DynamicModule) EnsureExecutable() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	srcInfo, err := os.Stat(dm.sourcePath)
	if err != nil {
		return fmt.Errorf("kaynak dosya okunamadı: %w", err)
	}

	switch dm.pluginType {
	case PluginTypeGoSource:
		binInfo, binErr := os.Stat(dm.execPath)
		needsBuild := binErr != nil || srcInfo.ModTime().After(binInfo.ModTime())

		if needsBuild {
			goCompiler := "go"
			if prefix := os.Getenv("PREFIX"); prefix != "" {
				termuxGo := filepath.Join(prefix, "bin", "go")
				if _, err := os.Stat(termuxGo); err == nil {
					goCompiler = termuxGo
				}
			}
			if _, lookErr := exec.LookPath(goCompiler); lookErr != nil {
				return fmt.Errorf("Go derleyicisi sistemde bulunamadı ($PATH üzerinde 'go' bulunmalıdır)")
			}
			_ = os.MkdirAll(filepath.Dir(dm.execPath), 0755)
			cmd := exec.Command(goCompiler, "build", "-o", dm.execPath, dm.sourcePath)
			output, buildErr := cmd.CombinedOutput()
			if buildErr != nil {
				return fmt.Errorf("derleme hatası:\n%s", string(output))
			}
		}

	case PluginTypeShell, PluginTypeBinary:
		if runtime.GOOS != "windows" {
			if srcInfo.Mode().Perm()&0111 == 0 {
				_ = os.Chmod(dm.sourcePath, 0755)
			}
		}
	}

	return nil
}
