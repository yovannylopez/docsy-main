# SDD 010 — Usuarios, auditoría y modales web (Fase 2)

**Estado:** Iteración A completada · Iteraciones B y C pendientes  
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

### B — Usuarios (pendiente)

- `UsersPageHandler`: list, create, edit
- Rutas: `GET /usuarios`, `GET /usuarios/nuevo`, `GET /usuarios/:id/editar`, `POST /usuarios`, etc.
- Templates: `web/templates/users/`
- `WebRequirePermission` para `users.read` / `users.create` / `users.update`

### C — Auditoría (pendiente)

- `AuditPageHandler`: list con filtros HTMX
- Ruta: `GET /auditoria`
- Template: `web/templates/audit/list.gohtml`
- Permiso: `audit.read`

## Convenciones

- Design tokens Tailwind (`bg-card`, `text-foreground`, `border-border`) en lugar de `gray-*`.
- Assets vía `assetURL` en plantillas.
- Modales: mostrar con `openModal('modal-id')`, cerrar con `closeModal` o `hx-on::after-request`.
