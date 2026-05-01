# RDK-TEST-000 — Estrategia y toolchain de testing para Raidark.

## Ubicación
- **Repositorio:** Raidark (`github.com/r0x16/Raidark`)
- **Componente:** Cross-cutting (`scripts/`, `Makefile`, CI, `docs/testing.md`).
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Completed
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

## Bitácora make

### 2026-04-30 — sesión 1

- Creado `Makefile` con targets `test`, `test-integration` y `test-cover`.
- Los targets de test fijan `GOTOOLCHAIN` al `GOVERSION` resuelto por Go para evitar mezcla entre el toolchain local `go1.25.6` y el `toolchain go1.25.9` declarado en `go.mod`.
- Agregado `github.com/stretchr/testify v1.11.1` como dependencia directa para tests.
- Creado `shared/internal/testutil/db` con helper `NewSQLite(t, models...)` para SQLite in-memory con cleanup y `AutoMigrate`.
- Creado `shared/internal/testutil/echo` con builder de contexto Echo basado en `httptest`.
- Creado `shared/internal/testutil/fixtures` con fixtures embebidas desde `testdata`.
- Agregados smoke tests para `db`, `echo` y `fixtures`.
- Creada documentación `docs/testing.md` con comandos, layout, convenciones, fakes/mocks, coverage e integración.
- Creado workflow `.github/workflows/test.yml` que corre `go test ./...` en PR y push a `main`, con job de integración documentado como skeleton opt-in.
- Agregado `coverage.html` a `.gitignore`; `coverage.out` ya queda cubierto por `*.out`.

**Archivos tocados:**
- `.github/workflows/test.yml` (nuevo)
- `.gitignore`
- `Makefile` (nuevo)
- `docs/testing.md` (nuevo)
- `go.mod`
- `shared/internal/testutil/db/db.go` (nuevo)
- `shared/internal/testutil/db/db_test.go` (nuevo)
- `shared/internal/testutil/echo/echo.go` (nuevo)
- `shared/internal/testutil/echo/echo_test.go` (nuevo)
- `shared/internal/testutil/fixtures/fixtures.go` (nuevo)
- `shared/internal/testutil/fixtures/fixtures_test.go` (nuevo)
- `shared/internal/testutil/fixtures/testdata/sample.txt` (nuevo)

**Tests:**
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make test-integration` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make test-cover` — exitoso; generó `coverage.out` y `coverage.html` ignorados por git.
- Esta tarea es la base de testing del proyecto; no tiene tarea hermana `*-TEST` asociada.

**Pendiente / dudas:**
- Ninguna. Implementación completa según criterio de aceptación.

### 2026-05-01 — sesión 2

- Procesadas respuestas de la encuesta de cierre.
- Anotación del usuario: el target `test-cover` fue renombrado manualmente a `coverage` en `Makefile`; `docs/testing.md` ya referencia `make coverage`.
- Agregados comentarios de paquete/archivo y comentarios por función, método o test en los helpers de `shared/internal/testutil`.
- La tarea sigue en `In Progress` porque el usuario respondió que no se debe cerrar hasta corregir lo indicado.

**Archivos tocados:**
- `shared/internal/testutil/db/db.go`
- `shared/internal/testutil/db/db_test.go`
- `shared/internal/testutil/echo/echo.go`
- `shared/internal/testutil/echo/echo_test.go`
- `shared/internal/testutil/fixtures/fixtures.go`
- `shared/internal/testutil/fixtures/fixtures_test.go`

**Tests:**
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make test-integration` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make coverage` — exitoso; generó `coverage.out` y `coverage.html` ignorados por git.
- Esta tarea es la base de testing del proyecto; no tiene tarea hermana `*-TEST` asociada.

**Pendiente / dudas:**
- Ninguna. Queda pendiente la confirmación de cierre del usuario.

### 2026-05-01 — cierre

- Encuesta Iteración 2 respondida por el usuario con cumplimiento confirmado, sin pendientes y cierre aprobado.
- Queda establecida la base de testing de Raidark: helpers reutilizables en `shared/internal/testutil`, comandos `make test`, `make test-integration` y `make coverage`, documentación en `docs/testing.md` y workflow CI unitario en `.github/workflows/test.yml`.
- Los smoke tests de `db`, `echo` y `fixtures` verifican que los helpers funcionan.
- La tarea no tiene tarea de tests hermana porque es la tarea fundacional de testing.

**Archivos finales relevantes:**
- `.github/workflows/test.yml`
- `.gitignore`
- `Makefile`
- `docs/testing.md`
- `go.mod`
- `shared/internal/testutil/db/db.go`
- `shared/internal/testutil/db/db_test.go`
- `shared/internal/testutil/echo/echo.go`
- `shared/internal/testutil/echo/echo_test.go`
- `shared/internal/testutil/fixtures/fixtures.go`
- `shared/internal/testutil/fixtures/fixtures_test.go`
- `shared/internal/testutil/fixtures/testdata/sample.txt`

**Verificación final:**
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make test-integration` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make coverage` — exitoso.

## Encuesta de cierre

### Iteración 1 (respondida)

> Responde inline las preguntas escribiendo después de cada `**Respuesta:**`.
> Cuando termines, vuelve a invocar `/make` y elige esta tarea para que el agente procese tus respuestas.

1. **¿La implementación cumple el criterio de aceptación tal como está hoy en el archivo?** (sí / no / parcial)
   **Respuesta:** Sí

2. **¿Hay algo que falte, sobre o esté mal hecho?** (texto libre, o "nada")
   **Respuesta:** nada, ya cambié lo que estimé apropiado, en este caso solo cambié test-cover por coverage en Makefile (Crealo como anotación)

3. **¿Quieres iterar sobre algún punto en particular?** (texto libre, o "no")
   **Respuesta:** Si, añade comentarios al código para entender cuál es la función de cada archivo y cada función/método

4. **¿Damos la tarea por cerrada?** (sí / no)
   **Respuesta:** no, hasta que se corrija lo que indiqué en las respuestas anteriores.

### Iteración 2

> Responde inline las preguntas escribiendo después de cada `**Respuesta:**`.
> Cuando termines, vuelve a invocar `/make` y elige esta tarea para que el agente procese tus respuestas.

1. **¿La implementación cumple el criterio de aceptación tal como está hoy en el archivo?** (sí / no / parcial)
   **Respuesta:** sí

2. **¿Hay algo que falte, sobre o esté mal hecho?** (texto libre, o "nada")
   **Respuesta:** no

3. **¿Quieres iterar sobre algún punto en particular?** (texto libre, o "no")
   **Respuesta:** no

4. **¿Damos la tarea por cerrada?** (sí / no)
   **Respuesta:** sí
