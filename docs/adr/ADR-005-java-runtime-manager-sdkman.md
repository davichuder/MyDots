# ADR-005: Java Runtime Manager — sdkman

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

La gestión de versiones de Java requiere cambiar entre versiones según el proyecto.

## Opciones consideradas

| Opción | Resultado |
| ------ | --------- |
| `brew install java` | Rechazado — sin soporte multi-versión |
| jabba | Rechazado — no mantenido |
| mise | Rechazado — sdkman es el estándar del ecosistema JVM |

## Decisión

sdkman instalado via script curl oficial. Java 25 instalado via `sdk install java 25-open`.

## Consecuencias

sdkman agrega `~/.sdkman` y hooks de shell init a `.zshrc`. Solo se instala si el usuario selecciona Java en el Config Menu.
