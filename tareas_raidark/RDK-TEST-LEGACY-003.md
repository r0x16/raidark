# RDK-TEST-LEGACY-003 — Tests para `shared/api` (EchoApiProvider, ApplicationBundle, modules, util).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/api/...`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests el `EchoApiProvider`, el `ApplicationBundle`, los `EchoModule` (Main, Auth, ApiMain) y los utils `RequestDataConversion` y `RequestDataSanitization`.
- **Cómo:**
  - **`EchoApiProvider`**:
    - Bootstrap con set mínimo de módulos retorna instancia Echo lista.
    - Health-check responde `200`.
    - `/csrf-token` montado cuando CSRF está activo (toggle existente; el toggle nuevo es `RDK-004-TEST`).
  - **`ApplicationBundle`**:
    - Composición de providers en orden correcto.
    - Falla con error claro si falta provider obligatorio.
  - **Modules (`api/driver/modules/`)**:
    - `EchoMainModule`: registra rutas base esperadas.
    - `EchoAuthModule`: registra endpoints de exchange/refresh/logout.
    - `EchoApiMainModule`: integración mínima con módulos custom registrados.
  - **Utils (`api/driver/util/`)**:
    - `RequestDataConversion`: parseo de query/body a structs, errores claros con tipos no coincidentes.
    - `RequestDataSanitization`: trim, lower, escape, lo que cada función prometa. Tests de edge cases (entrada vacía, unicode, control chars).
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 70% en `shared/api/`.
- Tests de utils con tabla de casos (table-driven).
- Tests de bundle/modules con `httptest` levantando Echo en proceso.

## Fuera de alcance
- Tests de los nuevos middlewares (envelope, paginación, correlación, permisos) — viven en sus respectivas tareas `*-TEST`.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre la capa HTTP base (provider, bundle, modules y utils), para que cambios en Echo o en el orden de bootstrap no rompan silenciosamente las rutas montadas.
- **Valor esperado:** Los servicios construidos sobre Raidark heredan una capa HTTP verificada por tests automatizados.
