---
description: Genera plan de arquitectura y borrador de SDD alineado al boilerplate.
handoffs:
  - label: Implementar
    agent: go-implement
    prompt: Implementa según el SDD acordado
    send: true
---

## Entrada del usuario

```text
$ARGUMENTS
```

Incluye el requerimiento (salida de `/go-hu-extract` o texto pegado), **bounded context** y restricciones. Si falta bounded context, pregunta antes de proponer rutas bajo `internal/`.

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture (`internal/<contexto>/`, `pkg/`, OpenAPI).
- Lee y respeta `docs/ARCHITECTURE.md` y la estructura / tono de `docs/SDD.md` como referencia canónica.
- No copies párrafos largos de `.cursor/skills/`; enlaza qué skills aplicarán en implementación (go-conventions, go-api-rest, database-queries, etc.) según el caso.

## Objetivo

1. Leer `docs/ARCHITECTURE.md` (y `docs/SDD.md` como guía de formato).
2. Proponer diseño: bounded context, capas afectadas, contratos HTTP/OpenAPI si aplica, datos/migraciones si aplica, riesgos y decisiones (ADRs breves si ayudan).
3. Entregar un **borrador de SDD** (markdown listo para pegar en `docs/` o ruta sugerida) alineado al estilo del SDD del repo.
4. Mantén **una sola fuente de verdad** por cambio: el SDD o plan acordado en `docs/` (o en la entrada del usuario), sin contradecir `docs/ARCHITECTURE.md`.
