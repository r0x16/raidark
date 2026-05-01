# RDK-TEST-LEGACY-007 — Tests para `shared/cmd` (CLI Cobra: api, dbmigrate, seed, root).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/cmd/{api.go,dbmigrate.go,seed.go,root.go}`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests los comandos Cobra: `api`, `dbmigrate`, `dbmigrate seed` y `root`.
- **Cómo:**
  - **`root.go`**:
    - Help text se imprime correctamente.
    - Subcommands registrados (`api`, `dbmigrate`).
    - Flags globales (si los hay) parsean OK.
  - **`api.go`**:
    - Invocación monta el provider hub y arranca Echo en puerto configurable.
    - Test usa puerto efímero y termina con cancel del contexto.
  - **`dbmigrate.go`**:
    - Invocación corre migraciones contra SQLite in-memory.
    - Subcomando `seed` invoca seeder.
  - Tests usan `cobra.Command.SetArgs([]string{...})` y `bytes.Buffer` para capturar stdout/stderr.
- **Cuándo:** Bloqueante: `RDK-TEST-000`, `RDK-TEST-LEGACY-006`.

## Criterio de aceptación
- Cobertura ≥ 60% en `shared/cmd/` (los entry points no dan mucho margen, pero los flujos principales deben ejercitarse).
- Smoke test: `raidark api` arranca y se cierra limpiamente.

## Fuera de alcance
- Tests end-to-end con `go run ./main` (queda como smoke manual).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre los comandos Cobra, para que un cambio en flags o en el orden de bootstrap no rompa el CLI silenciosamente.
- **Valor esperado:** El CLI queda verificado; agregar comandos nuevos parte de una base estable.
