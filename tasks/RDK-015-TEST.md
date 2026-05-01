# RDK-015-TEST — Tests para Idempotency-Key store (multi-engine).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/api/idempotency`
- **Tarea madre:** [`RDK-015`](RDK-015.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del helper `WithIdempotency` contra los tres motores soportados.
- **Cómo:**
  - **Suite común** (parametrizada por motor: Postgres testcontainer, MySQL testcontainer, SQLite in-memory):
    - Primera invocación: ejecuta `fn`, persiste status+body+request_hash.
    - Segunda invocación misma key + mismo body → cache hit, no re-ejecuta.
    - Misma key + body distinto → 409 `idempotency.mismatch`.
    - Key expirado → re-ejecuta como nuevo.
  - **Concurrencia (sólo Postgres y MySQL):**
    - Dos requests con misma key llegan en paralelo (`sync.WaitGroup`); sólo una ejecuta `fn`, la otra recibe respuesta cacheada.
  - **SQLite:**
    - Documentar limitación (no `FOR UPDATE`); test verifica que bajo concurrencia se detecta `SQLITE_BUSY` y se maneja con retry breve o error claro.
  - **Sin `Idempotency-Key`:**
    - Si la política exige el header → 400 `idempotency.required`.
    - Si la política no lo exige → ejecuta `fn` directamente sin persistir.
- **Cuándo:** Junto con `RDK-015`. Bloqueante: `RDK-TEST-000`, `RDK-002-TEST`, `RDK-TEST-LEGACY-004`.

## Criterio de aceptación
- Cobertura ≥ 85% en `shared/api/idempotency`.
- Tests parametrizados corren para los tres motores en CI.
- Test de concurrencia (Postgres/MySQL) pasa con `-race`.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del helper de Idempotency-Key contra los tres motores, para verificar que las garantías difieren por motor exactamente como se documenta y que Postgres/MySQL sí garantizan exactly-once bajo concurrencia.
- **Valor esperado:** El consumidor sabe qué motor le da qué garantías, con respaldo de tests automatizados.
