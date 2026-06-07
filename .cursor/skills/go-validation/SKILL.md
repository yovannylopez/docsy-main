---
name: go-validation
description: >-
  Validación en transport vs dominio/use cases, pkg/validators, Echo Bind,
  respuestas pkg/responses; alineado a Clean Architecture en docs/ARCHITECTURE.md
  y auditoría pkg en docs/architecture/.
  Usar al validar DTOs, reglas de negocio o errores HTTP.
---

# Go: validación (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** el **dominio** no debe asumir entrada “sucia”; validar en **transport** (forma) y en **dominio / casos de uso** (invariantes e invariantes de negocio).
- Informes de alineación del paquete: **`docs/architecture/pkg_validators_audit_report.md`** (cuando exista).

## Capas

1. **Transport (handler):** formato del cuerpo, tipos básicos, requeridos (tras `c.Bind` o lectura JSON).
2. **Dominio / casos de uso:** reglas de negocio, políticas, estado (p. ej. `domain/policies/` en auth).

## Paquete `pkg/validators`

- **`ValidationError`** (`Field`, `Message`, `Value`); **`ClientMessage()`** / **`ErrorClientMessage(err)`** para respuestas sin filtrar datos innecesarios.
- Reutilizar validadores del paquete antes de duplicar lógica.

## go-playground/validator

- **Opcional:** `github.com/go-playground/validator/v10` en DTOs de entrada.
- **`pkg/responses.Validate`** está **deprecated**; validar en handler o use case.

## Respuestas HTTP

- **`pkg/responses`** y **`pkg/http_status`**; códigos **400 / 422** según patrones ya establecidos en el API.

## Resumen

- Transport + dominio; `pkg/validators` + reglas en la capa correcta.

## Skills relacionados

- `go-conventions`
- `go-errors`
- `go-api-rest`
