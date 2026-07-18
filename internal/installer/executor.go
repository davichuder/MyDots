package installer

import (
	"os"
	"path/filepath"

	"github.com/davichuder/MyDots/internal/audit"
)

// ProgressEvent is sent to the TUI channel during execution.
// Each event represents a change in a module's install status.
type ProgressEvent struct {
	ModuleID ModuleID
	Status   InstallStatus
	LogLine  string
	Err      error
}

// Run executes the install plan, sending progress events to ch.
// It defers close(ch) so callers can range over ch.
// On critical failure or context cancellation it stops the pipeline.
func Run(plan []Module, ctx InstallContext, ch chan<- ProgressEvent) {
	defer close(ch)

	if ctx.Cancel.Err() != nil {
		return
	}

	failedIDs := map[ModuleID]bool{}
	for _, mod := range plan {
		if err := runOne(mod, ctx, ch, failedIDs); err != nil {
			return
		}
	}
}

// auditPath returns the path to the audit file.
// Tests can override by setting HOME in the environment.
func auditPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".mydots-audit.json")
}

// runOne handles a single module: dependency check, idempotence check,
// install, and audit logging. It returns an error only for Critical failures.
func runOne(mod Module, ctx InstallContext, ch chan<- ProgressEvent, failedIDs map[ModuleID]bool) error {
	auditFile := auditPath()

	// Dependency check
	for _, dep := range mod.Dependencies() {
		if failedIDs[dep] {
			evt := ProgressEvent{ModuleID: mod.ID(), Status: StatusSkippedDependencyFailed}
			ch <- evt
			_ = audit.Append(auditFile, audit.Entry{
				ModuleID: string(mod.ID()),
				Name:     mod.Name(),
				Status:   audit.InstallStatus(StatusSkippedDependencyFailed),
				OS:       string(ctx.Platform.OS),
			})
			return nil
		}
	}

	// Idempotence check
	if mod.IsInstalled(ctx.Platform) {
		evt := ProgressEvent{ModuleID: mod.ID(), Status: StatusSkipped}
		ch <- evt
		_ = audit.Append(auditFile, audit.Entry{
			ModuleID: string(mod.ID()),
			Name:     mod.Name(),
			Version:  mod.AuditInfo(),
			Status:   audit.InstallStatus(StatusSkipped),
			OS:       string(ctx.Platform.OS),
		})
		return nil
	}

	// Install
	err := mod.Install(ctx)
	if err != nil {
		failedIDs[mod.ID()] = true
		ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusFailed, Err: err}
		_ = audit.Append(auditFile, audit.Entry{
			ModuleID: string(mod.ID()),
			Name:     mod.Name(),
			Status:   audit.InstallStatus(StatusFailed),
			Error:    err.Error(),
			OS:       string(ctx.Platform.OS),
		})
		if mod.Criticality() == Critical {
			return err
		}
		return nil
	}

	ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusInstalled}
	_ = audit.Append(auditFile, audit.Entry{
		ModuleID: string(mod.ID()),
		Name:     mod.Name(),
		Version:  mod.AuditInfo(),
		Status:   audit.InstallStatus(StatusInstalled),
		OS:       string(ctx.Platform.OS),
	})
	return nil
}
