# Testing en Raidark

Raidark usa `go test` estándar como runner único. Los tests viven junto al código que verifican y los helpers compartidos viven bajo `shared/internal/testutil`.

## Comandos

```bash
make test
make test-integration
make coverage
```

- `make test` ejecuta `go test ./...` y debe mantenerse libre de dependencias externas.
- `make test-integration` ejecuta `go test -tags=integration ./...` para pruebas que requieren servicios externos o containers.
- `make coverage` genera `coverage.out` y `coverage.html`.

La meta inicial de cobertura es `>=70%` por paquete. Las tareas específicas pueden exigir umbrales más altos.

## Layout

Los tests se escriben al lado del código:

```text
shared/foo/bar.go
shared/foo/bar_test.go
```

Los helpers reutilizables se organizan por subpaquete:

- `shared/internal/testutil/db`: bases de datos efímeras para tests. `db.NewSQLite(t, models...)` abre SQLite en memoria, registra cleanup y aplica `AutoMigrate`.
- `shared/internal/testutil/echo`: construcción de `echo.Context` con `httptest`.
- `shared/internal/testutil/fixtures`: lectura de bytes embebidos desde `testdata`.

## Convenciones

- Usa `github.com/stretchr/testify/require` para precondiciones que deben cortar el test.
- Usa `github.com/stretchr/testify/assert` para verificaciones acumulables.
- Nombra tests como `TestNombreFuncion_caso`.
- Prefiere fakes simples escritos a mano antes que mocks cuando la interfaz es pequeña.
- Usa `github.com/stretchr/testify/mock` solo cuando el fake manual agregue más ruido que claridad.
- Usa `net/http/httptest` para handlers HTTP.
- Para Echo, crea contexto con `echo.New().NewContext(req, rec)` o usa `testutil/echo`.

## Integración

Los tests que requieren containers deben usar build tag `integration`:

```go
//go:build integration
// +build integration
```

La librería estándar para esos casos es `github.com/testcontainers/testcontainers-go`. Los tests unitarios no deben arrancar containers ni depender de red.

## Fixtures

Los fixtures viven en `testdata` dentro del paquete que los consume o en `shared/internal/testutil/fixtures` si son reutilizables por varios paquetes. Usa bytes pequeños y determinísticos; evita fixtures generados dinámicamente salvo que el test lo necesite.
