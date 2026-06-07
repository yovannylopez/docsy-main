---
name: go-testing
description: >-
  Object Mother en internal/<slice>/test_utils, Clone*, escenarios, table-driven, Mockery,
  dupl/golangci-lint. Usar al escribir tests, regenerar mocks o revisar cobertura.
---

# Go: tests unitarios (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** testing **aislado por módulo**, mocks por vertical slice, interfaces en `domain/ports` para inyectar dobles.
- Mantener **cohesión:** tests y mocks junto al bounded context que ejercitan (ver sección de mocks en el doc de arquitectura).

## Ubicación

- `*_test.go` junto al código o `package xxx_test` cuando convenga.
- **Table-driven** con `t.Run`.
- HTTP: `net/http/httptest`; Echo: `internal/shared/test_utils` y ejemplos en handlers existentes.

## Mocks (Mockery)

- **`internal/<modulo>/mocks/`** por bounded context (ej. `internal/auth/mocks/`).
- **`make generate-<modulo>-mocks`** / **`make generate-mocks`** (`Makefile`); `--dir=.../domain/ports`, `--output=.../mocks`.
- No editar generados a mano.

## Cobertura

- `make test` / `make test-coverage` / `make coverage-html`.

## Testify

- **`github.com/stretchr/testify`** — uso consistente por paquete.

## Object Mother (obligatorio en `internal/`)

**MUST:** En cada bounded context bajo `internal/<slice>/`, los tests que necesiten **entidades o DTOs de dominio del slice** deben obtenerlos desde **`internal/<slice>/test_utils`** (factories tipo `NewAuthStubs`, `NewUsersStubs`, escenarios prearmados, `Clone*` cuando haga falta variar un campo). No volcar structs completos en el cuerpo del test salvo **delta mínimo** sobre una copia devuelta por el mother.

- **Auth:** [`internal/auth/test_utils`](../../../internal/auth/test_utils) — `NewAuthStubs()`, `Entities`, `DTOs`, `Scenarios` (`AuthLoginScenario`, etc.); clones: `CloneUser`, `CloneRole`, `CloneSession`, `CloneAuthToken`, `CloneLoginRequest`, `CloneAuditLog`, `CloneSignupRequest` (copias seguras; UUIDs dinámicos del mother no deben mutarse compartiendo punteros). **`StringPtr(s string) *string`** vive aquí y es el helper canónico para punteros a string en tests de cualquier slice que ya importe este paquete (p. ej. users vía `authtest.StringPtr`).
- **Users:** [`internal/users/test_utils`](../../../internal/users/test_utils) — `NewUsersStubs()`, DTOs de alta/búsqueda/actualización; entidades de listado reutilizan clones del mother de auth donde el dominio es el mismo `User`.
- **Shared:** `internal/shared/test_utils` — Echo, HTTP, config; si hace falta un usuario/DTO de otro slice, **importar** el mother de ese slice, no duplicar payloads de dominio en shared.

**MUST NOT:** Sustituir Mockery; Object Mother es para **datos** (inputs/outputs), no para dobles de puertos.

### Excepciones (documentadas)

- **`pkg/`:** literales OK en tests mínimos de una sola función. Si el mismo struct de prueba se repite en **varios** `*_test.go` del módulo o el payload es grande, añadir helpers en `pkg/<mod>/testutil` o helpers compartidos en el paquete.
- **Deltas intencionales:** `u := CloneUser(stubs.Entities.ValidUser); u.Email = "edge@example.com"` está permitido; construir `&entities.User{...}` con 20 campos desde cero **no**, salvo que aún no exista variante en test_utils (entonces ampliar el mother primero).
- **`cmd/composition`:** si importar `internal/*/test_utils` introduce ciclo, mantener el test con el mínimo literal y anotar en comentario breve `// test_utils: avoid import cycle`.

### Código de test duplicado (`dupl`)

golangci-lint puede marcar **dos funciones `Test*` con el mismo cuerpo** (p. ej. tras alinear ambas al Object Mother). **NO** copiar-pegar el bloque: extrae un helper en el mismo `*_test.go` con `t.Helper()`, o unifica en **table-driven** con `t.Run`. Los datos del caso siguen saliendo del mother; el helper solo orquesta mocks + `Execute` + asserts.

## Verificación antes de PR

- `make test` y **`make verify`** (incluye `dupl` y el resto de linters del perfil).

## Resumen

- `domain/ports` + mocks por módulo + **Object Mother por slice** + comandos del `Makefile`.

## Skills relacionados

- `go-conventions`
- `go-api-rest`
- `go-tooling`
