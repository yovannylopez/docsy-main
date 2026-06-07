# Instrucciones para agentes (docsy-main)

## Contexto del proyecto

API Go con **Clean Architecture**, **Echo**, **PostgreSQL/sqlx**, JWT y paquetes compartidos en `pkg/`. Convenciones completas en **`.cursor/rules/`** y **`.cursor/skills/`**.

## Arquitectura canónica

- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — vertical slicing, capas, flujo HTTP → transport → use cases → domain → infrastructure, contenedores y patrones.
- **[docs/SDD.md](docs/SDD.md)** — Software Design Document completo del boilerplate.

Ante duda sobre convenciones, prevalece la documentación en `docs/` y los skills en `.cursor/skills/`.

- **Constitución del proyecto** (principios y stack): **[docs/CONSTITUTION.md](docs/CONSTITUTION.md)**.
- **Human-in-the-loop, tabla de reglas `.mdc`, tests/Object Mother, módulo Go**: **[docs/agent-guide.md](docs/agent-guide.md)**.

## Comandos Cursor (flujo HU → PR)

Comandos en `.cursor/commands/` (invocar con `/` en el chat). Cadena típica: extracto de HU → SDD → implementación → commit → PR → revisión. Para **bugs** o **refactors** acotados, ver **`/go-bugfix`** y **`/go-refactor`**.

| Comando | Archivo | Uso |
|---------|---------|-----|
| `/go-hu-extract` | [.cursor/commands/go-hu-extract.md](.cursor/commands/go-hu-extract.md) | Leer HU y extraer requerimiento estructurado |
| `/go-sdd` | [.cursor/commands/go-sdd.md](.cursor/commands/go-sdd.md) | Plan de arquitectura y borrador de SDD |
| `/go-implement` | [.cursor/commands/go-implement.md](.cursor/commands/go-implement.md) | Implementar según SDD o plan acordado |
| `/go-bugfix` | [.cursor/commands/go-bugfix.md](.cursor/commands/go-bugfix.md) | Flujo ligero: reproducir → causa raíz → parche mínimo |
| `/go-refactor` | [.cursor/commands/go-refactor.md](.cursor/commands/go-refactor.md) | Refactor incremental; escalar a SDD si el alcance es grande |
| `/go-suggest-commit` | [.cursor/commands/go-suggest-commit.md](.cursor/commands/go-suggest-commit.md) | Proponer mensaje de commit |
| `/go-generate-pr` | [.cursor/commands/go-generate-pr.md](.cursor/commands/go-generate-pr.md) | Redactar y crear PR |
| `/go-review-pr` | [.cursor/commands/go-review-pr.md](.cursor/commands/go-review-pr.md) | Revisar PR frente a arquitectura y calidad |

## Skills del repositorio (léelos cuando el trabajo lo requiera)

Los skills viven en **`.cursor/skills/<nombre>/SKILL.md`**. Complementan las reglas con guías prácticas para cada área técnica.

| Skill | Ruta | Cuándo usarlo |
|----------|------|----------------|
| **go-tooling** | [.cursor/skills/go-tooling/SKILL.md](.cursor/skills/go-tooling/SKILL.md) | Lint, formato, `make verify` / `verify-dev`, `golangci-lint` |
| **go-testing** | [.cursor/skills/go-testing/SKILL.md](.cursor/skills/go-testing/SKILL.md) | Tests, Object Mother (`test_utils`), Mockery, cobertura, `make generate-*-mocks` |
| **go-conventions** | [.cursor/skills/go-conventions/SKILL.md](.cursor/skills/go-conventions/SKILL.md) | Estructura `internal/<slice>/`, puertos, capas |
| **go-generics** | [.cursor/skills/go-generics/SKILL.md](.cursor/skills/go-generics/SKILL.md) | Type parameters y helpers genéricos |
| **go-api-rest** | [.cursor/skills/go-api-rest/SKILL.md](.cursor/skills/go-api-rest/SKILL.md) | Handlers Echo, `pkg/responses`, HTTP |
| **database-queries** | [.cursor/skills/database-queries/SKILL.md](.cursor/skills/database-queries/SKILL.md) | sqlx, SQL, migraciones |
| **go-validation** | [.cursor/skills/go-validation/SKILL.md](.cursor/skills/go-validation/SKILL.md) | `pkg/validators`, DTOs, errores de validación |
| **go-context** | [.cursor/skills/go-context/SKILL.md](.cursor/skills/go-context/SKILL.md) | `context.Context`, timeouts, shutdown |
| **go-errors** | [.cursor/skills/go-errors/SKILL.md](.cursor/skills/go-errors/SKILL.md) | `%w`, mapeo a respuestas HTTP |
| **go-logging** | [.cursor/skills/go-logging/SKILL.md](.cursor/skills/go-logging/SKILL.md) | `pkg/logging` (zap) |

Detalle de reglas, human-in-the-loop y tests: **[docs/agent-guide.md](docs/agent-guide.md)**.
