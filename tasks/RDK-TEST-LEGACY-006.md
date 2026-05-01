# RDK-TEST-LEGACY-006 — Tests para `shared/migration` (migrate, seed, controllers).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/migration/{driver,domain}/...`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests la migración (`migrate.go`), el seeder (`seed.go`) y sus controllers.
- **Cómo:**
  - **`migrate.go`**:
    - Con SQLite in-memory: módulo con `GetModel()` no vacío → tablas creadas.
    - Módulo sin modelos → no falla.
    - Múltiples módulos → todas las tablas presentes.
  - **`seed.go`**:
    - Módulo con `GetSeedData()` poblado → datos insertados.
    - Idempotencia: correr seed dos veces no duplica filas (cuando el módulo declara dedup) o falla con error claro (cuando no).
  - **Controllers (`migration/driver/controller/`)**:
    - `DbMigrationController`: invocación con módulos mock, verifica orden y errores.
    - `SeederController`: análogo.
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 75% en `shared/migration/`.
- Tests usan SQLite in-memory; no requieren testcontainers.

## Fuera de alcance
- Tests del nuevo soporte de SQL custom (vive en `RDK-020-TEST`).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre la pipeline de migraciones y seeds, para garantizar que `dbmigrate` no corra dos veces el mismo schema y que los seeds no se rompan al agregar módulos nuevos.
- **Valor esperado:** Las migraciones del framework quedan verificadas antes de extenderlas con SQL custom (RDK-020).
