# MyDots — Onboarding

Levantar el proyecto en menos de 15 minutos.

---

## Prerequisitos

| Herramienta | Versión | Cómo verificarlo |
| ----------- | ------- | ---------------- |
| Go | 1.26.3+ | `go version` |
| Git | cualquier versión moderna | `git --version` |
| Make | cualquier versión | `make --version` |
| shellcheck | cualquiera | `shellcheck --version` (opcional, para validar scripts) |

---

## Setup

```bash
# 1. Clonar
git clone https://github.com/davichuder/MyDots
cd MyDots

# 2. Compilar
go build -o mydots .

# 3. Ejecutar
./mydots

#   Para desarrollo:
go run .
```

---

## Comandos útiles

```bash
go build ./...       # compilar todo
go test ./...        # tests unitarios
go vet ./...         # análisis estático
golangci-lint run    # linter completo
```

---

## Tests

```bash
# Unitarios (rápidos)
go test ./...

# Tests de integración (más lentos, requieren setup)
go test -tags=integration ./...

# Tests BDD (godog)
make test-bdd

# Actualizar golden files
make update-golden

# Verificar shell scripts
shellcheck --shell=sh assets/scripts/*.sh
```

---

## Makefile targets

| Target | Descripción |
| ------ | ----------- |
| `build` | Compila el binario |
| `test` | Tests unitarios |
| `test-integration` | Tests con tag integration |
| `test-bdd` | Tests BDD con godog |
| `lint` | golangci-lint |
| `shellcheck` | Verifica scripts POSIX |
| `update-golden` | Actualiza golden files de tests TUI |

---

## Estructura del proyecto

```text
mydots/
├── main.go              ← entry point
├── assets.go            ← embebidos
├── assets/
│   ├── cheatsheets/     ← documentación de cada tool
│   └── scripts/         ← shell scripts POSIX
└── internal/
    ├── platform/        ← detección de SO
    ├── config/          ← mydots-config.json
    ├── audit/           ← mydots-audit.json
    ├── backup/          ← backup de configs existentes
    ├── sudo/            ← elevación + keepalive
    ├── runtime/         ← gestores de runtime (fnm, uv, sdkman)
    ├── installer/       ← pipeline de instalación
    │   ├── runner/      ← exec.Command wrappers
    │   └── modules/     ← módulos complejos
    ├── errors/          ← errores con What/Why/Fix
    └── tui/             ← Bubble Tea v2
        ├── screens/
        └── components/
```

---

## Docs

| Archivo | Para qué leerlo |
| ------- | --------------- |
| `docs/proposal.md` | Visión, alcance y roadmap del producto |
| `docs/specs.md` | Qué hace el sistema (fuente de verdad) |
| `docs/design.md` | Arquitectura técnica |
| `docs/tasks.md` | Qué falta implementar |
| `docs/glossary.md` | Lenguaje del dominio |
| `docs/styleguide.md` | Convenciones de código |
| `docs/tests.md` | Estrategia de testing |
| `docs/adr/` | Decisiones arquitectónicas |

---

## Debugging

```bash
# Error común: falta de Homebrew
./mydots --unattended
# → "Error: no config found. Run 'mydots' to configure first."

# Generar config por defecto e instalar
./mydots --default
```
