# ADR-010: Nerd Fonts on Linux — Manual Install, Not brew --cask

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

Homebrew en Linux (Linuxbrew) no soporta `--cask`. Los casks son una funcionalidad exclusiva de macOS.

## Decisión

En Darwin: `brew install --cask font-<name>-nerd-font`. En Linux: descargar zip desde GitHub Releases (nerdfonts.com), extraer a `~/.local/share/fonts/`, ejecutar `fc-cache -fv`.

## Consecuencias

Dos caminos de código diferentes para instalar fuentes. Es inevitable dada la arquitectura de Homebrew.
