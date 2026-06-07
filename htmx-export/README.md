# htmx-export — UI reutilizable para Go + HTMX

Exportación autocontenida del template visual **docsy-dashboard** (Angular) para construir el frontend de **docsy-main** con `html/template` + HTMX.

## Propósito

Esta carpeta contiene HTML plano, CSS Tailwind compilado, JavaScript vanilla y assets estáticos extraídos del dashboard Angular. **No reemplaza** la app Angular; es una referencia visual y punto de partida para el backend Go.

## Cómo copiar a docsy-main

```bash
# Desde la raíz de docsy-dashboard
cp -r htmx-export/ /ruta/a/docsy-main/web/

# Estructura sugerida en docsy-main
web/
├── static/          # ← css/, js/, assets/ de htmx-export
├── templates/       # ← html/ de htmx-export
└── ...
```

En Go, servir estáticos:

```go
router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
```

Parsear templates:

```go
tmpl, err := template.ParseGlob("web/templates/**/*.html")
// Ejecutar: tmpl.ExecuteTemplate(w, "login", data)
```

### Convención de composición (html/template)

Los layouts usan **parciales encadenados** (sin bloques globales conflictivos):

```
base-head → [auth-left-panel + auth-right-start + contenido + auth-right-end] → base-foot
base-head → [app-shell-start + contenido + app-shell-end] → base-foot
```

Cada página define un template de nivel superior único (`login`, `users-list`, etc.).

### FuncMap recomendado en Go

```go
funcMap := template.FuncMap{
    "hasPrefix": strings.HasPrefix,
}
```

## Mapeo Angular → exportado

| Origen Angular | Exportado |
|----------------|-----------|
| `src/app/modules/auth/auth.component.html` | `html/layouts/auth.html` (parciales `auth-left-panel`, `auth-right-start/end`) |
| `src/app/modules/auth/pages/sign-in/sign-in.component.html` | `html/pages/auth/login.html` |
| `src/app/modules/layout/layout.component.html` | `html/layouts/app.html` (`app-shell-start/end`) |
| `src/app/modules/layout/components/sidebar/*.html` | `html/partials/sidebar.html` |
| `src/app/modules/layout/components/navbar/navbar.component.html` + search + profile | `html/partials/navbar.html` |
| `src/app/modules/layout/components/footer/footer.component.html` | `html/partials/footer.html` |
| `src/app/modules/users/pages/users-list/users-list.component.html` | `html/pages/users/list.html` |
| `src/app/modules/users/pages/create-user/create-user.component.html` | `html/pages/users/create.html` |
| `src/app/modules/users/pages/edit-user/edit-user.component.html` | `html/pages/users/edit.html` |
| `src/app/modules/audit/pages/list-audit-logs/list-audit-logs.component.html` | `html/pages/audit/list.html` |
| `src/app/shared/components/confirmation-modal/` | `html/components/confirmation-modal.html` |
| `src/app/shared/components/success-modal/` | `html/components/success-modal.html` |
| `src/app/shared/components/button/button.component.ts` | Clases inline en botones (ver sección Botones) |
| `src/app/shared/components/secondary-header/` | `html/partials/secondary-header.html` |
| `src/app/modules/error/pages/error404/` | `html/pages/errors/404.html` |
| `src/app/modules/error/pages/error500/` | `html/pages/errors/500.html` |
| `src/styles.css` + `src/styles/headers.css` | `css/source/app.css` + `headers.css` + `components.css` |
| `src/assets/icons/` | `assets/icons/` |
| `src/assets/images/auth-screens.png` | `assets/images/auth-screens.png` *(ver nota abajo)* |
| — | `js/app.js`, `js/htmx.min.js` (v2.0.4) |

## Mapeo rutas Angular → Go

