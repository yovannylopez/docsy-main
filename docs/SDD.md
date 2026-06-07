# Software Design Document (SDD)
## Go API Docsy Main — docsy-main

| Campo | Valor |
|-------|-------|
| **Versión** | 1.0.0 |
| **Fecha** | Marzo 2026 |
| **Estado** | Aprobado |
| **Autor** | Tu equipo |
| **Repositorio** | `github.com/yovannylopez/docsy-main` |

---

## Tabla de Contenidos

1. [Propósito y Alcance](#1-propósito-y-alcance)
2. [Principios Arquitectónicos](#2-principios-arquitectónicos)
3. [Vista de Componentes](#3-vista-de-componentes)
4. [Descripción de Capas](#4-descripción-de-capas)
5. [Paquetes Compartidos (pkg/)](#5-paquetes-compartidos-pkg)
6. [Flujo de Datos](#6-flujo-de-datos)
7. [Diseño de Seguridad](#7-diseño-de-seguridad)
8. [Diseño de Base de Datos](#8-diseño-de-base-de-datos)
9. [Manejo de Errores](#9-manejo-de-errores)
10. [Estrategia de Testing](#10-estrategia-de-testing)
11. [Observabilidad](#11-observabilidad)
12. [Configuración y Entornos](#12-configuración-y-entornos)
13. [Cómo Agregar un Nuevo Módulo](#13-cómo-agregar-un-nuevo-módulo)
14. [Decisiones de Arquitectura (ADRs)](#14-decisiones-de-arquitectura-adrs)
15. [Glosario](#15-glosario)

---

## 1. Propósito y Alcance

### 1.1 Propósito

Este documento describe el diseño técnico del **Go API Docsy Main** encapsulado en `docsy-main`. El boilerplate provee una base de código lista para producción que resuelve los problemas de infraestructura comunes (autenticación, base de datos, logging, linting, testing) para que los equipos puedan concentrarse exclusivamente en la lógica de negocio.

### 1.2 Alcance

El boilerplate cubre:

- **Estructura del proyecto:** Vertical Slicing + Clean Architecture
- **Autenticación:** JWT con access/refresh tokens, gestión de sesiones y auditoría
- **Acceso a datos:** PostgreSQL con sqlx, pool de conexiones y circuit breaker
- **HTTP:** Echo v4 con middleware estándar y graceful shutdown
- **Observabilidad:** Logging estructurado (zap)
- **Calidad:** golangci-lint, gofumpt, cobertura de tests
- **Documentación de API:** OpenAPI/Swagger auto-generado

El boilerplate **no** cubre: lógica de negocio específica, integraciones de terceros propias del dominio, ni arquitectura de microservicios (es monolito modular).

### 1.3 Audiencia

- Desarrolladores Go que crean una nueva API
- Arquitectos que evalúan el boilerplate para un proyecto
- Ingenieros de DevOps que configuran CI/CD

---

## 2. Principios Arquitectónicos

### P1 — Clean Architecture

Las dependencias fluyen de afuera hacia adentro. El dominio nunca importa infraestructura.

```
┌─────────────────────────────────────────────┐
│  Transport (HTTP)                           │
│  ┌───────────────────────────────────────┐  │
│  │  Use Cases (Application Logic)        │  │
│  │  ┌─────────────────────────────────┐  │  │
│  │  │  Domain (Entities + Ports)      │  │  │
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
│  Infrastructure (DB, JWT, Email, etc.)      │
└─────────────────────────────────────────────┘
```

**Regla de dependencias:** Transport → Use Cases → Domain ← Infrastructure

### P2 — Vertical Slicing

Cada módulo de negocio es un corte vertical completo e independiente.

```
internal/
├── auth/          ← corte vertical completo
│   ├── domain/
│   ├── usecases/
│   ├── infrastructure/
│   ├── transport/
│   └── mocks/
├── products/      ← nuevo módulo, mismo patrón
└── shared/        ← shared kernel (no es un slice de negocio)
```

**Regla:** Los módulos de negocio no se importan entre sí. Si necesitan datos compartidos, usan `internal/shared/`.

### P3 — Puertos y Adaptadores (Hexagonal)

Las interfaces (puertos) se definen en `domain/ports/`. Las implementaciones (adaptadores) viven en `infrastructure/`.

```go
// Puerto — pertenece al dominio, no sabe nada de sqlx ni de PostgreSQL
type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
}

// Adaptador — pertenece a infrastructure, implementa el puerto
type UserRepositoryAdapter struct {
    db *sqlx.DB
}
```

### P4 — Inyección de Dependencias por Container

Cada módulo tiene un `Container` que instancia y cablea sus dependencias. El `RootContainer` en `cmd/composition/` compone todos los módulos.

### P5 — Fail Fast al Inicio

El servidor no arranca si las migraciones fallan, la DB no está disponible, o la configuración es inválida.

### P6 — Errores Explícitos con Wrapping

```go
// Siempre usar %w para mantener la cadena de error
return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
```

---

## 3. Vista de Componentes

### 3.1 Diagrama de Componentes

```
┌──────────────────────────────────────────────────────────────────┐
│                        cmd/composition/                          │
│                                                                  │
│  main.go → Application → Container → Router → OpenAPI           │
└──────────────────┬───────────────────────────────────────────────┘
                   │ instancia y cablea
         ┌─────────┼─────────────────────────────────┐
         │         │                                 │
         ▼         ▼                                 ▼
  ┌──────────┐ ┌──────────┐                  ┌─────────────┐
  │  shared  │ │   auth   │  ... módulos ...  │  <módulo N> │
  │ container│ │ container│                  │  container  │
  └──────────┘ └──────────┘                  └─────────────┘
         │
         ▼
  ┌──────────────────────────────────────────────────┐
  │                     pkg/                         │
  │  config  databases  logging  responses  errors   │
  │  validators openapi  pagination ratelimit        │
  └──────────────────────────────────────────────────┘
         │
         ▼
  ┌──────────────┐
  │  PostgreSQL  │
  └──────────────┘
```

### 3.2 Mapa de Rutas HTTP (Base)

```
/api/
├── public/
│   ├── GET  /health            ← HealthHandler (shared)
│   └── GET  /ready             ← ReadinessHandler (shared)
│
├── auth/
│   ├── POST /signup            ← SignupHandler
│   ├── POST /login             ← LoginHandler
│   ├── POST /refresh           ← RefreshTokenHandler
│   ├── POST /logout            ← LogoutHandler
│   └── GET  /me                ← ProfileHandler (requiere JWT)
│
└── <módulo>/                   ← Rutas del módulo de negocio
    ├── GET  /                  ← List
    ├── POST /                  ← Create
    ├── GET  /:id               ← GetByID
    ├── PUT  /:id               ← Update
    └── DELETE /:id             ← Delete

/swagger/
└── index.html                  ← UI de OpenAPI
```

---

## 4. Descripción de Capas

### 4.1 Capa Domain

**Propósito:** Define el modelo de negocio. No tiene dependencias externas.

**Contenido:**

```
domain/
├── entities/          # Structs de negocio con tags json y db
├── dtos/              # Data Transfer Objects para la API
└── ports/             # Interfaces (contratos)
```

**Reglas:**
- Solo puede importar paquetes de la stdlib de Go
- Nunca importa `pkg/`, `infrastructure/`, ni `transport/`
- Las entidades usan tags `json:"..."` y `db:"..."` para serialización

**Ejemplo de entidad:**

```go
type User struct {
    ID        string    `json:"id"         db:"id"`
    Email     string    `json:"email"      db:"email"`
    Password  string    `json:"-"          db:"password"`
    FirstName string    `json:"first_name" db:"first_name"`
    LastName  string    `json:"last_name"  db:"last_name"`
    IsActive  bool      `json:"is_active"  db:"is_active"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
    Roles     []Role    `json:"roles"      db:"-"`
}
```

### 4.2 Capa Use Cases

**Propósito:** Orquesta la lógica de aplicación usando los puertos del dominio.

**Contenido:**

```
usecases/
├── create_<entity>_usecase.go
├── get_<entity>_usecase.go
├── update_<entity>_usecase.go
└── delete_<entity>_usecase.go
```

**Reglas:**
- Solo importa `domain/entities`, `domain/ports` y `pkg/errors`
- No sabe nada de HTTP, base de datos, ni JWT
- Cada use case es una struct con un método `Execute`
- Usa `context.Context` en todos los métodos

**Patrón estándar de un use case:**

```go
type CreateProductUseCase struct {
    repo    ports.ProductRepository
    logger  logging.Logger
}

func NewCreateProductUseCase(repo ports.ProductRepository) *CreateProductUseCase {
    return &CreateProductUseCase{repo: repo}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, dto *dtos.CreateProductRequest) (*entities.Product, error) {
    // 1. Validar reglas de negocio
    if dto.Price <= 0 {
        return nil, domainerrors.ErrInvalidPrice
    }

    // 2. Crear entidad
    product := &entities.Product{
        ID:    uuid.New().String(),
        Name:  dto.Name,
        Price: dto.Price,
    }

    // 3. Persistir
    if err := uc.repo.Create(ctx, product); err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    return product, nil
}
```

### 4.3 Capa Infrastructure

**Propósito:** Implementaciones concretas de los puertos del dominio.

**Contenido:**

```
infrastructure/
├── container/         # DI container del módulo
├── repositories/      # Adaptadores de base de datos (sqlx)
└── security/          # JWT, bcrypt (solo en auth)
```

**Reglas:**
- Implementa las interfaces definidas en `domain/ports/`
- Puede importar `pkg/databases`, `pkg/logging`, sqlx, etc.
- Los repositorios siempre usan `Context` en sus métodos

**Patrón de repositorio:**

```go
type ProductRepositoryAdapter struct {
    db *sqlx.DB
}

func (r *ProductRepositoryAdapter) Create(ctx context.Context, p *entities.Product) error {
    const query = `
        INSERT INTO products (id, name, price, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    _, err := r.db.ExecContext(ctx, query, p.ID, p.Name, p.Price, p.CreatedAt, p.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to insert product: %w", err)
    }
    return nil
}

func (r *ProductRepositoryAdapter) GetByID(ctx context.Context, id string) (*entities.Product, error) {
    var p entities.Product
    const query = `SELECT id, name, price, created_at, updated_at FROM products WHERE id = $1`
    if err := r.db.GetContext(ctx, &p, query, id); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domainerrors.ErrNotFound
        }
        return nil, fmt.Errorf("failed to get product %s: %w", id, err)
    }
    return &p, nil
}
```

### 4.4 Capa Transport

**Propósito:** Interfaz HTTP. Traduce requests/responses entre HTTP y use cases.

**Contenido:**

```
transport/
├── handlers/          # Echo handlers
├── middleware/         # Middleware específico del módulo
└── routes/            # Registro de rutas en Echo
```

**Reglas:**
- Solo puede importar use cases (a través de puertos/interfaces) y `pkg/responses`
- Nunca importa repositorios directamente
- El binding y la validación se hacen aquí, antes de llamar al use case
- Los errores de dominio se mapean a HTTP status codes aquí

**Patrón de handler:**

```go
type ProductsHandler struct {
    createUC ports.CreateProductService
    getUC    ports.GetProductService
}

func (h *ProductsHandler) Create(c echo.Context) error {
    var req dtos.CreateProductRequest
    if err := c.Bind(&req); err != nil {
        return responses.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
    }

    if errs := validators.ValidateStruct(req); len(errs) > 0 {
        return responses.NewValidationErrorResponse(c, errs)
    }

    product, err := h.createUC.Execute(c.Request().Context(), &req)
    if err != nil {
        return httpstatus.MapDomainError(c, err)
    }

    return responses.NewSuccessResponse(c, http.StatusCreated, product)
}
```

### 4.5 Container (DI)

**Propósito:** Instancia y cablea todas las dependencias de un módulo.

```go
type ProductsContainer struct {
    Repository    ports.ProductRepository
    CreateUseCase ports.CreateProductService
    GetUseCase    ports.GetProductService
    Handler       *handlers.ProductsHandler
}

func NewProductsContainer(db *sqlx.DB) (*ProductsContainer, error) {
    repo := repositories.NewProductRepositoryAdapter(db)

    createUC := usecases.NewCreateProductUseCase(repo)
    getUC := usecases.NewGetProductUseCase(repo)

    handler := handlers.NewProductsHandler(createUC, getUC)

    return &ProductsContainer{
        Repository:    repo,
        CreateUseCase: createUC,
        GetUseCase:    getUC,
        Handler:       handler,
    }, nil
}
```

---

## 5. Paquetes Compartidos (pkg/)

Todos los paquetes en `pkg/` son módulos Go independientes (tienen su propio `go.mod`) y están referenciados mediante `go.work` + directivas `replace`.

### 5.1 `pkg/config`

Carga la configuración del servidor desde variables de entorno y archivos `.env`.

```go
type CoreConfig struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Log      LogConfig
}

cfg, err := config.NewCoreConfig(".env")
```

### 5.2 `pkg/constants`

Constantes globales de la aplicación: duraciones de tokens, límites de paginación, nombres de claims JWT.

```go
const (
    AccessTokenExpirationHours  = 24
    RefreshTokenExpirationHours = 168
    DefaultPageSize             = 20
    MaxPageSize                 = 100
)
```

### 5.3 `pkg/databases`

Pool de conexiones PostgreSQL con circuit breaker integrado.

```go
// Conectar con circuit breaker
db, err := databases.NewPostgresDB(databases.Config{
    DSN:             cfg.Database.DSN(),
    MaxOpenConns:    25,
    MaxIdleConns:    10,
    ConnMaxLifetime: 5 * time.Minute,
})

// El circuit breaker se activa automáticamente ante fallos consecutivos
```

**Circuit breaker:** Implementado con `gobreaker`. Se abre después de N fallos en una ventana de tiempo y permite recuperación parcial (half-open).

### 5.4 `pkg/errors`

Tipos de error de dominio y utilidades de clasificación.

```go
var (
    ErrNotFound          = errors.New("not found")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrConflict          = errors.New("conflict")
    ErrValidation        = errors.New("validation error")
    ErrInternalServer    = errors.New("internal server error")
)

// Verificar tipo de error
if errors.Is(err, pkgerrors.ErrNotFound) {
    return responses.NewErrorResponse(c, http.StatusNotFound, err.Error())
}
```

### 5.5 `pkg/http_status`

Mapea errores de dominio a HTTP status codes automáticamente.

```go
// En el handler, una sola línea para mapear cualquier error de dominio
return httpstatus.MapDomainError(c, err)
```

### 5.6 `pkg/logging`

Wrapper sobre `go.uber.org/zap` con campos estructurados predefinidos.

```go
logging.Init(false)  // false = producción, true = desarrollo (verbose)

logging.Info("User created", 
    logging.WithRequestID(reqID),
    zap.String("user_id", userID),
)
logging.Error("Database error", zap.Error(err))
logging.Warn("Slow query detected", zap.Duration("duration", d))
```

### 5.7 `pkg/openapi`

Generador de especificaciones OpenAPI 3.0 a partir de las rutas de Echo.

```go
gen := openapi.NewGenerator("MyAPI", "Description", "1.0.0")
gen.AddServer("http://localhost:8100", "Development")
gen.AddTag("products", "Product operations")

// Auto-generar desde rutas registradas en Echo
gen.GenerateFromEcho(e)

// Agregar especificaciones manuales (para request/response bodies)
gen.AddOperation("POST", "/api/products", openapi.OperationSpec{...})
```

### 5.8 `pkg/pagination`

Utilidades de paginación para endpoints de listado.

```go
// Parsear parámetros de la request
page := pagination.ParsePage(c)  // page, limit, offset

// Construir respuesta paginada
return responses.NewPaginatedResponse(c, http.StatusOK, items, page)
```

### 5.9 `pkg/responses`

Formatos de respuesta HTTP estandarizados.

```go
// Éxito con datos
responses.NewSuccessResponse(c, http.StatusOK, data)

// Error simple
responses.NewErrorResponse(c, http.StatusBadRequest, "message")

// Error de validación con campos
responses.NewValidationErrorResponse(c, []validators.FieldError{...})

// Lista paginada
responses.NewPaginatedResponse(c, http.StatusOK, items, pageInfo)
```

Formato de respuesta de error:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "email", "message": "must be a valid email" }
    ]
  }
}
```

### 5.10 `pkg/validators`

Validación de structs con mensajes de error enriquecidos.

```go
type CreateProductRequest struct {
    Name  string  `json:"name"  validate:"required,min=2,max=100"`
    Price float64 `json:"price" validate:"required,gt=0"`
}

errs := validators.ValidateStruct(req)
// []FieldError{ {Field: "price", Message: "must be greater than 0"} }
```

### 5.12 `pkg/ratelimit`

Middleware de rate limiting por IP o por token JWT.

```go
// En el router de auth
authGroup.Use(ratelimit.NewAuthRateLimiter(5, time.Minute))  // 5 req/min
```

---

## 6. Flujo de Datos

### 6.1 Flujo de una Request HTTP

```
Cliente HTTP
    │
    ▼ POST /api/products
┌─────────────────────────────────────────────────────┐
│  Echo Framework (transport layer)                   │
│                                                     │
│  1. Middleware pipeline:                            │
│     - Recovery (panic → 500)                        │
│     - CORS                                          │
│     - RequestID (genera X-Request-ID)               │
│     - Logger (zap, logs request/response)           │
│     - JWT (valida token si la ruta lo requiere)     │
│                                                     │
│  2. Handler.Create(c echo.Context)                  │
│     a. c.Bind(&req)        → parse JSON             │
│     b. ValidateStruct(req) → validar campos         │
│     c. createUC.Execute(ctx, &req) ──────────────┐  │
└──────────────────────────────────────────────────│──┘
                                                   │
┌──────────────────────────────────────────────────▼──┐
│  Use Case (application layer)                       │
│                                                     │
│  Execute(ctx, dto):                                 │
│     a. Validar reglas de negocio                    │
│     b. Construir entidad de dominio                 │
│     c. repo.Create(ctx, entity) ─────────────────┐  │
└──────────────────────────────────────────────────│──┘
                                                   │
┌──────────────────────────────────────────────────▼──┐
│  Repository Adapter (infrastructure layer)          │
│                                                     │
│  Create(ctx, entity):                               │
│     a. Construir query SQL parametrizado            │
│     b. db.ExecContext(ctx, query, params...)        │
│     c. Retornar error wrapeado si aplica            │
└─────────────────────────────────────────────────────┘
    │
    ▼
PostgreSQL
    │
    ▼ (respuesta viaja de vuelta por el mismo stack)
Cliente HTTP ← 201 Created { "success": true, "data": {...} }
```

### 6.2 Flujo de Autenticación

```
POST /api/v1/auth/login  { email, password }
    │
    ▼ LoginHandler.Handle(c)
    │  ├── Bind + Validate DTO
    │  └── loginUC.Execute(ctx, email, password)
    │
    ▼ LoginUseCase.Execute(ctx, email, password)
    │  ├── userRepo.GetByEmail(ctx, email)      → *User o ErrNotFound
    │  ├── passwordHasher.Compare(plain, hash)  → bool
    │  └── tokenGenerator.Generate(*User)       → *AuthToken
    │
    ▼ TokenGeneratorAdapter.Generate(*User)
    │  ├── Crear JWT claims (user_id, email, role, exp, iat)
    │  ├── Firmar con HMAC-SHA256 + secretKey
    │  ├── Crear refresh token (UUID firmado)
    │  └── Guardar sesión en DB (session_repository)
    │
    ▼ Response 200 OK
    {
      "access_token":  "eyJ...",
      "refresh_token": "eyJ...",
      "token_type":    "Bearer",
      "expires_at":    "2026-04-01T12:00:00Z"
    }
```

---

## 7. Diseño de Seguridad

### 7.1 Autenticación JWT

| Aspecto | Decisión |
|---------|----------|
| Algoritmo | HMAC-SHA256 (HS256) |
| Access token TTL | 24h (configurable via `TOKEN_EXPIRATION`) |
| Refresh token TTL | 168h / 7 días (configurable via `REFRESH_DURATION`) |
| Claims estándar | `user_id`, `email`, `role`, `exp`, `iat` |
| Almacenamiento de sesión | PostgreSQL tabla `sessions` (refresh token hasheado) |
| Rotación de refresh | Al usar el refresh token se genera un par nuevo |

**Flujo de validación JWT:**

```go
// pkg/ratelimit o internal/auth/transport/middleware
func JWTMiddleware(secret string) echo.MiddlewareFunc {
    return echojwt.WithConfig(echojwt.Config{
        SigningKey:  []byte(secret),
        TokenLookup: "header:Authorization:Bearer ",
        ErrorHandler: func(c echo.Context, err error) error {
            return responses.NewErrorResponse(c, http.StatusUnauthorized, "invalid or expired token")
        },
    })
}
```

### 7.2 Hashing de Contraseñas

- Algoritmo: **bcrypt** con cost factor 12
- Nunca se almacena la contraseña en texto plano
- La comparación se hace en tiempo constante para prevenir timing attacks

```go
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}

func (h *BcryptPasswordHasher) Compare(plain, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
```

### 7.3 Auditoría

Todas las operaciones de autenticación (login, logout, signup, refresh, failed attempts) se registran en la tabla `audit_logs` con:
- `user_id` (puede ser null si el usuario no existe)
- `action` (LOGIN, LOGOUT, SIGNUP, REFRESH, FAILED_LOGIN)
- `ip_address`
- `user_agent`
- `created_at`

### 7.4 Rate Limiting

El módulo `pkg/ratelimit` aplica rate limiting en endpoints sensibles:
- `/api/v1/auth/login` → 5 intentos / minuto por IP
- `/api/v1/auth/signup` → 3 intentos / minuto por IP

### 7.5 CORS

Configurado en `internal/shared/transport/middleware/cors.go`:

```go
middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: cfg.AllowedOrigins,  // desde .env
    AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
})
```

---

## 8. Diseño de Base de Datos

### 8.1 Esquema Base (siempre presente en el boilerplate)

```sql
-- 000001: Tipos enum base
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE role_name   AS ENUM ('admin', 'user', 'viewer');
CREATE TYPE audit_action AS ENUM ('LOGIN', 'LOGOUT', 'SIGNUP', 'REFRESH', 'FAILED_LOGIN');

-- 000002: Módulo Auth
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    first_name  VARCHAR(100) NOT NULL,
    last_name   VARCHAR(100) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    status      user_status NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        role_name UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- 000003: Sesiones
CREATE TABLE sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) NOT NULL,
    ip_address    INET,
    user_agent    TEXT,
    expires_at    TIMESTAMPTZ NOT NULL,
    revoked       BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 000004: Audit Logs
CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID REFERENCES users(id),
    action      audit_action NOT NULL,
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 8.2 Convenciones de Nomenclatura

| Elemento | Convención | Ejemplo |
|----------|-----------|---------|
| Tablas | `snake_case` plural | `audit_logs` |
| Columnas | `snake_case` | `first_name` |
| PKs | UUID con `gen_random_uuid()` | `id UUID PRIMARY KEY DEFAULT gen_random_uuid()` |
| FKs | `<tabla_ref>_id` | `user_id UUID REFERENCES users(id)` |
| Índices | `idx_<tabla>_<columna>` | `idx_users_email` |
| Migraciones | `<secuencia>_<descripcion>.<up|down>.sql` | `000005_create_products.up.sql` |

### 8.3 Pool de Conexiones

| Parámetro | Valor por defecto | Variable de entorno |
|-----------|------------------|---------------------|
| MaxOpenConns | 25 | `DB_MAX_OPEN_CONNS` |
| MaxIdleConns | 10 | `DB_MAX_IDLE_CONNS` |
| ConnMaxLifetime | 5m | `DB_CONN_MAX_LIFETIME` |

### 8.4 Circuit Breaker

Configurado en `pkg/databases/circuitbreaker.go`:

| Estado | Condición de transición |
|--------|------------------------|
| Closed (normal) | → Open: 5 fallos consecutivos en 60 segundos |
| Open (rechaza requests) | → Half-Open: después de 30 segundos |
| Half-Open (prueba) | → Closed: 1 éxito / → Open: 1 fallo |

---

## 9. Manejo de Errores

### 9.1 Jerarquía de Errores

```
pkg/errors/
├── ErrNotFound          → HTTP 404
├── ErrUnauthorized      → HTTP 401
├── ErrForbidden         → HTTP 403
├── ErrConflict          → HTTP 409
├── ErrValidation        → HTTP 422
├── ErrBadRequest        → HTTP 400
└── ErrInternalServer    → HTTP 500
```

### 9.2 Propagación de Errores por Capa

```
Infrastructure (repositorio)
    │ fmt.Errorf("failed to get product %s: %w", id, sql.ErrNoRows)
    │                                                 ↑ wrapeado
    ▼
Use Case
    │ Detecta: errors.Is(err, sql.ErrNoRows) → retorna domainerrors.ErrNotFound
    │ O propaga: fmt.Errorf("get product use case: %w", err)
    ▼
Transport (handler)
    │ httpstatus.MapDomainError(c, err) → mapea ErrNotFound → 404
    ▼
Cliente HTTP ← 404 { "success": false, "error": { "code": "NOT_FOUND", ... } }
```

### 9.3 Formato de Respuesta de Error

```json
{
  "success": false,
  "error": {
    "code":    "NOT_FOUND",
    "message": "product with id abc123 not found"
  }
}
```

Para errores de validación:
```json
{
  "success": false,
  "error": {
    "code":    "VALIDATION_ERROR",
    "message": "validation failed",
    "details": [
      { "field": "name",  "message": "name is required" },
      { "field": "price", "message": "price must be greater than 0" }
    ]
  }
}
```

### 9.4 Panic Recovery

El middleware de Echo `middleware.Recover()` captura panics y los convierte en respuestas HTTP 500 sin crashear el servidor.

---

## 10. Estrategia de Testing

### 10.1 Pirámide de Tests

```
          /\
         /  \     E2E / Integration (pocos, lentos)
        /────\
       /      \   Handler Tests con Echo + httptest
      /────────\
     /          \  Use Case Tests (mocks de repositorios)
    /────────────\
   /              \ Unit Tests (dominio, pure functions)
  /────────────────\
```

### 10.2 Tests Unitarios — Use Cases

```go
func TestCreateProductUseCase_Execute(t *testing.T) {
    tests := []struct {
        name    string
        dto     *dtos.CreateProductRequest
        mockFn  func(*mocks.ProductRepository)
        wantErr bool
    }{
        {
            name: "success",
            dto:  &dtos.CreateProductRequest{Name: "Widget", Price: 9.99},
            mockFn: func(m *mocks.ProductRepository) {
                m.On("Create", mock.Anything, mock.AnythingOfType("*entities.Product")).Return(nil)
            },
            wantErr: false,
        },
        {
            name:    "invalid price",
            dto:     &dtos.CreateProductRequest{Name: "Widget", Price: -1},
            mockFn:  func(m *mocks.ProductRepository) {},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &mocks.ProductRepository{}
            tt.mockFn(mockRepo)

            uc := usecases.NewCreateProductUseCase(mockRepo)
            product, err := uc.Execute(context.Background(), tt.dto)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, product)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, product)
            }
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### 10.3 Tests de Handlers (HTTP)

```go
func TestProductsHandler_Create(t *testing.T) {
    e := echo.New()
    mockCreateUC := &mocks.CreateProductService{}

    mockCreateUC.On("Execute", mock.Anything, mock.Anything).
        Return(&entities.Product{ID: "123", Name: "Widget", Price: 9.99}, nil)

    handler := handlers.NewProductsHandler(mockCreateUC, nil)

    body := `{"name":"Widget","price":9.99}`
    req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(body))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()

    c := e.NewContext(req, rec)
    err := handler.Create(c)

    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, rec.Code)
    mockCreateUC.AssertExpectations(t)
}
```

### 10.4 Generación de Mocks

Los mocks se generan con [Mockery](https://github.com/vektra/mockery):

```bash
# Generar mocks para un módulo
make generate-products-mocks

# El Makefile ejecuta:
mockery --dir=internal/products/domain/ports \
        --output=internal/products/mocks \
        --outpkg=mocks \
        --all
```

### 10.5 Cobertura

```bash
make test-coverage       # Genera coverage.out
make coverage-html       # Abre reporte HTML en el browser
```

Meta de cobertura mínima: **70%** en use cases y handlers.

---

## 11. Observabilidad

### 11.1 Logging

Todos los logs son JSON estructurado (zap en modo production).

```go
logging.Info("HTTP request",
    logging.WithRequestID(c.Request().Header.Get(echo.HeaderXRequestID)),
    zap.String("method", c.Request().Method),
    zap.String("path", c.Request().URL.Path),
    zap.Int("status", c.Response().Status),
    zap.Duration("latency", time.Since(start)),
)
```

**Niveles de log:**

| Nivel | Cuándo usarlo |
|-------|--------------|
| `Debug` | Solo en desarrollo, detalles internos |
| `Info` | Eventos de negocio importantes (startup, request, auth) |
| `Warn` | Situaciones degradadas pero recuperables |
| `Error` | Fallos que requieren atención (DB errors, panics recuperados) |

### 11.2 Health Checks

```
GET /api/public/health  → { "status": "ok", "timestamp": "..." }
GET /api/public/ready   → { "status": "ok", "database": "ok" }  (verifica conexión DB)
```

---

## 12. Configuración y Entornos

### 12.1 Jerarquía de Configuración

```
Valores default en código
    ↓ (sobreescrito por)
Archivo .env
    ↓ (sobreescrito por)
Variables de entorno del sistema
```

### 12.2 Referencia de Variables

```env
# ─── Servidor ────────────────────────────────────────
SERVER_HOST=0.0.0.0
SERVER_PORT=8100
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s

# ─── Base de datos (opción A: campos individuales) ───
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=myapp_db
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m

# ─── Base de datos (opción B: URL completa) ───────────
# DATABASE_URL tiene prioridad sobre DB_* individuales
DATABASE_URL=postgresql://user:pass@host:5432/db?sslmode=require

# ─── JWT ─────────────────────────────────────────────
JWT_SECRET=cambia-esto-en-produccion-min-32-chars
TOKEN_EXPIRATION=24h
REFRESH_DURATION=168h

# ─── Migraciones ─────────────────────────────────────
MIGRATIONS_PATH=migrations/core          # opcional, auto-detectado

# ─── Logging ─────────────────────────────────────────
LOG_LEVEL=info                           # debug | info | warn | error
LOG_DEVELOPMENT=false                    # true para logs legibles

# ─── Tracing (opcional) ──────────────────────────────
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_SERVICE_NAME=my-service

# ─── CORS ────────────────────────────────────────────
ALLOWED_ORIGINS=http://localhost:3000,https://myapp.com
```

### 12.3 Entornos

| Entorno | `LOG_DEVELOPMENT` | `DB_SSLMODE` | `JWT_SECRET` |
|---------|---------------------|--------------|--------------|
| Development | `true` | `disable` | Cualquiera |
| Staging | `false` | `require` | Secreto seguro |
| Production | `false` | `require` | Secreto rotado |

---

## 13. Cómo Agregar un Nuevo Módulo

### 13.1 Usando el Script de Scaffolding

```bash
scripts/scaffold_module.sh orders
```

El script genera la estructura completa en `internal/orders/`.

### 13.2 Estructura Generada

```
internal/orders/
├── domain/
│   ├── dtos/
│   │   ├── create_order_request.go
│   │   └── order_response.go
│   ├── entities/
│   │   └── order.go
│   └── ports/
│       ├── order_repository.go
│       └── order_service.go
│
├── infrastructure/
│   ├── container/
│   │   └── orders_container.go
│   └── repositories/
│       └── order_repository_adapter.go
│
├── transport/
│   ├── handlers/
│   │   └── orders_handler.go
│   ├── middleware/               # vacío, se agrega si se necesita
│   └── routes/
│       └── orders_routes.go
│
├── usecases/
│   ├── create_order_usecase.go
│   └── get_order_usecase.go
│
└── mocks/                        # generado por Mockery
```

### 13.3 Registrar en la Composición

**1. Agregar el container al Root Container** (`cmd/composition/container.go`):

```go
type Container struct {
    Shared  *shared.SharedContainer
    Auth    *auth.AuthContainer
    Orders  *orders.OrdersContainer   // ← agregar
}

func NewContainer(cfg *config.CoreConfig) (*Container, error) {
    // ...
    ordersContainer, err := orders.NewOrdersContainer(shared.DB)
    // ...
    return &Container{
        Orders: ordersContainer,
    }, nil
}
```

**2. Registrar rutas** (`cmd/composition/router.go`):

```go
func (r *Router) SetupRoutes() {
    // ...rutas existentes
    ordersRoutes.NewOrdersRoutes(r.e, r.container.Orders).Register(jwtMiddleware)
}
```

**3. Agregar tag OpenAPI** (`cmd/composition/bootstrap.go`):

```go
openapiGen.AddTag("orders", "Order management operations")
```

**4. Crear migración:**

```bash
# Nombrar con el siguiente número de secuencia disponible
touch migrations/core/000005_create_orders.up.sql
touch migrations/core/000005_create_orders.down.sql
```

**5. Agregar target de mocks en Makefile:**

```makefile
generate-orders-mocks:
    mockery --dir=internal/orders/domain/ports \
            --output=internal/orders/mocks \
            --outpkg=mocks --all
```

**6. Verificar:**

```bash
make generate-orders-mocks
make format
make verify
make test
```

---

## 14. Decisiones de Arquitectura (ADRs)

### ADR-001: Echo como framework HTTP

**Estado:** Aceptado

**Contexto:** Necesitamos un framework HTTP con buen soporte de middleware, routing con grupos y parámetros, y bajo overhead.

**Decisión:** Usar Echo v4.

**Consecuencias:**
- (+) API limpia, grupos de rutas, middleware configurable
- (+) Bind/Validate integrado
- (+) Buena performance (comparable con Gin)
- (-) Dependencia de terceros (alternativa sería `net/http` puro)

---

### ADR-002: sqlx en lugar de ORM

**Estado:** Aceptado

**Contexto:** Necesitamos acceso a PostgreSQL con control total del SQL y sin magia de un ORM.

**Decisión:** Usar sqlx sobre `database/sql`.

**Consecuencias:**
- (+) SQL explícito, fácil de auditar
- (+) Mapeo directo con tags `db:"..."` en structs
- (+) Sin overhead de reflexión de ORM
- (-) Más código de boilerplate en repositorios
- (-) Sin migraciones automáticas de struct → usar go-migrate

---

### ADR-003: go.work + módulos locales en pkg/

**Estado:** Aceptado

**Contexto:** Los paquetes de infraestructura necesitan ser reutilizables y eventualmente publicables.

**Decisión:** Cada `pkg/*` es un módulo Go independiente con su propio `go.mod`, referenciado via `go.work` y `replace` en desarrollo.

**Consecuencias:**
- (+) Publicables como librerías independientes
- (+) Versionables de forma independiente
- (-) `go work sync` necesario al agregar dependencias
- (-) IDE necesita configuración adicional para go.work

---

### ADR-004: Auto-migración al inicio del servidor

**Estado:** Aceptado con condiciones

**Contexto:** Necesitamos que las migraciones se apliquen antes de aceptar tráfico.

**Decisión:** `app.Run()` aplica migraciones con rollback automático si fallan.

**Condición:** En producción multi-replica, usar la variable `MIGRATIONS_PATH=""` + un job de migración separado, o `advisory_lock` de PostgreSQL para serializar migraciones.

---

### ADR-005: Mocks generados por Mockery

**Estado:** Aceptado

**Contexto:** Los tests necesitan mocks de las interfaces de dominio.

**Decisión:** Usar Mockery para generar mocks automáticamente desde las interfaces en `domain/ports/`.

**Consecuencias:**
- (+) Mocks siempre sincronizados con las interfaces
- (+) `make generate-mocks` como step explícito
- (-) Dependencia de la herramienta Mockery en el entorno de desarrollo

---

## 15. Glosario

| Término | Definición |
|---------|-----------|
| **Bounded Context** | Límite explícito dentro del cual un modelo de dominio es consistente. En este proyecto cada carpeta en `internal/` es un bounded context |
| **Vertical Slice** | Organización del código por funcionalidad de negocio (vs. por capa técnica). Cada slice incluye todas sus capas |
| **Clean Architecture** | Patrón donde las dependencias fluyen hacia el centro (dominio), con capas: Domain → Use Cases → Infrastructure/Transport |
| **Puerto** | Interface definida en el dominio que representa una capacidad abstracta (ej. `UserRepository`) |
| **Adaptador** | Implementación concreta de un puerto. Vive en infrastructure |
| **Container** | Struct que instancia y cablea todas las dependencias de un módulo (DI manual) |
| **Shared Kernel** | Código compartido entre módulos que no pertenece a ningún dominio de negocio (`internal/shared/`) |
| **Circuit Breaker** | Patrón de resiliencia que detiene las llamadas a un servicio degradado para evitar cascada de fallos |
| **DTOs** | Data Transfer Objects. Structs que representan los datos de entrada/salida de la API, distintos de las entidades de dominio |
| **go.work** | Archivo de workspace Go que agrupa múltiples módulos en un mismo directorio para desarrollo local |
| **Mockery** | Herramienta de generación de mocks para interfaces Go, compatible con testify/mock |
