package kernel

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rkqze10/mt-tui/pkg/plugin"
	"github.com/rkqze10/mt-tui/pkg/superuser"
)

// KernelEngine manages built-in commands, dynamic modules, interceptors, and command resolution.
type KernelEngine struct {
	mu           sync.RWMutex
	builtins     []plugin.LowLevelPlugin
	dynamicMods  map[string]*plugin.DynamicModule
	runtimeCmds  map[string]plugin.LowLevelPlugin
	patchedCmds  map[string]string
	interceptors []plugin.LowLevelPlugin
	host         superuser.SuperuserHost
}

// New creates a new KernelEngine instance.
func New(host superuser.SuperuserHost) *KernelEngine {
	return &KernelEngine{
		dynamicMods: make(map[string]*plugin.DynamicModule),
		runtimeCmds: make(map[string]plugin.LowLevelPlugin),
		patchedCmds: make(map[string]string),
		host:        host,
	}
}

// RegisterBuiltin registers an internal LowLevelPlugin into the kernel.
func (ke *KernelEngine) RegisterBuiltin(p plugin.LowLevelPlugin) error {
	ke.mu.Lock()
	defer ke.mu.Unlock()

	if err := p.Init(ke.host); err != nil {
		return err
	}
	ke.builtins = append(ke.builtins, p)

	for _, m := range p.Manifests() {
		if m.Capabilities&plugin.CapInterceptor != 0 {
			alreadyRegistered := false
			for _, existing := range ke.interceptors {
				if existing == p {
					alreadyRegistered = true
					break
				}
			}
			if !alreadyRegistered {
				ke.interceptors = append(ke.interceptors, p)
			}
		}
	}
	return nil
}

// Builtins returns a slice of registered built-in plugins.
func (ke *KernelEngine) Builtins() []plugin.LowLevelPlugin {
	ke.mu.RLock()
	defer ke.mu.RUnlock()
	res := make([]plugin.LowLevelPlugin, len(ke.builtins))
	copy(res, ke.builtins)
	return res
}

// RegisterRuntimeCommand adds or overrides an in-memory command at runtime.
func (ke *KernelEngine) RegisterRuntimeCommand(cmd string, p plugin.LowLevelPlugin) {
	ke.mu.Lock()
	defer ke.mu.Unlock()
	clean := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(cmd), "/"))
	if clean != "" {
		ke.runtimeCmds[clean] = p
	}
}

// UnregisterRuntimeCommand removes an in-memory runtime command.
func (ke *KernelEngine) UnregisterRuntimeCommand(cmd string) {
	ke.mu.Lock()
	defer ke.mu.Unlock()
	clean := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(cmd), "/"))
	delete(ke.runtimeCmds, clean)
}

// PatchCommand maps a command alias or rewrite to another command.
func (ke *KernelEngine) PatchCommand(targetCmd, rewriteCmd string) {
	ke.mu.Lock()
	defer ke.mu.Unlock()
	cleanTarget := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(targetCmd), "/"))
	cleanRewrite := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(rewriteCmd), "/"))
	if cleanTarget != "" && cleanRewrite != "" && cleanTarget != cleanRewrite {
		ke.patchedCmds[cleanTarget] = cleanRewrite
	}
}

// UnpatchCommand removes a command rewrite.
func (ke *KernelEngine) UnpatchCommand(targetCmd string) {
	ke.mu.Lock()
	defer ke.mu.Unlock()
	cleanTarget := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(targetCmd), "/"))
	delete(ke.patchedCmds, cleanTarget)
}

