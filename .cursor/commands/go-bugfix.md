---
description: Flujo ligero para corregir un bug (reproducir, causa raíz, parche mínimo).
handoffs:
  - label: Sugerir commit
    agent: go-suggest-commit
    prompt: Sugiere mensaje de commit para el parche
    send: true
---

## Entrada del usuario

```text
$ARGUMENTS
```

Incluye síntoma, entorno si aplica (versión, flags), logs o test que falla, y **bounded context** si lo sabes. Si falta información para no adivinar, pregunta.

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture (`internal/<contexto>/`, `pkg/`).
- Este comando **no** exige SDD extenso ni lista de tareas formal; el equipo puede añadir trazabilidad aparte si la necesita.
- Alineación obligatoria con `docs/ARCHITECTURE.md`, `docs/SDD.md` y [`docsy-main-guardrails.mdc`](../rules/docsy-main-guardrails.mdc).
- Skills típicos: [`go-errors`](../skills/go-errors/SKILL.md), [`go-testing`](../skills/go-testing/SKILL.md), [`go-api-rest`](../skills/go-api-rest/SKILL.md) según el fallo.

## Objetivo (orden sugerido)

1. **Reproducir** el fallo: test que falla, request mínimo, o pasos claros; si no hay test, valorar añadir uno que falle antes del fix.
2. **Aislar** el alcance: archivo/capa (transport vs use case vs repo) sin ampliar scope a refactors no pedidos.
3. **Causa raíz**: explicación breve (por qué ocurre), sin culpar capas arbitrariamente.
4. **Parche mínimo** que corrige el bug; preservar comportamiento existente salvo lo necesario (edición aditiva por defecto; ver guardrails).
5. **`make test`** y **`make verify`** (o el target acordado); si falla, aplicar la heurística de corrección en guardrails antes de pedir ayuda humana.
6. No ejecutar `git commit` salvo que el usuario lo pida explícitamente.

## Referencias

- Implementación según plan o SDD: [`/go-implement`](go-implement.md)
