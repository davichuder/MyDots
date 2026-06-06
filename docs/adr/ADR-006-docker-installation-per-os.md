# ADR-006: Docker Installation per OS

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

Docker Engine no puede instalarse via brew en macOS (brew solo provee el CLI client). En Linux, Docker requiere su propio repositorio apt y modifica archivos del sistema fuera de `$HOME`.

## Decisión

- **Darwin:** `brew install --cask docker-desktop` (plan Personal, gratuito).
- **Ubuntu native / WSL2:** Repositorio apt oficial de Docker + `docker-ce` + `docker-ce-cli` + `containerd.io`. Usuario agregado al grupo `docker`. Escribe en `/etc/apt/sources.list.d/docker.list` y `/etc/group` — excepción documentada a NFR-06.

## Consecuencias

En Darwin, Docker Desktop requiere un primer lanzamiento manual para aceptar términos. En WSL2, Docker requiere systemd (verificado antes de instalar).
