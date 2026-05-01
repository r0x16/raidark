# RDK-006-TEST — Tests para pipeline de imágenes.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/imaging`
- **Tarea madre:** [`RDK-006`](RDK-006.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del pipeline de imágenes seguras.
- **Cómo:**
  - **Validación:**
    - Magic bytes inválidos (txt, gif) → `ErrValidation`.
    - JPEG/PNG/WebP válidos → procesa.
    - Archivo > tamaño máximo → `ErrValidation`.
    - Dimensiones bajo mínimo o sobre máximo → `ErrValidation`.
  - **Strip metadata:**
    - JPEG con EXIF GPS de fixture → output re-leído no contiene tags GPS ni IPTC ni XMP.
    - PNG con metadata textual → idem.
  - **Variantes:**
    - Cada `VariantSpec` produce archivo con dimensiones ≤ `MaxWidth/Height`, conservando aspect ratio.
    - Output WebP y JPEG ambos generados según política.
  - **Persistencia:**
    - Cada variante persistida con key `{namespace}/{usage}/{año}/{mes}/{uuid}_{variant}.{ext}` (verificable en storage mock).
    - `HashSha256` corresponde al input original.
  - **Memoria:**
    - Imagen de 20MB no spike-ea más de N MB en RAM (medido con `MemStats`).
- **Cuándo:** Junto con `RDK-006`. Bloqueante: `RDK-TEST-000`, `RDK-005-TEST`.

## Criterio de aceptación
- Cobertura ≥ 80% en `shared/imaging`.
- `testdata/` con: jpeg con EXIF GPS, png con metadata, webp válido, gif inválido, jpeg de 20MB.
- Tests verifican strip leyendo metadata del output con `goexif` o equivalente.

## Fuera de alcance
- Tests de instalación de libvips (responsabilidad de CI/Dockerfile).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre el pipeline de imágenes, para asegurar que el strip de EXIF y la validación de dimensiones nunca sean bypasseables.
- **Valor esperado:** Cualquier servicio que reciba uploads queda protegido sin lógica propia, con confianza verificable.
