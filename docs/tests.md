# MyDots — Testing Strategy

**Version:** 1.0.0-draft  
**Última actualización:** 2026-06-05

> Estrategia de testing organizada por tipo y stack.
> La trazabilidad entre specs y tests vive en los tests mismos
> mediante tags `@spec:SPEC-XXX`, no en markdown manual.

---

## 1. Stack de testing

| Tipo | Framework | Cobertura |
| ---- | --------- | --------- |
| Unitario | `go test` estándar | Todo `internal/` |
| Aceptación (BDD) | godog | Escenarios `SC-01` a `SC-15` |
| TUI | teatest + golden files | Pantallas y navegación |
| Linting | golangci-lint | Todo el código Go |
| Shell | shellcheck | `assets/scripts/*.sh` |

---

## 2. Tests unitarios

Cada paquete bajo `internal/` tiene su `_test.go`. Ningún `.go` (excepto
auto-generados) se entrega sin tests.

### 2.1 Principios

- **Triangulación:** múltiples inputs para cada función (no un solo caso
  feliz)
- **Caja negra:** se testea comportamiento, no implementación interna
- **Path coverage:** todas las ramas de cada función
- **Edge cases reales:** archivos vacíos, permisos denegados, formatos
  inesperados

### 2.2 Paquetes y qué testean

| Paquete | Qué testear |
| ------- | ----------- |
| `platform` | `Detect()` con `/proc/version` mockeado — Darwin, Linux native, WSL2, Windows, desconocido |
| `config` | `Load`, `Save`, `Validate` — JSON válido, mal formado, campos inválidos, URLs rotas |
| `audit` | `Append` — escritura atómica, directorio faltante, concurrencia (N gorutinas) |
| `backup` | `BackupFile`, `BackupDir`, `ListBackups`, `DeleteBackup` — preservación de paths, symlinks |
| `sudo` | `StartKeepalive` — goroutine se detiene limpiamente, stdout va a `io.Discard` |
| `runtime` | Cada manager: `IsInstalled`, `Install`, `InstallRuntime`, `AuditInfo` con runner mockeado |
| `installer/runner` | `Run`, `Brew`, `BrewCask`, `Script`, `CommandExists`, `CaptureOutput` — salida línea por línea, cancelación por contexto |
| `installer/brew_module` | `BrewModule.IsInstalled`, `Install`, `AuditInfo` con runner mockeado |
| `installer/catalogue` | 48 módulos sin IDs duplicados, orden canónico, dependencias válidas |
| `installer/planner` | `BuildPlan` — todas las combinaciones de config producen la lista correcta |
| `installer/executor` | `Run` — éxito total, fallo no crítico, fallo crítico, dependencia fallida, el canal siempre se cierra |
| `installer/modules` | Cada módulo complejo: todas las variantes de SO, idempotencia, caminos de fallo |

### 2.3 Mock runner

Ningún test llama a herramientas reales del sistema.

```go
runner.SetExecutor(fakeExecutor)
```

El fake executor registra llamadas y devuelve resultados configurados
por el test.

---

## 3. Tests de aceptación (BDD)

Cada escenario `SC-*` de `specs.md` §9 se implementa como un archivo
`.feature` con godog.

### 3.1 Feature files

```text
internal/test/features/
├── fresh_install_macos.feature       (SC-01)
├── fresh_install_ubuntu.feature      (SC-02)
├── fresh_install_wsl2.feature        (SC-03)
├── idempotent_rerun.feature          (SC-04)
├── noncritical_failure.feature       (SC-05)
├── critical_failure.feature          (SC-06)
├── windows_detection.feature         (SC-07)
├── install_without_config.feature     (SC-08)
├── personal_nvim.feature             (SC-09)
├── nvim_framework.feature            (SC-10)
├── sudo_keepalive.feature            (SC-11)
├── default_flag.feature              (SC-12)
├── unattended_no_config.feature      (SC-13)
├── chezmoi_failure.feature           (SC-14)
└── theme_idempotent.feature          (SC-15)
```

### 3.2 Step definitions

Las definiciones de pasos viven en `internal/test/steps/` e inyectan la
plataforma mediante `Detect(goos)` con un GOOS mockeado y un directorio
home temporal. El runner está mockeado; los escritores de audit, backup
y config son reales (operan sobre el home temporal).

### 3.3 Ejecución

```bash
make test-bdd
```

---

## 4. Tests de TUI

Usan **teatest** (`github.com/charmbracelet/x/exp/teatest`) para
programas Bubble Tea completos y **golden files** para regresión visual.

### 4.1 Unitarios (rápidos, sin teatest)

El estado del modelo se manipula directamente y se verifica:

- Salida de `View()` contiene/no contiene texto esperado
- `Update()` retorna el mensaje de cambio de pantalla correcto
- Transiciones de estado se reflejan en los campos del modelo

### 4.2 Integración (teatest)

Se lanza el programa completo con `teatest.NewProgram` y se espera
hasta que cierto texto aparezca en la salida:

```go
tm := teatest.NewProgram(t, NewApp(platform))
teatest.WaitFor(t, tm.Output(), "Main Menu", time.Second)
```

### 4.3 Golden files

Las salidas de `View()` se comparan con archivos golden en
`testdata/golden/`:

```text
testdata/golden/
├── main-menu.golden
├── config-step-1.golden
├── install-progress-row-ok.golden
├── install-progress-row-fail.golden
└── result-success.golden
```

Actualizar con:

```bash
make update-golden
```

---

## 5. Integración (build tag)

Los tests marcados con `//go:build integration` se ejecutan por
separado:

```bash
go test -tags=integration ./...
```

No se incluyen en `go test ./...` estándar porque son más lentos o
requieren setup específico (ej. un home directory temporal completo con
estructura simulada).

---

## 6. CI Pipeline

En cada PR (`.github/workflows/ci.yml`):

```yaml
- go test ./...
- golangci-lint run
- shellcheck assets/scripts/*.sh
```

En tag `v*` (release workflow):

```yaml
- go test ./...
- go test -tags=integration ./...
- make test-bdd
- goreleaser release
```

---

## 7. Trazabilidad

Los tests individuales referencian especificaciones mediante tags:

```go
// @spec:SPEC-001
// @spec:FR-05
func TestInstallWithoutConfig(t *testing.T) { ... }
```

No existe un archivo de trazabilidad mantenido a mano. La relación
spec → test se verifica en el pipeline.

---

## 8. Cobertura

- Mínimo aceptable: 80% de cobertura en paquetes de lógica de negocio
  (`internal/installer`, `internal/config`, `internal/audit`,
  `internal/backup`)
- Paquetes de infraestructura TUI (`internal/tui/screens/`): 60% mínimo
  (unittest de modelos + teatest de integración)
- Cobertura de `internal/errors/`, `internal/sudo/`: 100% de los tipos
  de error y funciones exportadas
