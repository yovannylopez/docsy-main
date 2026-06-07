---
name: database-queries
description: >-
  SQL parametrizado sqlx, PostgreSQL, pkg/databases (pool, circuit breaker),
  migraciones migrations/core, wrapper shared/migrations y data_schema.md.
  Alineado a docs/architecture/database_infrastructure_context.md.
  Usar al escribir repositorios, migraciones SQL o revisar rendimiento.
---

# Base de datos y consultas (docsy-main)

## Arquitectura de referencia

- **[`docs/architecture/database_infrastructure_context.md`](../../../docs/architecture/database_infrastructure_context.md):** `pkg/databases` (conexión PostgreSQL, pool, **circuit breaker**, operaciones **context-aware**), migraciones en **`migrations/core/`**, wrapper en **`internal/shared/infrastructure/migrations`**.
- **`docs/specs/data_schema.md`**: fuente de verdad del esquema antes de cambiar structs o SQL.

## PostgreSQL + sqlx

- **Siempre** consultas parametrizadas; no concatenar datos de usuario en SQL.
- Preferir `GetContext`, `SelectContext`, `ExecContext` con `context.Context` (cancelación y timeouts propagados desde handlers).
- Columnas NULL en SQL → punteros o `sql.Null*` en Go cuando corresponda.
- Manejar `sql.ErrNoRows` explícitamente en lecturas de una fila.

## Capa de infraestructura compartida

- Conexión y resiliencia: usar **`pkg/databases`** según el arranque del servicio (incl. circuit breaker donde esté configurado). No duplicar lógica de pool/retry si ya la centraliza ese paquete.

## Transacciones

- `BeginTxx` / transacciones explícitas para operaciones multi-tabla o multi-paso; **rollback** en error.

## Migraciones

- Ficheros SQL en **`migrations/core/`** (up/down numerados, idempotentes cuando el proyecto lo exige).
- Ejecución y políticas: coherentes con **`database_infrastructure_context.md`** y el código en `pkg/databases/migrate` / shared.

## Rendimiento

- Índices alineados con filtros y ordenaciones frecuentes; `EXPLAIN` en entornos adecuados.
- Timeouts: ver **go-context**; las operaciones deben respetar el `ctx` recibido.

## No incluido en el stack principal

- **GORM** y **MongoDB** no son el stack principal del Core; si se añadiera otro almacén, documentarlo y mantener parametrización y contexto.

## Resumen

- sqlx + pq, `$n`, transacciones, migraciones versionadas, `pkg/databases` y `data_schema.md`.

## Skills relacionados

- `go-context`
- `go-api-rest`
