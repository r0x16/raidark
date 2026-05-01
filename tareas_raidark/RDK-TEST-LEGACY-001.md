# RDK-TEST-LEGACY-001 — Tests para `shared/env`.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/env/driver/EnvProvider.go`
- **Épica técnica:** `EP-RDK-QUALITY` — Calidad y testing

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Cubrir con tests el `EnvProvider` que carga variables de entorno (vía `godotenv` + `viper`). Hoy no tiene tests.
- **Cómo:**
  - Tests unitarios contra:
    - Lectura de un `.env` real desde un directorio temporal.
    - Precedencia: `os.Setenv` gana sobre `.env`.
    - Variables ausentes devuelven default cuando hay default; cadena vacía cuando no.
    - Conversión de tipos (string → int, bool, duration) cuando esté implementada.
    - Errores claros si el archivo `.env` referenciado no existe (cuando se exige).
- **Cuándo:** Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 80% en `shared/env/driver/`.
- Tests usan archivos `.env` de fixture en `testdata/`.
- Cada función pública del provider tiene al menos un test happy-path y uno de error.

## Fuera de alcance
- Tests de cómo lo consumen otros componentes (los suben sus propias tareas).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre el `EnvProvider`, para detectar regresiones cuando se cambie el orden de precedencia o el manejo de valores faltantes.
- **Valor esperado:** El `EnvProvider` queda como referencia confiable para todos los demás providers que dependen de él.
