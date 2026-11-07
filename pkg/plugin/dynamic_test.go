package plugin

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseDynamicModule_Types(t *testing.T) {
	tmpDir := t.TempDir()

	files := []struct {
		name     string
		content  string
		expected PluginType
	}{
		{"test.py", "# @name: py_test\nprint('hi')", PluginTypePython},
		{"test.sh", "#!/bin/bash\n# @name: sh_test\necho hi", PluginTypeShell},
		{"test.go", "package main\n// @name: go_test\nfunc main() {}", PluginTypeGoSource},
		{"test.js", "// @name: js_test\nconsole.log('hi')", PluginTypeNode},
		{"test.mjs", "// @name: mjs_test\nconsole.log('hi')", PluginTypeNode},
		{"test.cjs", "// @name: cjs_test\nconsole.log('hi')", PluginTypeNode},
	}

	for _, f := range files {
		p := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(p, []byte(f.content), 0755); err != nil {
			t.Fatalf("Failed to write file %s: %v", f.name, err)
		}
		dm, err := ParseDynamicModule(p)
		if err != nil {
			t.Fatalf("ParseDynamicModule failed on %s: %v", f.name, err)
		}
		if dm.pluginType != f.expected {
			t.Errorf("File %s: expected %v, got %v", f.name, f.expected, dm.pluginType)
		}
	}
}

func TestParseDynamicModule_NodeManifest(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "node_test.js")

	content := `#!/usr/bin/env node
// @name: nodetest
// @title: Node Test Plugin
// @category: AI-NODE
// @desc: Node.js plugin for testing
// @perm: net, stream, eventbus
// @hook: on_load, on_unload
// @tags: node, test, ai
// @timeout: 20s

console.log("hello");
`
	if err := os.WriteFile(jsFile, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	dm, err := ParseDynamicModule(jsFile)
	if err != nil {
		t.Fatalf("ParseDynamicModule failed: %v", err)
	}

	if dm.pluginType != PluginTypeNode {
		t.Fatalf("Expected PluginTypeNode, got %v", dm.pluginType)
	}

	manifests := dm.Manifests()
	if len(manifests) == 0 {
		t.Fatalf("Expected at least 1 manifest")
	}

	m := manifests[0]
	if m.Command != "nodetest" {
		t.Errorf("Expected command 'nodetest', got %q", m.Command)
	}
	if m.Category != "AI-NODE" {
		t.Errorf("Expected category 'AI-NODE', got %q", m.Category)
	}
	if m.Timeout != 20*time.Second {
		t.Errorf("Expected timeout 20s, got %v", m.Timeout)
	}
	if !m.HasCapability(CapNetwork) || !m.HasCapability(CapStream) || !m.HasCapability(CapEventBus) {
		t.Errorf("Capabilities mismatch: %v", m.Capabilities)
	}
	if len(m.Permissions) != 3 {
		t.Errorf("Expected 3 permissions, got %v", m.Permissions)
	}
	if len(m.Hooks) != 2 {
		t.Errorf("Expected 2 hooks, got %v", m.Hooks)
	}
	if len(m.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %v", m.Tags)
	}
}
