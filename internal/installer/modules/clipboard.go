package modules

import (
	"fmt"
	"os"

	"github.com/davichuder/MyDots/internal/installer/runner"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// lookupEnv is an injectable wrapper around os.Getenv for testing.
var lookupEnv = os.Getenv

// ClipboardModule implements the Clipboard tool (M-37) installation module.
// On Darwin pbcopy is built-in, so no install is needed. On Linux it installs
// wl-clipboard (Wayland) or xclip (X11) via Homebrew. On WSL2 it requires
// an active Wayland session (WSLg); if unavailable the module is skipped.
type ClipboardModule struct{}

// Compile-time interface check.
var _ types.Module = ClipboardModule{}

// Clipboard is the exported package-level instance used by the catalogue.
var Clipboard types.Module = ClipboardModule{}

// ID returns the stable module identifier M-37.
func (m ClipboardModule) ID() types.ModuleID { return types.ModClipboard }

// Name returns a human-readable module name.
func (m ClipboardModule) Name() string { return "Clipboard" }

// Criticality returns NonCritical — clipboard failure does not stop the pipeline.
func (m ClipboardModule) Criticality() types.Criticality { return types.NonCritical }

// Dependencies returns Homebrew as a dependency.
func (m ClipboardModule) Dependencies() []types.ModuleID {
	return []types.ModuleID{types.ModHomebrew}
}

// IsInstalled checks whether the clipboard tool is available.
// On Darwin pbcopy is always built-in. On Linux it checks for wl-copy or xclip.
func (m ClipboardModule) IsInstalled(p platform.Platform) bool {
	if p.OS == platform.Darwin {
		return true
	}
	return runner.CommandExists("wl-copy") || runner.CommandExists("xclip")
}

// Install installs the clipboard tool based on platform and environment.
//   - Darwin: no-op (pbcopy is built-in)
//   - Ubuntu native + WAYLAND_DISPLAY: brew install wl-clipboard
//   - Ubuntu native + no WAYLAND_DISPLAY: brew install xclip (X11 fallback)
//   - WSL2 + WAYLAND_DISPLAY: brew install wl-clipboard
//   - WSL2 + no WAYLAND_DISPLAY: log warning, return nil (skipped)
func (m ClipboardModule) Install(ctx types.InstallContext) error {
	if ctx.Platform.OS == platform.Darwin {
		return nil // pbcopy is built-in
	}

	wayland := lookupEnv("WAYLAND_DISPLAY")
	isWSL2 := ctx.Platform.Variant == platform.WSL2

	if isWSL2 && wayland == "" {
		fmt.Fprintln(ctx.Log, "[WARN] WSL2 without Wayland (WSLg) detected — clipboard module skipped.")
		fmt.Fprintln(ctx.Log, "[WARN] Install WSLg or set WAYLAND_DISPLAY to enable clipboard support.")
		return nil
	}

	if wayland != "" {
		return runner.BrewAt(ctx.Cancel, ctx.Log, brewPath(ctx), "install", "wl-clipboard")
	}
	return runner.BrewAt(ctx.Cancel, ctx.Log, brewPath(ctx), "install", "xclip")
}

// AuditInfo returns version info about the installed clipboard tool.
// Tries wl-copy first, then xclip (Linux), or "built-in" on Darwin.
func (m ClipboardModule) AuditInfo() string {
	if v := runner.CaptureOutput("wl-copy", "--version"); v != "" {
		return v
	}
	if v := runner.CaptureOutput("xclip", "-version"); v != "" {
		return v
	}
	return "built-in"
}
