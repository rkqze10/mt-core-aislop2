package plugin

import "time"

// Capability bitmask defines what privileges a module or plugin requires.
type Capability uint32

const (
	CapUI          Capability = 1 << 0
	CapDaemon      Capability = 1 << 1
	CapInterceptor Capability = 1 << 2
	CapOSProcess   Capability = 1 << 3
	CapSuperuser   Capability = 1 << 4
	CapNetwork     Capability = 1 << 5
	CapFileSystem  Capability = 1 << 6
	CapStream      Capability = 1 << 7
	CapEventBus    Capability = 1 << 8
)

// Manifest contains rich metadata describing a command or plugin.
type Manifest struct {
	Command      string        `json:"command"`
	Title        string        `json:"title"`
	Category     string        `json:"category"`
	Description  string        `json:"description"`
	Author       string        `json:"author"`
	Version      string        `json:"version"`
	Icon         string        `json:"icon"`
	Aliases      []string      `json:"aliases"`
	Requires     []string      `json:"requires"`
	Permissions  []string      `json:"permissions"`
	Hooks        []string      `json:"hooks"`
	Tags         []string      `json:"tags"`
	Timeout      time.Duration `json:"timeout"`
	Capabilities Capability    `json:"capabilities"`
}

// HasCapability checks if the manifest has the given capability flag set.
func (m Manifest) HasCapability(cap Capability) bool {
	return (m.Capabilities & cap) == cap
}

