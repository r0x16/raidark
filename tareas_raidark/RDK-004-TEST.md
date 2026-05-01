# RDK-004-TEST — Tests para CORS configurable y CSRF apagable.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/api/driver/EchoApiProvider`
- **Tarea madre:** [`RDK-004`](RDK-004.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
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
