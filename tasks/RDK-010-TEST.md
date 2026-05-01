# RDK-010-TEST — Tests para JetStream consumer (DLQ, backoff, broadcast).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/events/driver/JetStreamDomainEventsProvider`
- **Tarea madre:** [`RDK-010`](RDK-010.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests de integración del consumer JetStream en sus dos modos.
- **Cómo:**
  - **Modo durable:**
    - Handler exitoso → mensaje ack-eado, no aparece en DLQ.
    - Handler retorna `ErrTransient` → `nak`, se reintenta. Tras `max_deliver` agotado → mensaje a DLQ con headers `delivery_count`, `last_error`, `last_attempt_at`, `original_subject`.
    - Handler retorna `ErrPermanent` → ack inmediato + DLQ.
    - Handler retorna error sin clasificar → tratado como transitorio hasta `max_deliver`, luego DLQ.
    - Backoff: tiempos entre reintentos respetan el schedule configurado (verificar con `time.Now()` y tolerancia ±20%).
  - **Modo broadcast:**
    - Dos consumers con `Mode=Broadcast` en el mismo subject → ambos reciben el mismo mensaje (no compete queue group).
    - Cada instancia tiene su propio durable efímero.
  - **Trace:** handler recibe contexto con `trace_id`/`span_id` reconstruidos desde el header `traceparent` del mensaje.
  - **DLQ subject mapper custom**: respeta función provista.
  - **Concurrency:** múltiples mensajes en paralelo procesados sin race conditions.
- **Cuándo:** Junto con `RDK-010`. Bloqueante: `RDK-TEST-000`, `RDK-008-TEST`, `RDK-009-TEST`.

## Criterio de aceptación
- Cobertura ≥ 80% en consumer.
- Tests con build tag `integration`.
- Test de backoff verifica el schedule completo `1s,4s,15s,60s,300s` con tolerancia (puede testearse acelerando el reloj o reduciendo schedule en config de test).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del consumer cubriendo DLQ, backoff, clasificación de errores y modo broadcast, para que las garantías declaradas se mantengan en presencia de errores y de múltiples instancias.
- **Valor esperado:** El consumer queda verificado en condiciones reales de operación.
