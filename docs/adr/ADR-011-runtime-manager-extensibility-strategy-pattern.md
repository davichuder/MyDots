# ADR-011: Runtime Manager Extensibility — Strategy Pattern

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

Los gestores de runtime (fnm para Node, uv para Python, sdkman para Java) pueden cambiar en el futuro. El usuario podría querer cambiar fnm por bun, o agregar gradle en lugar de maven.

## Decisión

Cada gestor de runtime se implementa como un struct Go que satisface la interfaz `RuntimeManager`. Agregar un nuevo gestor requiere solo un struct nuevo, sin cambios en la lógica de orquestación.

**Contrato de la interfaz:** `Install() error`, `IsInstalled() bool`, `InstallRuntime(version string) error`, `AuditVersion() string`.

## Consecuencias

v1.0.0 incluye fnm, uv y sdkman como implementaciones concretas. Cambiar o agregar gestores en el futuro está aislado a un solo archivo nuevo.
