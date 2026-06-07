---
name: go-conventions
description: >-
  Vertical slicing y bounded contexts en internal/, Shared Kernel en shared/,
  domain/usecases/infrastructure/transport, puertos en domain/ports, pkg/ y
  lineamientos de docs/ARCHITECTURE.md y docs/architecture/.
  Usar al crear módulos, paquetes o refactors de capas.
---

# Go: estructura y convenciones (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** vertical slicing, Clean Architecture (domain → use cases → infrastructure → transport), inversión de dependencias (puertos en dominio, adaptadores en infra), flujo HTTP → transport → use cases → domain → infrastructure → DB.
- **[`docs/architecture/`](../../../docs/architecture/):** contratos por **bounded context** (p. ej. `auth_context.md`), **Shared Kernel** (`shared_context.md`), infraestructura de BD (`database_infrastructure_context.md`), informes de auditoría de módulos y `pkg/*`.

## Vertical slicing y bounded contexts

- Cada carpeta bajo **`internal/<modulo>/`** (salvo **`shared`**) es un **módulo de negocio** con todas las capas; en DDD se documentan como **bounded contexts** (ver `docs/architecture/*_context.md` del módulo).
- **`internal/shared`** es **Shared Kernel**: infraestructura técnica compartida (config, servidor Echo, migraciones, middleware transversal, health). **No** debe convertirse en depósito de reglas de negocio de dominios — ver `docs/architecture/shared_context.md`.
- **Principio:** evitar **dependencias directas entre módulos de negocio**; integrar vía composición en `cmd` / contenedor, contratos en `domain/ports` o mecanismos explícitos. Los informes `*_audit_report.md` en `docs/architecture/` documentan desviaciones y trabajo de alineación.

## Raíz del repositorio

- **`cmd/`** — punto de entrada.
- **`internal/`** — código privado del módulo Go.
- **`pkg/`** — librerías reutilizables con `replace` en `go.mod` (config, responses, validators, logging, databases, etc.).
- **`migrations/core/`** — SQL versionado; coherente con **`docs/specs/data_schema.md`**.

## Por módulo (`internal/<modulo>/`)

Estructura alineada a `docs/ARCHITECTURE.md` y a los contextos existentes:

- **`domain/entities/`**, **`domain/dtos/`**, **`domain/ports/`** — contratos que el dominio exige a la infraestructura.
- **`domain/policies/`**, **`domain/repositories/`**, **`domain/services/`** — solo si el módulo los usa (algunos contextos declaran repositorios/servicios de dominio en documentación).
- **`usecases/`** — casos de uso / aplicación.
- **`infrastructure/`** — adaptadores (`repositories/`, `security/`, `container/`, `openapi/`, etc.).
- **`transport/handlers/`**, **`transport/middleware/`**, **`transport/routes/`** — HTTP (Echo).
- **`mocks/`** — Mockery por módulo.

Interfaces en **`domain/ports`:** nombres descriptivos (`UserRepository`, `LoginService`); no es obligatorio el sufijo `Port`.

## Nombres

- **Exportados:** `PascalCase`; no exportados: `camelCase`.
- **JSON / DB:** `snake_case` en tags según esquema (`docs/specs/data_schema.md`).
- **Archivos:** `snake_case` (`auth_handler.go`, `login_usecase.go`).

## Puertos y dependencias

- **Dependency inversion:** el dominio define interfaces; la infraestructura implementa.
- **Container pattern:** contenedores por módulo (`*_container.go`) + composición global según `docs/ARCHITECTURE.md`.

## Resumen

| Área | Ubicación típica |
|------|------------------|
| Entidades | `internal/<modulo>/domain/entities/` |
| Contratos | `internal/<modulo>/domain/ports/` |
| Casos de uso | `internal/<modulo>/usecases/` |
| Adaptadores | `internal/<modulo>/infrastructure/` |
| HTTP | `internal/<modulo>/transport/` |
| Infra transversal | `internal/shared/` (Shared Kernel) |

## Skills relacionados

- `go-errors`
- `go-validation`
- `go-tooling`
