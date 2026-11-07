package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// SharedStateStore manages persistent key-value pairs shared between plugins and core.
type SharedStateStore struct {
	mu       sync.RWMutex
	data     map[string]string
	filePath string
}

// NewSharedStateStore initializes a new or existing KV store.
func NewSharedStateStore(filePath string) *SharedStateStore {
	store := &SharedStateStore{
		data:     make(map[string]string),
		filePath: filePath,
	}
	store.load()
	return store
}

func (s *SharedStateStore) load() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.filePath == "" {
		return
	}
	bytes, err := os.ReadFile(s.filePath)
	if err == nil && len(bytes) > 0 {
		var loaded map[string]string
		if err := json.Unmarshal(bytes, &loaded); err == nil && loaded != nil {
			s.data = loaded
		}
	}
}

func (s *SharedStateStore) save() {
	if s.filePath == "" {
		return
	}

	s.mu.RLock()
	dataCopy := make(map[string]string, len(s.data))
	for k, v := range s.data {
		dataCopy[k] = v
	}
	s.mu.RUnlock()

	bytes, err := json.MarshalIndent(dataCopy, "", "  ")
	if err != nil {
		return
	}

	dir := filepath.Dir(s.filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return
		}
	}

	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, bytes, 0644); err == nil {
		if err := os.Rename(tmp, s.filePath); err != nil {
			_ = os.Remove(s.filePath)
			_ = os.Rename(tmp, s.filePath)
		}
	}
}

// Set stores a key-value pair and persists it.
func (s *SharedStateStore) Set(key, val string) {
	s.mu.Lock()
	s.data[key] = val
	s.mu.Unlock()
	s.save()
}

// Get retrieves a value by key.
func (s *SharedStateStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Delete removes a key and saves the store.
func (s *SharedStateStore) Delete(key string) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
	s.save()
}

// GetAll returns a copy of all key-value pairs.
func (s *SharedStateStore) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[string]string, len(s.data))
	for k, v := range s.data {
		res[k] = v
	}
	return res
}

// ExportEnv converts all stored keys into shell-safe environment variables (MT_KV_<KEY>=<VALUE>).
func (s *SharedStateStore) ExportEnv() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	envMap := make(map[string]string)

	for k, v := range s.data {
		var b strings.Builder
		for _, r := range k {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			} else {
				b.WriteRune('_')
			}
		}
		cleanKey := strings.ToUpper(b.String())
		cleanKey = strings.Trim(cleanKey, "_")
		if cleanKey != "" {
			// Ensure it doesn't start with a number if standalone, though MT_KV_ prefix guarantees leading letter
			cleanVal := strings.ReplaceAll(strings.ReplaceAll(v, "\x00", ""), "\r", "")
			envVar := fmt.Sprintf("MT_KV_%s", cleanKey)
			envMap[envVar] = cleanVal
		}
	}

	var envs []string
	for k, v := range envMap {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(envs)
	return envs
}

