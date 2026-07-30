package modules

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davichuder/MyDots/internal/backup"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// Injectable file I/O for testing
// ---------------------------------------------------------------------------

var themeHomeDir = os.UserHomeDir
var themeReadFile = os.ReadFile
var themeWriteFile = os.WriteFile
var themeBackupFile = backup.BackupFile

// ---------------------------------------------------------------------------
// Theme config targets
// ---------------------------------------------------------------------------

// themeFileTarget describes one config file that receives the theme marker block.
type themeFileTarget struct {
	// path is relative to $HOME.
	path string
	// startMarker is the comment line that opens the theme block (e.g. "# MYDOTS_THEME_START").
	startMarker string
	// endMarker is the comment line that closes the theme block.
	endMarker string
	// themeLine returns the theme directive line(s) for the given theme name.
	themeLine func(theme string) string
}

// themeTargets lists all config files managed by the Theme module (M-46).
// Each entry defines the target path, marker format, and theme directive template.
var themeTargets = []themeFileTarget{
	{
		path:        filepath.Join(".config", "nvim", "lua", "plugins", "colorscheme.lua"),
		startMarker: "-- MYDOTS_THEME_START",
		endMarker:   "-- MYDOTS_THEME_END",
		themeLine:   func(t string) string { return fmt.Sprintf("vim.cmd.colorscheme(%q)", t) },
	},
	{
		path:        ".zshrc",
		startMarker: "# MYDOTS_THEME_START",
		endMarker:   "# MYDOTS_THEME_END",
		themeLine:   func(t string) string { return fmt.Sprintf("ZSH_THEME=%q", t) },
	},
	{
		path:        filepath.Join(".config", "zellij", "config.kdl"),
		startMarker: "// MYDOTS_THEME_START",
		endMarker:   "// MYDOTS_THEME_END",
		themeLine:   func(t string) string { return fmt.Sprintf("theme %q", t) },
	},
	{
		path:        filepath.Join(".config", "ghostty", "config"),
		startMarker: "# MYDOTS_THEME_START",
		endMarker:   "# MYDOTS_THEME_END",
		themeLine:   func(t string) string { return fmt.Sprintf("theme = %s", t) },
	},
}

// ---------------------------------------------------------------------------
// ThemeModule (M-46)
// ---------------------------------------------------------------------------

// ThemeModule implements theme application (M-46) for MyDots.
// It writes the selected theme to four config files using idempotent
// # MYDOTS_THEME_START / # MYDOTS_THEME_END marker blocks.
type ThemeModule struct{}

// Compile-time interface check.
var _ types.Module = ThemeModule{}

// Theme is the exported package-level instance used by the catalogue.
var Theme types.Module = ThemeModule{}

// ID returns the stable module identifier M-46.
func (m ThemeModule) ID() types.ModuleID { return types.ModTheme }

// Name returns a human-readable module name.
func (m ThemeModule) Name() string { return "Theme" }

// Criticality returns NonCritical — theme failure does not stop the pipeline.
func (m ThemeModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns the modules that must run before theme (M-17, M-03, M-04, M-48).
func (m ThemeModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModNeovim, types.ModOhMyZsh, types.ModZellij, types.ModGhostty}
}

// IsInstalled checks the default desired theme for direct callers. The executor
// uses IsInstalledForConfig with the active session configuration.
func (m ThemeModule) IsInstalled(p platform.Platform) bool {
	return m.IsInstalledForConfig(p, config.DefaultConfig())
}

// IsInstalledForConfig checks whether every existing managed file contains the
// selected theme block. Missing targets are optional because Install skips them.
func (m ThemeModule) IsInstalledForConfig(_ platform.Platform, cfg config.Config) bool {
	home, err := themeHomeDir()
	if err != nil {
		return false
	}

	for _, target := range themeTargets {
		absPath := filepath.Join(home, target.path)
		data, err := themeReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false
		}
		block := target.startMarker + "\n" + target.themeLine(string(cfg.Theme)) + "\n" + target.endMarker
		if !bytes.Contains(data, []byte(block)) {
			return false
		}
	}

	return len(themeTargets) > 0
}

