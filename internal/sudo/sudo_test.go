package sudo

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

func restoreGlobals() {
	runCmd = func(cmd *exec.Cmd) error { return cmd.Run() }
	keepaliveInterval = 45 * time.Second
}

func TestRequestElevation_SetsTerminalStreams(t *testing.T) {
	var capturedStdin io.Reader
	var capturedStdout io.Writer
	var capturedStderr io.Writer

	runCmd = func(cmd *exec.Cmd) error {
		capturedStdin = cmd.Stdin
		capturedStdout = cmd.Stdout
		capturedStderr = cmd.Stderr
		return nil
	}
	defer restoreGlobals()

	err := RequestElevation(context.Background())
	if err != nil {
		t.Fatalf("RequestElevation returned error: %v", err)
	}

	if capturedStdin != os.Stdin {
		t.Error("Stdin should be os.Stdin")
	}
	if capturedStdout != os.Stdout {
		t.Error("Stdout should be os.Stdout")
	}
	if capturedStderr != os.Stderr {
		t.Error("Stderr should be os.Stderr")
	}
}

func TestRequestElevation_ExecutesSudoV(t *testing.T) {
	var capturedArgs []string
	runCmd = func(cmd *exec.Cmd) error {
		capturedArgs = cmd.Args
		return nil
	}
	defer restoreGlobals()

	_ = RequestElevation(context.Background())

	if len(capturedArgs) < 2 {
		t.Fatalf("expected at least 2 args, got: %v", capturedArgs)
	}
	if capturedArgs[0] != "sudo" {
		t.Errorf("expected command 'sudo', got: %q", capturedArgs[0])
	}
	if capturedArgs[1] != "-v" {
		t.Errorf("expected arg '-v', got: %q", capturedArgs[1])
	}
}

func TestRequestElevation_ReturnsError(t *testing.T) {
	expectedErr := errors.New("sudo not available")
	runCmd = func(*exec.Cmd) error {
		return expectedErr
	}
	defer restoreGlobals()

	err := RequestElevation(context.Background())
	if err != expectedErr {
		t.Errorf("expected error %v, got: %v", expectedErr, err)
	}
}

func TestDefaultKeepaliveInterval(t *testing.T) {
	defer restoreGlobals()

	if keepaliveInterval != 45*time.Second {
		t.Errorf("default keepaliveInterval = %v, want 45s", keepaliveInterval)
	}
}

func TestStartKeepalive_OutputToDiscard(t *testing.T) {
	var capturedStdout, capturedStderr io.Writer
	called := make(chan struct{}, 1)
	runCmd = func(cmd *exec.Cmd) error {
		capturedStdout = cmd.Stdout
		capturedStderr = cmd.Stderr
		called <- struct{}{}
		return nil
	}
	keepaliveInterval = 10 * time.Millisecond
	defer restoreGlobals()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handle, err := StartKeepalive(ctx)
	if err != nil {
		t.Fatalf("StartKeepalive returned error: %v", err)
	}
	defer func() { _ = handle.Stop(context.Background()) }()

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("keepalive did not execute sudo validation")
	}

	if capturedStdout != io.Discard {
		t.Error("sudo stdout should go to io.Discard")
	}
	if capturedStderr != io.Discard {
		t.Error("sudo stderr should go to io.Discard")
	}
}

func TestKeepaliveReportsRefreshFailureAndJoins(t *testing.T) {
	refresh := make(chan time.Time, 1)
	wantErr := errors.New("sudo refresh failed")

	handle, err := startKeepaliveWith(context.Background(), refresh, func(context.Context) error { return wantErr })
	if err != nil {
		t.Fatalf("startKeepaliveWith returned error: %v", err)
	}
	refresh <- time.Now()
	select {
	case <-handle.Done():
	case <-time.After(time.Second):
		t.Fatal("keepalive did not terminate after refresh failure")
	}
	if !errors.Is(handle.Err(), wantErr) {
		t.Errorf("keepalive error = %v, want %v", handle.Err(), wantErr)
	}
	_ = handle.Stop(context.Background())
	_ = handle.Stop(context.Background())
}

