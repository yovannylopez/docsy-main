# Módulo Users (`internal/users`)

Bounded context de **gestión de usuarios y perfiles** expuesto por HTTP. La **persistencia canónica** y las entidades de usuario viven en **auth**; aquí se aplica el patrón **Anti-Corruption Layer** mediante el puerto `UserProfileRepository` y el adaptador hacia `auth/domain/ports.UserRepository`.

## Rol del BC vs Auth

| Auth | Users (este módulo) |
|------|---------------------|
| Login, sesiones, JWT, políticas de seguridad en repositorio | API administrativa: listado, búsqueda, alta masiva, actualización de perfil vía casos de uso |
| `UserRepository` completo (seguridad + datos) | Solo métodos de **perfil / listado** en `UserProfileRepository` |

Las entidades en `domain/entities` son **alias de tipos** hacia `internal/auth/domain/entities` (opción A del informe de auditoría): una sola definición canónica.

## Estructura real del código

```
internal/users/
├── domain/
│   ├── entities/
│   │   ├── user.go                 # type User = authEntities.User (y Role, …)
│   │   └── identification_type.go # tipos de identificación válidos + ValidateIdentificationType
│   ├── errors/user_errors.go      # errores de dominio (ErrUserNotFound, ErrBatchSizeExceeded, …)
│   ├── dtos/
│   └── ports/
│       ├── user_profile_repository.go
│       ├── user_repository.go     # PasswordHistoryRepository, UserService (otros contratos)
│       └── password_hasher.go
├── usecases/
│   ├── create_users_usecase.go    # PasswordHasher inyectado; límites de lote vía pkg/constants.MaxUsersBatchSize
│   ├── get_users_usecase.go
│   ├── get_user_by_id_usecase.go
│   ├── search_users_usecase.go
│   └── update_user_usecase.go
├── infrastructure/
│   ├── repositories/user_repository_adapter.go  # UserProfileRepository → auth.UserRepository
│   ├── adapters/
│   ├── container/users_container.go
│   └── openapi/
├── mocks/                         # mockery: UserProfileRepository, PasswordHasher, …
└── transport/
    ├── handlers/users_handler.go  # recibe use cases inyectados (no los instancia)
    └── routes/users_routes.go
```

## Dependencias

- **`auth/domain/ports.UserRepository`**: inyectado en `NewUsersContainer` desde `cmd/composition` (implementación concreta del BC auth).
- **`ports.PasswordHasher`**: misma abstracción que en auth; implementación bcrypt fuera del caso de uso.
- **No** se importan entidades de auth en el adaptador salvo indirectamente vía tipos alias en `entities.User` (misma estructura en memoria).

## Rutas HTTP (`/api/v1/users`)

Registradas en `transport/routes/users_routes.go`:

| Método | Ruta | Handler | Notas |
|--------|------|---------|--------|
| GET | `/api/v1/users` | Listado paginado | `limit` / `offset` |
| GET | `/api/v1/users/search` | Búsqueda | query `q`, paginación, `activo` opcional |
| GET | `/api/v1/users/:id` | Detalle | |
| POST | `/api/v1/users` | Creación individual o batch | body: un usuario o `{ "usuarios": [...] }` |
| PATCH | `/api/v1/users/:id` | Actualización parcial | header `X-User-ID` |
| GET | `/api/v1/users/profile` | Perfil del actor | `X-User-ID` |
| PUT | `/api/v1/users/profile` | **501** (no implementado) | |
| POST | `/api/v1/users/change-password` | **501** | responsabilidad típica de auth |
| POST | `/api/v1/users/reset-password` | **501** | idem |

No hay **DELETE** de usuario en este router; no figura en el diseño actual.

## Composición (ejemplo)

```go
usersC := usersContainer.NewUsersContainer(authUserRepo, passwordHasher)
routes.RegisterUserRoutes(e, usersC.GetUsersHandler())
```

## Testing

- Use cases: `go test ./internal/users/usecases/...`
- Handlers: `go test ./internal/users/transport/handlers/...`
- Dominio (`identification_type`): `go test ./internal/users/domain/entities/...`
- Mocks: `make generate-users-mocks` (ver `Makefile`)

## Documentación de arquitectura

- Informe actualizado: `docs/architecture/users_audit_report.md`

## Referencias

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- DDD: Shared Kernel / ACL entre users y auth según decisión documentada en el informe de auditoría.