| Ruta Angular | Ruta web Go propuesta |
|--------------|----------------------|
| `/auth/sign-in` | `/login` |
| `/auth/forgot-password` | `/recuperar-contrasena` |
| `/dashboard/sirco/estadisticas/statistics` | `/estadisticas` |
| `/dashboard/sirco/comunicaciones/crear` | `/comunicaciones/crear` |
| `/dashboard/sirco/comunicaciones/listar` | `/comunicaciones` |
| `/dashboard/sirco/comunicaciones/asignaciones` | `/comunicaciones/asignaciones` |
| `/dashboard/sirco/comunicaciones/reportes` | `/comunicaciones/reportes` |
| `/dashboard/sirco/notificaciones` | `/notificaciones` |
| `/dashboard/sirco/auditoria/listar` | `/auditoria` |
| `/dashboard/sirco/configuracion/tipos-comunicaciones` | `/configuracion/tipos-comunicaciones` |
| `/dashboard/sirco/configuracion/estados-comunicaciones` | `/configuracion/estados-comunicaciones` |
| `/dashboard/administracion/usuarios` | `/usuarios` |
| `/dashboard/administracion/dependencias/create` | `/dependencias/crear` |
| `/dashboard/administracion/dependencias/list` | `/dependencias` |
| `/dashboard/administracion/empresas` | `/empresas` |
| `/dashboard/clasificacion-documental/series` | `/series-documentales` |
| `/dashboard/users/editar/:id` | `/usuarios/{{.ID}}/editar` |

## Placeholders Go template por página

### Layout común (app)

| Campo | Tipo | Uso |
|-------|------|-----|
| `.Title` | string | `<title>` |
| `.ThemeClass` | string | Clase en `<html>` (ej. `dark`) |
| `.ActiveRoute` | string | Resaltar ítem activo en sidebar |
| `.SidebarExpanded` | bool | Estado inicial sidebar |
| `.AppVersion` | string | Badge versión en sidebar |
| `.UnreadNotifications` | int | Badge notificaciones |
| `.UserName` | string | Navbar profile menu |
| `.Year` | int | Footer copyright |

### Login (`login`)

| Campo | Tipo |
|-------|------|
| `.Email` | string |
| `.RememberMe` | bool |
| `.Loading` | bool |
| `.Error` | string |
| `.EmailError` | string |
| `.PasswordError` | string |

### Usuarios list (`users-list`)

| Campo | Tipo |
|-------|------|
| `.Users` | `[]User` (ID, PrimerNombre, SegundoNombre, NombreUsuario, Email, Telefono, EstaActivo, EstaVerificado) |
| `.Total` | int |
| `.Loading`, `.Error` | bool/string |
| `.Pagination` | Offset, Limit, Total, PageStart, PageEnd, HasPrevious, HasNext, PrevURL, NextURL |

### Usuarios create (`users-create`)

| Campo | Tipo |
|-------|------|
| `.Form.*` | campos del formulario + errores |
| `.Roles` | `[]string` |
| `.TiposIdentificacion` | `[]{Value, Label}` |
| `.Success`, `.Error` | alertas |

### Usuarios edit (`users-edit`)

| Campo | Tipo |
|-------|------|
| `.User.*` | Email, FirstName, LastName, IdentificationType, etc. |
| `.LoadingData` | bool |

### Auditoría (`audit-list`)

| Campo | Tipo |
|-------|------|
| `.Filters.*` | UserID, Action, Resource, Result |
| `.AuditLogs` | filas pre-formateadas (CreatedAtFormatted, ActionLabel, ResultBadgeClass, etc.) |
| `.ActionOptions`, `.ResourceOptions`, `.ResultOptions` | selects |

## Botones (equivalente `app-button`)

Clases extraídas de `button.component.ts`:

| Variante | Clases Tailwind |
|----------|-----------------|
| Primary bold medium rounded full | `font-semibold focus-visible:outline-none flex items-center justify-center focus-visible:ring-2 focus-visible:ring-offset-2 active:translate-y-px disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 focus-visible:ring-primary px-5 py-2 text-sm rounded-lg w-full` |
| Primary none small rounded (link) | `... bg-transparent text-primary hover:bg-primary/10 focus-visible:ring-primary px-3 py-1 text-xs rounded-lg` |
| Secundario (formularios legacy) | `.btn.btn-secondary` (ver `components.css`) |

