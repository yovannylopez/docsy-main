# TEMPLATE_MODULE — Plantilla de Módulo

Esta carpeta es la **plantilla canónica** para crear un nuevo módulo vertical en el boilerplate.

## Uso

```bash
# Copiar y renombrar usando el script de scaffolding (recomendado)
scripts/scaffold_module.sh <nombre_del_modulo>

# O manualmente
cp -r _template/TEMPLATE_MODULE internal/products
find internal/products -type f | xargs sed -i 's/TEMPLATE_MODULE/products/g'
find internal/products -type f | xargs sed -i 's/Template/Product/g'
find internal/products -type f | xargs sed -i 's/template/product/g'
```

## Estructura

```
TEMPLATE_MODULE/
├── domain/
│   ├── dtos/               ← Request/Response DTOs de la API
│   ├── entities/           ← Entidades de dominio (structs de negocio)
│   └── ports/              ← Interfaces: repository + service ports
│
├── infrastructure/
│   ├── container/          ← DI container: cablea todas las dependencias
│   └── repositories/       ← Adaptadores sqlx para PostgreSQL
│
├── transport/
│   ├── handlers/           ← Echo handlers (HTTP in/out)
│   └── routes/             ← Registro de rutas en Echo
│
└── usecases/               ← Lógica de aplicación (orquesta ports)
```

## Checklist al Crear un Módulo Nuevo

```
[ ] Renombrar: TEMPLATE_MODULE → nombre_modulo (snake_case en carpetas/SQL)
[ ] Renombrar: Template → NombreModulo (PascalCase en Go)
[ ] Actualizar imports: github.com/org/repo/internal/TEMPLATE_MODULE → /nombre_modulo
[ ] Crear migración SQL en migrations/core/<N+1>_create_nombre_modulo.up.sql
[ ] Agregar container al RootContainer en cmd/composition/container.go
[ ] Registrar rutas en cmd/composition/router.go
[ ] Agregar tag OpenAPI en cmd/composition/bootstrap.go
[ ] Agregar target generate-<modulo>-mocks en Makefile
[ ] Ejecutar make generate-<modulo>-mocks
[ ] make format && make verify && make test
```
