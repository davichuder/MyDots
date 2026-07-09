# ADR-007: Neovim Frameworks via Git Clone

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

Los scripts de instalación oficiales de algunos frameworks escriben directamente en `~/.config/nvim`, sobrescribiendo la config base.

## Decisión

Todos los frameworks se clonan a `~/.config/nvim-<framework-name>` usando aislamiento por `NVIM_APPNAME`. Cada uno recibe un alias de shell. El comando `nvim` siempre abre la config base o personal.

## Consecuencias

Algunas funcionalidades del framework que asumen `~/.config/nvim` como home pueden comportarse distinto. Trade-off aceptado por la coexistencia.
