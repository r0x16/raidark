# RDK-008-TEST — Tests para envelope estándar de evento.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/events/domain`
- **Tarea madre:** [`RDK-008`](RDK-008.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del `EventEnvelope` y del helper `NewEnvelope`.
- **Cómo:**
  - **Round-trip JSON:**
    - Envelope → JSON → Envelope produce mismos bytes (snapshot test contra fixture).
    - Campos opcionales ausentes (sin `correlation_id`) se omiten u quedan vacíos según convención.
    - Decoder ignora campos desconocidos (forward compat).
  - **`NewEnvelope`:**
    - `event_id` válido UUIDv7.
    - `event_name` proviene de `DomainEvent.Name()`.
    - `occurred_at` proviene de `DomainEvent.OccurredAt()`.
    - `published_at` vacío en construcción.
    - `idempotency_key` default = `event_id`.
    - `trace_id` y `span_id` extraídos del context cuando están; vacíos cuando no.
    - `producer` corresponde al parámetro recibido.
  - **Versionado:**
    - Helper o constante de `event_version` por default = 1.
    - Documentar (en test comments) que cambio incompatible exige nuevo subject; no se testea el subject porque no es responsabilidad de Raidark.
- **Cuándo:** Junto con `RDK-008`. Bloqueante: `RDK-TEST-000`, `RDK-001-TEST`.

## Criterio de aceptación
- Cobertura ≥ 95% en `shared/events/domain` (envelope).
- Snapshot test contra fixture JSON estable.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests del envelope para que ningún cambio accidental rompa la serialización que consumidores externos esperan.
- **Valor esperado:** El envelope queda como contrato verificado; cambios futuros se detectan en CI.
