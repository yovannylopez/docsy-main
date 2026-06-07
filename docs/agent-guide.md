# Guía ampliada para agentes (docsy-main)

Punto de entrada corto: **[AGENTS.md](../AGENTS.md)** en la raíz del repo. Este documento amplía **human-in-the-loop**, **reglas Cursor**, **tests/Object Mother** y **módulo Go**.

## Human-in-the-loop (cuándo parar al agente)

El asistente (LLM) **propone** texto, diseño y código; quien **decide, revisa y aprueba** es la persona (y el equipo en PR y release). **La responsabilidad final del código que llega a producción es humana**.

Puntos prácticos donde suele merecer la pena **pausar** y mirar con calma (cada equipo marca su ritmo):

1. **Tras extraer o acordar requisitos** (`/go-hu-extract` o texto equivalente): revisar alcance y criterios de aceptación antes de planificar en profundidad.
2. **Tras un borrador de SDD o plan** (`/go-sdd` o documento en `docs/`): revisar diseño (capas, contratos, migraciones) antes de implementar.
3. **Antes de fusionar cambios grandes**: checklist interno del equipo (tests, seguridad, migraciones) según aplique.
4. **Antes de abrir o fusionar un PR**: `make verify`, diff de OpenAPI/config/tests, y `/go-review-pr` o revisión entre pares — el modelo sugiere; **la aprobación es humana**.
5. **Secretos y env**: no integrar cambios que añadan variables sin reflejarlas en `.env.example` y en la config correspondiente.

Verificación del entorno local y de MCP (Go, make, `.env`, `mcp.json` sin exponer tokens).

## Reglas Cursor (`.cursor/rules/`)

| Archivo | Ámbito |
|---------|--------|
| `specify-rules.mdc` | `**/*` (siempre) — stack, estructura y comandos del proyecto |
| `docsy-main-guardrails.mdc` | `**/*` (siempre) — capas Clean Architecture, errores `%w`, env/secretos, verificación local, edición incremental |
| `docsy-main-openapi-contracts.mdc` | transport/openapi/composition — sincronizar OpenAPI al cambiar HTTP |
| `docsy-main-skills.mdc` | Archivos `**/*.go` — índice de skills según tarea |
| `docsy-main-migrations.mdc` | `migrations/**/*.sql` — skill database-queries / migraciones |
| `docsy-main-makefile.mdc` | `Makefile` — targets, lint, mocks |
| `docsy-main-gomod.mdc` | `**/go.mod` — dependencias y `replace` |

## Tests y Object Mother

Al añadir o modificar `*_test.go` bajo **`internal/<slice>/`**:

1. Lee **[`.cursor/skills/go-testing/SKILL.md`](../.cursor/skills/go-testing/SKILL.md)**.
2. Obtén datos de dominio desde **`internal/<slice>/test_utils`** (`NewAuthStubs` / `NewUsersStubs`, `Scenarios`, **`Clone*`** para variar un campo sin literales masivos).
3. Tras cambios, ejecuta **`make test`** y **`make verify`** (golangci-lint, incluye **`dupl`**: si dos tests quedan idénticos, extrae helper con `t.Helper()` o table-driven; no dupliques cuerpos de test).

Excepciones (`pkg/`, `cmd/`, ciclos de import): en el mismo skill y en [`docs/ARCHITECTURE.md`](ARCHITECTURE.md).

## Módulo path

```
github.com/yovannylopez/docsy-main
```

Al crear un proyecto nuevo a partir de este boilerplate, reemplaza este path en `go.mod`, `go.work` y todos los `pkg/*/go.mod`.
