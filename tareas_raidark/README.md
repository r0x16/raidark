# Tareas Raidark

Esta carpeta contiene un set de tareas de desarrollo para el framework Go
**Raidark** (`github.com/r0x16/Raidark`). Las tareas son **independientes del
producto**: pueden trasladarse al repositorio de Raidark e implementarse sin
necesidad de conocer ningún consumidor concreto del framework.

## 1. Qué es Raidark

Raidark es un framework Go de servicios web diseñado para ser la base
transversal de APIs y servicios HTTP. Hoy provee:

- HTTP server sobre Echo con health-check y CSRF opcional.
- Provider hub para inyección de dependencias.
- Database providers GORM (sqlite, postgres, mysql).
- Auth providers (`array` mock + `casdoor`).
- Module hooks (`Setup`, `GetModel`, `GetSeedData`, `GetEventListeners`).
- Event provider in-memory (`InMemoryDomainEventsProvider`).
- CLI (`api`, `dbmigrate`, `dbmigrate seed`).

## 2. Objetivo de este conjunto de tareas

Convertir a Raidark en una base completa para construir servicios web
modernos que necesiten:

- Eventos de dominio sobre NATS + JetStream con outbox transaccional, dedup
  en consumer y DLQ.
- Almacenamiento de archivos con visibilidad pública/privada y URLs firmadas.
- Pipelines seguros de imágenes y PDFs.
- Convenciones REST estándar (envelope de error, paginación keyset, request
  id propagado).
- Observabilidad base (logs estructurados con trace_id, métricas
  Prometheus, propagación W3C trace context).
- Helpers de auth basados en permisos.
- Soporte WebSocket.
- Cliente HTTP inter-servicios estandarizado.
- EmailSender abstracto.
- Idempotency-Key store.
- Particionado mensual nativo de Postgres.
- Sanitizador markdown.
- Migraciones SQL custom además del auto-migrate de GORM.

Todas las funcionalidades nuevas deben ser **opcionales** y **componibles**:
quien use Raidark elige qué providers activar y cómo configurarlos.

## 3. Estructura de las tareas

Cada tarea sigue el formato `RDK-NNN.md` con:

- **Ubicación**: rama y épica técnica de Raidark.
- **Tarea técnica**: tipo (DEV / DECISION / DATA / SPIKE), estado, quién,
  qué, cómo, por qué, cuándo (dependencias).
- **Criterio de aceptación**: lista verificable.
- **Fuera de alcance**: lo que la tarea explícitamente no cubre.
- **Historia de usuario relacionada**: actor, historia, valor esperado.

## 4. Índice de tareas

Orden recomendado de implementación (los items más arriba desbloquean a los
de abajo):

| ID | Tarea | Tipo | Bloqueantes |
|---|---|---|---|
| `RDK-001` | Helper UUIDv7 | DEV | — |
| `RDK-002` | Convenciones REST: envelope de error y paginación keyset | DEV | — |
| `RDK-003` | Observabilidad base: logs estructurados, métricas Prometheus, W3C trace | DEV | `RDK-002` |
| `RDK-004` | CORS configurable y CSRF apagable por servicio | DEV | — |
| `RDK-005` | Storage adapter abstracto (interfaz + driver filesystem) | DEV | `RDK-001` |
| `RDK-006` | Pipeline de imágenes seguras | DEV | `RDK-005` |
| `RDK-007` | Pipeline de PDFs seguros | DEV | `RDK-005` |
| `RDK-008` | Envelope estándar de evento de dominio | DEV | `RDK-001` |
| `RDK-009` | Driver JetStream publisher | DEV | `RDK-008` |
| `RDK-010` | Driver JetStream consumer (DLQ, backoff, clasificación, broadcast) | DEV | `RDK-008`, `RDK-009` |
| `RDK-011` | Outbox transaccional como librería reutilizable | DEV | `RDK-008`, `RDK-009` |
| `RDK-012` | Dedup en consumer (`processed_events`) | DEV | `RDK-010` |
| `RDK-013` | Helpers de auth basados en permisos | DEV | — |
| `RDK-014` | Sanitizador markdown | DEV | — |
| `RDK-015` | Idempotency-Key store | DEV | `RDK-002` |
| `RDK-016` | Cliente HTTP inter-servicios estandarizado | DEV | `RDK-002`, `RDK-003` |
| `RDK-017` | Soporte WebSocket | DEV | — |
| `RDK-018` | EmailSender abstracto con drivers SMTP y HTTP | DEV | — |
| `RDK-019` | Helper de particionado mensual nativo de Postgres | DEV | — |
| `RDK-020` | Soporte para migraciones SQL custom además de auto-migrate GORM | DEV | — |
| `RDK-021` | Drivers exportables y componibles desde el exterior | DEV | todas las anteriores |

### 4.bis Tareas de testing

Raidark hoy **no tiene tests**. Antes de tocar funcionalidad nueva hay que
levantar la base de testing, y cada feature nueva entrega su propia suite.

#### Estrategia y cobertura del código existente

