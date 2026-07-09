# ADR-002: Primary Package Manager — Homebrew

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

El instalador apunta tanto a macOS como a Linux. Una herramienta unificada reduce la superficie de estrategias de instalación.

## Opciones consideradas

| Opción | Resultado |
| ------ | --------- |
| apt-only | Rechazado — incompatible con Darwin |
| nix | Rechazado — curva de aprendizaje demasiado pronunciada para contribuidores |
| mise | Rechazado — bueno para runtimes solamente, no para herramientas generales |

## Decisión

Homebrew es el gestor por defecto para todas las herramientas en ambas plataformas. Gestores nativos (apt) se usan solo cuando brew no puede manejar la instalación (ej. Docker Engine, dependencias C/C++ a nivel sistema en Ubuntu).

## Consecuencias

`--cask` es solo macOS. Las instalaciones de fuentes Linux y Docker requieren estrategias nativas del SO. Esto se maneja por módulo en la matriz de estrategias de instalación.
