# Linting Options for Docsy Main

Este documento describe las opciones de linting disponibles y cuándo usar cada una.

## Opciones disponibles

### `make verify` — Desarrollo diario ⭐ RECOMENDADO

Usa `.golangci.yml`: activa todos los linters esenciales, omite los que generan falsos positivos en desarrollo (`wsl`, `lll`, `misspell`, `gomoddirectives`). Exluye también errores de `gomoddirectives` en `go.mod` cuando se trabaja con paquetes locales.

```bash
make verify
```

### `make verify` — CI / producción

Usa `.golangci.yml`: versión completa y estricta con todos los linters habilitados (`wsl`, `lll`, `misspell`, `gomoddirectives` incluidos).

```bash
make verify
```

## Recomendaciones por escenario

| Escenario | Comando |
|-----------|---------|
| Desarrollo diario | `make verify` |
| Pre-commit | `make verify` |
| CI/CD Pipeline | `make verify` |
| Edición rápida (solo fmt/imports) | `make lint-basic` |

## Archivos de configuración

| Archivo | Propósito |
|---------|-----------|
| `.golangci.yml` | CI / producción — máxima calidad |
| `.golangci-dev.yml` | Desarrollo diario — sin linters ruidosos |

## Instalación de herramientas

```bash
make install-lint         # golangci-lint
make install-basic-tools  # gofumpt, goimports
```

## Flujo de trabajo recomendado

1. **Desarrollo**: `make verify`
2. **Pre-commit**: `make verify`
3. **CI/CD**: `make verify`

## Notas

- `wsl`, `lll` y `misspell` están deshabilitados en `verify` para reducir el ruido en desarrollo.
- `gomoddirectives` se excluye en `verify` cuando se usan paquetes locales con `replace`.

## Arquitectura hexagonal — `depguard`

El linter **`depguard`** (habilitado en `.golangci.yml`) aplica la regla **`domain_no_outer_layers`** a todos los archivos bajo **`internal/<bounded_context>/domain/**`** (incluye `internal/auth/domain/`, `internal/users/domain/`, `internal/shared/domain/`, etc.).

Queda **prohibido** importar, entre otros:

- `.../internal/<módulo>/infrastructure` y `.../internal/<módulo>/transport` para cada bounded context listado (hoy: `auth`, `users`, `shared`).
- `github.com/labstack/echo/v4`, `github.com/jmoiron/sqlx`, `database/sql`.

**Nuevo módulo** (`internal/products/`, …): añade en `.golangci.yml` dos entradas `deny` con prefijo de import del módulo, por ejemplo `github.com/yovannylopez/docsy-main/internal/products/infrastructure` y `.../internal/products/transport` (ajusta el path del módulo en `go.mod` si el fork cambia el nombre).

Los **use cases** no tienen regla `depguard` estricta todavía: el código actual importa helpers desde `infrastructure/security` en algunos flujos; cerrar eso es refactor opcional.