## Compilar CSS

### Ya compilado

`css/app.css` fue generado con Tailwind v4.3.0 desde este repo.

### Recompilar en docsy-main

```bash
cd web/static/css/source

# Requiere Node.js y dependencias Tailwind v4
npm install tailwindcss@^4 @tailwindcss/cli @tailwindcss/forms @tailwindcss/typography @tailwindcss/aspect-ratio tailwind-scrollbar

npx @tailwindcss/cli -i app.css -o ../app.css --minify
```

Plugins declarados en `css/source/app.css`:

- `@tailwindcss/forms`
- `@tailwindcss/typography`
- `@tailwindcss/aspect-ratio`
- `tailwind-scrollbar`

Variables de tema en `@theme` y `:root` / `.dark` / `data-theme`: `primary`, `background`, `foreground`, `muted`, `destructive`, `card`.

## JavaScript (`js/app.js`)

| Función | Descripción |
|---------|-------------|
| `togglePassword(inputId, buttonEl)` | Show/hide password |
| `toggleSidebar()` | Colapsar sidebar (localStorage `docsy-sidebar-expanded`) |
| `toggleProfileMenu()` | Dropdown perfil |
| `setThemeMode('light'\|'dark')` | Dark mode |
| `openModal(id)` / `closeModal(id)` | Modales |
| HTMX `htmx:configRequest` | Inyecta `Authorization: Bearer` desde `sessionStorage.auth_token` *(temporal)* |

## HTMX — comentarios en HTML

Los formularios incluyen comentarios `<!-- HTMX: hx-post="..." -->` donde aplica interacción parcial. Ejemplo login:

```html
<!-- HTMX: hx-post="/login" hx-target="#login-feedback" hx-swap="innerHTML" -->
```

## Incluido vs excluido

### Incluido (MVP + referencia Fase 2)

- Layout auth (dos columnas) + login
- Layout app (sidebar, navbar, footer)
- Estilos Tailwind v4 + headers + componentes custom
- Icons heroicons, logo, ilustraciones 404/500
- HTMX 2.0.4 vendored
- Usuarios: list, create, edit
- Auditoría: list con filtros (referencia)
- Modales: confirmation, success, modal-shell
- Paginación reutilizable

### Excluido (ruido del template)

| Módulo | Motivo |
|--------|--------|
| `uikit/` completo | Demo de componentes, no MVP |
| Auth: sign-up, two-steps, forgot-password, new-password | Fuera de MVP *(existen en Angular como referencia opcional)* |
| Dashboard demo: `management-auctions-*`, `management-chart-card` | Data fake |
| `context-debug`, `responsive-helper` | Herramientas dev |
| `apexchart.css` | Gráficos no incluidos en MVP |
| `app-session-notification` | Lógica Angular específica |

## Assets pendientes

- **`auth-screens.png`**: referenciado en Angular pero **no existe en el repo**. Ver `assets/images/README-auth-screens.md`.
- Avatar: se usa `avatar-default.svg` como placeholder (Angular usa `avt-02.png`).

## Páginas auth opcionales (solo en Angular)

Si necesitas convertir más pantallas auth, los originales están en:

- `src/app/modules/auth/pages/sign-up/`
- `src/app/modules/auth/pages/forgot-password/`
- `src/app/modules/auth/pages/new-password/`
- `src/app/modules/auth/pages/two-steps/`

## Verificación

```bash
# Sin sintaxis Angular en HTML exportado
grep -rE '\*ngIf|\*ngFor|routerLink|formControlName|\[ngClass\]|\(click\)|svg-icon|@if' htmx-export/html && echo "FAIL" || echo "OK"

# Estructura
find htmx-export -type f | head -50
```
