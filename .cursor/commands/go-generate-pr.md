---
description: Redacta y crea un PR (gh o pasos manuales) con checklist de verificación.
---

## Entrada del usuario

```text
$ARGUMENTS
```

Incluye si aplica: rama base (por defecto `main`), título deseado, enlaces a HU/SDD, notas para revisores.

Tras generar el PR, puedes ejecutar **`/go-review-pr`** con la URL o número del PR si quieres revisión en el mismo hilo.

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture (`internal/<contexto>/`, `pkg/`, OpenAPI).
- Referencias: `docs/ARCHITECTURE.md`, `docs/SDD.md`.

## Objetivo

1. Confirmar rama actual y rama base; resumir cambios (`git log`, `git diff` vs base).
2. Redactar **título y cuerpo** del PR: qué cambia, por qué, cómo probarlo, riesgos. Incluir enlaces a HU/SDD si existen.
3. Incluir en el cuerpo un **checklist** de verificación (ej. `make verify`, tests manuales, migraciones).
4. Si `gh` está disponible y el usuario autoriza operaciones en remoto, crear el PR con `gh pr create`; si no, dar los pasos exactos (GitHub UI o `gh` con flags).
5. No forzar push ni crear PR destructivo sin confirmación explícita.
