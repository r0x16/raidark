# RDK-003-TEST — Tests para observabilidad base.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/observability`
- **Tarea madre:** [`RDK-003`](RDK-003.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Completed
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

## Bitácora make

### 2026-05-01 — sesión 1

Se implementó la cobertura de observabilidad tomando como base tanto el archivo
`RDK-003-TEST` como las decisiones aplicadas en `RDK-003` sesión 2.

**Cobertura agregada:**
- Logs JSON context-aware:
  - `Logger.FromContext(ctx)` agrega `trace_id`, `span_id`, `service` y `event_id`.
  - Sin campos de trace/event en contexto, el logger no emite claves vacías.
  - Salida verificada con `bytes.Buffer` y `encoding/json`.
  - Niveles (`Debug`, `Info`, `Warning`, `Error`, `Critical`) y `SetLogLevel`.
  - `DataSanitizer` compartido: redacción de campos sensibles, mapas y valores complejos truncados.
- Métricas Prometheus:
  - `HTTPMetrics` incrementa `http_requests_total{status,endpoint}` con route pattern de Echo.
  - Histograma `http_request_duration_ms` acumula muestras de requests reales.
  - Métricas de eventos (`events_published_total`, `events_consumed_total`, `events_redeliveries_total`, `event_processing_duration_ms`, `outbox_pending_gauge`).
  - Tests usan `prometheus/testutil` para counters/gauge y validan histogramas recolectando métricas Prometheus.
- W3C trace context:
  - `traceparent` válido conserva `trace_id`, genera span local nuevo y propaga `tracestate`.
  - Sin `traceparent` genera trace W3C válido.
  - `traceparent` malformado no falla y genera uno nuevo.
  - `X-Correlation-ID` UUIDv7-compatible se promueve a `trace_id`, alineado con la decisión de `RDK-003` / `RDK-002`.
  - `InjectTrace` / `ExtractTrace` conservan `trace_id`, `span_id`, flags y state usando `MapCarrier` (caso NATS / headers map).
  - Validaciones de wire format rechazan versiones no soportadas, IDs cero, uppercase y flags inválidos.
- Wiring derivado de la tarea principal:
  - `PrometheusMetricsProvider` expone `/metrics` desde registry privado.
  - `MetricsProviderFactory` registra provider sólo con `METRICS_ENABLED=true`, conserva `SERVICE_NAME` global.
  - `EchoMetricsModule` monta `/metrics` siguiendo el patrón ApiModule y es no-op sin provider.
  - `EchoApiProvider` registra `HTTPMetrics` cuando existe `MetricsProvider`.

**Archivos tocados:**
- `shared/observability/trace_test.go` (nuevo)
- `shared/observability/metrics_test.go` (nuevo)
- `shared/observability/log/log_test.go` (nuevo)
- `shared/observability/driver/PrometheusMetricsProvider_test.go` (nuevo)
- `shared/providers/driver/MetricsProviderFactory_test.go` (nuevo)
- `shared/api/driver/modules/EchoMetricsModule_test.go` (nuevo)
- `shared/api/driver/EchoApiProvider_observability_test.go` (nuevo)
- `go.mod` / `go.sum` (agrega `github.com/kylelemons/godebug`, dependencia requerida por `prometheus/testutil`)

**Tests:**
- `go test ./shared/observability/... -cover` ✓
  - `shared/observability`: 81.8%
  - `shared/observability/driver`: 100.0%
  - `shared/observability/log`: 89.0%
- `go test ./shared/api/driver -cover` ✓
- `go test ./shared/api/driver/modules -cover` ✓
- `go test ./shared/providers/driver -cover` ✓
- `go test ./shared/...` ✓
- `go test ./...` ✓

**Pendiente / dudas:**
- Sin dudas pendientes. La tarea queda lista para revisión del usuario.

### 2026-05-02 — sesión 2 (correcciones iteración 1)

Se procesó la encuesta de cierre iteración 1. El usuario pidió subir a 100%
los packages:
- `shared/observability/log`
- `shared/observability`

**Correcciones aplicadas:**
- Agregados tests unitarios directos para helpers de contexto:
  - `SetDefaultServiceName` / `GetDefaultServiceName`.
  - getters nil-safe (`GetTraceID`, `GetSpanID`, `GetTraceFlags`, `GetTraceState`, `GetEventID`).
  - `WithServiceName` / `GetServiceName`.
  - `WithEventID` / `GetEventID`.
- Agregados tests directos de trace/propagation:
  - `InjectTrace` sin `span_id`, sin `tracestate`, y flags default.
  - `ExtractTrace` sin `traceparent` y con `traceparent` válido sin `tracestate`.
  - más casos inválidos de `parseTraceParent`.
  - defaults de `formatTraceParent`.
  - normalización y rechazo de `traceIDFromCorrelation`.
- Agregado test de `HTTPMetrics` cuando el handler devuelve error antes de escribir status.
- Ampliados tests de `shared/observability/log`:
  - constructor `New`.
  - `NewWithWriter` con formato text.
  - `FromContext(nil)`.
  - `With(nil)` / `With(map vacío)`.
  - copia de campos estáticos existentes en `FromContext` y `With`.
  - guards de nivel para `Debug` y `Warning`.
  - `DataSanitizer` con valores complejos no truncados.

**Resultado de coverage:**
- `shared/observability/log`: 100.0% ✓
- `shared/observability`: 98.1%

**Límite detectado para `shared/observability` al intentar llegar a 100% sólo con tests:**
- `trace.go:newTraceID` y `trace.go:newSpanID` quedan en 75% porque sus ramas de fallback están después de `crypto/rand.Read`. En Go 1.25, al reemplazar `crypto/rand.Reader` por un reader que falla para probar esa rama, `crypto/rand.Read` aborta el proceso con `fatal error` en vez de devolver el error al caller. Por eso esas líneas no son cubribles por un test unitario normal.
- `middleware_trace.go:W3CTrace` queda en 95.7% por la asignación:
  `if !ok && tc.Flags == "" { tc.Flags = defaultTraceFlags }`.
  Esa rama no es alcanzable con el código actual, porque `resolveTraceContext` siempre retorna `Flags` poblado cuando `ok == false`.
- Para llegar a 100% en `shared/observability` se requiere una corrección de código productivo en la tarea madre (`RDK-003`), por ejemplo:
  - hacer testable/efectivo el fallback de random usando `io.ReadFull(rand.Reader, ...)` en vez de `crypto/rand.Read`, o eliminar el fallback muerto;
  - eliminar la rama redundante de `W3CTrace` o mover el default de flags a un único lugar.
- Esta tarea es `-TEST`, por lo que no se modificó código productivo.

**Archivos tocados en sesión 2:**
- `shared/observability/context_test.go` (nuevo)
- `shared/observability/trace_test.go` (ampliado)
- `shared/observability/metrics_test.go` (ampliado)
- `shared/observability/log/log_test.go` (ampliado)

**Tests:**
- `go test ./shared/observability -coverprofile=/tmp/observability.cover && go tool cover -func=/tmp/observability.cover` ✓
- `go test ./shared/observability/log -coverprofile=/tmp/obslog.cover && go tool cover -func=/tmp/obslog.cover` ✓
- `go test ./shared/observability/... -cover` ✓
- `go test ./...` ✓

**Pendiente / dudas:**
- Queda pendiente decisión del usuario sobre si se autoriza corregir el código productivo en la tarea madre para hacer alcanzable el 100% del paquete raíz.

### 2026-05-02 — cierre

Tarea cerrada por instrucción del usuario tras procesar la iteración 2.

**Resultado final consolidado:**
- Tests de observabilidad base implementados para logs JSON, métricas Prometheus y propagación W3C trace context.
- Coverage final verificado:
  - `shared/observability`: 98.1%.
  - `shared/observability/log`: 100.0%.
  - `shared/observability/driver`: 100.0%.
- La meta original de aceptación (`Cobertura ≥ 80% en shared/observability`) queda cumplida.
- Se aceptó explícitamente no modificar código productivo desde esta tarea `-TEST`; por eso el paquete raíz queda en 98.1% y no en 100%.

**Decisión tomada:**
- No se harán más cambios en esta tarea.
- Se acepta `shared/observability` en 98.1% porque los statements restantes requieren corregir o eliminar ramas no alcanzables en código productivo (`RDK-003`), fuera del alcance de una tarea de tests.

**Archivos finales de tests agregados / modificados:**
- `shared/observability/context_test.go`
- `shared/observability/trace_test.go`
- `shared/observability/metrics_test.go`
- `shared/observability/log/log_test.go`
- `shared/observability/driver/PrometheusMetricsProvider_test.go`
- `shared/providers/driver/MetricsProviderFactory_test.go`
- `shared/api/driver/modules/EchoMetricsModule_test.go`
- `shared/api/driver/EchoApiProvider_observability_test.go`

**Verificación final:**
- `go test ./shared/observability/... -cover` ✓
- `go test ./...` ✓

**Tarea de tests cerrada:**
- `RDK-003-TEST` queda en `Completed`.

## Encuesta de cierre

### Iteración 1 (respondida)

> Responde inline las preguntas escribiendo después de cada `**Respuesta:**`.
> Cuando termines, vuelve a invocar `/make` y elige esta tarea para que el agente procese tus respuestas.

1. **¿La implementación cumple el criterio de aceptación tal como está hoy en el archivo?** (sí / no / parcial)
   **Respuesta:** sí

2. **¿Hay algo que falte, sobre o esté mal hecho?** (texto libre, o "nada")
   **Respuesta:** Necesitoq ue aumentes el coverage más probando las funciones que fueron "probadas indirectamente", quiero que tenga sus pruebas independientes en caso de que cambien las funciones que lo llaman.

3. **¿Quieres iterar sobre algún punto en particular?** (texto libre, o "no")
   **Respuesta:**
   Lleva estos packages a un coverage de 100%
   - shared/observability/log
   - shared/observability

4. **¿Damos la tarea por cerrada?** (sí / no)
   **Respuesta:** no

### Iteración 2 (respondida)

> Surgió una decisión fuera del alcance estricto de esta tarea de tests.
> Responde inline cada pregunta después de `**Respuesta:**` y vuelve a invocar
> `/make` con esta tarea. Hasta resolverla no se cierra la encuesta estándar.

1. **[¿Autorizas una corrección en código productivo para poder llegar a 100% en `shared/observability`?]**
   Contexto: con sólo tests, el paquete raíz subió de 81.8% a 98.1%. Los statements restantes están en ramas no alcanzables: fallback de `crypto/rand.Read` que en Go 1.25 aborta el proceso al fallar, y una asignación redundante de `trace_flags` en `W3CTrace` que `resolveTraceContext` nunca deja sin valor.
   Opciones tentativas:
   - a) Sí, corregir en `RDK-003` / código productivo para hacer esas ramas alcanzables o remover código muerto.
   - b) No, aceptar `shared/observability` en 98.1% y `shared/observability/log` en 100.0%.
   - c) Crear una tarea nueva específica para limpiar esas ramas de observabilidad.
   **Respuesta:** b

2. **¿Quieres hacer alguna corrección o ajuste a lo implementado en esta sesión?** (sí / no — si sí, descríbela en Detalle)
   **Respuesta:** no
   **Detalle:** _

### Iteración 3 (respondida)

> Respuestas default de término solicitadas por el usuario.

1. **¿La implementación cumple el criterio de aceptación tal como está hoy en el archivo?** (sí / no / parcial)
   **Respuesta:** sí

2. **¿Hay algo que falte, sobre o esté mal hecho?** (texto libre, o "nada")
   **Respuesta:** nada

3. **¿Quieres iterar sobre algún punto en particular?** (texto libre, o "no")
   **Respuesta:** no

4. **¿Damos la tarea por cerrada?** (sí / no)
   **Respuesta:** sí
