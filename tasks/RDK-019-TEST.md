# RDK-019-TEST — Tests para particionado mensual de Postgres.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/migration/partitioning`
- **Tarea madre:** [`RDK-019`](RDK-019.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del helper de particionado contra Postgres testcontainer.
- **Cómo:**
  - **Setup:** parent table creado con `PARTITION BY RANGE (created_at)` mediante SQL custom (RDK-020).
  - **Casos:**
    - `CreateMonthlyPartition` para mes M → tabla `parent_YYYY_MM` existe con bordes correctos (`pg_catalog.pg_inherits` + `pg_get_expr`).
    - Insert con `created_at` dentro de M → cae en partición correcta.
    - Insert fuera de cualquier partición declarada → error nativo de Postgres.
    - `DetachMonthlyPartition` → partición desligada, datos siguen accesibles directamente.
    - `PartitionName` produce identificadores deterministas y válidos.
  - **Job:**
    - `StartPartitionAheadJob` con `MonthsAhead=2` y `Interval=100ms` crea M+1 y M+2 si faltan; no recrea las que existen.
    - `cancel()` retornado detiene el job limpiamente.
  - **Validación:**
    - `parentTable` con caracteres SQL no válidos → rechazado.
    - `month` se reduce a primer día UTC.
- **Cuándo:** Junto con `RDK-019`. Bloqueante: `RDK-TEST-000`, `RDK-TEST-LEGACY-004`, `RDK-020-TEST`.

## Criterio de aceptación
- Cobertura ≥ 85% en `shared/migration/partitioning`.
- Tests con build tag `integration`.
- Test del job verifica que no spamea CREATE cuando ya existen las particiones.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del helper de particionado contra Postgres real, para que los servicios consumidores puedan confiar en que `CreateMonthlyPartition` no rompe inserts ni produce particiones con bordes mal calculados.
- **Valor esperado:** El helper queda verificado para tablas de alto volumen mensual.
