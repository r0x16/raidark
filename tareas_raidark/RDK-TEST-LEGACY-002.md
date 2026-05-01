# RDK-TEST-LEGACY-002 — Tests para `shared/auth` (Casdoor + Array).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/auth/{driver,service,domain}/...`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests los providers `ArrayAuthProvider` y `CasdoorAuthProvider`, los servicios `AuthExchangeService`, `AuthLogoutService`, `AuthRefreshService`, los controllers HTTP y el `GormSessionRepository`.
- **Cómo:**
  - **Domain (`auth/domain/`)**: tests puros de modelos (`User`, `Session`) y validación de claims.
  - **`ArrayAuthProvider`**: validar que retorna usuarios mockeados según mapping configurado, lookup por id, lookup por token.
  - **`CasdoorAuthProvider`**:
    - Verificación de JWT contra cert pegado en config — usar par de claves de prueba en `testdata/`.
    - Token inválido → error tipado.
    - Token expirado → error tipado.
    - `CasdoorError` mapping de errores.
  - **Servicios (`auth/service/`)**:
    - `AuthExchangeService`: mock del provider Casdoor, verificar intercambio code→token y persistencia de sesión.
    - `AuthRefreshService`: refresh con sesión vigente, error con sesión expirada.
    - `AuthLogoutService`: borra/invalida sesión.
  - **Controllers (`auth/driver/controller/`)**:
    - Tests de Echo handler con `httptest`: payloads válidos/ inválidos, códigos HTTP esperados.
  - **Repositorio (`GormSessionRepository`)**:
    - Test contra SQLite in-memory: create, get, delete, expirar.
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 75% en `shared/auth/`.
- `testdata/` con par de claves RSA de juguete y JWT firmado de muestra para tests del provider Casdoor.
- Cada controller tiene al menos un test feliz y uno de error.

## Fuera de alcance
- Levantar un Casdoor real en CI (se mockea la verificación con cert local).
- Tests de los nuevos middlewares `RequirePermission` (esos viven en `RDK-013-TEST`).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre auth (providers, servicios, controllers, repo de sesiones), para no romper la verificación de tokens al actualizar el SDK de Casdoor o al cambiar el modelo de sesión.
- **Valor esperado:** El módulo de auth queda blindado contra regresiones silenciosas en uno de los puntos más sensibles del framework.
