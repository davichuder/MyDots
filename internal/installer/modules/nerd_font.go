package modules

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davichuder/MyDots/internal/backup"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// Injectable dependencies for testing
// ---------------------------------------------------------------------------

var nerdFontHomeDir = os.UserHomeDir
var nerdFontReadFile = os.ReadFile
var nerdFontWriteFile = os.WriteFile
var nerdFontBackupFile = backup.BackupFile

// nerdFontRunScript wraps runner.Script and is injectable for tests.
var nerdFontRunScript = runner.Script

// ---------------------------------------------------------------------------
// Font name maps
// ---------------------------------------------------------------------------

// nerdFontCaskName maps config font choices to brew cask names on Darwin.
var nerdFontCaskName = map[config.FontChoice]string{
	config.FontJetBrainsMono: "font-jetbrains-mono-nerd-font",
	config.FontFiraCode:      "font-fira-code-nerd-font",
	config.FontCascadiaCode:  "font-caskaydia-cove-nerd-font",
	config.FontHack:          "font-hack-nerd-font",
	config.FontIosevka:       "font-iosevka-nerd-font",
}

// nerdFontZipName maps config font choices to GitHub release zip filenames.
var nerdFontZipName = map[config.FontChoice]string{
	config.FontJetBrainsMono: "JetBrainsMono.zip",
	config.FontFiraCode:      "FiraCode.zip",
	config.FontCascadiaCode:  "CascadiaCode.zip",
	config.FontHack:          "Hack.zip",
	config.FontIosevka:       "Iosevka.zip",
}

// ---------------------------------------------------------------------------
// NerdFontModule (M-47)
// ---------------------------------------------------------------------------

// NerdFontModule implements the Nerd Font installation (M-47) for MyDots.
// On Darwin it installs via brew cask. On Linux it runs the embedded
// font-linux.sh script. Post-install, it updates Ghostty config with
// an idempotent font-family marker block.
type NerdFontModule struct{}

// Compile-time interface check.
var _ types.Module = NerdFontModule{}

// NerdFont is the exported package-level instance used by the catalogue.
var NerdFont types.Module = NerdFontModule{}

// ID returns the stable module identifier M-47.
func (m NerdFontModule) ID() types.ModuleID { return types.ModNerdFont }

// Name returns a human-readable module name.
func (m NerdFontModule) Name() string { return "Nerd Font" }

// Criticality returns NonCritical — font failure does not stop the pipeline.
func (m NerdFontModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns no dependencies — Nerd Font can install independently.
func (m NerdFontModule) Dependencies() []types.ModuleID { return nil }

// IsInstalled checks whether a Nerd Font is already installed on the system.
//   - Darwin: checks if any nerd font cask appears in `brew list --cask`
//   - Linux: checks if `fc-list` output contains "Nerd Font"
func (m NerdFontModule) IsInstalled(p platform.Platform) bool {
	if p.OS == platform.Darwin {
		out := runner.CaptureOutput("brew", "list", "--cask")
		return strings.Contains(out, "nerd-font")
	}

	// Linux — check fontconfig for Nerd Font.
	out := runner.CaptureOutput("fc-list")
	return strings.Contains(out, "Nerd Font")
}

// Install installs the selected Nerd Font and updates Ghostty config.
func (m NerdFontModule) Install(ctx types.InstallContext) error {
	fontName := ctx.Config.Font

	if ctx.Platform.OS == platform.Darwin {
		// Darwin: brew cask.
		cask, ok := nerdFontCaskName[fontName]
		if !ok {
			return fmt.Errorf("unknown font %q for darwin cask", fontName)
		}
		if err := runner.BrewCask(ctx.Cancel, ctx.Log, ctx.Platform, cask); err != nil {
			return err
		}
	} else {
		// Linux: run font-linux.sh script.
		zipName, ok := nerdFontZipName[fontName]
		if !ok {
			return fmt.Errorf("unknown font %q for linux zip", fontName)
		}
		env := map[string]string{"FONT_NAME": zipName}
		if err := nerdFontRunScript(ctx.Cancel, ctx.Log, ctx.Assets, "assets/scripts/font-linux.sh", env); err != nil {
			return err
		}
	}

	// Post-install: update Ghostty config with font-family marker.
	if err := m.updateGhosttyFont(ctx); err != nil {
		return fmt.Errorf("update ghostty font config: %w", err)
	}

	return nil
}

// updateGhosttyFont writes font-family to ghostty config using idempotent
// MYDOTS_FONT_START / MYDOTS_FONT_END markers.
func (m NerdFontModule) updateGhosttyFont(ctx types.InstallContext) error {
	home, err := nerdFontHomeDir()
	if err != nil {
		return err
	}

	fontDisplayName := string(ctx.Config.Font) + " Nerd Font"
	ghosttyConfig := filepath.Join(home, ".config", "ghostty", "config")

	startMarker := "# MYDOTS_FONT_START"
	endMarker := "# MYDOTS_FONT_END"
	block := startMarker + "\nfont-family = " + fontDisplayName + "\n" + endMarker + "\n"

	data, readErr := nerdFontReadFile(ghosttyConfig)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil // skip if ghostty not configured yet
		}
		return fmt.Errorf("read ghostty config: %w", readErr)
	}

	// Check for existing font block.
	startIdx := bytes.Index(data, []byte(startMarker))
	if startIdx >= 0 {
		afterStart := data[startIdx+len(startMarker):]
		endIdx := bytes.Index(afterStart, []byte(endMarker))
		if endIdx >= 0 {
			endIdx += startIdx + len(startMarker) + len(endMarker)
			existingBlock := string(data[startIdx:endIdx])

			if strings.TrimSpace(existingBlock) == strings.TrimSpace(block) {
				return nil // unchanged
			}

			// Font changed — replace block.
			if err := nerdFontBackupFile(ghosttyConfig, ctx.SessionTimestamp); err != nil {
				return fmt.Errorf("backup ghostty config: %w", err)
			}

			newData := make([]byte, 0, len(data)-endIdx+startIdx+len(block))
			newData = append(newData, data[:startIdx]...)
			newData = append(newData, []byte(block)...)
			newData = append(newData, data[endIdx:]...)
			return nerdFontWriteFile(ghosttyConfig, newData, 0644)
		}
	}

	// No existing font block — append.
	if err := nerdFontBackupFile(ghosttyConfig, ctx.SessionTimestamp); err != nil {
		return fmt.Errorf("backup ghostty config: %w", err)
	}

	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, '\n')
	}
	data = append(data, []byte(block)...)
	return nerdFontWriteFile(ghosttyConfig, data, 0644)
}

// AuditInfo returns the installed font name by reading ghostty config.
func (m NerdFontModule) AuditInfo() string {
	home, err := nerdFontHomeDir()
	if err != nil {
		return ""
	}

	ghosttyConfig := filepath.Join(home, ".config", "ghostty", "config")
	data, err := nerdFontReadFile(ghosttyConfig)
	if err != nil {
		return ""
	}

	startMarker := "# MYDOTS_FONT_START"
	startIdx := bytes.Index(data, []byte(startMarker))
	if startIdx < 0 {
		return ""
	}

	afterStart := data[startIdx+len(startMarker):]
	// Find first newline after marker.
	newlineIdx := bytes.Index(afterStart, []byte("\n"))
	if newlineIdx < 0 {
		return ""
	}

	fontLine := strings.TrimSpace(string(afterStart[1:newlineIdx]))
	// Extract value after "font-family = "
	if parts := strings.SplitN(fontLine, "=", 2); len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
