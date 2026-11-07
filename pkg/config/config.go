package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// PersistentConfig holds user settings stored on disk.
type PersistentConfig struct {
	Theme   string   `json:"theme"`
	History []string `json:"history"`
}

var (
	baseDirOnce sync.Once
	cachedBase  string
)

// GetBaseDir locates or creates the base directory for MT Core data.
// It prioritizes current workspace directory if plugins/ exists or can be created locally.
func GetBaseDir() string {
	baseDirOnce.Do(func() {
		cwd, err := os.Getwd()
		if err == nil {
			localPlugins := filepath.Join(cwd, "plugins")
			if info, err := os.Stat(localPlugins); err == nil && info.IsDir() {
				cachedBase = cwd
				return
			}
			if _, err := os.Stat(filepath.Join(cwd, "main.go")); err == nil {
				_ = os.MkdirAll(localPlugins, 0755)
				cachedBase = cwd
				return
			}
			if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
				_ = os.MkdirAll(localPlugins, 0755)
				cachedBase = cwd
				return
			}
		}

		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			execPlugins := filepath.Join(execDir, "plugins")
			if info, err := os.Stat(execPlugins); err == nil && info.IsDir() {
				cachedBase = execDir
				return
			}
		}

		if userConfig, err := os.UserConfigDir(); err == nil {
			dir := filepath.Join(userConfig, "mt")
			if err := os.MkdirAll(filepath.Join(dir, "plugins"), 0755); err == nil {
				cachedBase = dir
				return
			}
		}

		// Termux specific prefix check
		if prefix := os.Getenv("PREFIX"); prefix != "" && strings.Contains(prefix, "com.termux") {
			termuxHome := os.Getenv("HOME")
			if termuxHome != "" {
				dir := filepath.Join(termuxHome, ".mt")
				if err := os.MkdirAll(filepath.Join(dir, "plugins"), 0755); err == nil {
					cachedBase = dir
					return
				}
			}
		}

		if home, err := os.UserHomeDir(); err == nil {
			dir := filepath.Join(home, ".mt")
			if err := os.MkdirAll(filepath.Join(dir, "plugins"), 0755); err == nil {
				cachedBase = dir
				return
			}
		}

		_ = os.MkdirAll("plugins", 0755)
		absCur, _ := filepath.Abs(".")
		cachedBase = absCur
	})

	return cachedBase
}

// GetPluginsDir returns the absolute path to the plugins directory.
func GetPluginsDir() string {
	p := filepath.Join(GetBaseDir(), "plugins")
	_ = os.MkdirAll(p, 0755)
	absP, _ := filepath.Abs(p)
	return absP
}

// GetBinDir returns the absolute path to the pre-compiled binary cache directory.
func GetBinDir() string {
	b := filepath.Join(GetBaseDir(), ".mt_bin")
	_ = os.MkdirAll(b, 0755)
	absB, _ := filepath.Abs(b)
	return absB
}

// GetConfigPath returns the path to config.json.
func GetConfigPath() string {
	return filepath.Join(GetBaseDir(), "config.json")
}

// GetStatePath returns the path to state storage file.
func GetStatePath() string {
	return filepath.Join(GetBaseDir(), ".mt_state.json")
}

var saveMu sync.Mutex

// LoadConfig reads the configuration from disk with safe defaults.
func LoadConfig() PersistentConfig {
	cfg := PersistentConfig{
		Theme:   "amber",
		History: []string{},
	}

	data, err := os.ReadFile(GetConfigPath())
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	if cfg.History == nil {
		cfg.History = []string{}
	}
	validThemes := map[string]bool{"amber": true, "emerald": true, "cyber": true, "mono": true}
	if !validThemes[cfg.Theme] {
		cfg.Theme = "amber"
	}

	return cfg
}

// SaveConfig writes the configuration atomically to disk with concurrency locking.
func SaveConfig(cfg PersistentConfig) error {
	saveMu.Lock()
	defer saveMu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	target := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(target)
		return os.Rename(tmp, target)
	}
	return nil
}
