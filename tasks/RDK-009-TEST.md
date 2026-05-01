# RDK-009-TEST — Tests para JetStream publisher.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/events/driver/JetStreamDomainEventsProvider`
- **Tarea madre:** [`RDK-009`](RDK-009.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests de integración del publisher contra NATS embebido.
- **Cómo:**
  - **Setup:** helper `testutil/nats` levanta NATS con JetStream embedded; crea stream y subject por test.
  - **Casos:**
    - Publish exitoso: `PubAck` recibido, mensaje aparece en stream con header `Nats-Msg-Id` = `event_id`.
    - **Dedup nativo:** publicar mismo `event_id` dos veces → stream tiene un solo mensaje.
    - **Headers:** mensaje recibido contiene `trace_id`, `span_id`, `producer`, `event_version`, `traceparent` esperados.
    - **Subject mapper:** mapper custom retornando subject distinto se respeta.
    - **Backpressure:** llenar canal interno → siguiente publish retorna `ErrTransient`.
    - **Timeout:** publish con servidor que no acka dentro del timeout → `ErrTransient`.
    - **Close:** `Close()` flush mensajes pendientes y cierra conexión limpiamente.
- **Cuándo:** Junto con `RDK-009`. Bloqueante: `RDK-TEST-000`, `RDK-008-TEST`.

## Criterio de aceptación
- Cobertura ≥ 80% en el driver.
- Tests viven detrás de build tag `integration`.
- Tests corren con `-race` sin warnings.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests de integración del publisher contra NATS real (embedded), para verificar dedup nativo, headers y backpressure antes de que servicios consumidores construyan sobre el driver.
- **Valor esperado:** El driver queda verificado en condiciones reales; bugs en headers o dedup se detectan en CI.
