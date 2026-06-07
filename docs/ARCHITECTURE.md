# Arquitectura — docsy-main

## Patrón: Vertical Slicing + Clean Architecture

Cada módulo de negocio es un **vertical slice** completamente independiente que contiene todas sus capas. No existen dependencias cruzadas entre módulos de negocio.

```
docsy-main/
├── cmd/
│   ├── main.go                        # Punto de entrada
│   └── composition/                   # Wiring: container, router, bootstrap, OpenAPI
│       ├── bootstrap.go               # Application lifecycle
│       ├── container.go               # DI raíz — instancia todos los módulos
│       ├── router.go                  # Registro de rutas por módulo
│       └── openapi_setup.go           # Agregación de specs OpenAPI
│
├── internal/
│   ├── auth/                          # Módulo: autenticación y auditoría
│   │   ├── domain/
│   │   │   ├── dtos/                  # Request / Response DTOs
│   │   │   ├── entities/              # User, Role, Session, AuthToken
│   │   │   └── ports/                 # Interfaces: UserRepository, LoginService…
│   │   ├── infrastructure/
│   │   │   ├── container/             # AuthContainer (DI)
│   │   │   ├── identity/              # LDAP provider, noop provider
│   │   │   ├── openapi/               # Specs OpenAPI del módulo
│   │   │   ├── repositories/          # sqlx adapters: user, session, audit
│   │   │   ├── security/              # JWT token generator, bcrypt hasher
│   │   │   └── services/              # AuditService
│   │   ├── mocks/                     # Mocks generados por Mockery
│   │   ├── transport/
│   │   │   ├── handlers/              # AuthHandler, SignupHandler, AuditHandler
│   │   │   ├── middleware/            # JWT auth middleware, permission checks
│   │   │   └── routes/                # AuthRoutes, AuditRoutes
│   │   ├── test_utils/                # Stubs y helpers de test del módulo
│   │   └── usecases/                  # LoginUseCase, SignupUseCase, ListAuditLogsUseCase
│   │
│   ├── users/                         # Módulo: perfiles y gestión de usuarios
│   │   ├── domain/
│   │   │   ├── entities/              # User profile entity
│   │   │   └── ports/                 # UserRepository, UserService…
│   │   ├── infrastructure/
│   │   │   ├── adapters/              # UserInfoProviderAdapter
│   │   │   ├── container/             # UsersContainer (DI)
│   │   │   ├── openapi/               # Specs OpenAPI del módulo
│   │   │   └── repositories/          # sqlx adapter
│   │   ├── mocks/                     # Mocks generados por Mockery
│   │   ├── test_utils/                # Object Mother: DTOs/requests de users para tests
│   │   ├── transport/
│   │   │   ├── handlers/              # UsersHandler
│   │   │   └── routes/                # RegisterUserRoutes
│   │   └── usecases/                  # GetUsers, GetUserByID, CreateUser, UpdateUser, SearchUsers
│   │
│   └── shared/                        # Shared Kernel — solo utilidades transversales
│       ├── infrastructure/
│       │   ├── config/                # CoreConfig (env vars, pool, Redis, LDAP)
│       │   ├── migrations/            # Wrapper go-migrate
│       │   └── openapi/               # Health spec OpenAPI
│       ├── mocks/                     # Mocks compartidos
│       ├── test_utils/                # Echo stubs y helpers de test
│       └── transport/
│           ├── handlers/              # HealthHandler (único handler genérico)
│           ├── middleware/            # CentralHTTPErrorHandler, CORS
│           └── routes/                # HealthRoutes
│
├── migrations/core/                   # SQL 000000–000004 (base del boilerplate)
│   ├── 000000_initial.*               # Schema mínimo
│   ├── 000001_create_initial_enum_types.*
│   ├── 000002_create_auth_module_tables.*
│   ├── 000003_create_sessions_and_security.*
│   └── 000004_create_audit_logs.*
│
├── pkg/                               # Módulos Go reutilizables (go.work workspace)
│   ├── config/                        # BaseConfig, GetBaseConfig, GetEnv* helpers
│   ├── constants/                     # Constantes globales
│   ├── databases/                     # Pool PostgreSQL + circuit breaker + migrate
│   ├── errors/                        # AppError tipado con ErrorType
│   ├── http_status/                   # Status semánticos (OK, Created, NotFound…)
│   ├── logging/                       # Fachada zap + campos estándar
│   ├── openapi/                       # Generador OpenAPI + Swagger middleware
│   ├── pagination/                    # Params, Metadata, Response
│   ├── ratelimit/                     # Middleware auth rate limit (Redis/memory)
│   ├── responses/                     # OK, Created, BadRequest, MapDomainError…
│   └── validators/                    # StringValidator, EmailValidator…
│
├── _template/TEMPLATE_MODULE/         # Plantilla canónica para nuevos módulos
├── scripts/scaffold_module.sh         # Generador automático: make scaffold MODULE=x
├── .cursor/
│   ├── rules/                         # Reglas Cursor por ámbito (Go, Makefile, go.mod)
│   └── skills/                        # 10 skills con guías prácticas por área técnica
└── docs/                              # Arquitectura, SDD, linting, circuit breaker
```

