// Package sudo handles elevation requests and keepalive goroutines
// so install pipeline commands can run with sudo privileges.
package sudo

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

var ErrShutdownTimeout = errors.New("sudo keepalive shutdown timed out")

var runCmd = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

var keepaliveInterval = 45 * time.Second

func RequestElevation(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return runCmd(cmd)
}

// Keepalive owns a running sudo refresh loop. Stop cancels the loop and joins it
// synchronously when its caller supplies a live context.
type Keepalive interface {
	Stop(context.Context) error
	Done() <-chan struct{}
	Err() error
}

type keepalive struct {
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
	mu     sync.RWMutex
	err    error
}

type refreshRunner func(context.Context) error

func StartKeepalive(ctx context.Context) (Keepalive, error) {
	ticker := time.NewTicker(keepaliveInterval)
	return startKeepalive(ctx, ticker.C, ticker.Stop, refreshSudo)
}

func startKeepaliveWith(ctx context.Context, refresh <-chan time.Time, runRefresh refreshRunner) (Keepalive, error) {
	return startKeepalive(ctx, refresh, func() {}, runRefresh)
}

func startKeepalive(ctx context.Context, refresh <-chan time.Time, stopTicker func(), runRefresh refreshRunner) (Keepalive, error) {
	if err := ctx.Err(); err != nil {
		stopTicker()
		return nil, err
	}
	keepaliveContext, cancel := context.WithCancel(ctx)
	handle := &keepalive{cancel: cancel, done: make(chan struct{})}
	go func() {
		defer stopTicker()
		defer close(handle.done)
		for {
			select {
			case <-keepaliveContext.Done():
				return
			case <-refresh:
				if err := runRefresh(keepaliveContext); err != nil {
					if keepaliveContext.Err() != nil {
						return
					}
					handle.mu.Lock()
					handle.err = err
					handle.mu.Unlock()
					cancel()
					return
				}
			}
		}
	}()
	return handle, nil
}

func refreshSudo(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sudo", "-v")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return runCmd(cmd)
}

func (handle *keepalive) Stop(ctx context.Context) error {
	handle.once.Do(handle.cancel)
	select {
	case <-handle.done:
		return nil
	default:
	}
	select {
	case <-handle.done:
		return nil
	case <-ctx.Done():
		return ErrShutdownTimeout
	}
}

func (handle *keepalive) Done() <-chan struct{} { return handle.done }
func (handle *keepalive) Err() error {
	handle.mu.RLock()
	defer handle.mu.RUnlock()
	return handle.err
}