// CheckAndReload scans the given directory for new, updated, or removed plugin scripts.
func (ke *KernelEngine) CheckAndReload(dirPath string) (bool, int) {
	absDir, err := filepath.Abs(dirPath)
	if err != nil {
		absDir = dirPath
	}

	_ = os.MkdirAll(absDir, 0755)
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return false, 0
	}

	type candidateInfo struct {
		path    string
		modTime time.Time
	}
	var candidates []candidateInfo
	currentFiles := make(map[string]bool)

	ignoredExts := map[string]bool{
		".md": true, ".txt": true, ".json": true, ".tmp": true,
		".log": true, ".yml": true, ".yaml": true, ".png": true,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ignoredExts[ext] {
			continue
		}

		fullPath := filepath.Join(absDir, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		isCandidate := strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, ".py") ||
			strings.HasSuffix(name, ".sh") ||
			(info.Mode().Perm()&0111 != 0 && runtime.GOOS != "windows") ||
			(strings.HasSuffix(name, ".exe") && runtime.GOOS == "windows")

		if isCandidate {
			currentFiles[fullPath] = true
			candidates = append(candidates, candidateInfo{path: fullPath, modTime: info.ModTime()})
		}
	}

	var toUpdate []*plugin.DynamicModule
	for _, c := range candidates {
		ke.mu.RLock()
		existing, exists := ke.dynamicMods[c.path]
		needReload := !exists || c.modTime.After(existing.LastMod())
		ke.mu.RUnlock()

		if needReload {
			if mod, err := plugin.ParseDynamicModule(c.path); err == nil {
				toUpdate = append(toUpdate, mod)
			}
		}
	}

	var initializedMods []*plugin.DynamicModule
	for _, mod := range toUpdate {
		if err := mod.Init(ke.host); err == nil {
			initializedMods = append(initializedMods, mod)
		}
	}

	ke.mu.Lock()
	defer ke.mu.Unlock()

	changed := false
	for _, mod := range initializedMods {
		if existing, exists := ke.dynamicMods[mod.SourcePath()]; exists && existing != nil {
			_ = existing.Destroy()
		}
		ke.dynamicMods[mod.SourcePath()] = mod
		changed = true
	}

	for path, mod := range ke.dynamicMods {
		if !currentFiles[path] {
			if mod != nil {
				_ = mod.Destroy()
			}
			delete(ke.dynamicMods, path)
			changed = true
		}
	}

	return changed, len(ke.dynamicMods)
}

// FilterCommand applies command patches and passes the command through registered interceptors.
func (ke *KernelEngine) FilterCommand(input string) (bool, string) {
	ke.mu.RLock()
	defer ke.mu.RUnlock()

	current := strings.TrimSpace(input)
	if current == "" {
		return true, ""
	}

	// Apply patches with cycle detection (max 5 hops)
	visited := make(map[string]bool)
	for hops := 0; hops < 5; hops++ {
		firstWord := strings.Split(current, " ")[0]
		cleanCmd := strings.ToLower(strings.TrimPrefix(firstWord, "/"))
		if cleanCmd == "" || visited[cleanCmd] {
			break
		}
		visited[cleanCmd] = true

		if rewrite, ok := ke.patchedCmds[cleanCmd]; ok && rewrite != "" {
			parts := strings.SplitN(current, " ", 2)
			cleanRewrite := strings.TrimPrefix(rewrite, "/")
			if len(parts) > 1 {
				current = "/" + cleanRewrite + " " + parts[1]
			} else {
				current = "/" + cleanRewrite
			}
		} else {
			break
		}
	}

	for _, it := range ke.interceptors {
		proceed, modified := it.Intercept(current)
		if !proceed {
			return false, ""
		}
		current = modified
	}
	return true, current
}

// Resolve looks up a command by name across runtime commands, dynamic modules, and built-ins.
// It matches primary command names and declared aliases.
func (ke *KernelEngine) Resolve(cmd string) (plugin.LowLevelPlugin, bool) {
	ke.mu.RLock()
	defer ke.mu.RUnlock()

	target := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(cmd), "/"))
	if target == "" {
		return nil, false
	}

	// Helper to check manifest command and aliases
	matchesTarget := func(m plugin.Manifest) bool {
		if strings.ToLower(m.Command) == target {
			return true
		}
		for _, a := range m.Aliases {
			if strings.ToLower(strings.TrimSpace(a)) == target {
				return true
			}
		}
		return false
	}

	// 1. Check runtime dynamic commands (highest priority)
	if rc, ok := ke.runtimeCmds[target]; ok {
		return rc, true
	}
	for _, rc := range ke.runtimeCmds {
		for _, m := range rc.Manifests() {
			if matchesTarget(m) {
				return rc, true
			}
		}
	}

	// 2. Check dynamic on-disk modules
	var paths []string
	for p := range ke.dynamicMods {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		dm := ke.dynamicMods[p]
		for _, m := range dm.Manifests() {
			if matchesTarget(m) {
				return dm, true
			}
		}
	}

	// 3. Check internal built-ins
	for _, b := range ke.builtins {
		for _, m := range b.Manifests() {
			if matchesTarget(m) {
				return b, true
			}
		}
	}

	return nil, false
}

