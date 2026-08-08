# Capa web HTMX (server-rendered)

Documentación de la interfaz web server-rendered integrada con el módulo Auth.

## Arquitectura

Patrón **dual transport** por bounded context:

| Capa | Archivos | Responsabilidad |
|------|----------|-----------------|
| API JSON | `transport/handlers/*_handler.go`, `transport/routes/*_routes.go` | REST `/api/v1/...` |
| Web HTML | `transport/handlers/*_page_handler.go`, `transport/routes/web_*_routes.go` | HTMX + `html/template` |

Los page handlers **no duplican lógica de negocio**: llaman directamente a los puertos existentes (`ports.LoginService`, `ports.AuthenticationService`, etc.).

```
Browser (HTMX)
    │
    ▼
Echo Router
    ├── /static/*          → web/static (CSS, JS, assets)
    ├── /login             → LoginPageHandler (HTML)
    ├── /                  → ShowHome (protegida, WebAuthMiddleware)
    └── /api/v1/auth/login → AuthHandler (JSON, sin cambios)
```

## Estructura de archivos

```
web/
├── styles/app.css          # Fuente Tailwind
├── static/
│   ├── css/app.css         # CSS compilado (make web-build)
│   ├── js/htmx.min.js
│   ├── js/app.js
│   └── assets/             # Iconos e imágenes
├── templates/
│   ├── layouts/            # base, auth, app
│   ├── components/         # modales reutilizables (confirmación, éxito, shell)
│   ├── partials/           # alerts, form-field, pagination, navbar, sidebar, etc.
│   ├── auth/               # login, login_result, mfa_unavailable
│   └── home.gohtml
└── package.json            # Solo dev: Tailwind CLI

internal/shared/infrastructure/templates/renderer.go  # echo.Renderer
internal/auth/transport/handlers/login_page_handler.go
internal/auth/transport/routes/web_auth_routes.go
internal/auth/transport/middleware/web_auth.go
```

## Autenticación web (MVP)

### Login HTMX

1. `GET /login` — página completa con formulario.
2. `POST /login` — partial HTMX (`hx-target="#login-feedback"`).
3. Éxito: header `HX-Redirect: /` + partial `login_result` con script que guarda tokens en `sessionStorage`.
4. Errores de dominio → partial `partials/alerts` (422).

### Sesión

Tras login exitoso se establecen **dos mecanismos** complementarios:

1. **Cookie `access_token` (HttpOnly, SameSite=Lax):** permite que `GET /` y otras navegaciones completas pasen `WebAuthMiddleware`.
2. **`sessionStorage.access_token`:** inyectado vía `htmx:configRequest` para peticiones HTMX/AJAX con header `Authorization`.

En logout se borran ambos (cookie en servidor, `sessionStorage` en cliente).

### Rutas protegidas

`WebAuthMiddleware.RequireAuth()` aplica a rutas web protegidas (`/`, `/usuarios`, `/auditoria`, etc.).  
`WebAuthMiddleware.RequirePermission("<perm>")` valida RBAC en rutas web (403 HTML si se deniega). No afecta `/api` ni `/static`.

## Variables de entorno

| Variable | Default | Uso |
|----------|---------|-----|
| `TEMPLATES_PATH` | `./web/templates` (búsqueda desde cwd) | Directorio de plantillas `.gohtml` |
| `STATIC_PATH` | `./web/static` | Assets estáticos servidos en `/static` |

## Desarrollo

```bash
# Compilar CSS (requiere Node solo en dev)
make web-build

# Watch CSS durante desarrollo
make web-watch

# Arrancar servidor
make run
```

Abrir `http://localhost:8100/login` (o el puerto configurado en `SERVER_PORT`).

## Tests

```bash
go test ./internal/auth/...
go test ./internal/shared/infrastructure/templates/...
```

Cobertura principal:

- `login_page_handler_test.go` — GET/POST login, credenciales inválidas y válidas.
- `renderer_test.go` — parse de plantillas.
- `web_auth_routes_test.go` — registro de rutas web.

## Menú de perfil de usuario

El dropdown del navbar (`partials/profile-menu.gohtml`) incluye:

- Cabecera con avatar y nombre del usuario autenticado.
- Enlaces: **Perfil** (`/perfil`), **Configuración** (`/configuracion`), **Cerrar sesión** (`POST /logout`).
- **Variantes de color** (7 temas vía `data-theme` en `<html>`, solo ícono circular).
- **Modo** claro/oscuro (`class="dark"`).

Las preferencias se persisten en `localStorage` (`docsy-theme-color`, `docsy-theme-mode`). No hay round-trip al servidor en esta fase.

Layout compartido: `internal/shared/transport/web/layout_data.go` (`AppLayoutData`).

Ver SDD: [`docs/sdd/009-web-profile-menu.md`](sdd/009-web-profile-menu.md).

## Componentes compartidos (Fase 2 — Iteración A)

Plantillas reutilizables para formularios, tablas y modales (usuarios y auditoría):

