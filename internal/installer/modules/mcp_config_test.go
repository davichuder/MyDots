package modules

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// McpConfigModule tests — T-065 (RED)
// ---------------------------------------------------------------------------

func TestMcpConfig_ID(t *testing.T) {
	m := McpConfigModule{}
	if got := m.ID(); got != types.ModMcpConfig {
		t.Errorf("ID() = %v, want %v", got, types.ModMcpConfig)
	}
}

func TestMcpConfig_Name(t *testing.T) {
	m := McpConfigModule{}
	if got := m.Name(); got != "MCP Config" {
		t.Errorf("Name() = %q, want %q", got, "MCP Config")
	}
}

func TestMcpConfig_Criticality(t *testing.T) {
	m := McpConfigModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestMcpConfig_Dependencies(t *testing.T) {
	m := McpConfigModule{}
	deps := m.Dependencies()
	if len(deps) != 1 || deps[0] != types.ModOpencode {
		t.Errorf("Dependencies() = %v, want [%v]", deps, types.ModOpencode)
	}
}

// --- IsInstalled ---

func TestMcpConfig_IsInstalled(t *testing.T) {
	t.Run("all 6 mcp keys present returns true", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { mcpHomeDir = origHome })

		// Create opencode.json with all 6 MCP entries
		cfg := map[string]interface{}{
			"$schema": "https://opencode.ai/config.json",
			"mcp": map[string]interface{}{
				"supabase":   map[string]interface{}{"type": "local", "enabled": true},
				"angular":    map[string]interface{}{"type": "local", "enabled": true},
				"primeng":    map[string]interface{}{"type": "local", "enabled": true},
				"postman":    map[string]interface{}{"type": "local", "enabled": true},
				"context7":   map[string]interface{}{"type": "remote", "enabled": true},
				"playwright": map[string]interface{}{"type": "local", "enabled": true},
			},
		}
		writeTestJSON(t, tmpDir, cfg)

		m := McpConfigModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when all 6 MCP keys are present")
		}
	})

	t.Run("missing one key returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { mcpHomeDir = origHome })

		cfg := map[string]interface{}{
			"$schema": "https://opencode.ai/config.json",
			"mcp": map[string]interface{}{
				"supabase": map[string]interface{}{"type": "local", "enabled": true},
				// only 1 of 6 keys present
			},
		}
		writeTestJSON(t, tmpDir, cfg)

		m := McpConfigModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when some MCP keys are missing")
		}
	})

	t.Run("no opencode.json returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { mcpHomeDir = origHome })

		m := McpConfigModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when opencode.json does not exist")
		}
	})
}

// --- Install ---

