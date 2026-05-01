# RDK-005-TEST — Tests para storage adapter (filesystem).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/storage`
- **Tarea madre:** [`RDK-005`](RDK-005.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
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
