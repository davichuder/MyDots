package modules

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/davichuder/MyDots/internal/config"
	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

// ---------------------------------------------------------------------------
// NeovimFrameworkModule tests — T-053 (RED)
// ---------------------------------------------------------------------------

func TestNeovimFramework_ID(t *testing.T) {
	m := NeovimFrameworkModule{}
	if got := m.ID(); got != types.ModNeovimFramework {
		t.Errorf("ID() = %v, want %v", got, types.ModNeovimFramework)
	}
}

func TestNeovimFramework_Name(t *testing.T) {
	m := NeovimFrameworkModule{}
	if got := m.Name(); got != "Neovim Framework" {
		t.Errorf("Name() = %q, want %q", got, "Neovim Framework")
	}
}

func TestNeovimFramework_Criticality(t *testing.T) {
	m := NeovimFrameworkModule{}
	if got := m.Criticality(); got != types.NonCritical {
		t.Errorf("Criticality() = %v, want %v", got, types.NonCritical)
	}
}

func TestNeovimFramework_Dependencies(t *testing.T) {
	m := NeovimFrameworkModule{}
	got := m.Dependencies()
	if len(got) != 1 || got[0] != types.ModNeovim {
		t.Errorf("Dependencies() = %v, want [%v]", got, types.ModNeovim)
	}
}

// --- IsInstalled ---

func TestNeovimFramework_IsInstalled(t *testing.T) {
	t.Run("any framework dir exists returns true", func(t *testing.T) {
		tmpDir := t.TempDir()
		fwDir := filepath.Join(tmpDir, ".config", "nvim-lazyvim")
		if err := os.MkdirAll(fwDir, 0755); err != nil {
			t.Fatal(err)
		}

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		m := NeovimFrameworkModule{}
		if !m.IsInstalled(platform.Platform{}) {
			t.Error("expected true when framework dir exists")
		}
	})

	t.Run("no framework dir returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		m := NeovimFrameworkModule{}
		if m.IsInstalled(platform.Platform{}) {
			t.Error("expected false when no framework dir exists")
		}
	})
}

// --- Install ---

