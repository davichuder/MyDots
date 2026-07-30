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
	t.Run("framework directory and managed alias return true", func(t *testing.T) {
		tmpDir := t.TempDir()
		fwDir := filepath.Join(tmpDir, ".config", "nvim-lazyvim")
		createFrameworkClone(t, fwDir)

		origHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		origRead := zshrcReadFile
		zshrcReadFile = os.ReadFile
		t.Cleanup(func() { nvimHomeDir, zshrcReadFile = origHome, origRead })
		mustWriteFile(t, filepath.Join(tmpDir, ".zshrc"), frameworkAlias(config.NvimFrameworkLazyVim)+"\n")

		m := NeovimFrameworkModule{}
		if !m.IsInstalledForConfig(platform.Platform{}, config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}) {
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

	origBackup := zshrcBackupFile
	zshrcBackupFile = func(string, string) error { return nil }
	t.Cleanup(func() { zshrcBackupFile = origBackup })

	t.Run("clones repo and appends alias to zshrc", func(t *testing.T) {
		tmpDir := t.TempDir()

		var cloneCmd []string
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && len(args) >= 2 && args[0] == "clone" {
					cloneCmd = append(cloneCmd, args...)
					// Create target dir to simulate successful clone
					targetDir := args[len(args)-1]
					createFrameworkClone(t, targetDir)
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
							createFrameworkClone(t, target)
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
					createFrameworkClone(t, target)
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

	t.Run("complete existing clone is idempotent", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, ".config", "nvim-lazyvim")
		createFrameworkClone(t, targetDir)

		cloneCalled := false
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, _ ...string) ([]byte, error) {
				cloneCalled = true
				return nil, errors.New("unexpected command: " + name)
			},
		})

		originalHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = originalHome })

		originalRead, originalWrite := zshrcReadFile, zshrcWriteFile
		zshrcReadFile = func(string) ([]byte, error) {
			return []byte(frameworkAlias(config.NvimFrameworkLazyVim) + "\n"), nil
		}
		writeCalled := false
		zshrcWriteFile = func(string, []byte, os.FileMode) error {
			writeCalled = true
			return nil
		}
		t.Cleanup(func() { zshrcReadFile, zshrcWriteFile = originalRead, originalWrite })

		ctx := types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, Config: config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}}
		if err := (NeovimFrameworkModule{}).Install(ctx); err != nil {
			t.Fatalf("Install() = %v, want nil", err)
		}
		if cloneCalled {
			t.Error("Install() attempted to clone an existing complete clone")
		}
		if writeCalled {
			t.Error("Install() rewrote an existing managed alias")
		}
	})

	t.Run("git clone failure cleans up target dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, ".config", "nvim-lazyvim")
		cloneAttempts := 0

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" {
					cloneAttempts++
					// Create partial dir to simulate failed clone
					if err := os.MkdirAll(targetDir, 0755); err != nil {
						return nil, err
					}
					if cloneAttempts == 1 {
						return nil, errors.New("clone failed")
					}
					createFrameworkClone(t, targetDir)
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
		if err := m.Install(ctx); err != nil {
			t.Fatalf("retry Install() = %v, want nil", err)
		}
		if cloneAttempts != 2 {
			t.Fatalf("clone attempts after retry = %d, want 2", cloneAttempts)
		}
	})

	t.Run("clone failure cleanup joins both errors and leaves retry state explicit", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, ".config", "nvim-lazyvim")
		cloneErr := errors.New("clone failed")
		cleanupErr := errors.New("cleanup failed")
		originalRemoveAll := removeFrameworkDir
		removeFrameworkDir = func(path string) error {
			if path != targetDir {
				t.Fatalf("cleanup path = %q, want %q", path, targetDir)
			}
			return cleanupErr
		}
		t.Cleanup(func() { removeFrameworkDir = originalRemoveAll })

		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, _ ...string) ([]byte, error) {
				if name != "git" {
					return nil, errors.New("unexpected command: " + name)
				}
				if err := os.MkdirAll(targetDir, 0755); err != nil {
					return nil, err
				}
				return nil, cloneErr
			},
		})

		originalHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = originalHome })

		ctx := types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, Config: config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}}
		err := (NeovimFrameworkModule{}).Install(ctx)
		if !errors.Is(err, cloneErr) {
			t.Fatalf("Install() error = %v, want clone error", err)
		}
		if !errors.Is(err, cleanupErr) {
			t.Fatalf("Install() error = %v, want cleanup error", err)
		}
		if _, statErr := os.Stat(targetDir); statErr != nil {
			t.Fatalf("partial clone state = %v, want retained directory after cleanup failure", statErr)
		}
	})

	t.Run("retry removes partial clone left after cleanup failure before recloning", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, ".config", "nvim-lazyvim")
		cloneErr := errors.New("clone failed")
		cleanupErr := errors.New("cleanup failed")
		cleanupAttempts := 0
		originalRemoveAll := removeFrameworkDir
		removeFrameworkDir = func(path string) error {
			cleanupAttempts++
			if path != targetDir {
				t.Fatalf("cleanup path = %q, want %q", path, targetDir)
			}
			if cleanupAttempts == 1 {
				return cleanupErr
			}
			return os.RemoveAll(path)
		}
		t.Cleanup(func() { removeFrameworkDir = originalRemoveAll })

		cloneAttempts := 0
		withMockExecutor(t, &mockExecutor{
			executeFunc: func(_ context.Context, name string, _ ...string) ([]byte, error) {
				if name != "git" {
					return nil, errors.New("unexpected command: " + name)
				}
				cloneAttempts++
				if err := os.MkdirAll(targetDir, 0755); err != nil {
					return nil, err
				}
				if cloneAttempts == 1 {
					return nil, cloneErr
				}
				createFrameworkClone(t, targetDir)
				return nil, nil
			},
		})

		originalHome := nvimHomeDir
		nvimHomeDir = func() (string, error) { return tmpDir, nil }
		t.Cleanup(func() { nvimHomeDir = originalHome })

		originalRead, originalWrite := zshrcReadFile, zshrcWriteFile
		zshrcReadFile = func(string) ([]byte, error) { return []byte{}, nil }
		writeCalls := 0
		zshrcWriteFile = func(string, []byte, os.FileMode) error {
			writeCalls++
			return nil
		}
		t.Cleanup(func() { zshrcReadFile, zshrcWriteFile = originalRead, originalWrite })

		ctx := types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, Config: config.Config{Nvim: config.NvimOptions{Framework: config.NvimFrameworkLazyVim}}}
		err := (NeovimFrameworkModule{}).Install(ctx)
		if !errors.Is(err, cloneErr) || !errors.Is(err, cleanupErr) {
			t.Fatalf("Install() error = %v, want clone and cleanup errors", err)
		}
		if writeCalls != 0 {
			t.Fatalf("alias writes after failed clone = %d, want 0", writeCalls)
		}

		if err := (NeovimFrameworkModule{}).Install(ctx); err != nil {
			t.Fatalf("retry Install() = %v, want nil", err)
		}
		if cloneAttempts != 2 {
			t.Errorf("clone attempts = %d, want 2", cloneAttempts)
		}
		if cleanupAttempts != 2 {
			t.Errorf("cleanup attempts = %d, want 2", cleanupAttempts)
		}
		if writeCalls != 1 {
			t.Errorf("alias writes after successful retry = %d, want 1", writeCalls)
		}
	})
}

func createFrameworkClone(t *testing.T, targetDir string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(targetDir, ".git", "HEAD"), "ref: refs/heads/main\n")
}

// --- AuditInfo ---

func TestNeovimFramework_AuditInfo(t *testing.T) {
	m := NeovimFrameworkModule{}
	got := m.AuditInfo()
	if got == "" {
		t.Error("AuditInfo() should not be empty")
	}
}
