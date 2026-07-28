package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// Injectable dependencies for testing
// ---------------------------------------------------------------------------

var mcpHomeDir = os.UserHomeDir
var mcpReadFile = os.ReadFile
var mcpWriteFile = os.WriteFile
var mcpMkdirAll = os.MkdirAll

// ---------------------------------------------------------------------------
// MCP server definitions
// ---------------------------------------------------------------------------

// mcpServerEntry represents a single MCP server in opencode.json.
type mcpServerEntry struct {
	Type    string   `json:"type"`
	Command []string `json:"command,omitempty"`
	URL     string   `json:"url,omitempty"`
	Enabled bool     `json:"enabled"`
}

// defaultMcpServers defines the 6 managed MCP server entries written by MyDots.
// Context7 is remote; all others are local (npx-based).
var defaultMcpServers = map[string]mcpServerEntry{
	"supabase": {
		Type:    "local",
		Command: []string{"npx", "-y", "@supabase/mcp-server-supabase@latest"},
		Enabled: true,
	},
	"angular": {
		Type:    "local",
		Command: []string{"npx", "-y", "@angular/mcp@latest"},
		Enabled: true,
	},
	"primeng": {
		Type:    "local",
		Command: []string{"npx", "-y", "primeng-mcp@latest"},
		Enabled: true,
	},
	"postman": {
		Type:    "local",
		Command: []string{"npx", "-y", "@postman/mcp-server-local@latest"},
		Enabled: true,
	},
	"context7": {
		Type:    "remote",
		URL:     "https://mcp.context7.com/mcp",
		Enabled: true,
	},
	"playwright": {
		Type:    "local",
		Command: []string{"npx", "-y", "@playwright/mcp@latest"},
		Enabled: true,
	},
}

// managedMcpKeys lists the 6 keys managed by this module (for IsInstalled check).
var managedMcpKeys = func() []string {
	keys := make([]string, 0, len(defaultMcpServers))
	for k := range defaultMcpServers {
		keys = append(keys, k)
	}
	return keys
}()

// ---------------------------------------------------------------------------
// McpConfigModule (M-44)
// ---------------------------------------------------------------------------

// McpConfigModule implements the MCP server config (M-44) for MyDots.
// It writes or merges 6 MCP server entries into ~/.config/opencode/opencode.json.
type McpConfigModule struct{}

// Compile-time interface check.
var _ types.Module = McpConfigModule{}

// McpConfig is the exported package-level instance used by the catalogue.
var McpConfig types.Module = McpConfigModule{}

// ID returns the stable module identifier M-44.
func (m McpConfigModule) ID() types.ModuleID { return types.ModMcpConfig }

// Name returns a human-readable module name.
func (m McpConfigModule) Name() string { return "MCP Config" }

// Criticality returns NonCritical — MCP config failure does not stop the pipeline.
func (m McpConfigModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns opencode as a dependency (M-40).
func (m McpConfigModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModOpencode}
}

// IsInstalled checks whether all 6 managed MCP keys exist in the mcp object
// of ~/.config/opencode/opencode.json.
func (m McpConfigModule) IsInstalled(_ platform.Platform) bool {
	home, err := mcpHomeDir()
	if err != nil {
		return false
	}

	path := filepath.Join(home, ".config", "opencode", "opencode.json")
	data, err := mcpReadFile(path)
	if err != nil {
		return false
	}

	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}

	mcpRaw, ok := cfg["mcp"]
	if !ok {
		return false
	}

	var mcpObj map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpObj); err != nil {
		return false
	}

	for _, key := range managedMcpKeys {
		if _, exists := mcpObj[key]; !exists {
			return false
		}
	}

	return true
}

// Install writes or merges the 6 managed MCP server entries into
// ~/.config/opencode/opencode.json. Existing unmanaged keys are preserved.
func (m McpConfigModule) Install(ctx types.InstallContext) error {
	home, err := mcpHomeDir()
	if err != nil {
		return err
	}

	opencodeDir := filepath.Join(home, ".config", "opencode")
	opencodeFile := filepath.Join(opencodeDir, "opencode.json")

	// Parse existing config — if file doesn't exist, start fresh.
	var root map[string]interface{}
	data, readErr := mcpReadFile(opencodeFile)
	if readErr == nil {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("malformed opencode.json: %w", err)
		}
	} else if !os.IsNotExist(readErr) {
		return fmt.Errorf("read opencode.json: %w", readErr)
	}

	if root == nil {
		root = make(map[string]interface{})
	}

	// Merge managed MCP entries into existing mcp object.
	var mcpObj map[string]interface{}
	if existing, ok := root["mcp"]; ok {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			mcpObj = existingMap
		}
	}
	if mcpObj == nil {
		mcpObj = make(map[string]interface{})
	}

	for key, entry := range defaultMcpServers {
		mcpObj[key] = entry
	}

	root["mcp"] = mcpObj

	// Ensure directory exists.
	if err := mcpMkdirAll(opencodeDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", opencodeDir, err)
	}

	// Write with indentation.
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal opencode.json: %w", err)
	}

	return mcpWriteFile(opencodeFile, out, 0644)
}

// AuditInfo returns "config" as the version identifier for MCP config.
func (m McpConfigModule) AuditInfo() string { return "config" }
