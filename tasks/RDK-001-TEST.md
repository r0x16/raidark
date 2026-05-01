# RDK-001-TEST — Tests para UUIDv7 helper.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/ids`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing
- **Tarea madre:** [`RDK-001`](RDK-001.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Completed
- **Quién:** DEV
- **Qué:** Implementar la batería de tests definida en el criterio de aceptación de `RDK-001` y extenderla con tests adicionales de robustez.
- **Cómo:**
  - **Unitarios:**
    - `NewV7()` produce UUIDs con bits de versión `7` y variante RFC.
    - Monotonicidad lógica: dos llamadas consecutivas tienen timestamp embebido no decreciente.
    - `IsValidV7` rechaza: cadena vacía, formato malformado, UUIDv4, UUIDs con variante incorrecta.
    - `IsValidV7` acepta UUIDs generados por `NewV7`.
  - **Concurrencia:**
    - `NewV7` desde N goroutines no produce colisiones (ejecutar 100k IDs y verificar unicidad).
    - Tests con `-race`.
  - **GORM integración:**
    - SQLite in-memory con un modelo cuyo PK es `ids.UUIDv7`.
    - Insertar registro sin asignar PK → BeforeCreate poblá un UUIDv7 válido.
    - Insertar registro con PK provisto → respeta el valor.
- **Cuándo:** Junto con `RDK-001`. Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 90% en `shared/ids`.
- Test de concurrencia genera ≥ 100k IDs únicos sin colisiones.
- Test GORM verifica autogeneración y respeto a valor explícito.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests exhaustivos para el helper UUIDv7, para garantizar que sea seguro como base de PKs e identificadores de mensajes.
- **Valor esperado:** El generador queda verificado para uso masivo en todos los servicios consumidores.

## Bitácora make

### 2026-05-01 — sesión 1

- Agregada batería de tests en `shared/ids/uuidv7_test.go` para generación UUIDv7, bits de versión, variante RFC, monotonía del timestamp embebido, validación de entradas inválidas y aceptación de IDs generados.
- Agregado test concurrente con 100.000 UUIDv7 generados desde 50 goroutines y verificación de unicidad sin colisiones.
- Agregados tests de integración GORM con SQLite in-memory usando `shared/datastore/domain.BaseModel`, cubriendo autogeneración de PK UUIDv7 y respeto de PK explícita.
- Agregados tests para `Scan` y `Value` de `ids.UUIDv7`, incluyendo lectura desde `string`, lectura desde `[]byte` y rechazo de tipos incompatibles.
- Corregido `ids.IsValidV7` para exigir variante RFC 4122 además de versión 7, porque el criterio de aceptación exige rechazar UUIDs con variante incorrecta.
- Agregada cabecera de paquete en `shared/ids/uuidv7.go` para documentar responsabilidad del paquete modificado.

**Archivos tocados:**
- `shared/ids/uuidv7.go`
- `shared/ids/uuidv7_test.go` (nuevo)

**Tests:**
- `GOCACHE=/tmp/raidark-gocache-final go test ./shared/ids -cover` — exitoso; cobertura `92.9%` en `shared/ids`.
- `GOCACHE=/tmp/raidark-gocache-final go test -race ./shared/ids` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.
- Esta tarea es una tarea de tests; no tiene tarea hermana `*-TEST` asociada.

**Pendiente / dudas:**
- Ninguna. La batería cubre el criterio de aceptación definido para `RDK-001-TEST`.

### 2026-05-01 — correcciones iteración 1

- Procesada corrección del usuario: los comentarios de código deben estar en inglés.
- Procesada observación de alcance: una tarea `*-TEST` no debe modificar funcionalidad productiva.
- Convertidos a inglés los comentarios de `shared/ids/uuidv7_test.go`.
- Retirado el cambio funcional que se había hecho en `shared/ids/uuidv7.go`; la tarea queda limitada a tests.
- El test `TestIsValidV7_rejectsInvalidInputs/wrong-variant` queda como detector del incumplimiento actual de `RDK-001`: `IsValidV7` todavía acepta UUIDv7 con variante no RFC.

**Archivos tocados:**
- `shared/ids/uuidv7_test.go`

**Tests:**
- `GOCACHE=/tmp/raidark-gocache-final go test ./shared/ids -run TestIsValidV7_rejectsInvalidInputs -count=1` — falla en `wrong-variant`, exponiendo un defecto de implementación fuera del alcance de esta tarea de tests.

**Pendiente / dudas:**
- Para que la batería completa pase, `RDK-001` debe corregir `IsValidV7` para exigir variante RFC además de versión 7.

### 2026-05-01 — sesión 2

- Procesada respuesta de la encuesta anterior: no hay comentarios adicionales sobre la batería de tests.
- Verificada la corrección aplicada en la tarea principal `RDK-001`: `ids.IsValidV7` ahora exige versión 7 y variante RFC 4122.
- La batería de `shared/ids` vuelve a pasar completa con cobertura superior al objetivo.

**Archivos tocados:**
- `tasks/RDK-001-TEST.md`

**Tests:**
- `GOCACHE=/tmp/raidark-gocache-final go test ./shared/ids -cover` — exitoso; cobertura `92.9%` en `shared/ids`.
- `GOCACHE=/tmp/raidark-gocache-final go test -race ./shared/ids` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.
- Esta tarea es una tarea de tests; no tiene tarea hermana `*-TEST` asociada.

**Pendiente / dudas:**
- Ninguna. Queda pendiente la confirmación de cierre del usuario.

### 2026-05-01 — cierre

- Encuesta Iteración 2 respondida por el usuario con cumplimiento confirmado, sin pendientes, sin iteraciones adicionales y cierre aprobado.
- Queda implementada la batería de tests para `shared/ids`: generación UUIDv7, versión, variante RFC, monotonía lógica, validación de entradas inválidas, aceptación de IDs generados, concurrencia con 100.000 IDs únicos, contratos SQL de `ids.UUIDv7` e integración GORM con `shared/datastore/domain.BaseModel`.
- La corrección productiva necesaria quedó aplicada en la tarea madre `RDK-001`: `IsValidV7` ahora valida versión 7 y variante RFC 4122, permitiendo que esta batería pase completa.
- Esta tarea es una tarea de tests y no tiene tarea hermana `*-TEST` asociada.

**Archivos finales relevantes:**
- `shared/ids/uuidv7_test.go`
- `shared/ids/uuidv7.go`
- `tasks/RDK-001-TEST.md`
- `tasks/RDK-001.md`

**Verificación final:**
- `GOCACHE=/tmp/raidark-gocache-final go test ./shared/ids -cover` — exitoso; cobertura `92.9%` en `shared/ids`.
- `GOCACHE=/tmp/raidark-gocache-final go test -race ./shared/ids` — exitoso.
- `GOCACHE=/tmp/raidark-gocache-final make test` — exitoso.

## Encuesta de cierre

### Iteración 1 (respondida)

> La implementación de esta sesión quedó registrada en la bitácora. Los tests aún no están listos para cerrar la tarea, pero puedes pedir correcciones.
> Responde inline y vuelve a invocar `/make` con esta tarea para que procese tus respuestas.

1. **¿Quieres hacer alguna corrección o ajuste a lo implementado en esta sesión?** (sí / no — si sí, descríbela en Detalle)
   **Respuesta:** No tengo comentarios.
   **Detalle:** Ya corregí con la tarea principal para poder soportar los tests que fallaban por el RFC.

### Iteración 2 (respondida)

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
