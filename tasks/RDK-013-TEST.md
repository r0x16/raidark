# RDK-013-TEST — Tests para helpers de auth basados en permisos.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/auth/permissions`
- **Tarea madre:** [`RDK-013`](RDK-013.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests de los middlewares `RequirePermission`, `RequireAny`, `RequireAll` con extractores Casdoor y Array.
- **Cómo:**
  - **`RequirePermission`:**
    - Sin claims en contexto → 401 con código `auth.unauthenticated`.
    - Claims sin permiso → 403 con código `auth.forbidden`.
    - Claims con permiso → next, status 200.
  - **`RequireAny`:** pasa con cualquier permiso del set; falla con ninguno.
  - **`RequireAll`:** pasa con todos; falla con uno faltante.
  - **Extractor Casdoor:**
    - JWT con claim `permissions=["a","b"]` → extractor devuelve `["a","b"]`.
    - JWT con claim `roles=["admin"]` y mapping rol→permisos en config → extractor devuelve los permisos resueltos.
  - **Extractor Array:** mock de usuario con permisos en memoria.
  - **`RegisterExtractor`:** override custom se invoca correctamente.
  - **Logs:** 401/403 emiten log con `user_id`, `endpoint`, `required_perm`, `trace_id`.
- **Cuándo:** Junto con `RDK-013`. Bloqueante: `RDK-TEST-000`, `RDK-002-TEST`, `RDK-TEST-LEGACY-002`.

## Criterio de aceptación
- Cobertura ≥ 90% en `shared/auth/permissions`.
- Snapshot del envelope de error 401/403 (RDK-002 shape).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre los middlewares de permisos, para que la diferencia 401 vs 403 sea consistente y para que el extractor por default de Casdoor/Array no derive entre versiones.
- **Valor esperado:** La protección de rutas en servicios consumidores tiene comportamiento verificable y reproducible.