| Plantilla | Uso |
|-----------|-----|
| `components/confirmation-modal` | Confirmar acciones destructivas (HTMX en botón confirmar) |
| `components/success-modal` | Resumen post-acción con secciones label/valor |
| `components/modal-shell` | Modal genérico con tonos `warning`, `danger`, `success`, `info` |
| `partials/form-field` | Input flotante con estado inválido y toggle de contraseña |
| `partials/table-pagination` | Paginación offset/limit con URLs prev/next |
| `partials/alert-success` | Banner inline de éxito |

View models en `internal/shared/transport/web/components_data.go`.  
Mostrar modales: `openModal('id')` · Cerrar: `closeModal('id')` (`web/static/js/app.js`).

Ver SDD: [`docs/sdd/010-web-users-audit-modals.md`](sdd/010-web-users-audit-modals.md).

## Usuarios y auditoría (Fase 2 — Iteraciones B y C)

| Ruta | Permiso | Handler |
|------|---------|---------|
| `GET /usuarios` | `users.read` | `UsersPageHandler.ListUsers` |
| `GET /usuarios/nuevo`, `POST /usuarios` | `users.create` | create form + submit |
| `GET/POST /usuarios/:id/editar` | `users.update` | edit form + submit |
| `GET /auditoria` | `audit.read` | `AuditPageHandler.List` (partial HTMX en `#audit-table`) |

Dual transport: mismos use cases que `/api/v1/users` y `GET /api/v1/auditoria`.

## Archivo personal (SDD 011 — Iteraciones A–C)

| Ruta | Permiso | Handler |
|------|---------|---------|
| `GET /archivo` | `archive.read` | `ArchivePageHandler.ShowArchive` (lista workspaces) |
| `GET/POST /archivo/hogares/nuevo` | `archive.manage` | crear hogar compartido |
| `GET /archivo/hogares/:id/miembros` | `archive.read` | listado de miembros |
| `POST /archivo/hogares/:id/miembros` | `archive.manage` | invitar miembro por email |
| `POST /archivo/hogares/:id/miembros/:userId/eliminar` | `archive.manage` | eliminar miembro |
| `GET /archivo/documentos` | `archive.read` | listado con filtros `q`, `category`, `status`, `workspace_id` |
| `GET/POST /archivo/documentos/nuevo` | `archive.write` | crear documento con adjunto obligatorio (multipart `file`, `workspace_id` opcional) |
| `GET/POST /archivo/documentos/:id/editar` | `archive.write` | formulario editar metadatos (`workspace_id` en query/form) |
| `POST /archivo/documentos/:id/archivos` | `archive.write` | subir adjunto (PDF, imagen, Word, Excel; multipart `file`) |
| `GET /archivo/documentos/:id/archivos/:fileId` | `archive.read` | descargar adjunto |
| `POST /archivo/documentos/:id/archivos/:fileId/eliminar` | `archive.write` | eliminar adjunto |
| `GET /api/v1/archive/workspaces/me` | `archive.read` | workspace personal |
| `GET /api/v1/archive/workspaces` | `archive.read` | listar workspaces del usuario |
| `POST /api/v1/archive/workspaces/household` | `archive.manage` | crear hogar |
| `GET /api/v1/archive/workspaces/:id/members` | `archive.read` | listar miembros |
| `POST /api/v1/archive/workspaces/:id/members` | `archive.manage` | invitar miembro |
| `PATCH /api/v1/archive/workspaces/:id/members/:userId` | `archive.manage` | cambiar rol |
| `DELETE /api/v1/archive/workspaces/:id/members/:userId` | `archive.manage` | eliminar miembro |
| `GET /api/v1/archive/categories` | `archive.read` | categorías seed |
| `GET/POST /api/v1/archive/documents` | `archive.read` / `archive.write` | listar / crear (`workspace_id` opcional) |
| `GET/PATCH /api/v1/archive/documents/:id` | `archive.read` / `archive.write` | detalle / actualizar |
| `POST /api/v1/archive/documents/:id/archive` | `archive.write` | soft-archive |
| `GET/POST /api/v1/archive/documents/:id/files` | `archive.read` / `archive.write` | listar / subir adjuntos |
| `GET/DELETE /api/v1/archive/documents/:id/files/:fileId` | `archive.read` / `archive.write` | descargar / eliminar |

Storage local: `DOCUMENT_PATH` (default `./storage/documents`), límite por archivo `MAX_FILE_SIZE` (default 10 MB), cuota blanda del sidebar `STORAGE_QUOTA_BYTES` (default 10 GiB; aún no se aplica en upload). Tipos: PDF, JPG/JPEG, PNG, TIFF, WebP, GIF, DOC/DOCX, XLS/XLSX.

Bounded context: `internal/archive/`. Ver [`docs/sdd/011-archive-personal-family.md`](sdd/011-archive-personal-family.md).

Las escrituras del archivo (documentos, adjuntos, hogares y miembros) quedan en `audit_logs` y se listan en `GET /auditoria` con acciones `archive.*`.

## API REST

La API JSON existente (`POST /api/v1/auth/login`, etc.) **no se modifica**. Ambos transports comparten el mismo `LoginUseCase`.