func TestMcpConfig_Install(t *testing.T) {
	t.Run("creates opencode.json with 6 MCP entries when file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { mcpHomeDir = origHome })

		origWrite := mcpWriteFile
		origMkdir := mcpMkdirAll
		t.Cleanup(func() {
			mcpWriteFile = origWrite
			mcpMkdirAll = origMkdir
		})
		mcpMkdirAll = func(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

		var writtenPath string
		var writtenData []byte
		mcpWriteFile = func(path string, data []byte, perm os.FileMode) error {
			writtenPath = path
			writtenData = data
			return os.WriteFile(path, data, perm)
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := McpConfigModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		if writtenPath == "" {
			t.Fatal("expected file to be written")
		}
		if !strings.HasSuffix(writtenPath, "opencode.json") {
			t.Errorf("expected opencode.json, got %q", writtenPath)
		}

		// Verify all 6 MCP keys are present
		var result map[string]interface{}
		if err := json.Unmarshal(writtenData, &result); err != nil {
			t.Fatalf("invalid JSON written: %v", err)
		}
		mcp, ok := result["mcp"].(map[string]interface{})
		if !ok {
			t.Fatal("written JSON missing 'mcp' key")
		}
		expectedKeys := []string{"supabase", "angular", "primeng", "postman", "context7", "playwright"}
		for _, k := range expectedKeys {
			if _, exists := mcp[k]; !exists {
				t.Errorf("missing MCP key %q in written output", k)
			}
		}
	})

	t.Run("merges entries into existing opencode.json preserving unmanaged keys", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { mcpHomeDir = origHome })

		origWrite := mcpWriteFile
		t.Cleanup(func() { mcpWriteFile = origWrite })

		// Pre-create opencode.json with unmanaged keys (theme) and a partial MCP entry
		existingCfg := `{
			"$schema": "https://opencode.ai/config.json",
			"theme": "dark",
			"mcp": {
				"supabase": {"type": "local", "command": ["npx", "-y", "@supabase/mcp-server-supabase@latest"], "enabled": true},
				"custom-server": {"type": "local", "command": ["node", "server.js"], "enabled": true}
			}
		}`
		opencodeDir := filepath.Join(tmpDir, ".config", "opencode")
		opencodeFile := filepath.Join(opencodeDir, "opencode.json")
		if err := os.MkdirAll(opencodeDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(opencodeFile, []byte(existingCfg), 0644); err != nil {
			t.Fatal(err)
		}

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := McpConfigModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Read back and verify
		data, err := os.ReadFile(opencodeFile)
		if err != nil {
			t.Fatalf("read opencode.json: %v", err)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// Unmanaged keys preserved
		if theme, ok := result["theme"]; !ok || theme != "dark" {
			t.Errorf("unmanaged key 'theme' should be preserved, got %v", result["theme"])
		}

		// Custom MCP server preserved
		mcp := result["mcp"].(map[string]interface{})
		if _, exists := mcp["custom-server"]; !exists {
			t.Error("custom MCP server should be preserved")
		}

		// All 6 managed keys added
		expectedKeys := []string{"supabase", "angular", "primeng", "postman", "context7", "playwright"}
		for _, k := range expectedKeys {
			if _, exists := mcp[k]; !exists {
				t.Errorf("missing managed MCP key %q after merge", k)
			}
		}
	})

	t.Run("creates ~/.config/opencode directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { mcpHomeDir = origHome })

		origWrite := mcpWriteFile
		t.Cleanup(func() { mcpWriteFile = origWrite })
		mcpWriteFile = os.WriteFile

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := McpConfigModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		opencodeDir := filepath.Join(tmpDir, ".config", "opencode")
		if _, err := os.Stat(opencodeDir); os.IsNotExist(err) {
			t.Error("expected ~/.config/opencode to be created")
		}
	})

	t.Run("malformed existing JSON returns error without overwriting", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := mcpHomeDir
		mcpHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { mcpHomeDir = origHome })

		origWrite := mcpWriteFile
		t.Cleanup(func() { mcpWriteFile = origWrite })
		mcpWriteFile = os.WriteFile

		// Write malformed JSON
		opencodeDir := filepath.Join(tmpDir, ".config", "opencode")
		opencodeFile := filepath.Join(opencodeDir, "opencode.json")
		if err := os.MkdirAll(opencodeDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(opencodeFile, []byte("{invalid json}"), 0644); err != nil {
			t.Fatal(err)
		}

		// Read the original content before Install
		origContent, _ := os.ReadFile(opencodeFile)

		ctx := types.InstallContext{
			Cancel:           context.Background(),
			Log:              &bytes.Buffer{},
			SessionTimestamp: "test-ts",
		}
		m := McpConfigModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}

		// Verify original content was not modified
		afterContent, _ := os.ReadFile(opencodeFile)
		if string(afterContent) != string(origContent) {
			t.Error("original malformed file should not be modified on error")
		}
	})
}

// --- AuditInfo ---

func TestMcpConfig_AuditInfo(t *testing.T) {
	t.Run("returns config", func(t *testing.T) {
		m := McpConfigModule{}
		got := m.AuditInfo()
		if got != "config" {
			t.Errorf("AuditInfo() = %q, want %q", got, "config")
		}
	})
}

// --- Golden file ---

func TestMcpConfig_GoldenMerge(t *testing.T) {
	t.Skip("golden file test — enable after mcp_config.go is running")
}

// --- helpers ---

// writeTestJSON writes a map as JSON to ~/.config/opencode/opencode.json under tmpDir.
func writeTestJSON(t *testing.T, tmpDir string, v interface{}) {
	t.Helper()
	dir := filepath.Join(tmpDir, ".config", "opencode")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}
