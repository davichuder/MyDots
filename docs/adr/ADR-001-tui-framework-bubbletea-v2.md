# ADR-001: TUI Framework — Bubble Tea v2

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

El instalador necesita una TUI interactiva con asistentes multi-paso, vistas de progreso y renderizado de markdown. Herramientas basadas en bash (gum, dialog) son limitadas y no componibles.

## Opciones consideradas

| Opción | Resultado |
| ------ | --------- |
| gum (bash) | Rechazado — no componible, difícil de testear, sin integración Go nativa |
| Bubble Tea v1 | Rechazado — v2 tiene mejor arquitectura multi-modelo y es la rama activa |
| survey (Go) | Rechazado — limitado a formularios, sin control de layout |

## Decisión

Bubble Tea v2 + lipgloss + huh? (ecosistema Charm).

## Consecuencias

Dependencia en una versión mayor pre-stable. Aceptado porque v2 está activamente mantenida y es una herramienta personal, no un producto enterprise.