func TestKeepaliveStopCancelsInFlightRefreshWithoutReportingHealthFailure(t *testing.T) {
	refresh := make(chan time.Time, 1)
	testContext, cancelTest := context.WithTimeout(t.Context(), time.Second)
	defer cancelTest()
	refreshStarted := make(chan struct{})
	refreshCanceled := make(chan struct{})

	handle, err := startKeepaliveWith(testContext, refresh, func(ctx context.Context) error {
		close(refreshStarted)
		select {
		case <-ctx.Done():
			close(refreshCanceled)
			return ctx.Err()
		case <-testContext.Done():
			return testContext.Err()
		}
	})
	if err != nil {
		t.Fatalf("startKeepaliveWith returned error: %v", err)
	}

	refresh <- time.Now()
	select {
	case <-refreshStarted:
	case <-testContext.Done():
		t.Fatal("keepalive did not start the refresh")
	}

	stopped := make(chan struct{})
	go func() {
		_ = handle.Stop(context.Background())
		close(stopped)
	}()

	select {
	case <-refreshCanceled:
	case <-testContext.Done():
		t.Fatal("Stop did not cancel the in-flight refresh")
	}
	select {
	case <-stopped:
	case <-testContext.Done():
		t.Fatal("Stop did not join the canceled refresh")
	}
	if err := handle.Err(); err != nil {
		t.Errorf("keepalive error after Stop = %v, want nil", err)
	}

	if err := handle.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() = %v, want nil", err)
	}
}

func TestKeepaliveStopReturnsTimeoutForNonCooperativeRefresh(t *testing.T) {
	refresh := make(chan time.Time, 1)
	refreshStarted := make(chan struct{})
	releaseRefresh := make(chan struct{})
	handle, err := startKeepaliveWith(context.Background(), refresh, func(context.Context) error {
		close(refreshStarted)
		<-releaseRefresh // Deliberately violates the cancellation contract.
		return nil
	})
	if err != nil {
		t.Fatalf("startKeepaliveWith returned error: %v", err)
	}
	refresh <- time.Time{}
	<-refreshStarted

	shutdownContext, cancelShutdown := context.WithCancel(context.Background())
	cancelShutdown()
	if err := handle.Stop(shutdownContext); !errors.Is(err, ErrShutdownTimeout) {
		t.Fatalf("Stop() error = %v, want ErrShutdownTimeout", err)
	}
	close(releaseRefresh)
	if err := handle.Stop(context.Background()); err != nil {
		t.Fatalf("idempotent Stop() = %v, want nil after join", err)
	}
}

func TestStartKeepalive_StopsOnCancel(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	runCmd = func(*exec.Cmd) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}
	keepaliveInterval = 10 * time.Millisecond
	defer restoreGlobals()

	ctx, cancel := context.WithCancel(context.Background())
	handle, err := StartKeepalive(ctx)
	if err != nil {
		t.Fatalf("StartKeepalive returned error: %v", err)
	}

	time.Sleep(25 * time.Millisecond)

	mu.Lock()
	before := callCount
	mu.Unlock()

	cancel()
	_ = handle.Stop(context.Background())

	// Wait long enough that an alive goroutine would tick several times
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	after := callCount
	mu.Unlock()

	// The goroutine may tick once more after cancel due to select randomness
	// (both ticker.C and ctx.Done() ready). More than 2 extra = still running.
	if after > before+2 {
		t.Errorf("goroutine should have stopped after cancel: before=%d, after=%d", before, after)
	}
}

func TestStartKeepalive_AlreadyCancelledContext(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	runCmd = func(*exec.Cmd) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}
	keepaliveInterval = 10 * time.Millisecond
	defer restoreGlobals()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before StartKeepalive

	handle, err := StartKeepalive(ctx)
	if !errors.Is(err, context.Canceled) || handle != nil {
		t.Fatalf("StartKeepalive() = (%v, %v), want (nil, context.Canceled)", handle, err)
	}

	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count > 0 {
		t.Errorf("expected 0 sudo calls with already cancelled context, got: %d", count)
	}
}

func TestStartKeepalive_MultipleCalls(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	runCmd = func(*exec.Cmd) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}
	keepaliveInterval = 10 * time.Millisecond
	defer restoreGlobals()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	first, err := StartKeepalive(ctx)
	if err != nil {
		t.Fatalf("first StartKeepalive returned error: %v", err)
	}
	second, err := StartKeepalive(ctx)
	if err != nil {
		t.Fatalf("second StartKeepalive returned error: %v", err)
	}

	time.Sleep(35 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	// Each goroutine ticks at 10ms, 20ms, 30ms = 3 each
	// Allow some slack due to goroutine scheduling
	if count < 2 {
		t.Errorf("expected at least 2 sudo calls across 2 goroutines, got: %d", count)
	}
	cancel()
	_ = first.Stop(context.Background())
	_ = second.Stop(context.Background())
}
