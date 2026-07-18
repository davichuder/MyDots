package installer

import "github.com/davichuder/MyDots/internal/platform"

// stubModule is a minimal Module placeholder for phases 2.6+.
type stubModule struct {
	id          ModuleID
	name        string
	criticality Criticality
	deps        []ModuleID
}

func (m stubModule) ID() ModuleID                              { return m.id }
func (m stubModule) Name() string                               { return m.name }
func (m stubModule) Criticality() Criticality                   { return m.criticality }
func (m stubModule) Dependencies() []ModuleID                   { return m.deps }
func (m stubModule) IsInstalled(_ platform.Platform) bool       { return false }
func (m stubModule) Install(_ InstallContext) error             { return nil }
func (m stubModule) AuditInfo() string                          { return "" }

// Complex module stubs. Each exported variable implements Module.
// These are placeholder implementations for phases 2.6+.
var (
	Homebrew          Module = stubModule{id: ModHomebrew,       name: "Homebrew",          criticality: Critical,    deps: nil}
	Ghostty           Module = stubModule{id: ModGhostty,        name: "Ghostty",           criticality: NonCritical, deps: nil}
	NerdFont          Module = stubModule{id: ModNerdFont,       name: "Nerd Font",         criticality: NonCritical, deps: nil}
	Git               Module = stubModule{id: ModGit,            name: "Git",               criticality: NonCritical, deps: []ModuleID{ModHomebrew}}
	GitCredOAuth      Module = stubModule{id: ModGitCredOAuth,   name: "git-credential-oauth", criticality: NonCritical, deps: []ModuleID{ModGit}}
	Chezmoi           Module = stubModule{id: ModChezmoi,        name: "Chezmoi",           criticality: NonCritical, deps: []ModuleID{ModGit}}
	Zsh               Module = stubModule{id: ModZsh,            name: "Zsh",               criticality: NonCritical, deps: nil}
	OhMyZsh           Module = stubModule{id: ModOhMyZsh,        name: "Oh My Zsh",         criticality: NonCritical, deps: []ModuleID{ModZsh}}
	Fnm               Module = stubModule{id: ModFnm,            name: "fnm",               criticality: NonCritical, deps: nil}
	Node              Module = stubModule{id: ModNode,           name: "Node 24",           criticality: NonCritical, deps: []ModuleID{ModFnm}}
	Uv                Module = stubModule{id: ModUv,             name: "uv",                criticality: NonCritical, deps: nil}
	Python            Module = stubModule{id: ModPython,         name: "Python 3.12",       criticality: NonCritical, deps: []ModuleID{ModUv}}
	CppToolchain      Module = stubModule{id: ModCppToolchain,   name: "C/C++ toolchain",   criticality: NonCritical, deps: nil}
	Sdkman            Module = stubModule{id: ModSdkman,         name: "sdkman",            criticality: NonCritical, deps: nil}
	Java              Module = stubModule{id: ModJava,           name: "Java 25",           criticality: NonCritical, deps: []ModuleID{ModSdkman}}
	Php               Module = stubModule{id: ModPhp,            name: "PHP",               criticality: NonCritical, deps: nil}
	Neovim            Module = stubModule{id: ModNeovim,         name: "Neovim",            criticality: NonCritical, deps: nil}
	NeovimPersonal    Module = stubModule{id: ModNeovimPersonal, name: "Nvim personal",     criticality: NonCritical, deps: []ModuleID{ModNeovim, ModChezmoi}}
	NeovimFramework   Module = stubModule{id: ModNeovimFramework,name: "Nvim framework",    criticality: NonCritical, deps: []ModuleID{ModNeovim}}
	Theme             Module = stubModule{id: ModTheme,          name: "Theme",             criticality: NonCritical, deps: []ModuleID{ModNeovim, ModOhMyZsh, ModZellij, ModGhostty}}
	Clipboard         Module = stubModule{id: ModClipboard,      name: "Clipboard",         criticality: NonCritical, deps: nil}
	Docker            Module = stubModule{id: ModDocker,         name: "Docker",            criticality: NonCritical, deps: nil}
	McpConfig         Module = stubModule{id: ModMcpConfig,      name: "MCP config",        criticality: NonCritical, deps: []ModuleID{ModOpencode}}
	Rtk               Module = stubModule{id: ModRtk,            name: "RTK",               criticality: NonCritical, deps: nil}
	Caveman           Module = stubModule{id: ModCaveman,        name: "Caveman",           criticality: NonCritical, deps: nil}
	GentleAi          Module = stubModule{id: ModGentleAi,       name: "Gentle AI",         criticality: NonCritical, deps: nil}
)
