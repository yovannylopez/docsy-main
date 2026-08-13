# SDD 013 — Categorías planas personalizadas (archive)

## Objetivo

Permitir gestión de categorías **sin jerarquía**: catálogo de sistema (seed) + categorías personalizadas por workspace.

## Modelo

`archive_document_categories`:

| Campo | Notas |
|-------|--------|
| `code` PK | Sistema: slug estable (`taxes`, …). Custom: `c_` + UUID |
| `workspace_id` | `NULL` = sistema; set = custom del workspace |
| `label_es` | Nombre visible |
| `is_system` | `true` solo lectura vía API/UI |
| `is_active` | Soft-delete en custom |

Documento sigue referenciando `category_code`. Listado = sistema activos ∪ custom del workspace.

## Reglas

- Sin padres/hijos.
- Máx. 20 custom por workspace.
- Label único (case-insensitive) entre activas visibles en el workspace.
- Sistema: no renombrar/desactivar por usuario.
- Desactivar custom solo si no hay documentos con ese code.
- Escritura: miembro con rol distinto de `viewer` + `archive.write`.
- **Super admin** (`super_admin`): puede **renombrar** categorías base (cambio global de `label_es`). No desactiva el seed desde la UI.

## API / Web

- `GET/POST /api/v1/archive/categories`
- `PATCH/DELETE /api/v1/archive/categories/:code`
- UI: `/archivo/categorias`

## Seed (sistema)

| code | label_es |
|------|----------|
| `identity` | Identidad |
| `health` | Salud |
| `finance` | Finanzas |
| `taxes` | Impuestos |
| `property` | Propiedades |
| `insurance` | Seguros |
| `education` | Educación |
| `work` | Trabajo |
| `legal` | Legal |
| `utilities` | Servicios públicos |
| `invoices` | Facturas de compra |
| `photos` | Fotografías |
| `other` | Otros |

Código `photos` (no `photographs`/`pictures`): corto y habitual en apps; label en UI **Fotografías**. Sin categoría `family` — el titular se modela aparte (miembro/etiqueta).
