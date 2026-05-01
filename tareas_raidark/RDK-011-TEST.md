# RDK-011-TEST — Tests para outbox transaccional.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/events/outbox`
- **Tarea madre:** [`RDK-011`](RDK-011.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del helper `Enqueue` y del dispatcher background.
- **Cómo:**
  - **Setup:** Postgres testcontainer + NATS embedded.
  - **Atomicidad:**
    - Tx con `Enqueue` que commitea → fila en `outbox_events` con `status=pending`.
    - Tx con `Enqueue` que rollea → no queda fila.
    - Mutación de entidad de negocio + `Enqueue` en misma tx: ambas o ninguna.
  - **Dispatcher:**
    - Fila `pending` → tras tick del dispatcher pasa a `published` con `published_at` set.
    - Publish que falla (NATS down) → `attempts++`, `last_error` poblado, queda `pending`.
    - Tras `MAX_OUTBOX_ATTEMPTS` agotados → `status=failed`.
    - Mensaje publicado tiene envelope correcto en JetStream (header `Nats-Msg-Id`, etc.).
  - **Orden por aggregate:**
    - Tres eventos del mismo `aggregate_id` → publicados en orden de `created_at`.
    - Eventos de aggregates distintos → no se bloquean entre sí (paralelizan o intercalan; verificar latencia agregada).
  - **Concurrency:**
    - Dos dispatchers en el mismo proceso (no se debe permitir, o se coordina) — comportamiento documentado y testeado.
- **Cuándo:** Junto con `RDK-011`. Bloqueante: `RDK-TEST-000`, `RDK-008-TEST`, `RDK-009-TEST`.

## Criterio de aceptación
- Cobertura ≥ 85% en `shared/events/outbox`.
- Tests con build tag `integration`.
- Test de atomicidad rollea con un panic dentro de la tx y verifica que no quede fila.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del outbox cubriendo atomicidad, reintentos y orden por aggregate, para que la garantía "no se pierde el evento" se mantenga ante fallas reales de red/NATS.
- **Valor esperado:** El outbox queda verificado como pieza clave de "exactly-once effective" en publish.
