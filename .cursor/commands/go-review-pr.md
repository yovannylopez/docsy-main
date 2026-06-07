---
description: Revisa un PR contra arquitectura, SDD, seguridad y calidad.
---

## Entrada del usuario

```text
$ARGUMENTS
```

URL del pull request, o número de PR en el repo actual, o rama a comparar con la base.

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture (`internal/<contexto>/`, `pkg/`, OpenAPI).
- Criterios: `docs/ARCHITECTURE.md`, `docs/SDD.md`; skills en `.cursor/skills/` para detalle técnico.

## Objetivo

1. Obtener el diff y el contexto del PR (`gh pr diff`, `gh pr view`, o equivalente). Si no hay acceso al remoto, pedir diff pegado o usar solo cambios locales.
2. Revisar frente a:
   - Alineación con capas y bounded contexts del proyecto
   - Contratos HTTP / OpenAPI si aplica
   - SQL/migraciones parametrizadas y convenciones sqlx
   - Errores, validación, logging, tests
   - **Object Mother:** en `internal/<slice>/`, tests que construyan DTOs/entidades de dominio del slice deben usar `internal/<slice>/test_utils` (p. ej. `NewAuthStubs`, `NewUsersStubs`, `Clone*`). **Crítico (pre-merge):** literales masivos de `entities.*` / `dtos.*` cuando ya existe `test_utils` en ese slice, salvo delta mínimo o excepción documentada en `.cursor/skills/go-testing/SKILL.md`.
   - **Tests y `dupl`:** dos o más `Test*` con el mismo cuerpo deben unificarse (helper con `t.Helper()` o table-driven); ver `.cursor/skills/go-testing/SKILL.md`.
   - **CI:** el PR debe poder pasar `make verify` (incluye golangci-lint).
   - Seguridad (auth, secretos, inputs)
3. Entregar hallazgos con severidad:
   - **Crítico:** debe corregirse antes de merge
   - **Sugerencia:** mejora recomendada
   - **Nit:** opcional / estilo
4. Resumir veredicto: aprobado con comentarios, cambios solicitados, o bloqueado.