---

## Capas por módulo

```
HTTP Request
    │
    ▼
transport/handlers     ← valida input, llama use case, formatea respuesta
    │
    ▼
usecases/              ← orquesta lógica de negocio, sin conocer HTTP ni DB
    │
    ▼
domain/ports/          ← interfaces (contratos)
    │
    ▼
infrastructure/        ← implementaciones: sqlx repos, JWT, bcrypt, email
    │
    ▼
PostgreSQL / Redis / LDAP
```

**Regla de dependencias:** las capas internas no importan capas externas. `domain` no conoce `infrastructure`. `usecases` no conoce `transport`.

---

## Principios clave

### 1. Vertical Slicing
Cada módulo (`auth/`, `users/`, `<tumodulo>/`) es autónomo. Agregar un módulo nuevo no toca código existente — solo `cmd/composition/`.

### 2. Ports & Adapters
- **Puertos** = interfaces en `domain/ports/`
- **Adaptadores** = implementaciones en `infrastructure/`
- Los use cases dependen de puertos, nunca de adaptadores concretos

### 3. Container Pattern
Cada módulo tiene su propio container (DI) que instancia sus dependencias. El container raíz en `cmd/composition/container.go` los compone.

### 4. Shared Kernel mínimo
`internal/shared/` contiene solo lo genuinamente transversal: health check, error handler central, config, migrations. **No** contiene lógica de dominio.

### 5. Datos de prueba (Object Mother) — obligatorio en `internal/`

Cada bounded context expone **`internal/<slice>/test_utils`** con factories y escenarios de prueba para entidades y DTOs de ese slice (referencias: auth `NewAuthStubs` + README en `internal/auth/test_utils`, users `NewUsersStubs` + `internal/users/test_utils/README.md`). Los tests del slice **deben** obtener objetos de dominio desde ahí, aplicando solo **deltas** sobre copias (`CloneUser`, `CloneLoginRequest`, etc.) cuando el caso lo requiera. Así se evita deriva de literales duplicados y se alinea con Mockery (mocks = puertos; mothers = datos).

- **Nuevos módulos:** al usar `make scaffold`, añade **`internal/<nombre>/test_utils/`** con el Object Mother del slice antes de acumular literales en tests.
- **Shared:** `internal/shared/test_utils` sigue centrado en Echo, HTTP y config; no duplicar aquí payloads de auth/users: importar el `test_utils` del slice correspondiente.
- **`pkg/`:** ver excepciones en `.cursor/skills/go-testing/SKILL.md` (tests mínimos con literales; `testutil` local si hay repetición).
- **Lint:** `make verify` ejecuta golangci-lint (p. ej. `dupl`). Tests con el mismo flujo deben compartir helper o tabla, no copiar el cuerpo de otra `Test*` (detalle en `.cursor/skills/go-testing/SKILL.md`).

---

## Agregar un nuevo módulo

```bash
make scaffold MODULE=products
```

El script genera la estructura completa. Luego registra en `cmd/composition/`:

1. **`container.go`** — instancia `ProductsContainer` y sus dependencias
2. **`router.go`** — registra las rutas bajo el grupo `protected`
3. **`openapi_setup.go`** — llama a `productsOpenAPI.SetupProductsSpec(generator)`
4. **`bootstrap.go`** — agrega `openapiGen.AddTag("products", "...")`
5. Crea migración SQL en `migrations/core/000005_create_products_table.*`

Ver [`_template/TEMPLATE_MODULE/README.md`](../_template/TEMPLATE_MODULE/README.md) para el checklist completo.

---

## Flujo de autenticación

```
POST /api/v1/auth/login
    → AuthHandler.Login
    → LoginUseCase.Execute(email, password)
    → UserRepository.GetByEmail          # Puerto → sqlx adapter
    → PasswordHasher.Compare             # Puerto → bcrypt adapter
    → SessionRepository.RevokeAllUserSessions  # sesión única (SDD 005)
    → TokenGenerator.GenerateToken       # Puerto → JWT adapter (claim session_id)
    → SessionRepository.Create           # refresh token hasheado en BD
    → AuthToken{AccessToken, RefreshToken}
```

Rutas protegidas requieren `Authorization: Bearer <token>`. El middleware `auth.go` llama a `ValidateToken`, que valida el JWT **y** el estado de la sesión en BD (`session_id`). Ver [`docs/sdd/005-single-session-and-session-validation.md`](sdd/005-single-session-and-session-validation.md).

---

## Documentación relacionada

| Documento | Contenido |
|-----------|-----------|
| [`docs/SDD.md`](SDD.md) | Software Design Document completo |
| [`docs/CIRCUIT_BREAKER.md`](CIRCUIT_BREAKER.md) | Pool PostgreSQL y circuit breaker |
| [`docs/LINTING.md`](LINTING.md) | Perfiles golangci-lint |
| [`_template/TEMPLATE_MODULE/README.md`](../_template/TEMPLATE_MODULE/README.md) | Guía de uso de la plantilla |
