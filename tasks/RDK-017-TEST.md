# RDK-017-TEST — Tests para soporte WebSocket.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/ws`
- **Tarea madre:** [`RDK-017`](RDK-017.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del helper `Upgrade` y del tipo `WSConn`.
- **Cómo:**
  - **Upgrade:**
    - Request con headers correctos de WebSocket → upgrade exitoso.
    - Request sin headers de upgrade → 400/426.
    - Sin claims de auth (cuando se requiere) → 401 con envelope estándar.
  - **Echo bidireccional:**
    - Cliente envía text frame → server lo recibe con `Read`.
    - Server envía binary frame → cliente lo recibe.
  - **Ping/pong:**
    - Cliente que no responde a ping en `2 * PingInterval` → conexión cerrada con código `1011` o equivalente.
  - **Cierre:**
    - `ctx.Cancel()` cierra con código `1001 Going Away`.
    - Cliente cierra → `Read` retorna error tipado de cierre limpio.
  - **Límites:**
    - Mensaje > `MaxMessageBytes` → close con código `1009`.
  - **Concurrency:**
    - 100 conexiones simultáneas, cada una recibe ping/pong correctamente. `-race` limpio.
- **Cuándo:** Junto con `RDK-017`. Bloqueante: `RDK-TEST-000`, `RDK-TEST-LEGACY-003`.

## Criterio de aceptación
- Cobertura ≥ 80% en `shared/ws`.
- Test de 100 conexiones simultáneas pasa con `-race` y termina en < 10s.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del helper WebSocket cubriendo upgrade, ping/pong, cierre limpio y concurrencia, para que servicios realtime sobre Raidark hereden estabilidad verificable.
- **Valor esperado:** El helper WS queda verificado para uso productivo en gateways realtime.
