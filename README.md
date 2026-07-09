# MyDots

> Un comando, cualquier máquina, mismo resultado, cada vez.

MyDots es un instalador de entorno de desarrollo con TUI escrito en Go.
Configurá una máquina nueva desde cero en **un solo comando**.

```bash
brew tap davichuder/homebrew-tap && brew install mydots && mydots
```

O sin Homebrew:

```bash
git clone https://github.com/davichuder/MyDots
cd MyDots && go run .
```

---

## Cómo funciona

1. **Preflight** — detecta el SO (macOS, Linux, WSL2). En Windows muestra una guía para configurar WSL2.
2. **Config** — elegís fuente, tema, lenguajes opcionales y config de Neovim. Un solo paso, una sola vez.
3. **Install** — todo lo demás corre solo. 48 módulos en orden, con progreso en tiempo real.
4. **Ready** — terminal lista para desarrollar.

## Qué instala (siempre)

| Categoría | Herramientas |
| --------- | ------------ |
| Package manager | Homebrew |
| Shell | Zsh + Oh My Zsh |
| Multiplexer | Zellij |
| Shell tools | Atuin, zoxide, lazygit, gh, wl-clipboard / pbcopy |
| AI tools | opencode, RTK, gentle-ai, caveman |
| Editor | Neovim (con base config y frameworks opcionales) |
| Version control | Git, git-credential-oauth |
| Dotfiles sync | chezmoi → tu repo de dotfiles |
| Containers | Docker Desktop (macOS) / Docker Engine (Linux) |
| Runtimes | Node 24 (fnm), Python 3.12 (uv), Go 1.26, C/C++ toolchain |
| MCP servers | supabase, angular, primeng, context7, postman, playwright |

## Opcional (elegís en Config)

- Java 25 (sdkman), PHP
- Neovim framework: LazyVim, LunarVim, AstroNvim, NvChad, Launch.nvim
- Config personal de Neovim desde tus dotfiles
- 5 temas y 5 Nerd Fonts para elegir

## Flags

| Flag | Qué hace |
| ---- | -------- |
| `mydots --default` | Genera config con defaults e instala sin TUI |
| `mydots --unattended` | Usa config guardada, instala sin TUI |
| `mydots --version` | Versión del binario |

## Plataformas soportadas

| Plataforma | v1.0.0 |
| ---------- | ------ |
| macOS Darwin | ✅ |
| Ubuntu / WSL2 | ✅ |
| Arch Linux | 🔜 v1.1.0 |
| Debian | 🔜 v1.1.0 |
| Fedora | 🔜 v1.1.0 |
| Windows (host) | ❌ Guía WSL2 |

## Documentación

| Archivo | Para qué |
| ------- | -------- |
| `docs/proposal.md` | Visión, alcance y roadmap |
| `docs/specs.md` | Qué hace el sistema (fuente de verdad) |
| `docs/design.md` | Arquitectura técnica |
| `docs/tasks.md` | Qué falta implementar |
| `docs/glossary.md` | Lenguaje del dominio |
| `docs/styleguide.md` | Convenciones de código |
| `docs/ONBOARDING.md` | Setup del proyecto en < 15 min |
| `docs/tests.md` | Estrategia de testing |
| `docs/adr/` | Decisiones arquitectónicas |
| `docs/CHANGELOG.md` | Historial de cambios |

## Principios

Este proyecto sigue **Spec-Driven Development (SDD)**.
Primero las specs, después el diseño, después las tareas, después el código.
Los principios de ingeniería están en `docs/principles/`.

---

**Hecho con ❤️ por David.**  
[Repo](https://github.com/davichuder/MyDots) — [Homebrew tap](https://github.com/davichuder/homebrew-tap)
