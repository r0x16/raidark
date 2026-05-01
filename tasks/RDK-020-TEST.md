# RDK-020-TEST — Tests para migraciones SQL custom.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/migration`
- **Tarea madre:** [`RDK-020`](RDK-020.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del soporte de SQL custom además del auto-migrate GORM.
- **Cómo:**
  - **Setup:** Postgres testcontainer (extensiones reales) + SQLite in-memory (casos no-extensión).
  - **Casos:**
    - Módulo con `GetCustomMigrations()` vacío (default heredado) → `dbmigrate` no falla.
    - Módulo con una migración `CREATE EXTENSION pg_trgm` → tras correr, extensión existe y `schema_migrations` tiene la fila.
    - Re-ejecutar `dbmigrate` → migración no se aplica de nuevo (idempotente).
    - Migración cuyo `Up` cambia (checksum distinto) → `dbmigrate` falla con error explícito y código no-cero.
    - Migración con SQL inválido → no se registra y `dbmigrate` retorna error.
    - Orden lexicográfico de `ID` se respeta entre migraciones del mismo módulo.
    - Múltiples módulos: cada uno corre las suyas, sin colisión por `ID`.
  - **CLI:**
    - `dbmigrate status` lista pendientes y aplicadas con formato esperado (snapshot test sobre stdout capturado).
- **Cuándo:** Junto con `RDK-020`. Bloqueante: `RDK-TEST-000`, `RDK-TEST-LEGACY-006`.

## Criterio de aceptación
- Cobertura ≥ 85% en el código nuevo de migraciones.
- Tests con build tag `integration` para extensiones Postgres reales.
- Snapshot test del output de `dbmigrate status`.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre las migraciones SQL custom, para garantizar que la idempotencia y la detección de checksum cambiado funcionan ante upgrades reales del schema.
- **Valor esperado:** Los servicios consumidores pueden confiar en `dbmigrate` para evolucionar schema sin scripts manuales.
