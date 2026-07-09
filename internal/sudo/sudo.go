package sudo

import (
	"context"
	"io"
	"os"
	"os/exec"
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

func StartKeepalive(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(keepaliveInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cmd := exec.Command("sudo", "-v")
				cmd.Stdout = io.Discard
				cmd.Stderr = io.Discard
				_ = runCmd(cmd)
			}
		}
	}()
}
