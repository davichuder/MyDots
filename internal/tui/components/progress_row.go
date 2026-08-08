// Package components provides reusable TUI rendering models.
package components

import "strings"

const (
	pendingIndicator   = "—"
	runningIndicator   = "⠋"
	installedIndicator = "✅"
	failedIndicator    = "❌"
	skippedIndicator   = "⏭"
)

// ProgressRow renders one module's installation status without terminal styling.
type ProgressRow struct {
	Name   string
	Status string
}

func NewProgressRow(name, status string) ProgressRow {
	return ProgressRow{Name: name, Status: status}
}

func (row ProgressRow) Indicator() string {
	switch {
	case row.Status == "running":
		return runningIndicator
	case row.Status == "installed":
		return installedIndicator
	case row.Status == "failed":
		return failedIndicator
	case strings.HasPrefix(row.Status, "skipped"):
		return skippedIndicator
	default:
		return pendingIndicator
	}
}

func (row ProgressRow) View() string {
	return row.Name + " " + row.Indicator() + "\n"
}
