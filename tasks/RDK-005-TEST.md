# RDK-005-TEST — Tests para storage adapter (filesystem).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/storage`
- **Tarea madre:** [`RDK-005`](RDK-005.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Completed
- **Quién:** DEV
- **Qué:** Tests del storage adapter abstracto y del driver filesystem.
- **Cómo:**
  - **Put/Get round-trip** con `io.Reader` que NO carga todo en memoria (test con archivo de 50MB en `t.TempDir()`; verificar memoria con `runtime.MemStats` o asegurando que el reader es streaming).
  - **Visibilidad pública:**
    - `Put` con `VisibilityPublic` → archivo bajo `STORAGE_PUBLIC_ROOT`.
    - `PublicURL` retorna `STORAGE_PUBLIC_BASE_URL + key`.
  - **Visibilidad privada:**
    - `Put` con `VisibilityPrivate` → archivo bajo `STORAGE_PRIVATE_ROOT`.
    - `SignedURL` con TTL futuro → válido al pegarle al handler estático interno.
    - `SignedURL` con TTL pasado → 403/expirada.
    - `SignedURL` con HMAC manipulado → 403.
    - `SignedURL` para key inexistente → 404.
  - **Delete + Exists**:
    - `Exists` retorna true tras Put, false tras Delete.
    - `Delete` de key inexistente: comportamiento documentado (idempotente vs error tipado).
  - **Convención de keys**: helper rechaza keys malformados (`../`, absolutos, vacíos).
- **Cuándo:** Junto con `RDK-005`. Bloqueante: `RDK-TEST-000`, `RDK-001-TEST`.

## Criterio de aceptación
- Cobertura ≥ 85% en `shared/storage`.
- Test de tamaño grande (50MB) corre en < 5s y no consume > 50MB de RAM (medido con `MemStats`).
- Tests del handler estático interno: 200 con firma válida, 403 con inválida, 404 inexistente.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del storage adapter para garantizar streaming, signed URLs HMAC y separación pública/privada, antes de que servicios consumidores empiecen a apoyarse en él.
- **Valor esperado:** El driver filesystem queda verificado y la interfaz queda fija para futuros drivers (S3/MinIO).

## Bitácora make

### 2026-05-02 — sesión 1

- Revisada la tarea madre `RDK-005` antes de escribir tests. Los tests incorporan los cambios finales de la tarea principal: driver basado en `afero.BasePathFs`, `EchoStorageModule` opt-in, separación completa entre `StorageProvider` y `DatastoreProvider`, `Delete` idempotente y signed URLs HMAC sobre `"{key}\n{expires_unix}"`.
- Agregados tests de helper de keys en `shared/storage/domain/KeyHelper_test.go`: `BuildKey`, aceptación de keys canónicos con y sin extensión, rechazo de keys vacíos, absolutos, con `.`/`..`, segmentos requeridos vacíos, año/mes inválidos y UUID inválido. También cubre que `a..b` se acepta porque no es traversal por segmento.
- Agregados tests del driver filesystem en `shared/storage/driver/FilesystemStorageProvider_test.go`: round-trip Put/Get con reader streaming de 50MB, verificación de ETag MD5, crecimiento de heap menor a 50MB, separación física entre raíz pública/privada, `PublicURL`, `Exists`, `Delete` idempotente y validación de configuración inválida.
- Agregados tests del handler interno de signed URLs: 200 con firma válida, 403 por URL expirada, 403 por HMAC manipulado y 404 para key inexistente.
- La tarea no modificó código productivo; solo agrega archivos `*_test.go`.

**Archivos tocados:**
- `shared/storage/domain/KeyHelper_test.go` (nuevo)
- `shared/storage/driver/FilesystemStorageProvider_test.go` (nuevo)
- `tasks/RDK-005-TEST.md` (bitácora y encuesta)

**Tests:**
- Bloqueantes verificados: `RDK-TEST-000` (`Completed`) y `RDK-001-TEST` (`Completed`).
- `GOCACHE=/tmp/raidark-gocache-final go test ./shared/storage/... -cover` — exitoso; cobertura `shared/storage/domain` 88.9% y `shared/storage/driver` 85.8%.
- El test de 50MB queda dentro del paquete `shared/storage/driver`, que completó en 0.334s durante la corrida con cobertura.
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.
- Esta tarea es una tarea de tests; no tiene tarea hermana `*-TEST` asociada.

**Pendiente / dudas:**
- Ninguna. La batería cubre el criterio de aceptación definido para `RDK-005-TEST`.

### 2026-05-02 — cierre

- Encuesta Iteración 1 respondida por el usuario con cumplimiento confirmado, sin pendientes, sin iteraciones adicionales y cierre aprobado.
- Queda implementada la batería de tests para `shared/storage`: helper de keys, driver filesystem, round-trip streaming de 50MB, separación de visibilidad pública/privada, `PublicURL`, signed URLs HMAC, handler interno con respuestas 200/403/404, `Exists`, `Delete` idempotente y validación de configuración.
- La tarea se mantuvo dentro del alcance de tests y no modificó código productivo.
- Esta tarea es una tarea de tests y no tiene tarea hermana `*-TEST` asociada.

**Archivos finales relevantes:**
- `shared/storage/domain/KeyHelper_test.go`
- `shared/storage/driver/FilesystemStorageProvider_test.go`
- `tasks/RDK-005-TEST.md`
- Tarea madre verificada: `tasks/RDK-005.md`

**Verificación final:**
- `GOCACHE=/tmp/raidark-gocache-final go test ./shared/storage/... -cover` — exitoso; cobertura `shared/storage/domain` 88.9% y `shared/storage/driver` 85.8%.
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.

## Encuesta de cierre

### Iteración 1

> Responde inline las preguntas escribiendo después de cada `**Respuesta:**`.
> Cuando termines, vuelve a invocar `/make` y elige esta tarea para que el agente procese tus respuestas.

1. **¿La implementación cumple el criterio de aceptación tal como está hoy en el archivo?** (sí / no / parcial)
   **Respuesta:** sí

2. **¿Hay algo que falte, sobre o esté mal hecho?** (texto libre, o "nada")
   **Respuesta:** nada

3. **¿Quieres iterar sobre algún punto en particular?** (texto libre, o "no")
   **Respuesta:** no

4. **¿Damos la tarea por cerrada?** (sí / no)
   **Respuesta:** sí
