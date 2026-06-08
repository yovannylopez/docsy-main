# SDD 010 — Usuarios, auditoría y modales web (Fase 2)

**Estado:** Iteraciones A, B y C completadas  
**Bounded context:** `shared` (componentes), `users`, `auth` (audit)  
**Referencia de diseño:** `htmx-export/html/` (no usado en runtime)

## Objetivo

Exponer listados y formularios de **usuarios** y **auditoría** como páginas server-rendered HTMX, reutilizando la API REST existente y los permisos RBAC (`users.read`, `users.create`, `users.update`, `audit.read`).

## Arquitectura (dual transport)

| Capa | Usuarios | Auditoría |
|------|----------|-----------|
| API JSON | `UsersHandler`, `users_routes.go` | `AuditHandler`, `audit_routes.go` |
| Web HTML | `UsersPageHandler`, `web_users_routes.go` | `AuditPageHandler`, `web_audit_routes.go` |

Middleware previsto: `WebRequirePermission` en rutas web protegidas por permiso.

## Iteraciones

### A — Componentes compartidos ✅

Plantillas portadas a `web/templates/`:

| Plantilla | Define | View model |
|-----------|--------|------------|
| `components/confirmation-modal.gohtml` | `components/confirmation-modal` | `ConfirmationModalData` |
| `components/success-modal.gohtml` | `components/success-modal` | `SuccessModalData` |
| `components/modal-shell.gohtml` | `components/modal-shell` | `ModalShellData` |
| `partials/form-field.gohtml` | `partials/form-field` | `FormFieldPageData` |
| `partials/table-pagination.gohtml` | `partials/table-pagination` | `PaginationData` |
| `partials/alert-success.gohtml` | `partials/alert-success` | `AlertSuccessData` |

View models en `internal/shared/transport/web/components_data.go`.  
Helpers JS existentes: `openModal`, `closeModal`, `togglePassword` (`web/static/js/app.js`).

Tests: `renderer_test.go`, `components_data_test.go`.

### B — Usuarios ✅

- `UsersPageHandler`: list, create, edit
- Rutas: `GET /usuarios`, `GET /usuarios/nuevo`, `POST /usuarios`, `GET/POST /usuarios/:id/editar`
- Templates: `web/templates/users/`, `partials/users-table`
- `WebAuthMiddleware.RequirePermission` para `users.read` / `users.create` / `users.update`

### C — Auditoría ✅

- `AuditPageHandler`: list con filtros HTMX (`hx-target="#audit-table"`)
- Ruta: `GET /auditoria`
- Templates: `web/templates/audit/list.gohtml`, `partials/audit-table`
- Permiso: `audit.read`

## Convenciones

- Design tokens Tailwind (`bg-card`, `text-foreground`, `border-border`) en lugar de `gray-*`.
- Assets vía `assetURL` en plantillas.
- Modales: mostrar con `openModal('modal-id')`, cerrar con `closeModal` o `hx-on::after-request`.
