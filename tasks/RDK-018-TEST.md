# RDK-018-TEST — Tests para EmailSender (SMTP, http_generic, http_gws, http_brevo).

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/email`
- **Tarea madre:** [`RDK-018`](RDK-018.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests por driver más una suite común de la interfaz.
- **Cómo:**
  - **Suite común** (ejecutada con cada driver tras un mock/embedded server):
    - `Send` con `EmailMessage` válido → no error.
    - `EmailMessage` sin `To` o sin `Subject` → error de validación tipado.
    - `From` por default proviene de `EMAIL_DEFAULT_FROM` cuando el campo está vacío.
  - **`smtp`** contra mailpit/mailhog testcontainer:
    - Mensaje recibido conserva `Subject`, `From`, `To`, `BodyText`, `BodyHTML`.
    - STARTTLS funciona contra el contenedor con TLS habilitado.
  - **`http_generic`** contra mock server:
    - `EMAIL_HTTP_PAYLOAD_TEMPLATE` se renderiza con todos los placeholders.
    - Header de auth se envía según config.
    - 5xx → `ErrTransient`. 4xx → `ErrPermanent`. 429 con `Retry-After` → `ErrTransient`.
  - **`http_gws`** contra mock de Gmail API:
    - Service account JSON parseado.
    - Token JWT generado y enviado como `Authorization: Bearer ...`.
    - Payload base64url del MIME.
  - **`http_brevo`** contra mock:
    - JSON con shape `{sender, to, subject, htmlContent, textContent}`.
    - Header `api-key` con valor de env.
- **Cuándo:** Junto con `RDK-018`. Bloqueante: `RDK-TEST-000`, `RDK-016-TEST`.

## Criterio de aceptación
- Cobertura ≥ 80% en `shared/email`.
- Cada driver tiene suite específica + suite común.
- Test SMTP con build tag `integration` (testcontainer mailpit).

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero tests por cada driver de email, para que cambiar de SMTP a Brevo o GWS sea seguro y se detecten regresiones provider-specific.
- **Valor esperado:** Los cuatro drivers quedan verificados; el consumidor elige por env sin sorpresas.
