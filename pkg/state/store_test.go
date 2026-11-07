package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSharedStateStore(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test_state.json")

	store := NewSharedStateStore(storePath)
	store.Set("my_key", "my_val")
	store.Set("counter", "42")

	val, ok := store.Get("my_key")
	if !ok || val != "my_val" {
		t.Errorf("expected 'my_val', got %q", val)
	}

	envs := store.ExportEnv()
	expectedPrefix := "MT_KV_MY_KEY=my_val"
	found := false
	for _, env := range envs {
		if env == expectedPrefix {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected env %q in %v", expectedPrefix, envs)
	}

	store2 := NewSharedStateStore(storePath)
	val2, ok2 := store2.Get("counter")
	if !ok2 || val2 != "42" {
		t.Errorf("expected persisted '42', got %q", val2)
	}

	_ = os.Remove(storePath)
}

func TestSharedStateStore_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test_dedup.json")

	store := NewSharedStateStore(storePath)
	// Keys that would collide in environment variable representation
	store.Set("app-port", "3000")
	store.Set("app_port", "8080")

	envs := store.ExportEnv()
	count := 0
	for _, e := range envs {
		if len(e) >= 15 && e[:15] == "MT_KV_APP_PORT=" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 deduplicated MT_KV_APP_PORT, got %d in %v", count, envs)
	}
}

func TestSharedStateStore_EmptyPathSafety(t *testing.T) {
	// Should safely handle in-memory store without writing random .tmp files
	store := NewSharedStateStore("")
	store.Set("temp", "value")
	val, ok := store.Get("temp")
	if !ok || val != "value" {
		t.Errorf("expected 'value', got %q", val)
	}
}
