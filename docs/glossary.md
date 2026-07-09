# MyDots — Glosario

**Version:** 1.0.0-draft  
**Última actualización:** 2026-06-05

> Lenguaje ubicuo del dominio. Emerge de `proposal.md` y se refina en `specs.md`.

---

## Términos del producto

| Término | Definición |
| ------- | ---------- |
| **MyDots** | Binario Go (`mydots`) que instala, configura y sincroniza un entorno de desarrollo completo en macOS y Linux mediante una TUI. |
| **Dotfiles** | Archivos de configuración personal del usuario gestionados por chezmoi y sincronizados desde un repositorio Git. |
| **Homebrew tap** | Repositorio personal de fórmulas Homebrew (`davichuder/homebrew-tap`) que permite instalar mydots con `brew install mydots`. |
| **chezmoi** | Gestor de dotfiles que sincroniza `~/.config/` con un repositorio Git remoto. Es la herramienta base para la config personal de Neovim. |
| **WSL2 guide** | Guía markdown embebida en el binario que se muestra al ejecutar mydots en Windows. Explica cómo configurar WSL2 manualmente. |

---

## Términos del sistema

| Término | Definición |
| ------- | ---------- |
| **Módulo** | Unidad atómica de instalación identificada por un ID (`M-01` a `M-48`). Cada módulo sabe detectar si ya está instalado, instalarse, y reportar su versión. |
| **Criticality** | Nivel de gravedad de un módulo: `critical` (detiene toda la instalación si falla) o `non-critical` (registra el error y continúa). |
| **InstallStatus** | Estado final de un módulo: `installed`, `skipped`, `skipped-disabled`, `skipped-dependency-failed`, `skipped-no-wayland`, `failed`. |
| **Plataforma** | Objeto que describe el SO destino: `OS` (darwin, linux), `Variant` (native, wsl2) y `Arch` (amd64, arm64). Se detecta en cada ejecución, nunca se guarda. |
| **Config** | Archivo `~/.config/mydots/mydots-config.json` con las preferencias del usuario (fuente, tema, lenguajes opcionales, config de Neovim, repo chezmoi). |
| **Audit** | Archivo `~/.mydots-audit.json` con el registro inmutable de cada módulo instalado: versión, método, estado y timestamp. |
| **Sudo keepalive** | Mecanismo que solicita `sudo -v` una vez al inicio del menú Install y renueva la sesión cada 45 segundos en segundo plano para evitar que caduque durante instalaciones largas. |
| **Idempotence** | Propiedad de que ejecutar mydots N veces produce el mismo resultado final. Módulos ya instalados se registran como `skipped`. |
| **Brew-first** | Estrategia por defecto: Homebrew es el instalador primario en todas las plataformas. Gestores nativos (apt) se usan solo cuando brew no puede o no debe. |
| **TUI** | Interfaz de terminal interactiva construida con Bubble Tea v2. Renderiza menús, progreso de instalación y resultados. |
| **Install pipeline** | Secuencia ordenada de 48 módulos ejecutados en orden fijo. El plan se filtra según la configuración del usuario eliminando módulos opcionales no seleccionados. |
| **Plan** | Lista de módulos a ejecutar, derivada del catálogo completo menos los módulos desactivados en config. |
| **Runner** | Capa de ejecución de comandos del sistema. Wrappers para `exec.CommandContext` con escritura línea por línea a un `io.Writer` para la TUI. |
| **Embed** | Directiva `//go:embed assets` que empaqueta shell scripts, cheatsheets y la guía WSL2 dentro del binario en tiempo de compilación. |

---

## Términos técnicos (arquitectura)

| Término | Definición |
| ------- | ---------- |
| **GenericBrewModule** | Struct de datos que implementa `Module` para herramientas instalables con un solo `brew install`. Declarado como data en `catalogue.go`, no como archivo individual. |
| **RuntimeManager** | Interfaz Go que abstrae gestores de runtimes (fnm, uv, sdkman). Permite agregar un nuevo gestor creando solo un struct, sin modificar lógica de orquestación. |
| **MyDotsError** | Interfaz de errores con tres métodos: `What()` (qué pasó), `Why()` (por qué pasó), `Fix()` (cómo arreglarlo). Todo error mostrado al usuario implementa esta interfaz. |
| **Platform detection** | Lógica que determina `Platform` en cada inicio. `runtime.GOOS` + lectura de `/proc/version` para detectar WSL2. Windows detiene la ejecución y muestra la WSL2 guide. |
| **ProgressEvent** | Struct que el executor envía por canal a la TUI por cada módulo: ID, estado, línea de log, y error si ocurrió. |
| **LogPane** | Ring buffer acotado (200 líneas) que la pantalla de instalación usa para mostrar el log en tiempo real. |
| **ChangeScreenMsg** | Mensaje Bubble Tea que los sub-modelos envían al modelo raíz para solicitar un cambio de pantalla. |
| **Catalogue** | Lista ordenada estáticamente de los 48 módulos en `catalogue.go`, que refleja el orden de ejecución canónico definido en `specs.md`. El orden no se calcula en runtime. |
