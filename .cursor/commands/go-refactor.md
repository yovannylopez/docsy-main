---
description: Refactor incremental con límites claros; escalar a SDD o plan cuando el alcance lo exija.
handoffs:
  - label: Plan SDD
    agent: go-sdd
    prompt: El refactor cruza bounded contexts o API pública; propón plan y SDD acotado
    send: false
  - label: Sugerir commit
    agent: go-suggest-commit
    prompt: Sugiere mensaje de commit para el refactor
    send: true
---

## Entrada del usuario

```text
$ARGUMENTS
```

Incluye qué código tocar, objetivo del refactor (legibilidad, duplicación, preparar feature), y **bounded context(s)**. Si el alcance es ambiguo, pregunta antes de cambiar APIs públicas o contratos.

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture.
- **Refactor localizado** puede hacerse con este comando; **refactor amplio** (varios contextos, OpenAPI, migraciones, contratos) debe pasar por **`/go-sdd`** o notas explícitas en `docs/`.
- [`go-conventions`](../skills/go-conventions/SKILL.md), [`go-testing`](../skills/go-testing/SKILL.md), [`go-errors`](../skills/go-errors/SKILL.md) suelen aplicar.

## Objetivo (orden sugerido)

1. **Delimitar superficie**: paquetes y archivos afectados; listar rompimientos de API potenciales.
2. **Tests de caracterización** si hace falta: cubrir comportamiento actual antes de mover código (sobre todo si no hay cobertura).
3. **Cambios incrementales** por pasos verificables; edición **aditiva** por defecto (ver [`docsy-main-guardrails.mdc`](../rules/docsy-main-guardrails.mdc)); evitar reescrituras completas de archivos grandes sin necesidad.
4. Tras cada paso relevante: `make test` / `make verify` según guardrails.
5. Si el refactor **cruza bounded contexts**, **expone nuevos endpoints**, o **cambia contratos** (OpenAPI, DB): **parar** y proponer **`/go-sdd`** o plan explícito con el usuario — no improvisar diseño amplio en un solo paso.
6. No ejecutar `git commit` salvo que el usuario lo pida explícitamente.

## Referencias

- Implementación: [`/go-implement`](go-implement.md)