// Install writes the selected theme to all four config files.
// Each file uses idempotent MYDOTS_THEME_START / MYDOTS_THEME_END marker blocks.
// Existing files are backed up before modification.
// Missing files are skipped (not an error) so the install continues with the rest.
func (m ThemeModule) Install(ctx types.InstallContext) error {
	home, err := themeHomeDir()
	if err != nil {
		return err
	}

	themeName := string(ctx.Config.Theme)

	for _, target := range themeTargets {
		absPath := filepath.Join(home, target.path)

		// Read existing content. If file doesn't exist, skip this target
		// (the Theme module only modifies existing config files; it does
		// not create them from scratch).
		data, readErr := themeReadFile(absPath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue // skip missing files gracefully
			}
			return fmt.Errorf("read %s: %w", target.path, readErr)
		}

		// Build the new marker block.
		block := target.startMarker + "\n" + target.themeLine(themeName) + "\n" + target.endMarker + "\n"

		// Check for existing markers.
		startIdx := bytes.Index(data, []byte(target.startMarker))
		if startIdx >= 0 {
			// Find end marker after start.
			afterStart := data[startIdx+len(target.startMarker):]
			endIdx := bytes.Index(afterStart, []byte(target.endMarker))
			if endIdx >= 0 {
				// Extract the existing block (inclusive of both markers).
				endIdx += startIdx + len(target.startMarker) + len(target.endMarker)
				existingBlock := string(data[startIdx:endIdx])

				// Compare with the new block (strip trailing newlines for comparison).
				if strings.TrimSpace(existingBlock) == strings.TrimSpace(block) {
					continue // no change needed
				}

				// Theme changed — backup existing file before modifying.
				if err := themeBackupFile(absPath, ctx.SessionTimestamp); err != nil {
					return fmt.Errorf("backup %s: %w", target.path, err)
				}

				// Replace existing block with new one.
				newData := make([]byte, 0, len(data)-endIdx+startIdx+len(block))
				newData = append(newData, data[:startIdx]...)
				newData = append(newData, []byte(block)...)
				newData = append(newData, data[endIdx:]...)
				if err := themeWriteFile(absPath, newData, 0644); err != nil {
					return fmt.Errorf("write %s: %w", target.path, err)
				}
				continue
			}
		}

		// No existing markers found — first-time write.
		// Backup existing file before modifying.
		if err := themeBackupFile(absPath, ctx.SessionTimestamp); err != nil {
			return fmt.Errorf("backup %s: %w", target.path, err)
		}

		// Ensure file ends with newline before appending block.
		if !bytes.HasSuffix(data, []byte("\n")) {
			data = append(data, '\n')
		}
		data = append(data, []byte(block)...)

		if err := themeWriteFile(absPath, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", target.path, err)
		}
	}

	return nil
}

// AuditInfo returns the installed theme name.
// Since IsInstalled checks all 4 files, this returns the theme name
// by reading the first config file found and extracting the theme line.
func (m ThemeModule) AuditInfo() string {
	home, err := themeHomeDir()
	if err != nil {
		return ""
	}

	// Try each target to find an existing theme block.
	for _, target := range themeTargets {
		absPath := filepath.Join(home, target.path)
		data, readErr := themeReadFile(absPath)
		if readErr != nil {
			continue
		}

		startIdx := bytes.Index(data, []byte(target.startMarker))
		if startIdx < 0 {
			continue
		}
		startIdx += len(target.startMarker) + 1 // skip past marker + newline

		endIdx := bytes.Index(data[startIdx:], []byte(target.endMarker))
		if endIdx < 0 {
			continue
		}

		themeLine := strings.TrimSpace(string(data[startIdx : startIdx+endIdx]))
		return themeLine
	}

	return ""
}
