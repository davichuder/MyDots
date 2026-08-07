package screens

import (
	"fmt"
	"strings"

	"charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/davichuder/MyDots/internal/config"
)

var configMenuFontOptions = []huh.Option[config.FontChoice]{
	huh.NewOption("JetBrains Mono", config.FontJetBrainsMono),
	huh.NewOption("Fira Code", config.FontFiraCode),
	huh.NewOption("Cascadia Code", config.FontCascadiaCode),
	huh.NewOption("Hack", config.FontHack),
	huh.NewOption("Iosevka", config.FontIosevka),
}

// ConfigMenu delegates six configuration steps to the Huh v2 Bubble Tea adapter.
type ConfigMenu struct {
	form        *huh.Form
	config      config.Config
	formConfig  *config.Config
	languages   *[]string
	configPath  string
	fontOptions []huh.Option[config.FontChoice]
	width       int
	height      int
}

// NewConfigMenu creates a configuration assistant initialized with application defaults.
func NewConfigMenu(configPath string) ConfigMenu {
	cfg := config.DefaultConfig()
	menu := ConfigMenu{
		config:      cfg,
		configPath:  configPath,
		fontOptions: configMenuFontOptions,
	}
	menu.form = menu.newForm()
	return menu
}

func (menu *ConfigMenu) newForm() *huh.Form {
	formConfig := menu.config
	menu.formConfig = &formConfig
	languages := selectedLanguages(menu.config.Languages)
	menu.languages = &languages

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[config.FontChoice]().
				Key("font").
				Title("Choose a font").
				Options(menu.fontOptions...).
				Value(&menu.formConfig.Font),
		),
		huh.NewGroup(
			huh.NewSelect[config.ThemeChoice]().
				Key("theme").
				Title("Choose a theme").
				Options(
					huh.NewOption("Tokyo Night", config.ThemeTokyoNight),
					huh.NewOption("Catppuccin Mocha", config.ThemeCatppuccin),
					huh.NewOption("Gruvbox Dark", config.ThemeGruvbox),
					huh.NewOption("Dracula", config.ThemeDracula),
					huh.NewOption("Kanagawa", config.ThemeKanagawa),
				).
				Value(&menu.formConfig.Theme),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("languages").
				Title("Optional languages").
				Options(
					huh.NewOption("Java", "java").Selected(menu.formConfig.Languages.Java),
					huh.NewOption("PHP", "php").Selected(menu.formConfig.Languages.PHP),
				).
				Value(menu.languages),
		),
		huh.NewGroup(
			huh.NewSelect[config.NvimConfig]().
				Key("nvim-config").
				Title("Choose a Neovim configuration").
				Options(
					huh.NewOption("Base", config.NvimConfigBase),
					huh.NewOption("Personal", config.NvimConfigPersonal),
				).
				Value(&menu.formConfig.Nvim.Config),
		),
		huh.NewGroup(
			huh.NewSelect[config.NvimFramework]().
				Key("nvim-framework").
				Title("Choose a Neovim framework").
				Options(
					huh.NewOption("None", config.NvimFrameworkNone),
					huh.NewOption("LazyVim", config.NvimFrameworkLazyVim),
					huh.NewOption("LunarVim", config.NvimFrameworkLunarVim),
					huh.NewOption("AstroNvim", config.NvimFrameworkAstroNvim),
					huh.NewOption("NvChad", config.NvimFrameworkNvChad),
					huh.NewOption("LaunchVim", config.NvimFrameworkLaunchVim),
				).
				Value(&menu.formConfig.Nvim.Framework),
		),
		huh.NewGroup(
			huh.NewInput().
				Key("chezmoi-url").
				Title("Chezmoi repository URL").
				Value(&menu.formConfig.Chezmoi.RepoURL).
				Validate(func(value string) error {
					if !strings.HasPrefix(value, "https://") {
						return fmt.Errorf("chezmoi repo_url must be an HTTPS URL")
					}
					return nil
				}),
		),
	)
}

func selectedLanguages(languages config.LanguageOptions) []string {
	selected := make([]string, 0, 2)
	if languages.Java {
		selected = append(selected, "java")
	}
	if languages.PHP {
		selected = append(selected, "php")
	}
	return selected
}

func languageOptions(selected []string) config.LanguageOptions {
	var languages config.LanguageOptions
	for _, language := range selected {
		switch language {
		case "java":
			languages.Java = true
		case "php":
			languages.PHP = true
		}
	}
	return languages
}

func (menu ConfigMenu) Init() tea.Cmd {
	return menu.form.Init()
}

func (menu ConfigMenu) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		menu.width = msg.Width
		menu.height = msg.Height
		menu.form = menu.form.WithWidth(msg.Width).WithHeight(msg.Height)
	}

	updated, command := menu.form.Update(message)
	if form, ok := updated.(*huh.Form); ok {
		menu.form = form
	}
	if menu.form.State == huh.StateCompleted {
		return menu, menu.saveAndReturn()
	}
	return menu, command
}

func (menu ConfigMenu) View() tea.View {
	return tea.NewView(ansi.Strip(menu.form.View()))
}

func (menu ConfigMenu) validate() error {
	return config.Validate(menu.config)
}

func (menu ConfigMenu) saveAndReturn() tea.Cmd {
	return func() tea.Msg {
		cfg := *menu.formConfig
		cfg.Languages = languageOptions(*menu.languages)
		if err := config.Validate(cfg); err != nil {
			return ConfigIssue{Err: err}
		}
		if err := config.Save(menu.configPath, cfg); err != nil {
			return ConfigIssue{Err: err}
		}
		return ChangeScreenMsg{Screen: ScreenMain}
	}
}
