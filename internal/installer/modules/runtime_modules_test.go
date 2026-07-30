package modules

import (
	"context"
	"io"
	"io/fs"
)

type mockRuntimeManager struct {
	installFunc        func(context.Context, io.Writer, fs.FS) error
	isInstalledFunc    func() bool
	installRuntimeFunc func(context.Context, string, io.Writer) error
	setDefaultFunc     func(context.Context, string, io.Writer) error
	auditInfoFunc      func() string
}

func (m *mockRuntimeManager) Install(ctx context.Context, logw io.Writer, assets fs.FS) error {
	if m.installFunc != nil {
		return m.installFunc(ctx, logw, assets)
	}
	return nil
}
func (m *mockRuntimeManager) IsInstalled() bool {
	if m.isInstalledFunc != nil {
		return m.isInstalledFunc()
	}
	return false
}
func (m *mockRuntimeManager) InstallRuntime(ctx context.Context, version string, logw io.Writer) error {
	if m.installRuntimeFunc != nil {
		return m.installRuntimeFunc(ctx, version, logw)
	}
	return nil
}
func (m *mockRuntimeManager) SetDefaultRuntime(ctx context.Context, version string, logw io.Writer) error {
	if m.setDefaultFunc != nil {
		return m.setDefaultFunc(ctx, version, logw)
	}
	return nil
}
func (m *mockRuntimeManager) AuditInfo() string {
	if m.auditInfoFunc != nil {
		return m.auditInfoFunc()
	}
	return ""
}
