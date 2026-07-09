# MyDots — Styleguide

**Version:** 1.0.0-draft  
**Última actualización:** 2026-06-05

> Convenciones verificables de código, linters, commits y branching.
> Las reglas aquí son automatizables. El razonamiento detrás de ellas
> vive en `docs/principles/`.

---

## 1. Go

### 1.1 Enums tipados

Todos los enums son `type X string` con constantes tipadas. No existen
strings literales para valores de enums en ningún lugar del código.

```go
// ✅ Correcto
type NvimConfig string
const (
    NvimConfigBase     NvimConfig = "base"
    NvimConfigPersonal NvimConfig = "personal"
)

// ❌ Incorrecto
nvimConfig := "base" // string literal sin tipo
```

### 1.2 Nombres

- `camelCase` para vars y funciones no exportadas
- `PascalCase` para tipos, interfaces y funciones exportadas
- `SCREAMING_SNAKE` para constantes solo cuando el valor tiene
  significado semántico especial (ej. IDs de módulos como `ModHomebrew`)
- Abreviaciones: una letra por scope pequeño (`i`, `n`), palabras
  completas para scope amplio (`moduleID`, `configPath`)
- Evitar prefijos de paquete en nombres exportados (`config.Config` es
  ruido → `config.Config` es aceptable solo cuando el nombre del tipo
  es distinto al del paquete)

### 1.3 Legibilidad (NFR-02)

Cada archivo Go debe ser entendible por un desarrollador junior sin
necesitar comentarios como muleta.

- Sin one-liners que sacrifiquen claridad
- `if` con `return` temprano preferido sobre `else` anidado
- Funciones de menos de 40 líneas idealmente
- Un propósito por función (SRP)

### 1.4 Magic strings

Ningún string literal con significado semántico aparece directamente
en el código. Todo valor que se repite, representa un dominio, o
tiene significado especial se declara como constante tipada.

```go
// ✅ Correcto — constantes con tipo explícito
const DefaultChezmoiRepoURL string = "https://github.com/davichuder/dotfiles"
const procVersionPath    string = "/proc/version"
const wsl2Marker         string = "microsoft"

// ❌ Incorrecto — magic string en medio del código
os.ReadFile("/proc/version")
strings.Contains(data, "microsoft")

// ✅ Correcto — constantes de path agrupadas
const configDirName  string = ".config"
const appDirName     string = "mydots"
const configFileName string = "mydots-config.json"

// ❌ Incorrecto — path segments como literales
filepath.Join(home, ".config", "mydots", "mydots-config.json")

// ✅ Correcto — error messages pueden ser literales (son el mensaje en sí)
return fmt.Errorf("invalid font: %q", cfg.Font)
```

**Excepciones:** Los mensajes de error (`fmt.Errorf`), format strings,
y valores únicos de prueba que solo aparecen una vez y son obvios
en contexto (ej. `"NonExistent"` en un test de validación).

**Regla práctica:** Si el string aparece en más de un lugar o
tiene significado de dominio (no es solo un mensaje), es constante.

### 1.5 Imports

Tres grupos separados por línea en blanco:

```go
import (
    // stdlib
    "os"
    "strings"

    // externos
    "charm.land/bubbletea/v2"
    "github.com/cucumber/godog"

    // internos
    "github.com/davichuder/MyDots/internal/config"
)
```

---

## 2. Commits

### 2.1 Formato

[Conventional Commits v1.0.0](https://www.conventionalcommits.org/en/v1.0.0/):

```text
tipo(alcance): descripción imperativa (≤ 50 chars)

Cuerpo opcional explicando el POR QUÉ, no el qué.
```

### 2.2 Tipos

| Tipo | Uso |
| ---- | --- |
| `feat` | Nueva funcionalidad |
| `fix` | Corrección de bug |
| `refactor` | Cambio que no agrega funcionalidad ni corrige bug |
| `test` | Agregar o corregir tests |
| `docs` | Documentación (README, docs/, comentarios) |
| `style` | Formato, linting, espacios, sin cambio de lógica |
| `chore` | Mantenimiento, dependencias, CI |
| `perf` | Mejora de rendimiento |
| `ci` | Cambios en CI/CD |
| `build` | Cambios en el sistema de build |

### 2.3 Cadencia

Cada par RED-GREEN es un commit atómico:

```bash
git commit -m "test(platform): detect WSL2 from /proc/version"
git commit -m "feat(platform): implement Detect()"
```

---

## 3. Testing

### 3.1 Ciclo TDD

```text
[RED]   Escribir test que falla → commit "test(paquete): ..."
[GREEN] Implementar mínimo para que pase → commit "feat(paquete): ..."
        Refactor si es necesario antes del siguiente par
```

### 3.2 Convenciones

- Ningún `.go` (excepto auto-generados) sin su `_test.go`
- Tests unitarios con `go test` estándar
- Tests de aceptación con **godog** (Gherkin + BDD)
- Tests de TUI con **teatest** + **golden files**
- Integración con build tag `//go:build integration`
- Uso de triangulación (múltiples inputs), caja negra (comportamiento,
  no implementación) y path coverage (todas las ramas)

### 3.3 Mock runner

Ningún test llama a `brew`, `apt` o `curl` reales. Usar
`runner.SetExecutor()` para inyectar un executor fake que registra
llamadas y devuelve resultados configurados.

---

## 4. Shell scripts

- POSIX-compliant (`shellcheck --shell=sh`)
- Sin bashismos (`[[ ]]`, `source`, arrays)
- Recibir parámetros via variables de entorno, no argumentos posicionales
- Terminar con `exit 0` en éxito, `exit 1` en error

---

## 5. Estructura de archivos

```text
mydots/
├── main.go
├── assets.go                // //go:embed assets
│
├── assets/
│   ├── wsl2-guide.md
│   ├── cheatsheets/
│   └── scripts/
│
└── internal/
    ├── platform/
    ├── config/
    ├── audit/
    ├── backup/
    ├── sudo/
    ├── runtime/
    ├── installer/
    │   ├── runner/
    │   └── modules/
    ├── errors/
    └── tui/
        ├── screens/
        └── components/
```

- Módulos simples (brew install único): se declaran como datos en
  `catalogue.go`, no tienen archivo individual
- Módulos complejos (lógica OS-specific, multi-step): un archivo por
  módulo en `internal/installer/modules/`

---

## 6. Linters

`.golangci.yml` debe habilitar como mínimo:

- `errcheck` — errores no manejados
- `govet` — sospechas de código
- `staticcheck` — análisis avanzado
- `nolintlint` — abuso de `//nolint`

---

## 7. Branching

```text
main        ← releases
  └─ dev    ← integración
       └─ feature/nombre   ← features
       └─ fix/nombre       ← bugs
```

Commits pequeños, rebase para limpiar antes de merge a `dev`.
