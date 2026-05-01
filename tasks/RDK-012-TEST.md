# RDK-012-TEST — Tests para dedup en consumer (`processed_events`).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/events/dedup`
- **Tarea madre:** [`RDK-012`](RDK-012.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del wrapper `WithDedup` contra Postgres + NATS embedded.
- **Cómo:**
  - **Idempotencia básica:**
    - Mensaje entregado dos veces (forzando redelivery por `nak` o por reinicio del consumer) → handler ejecuta una sola vez, fila en `processed_events`.
  - **Errores transitorios:**
    - Handler falla con `ErrTransient` → tx rollea, no se inserta en `processed_events`, mensaje se reintenta.
  - **Errores permanentes:**
    - Handler falla con `ErrPermanent` → tx rollea, no se inserta, mensaje va a DLQ.
  - **Race condition:**
    - Dos workers compiten por el mismo `idempotency_key` (mismo mensaje entregado a dos instancias del consumer durable). Sólo uno ejecuta el handler; el otro detecta la fila vía `SELECT ... FOR UPDATE` y hace ack sin re-ejecutar.
  - **Por consumer:**
    - Mismo `idempotency_key` con `consumer_name` distinto → cada consumer ejecuta independientemente (deliberado).
- **Cuándo:** Junto con `RDK-012`. Bloqueante: `RDK-TEST-000`, `RDK-010-TEST`.

## Criterio de aceptación
- Cobertura ≥ 85% en `shared/events/dedup`.
- Tests con build tag `integration`.
- Test de race usa `sync.WaitGroup` para forzar concurrencia real.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del wrapper de dedup, para que la promesa "el handler nunca ve un evento dos veces" se cumpla incluso bajo concurrencia y errores.
- **Valor esperado:** Los handlers de servicios consumidores pueden asumir "exactly-once effective" verificado.
