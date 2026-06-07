# Shared Kernel (`internal/shared`)

Este paquete **no es un bounded context**: es el **Shared Kernel** y la **infraestructura transversal** que pueden usar todos los módulos sin acoplarse entre sí por medio de `shared`.

## Qué debe vivir aquí

| Área | Contenido |
|------|-----------|
| **Dominio mínimo** | Tipos técnicos compartidos, p. ej. `AuditModel` (timestamps) en `domain/entities`. |
| **Configuración** | `CoreConfig` y derivados por ambiente (`infrastructure/config`). |
| **Persistencia transversal** | Migraciones del núcleo / esquema común (`infrastructure/migrations`). |
| **OpenAPI mínimo** | Spec de health u otros endpoints que **no** pertenecen a un BC (`infrastructure/openapi`). |
| **Transporte transversal** | Health checks, middleware global (CORS, errores, request ID), rutas solo de health (`transport/`). |
| **Test utils** | Stubs de **configuración** reutilizables (`test_utils/`). Sin entidades de auth ni de otros BC. |

## Qué no debe vivir aquí

- **Orquestación de bounded contexts**: contenedor único, registro de rutas de todos los módulos, `SetupAllSpecs` agregando specs de auth, empresas, etc. Eso vive en **`cmd/composition/`** (composition root).
- **Lógica o adaptadores propios de un BC**: por ejemplo hashing de contraseñas → `internal/auth`, repositorios de un dominio → su slice vertical.
- **Imports hacia `internal/<otro_contexto>`**: `shared` no debe importar bounded contexts. La dependencia correcta es BC → shared (p. ej. embebido de `AuditModel`), no al revés.

## Añadir infraestructura compartida (checklist)

1. ¿Es realmente **genérica** (varios BC) o es **específica de un dominio**? Si es la segunda, va en ese BC.
2. Si añades código en `shared`, confirma con `grep` que **no** aparecen imports a `internal/auth`, `internal/empresas`, etc.
3. Expón contratos estables (tipos claros, pocos puntos de extensión). Evita “god packages”.
4. Añade o amplía **tests** junto al código (config, middleware, migraciones ya tienen precedentes en este módulo).

## Estructura actual

```
internal/shared/
├── domain/entities/       # AuditModel
├── infrastructure/
│   ├── config/
│   ├── migrations/
│   └── openapi/           # health + nota de compatibilidad SetupAllSpecs → cmd/composition
├── transport/
│   ├── handlers/          # health
│   ├── middleware/        # CORS, ErrorHandler, RequestID, cadena HTTP
│   └── routes/            # solo health (`health_routes_test.go` integra Setup + handler)
└── test_utils/            # stubs de config para tests
```

## Referencias

- Composition root: `cmd/composition/` (`container.go`, `router.go`, `openapi_setup.go`, `bootstrap.go`).
- Auditoría: `docs/architecture/shared_audit_report.md`.
