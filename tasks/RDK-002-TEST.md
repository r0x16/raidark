# RDK-002-TEST — Tests para convenciones REST (envelope, paginación, correlation id).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/api/rest`
- **Tarea madre:** [`RDK-002`](RDK-002.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** In Progress
- **Quién:** DEV
- **Qué:** Implementar y extender la batería de tests del criterio de `RDK-002`.
- **Cómo:**
  - **Envelope de error:**
    - `RenderError` produce JSON con shape estable (snapshot test contra fixture).
    - Mapper devuelve status correcto para cada sentinel (`ErrNotFound`→404, `ErrConflict`→409, `ErrForbidden`→403, `ErrValidation`→400, `ErrTransient`→503, `ErrPermanent`→500).
    - Error desconocido → 500 con `code="internal.unexpected"` y mensaje genérico (no expone el error original al cliente).
    - `details` opcional: presente cuando se provee, ausente cuando no.
  - **Paginación:**
    - `Page[T]` round-trip: encode → decode → mismo cursor.
    - Cursor opaco resiste tampering: cambiar un byte invalida.
    - Limit por default si la query no lo trae; limit max-clamped.
  - **CorrelationID middleware:**
    - Header presente → se respeta.
    - Header ausente → genera UUIDv7 y lo retorna en response.
    - El valor se inyecta en `echo.Context` y es accesible vía helper.
  - **Tests con Echo + httptest.**
- **Cuándo:** Junto con `RDK-002`. Bloqueante: `RDK-TEST-000`, `RDK-001-TEST`.

## Criterio de aceptación
- Cobertura ≥ 90% en `shared/api/rest`.
- Snapshot tests del envelope cubren 6 sentinels.
- Test de tampering del cursor pasa.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre las convenciones REST, para que la forma de errores y páginas no derive accidentalmente entre versiones.
- **Valor esperado:** Clientes consumidores pueden parsear respuestas con código compartido sin sorpresas tras upgrades.

## Bitácora make

### 2026-05-01 — sesión 1

- Agregada batería de tests públicos para `shared/api/rest`:
  - `errors_test.go`: snapshots del envelope para los 6 sentinels, mapper con sentinel wrapeado, error desconocido genérico, `details` opcional, `RESTError.Error()` y bridge `EchoErrorHandler`.
  - `pagination_test.go`: shape JSON de `Page[T]`, round-trip de cursor, cursor tampered inválido, payload inválido y `ClampLimit` para default/max.
  - `correlation_test.go`: middleware con header presente, header ausente con UUIDv7 generado, response header y acceso vía `GetCorrelationID`.
- Agregada fixture `shared/api/rest/testdata/error_envelope_snapshots.json` para snapshot del envelope.
- Verificación enfocada:
  - `go test ./shared/api/rest -cover` → `coverage: 93.2% of statements`.
- Verificación ampliada:
  - `go test ./shared/...` → pasa.

**Archivos tocados:**
- `shared/api/rest/errors_test.go` (nuevo)
- `shared/api/rest/pagination_test.go` (nuevo)
- `shared/api/rest/correlation_test.go` (nuevo)
- `shared/api/rest/testdata/error_envelope_snapshots.json` (nuevo)
- `tasks/RDK-002-TEST.md` (estado, bitácora y encuesta)

**Tests:**
- Esta tarea es la tarea de tests asociada a `RDK-002`.
- `go test ./shared/api/rest -cover` pasa con cobertura `93.2%`.
- `go test ./shared/...` pasa.

**Pendiente / dudas:**
- Ninguna.

## Encuesta de cierre

> Responde inline las preguntas escribiendo después de cada `**Respuesta:**`.
> Cuando termines, vuelve a invocar `/make` y elige esta tarea para que el agente procese tus respuestas.

1. **¿La implementación cumple el criterio de aceptación tal como está hoy en el archivo?** (sí / no / parcial)
   **Respuesta:** sí

2. **¿Hay algo que falte, sobre o esté mal hecho?** (texto libre, o "nada")
   **Respuesta:** Verifica la tarea principal `RDK-002` y verifica si las modificaciones hechas sobre esta no añaden otras funciones que deben ser también probadas. ya uqe el coverage de 93% me dice que hay cosas que no consideraste.

3. **¿Quieres iterar sobre algún punto en particular?** (texto libre, o "no")
   **Respuesta:** no

4. **¿Damos la tarea por cerrada?** (sí / no)
   **Respuesta:** no
