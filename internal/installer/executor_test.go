package installer

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/davichuder/MyDots/internal/platform"
)

// testModule is a controllable stub for executor tests.
type testModule struct {
	id          ModuleID
	name        string
	criticality Criticality
	deps        []ModuleID
	installed   bool
	installErr  error
}

func (m *testModule) ID() ModuleID                       { return m.id }
func (m *testModule) Name() string                       { return m.name }
func (m *testModule) Criticality() Criticality            { return m.criticality }
func (m *testModule) Dependencies() []ModuleID            { return m.deps }
func (m *testModule) IsInstalled(platform.Platform) bool  { return m.installed }
func (m *testModule) Install(InstallContext) error        { return m.installErr }
func (m *testModule) AuditInfo() string                   { return "1.0.0" }

func mkPlan(modules ...Module) []Module {
	return modules
}

func defaultCtx() InstallContext {
	return InstallContext{
		Platform: platform.Platform{OS: platform.Darwin},
		Log:      io.Discard,
		Cancel:   context.Background(),
	}
}

func TestRun_SuccessPath(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	modA := &testModule{id: "M-01", name: "A", criticality: Critical}
	modB := &testModule{id: "M-02", name: "B", criticality: NonCritical}
	modC := &testModule{id: "M-03", name: "C", criticality: NonCritical}

	plan := mkPlan(modA, modB, modC)
	ctx := defaultCtx()
	ch := make(chan ProgressEvent, 50)

	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for evt := range ch {
		events = append(events, evt)
	}

	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
	if events[0].ModuleID != "M-01" || events[0].Status != StatusInstalled {
		t.Errorf("events[0] = %s/%s, want M-01 installed", events[0].ModuleID, events[0].Status)
	}
	if events[1].ModuleID != "M-02" || events[1].Status != StatusInstalled {
		t.Errorf("events[1] = %s/%s, want M-02 installed", events[1].ModuleID, events[1].Status)
	}
	if events[2].ModuleID != "M-03" || events[2].Status != StatusInstalled {
		t.Errorf("events[2] = %s/%s, want M-03 installed", events[2].ModuleID, events[2].Status)
	}
}

func TestRun_CriticalFailure(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	errBoom := errors.New("critical failure")
	modA := &testModule{id: "M-01", name: "A", criticality: Critical, installErr: errBoom}
	modB := &testModule{id: "M-02", name: "B", criticality: NonCritical}

	plan := mkPlan(modA, modB)
	ctx := defaultCtx()
	ch := make(chan ProgressEvent, 50)

	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for evt := range ch {
		events = append(events, evt)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event on critical failure, got %d", len(events))
	}
	if events[0].ModuleID != "M-01" || events[0].Status != StatusFailed {
		t.Errorf("events[0] = %s/%s, want M-01 failed", events[0].ModuleID, events[0].Status)
	}
	if events[0].Err != errBoom {
		t.Errorf("events[0].Err = %v, want %v", events[0].Err, errBoom)
	}
}

