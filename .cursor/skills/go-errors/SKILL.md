---
name: go-errors
description: >-
  %w, errors.Is/As, errores de dominio, pkg/errors (si aplica), mapeo en
  transport con pkg/responses y pkg/http_status según capas en docs/ARCHITECTURE.md.
  Usar al definir errores, clasificar fallos en handlers o propagar desde repos.
---

# Go: manejo de errores (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** errores de **dominio/use case** se traducen a **HTTP** en **transport** (adaptadores de salida), sin filtrar detalles técnicos al cliente.
- **`docs/architecture/pkg_errors_audit_report.md`:** lineamientos del paquete compartido **`pkg/errors`** (si el informe está presente).

## Wrapping

- **`fmt.Errorf("contexto: %w", err)`** para preservar cadena y `errors.Is` / `errors.As`.
- Evitar `%v`/`%s` en errores que deban inspeccionarse aguas arriba (sí en logs).

## Sentinelas y tipos

- `var ErrX = errors.New(...)` + **`errors.Is`**.
- Tipos personalizados en dominio o infra; **`errors.As`** para ramificar.

## Capas

- **Dominio / use cases:** errores de negocio o envueltos desde infra.
- **Transport:** mapeo con **`pkg/responses`** y **`pkg/http_status`** (`EchoError`, `EchoAppError`, `BadRequest`, etc.).
- No exponer mensajes crudos de driver SQL, paths internos ni datos sensibles.

## Paquete `pkg/errors`

- Usar helpers/tipos del paquete cuando el proyecto ya los emplee en ese flujo, para consistencia entre bounded contexts.

## errors.Join (Go 1.20+)

- Varios fallos de validación; considerar comportamiento de `Unwrap` múltiple.

## Evitar

- `panic` en flujos recuperables.
- `_ = err` sin justificación.

## Resumen

- `%w`, `Is`/`As`, mapeo en transport con el contrato del Core.

## Skills relacionados

- `go-conventions`
- `go-validation`
- `go-api-rest`
