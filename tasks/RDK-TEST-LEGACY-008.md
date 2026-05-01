# RDK-TEST-LEGACY-008 — Tests para `shared/providers` (provider hub, factories).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/providers/{driver,services,domain}/...`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests el provider hub y todas las factories: `ApiProviderFactory`, `AuthProviderFactory`, `DatastoreProviderFactory`, `DomainEventFactory`, `EnvProviderFactory`, `LoggerProviderFactory`, y el `ProviderHubFactory`.
- **Cómo:**
  - **`ProviderHubFactory`**:
    - Build con set mínimo de factories registradas → hub funcional.
    - Lookup por tipo retorna provider esperado.
    - Lookup de provider no registrado → error claro (no panic).
  - **Cada factory individual**:
    - Construye provider con env vars de prueba.
    - Falla con error claro si faltan variables obligatorias.
    - `DomainEventFactory`: con `DOMAIN_EVENT_PROVIDER_TYPE=in_memory` retorna `InMemoryDomainEventsProvider`. Casos para los demás drivers cuando se agreguen, viven en sus tareas.
    - `AuthProviderFactory`: con `AUTH_PROVIDER_TYPE=array` retorna `ArrayAuthProvider`; con `casdoor` retorna `CasdoorAuthProvider` (config válida) o error.
    - `DatastoreProviderFactory`: dispatch correcto entre Postgres/MySQL/SQLite según `DATASTORE_TYPE`.
- **Cuándo:** Bloqueante: `RDK-TEST-000`, `RDK-TEST-LEGACY-001`.

## Criterio de aceptación
- Cobertura ≥ 75% en `shared/providers/`.
- Cada factory tiene tests de happy path y de error de configuración.

## Fuera de alcance
- Tests de las factories nuevas (storage, jetstream, outbox, etc.) — viven en sus respectivas tareas `*-TEST`.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre el provider hub y sus factories, para que el bootstrap no falle silenciosamente cuando una env var no se respete o una factory cambie su firma.
- **Valor esperado:** El sistema de DI queda verificado y se vuelve la pieza más estable para extender con factories nuevas.
