# RDK-TEST-LEGACY-009 — Tests para `shared/logger` (StdOutLogger, LogDataSanitizer).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/logger/driver/{StdOutLogger,LogDataSanitizer}.go`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests el logger stdout y el sanitizador de datos sensibles.
- **Cómo:**
  - **`StdOutLogger`**:
    - Tests con `bytes.Buffer` como `io.Writer` (capturable).
    - Niveles (`Debug`, `Info`, `Warn`, `Error`) emiten o silencian según `LOG_LEVEL`.
    - Estructura del output (campos esperados).
  - **`LogDataSanitizer`**:
    - Redacta llaves sensibles (`password`, `token`, `secret`, `authorization`, `api_key`, etc.) en mapas / structs / strings JSON-like.
    - Preserva el resto del payload sin tocar.
    - Casos de borde: nil, slices, nested maps.
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 80% en `shared/logger/`.
- Test table-driven para sanitizer cubre al menos: map plano, map anidado, slice de maps, struct con tags, JSON string.

## Fuera de alcance
- Tests del adapter JSON nuevo y de inyección de `trace_id` (viven en `RDK-003-TEST`).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre el logger y el sanitizador, para garantizar que ningún secreto se filtre en logs aunque se agreguen nuevas llaves al payload.
- **Valor esperado:** El sanitizador queda blindado contra regresiones que filtren PII o tokens.
