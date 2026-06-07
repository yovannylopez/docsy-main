# docsy-main Constitution

## Core Principles

### I. Clean Architecture y vertical slicing

Las dependencias van de afuera hacia adentro: el dominio no importa infraestructura. Cada capacidad vive en un bounded context bajo `internal/<contexto>/` con capas **transport → use cases → domain → infrastructure**, puertos en **domain/ports** y adaptadores en infrastructure. El Shared Kernel permanece en **`shared/`**; utilidades transversales reutilizables en **`pkg/`**.

### II. Documentación canónica y skills

La implementación y las revisiones deben alinearse con **`docs/ARCHITECTURE.md`** y **`docs/SDD.md`**. **`AGENTS.md`** (entrada e índice de skills), **`docs/agent-guide.md`** (human-in-the-loop, tabla de reglas, tests, módulo path), **`.cursor/rules/`** y **`.cursor/skills/<área>/SKILL.md`** son guías operativas obligatorias. Si hay conflicto entre un borrador informal y el repo, **prevalecen `docs/` y `AGENTS.md`**.

### III. Calidad, tests y tooling

Antes de dar por cerrado un cambio sustancial, seguir **`go-tooling`**: formato, `go vet`, `golangci-lint` y **`make verify`** (o el target acordado en el skill). Nuevas piezas críticas llevan tests alineados a **`go-testing`** (tablas, mocks con Mockery donde aplique). No degradar cobertura ni patrones de test del módulo sin acuerdo explícito.

### IV. API, seguridad y observabilidad

HTTP con **Echo v4**; respuestas y errores según convenciones del repo (**`pkg/responses`**, **`go-errors`**). Autenticación/autorización acordes al diseño existente (p. ej. JWT). Entrada validada según **`go-validation`**. Logging estructurado con **`pkg/logging`** (**`go-logging`**). Contratos HTTP documentados en OpenAPI cuando el flujo del proyecto lo exija.

### V. Alcance y simplicidad

Cambios **acotados al alcance de la tarea o feature**: sin refactors amplios, renombres masivos ni archivos no relacionados. Preferir soluciones simples y coherentes con el código existente; complejidad nueva debe estar justificada en la documentación o plan acordado con el equipo.

## Stack y restricciones técnicas

- **Lenguaje**: Go (versión del `go.mod` del repositorio).
- **HTTP**: Echo v4.
- **Datos**: PostgreSQL con **sqlx**, migraciones y consultas según **`database-queries`** y **`go-context`** en I/O.
- **Módulo Go**: `github.com/yovannylopez/docsy-main` (en forks o derivados, el path puede sustituirse de forma consistente en `go.mod`, `go.work` y `pkg/*/go.mod`).

Las especificaciones y planes no deben asumir otro stack sin decisión explícita documentada (p. ej. ADR o actualización acordada de `docs/SDD.md`).

## Flujo de trabajo y skills por tarea

Antes de implementar, el agente debe **leer el `SKILL.md` adecuado** según la tabla de **`AGENTS.md`**, entre otros:

| Área | Skill |
|------|-------|
| Estructura de módulos y capas | go-conventions |
| Handlers y REST | go-api-rest |
| SQL y migraciones | database-queries |
| Validación de DTOs | go-validation |
| Errores y HTTP | go-errors |
| Contexto y timeouts | go-context |
| Logging | go-logging |
| Tests y mocks | go-testing |
| Generics en pkg | go-generics |
| Lint y Make | go-tooling |

## Governance

Esta constitución **complementa** `docs/ARCHITECTURE.md` y `docs/SDD.md`; no las sustituye. Cualquier excepción debe quedar escrita en el SDD, en notas de diseño o en un ADR. Las enmiendas a esta constitución deben reflejarse en este archivo y, si aplica, en la documentación del equipo.

**Version**: 1.0.0 | **Ratified**: 2026-04-02 | **Last Amended**: 2026-05-09
