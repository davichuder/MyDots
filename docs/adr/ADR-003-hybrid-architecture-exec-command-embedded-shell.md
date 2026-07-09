# ADR-003: Hybrid Architecture — Go exec.Command + Embedded Shell Scripts

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

Algunas instalaciones son un solo comando. Otras requieren lógica condicional, bucles o manipulación de paths.

## Decisión

Instalaciones de un solo comando usan `exec.Command` directamente en Go. Instalaciones multi-paso o condicionales usan scripts `.sh` embebidos via `embed.FS`.

Regla: si una instalación requiere más de un comando con lógica condicional, bucles o manipulación de paths → va en un script shell.

## Consecuencias

Los scripts shell deben ser POSIX-compliant y pasar `shellcheck`. Reciben parámetros runtime mediante variables de entorno establecidas por Go antes de la ejecución.
