---
name: go-api-rest
description: >-
  Echo v4, transport por bounded context en internal/<modulo>/transport,
  shared solo health/middleware genérico, pkg/responses y flujo hacia use cases
  según docs/ARCHITECTURE.md.
  Usar al crear o modificar handlers, middlewares, rutas o tests HTTP.
---

# Go: API HTTP / REST (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** flujo **HTTP Request → Transport → Use Cases → Domain → Infrastructure → Database** y respuesta inversa; **handlers de negocio** viven en cada **`internal/<modulo>/transport/`**.
- **`internal/shared/transport/`:** solo handlers genéricos (p. ej. health) y middleware compartido — no mezclar lógica de dominio de un módulo en `shared` sin consenso arquitectónico (ver `docs/architecture/shared_context.md`).

## Stack

- **Framework:** [Echo v4](https://echo.labstack.com/) (`github.com/labstack/echo/v4`).
- **Respuestas:** `pkg/responses` (envoltorio `status`, `message`, `data` / `error`), `pkg/http_status`, `pkg/pagination` cuando corresponda.
- **Auth:** JWT (`github.com/golang-jwt/jwt/v5`) según **`internal/auth/`** y contrato en `docs/architecture/auth_context.md`.

## Handlers

- Firma: `func(c echo.Context) error`.
- `c.Bind` / helpers del proyecto para JSON; validar tras el bind (**go-validation**).
- Propagar `c.Request().Context()` hacia use cases y repositorios (**go-context**).

## Seguridad

- Entrada validada; SQL parametrizado (**database-queries**).
- Secretos vía **`pkg/config`**, no en código.
- No exponer entidades de dominio con campos sensibles: usar DTOs de respuesta (patrones en `docs/architecture/*_context.md`).

## REST

- Códigos HTTP coherentes con `pkg/http_status` y el contrato de `responses`.
- No exponer detalles de DB, stack traces ni datos fuera del contrato del bounded context.

## Pruebas

- `httptest` + Echo `NewContext` o utilidades en `internal/shared/test_utils`.
- Table-driven tests; mocks en `internal/<modulo>/mocks/`.

## Base de datos

- **PostgreSQL** + **sqlx** + **`lib/pq`**. Placeholders `$1`, `$2`.
- Sin GORM en el stack principal del Core.

## Apagado

- **Graceful shutdown** con `e.Shutdown(ctx)` y timeout dedicado, implementado en `cmd/composition/bootstrap.go`.

## OpenAPI (obligatorio cuando cambia el contrato HTTP)

Si cambias rutas, status codes, JSON de entrada/salida o seguridad del endpoint, actualiza la spec en `internal/<modulo>/infrastructure/openapi/` y el registro en `cmd/composition/` como en `docs/ARCHITECTURE.md`. No cierres la tarea sin `make verify` si tocaste handlers expuestos.

## Skills relacionados

- `go-context`
- `go-errors`
- `go-validation`
- `database-queries`
