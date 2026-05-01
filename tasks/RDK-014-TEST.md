# RDK-014-TEST — Tests para sanitizador markdown.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `shared/sanitizer/markdown`
- **Tarea madre:** [`RDK-014`](RDK-014.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests de las tres políticas (`PolicyDefault`, `PolicyStrict`, `PolicyArticle`) y del sistema de policies custom.
- **Cómo:**
  - **Vectores XSS** (table-driven, ~30 casos basados en OWASP XSS Filter Evasion):
    - `<script>alert(1)</script>` → removido en las tres policies.
    - `<iframe src="...">` → removido en las tres.
    - `[link](javascript:alert(1))` → protocolo neutralizado.
    - `<img onerror="alert(1)" src="x">` → atributo on* removido.
    - `<a href="..." style="...">` → atributo `style` removido.
    - Encodings (HTML entities, percent encoding) que intenten bypass → no logran.
  - **PolicyStrict:**
    - `![alt](img.jpg)` → imagen removida.
    - `# H1` → renderizado pero sin tags h1 si la policy lo restringe; documentar y testear el rendering esperado.
  - **PolicyArticle:**
    - `![alt](https://cdn.example/img.jpg)` con `cdn.example` en allowlist → presente.
    - Misma URL con dominio fuera de allowlist → removida.
    - Headings, listas, code blocks, blockquotes → presentes.
  - **Round-trip:**
    - `Sanitize(Sanitize(input))` produce mismo output que `Sanitize(input)` (idempotente).
  - **Custom policy:**
    - `RegisterPolicy("news", customPolicy)` registra y permite usarla por nombre.
- **Cuándo:** Junto con `RDK-014`. Bloqueante: `RDK-TEST-000`.

## Criterio de aceptación
- Cobertura ≥ 90% en `shared/sanitizer/markdown`.
- Suite XSS con ≥ 25 vectores, todos pasan en las tres policies por default.

## Historia de usuario relacionada
- **Actor:** Equipo desarrollador de Raidark.
- **Historia:** Como desarrollador de Raidark, quiero una suite XSS exhaustiva sobre el sanitizador markdown, para que cualquier servicio consumidor herede protección verificable contra inyecciones HTML.
- **Valor esperado:** El sanitizador queda blindado por una batería de vectores conocidos; nuevos vectores se agregan al test suite con un solo cambio.