| ID | Tarea | Bloqueantes |
|---|---|---|
| `RDK-TEST-000` | Estrategia y toolchain de testing (helpers, Makefile, CI) | — |
| `RDK-TEST-LEGACY-001` | Tests para `shared/env` | `RDK-TEST-000` |
| `RDK-TEST-LEGACY-002` | Tests para `shared/auth` (Casdoor + Array) | `RDK-TEST-000` |
| `RDK-TEST-LEGACY-003` | Tests para `shared/api` (EchoApiProvider, modules, util) | `RDK-TEST-000` |
| `RDK-TEST-LEGACY-004` | Tests para `shared/datastore` (Postgres/MySQL/SQLite) | `RDK-TEST-000` |
| `RDK-TEST-LEGACY-005` | Tests para `shared/events` (in-memory provider) | `RDK-TEST-000` |
| `RDK-TEST-LEGACY-006` | Tests para `shared/migration` (migrate, seed) | `RDK-TEST-000` |
| `RDK-TEST-LEGACY-007` | Tests para `shared/cmd` (CLI Cobra) | `RDK-TEST-000`, `LEGACY-006` |
| `RDK-TEST-LEGACY-008` | Tests para `shared/providers` (hub + factories) | `RDK-TEST-000`, `LEGACY-001` |
| `RDK-TEST-LEGACY-009` | Tests para `shared/logger` (StdOut, sanitizer) | `RDK-TEST-000` |
| `RDK-TEST-LEGACY-010` | Tests para `shared/serverevents` (SSE Echo) | `RDK-TEST-000` |

#### Tests por feature nueva (uno por cada `RDK-NNN`)

| ID | Cubre | Bloqueantes principales |
|---|---|---|
| `RDK-001-TEST` | UUIDv7 helper | `RDK-TEST-000` |
| `RDK-002-TEST` | Convenciones REST | `RDK-TEST-000`, `RDK-001-TEST` |
| `RDK-003-TEST` | Observabilidad base | `RDK-TEST-000`, `RDK-002-TEST` |
| `RDK-004-TEST` | CORS/CSRF toggleable | `RDK-TEST-000`, `LEGACY-003` |
| `RDK-005-TEST` | Storage adapter (filesystem) | `RDK-TEST-000`, `RDK-001-TEST` |
| `RDK-006-TEST` | Pipeline imágenes | `RDK-005-TEST` |
| `RDK-007-TEST` | Pipeline PDFs | `RDK-005-TEST` |
| `RDK-008-TEST` | Envelope evento | `RDK-001-TEST` |
| `RDK-009-TEST` | JetStream publisher | `RDK-008-TEST` |
| `RDK-010-TEST` | JetStream consumer | `RDK-009-TEST` |
| `RDK-011-TEST` | Outbox transaccional | `RDK-009-TEST` |
| `RDK-012-TEST` | Dedup consumer | `RDK-010-TEST` |
| `RDK-013-TEST` | Helpers permisos | `LEGACY-002`, `RDK-002-TEST` |
| `RDK-014-TEST` | Sanitizador markdown | `RDK-TEST-000` |
| `RDK-015-TEST` | Idempotency-Key store | `LEGACY-004`, `RDK-002-TEST` |
| `RDK-016-TEST` | Cliente HTTP inter-servicios | `RDK-002-TEST`, `RDK-003-TEST` |
| `RDK-017-TEST` | WebSocket | `LEGACY-003` |
| `RDK-018-TEST` | EmailSender (SMTP/generic/GWS/Brevo) | `RDK-016-TEST` |
| `RDK-019-TEST` | Particionado Postgres | `LEGACY-004`, `RDK-020-TEST` |
| `RDK-020-TEST` | Migraciones SQL custom | `LEGACY-006` |
| `RDK-021-TEST` | Exportabilidad/composabilidad | todas las RDK-*-TEST |

> Cada tarea de testing entrega los tests definidos en el criterio de
> aceptación de su tarea madre, más cobertura adicional listada en su propio
> criterio.

## 5. Cómo trabajar estas tareas

- Cada tarea es autocontenida. Puede implementarse aislada siempre que se
  respeten los bloqueantes declarados en el campo `Cuándo`.
- Toda tarea agrega o actualiza documentación dentro de `docs/` del
  repositorio de Raidark, en la subcarpeta indicada en el criterio de
  aceptación.
- Toda tarea entrega tests: unitarios siempre, integración cuando aplique
  (drivers contra dependencias reales o emuladas, p. ej. NATS embebido,
  Postgres con testcontainers).
- Las funcionalidades nuevas se entregan como providers/drivers que el
  consumidor activa explícitamente. Nada se prende por defecto si rompe la
  superficie pública actual de Raidark.
- Variables de entorno declaradas por las tareas se documentan en
  `docs/configuration.md` y aparecen en el `.env.example` de referencia.

## 6. Convenciones técnicas

- **Lenguaje:** Go (mismo go.mod que Raidark hoy).
- **HTTP:** Echo (driver actual).
- **DB:** GORM con drivers existentes; Postgres es el target principal para
  features avanzadas (particionado, índices funcionales).
- **IDs:** UUIDv7 para PKs, mensajes y agregados (RDK-001).
- **Tiempo:** UTC en almacenamiento y en headers/payloads. La conversión a
  zonas locales es responsabilidad del consumidor.
- **Errores:** sentinels tipados (`ErrNotFound`, `ErrConflict`,
  `ErrForbidden`, `ErrValidation`, `ErrTransient`, `ErrPermanent`).
- **Configuración:** sólo por variables de entorno. Nunca hardcodear
  endpoints, secretos o paths.
- **Compatibilidad:** las tareas no deben romper la API pública actual de
  Raidark. Si algo cambia, debe declararse en el criterio de aceptación.
