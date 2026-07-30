package modules

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/davichuder/MyDots/internal/installer/types"
	"github.com/davichuder/MyDots/internal/platform"
)

func TestZshInstallLinuxUsesElevatedExactCommandSequence(t *testing.T) {
	t.Setenv("USER", "mydots")

	var calls [][]string
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			calls = append(calls, append([]string{name}, args...))
			switch {
			case name == "brew" && reflect.DeepEqual(args, []string{"install", "zsh"}):
				return nil, nil
			case name == "brew" && reflect.DeepEqual(args, []string{"--prefix"}):
				return []byte("/home/linuxbrew/.linuxbrew\n"), nil
			case name == "sudo" && reflect.DeepEqual(args, []string{"sh", "-c", `printf '%s\n' "$1" >> "$2"`, "sh", "/home/linuxbrew/.linuxbrew/bin/zsh", "/etc/shells"}):
				return nil, nil
			case name == "chsh" && reflect.DeepEqual(args, []string{"-s", "/home/linuxbrew/.linuxbrew/bin/zsh", "mydots"}):
				return nil, nil
			default:
				return nil, errors.New("unexpected command")
			}
		},
	})

	origRead := readShellsFile
	t.Cleanup(func() { readShellsFile = origRead })
	readShellsFile = func(string) ([]byte, error) { return []byte("/bin/zsh\n"), nil }

	ctx := types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native}}
	if err := (ZshModule{}).Install(ctx); err != nil {
		t.Fatalf("Install() = %v, want nil", err)
	}

	want := [][]string{
		{"brew", "install", "zsh"},
		{"brew", "--prefix"},
		{"sudo", "sh", "-c", `printf '%s\n' "$1" >> "$2"`, "sh", "/home/linuxbrew/.linuxbrew/bin/zsh", "/etc/shells"},
		{"chsh", "-s", "/home/linuxbrew/.linuxbrew/bin/zsh", "mydots"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("commands = %#v, want %#v", calls, want)
	}
}

func TestZshInstallStopsBeforeShellChangesWhenBrewPrefixFails(t *testing.T) {
	var calls [][]string
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			calls = append(calls, append([]string{name}, args...))
			if name == "brew" && reflect.DeepEqual(args, []string{"install", "zsh"}) {
				return nil, nil
			}
			if name == "brew" && reflect.DeepEqual(args, []string{"--prefix"}) {
				return nil, errors.New("brew prefix unavailable")
			}
			return nil, errors.New("must not run after prefix failure")
		},
	})

	ctx := types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native}}
	if err := (ZshModule{}).Install(ctx); err == nil {
		t.Fatal("Install() = nil, want brew --prefix error")
	}
	want := [][]string{{"brew", "install", "zsh"}, {"brew", "--prefix"}}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("commands = %#v, want %#v", calls, want)
	}
}

func TestZshInstallDoesNotElevateWhenPathAlreadyRegistered(t *testing.T) {
	t.Setenv("USER", "mydots")
	var sudoCalled bool
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			switch {
			case name == "brew" && reflect.DeepEqual(args, []string{"install", "zsh"}):
				return nil, nil
			case name == "brew" && reflect.DeepEqual(args, []string{"--prefix"}):
				return []byte("/home/linuxbrew/.linuxbrew"), nil
			case name == "sudo":
				sudoCalled = true
				return nil, nil
			case name == "chsh":
				return nil, nil
			default:
				return nil, errors.New("unexpected command")
			}
		},
	})

	origRead := readShellsFile
	t.Cleanup(func() { readShellsFile = origRead })
	readShellsFile = func(string) ([]byte, error) { return []byte("/home/linuxbrew/.linuxbrew/bin/zsh\n"), nil }

	ctx := types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native}}
	if err := (ZshModule{}).Install(ctx); err != nil {
		t.Fatalf("Install() = %v, want nil", err)
	}
	if sudoCalled {
		t.Error("sudo must not run when the brew zsh path is already registered")
	}
}

func TestZshInstallStopsBeforeChshWhenShellRegistrationFails(t *testing.T) {
	var chshCalled bool
	withMockExecutor(t, &mockExecutor{
		executeFunc: func(_ context.Context, name string, args ...string) ([]byte, error) {
			switch {
			case name == "brew" && reflect.DeepEqual(args, []string{"install", "zsh"}):
				return nil, nil
			case name == "brew" && reflect.DeepEqual(args, []string{"--prefix"}):
				return []byte("/home/linuxbrew/.linuxbrew"), nil
			case name == "sudo":
				return nil, errors.New("sudo denied")
			case name == "chsh":
				chshCalled = true
				return nil, nil
			default:
				return nil, errors.New("unexpected command")
			}
		},
	})

	origRead := readShellsFile
	t.Cleanup(func() { readShellsFile = origRead })
	readShellsFile = func(string) ([]byte, error) { return []byte("/bin/zsh\n"), nil }

	ctx := types.InstallContext{Cancel: context.Background(), Log: &bytes.Buffer{}, Platform: platform.Platform{OS: platform.Linux, Variant: platform.Native}}
	if err := (ZshModule{}).Install(ctx); err == nil {
		t.Fatal("Install() = nil, want registration error")
	}
	if chshCalled {
		t.Error("chsh must not run after shell registration failure")
	}
}