// Search returns all manifests matching the search query, intelligently ranked.
func (ke *KernelEngine) Search(rawQuery string) []plugin.Manifest {
	ke.mu.RLock()
	defer ke.mu.RUnlock()

	q := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(rawQuery), "/"))
	seen := make(map[string]bool)

	type matchItem struct {
		manifest plugin.Manifest
		score    int
	}
	var matches []matchItem

	addIfMatches := func(m plugin.Manifest) {
		cmdKey := strings.ToLower(m.Command)
		if seen[cmdKey] {
			return
		}

		score := -1
		if q == "" {
			score = 1
		} else if cmdKey == q {
			score = 0
		} else if strings.HasPrefix(cmdKey, q) {
			score = 1
		} else if strings.Contains(cmdKey, q) {
			score = 2
		} else {
			// Check aliases
			for _, a := range m.Aliases {
				cleanA := strings.ToLower(strings.TrimSpace(a))
				if cleanA == q {
					score = 1
					break
				} else if strings.HasPrefix(cleanA, q) {
					score = 2
					break
				}
			}
		}

		if score < 0 && (strings.Contains(strings.ToLower(m.Description), q) || strings.Contains(strings.ToLower(m.Category), q)) {
			score = 3
		}

		if score >= 0 {
			seen[cmdKey] = true
			matches = append(matches, matchItem{manifest: m, score: score})
		}
	}


	for _, rc := range ke.runtimeCmds {
		for _, m := range rc.Manifests() {
			addIfMatches(m)
		}
	}

	for _, b := range ke.builtins {
		for _, m := range b.Manifests() {
			addIfMatches(m)
		}
	}

	for _, dm := range ke.dynamicMods {
		for _, m := range dm.Manifests() {
			addIfMatches(m)
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score < matches[j].score
		}
		return matches[i].manifest.Command < matches[j].manifest.Command
	})

	res := make([]plugin.Manifest, len(matches))
	for i, m := range matches {
		res[i] = m.manifest
	}

	return res
}

// Inspect returns an in-depth dump of kernel state, loaded plugins, patched commands, and interceptors.
func (ke *KernelEngine) Inspect() map[string]any {
	ke.mu.RLock()
	defer ke.mu.RUnlock()

	var dynList []string
	for p := range ke.dynamicMods {
		dynList = append(dynList, filepath.Base(p))
	}
	sort.Strings(dynList)

	var runtimeList []string
	for c := range ke.runtimeCmds {
		runtimeList = append(runtimeList, c)
	}
	sort.Strings(runtimeList)

	patches := make(map[string]string)
	for k, v := range ke.patchedCmds {
		patches[k] = v
	}

	return map[string]any{
		"dynamic_modules_count": len(ke.dynamicMods),
		"dynamic_modules":       dynList,
		"runtime_commands":      runtimeList,
		"builtins_count":        len(ke.builtins),
		"interceptors_count":    len(ke.interceptors),
		"command_patches":       patches,
	}
}

// Shutdown gracefully stops all builtins, dynamic modules, and runtime commands.
func (ke *KernelEngine) Shutdown() {
	ke.mu.Lock()
	defer ke.mu.Unlock()
	for _, b := range ke.builtins {
		_ = b.Destroy()
	}
	for _, dm := range ke.dynamicMods {
		_ = dm.Destroy()
	}
	for _, rc := range ke.runtimeCmds {
		_ = rc.Destroy()
	}
}
