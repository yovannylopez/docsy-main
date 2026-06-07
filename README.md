# docsy-main

[![Go Version](https://img.shields.io/badge/Go-1.26+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2B%20Vertical%20Slicing-orange.svg)](docs/ARCHITECTURE.md)

API REST en **Go** listo para producción. Incluye autenticación JWT, gestión de usuarios, migraciones automáticas, circuit breaker, logging estructurado y un sistema de scaffolding para agregar módulos de negocio.

---

## Qué incluye

| Capa | Tecnología |
|---|---|
| Framework HTTP | [Echo v4](https://echo.labstack.com/) |
| Base de datos | PostgreSQL + [sqlx](https://github.com/jmoiron/sqlx) |
| Migraciones | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Autenticación | JWT (access + refresh) · bcrypt · LDAP opcional |
| Rate limiting | Redis (o en memoria como fallback) |
| Logging | [zap](https://github.com/uber-go/zap) estructurado |
| Circuit breaker | [gobreaker](https://github.com/sony/gobreaker) |
| Documentación API | OpenAPI / Swagger |
| Linting | golangci-lint |
| Mocks | [mockery](https://github.com/vektra/mockery) |

---

## Estructura

```
docsy-main/
├── cmd/
│   ├── main.go                  # Punto de entrada
│   └── composition/             # Wiring: container, router, bootstrap
├── internal/
│   ├── auth/                    # Autenticación, sesiones, auditoría
│   ├── users/                   # Perfiles y gestión de usuarios
│   └── shared/                  # Health, middleware, config, migraciones
├── migrations/core/             # SQL 000000–000006 (auth, sesiones, auditoría, MFA, …)
├── pkg/                         # Módulos Go reutilizables (go.work)
│   ├── config/                  # BaseConfig + helpers GetEnv*
│   ├── constants/               # Constantes globales
│   ├── databases/               # Pool PostgreSQL + circuit breaker
│   ├── errors/                  # AppError tipado
│   ├── http_status/             # Códigos HTTP semánticos
│   ├── logging/                 # Fachada zap
│   ├── openapi/                 # Generador OpenAPI
│   ├── pagination/              # Parámetros y metadata de paginación
│   ├── ratelimit/               # Middleware rate limiting
│   ├── responses/               # Envoltorio JSON de respuestas HTTP
│   └── validators/              # Validación de DTOs
├── _template/TEMPLATE_MODULE/   # Plantilla para nuevos módulos
├── scripts/
│   ├── scaffold_module.sh         # Generador de módulos
│   ├── setup-agent.sh             # Onboarding: Go, Node/npx (MCP), mcp.json
│   └── verify-dev-environment.sh  # Alias → setup-agent.sh
├── docs/                        # Arquitectura, SDD, linting, circuit breaker
├── .cursor/                     # Comandos /, reglas, skills (Cursor AI)
├── .cursorignore                # Exclusiones del índice de Cursor (build, coverage, .env)
├── .env.example
├── go.mod / go.work
└── Makefile
```

---

## Prerrequisitos

| Herramienta | Versión mínima | Obligatorio |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.26 (ver `go.mod`) | Sí |
| [PostgreSQL](https://www.postgresql.org/) | 14+ | Sí (o vía Docker) |
| [Docker](https://www.docker.com/) | 24+ | No (reemplazable por instancia local) |
| [make](https://www.gnu.org/software/make/) | cualquiera | No (los comandos equivalen a `go ...`) |
| [golangci-lint](https://golangci-lint.run/welcome/install/) | v1.60+ | No (solo para `make verify`) |
| [mockery](https://vektra.github.io/mockery/latest/) | v2 | No (solo para `make generate-mocks`) |

---

## Inicio rápido

### 1. Clonar

```bash
git clone https://github.com/yovannylopez/docsy-main.git
cd docsy-main
```

### 2. Variables de entorno

```bash
cp .env.example .env
# Edita .env con tus valores
```

Mínimo necesario para desarrollo local:

```env
ENVIRONMENT=development
SERVER_PORT=8100

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=go_boilerplate
DB_SSLMODE=disable

JWT_SECRET=change-this-to-a-long-random-secret
```

> **Producción:** define `DATABASE_URL=postgres://user:pass@host/db?sslmode=require` en lugar de los `DB_*` individuales.

### 3. Comprobar entorno y requisitos MCP (opcional)

Comprueba Go, `make`, `.env`, Node/npx (muchas integraciones MCP) y, si existe `~/.cursor/mcp.json`, lista servidores MCP sin mostrar secretos (y avisa env vacíos). Ver [`scripts/setup-agent.sh`](scripts/setup-agent.sh).

```bash
chmod +x scripts/setup-agent.sh
./scripts/setup-agent.sh
# equivalente: ./scripts/verify-dev-environment.sh
```

### 4. Dependencias

```bash
go mod download
go work sync
```

### 5. Ejecutar

**Opción A — local** (requiere PostgreSQL corriendo):

```bash
make dev        # compila y ejecuta
# o
go run ./cmd/
```

**Opción B — con Docker Compose** (levanta PostgreSQL + Redis automáticamente):

```bash
docker compose up -d        # levanta dependencias en segundo plano
make run                    # ejecuta la app apuntando a localhost
```

Para levantar el stack completo (app incluida):

```bash
docker compose --profile full up -d
```

Las migraciones se aplican automáticamente al arrancar.

---

## Endpoints base

### Autenticación

```bash
# Login (obtener access token)
curl -X POST http://localhost:8100/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}'
```

> El usuario administrador inicial se crea en la migración `migrations/core/000002_create_auth_module_tables.up.sql` (credenciales en el bloque seed).

### Usuarios (requiere token y permiso `users.create`)

```bash
# Listar usuarios
curl http://localhost:8100/api/v1/users \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

# Crear usuario (admin-only; por defecto is_active/is_verified = false)
curl -X POST http://localhost:8100/api/v1/users \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecureP@ss1",
    "first_name": "Jane",
    "last_name": "Doe",
    "role_name": "user"
  }'
```

### Health

```bash
curl http://localhost:8100/api/public/health
curl http://localhost:8100/api/public/ready
```

Documentación interactiva: **`/swagger/index.html`**

---

## Usar como base de un nuevo proyecto

```bash
# 1. Clonar sin historial
git clone --depth=1 https://github.com/yovannylopez/docsy-main.git my-api
cd my-api
rm -rf .git && git init

# 2. Renombrar el module path en todos los archivos
OLD="github.com/yovannylopez/docsy-main"
NEW="github.com/tuorg/my-api"

# macOS
find . -type f \( -name "*.go" -o -name "go.mod" -o -name "go.work" \) \
  -not -path "*/vendor/*" \
  -exec sed -i '' "s|$OLD|$NEW|g" {} +

# Linux
find . -type f \( -name "*.go" -o -name "go.mod" -o -name "go.work" \) \
  -not -path "*/vendor/*" \
  -exec sed -i "s|$OLD|$NEW|g" {} +

# 3. Sincronizar workspace
go work sync && go mod tidy
```

Luego actualiza el módulo path en [`docs/agent-guide.md`](docs/agent-guide.md) y el nombre del proyecto donde corresponda.

---

## Agregar un módulo de negocio

```bash
make scaffold MODULE=products
```

Genera la estructura completa (`domain`, `usecases`, `infrastructure`, `transport`, `mocks`) pre-cableada con Clean Architecture. Luego registra el módulo en `cmd/composition/` siguiendo los comentarios del código.

Más detalles en [`_template/TEMPLATE_MODULE/README.md`](_template/TEMPLATE_MODULE/README.md).

---

## Calidad

```bash
make test             # tests unitarios
make test-coverage    # cobertura HTML
make format           # gofumpt + goimports
make verify           # lint completo (CI)
make generate-mocks   # regenera mocks (requiere mockery)
```

---

## Flujo de trabajo con IA (Cursor)

El punto de entrada humano y para el agente es [`AGENTS.md`](AGENTS.md): arquitectura canónica (`docs/`), tabla de **skills**, **[`docs/CONSTITUTION.md`](docs/CONSTITUTION.md)** y **[`docs/agent-guide.md`](docs/agent-guide.md)** (human-in-the-loop, reglas, tests, módulo Go).

### Comandos en el chat (`/`)

En [`.cursor/commands/`](.cursor/commands/) está el flujo **`go-*`**: HU → SDD → implementación → commit → PR → revisión (`go-hu-extract`, `go-sdd`, `go-implement`, `go-bugfix`, `go-refactor`, etc.).

### Reglas (`.cursor/rules/`)

| Regla | Se activa al editar |
|---|---|
| `specify-rules.mdc` | Todo el repo (contexto de stack y estructura) |
| `docsy-main-skills.mdc` | Archivos `.go` |
| `docsy-main-makefile.mdc` | `Makefile` |
| `docsy-main-gomod.mdc` | `go.mod` |

### Skills (`.cursor/skills/`)

Guías que el agente debe leer según la tarea (índice detallado en `AGENTS.md`):

| Skill | Cuándo es relevante |
|---|---|
| `go-conventions` | Crear módulos, capas, bounded contexts |
| `go-api-rest` | Handlers Echo, rutas, middleware HTTP |
| `database-queries` | Repositorios sqlx, SQL, migraciones |
| `go-testing` | Tests, mocks (Mockery), cobertura |
| `go-errors` | Definir y propagar errores entre capas |
| `go-validation` | Validar DTOs, reglas de negocio |
| `go-context` | `context.Context`, timeouts, shutdown |
| `go-logging` | Logging estructurado con zap |
| `go-tooling` | Lint, formato, `make verify` |
| `go-generics` | Helpers y APIs genéricas en `pkg/` |

[`.cursorignore`](.cursorignore) reduce ruido en el índice (binarios, cobertura, `.env`, copias `.md.bak`).

---

## Extensiones recomendadas para producción

El boilerplate cubre el núcleo. Estas son las adiciones más comunes al llevarlo a producción:

| Área | Opción recomendada |
|------|--------------------|
| **Tracing distribuido** | [`go.opentelemetry.io/otel`](https://opentelemetry.io/docs/languages/go/) con exportador OTLP (Jaeger, Tempo, Datadog) |
| **Métricas** | `go.opentelemetry.io/otel/metric` + Prometheus scrape endpoint |
| **Feature flags** | [Flagsmith](https://flagsmith.com) o [Unleash](https://www.getunleash.io) |
| **Caché** | Redis con [`github.com/redis/go-redis/v9`](https://github.com/redis/go-redis) |
| **Colas / eventos** | NATS, Kafka o AWS SQS según escala |
| **Email transaccional** | [Resend](https://resend.com) o [SendGrid](https://sendgrid.com) |

---

## Documentación

| Documento | Contenido |
|---|---|
| [`AGENTS.md`](AGENTS.md) | Índice corto para IA: comandos Cursor, skills |
| [`docs/CONSTITUTION.md`](docs/CONSTITUTION.md) | Principios del proyecto, stack y skills obligatorios |
| [`docs/agent-guide.md`](docs/agent-guide.md) | Human-in-the-loop, reglas `.mdc`, tests, módulo path |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Visión general, capas, patrones |
| [`docs/SDD.md`](docs/SDD.md) | Software Design Document completo |
| [`docs/CIRCUIT_BREAKER.md`](docs/CIRCUIT_BREAKER.md) | Circuit breaker y pool de BD |
| [`docs/LINTING.md`](docs/LINTING.md) | Configuración de golangci-lint |
| [`docs/specs/data_schema.md`](docs/specs/data_schema.md) | Esquema de base de datos |

---

## Licencia

MIT — ver [LICENSE](LICENSE).
