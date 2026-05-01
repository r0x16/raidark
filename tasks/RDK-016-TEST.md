# RDK-016-TEST — Tests para cliente HTTP inter-servicios.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/httpx`
- **Tarea madre:** [`RDK-016`](RDK-016.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del `ServiceClient` con `httptest.Server`.
- **Cómo:**
  - **Propagación:**
    - Contexto con `Authorization` token → header presente en request salida.
    - Contexto con `trace_id` → headers `X-Correlation-ID` y `traceparent` propagados.
    - Sin Authorization en contexto → request va sin el header (no inventa).
  - **Retries:**
    - Servidor que retorna `503` dos veces y luego `200` → retry con backoff esperado, request final exitosa.
    - Servidor que retorna `400` → no retry, error tipado.
    - Servidor que cuelga > timeout → retry hasta agotar.
    - `ctx.Cancel()` durante el wait de backoff → abort inmediato.
  - **Mapeo de errores:**
    - Servidor responde envelope `RESTError` (RDK-002) con status 404 → cliente devuelve `*RESTError` tipado con `code="..."`.
    - Servidor responde body sin shape estándar → error genérico con `Body` raw accesible.
  - **Helpers genéricos:**
    - `GetJSON[T any]` decodifica response correctamente.
    - `PostJSON[T any]` serializa body y decodifica response.
- **Cuándo:** Junto con `RDK-016`. Bloqueante: `RDK-TEST-000`, `RDK-002-TEST`, `RDK-003-TEST`.

## Criterio de aceptación
- Cobertura ≥ 85% en `shared/httpx`.
- Test de retries verifica tiempos de backoff con tolerancia.
- Test de cancelación pasa con `-race`.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del cliente HTTP inter-servicios, para que retries inteligentes y propagación de auth/trazas se mantengan en cualquier upgrade.
- **Valor esperado:** Las llamadas inter-servicios quedan robustas y observables por construcción.
