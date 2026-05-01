# RDK-TEST-LEGACY-010 — Tests para `shared/serverevents` (SSE sobre Echo).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/serverevents/driver/{ServerEventEcho,EventClientEcho}.go`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests el server-sent events sobre Echo (`ServerEventEcho`) y el cliente que los consume (`EventClientEcho`).
- **Cómo:**
  - **`ServerEventEcho`**:
    - Handler emite eventos con formato SSE válido (`data: ...\n\n`).
    - Cierre limpio cuando el cliente desconecta.
    - Múltiples clientes reciben los mismos eventos sin bloquear.
  - **`EventClientEcho`**:
    - Cliente conecta, recibe N eventos, cierra.
    - Reconexión tras error transitorio (si está implementada).
    - Cancelación por contexto detiene la recepción.
  - Tests usan `httptest.Server` + cliente real, con timeout corto.
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 70% en `shared/serverevents/`.
- Tests corren con `-race` sin warnings.
- Smoke test: server emite N eventos y cliente los recibe en orden.

## Fuera de alcance
- Tests de WebSocket (otra capa, vive en `RDK-017-TEST`).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre el SSE existente, para no romper el canal realtime cuando se introduzca el WebSocket helper (RDK-017) ni cuando se cambien middlewares globales.
- **Valor esperado:** El SSE de Raidark queda verificado y coexiste con el nuevo soporte WebSocket sin regresiones.
