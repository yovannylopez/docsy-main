# SDD 009 — Menú de perfil de usuario (capa web HTMX)

| Campo | Valor |
|-------|-------|
| **Versión** | 1.0.0 |
| **Fecha** | Junio 2026 |
| **Estado** | Implementado (iteración 1) |
| **Bounded context** | Web (HTMX) + auth (sesión) + users (perfil, fase 2) |
| **Depende de** | `docs/WEB_FRONTEND.md`, login web funcional |

---

## 1. Propósito

Completar el dropdown de perfil en la barra superior del layout `app`, alineado al dashboard Angular y al export `htmx-export`.

## 2. Alcance implementado (iteración 1)

- Partial `partials/profile-menu.gohtml` reutilizable.
- Enlaces: Perfil, Configuración, Cerrar sesión (HTMX).
- Preferencias UI en cliente: 7 variantes de color, modo claro/oscuro, distribución LTR/RTL.
- `AppLayoutData` compartido en `internal/shared/transport/web/`.
- Rutas placeholder: `GET /perfil`, `GET /configuracion`.

## 3. Fuera de alcance (iteración 2)

- Formulario editable de perfil con datos de `GetUserProfile`.
- Persistencia de preferencias en BD.

## 4. Archivos clave

| Archivo | Rol |
|---------|-----|
| `web/templates/partials/profile-menu.gohtml` | Dropdown completo |
| `web/templates/partials/navbar.gohtml` | Incluye profile-menu |
| `web/static/js/app.js` | `setThemeColor`, `setThemeMode`, `setTextDirection` |
| `internal/shared/transport/web/layout_data.go` | DTO de layout |
| `internal/auth/transport/handlers/login_page_handler.go` | `buildAppLayoutData`, placeholders |

## 5. Criterios de aceptación

- [x] Dropdown con secciones del screenshot (español).
- [x] Nombre del usuario en cabecera del menú.
- [x] Logout HTMX + limpieza cookie/sessionStorage.
- [x] Temas y dirección persisten en `localStorage`.
- [x] Estado activo visible en toggles.
- [x] Tests de render y rutas.
