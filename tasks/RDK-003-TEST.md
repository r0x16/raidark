# RDK-003-TEST — Tests para observabilidad base.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/observability`
- **Tarea madre:** [`RDK-003`](RDK-003.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests de logs JSON, métricas Prometheus y propagación W3C trace.
- **Cómo:**
  - **Logs:**
    - `log.FromContext(ctx)` con `trace_id` en contexto → log emitido lo incluye.
    - Sin `trace_id` → log se emite sin campo (no panic, no string vacío).
    - Adapter JSON: cada línea es JSON parseable con campos esperados.
    - Niveles respetan `LOG_LEVEL`.
  - **Métricas Prometheus:**
    - Tras N requests HTTP, `http_requests_total` incrementa N con labels correctos.
    - Histogram de duración acumula muestras en buckets esperados.
    - Counters de eventos (`events_published_total`, `events_consumed_total`) se incrementan al instrumentarlos.
    - `/metrics` retorna formato Prometheus parseable.
  - **W3C trace:**
    - Middleware con `traceparent` válido entrante propagá `trace_id`/`span_id` correctos.
    - Sin `traceparent` → genera uno nuevo con formato W3C correcto.
    - `traceparent` malformado → genera uno nuevo, no falla.
    - Helper que serializa headers para NATS conserva `trace_id` round-trip.
- **Cuándo:** Junto con `RDK-003`. Bloqueante: `RDK-TEST-000`, `RDK-002-TEST`.

## Criterio de aceptación
- Cobertura ≥ 80% en `shared/observability`.
- Tests usan `prometheus/testutil` para verificar métricas.
- Logs verificados con `bytes.Buffer` y parseados como JSON.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre logs, métricas y trace, para garantizar que la observabilidad transversal no pierda campos al cambiar middlewares globales.
- **Valor esperado:** La capa de observabilidad queda verificada para todos los servicios construidos sobre Raidark.
