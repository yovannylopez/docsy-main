---
name: go-logging
description: >-
  pkg/logging (zap), middleware en shared; cross-cutting
  alineado a Shared Kernel y docs/architecture/pkg_logging_audit_report.md.
  Usar al añadir logs, instrumentar requests o revisar observabilidad.
---

# Go: logging y trazas (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** logging/recovery como **middleware** en **`internal/shared`**; **graceful shutdown** y liberación de recursos al cerrar.
- **`docs/architecture/shared_context.md`:** middleware de logging como parte del Shared Kernel (técnico, no dominio).
- **`docs/architecture/pkg_logging_audit_report.md`:** auditoría del paquete **`pkg/logging`**.

## Logging (`pkg/logging`)

- **`go.uber.org/zap`** — **`Init(production bool)`**, **`Sync()`** al apagar.
- API global: `Info`, `Error`, `Warn`, `Debug`, `Logger()`; campos en **`fields.go`** (`WithRequestID`, `WithError`, etc.).
- Hasta **Init**, el logger es `zap.NewNop()` (sin panic).

## Buenas prácticas

- No loguear secretos, tokens ni PII innecesaria.
- Campos estructurados de Zap; atributos estables.

## Niveles (cuándo usar cada uno)

- **Debug**: diagnóstico detallado solo en desarrollo o troubleshooting (no depender de Debug en producción para lógica de negocio).
- **Info**: eventos de negocio o técnicos esperados (arranque, migración OK, operación completada).
- **Warn**: degradación o situación anómala no fatal (reintentos, fallback, rate limit cercano).
- **Error**: fallo que requiere atención (I/O fallida tras reintentos, invariantes rotas); incluir `WithError` / request id, **sin** secretos ni PII innecesaria.

## Resumen

- **zap** vía `pkg/logging`.

## Skills relacionados

- `go-context`
- `go-errors`
- `go-conventions`
