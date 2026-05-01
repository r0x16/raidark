# RDK-007-TEST — Tests para pipeline de PDFs.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/pdf`
- **Tarea madre:** [`RDK-007`](RDK-007.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests del pipeline de PDFs seguros.
- **Cómo:**
  - **Validación:**
    - Magic bytes no `%PDF-` → `ErrValidation` con código `pdf.invalid`.
    - PDF cifrado → `ErrValidation` con código `pdf.encrypted`.
    - Tamaño > `MaxBytes` → `ErrValidation` con código `pdf.too_large`.
    - Páginas > `MaxPages` → `ErrValidation` con código `pdf.too_many_pages`.
  - **Sanitización:**
    - PDF de fixture con `/JavaScript` → output re-parseado no contiene `/JS` ni `/JavaScript`.
    - PDF con `/OpenAction` → removido.
    - PDF con AcroForm → removido.
    - PDF con attachment embebido → removido.
  - **Persistencia:**
    - Output persistido con `Visibility=Private`.
    - Acceso vía `SignedURL` con TTL respeta el default.
  - **Output:**
    - `PageCount` real coincide con páginas del input.
    - `HashSha256` corresponde al input.
- **Cuándo:** Junto con `RDK-007`. Bloqueante: `RDK-TEST-000`, `RDK-005-TEST`.

## Criterio de aceptación
- Cobertura ≥ 80% en `shared/pdf`.
- `testdata/` con: pdf con JavaScript, pdf cifrado, pdf con AcroForm, pdf con attachment, pdf limpio multi-página.
- Tests verifican sanitización re-parseando el output con `pdfcpu` o `qpdf`.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests sobre el pipeline de PDFs, para garantizar que JavaScript y otras construcciones peligrosas nunca lleguen al storage final.
- **Valor esperado:** Cualquier servicio que reciba PDFs los entrega seguros, sin lógica propia y con cobertura automatizada.
