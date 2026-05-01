# RDK-001-TEST — Tests para UUIDv7 helper.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/ids`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing
- **Tarea madre:** [`RDK-001`](RDK-001.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
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