func TestRun_NonCriticalFailure(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	modA := &testModule{id: "M-01", name: "A", criticality: NonCritical, installErr: errors.New("non-critical fail")}
	modB := &testModule{id: "M-02", name: "B", criticality: NonCritical}

	plan := mkPlan(modA, modB)
	ctx := defaultCtx()
	ch := make(chan ProgressEvent, 50)

	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for evt := range ch {
		events = append(events, evt)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
	if events[0].ModuleID != "M-01" || events[0].Status != StatusFailed {
		t.Errorf("events[0] = %s/%s, want M-01 failed", events[0].ModuleID, events[0].Status)
	}
	if events[1].ModuleID != "M-02" || events[1].Status != StatusInstalled {
		t.Errorf("events[1] = %s/%s, want M-02 installed", events[1].ModuleID, events[1].Status)
	}
}

func TestRun_DependencyFailure(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	errBoom := errors.New("dep failure")
	modA := &testModule{id: "M-01", name: "A", criticality: NonCritical, installErr: errBoom}
	modB := &testModule{id: "M-02", name: "B", criticality: NonCritical, deps: []ModuleID{"M-01"}}

	plan := mkPlan(modA, modB)
	ctx := defaultCtx()
	ch := make(chan ProgressEvent, 50)

	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for evt := range ch {
		events = append(events, evt)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
	if events[0].ModuleID != "M-01" || events[0].Status != StatusFailed {
		t.Errorf("events[0] = %s/%s, want M-01 failed", events[0].ModuleID, events[0].Status)
	}
	if events[1].ModuleID != "M-02" || events[1].Status != StatusSkippedDependencyFailed {
		t.Errorf("events[1] = %s/%s, want M-02 dependency-skipped", events[1].ModuleID, events[1].Status)
	}
}

func TestRun_AlreadyInstalled(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	modA := &testModule{id: "M-01", name: "A", criticality: Critical, installed: true}
	modB := &testModule{id: "M-02", name: "B", criticality: NonCritical}

	plan := mkPlan(modA, modB)
	ctx := defaultCtx()
	ch := make(chan ProgressEvent, 50)

	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for evt := range ch {
		events = append(events, evt)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
	if events[0].ModuleID != "M-01" || events[0].Status != StatusSkipped {
		t.Errorf("events[0] = %s/%s, want M-01 skipped", events[0].ModuleID, events[0].Status)
	}
	if events[1].ModuleID != "M-02" || events[1].Status != StatusInstalled {
		t.Errorf("events[1] = %s/%s, want M-02 installed", events[1].ModuleID, events[1].Status)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	modA := &testModule{id: "M-01", name: "A", criticality: Critical}
	plan := mkPlan(modA)
	ctx := InstallContext{
		Platform: platform.Platform{OS: platform.Darwin},
		Log:      io.Discard,
		Cancel:   cancelledCtx,
	}
	ch := make(chan ProgressEvent, 50)

	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for evt := range ch {
		events = append(events, evt)
	}

	if len(events) != 0 {
		t.Errorf("expected 0 events on cancelled context, got %d", len(events))
	}
}

func TestRun_ChannelAlwaysClosed(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	t.Run("after critical failure", func(t *testing.T) {
		mod := &testModule{id: "M-01", name: "A", criticality: Critical, installErr: errors.New("fail")}
		plan := mkPlan(mod)
		ctx := defaultCtx()
		ch := make(chan ProgressEvent, 50)

		go Run(plan, ctx, ch)

		var events []ProgressEvent
		for evt := range ch {
			events = append(events, evt)
		}
		_ = events
	})

	t.Run("after success", func(t *testing.T) {
		mod := &testModule{id: "M-01", name: "A", criticality: NonCritical}
		plan := mkPlan(mod)
		ctx := defaultCtx()
		ch := make(chan ProgressEvent, 50)

		go Run(plan, ctx, ch)

		var events []ProgressEvent
		for evt := range ch {
			events = append(events, evt)
		}
		_ = events
	})

	t.Run("on cancelled context", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		mod := &testModule{id: "M-01", name: "A", criticality: Critical}
		plan := mkPlan(mod)
		ctx := InstallContext{
			Platform: platform.Platform{OS: platform.Darwin},
			Log:      io.Discard,
			Cancel:   cancelledCtx,
		}
		ch := make(chan ProgressEvent, 50)

		go Run(plan, ctx, ch)

		var events []ProgressEvent
		for evt := range ch {
			events = append(events, evt)
		}
		_ = events
	})
}

func TestRun_MixedPlan(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// M-01 (Critical) already installed → skipped
	// M-02 (NonCritical) fails → continues
	// M-03 depends on M-02 → dep-skipped
	// M-04 (NonCritical) succeeds
	modA := &testModule{id: "M-01", name: "A", criticality: Critical, installed: true}
	modB := &testModule{id: "M-02", name: "B", criticality: NonCritical, installErr: errors.New("non-critical fail")}
	modC := &testModule{id: "M-03", name: "C", criticality: NonCritical, deps: []ModuleID{"M-02"}}
	modD := &testModule{id: "M-04", name: "D", criticality: NonCritical}

	plan := mkPlan(modA, modB, modC, modD)
	ctx := defaultCtx()
	ch := make(chan ProgressEvent, 50)

	go Run(plan, ctx, ch)

	var events []ProgressEvent
	for evt := range ch {
		events = append(events, evt)
	}

	if len(events) != 4 {
		t.Errorf("expected 4 events, got %d", len(events))
	}
	if events[0].ModuleID != "M-01" || events[0].Status != StatusSkipped {
		t.Errorf("events[0] = %s/%s, want M-01 skipped", events[0].ModuleID, events[0].Status)
	}
	if events[1].ModuleID != "M-02" || events[1].Status != StatusFailed {
		t.Errorf("events[1] = %s/%s, want M-02 failed", events[1].ModuleID, events[1].Status)
	}
	if events[2].ModuleID != "M-03" || events[2].Status != StatusSkippedDependencyFailed {
		t.Errorf("events[2] = %s/%s, want M-03 dep-skipped", events[2].ModuleID, events[2].Status)
	}
	if events[3].ModuleID != "M-04" || events[3].Status != StatusInstalled {
		t.Errorf("events[3] = %s/%s, want M-04 installed", events[3].ModuleID, events[3].Status)
	}
}