func TestNeovimFramework_Install(t *testing.T) {
	skipIfWindows(t)

	t.Run("clones repo and appends alias to zshrc", func(t *testing.T) {
		tmpDir := t.TempDir()

		var cloneCmd []string
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && len(args) >= 2 && args[0] == "clone" {
					cloneCmd = append(cloneCmd, args...)
					// Create target dir to simulate successful clone
					targetDir := args[len(args)-1]
					if err := os.MkdirAll(targetDir, 0755); err != nil {
						return nil, err
					}
					return nil, nil
				}
				return nil, errors.New("unexpected command: " + name)
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		origRead := zshrcReadFile
		origWrite := zshrcWriteFile
		t.Cleanup(func() {
			zshrcReadFile = origRead
			zshrcWriteFile = origWrite
		})

		var writtenData string
		zshrcReadFile = func(path string) ([]byte, error) {
			return []byte("export EDITOR=nvim\n"), nil
		}
		zshrcWriteFile = func(path string, data []byte, perm os.FileMode) error {
			writtenData = string(data)
			return nil
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Config: config.Config{
				Nvim: config.NvimOptions{
					Framework: config.NvimFrameworkLazyVim,
				},
			},
		}
		m := NeovimFrameworkModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}

		// Verify clone: git clone <repo-url> <target>
		if len(cloneCmd) < 3 {
			t.Fatalf("expected git clone command, got: %v", cloneCmd)
		}
		wantURL := "https://github.com/LazyVim/starter"
		if cloneCmd[1] != wantURL {
			t.Errorf("clone URL = %q, want %q", cloneCmd[1], wantURL)
		}

		// Verify alias appended
		wantAlias := "alias lazyvim='NVIM_APPNAME=nvim-lazyvim nvim'"
		if !bytes.Contains([]byte(writtenData), []byte(wantAlias)) {
			t.Errorf("alias not found in written zshrc.\ngot:  %q\nwant: %q", writtenData, wantAlias)
		}
	})

	t.Run("triangulation — all 5 frameworks map to correct URL", func(t *testing.T) {
		cases := []struct {
			framework config.NvimFramework
			wantURL   string
			wantAlias string
		}{
			{config.NvimFrameworkLazyVim, "https://github.com/LazyVim/starter", "alias lazyvim="},
			{config.NvimFrameworkLunarVim, "https://github.com/LunarVim/Launch.nvim", "alias lunarvim="},
			{config.NvimFrameworkAstroNvim, "https://github.com/AstroNvim/template", "alias astronvim="},
			{config.NvimFrameworkNvChad, "https://github.com/NvChad/starter", "alias nvchad="},
			{config.NvimFrameworkLaunchVim, "https://github.com/LaunchVim/launch.nvim", "alias launchnvim="},
		}

		for _, tc := range cases {
			t.Run(string(tc.framework), func(t *testing.T) {
				tmpDir := t.TempDir()

				var cloneArgs []string
				withMockExecutor(t, &mockExecutor{
					executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
						if name == "git" && len(args) >= 2 && args[0] == "clone" {
							cloneArgs = args
							target := args[len(args)-1]
							os.MkdirAll(target, 0755)
							return nil, nil
						}
						return nil, nil
					},
				})

				origHome := nvimHomeDir
				nvimHomeDir = func() (string, error) { return tmpDir, nil }
				t.Cleanup(func() { nvimHomeDir = origHome })

				origRead := zshrcReadFile
				origWrite := zshrcWriteFile
				t.Cleanup(func() {
					zshrcReadFile = origRead
					zshrcWriteFile = origWrite
				})

				zshrcReadFile = func(path string) ([]byte, error) {
					return []byte(""), nil
				}
				var written string
				zshrcWriteFile = func(path string, data []byte, perm os.FileMode) error {
					written = string(data)
					return nil
				}

				ctx := types.InstallContext{
					Cancel: context.Background(),
					Log:    &bytes.Buffer{},
					Config: config.Config{
						Nvim: config.NvimOptions{
							Framework: tc.framework,
						},
					},
				}
				m := NeovimFrameworkModule{}
				if err := m.Install(ctx); err != nil {
					t.Fatalf("Install(%s) = %v, want nil", tc.framework, err)
				}

				if len(cloneArgs) < 2 || cloneArgs[1] != tc.wantURL {
					t.Errorf("clone URL = %q, want %q", cloneArgs[1], tc.wantURL)
				}
				if !bytes.Contains([]byte(written), []byte(tc.wantAlias)) {
					t.Errorf("alias not found in zshrc.\ngot: %q\nwant contains: %q", written, tc.wantAlias)
				}
			})
		}
	})

	t.Run("alias already in zshrc is not duplicated", func(t *testing.T) {
		tmpDir := t.TempDir()
		existingAlias := "alias lazyvim='NVIM_APPNAME=nvim-lazyvim nvim'"

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && len(args) >= 2 && args[0] == "clone" {
					target := args[len(args)-1]
					os.MkdirAll(target, 0755)
					return nil, nil
				}
				return nil, nil
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		origRead := zshrcReadFile
		origWrite := zshrcWriteFile
		t.Cleanup(func() {
			zshrcReadFile = origRead
			zshrcWriteFile = origWrite
		})

		zshrcReadFile = func(path string) ([]byte, error) {
			return []byte(existingAlias + "\n"), nil
		}
		var writeCalled bool
		zshrcWriteFile = func(path string, data []byte, perm os.FileMode) error {
			writeCalled = true
			return nil
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Config: config.Config{
				Nvim: config.NvimOptions{
					Framework: config.NvimFrameworkLazyVim,
				},
			},
		}
		m := NeovimFrameworkModule{}
		if err := m.Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if writeCalled {
			t.Error("zshrc should not be written when alias already exists")
		}
	})

	t.Run("git clone failure cleans up target dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, ".config", "nvim-lazyvim")

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" {
					// Create partial dir to simulate failed clone
					os.MkdirAll(targetDir, 0755)
					return nil, errors.New("clone failed")
				}
				return nil, nil
			},
		})

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = origHome })

		origRead := zshrcReadFile
		origWrite := zshrcWriteFile
		t.Cleanup(func() {
			zshrcReadFile = origRead
			zshrcWriteFile = origWrite
		})
		zshrcReadFile = func(path string) ([]byte, error) {
			return []byte(""), nil
		}
		zshrcWriteFile = func(path string, data []byte, perm os.FileMode) error {
			return nil
		}

		ctx := types.InstallContext{
			Cancel: context.Background(),
			Log:    &bytes.Buffer{},
			Config: config.Config{
				Nvim: config.NvimOptions{
					Framework: config.NvimFrameworkLazyVim,
				},
			},
		}
		m := NeovimFrameworkModule{}
		if err := m.Install(ctx); err == nil {
			t.Fatal("expected error, got nil")
		}

		// Verify cleanup
		if _, err := os.Stat(targetDir); err == nil {
			t.Error("target dir should be removed after clone failure")
		}
	})
}

// --- AuditInfo ---

func TestNeovimFramework_AuditInfo(t *testing.T) {
	m := NeovimFrameworkModule{}
	got := m.AuditInfo()
	if got == "" {
		t.Error("AuditInfo() should not be empty")
	}
}
