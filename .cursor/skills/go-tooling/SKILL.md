---
name: go-tooling
description: >-
  gofumpt, goimports, go vet, golangci-lint (.golangci*.yml), make verify/format;
  coherencia con lineamientos de docs/ARCHITECTURE.md.
  Usar al configurar CI, corregir lint, revisar go.mod o antes de commit.
---

# Go: linting, formateo y módulos (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md)** define el estándar del proyecto (Go, Echo, sqlx, herramientas).
- El código debe poder **compilar y pasar verificación** (`make verify` / `verify-dev`) sin degradar límites de complejidad acordados en golangci-lint.

Repo: `github.com/yovannylopez/docsy-main`, Go **1.26+** (`go.mod` / `go.work`).

## Formateo

- **`make format`:** `gofumpt` + `goimports` (preferido antes de commit).
- Alternativa: `gofmt -s -w .`

## Módulos

- Tras cambiar dependencias: `go mod tidy`.
- `replace` hacia `pkg/*` locales: no quitar sin revisar imports.

## Verificación

- **`make verify`:** `go vet ./...` + `golangci-lint run` (`.golangci.yml`) — perfil estricto, usado en CI.

## golangci-lint

- **`depguard`:** en `**/internal/**/domain/**/*.go` bloquea imports hacia `infrastructure`/`transport` del mismo u otros bounded contexts, Echo, sqlx y `database/sql` (ver `.golangci.yml`, regla `domain_no_outer_layers`). Al crear `internal/<nuevo>/`, amplía la lista `deny` en ese archivo.

- Configs en la **raíz** del repo; un solo lugar de verdad por entorno.
- El linter **`dupl`** detecta funciones duplicadas (incluidos tests). Si choca con dos `Test*` casi iguales: helper con `t.Helper()` o tests table-driven (ver **`.cursor/skills/go-testing/SKILL.md`** → sección *Código de test duplicado*).

## Resumen

- `make format` + `make verify`.

## Skills relacionados

- `go-conventions`
