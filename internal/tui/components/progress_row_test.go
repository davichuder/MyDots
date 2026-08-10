package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProgressRowStatusIndicators(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		indicator string
	}{
		{name: "pending", status: "pending", indicator: "—"},
		{name: "running", status: "running", indicator: "⠋"},
		{name: "installed", status: "installed", indicator: "✅"},
		{name: "failed", status: "failed", indicator: "❌"},
		{name: "skipped", status: "skipped", indicator: "⏭"},
		{name: "disabled", status: "disabled", indicator: "—"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewProgressRow("Homebrew", tt.status)

			if got := row.Indicator(); got != tt.indicator {
				t.Errorf("Indicator() = %q, want %q", got, tt.indicator)
			}
			if got := row.View(); !strings.Contains(got, "Homebrew") {
				t.Errorf("View() = %q, want module name Homebrew", got)
			}
		})
	}
}

func TestProgressRowViewGoldens(t *testing.T) {
	const terminalWidth = 80
	const terminalHeight = 24

	tests := []struct {
		name   string
		status string
	}{
		{name: "pending", status: "pending"},
		{name: "running", status: "running"},
		{name: "installed", status: "installed"},
		{name: "failed", status: "failed"},
		{name: "skipped", status: "skipped"},
		{name: "disabled", status: "disabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewProgressRow("Homebrew", tt.status)
			want, err := os.ReadFile(filepath.Join("testdata", "progress_row", tt.name+".golden"))
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}

			if got := row.View(); got != string(want) {
				t.Errorf("View() at %dx%d = %q, want %q", terminalWidth, terminalHeight, got, want)
			}
		})
	}
}
