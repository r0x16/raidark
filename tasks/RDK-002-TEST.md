# RDK-002-TEST — Tests para convenciones REST (envelope, paginación, correlation id).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/api/rest`
- **Tarea madre:** [`RDK-002`](RDK-002.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
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
