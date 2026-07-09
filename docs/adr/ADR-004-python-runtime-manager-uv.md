# ADR-004: Python Runtime Manager — uv

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

La gestión de versiones de Python ha estado históricamente fragmentada (pyenv, conda, python del sistema, brew python).

## Opciones consideradas

| Opción | Resultado |
| ------ | --------- |
| pyenv | Rechazado — más lento, requiere hooks de shell init, está siendo reemplazado por uv |
| `brew install python@3.12` | Rechazado — funciona pero no da un gestor de versiones para uso futuro |
| conda | Rechazado — excesivo para un instalador de entorno de desarrollo |

## Decisión

`uv` se instala via brew y se usa para instalar Python 3.12 mediante `uv python install 3.12`.

## Consecuencias

uv se convierte en el gestor de runtime de Python. La lógica de instalación sigue el patrón Strategy (ADR-011), haciendo trivial cambiar uv por otro gestor en el futuro.
