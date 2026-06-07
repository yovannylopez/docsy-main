---
name: go-context
description: >-
  context.Context en el flujo HTTP → use cases → infra (sqlx, pkg/databases),
  timeouts, cancelación y shutdown; alineado a docs/ARCHITECTURE.md y
  database_infrastructure_context (context-aware, circuit breaker).
  Usar al implementar I/O, queries o goroutines ligadas a requests.
---

# Go: context.Context (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** propagar `ctx` en todo el camino hasta **infrastructure** y **base de datos**; **graceful shutdown** del servidor.
- **[`docs/architecture/database_infrastructure_context.md`](../../../docs/architecture/database_infrastructure_context.md):** conexión y operaciones **context-aware**; operaciones bajo **circuit breaker** deben respetar el `ctx` recibido.

## Propagación

- Primer parámetro `ctx context.Context` en I/O (DB, HTTP externo, colas).
- **Echo:** `c.Request().Context()` hacia use cases y repositorios.

## Timeouts

- `context.WithTimeout` (o deadline del request) + **`defer cancel()`** en repos y llamadas lentas.
- Valores por SLA / entorno; evitar magia sin criterio.

## main y shutdown

- Cancelación en SIGINT/SIGTERM.
- **`e.Shutdown(ctx)`** con timeout dedicado; implementado en **`cmd/composition/bootstrap.go`**.

## Goroutines en background

- No usar solo el `ctx` del request si el trabajo debe seguir tras cerrar la conexión; crear contexto propio con timeout o política explícita.

## Resumen

- Handler → use case → repositorio → DB; `ctx` en todas las fronteras.

## Skills relacionados

- `go-logging`
- `go-conventions`
- `database-queries`
