# RDK-021-TEST — Tests de exportabilidad/composabilidad de drivers.

## Ubicación
- **Repositorio:** Raidark
- **Componente:** `examples/external-factory/`, paquetes públicos de RDK-005..RDK-020.
- **Tarea madre:** [`RDK-021`](RDK-021.md)

## Tarea técnica
- **Tipo:** DEVELOPMENT
- **Estado:** Ready
- **Quién:** DEV
- **Qué:** Tests que verifican que un consumidor externo puede componer factories propias sobre los drivers de Raidark sin tocar el árbol `shared/`.
- **Cómo:**
  - **Test de superficie pública:**
    - Para cada driver (storage, image pipeline, pdf pipeline, jetstream pub/sub, outbox, dedup, idempotency, http client, ws, email, partitioning, sql migrations), confirmar que tipos, opciones y constructor son exportados (`PascalCase`).
    - Test con import explícito desde `examples/external-factory/` que instancia cada uno desde fuera de `shared/`.
  - **Factory externa de ejemplo:**
    - `examples/external-factory/opinated_outbox.go` define una factory que envuelve `outbox.NewDispatcher` con valores opinados.
    - Test que registra la factory en el provider hub y verifica que el dispatcher se construye con los defaults.
  - **Múltiples instancias:**
    - Crear dos `JetStreamPublisher` con configs distintas en el mismo proceso → ambos funcionan.
    - Crear dos `httpx.ServiceClient` con baseURL distinta → ambos hacen requests independientes.
  - **Lifecycle:**
    - Para cada driver con `Close()`/`Stop()`: el método es público y testeable.
    - Provider hub permite registrar provider post-bootstrap (test con timing controlado).
- **Cuándo:** Junto con `RDK-021`. Bloqueante: todas las tareas de testing de RDK-005..RDK-020.

## Criterio de aceptación
- `examples/external-factory/` compila y sus tests pasan.
- Test parametrizado recorre la lista de drivers y verifica para cada uno que existan los símbolos públicos esperados (puede usar `reflect` + lista de nombres).
- Documentado en `docs/extending.md` con referencia a los tests como "ejemplos vivos".

## Historia de usuario relacionada
- **Actor:** Mantenedor de un proyecto base sobre Raidark.
- **Historia:** Como mantenedor de un proyecto que estandariza Raidark para varios servicios, quiero tests que verifiquen exportabilidad y composabilidad, para detectar inmediatamente si un cambio en Raidark privatiza un símbolo del que dependo.
- **Valor esperado:** La superficie pública queda contractual: cualquier ruptura se detecta en CI antes de un release.
