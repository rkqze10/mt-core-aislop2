package superuser

import (
	"sync"

	"github.com/rkqze10/mt-tui/pkg/framebuffer"
)

// SuperuserHookManager stores dynamic text and pixel pipeline hooks.
type SuperuserHookManager struct {
	mu         sync.RWMutex
	textHooks  map[string]string
	pixelHooks map[string]DynamicPixelHook
}

// NewSuperuserHookManager initializes a new hook manager.
func NewSuperuserHookManager() *SuperuserHookManager {
	return &SuperuserHookManager{
		textHooks:  make(map[string]string),
		pixelHooks: make(map[string]DynamicPixelHook),
	}
}

// SetTextHook replaces the content of a designated text target (e.g., "banner", "status").
func (shm *SuperuserHookManager) SetTextHook(target, content string) {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	shm.textHooks[target] = content
}

// ClearTextHook removes a specific text hook.
func (shm *SuperuserHookManager) ClearTextHook(target string) {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	delete(shm.textHooks, target)
}

// GetTextHook retrieves a text hook value or returns the fallback.
func (shm *SuperuserHookManager) GetTextHook(target, fallback string) string {
	shm.mu.RLock()
	defer shm.mu.RUnlock()
	if c, ok := shm.textHooks[target]; ok && c != "" {
		return c
	}
	return fallback
}

// RegisterPixelHook adds or replaces a dynamic pixel rendering pipeline.
func (shm *SuperuserHookManager) RegisterPixelHook(id string, hook DynamicPixelHook) {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	shm.pixelHooks[id] = hook
}

// RemovePixelHook unregisters a dynamic pixel pipeline.
func (shm *SuperuserHookManager) RemovePixelHook(id string) {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	delete(shm.pixelHooks, id)
}

// HasPixelHooks checks if any pixel pipelines are registered.
func (shm *SuperuserHookManager) HasPixelHooks() bool {
	shm.mu.RLock()
	defer shm.mu.RUnlock()
	return len(shm.pixelHooks) > 0
}

// ExecutePixelPipelines invokes all registered pixel pipelines in deterministic order with panic protection.
func (shm *SuperuserHookManager) ExecutePixelPipelines(fb *framebuffer.VirtualFramebuffer) {
	shm.mu.RLock()
	ids := make([]string, 0, len(shm.pixelHooks))
	for id, h := range shm.pixelHooks {
		if h != nil {
			ids = append(ids, id)
		}
	}
	// Sort IDs for deterministic render order across frames
	for i := 0; i < len(ids)-1; i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[i] > ids[j] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	hooks := make([]DynamicPixelHook, 0, len(ids))
	for _, id := range ids {
		hooks = append(hooks, shm.pixelHooks[id])
	}
	shm.mu.RUnlock()

	for _, hook := range hooks {
		func() {
			defer func() {
				_ = recover() // Protect hypervisor against panicking plugin hooks
			}()
			hook(fb, fb.Width, fb.Height)
		}()
	}
}
