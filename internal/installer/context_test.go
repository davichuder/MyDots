package installer

import (
	"bytes"
	"context"
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/platform"
)

func TestInstallContext_ZeroValue(t *testing.T) {
	// A zero-value InstallContext must not panic when accessing any field.
	var ctx InstallContext

	_ = ctx.Platform
	_ = ctx.Config
	_ = ctx.SessionTimestamp
	_ = ctx.Log
	_ = ctx.Cancel
}

func TestInstallContext_FieldsAccessible(t *testing.T) {
	p, err := platform.Detect("darwin")
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	var buf bytes.Buffer
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx := InstallContext{
		Platform:         p,
		Config:           cfg,
		SessionTimestamp: "2026-05-23T10-30-00Z",
		Log:              &buf,
		Cancel:           cancelCtx,
	}

	if ctx.Platform.OS != platform.Darwin {
		t.Errorf("expected Platform.OS = Darwin, got: %v", ctx.Platform.OS)
	}
	if ctx.Config.Version != config.ConfigVersion {
		t.Errorf("expected Config.Version = %q, got: %q", config.ConfigVersion, ctx.Config.Version)
	}
	if ctx.SessionTimestamp != "2026-05-23T10-30-00Z" {
		t.Errorf("expected SessionTimestamp = %q, got: %q", "2026-05-23T10-30-00Z", ctx.SessionTimestamp)
	}
	if ctx.Log == nil {
		t.Error("Log should not be nil after assignment")
	}
	if ctx.Cancel == nil {
		t.Error("Cancel should not be nil after assignment")
	}
}
