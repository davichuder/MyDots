package installer

// ProgressEvent is sent through the install channel to report module progress.
// The TUI install screen reads these events to update the progress display.
type ProgressEvent struct {
	ModuleID ModuleID
	Status   InstallStatus
	LogLine  string
	Err      error
}

// Run executes the install plan, sending progress events to ch for each module.
// The channel is closed when all modules have been processed or when a critical
// module fails.
//
// Callers must create the channel and read from it until it closes:
//
//	ch := make(chan ProgressEvent, bufferSize)
//	go Run(plan, ctx, ch)
//	for e := range ch { /* handle event */ }
func Run(plan []Module, ctx InstallContext, ch chan ProgressEvent) {
	defer close(ch)

	failedIDs := map[ModuleID]bool{}

	for _, mod := range plan {
		if err := runOne(mod, ctx, ch, failedIDs); err != nil {
			// Only critical modules return a non-nil error.
			return
		}
	}
}

// runOne handles a single module: dependency check, idempotence check,
// and installation. It returns a non-nil error only when a critical
// module fails, which signals Run() to stop the pipeline.
//
// Audit persistence is deferred until the pipeline is fully wired.
// When ready, add: audit.Append(audit.DefaultPath(), entry)
func runOne(mod Module, ctx InstallContext, ch chan ProgressEvent, failedIDs map[ModuleID]bool) error {
	// Dependency check — skip if a dependency has failed.
	// Failed dependencies propagate transitively through failedIDs:
	// if A fails, and B depends on A, B is added to failedIDs so
	// that C (which depends on B) is also skipped.
	for _, dep := range mod.Dependencies() {
		if failedIDs[dep] {
			ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusSkippedDependencyFailed}
			failedIDs[mod.ID()] = true
			return nil
		}
	}

	// Idempotence check — configuration-aware modules must match the current
	// desired state rather than merely any prior installation.
	installed := mod.IsInstalled(ctx.Platform)
	if configured, ok := mod.(ConfiguredStateModule); ok {
		installed = configured.IsInstalledForConfig(ctx.Platform, ctx.Config)
	}
	if installed {
		ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusSkipped}
		return nil
	}

	// Install.
	ch <- ProgressEvent{ModuleID: mod.ID(), Status: "running"}
	if err := mod.Install(ctx); err != nil {
		failedIDs[mod.ID()] = true
		ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusFailed, Err: err}
		if mod.Criticality() == Critical {
			return err
		}
		return nil
	}

	ch <- ProgressEvent{ModuleID: mod.ID(), Status: StatusInstalled}
	return nil
}
