# RDK-TEST-LEGACY-004 — Tests para `shared/datastore` (GORM Postgres/MySQL/SQLite).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/datastore/driver/connection/{GormPostgresConnection,GormMysqlConnection,GormSqliteConnection}.go`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests las conexiones GORM para los tres motores soportados.
- **Cómo:**
  - **SQLite**: tests unitarios con DB in-memory (`:memory:`) — happy path y error de DSN inválido.
  - **Postgres**: tests de integración con testcontainer Postgres — abre conexión, ejecuta `SELECT 1`, valida que respeta variables `DB_HOST/PORT/USER/PASSWORD/DATABASE`.
  - **MySQL**: análogo a Postgres con testcontainer MySQL.
  - **Lifecycle**: cada test cierra explícitamente la conexión y verifica.
  - **Casos de error**: host inalcanzable, credenciales inválidas, base inexistente — todos retornan errores tipados/útiles.
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 70% en `shared/datastore/`.
- Tests de Postgres y MySQL viven detrás de build tag `integration`.
- Helper `testutil/db` (de `RDK-TEST-000`) consume estos tests indirectamente: si los tests pasan, el helper también funciona.

## Fuera de alcance
- Tests de migraciones (cubiertos por `RDK-TEST-LEGACY-006`).
- Tests del helper de particionado (cubiertos por `RDK-019-TEST`).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre las conexiones GORM, para detectar regresiones cuando se actualice GORM o un driver y para confirmar que las variables de entorno se respetan en cada motor.
- **Valor esperado:** Los tres motores quedan verificados; nuevas tareas que apoyen sobre GORM (Idempotency-Key, outbox, particionado) parten desde una conexión confiable.
