# RDK-TEST-LEGACY-005 — Tests para `shared/events` (in-memory provider).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/events/driver/InMemoryDomainEventsProvider.go` y `shared/events/domain/`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests el provider in-memory existente y los tipos del dominio de eventos (`DomainEvent`, `EventListener`).
- **Cómo:**
  - **`DomainEvent` interface**: implementación dummy en test, validar `Name()` y `OccurredAt()`.
  - **`InMemoryDomainEventsProvider`**:
    - Subscribe + Publish entrega evento al listener registrado.
    - Múltiples listeners reciben todos el mismo evento.
    - Publish sin listeners no panickea.
    - Concurrency: publish desde múltiples goroutines no rompe estado interno (race detector limpio).
    - `Close()` flush si aplica.
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 85% en `shared/events/`.
- Tests corren con `-race`.

## Fuera de alcance
- Tests del envelope, JetStream publisher/consumer, outbox, dedup — viven en sus tareas `RDK-008-TEST`..`RDK-012-TEST`.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre el provider in-memory de eventos, para que siga funcionando como driver de tests aún después de que se introduzca el driver JetStream.
- **Valor esperado:** El provider in-memory queda verificado y sigue siendo el default usable en suites de testing de servicios construidos sobre Raidark.
