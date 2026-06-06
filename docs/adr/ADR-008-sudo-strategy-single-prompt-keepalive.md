# ADR-008: Sudo Strategy — Single Prompt + Keepalive

**Fecha:** 2026-05-23  
**Estado:** Aceptado

## Contexto

Bubble Tea renderiza la TUI en modo raw terminal. Un prompt de contraseña sudo apareciendo en medio del renderizado corrompe la pantalla.

## Decisión

Sudo se solicita una vez al inicio del menú Install mediante `sudo -v`. Un goroutine en segundo plano ejecuta `sudo -v` cada 45 segundos para mantener la sesión activa, con stdout/stderr redirigidos a `/dev/null` para evitar corrupción de la TUI. La sesión expira naturalmente al completar la instalación.

## Consecuencias

Si el usuario cancela el prompt sudo, la instalación se aborta y se retorna al Main Menu. Los módulos que no requieren sudo (la mayoría) no se ven afectados.
