package tui

import (
	"io/fs"

	tea "charm.land/bubbletea/v2"
	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
	"github.com/davichuder/MyDots/internal/tui/screens"
)

// Dependencies provides the deterministic environment required by the root TUI.
type Dependencies struct {
	Platform    platform.Platform
	Assets      fs.FS
	ConfigStore screens.ConfigStore
	ConfigPath  string
	// InstallFactory defers construction of the installation screen until the user explicitly selects Install.
	InstallFactory func(config.Config) tea.Model
}
