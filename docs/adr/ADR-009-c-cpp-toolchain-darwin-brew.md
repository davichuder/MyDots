# ADR-009: C/C++ Toolchain on Darwin — brew sobre xcode-select

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

`xcode-select --install` abre un diálogo GUI nativo de macOS que requiere interacción del usuario, incompatible con la automatización de la TUI.

## Opciones consideradas

| Opción | Resultado |
| ------ | --------- |
| `xcode-select --install` | Rechazado — abre diálogo GUI, incompatible con TUI |
| Descargar `.pkg` silenciosamente | Rechazado — requiere tooling adicional y autenticación Apple |

## Decisión

En Darwin, el toolchain C/C++ se instala completamente via Homebrew: `brew install gcc cmake llvm`. Esto provee GCC, CMake y LLVM (que incluye `clangd` para soporte LSP de nvim).

## Consecuencias

Usa GCC en lugar de Apple Clang. Para desarrollo (LSP de nvim, proyectos cmake), es equivalente. `lldb` no se instala; GDB está disponible como alternativa.
