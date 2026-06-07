---
description: Propone mensaje de commit convencional a partir del diff y el estado del repo.
handoffs:
  - label: Generar PR
    agent: go-generate-pr
    prompt: Genera el pull request para esta rama
    send: true
---

## Entrada del usuario

```text
$ARGUMENTS
```

Opcional: alcance del commit, ticket/HU, o instrucciones (ej. solo un archivo).

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture (`internal/<contexto>/`, `pkg/`, OpenAPI).
- Convenciones: `docs/ARCHITECTURE.md`, `docs/SDD.md`, `AGENTS.md`.

## Objetivo

1. Revisar `git status` y `git diff` (staged y unstaged según lo que el usuario quiera incluir).
2. Proponer uno o más mensajes de commit en estilo **Conventional Commits** (ej. `feat(auth): ...`, `fix: ...`).
3. El cuerpo del mensaje debe ser claro y en oraciones completas donde aplique.
4. **No** ejecutar `git commit` ni `git push` a menos que el usuario lo pida explícitamente en el chat.
