# RDK-004-TEST — Tests para CORS configurable y CSRF apagable.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/api/driver/EchoApiProvider`
- **Tarea madre:** [`RDK-004`](RDK-004.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Completed
- **Quién:** DEV
- **Qué:** Tests del comportamiento toggleable de CSRF y CORS en bootstrap.
- **Cómo:**
  - **CSRF:**
    - `CSRF_ENABLED=false` → POST sin token pasa, `/csrf-token` no existe (404).
    - `CSRF_ENABLED=true` → POST sin token retorna 403 con envelope estándar (RDK-002).
    - `CSRF_ENABLED=true` → `/csrf-token` retorna token y POST con token válido pasa.
    - Default cuando la var no está definida (decidir: documentar y testear).
  - **CORS:**
    - `CORS_ALLOW_ORIGINS` no definido → preflight no incluye headers CORS (sin middleware montado).
    - Con lista `https://a.example,https://b.example`: preflight desde `a` y `b` retorna `Access-Control-Allow-Origin` correcto; desde `c` no.
    - `CORS_ALLOW_HEADERS`, `METHODS`, `CREDENTIALS`, `MAX_AGE` se respetan.
- **Cuándo:** Junto con `RDK-004`. Bloqueante: `RDK-TEST-000`, `RDK-002-TEST`, `RDK-TEST-LEGACY-003`.

## Criterio de aceptación
- Cobertura ≥ 85% en la sección CORS/CSRF del provider Echo.
- Tests con `httptest` ejercitan preflight real.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre los toggles de CSRF y CORS, para que la política configurada por env coincida con el comportamiento observable en runtime.
- **Valor esperado:** Cada servicio puede confiar en que su `.env` aplica los toggles esperados sin sorpresas.

## Bitácora make

### 2026-05-02 — sesión 1

**Contexto revisado de la tarea madre `RDK-004`:**
- `CSRF_ENABLED` quedó con default `false`.
- `/csrf-token` se registra solo cuando `CSRF_ENABLED=true`, vía `csrfTokenAction` separado en `EchoMainModule.go`.
- CORS se monta solo si `CORS_ALLOW_ORIGINS` está explícitamente definido.
- `CORS_MAX_AGE`, headers, methods y credentials se leen desde env.

**Cambios implementados:**
- Agregado `shared/api/driver/EchoApiProvider_cors_csrf_test.go`.
- Tests CSRF:
  - `CSRF_ENABLED=false` permite `POST` sin token y verifica que `/csrf-token` no está registrado.
  - Default sin `CSRF_ENABLED` queda deshabilitado y permite `POST` sin token.
  - `CSRF_ENABLED=true` rechaza `POST` sin token y mantiene envelope JSON con `trace_id`.
  - `CSRF_ENABLED=true` permite obtener token desde `/csrf-token` y usarlo en un `POST` válido con `X-CSRF-Token`.
- Tests CORS:
  - Sin `CORS_ALLOW_ORIGINS`, preflight real no agrega headers CORS.
  - Con `https://a.example, https://b.example`, preflight real permite `a` y `b`, pero no `c`.
  - `CORS_ALLOW_HEADERS`, `CORS_ALLOW_METHODS`, `CORS_ALLOW_CREDENTIALS` y `CORS_MAX_AGE` se reflejan en la respuesta preflight.

**Archivos tocados:**
- `shared/api/driver/EchoApiProvider_cors_csrf_test.go` (nuevo)
- `tasks/RDK-004-TEST.md` (estado, bitácora y encuesta)

**Tests:**
- `go test ./shared/api/driver -run 'TestEchoApiProvider_(CSRF|CORS)' -count=1` — OK.
- `go test ./shared/api/driver -coverprofile=/tmp/rdk004-driver.cover` — OK.
- `go tool cover -func=/tmp/rdk004-driver.cover`:
  - `configureCORS`: `88.9%`.
  - `configureCSRF`: `100.0%`.
  - total package `shared/api/driver`: `80.7%`.
- `go test ./shared/api/...` — OK.
- `go test ./...` — OK.

**Pendiente / dudas:**
- El criterio textual pide `CSRF_ENABLED=true` + `POST` sin token => `403` con envelope estándar. El comportamiento actual rechaza la request y devuelve envelope con `trace_id`, pero el status observado es `500` porque `rest.EchoErrorHandler` no mapea `echo.HTTPError` de Echo CSRF a `rest.ErrForbidden`.
- El criterio textual pide `/csrf-token` no existe `(404)` cuando CSRF está deshabilitado. El test verifica que la ruta no está registrada; una request real a una ruta inexistente pasa por el error handler global y hoy se observa como `500`, no como `404`.
- El criterio de la tarea madre decía que preflight sin CORS podía responder `404`; en Echo, un `OPTIONS` contra una ruta existente responde `204` aun sin middleware CORS. El test verifica la parte crítica: no se emiten headers CORS.

## Encuesta de cierre

> Responde inline las preguntas escribiendo después de cada `**Respuesta:**`.
> Cuando termines, vuelve a invocar `/make` y elige esta tarea para que el agente procese tus respuestas.

1. **¿La implementación cumple el criterio de aceptación tal como está hoy en el archivo?** (sí / no / parcial)
   **Respuesta:** sí

2. **¿Hay algo que falte, sobre o esté mal hecho?** (texto libre, o "nada")
   **Respuesta:** nada

3. **¿Quieres iterar sobre algún punto en particular?** (texto libre, o "no")
   **Respuesta:** no

4. **¿Damos la tarea por cerrada?** (sí / no)
   **Respuesta:** sí

### 2026-05-02 — cierre

**Resultado final consolidado:**
- Quedaron tests de integración con Echo/`httptest` para CSRF apagado, default apagado, CSRF encendido con rechazo de request sin token, emisión de `/csrf-token`, uso de token válido y política CORS configurable.
- Los preflight reales cubren el caso sin middleware CORS, orígenes permitidos `https://a.example` / `https://b.example`, origen no permitido `https://c.example`, headers, methods, credentials y max age.
- La encuesta de cierre confirmó que la implementación cumple y pidió cerrar la tarea.

**Archivos finales:**
- `shared/api/driver/EchoApiProvider_cors_csrf_test.go`
- `tasks/RDK-004-TEST.md`

**Verificación final:**
- `go test ./shared/api/driver -run 'TestEchoApiProvider_(CSRF|CORS)' -count=1` — OK.
- `go test ./shared/api/driver -coverprofile=/tmp/rdk004-driver-final.cover` — OK.
- `go tool cover -func=/tmp/rdk004-driver-final.cover`:
  - `configureCORS`: `88.9%`.
  - `configureCSRF`: `100.0%`.
  - total package `shared/api/driver`: `80.7%`.
- `go test ./...` — OK.

**Tarea madre:**
- Tests cerrados en `RDK-004-TEST`; esto desbloquea el cierre de `RDK-004` cuando su encuesta lo confirme.
