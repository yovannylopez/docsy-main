---
description: Implementa cambios según el SDD o plan acordado.
handoffs:
  - label: Sugerir commit
    agent: go-suggest-commit
    prompt: Sugiere mensaje de commit para los cambios realizados
    send: true
---

## Entrada del usuario

```text
$ARGUMENTS
```

Incluye ruta al SDD en `docs/`, resumen del plan y el **bounded context**. Si falta información para implementar sin adivinar, pregunta.

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture (`internal/<contexto>/`, `pkg/`, OpenAPI).
- Implementación alineada con `docs/ARCHITECTURE.md` y `docs/SDD.md`.
- Antes de tocar código, lee los **skills** pertinentes desde `.cursor/skills/` (ver `AGENTS.md` e índice en `.cursor/rules/docsy-main-skills.mdc`).

## Flujos ligeros

- **Bug hotfix** sin SDD extenso: [`/go-bugfix`](go-bugfix.md).
- **Refactor** acotado (sin cruzar varios contextos): [`/go-refactor`](go-refactor.md). Cuando el alcance sea grande, conviene **`/go-sdd`** o plan explícito en `docs/` antes de implementar.

## Objetivo

1. Confirmar alcance y bounded context; no ampliar scope sin acuerdo explícito del usuario.
2. Implementar siguiendo capas del proyecto (transport → use cases → domain → infrastructure), OpenAPI por feature si aplica, migraciones en `migrations/core/` si aplica.
3. Añadir o actualizar tests acorde al cambio; ejecutar verificación local razonable (`make verify` o lo que el usuario indique).
4. No ejecutar `git commit` salvo que el usuario lo pida explícitamente en el mensaje.
