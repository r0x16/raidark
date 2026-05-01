# RDK-TEST-000 — Estrategia y toolchain de testing para Raidark.

## Ubicación
- **Repositorio:** Raidark (`github.com/r0x16/Raidark`)
- **Componente:** Cross-cutting (`scripts/`, `Makefile`, CI, `docs/testing.md`).
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Establecer la base de testing de Raidark: convenciones de layout, librerías, helpers compartidos, comando único para correr la batería completa y CI mínimo. Hoy Raidark **no tiene un solo `*_test.go`**; este es el primer paso obligatorio para que las demás tareas de testing apoyen sobre algo común.
- **Cómo:**
  - **Toolchain:**
    - `go test` estándar (Go 1.25 ya en `go.mod`).
    - Aserciones con `github.com/stretchr/testify/assert` y `require`.
    - Mocks con `github.com/stretchr/testify/mock` o interfaces fake escritas a mano (preferir las segundas cuando son simples).
    - Containers con `github.com/testcontainers/testcontainers-go` para Postgres, MySQL, NATS y mailpit.
    - HTTP fake server con `net/http/httptest`.
    - Echo: usar `echo.New().NewContext(req, rec)` con `httptest.NewRecorder` para tests de middleware/handler.
  - **Layout:**
    - Tests viven al lado del código (`shared/foo/bar.go` ↔ `shared/foo/bar_test.go`).
    - Helpers compartidos en `shared/internal/testutil/` (paquete `testutil`):
      - `testutil/db` — spinners de Postgres/MySQL/SQLite con migración aplicada.
      - `testutil/nats` — embedded NATS server (`nats-server` empaquetado o testcontainer JetStream).
      - `testutil/echo` — builder de `echo.Context` con auth/claims preconfigurados.
      - `testutil/fixtures` — bytes de fixtures (jpeg con EXIF, pdf con JS, etc.).
  - **Tags de build:** `// +build integration` para tests que requieran containers; `go test ./...` corre sólo unitarios; `go test -tags=integration ./...` corre todo.
  - **Comando único:** agregar a `raidark.sh` (o crear `Makefile`) los targets:
    - `make test` (unitarios).
    - `make test-integration` (con tag).
    - `make test-cover` (con `-cover` y reporte HTML).
  - **Coverage objetivo:** declarar `>=70%` por paquete como meta inicial; los criterios de cada tarea de testing afinan dónde se exige más.
  - **CI:** workflow GitHub Actions (`.github/workflows/test.yml`) que corre `go test ./...` en cada PR. Job adicional opcional con tag `integration` que arranca containers; si no hay infra de CI todavía, dejar el workflow como skeleton sin habilitar.
  - **Documentación:** `docs/testing.md` con:
    - Cómo correr cada subset.
    - Cómo escribir tests usando los helpers.
    - Convención de nombres (`TestNombreFuncion_caso`).
    - Política de fakes vs mocks.

## Criterio de aceptación
- `shared/internal/testutil/` existe con al menos los subpaquetes `db`, `echo`, `fixtures` poblados con helpers reutilizables.
- `Makefile` (o equivalente) con `test`, `test-integration`, `test-cover`.
- `docs/testing.md` cubre toolchain, layout y convenciones.
- Workflow CI en repo (aunque sea skeleton).
- Un test "smoke" en cada subpaquete de `testutil` que demuestra que el helper funciona.

## Fuera de alcance
- Mutation testing.
- Property-based testing (queda como mejora futura si surge necesidad).
- Benchmarks (cada tarea individual decide si los necesita).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero una base común de testing (toolchain, helpers, comandos, CI) antes de escribir tests por funcionalidad, para no inventar convenciones distintas en cada paquete y para que todos los tests futuros se ejecuten con un único comando.
- **Valor esperado:** Cualquier tarea de testing posterior empieza con una base estable y predecible; agregar tests a un paquete nuevo no obliga a re-decidir toolchain.
