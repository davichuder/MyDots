// Package sudo handles elevation requests and keepalive goroutines
// so install pipeline commands can run with sudo privileges.
package sudo

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

var runCmd = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

var keepaliveInterval = 45 * time.Second

func RequestElevation() error {
	cmd := exec.Command("sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return runCmd(cmd)
}

// Keepalive owns a running sudo refresh loop. Stop is idempotent and waits for
// its goroutine to finish, so callers can safely return only after cleanup.
type Keepalive interface{ Stop() }

type keepalive struct {
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

func StartKeepalive(ctx context.Context) (Keepalive, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	keepaliveContext, cancel := context.WithCancel(ctx)
	handle := &keepalive{cancel: cancel, done: make(chan struct{})}
	go func() {
		ticker := time.NewTicker(keepaliveInterval)
		defer ticker.Stop()
		defer close(handle.done)
		for {
			select {
			case <-keepaliveContext.Done():
				return
			case <-ticker.C:
				cmd := exec.Command("sudo", "-v")
				cmd.Stdout = io.Discard
				cmd.Stderr = io.Discard
				_ = runCmd(cmd)
			}
		}
	}()
	return handle, nil
}

func (handle *keepalive) Stop() {
	handle.once.Do(func() {
		handle.cancel()
		<-handle.done
	})
}
