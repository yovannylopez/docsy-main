# SDD 011 — Archivo personal / familiar (primer módulo de producto)

| Campo | Valor |
|-------|-------|
| **Versión** | 0.2.0 |
| **Fecha** | Agosto 2026 |
| **Estado** | **Aprobado** (decisiones congeladas) |
| **Bounded context** | `archive` (`internal/archive/`) |
| **Depende de** | `auth`, `users`, capa web HTMX (`docs/WEB_FRONTEND.md`), SDD 010 |
| **Visión de producto** | Docsy = GED configurable PaaS/SaaS; este SDD es el **primer módulo de negocio** |

---

## 1. Propósito

Diseñar e implementar el **gestor de archivo personal**: organizar documentos de una persona (luego hogar) — facturas, servicios públicos, certificados, salud, impuestos, pagos/extracurriculares, etc. — sobre el monolito existente, con Clean Architecture y dual transport (API JSON + HTMX).

La visión PaaS/SaaS (módulos activables por tenant / super admin) queda **fuera** de este SDD; se documenta como trabajo futuro (`platform`).

---

## 2. Decisiones congeladas (§9)

| # | Decisión | Elección |
|---|----------|----------|
| 1 | Corte del BC | **`archive` único** (workspace + membership + documents) |
| 2 | Binarios en MVP | **Solo metadatos** + puerto `DocumentStorage` (stub/noop); upload real = Iteración C |
| 3 | Hogar compartido | **Solo personal** en A–B; hogar = Iteración D |
| 4 | Catálogo de módulos PaaS | **SDD siguiente** (`platform`); archivo “siempre on” con permisos RBAC |

---

## 3. Tipología real (muestras del producto)

Referencias aportadas por el dueño del producto (fotos WhatsApp + PDFs en Downloads). Informan categorías y campos; **no** implican OCR en el MVP.

| Tipo real | Ejemplo | Categoría seed | Campos que el usuario querrá capturar |
|-----------|---------|----------------|----------------------------------------|
| Impuesto predial | Municipio Dosquebradas | `taxes` | Emisor, Nº factura/predio, fecha emisión, vencimiento, valor a pagar, referencia |
| Servicios públicos | SERVICIUDAD (acueducto/aseo), Efigas (gas) | `utilities` | Emisor, cuenta/contrato, periodo, vencimiento, total, referencia de pago |
| Factura electrónica de venta | TecnoTienda / DIAN (CUFE) | `invoices` | Emisor (NIT), Nº factura, fecha, total, CUFE (opcional en notas/ref) |
| Recibo de pago / mensualidad | Escuela de fútbol | `education` (o custom) | Emisor, concepto, fecha, valor, Nº recibo |
| Certificado de estudios | PDF académico | `education` | Emisor (institución), titular, fecha, tipo de certificado |
| Contrato laboral | Empresa | `work` | Emisor, titular, fecha, notas |
| Salud (previsto) | Citas, órdenes | `health` | Emisor, fecha, paciente, notas |
| Cédula / pasaporte | Documento de identidad | `identity` | Titular, Nº documento, vencimiento |
| Póliza / SOAT | Seguro | `insurance` | Emisor, póliza, vencimiento |

**Implicación de diseño:** el MVP captura **metadatos manuales** ricos (no solo título). El PDF/imagen se adjunta en Iteración C. OCR/extracción automática = SDD futuro.

---

## 4. Bounded context

**Nombre:** `archive`  
**Scaffold:** `make scaffold MODULE=archive`

Contiene:

- **Workspace** — contenedor del archivo (`personal` ahora; `household` / `organization` reservados)
- **WorkspaceMember** — membresía (owner ahora; member/viewer en D)
- **Document** + **DocumentCategory** — metadatos indexables
- **DocumentFile** — tabla lista en migración, **sin** upload en A–B
- Puerto **`DocumentStorage`** — stub hasta Iteración C

---

## 5. Alcance

### Incluye (Iteraciones A + B)

1. `EnsurePersonalWorkspace` al entrar a `/archivo` (1 workspace `personal` por usuario).
2. CRUD de documentos por metadatos dentro del workspace.
3. Categorías seed (ver §6).
4. Campos de documento (ver §6) — incluyen monto y vencimiento porque aparecen en casi todas las muestras.
5. Multi-tenant por fila: `workspace_id` + membership.
6. Dual transport API + HTMX.
7. Permisos: `archive.read`, `archive.write`, `archive.manage`.
8. Auditoría de create/update/archive vía audit existente (`auth.AuditRepository`; acciones `archive.*`; fallo de audit no bloquea la operación).
9. Puerto `DocumentStorage` + implementación local en disco (`DOCUMENT_PATH`).

### Excluye

- OCR / DIAN / CUFE automático.
- Recordatorios de vencimiento por email/push (futuro). MVP in-app: badges «Vence pronto» / «Vencido», franja resumen y filtro `?due=upcoming|expired` (ventana 7 días).
- Hogar e invitaciones (Iteración D — **hecho**).
- Panel super admin / catálogo de módulos (SDD `platform`).
- TRD/AGN, planes SaaS, cuotas de almacenamiento en nube (S3/minio = futuro).

---

## 6. Modelo de dominio

```text
Workspace
  id, name, type (personal | household | organization), owner_user_id,
  is_active, created_at, updated_at

WorkspaceMember
  workspace_id, user_id, role (owner | member | viewer), joined_at
  UNIQUE (workspace_id, user_id)

DocumentCategory  (seed)
  code PK, label_es, sort_order, is_active

Document
  id, workspace_id, category_code FK,
  title,                    -- obligatorio
  document_date,            -- fecha del documento / emisión
  due_date?,                -- vencimiento / “págese hasta” (nullable)
  issuer?,                  -- entidad emisora (SERVICIUDAD, Municipio, TecnoTienda…)
  reference_number?,        -- factura Nº, cuenta, contrato, recibo Nº, predio…
  amount_cents?,            -- valor principal en centavos COP (nullable)
  currency,                 -- default 'COP'
  notes?,                   -- texto libre (CUFE, conceptos, etc.)
  status (active | archived),
  created_by, updated_by, created_at, updated_at

DocumentFile  (schema en A; uso en C)
  id, document_id, storage_key, original_name, content_type, size_bytes,
  uploaded_by, uploaded_at
```

### Categorías seed

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

Las categorías son **planas** (sin jerarquía). Además del seed de sistema, cada workspace puede crear categorías personalizadas (ver SDD 013).

### Reglas

- Un usuario → un workspace `personal` (owner) en el MVP.
- Solo miembros del workspace operan documentos (más permiso RBAC del módulo).
- Soft-archive (`status=archived`) preferible a hard-delete.
- `amount_cents` evita float; UI muestra pesos.

---

## 7. Capas afectadas

| Capa | Qué |
|------|-----|
| `internal/archive/domain/` | Entities, DTOs, ports |
| `internal/archive/usecases/` | EnsurePersonalWorkspace, List/Create/Update/Archive Document |
| `internal/archive/infrastructure/` | sqlx, storage stub, container, openapi |
| `internal/archive/transport/` | JSON + page handlers HTMX + routes |
| `migrations/core/` | `000008_create_archive_module.up/down.sql` |
| `cmd/composition/` | DI + router + OpenAPI |
| `.golangci.yml` | depguard para `archive` |
| `web/templates/archive/` | home, list, form |
| Seed permisos | `archive.read/write/manage` en migración roles/permissions |

**Skills:** `go-conventions`, `go-api-rest`, `database-queries`, `go-validation`, `go-errors`, `go-context`, `go-testing`, `go-tooling`, `go-logging`.

---

## 8. Contratos HTTP

### API JSON (`/api/v1/archive`)

| Método | Ruta | Permiso |
|--------|------|---------|
| GET | `/workspaces/me` | `archive.read` |
| GET | `/workspaces` | `archive.read` |
| POST | `/workspaces/household` | `archive.manage` |
| GET | `/workspaces/:id/members` | `archive.read` |
| POST | `/workspaces/:id/members` | `archive.manage` |
| PATCH | `/workspaces/:id/members/:userId` | `archive.manage` |
| DELETE | `/workspaces/:id/members/:userId` | `archive.manage` |
| GET | `/categories` | `archive.read` |
| GET | `/documents` | `archive.read` |
| POST | `/documents` | `archive.write` |
| GET | `/documents/:id` | `archive.read` |
| PATCH | `/documents/:id` | `archive.write` |
| POST | `/documents/:id/archive` | `archive.write` |
| GET | `/documents/:id/files` | `archive.read` |
| POST | `/documents/:id/files` | `archive.write` |
| GET | `/documents/:id/files/:fileId` | `archive.read` |
| DELETE | `/documents/:id/files/:fileId` | `archive.write` |

Filtros listado: `workspace_id`, `category`, `q`, `from`, `to`, `due_before`, `status`, `limit`, `offset`.

### Web HTMX

| Método | Ruta | Permiso |
|--------|------|---------|
| GET | `/archivo` | `archive.read` |
| GET/POST | `/archivo/hogares/nuevo` | `archive.manage` |
| GET | `/archivo/hogares/:id/miembros` | `archive.read` |
| POST | `/archivo/hogares/:id/miembros` | `archive.manage` |
| POST | `/archivo/hogares/:id/miembros/:userId/eliminar` | `archive.manage` |
| GET | `/archivo/documentos` | `archive.read` |
| GET/POST | `/archivo/documentos/nuevo` | `archive.write` |
| GET/POST | `/archivo/documentos/:id/editar` | `archive.write` |
| POST | `/archivo/documentos/:id/archivos` | `archive.write` |
| GET | `/archivo/documentos/:id/archivos/:fileId` | `archive.read` |
| POST | `/archivo/documentos/:id/archivos/:fileId/eliminar` | `archive.write` |

---

## 9. Iteraciones

| Iteración | Entrega | Estado |
|-----------|---------|--------|
| **A — Cimiento** | Scaffold, migración, seed categorías + permisos, EnsurePersonalWorkspace, `GET /archivo` | **Hecho** |
| **B — Documentos** | CRUD metadatos (campos §6) API + HTMX list/form | **Hecho** |
| **C — Binarios** | Upload PDF/imagen vía `DocumentStorage` (disco local primero) | **Hecho** |
| **D — Hogar** | `household` + invitar miembros | **Hecho** |
| **E — Platform** | Otro SDD: catálogo de módulos + activación por tenant | Futuro |

---

## 10. Criterios de aceptación (A + B + C + D + auditoría)

- [x] `make scaffold MODULE=archive` + composition + depguard + permisos seed.
- [x] Usuario autenticado obtiene workspace personal en `/archivo`.
- [x] Crear/listar documentos con categoría, emisor, fechas, referencia y monto.
- [x] Aislamiento por `workspace_id` (sin membership → 403/404).
- [x] OpenAPI del módulo; tests use case + handler + render HTMX básico.
- [x] Linter y tests del módulo en verde.
- [x] Upload PDF/imagen vía `DocumentStorage` local (`DOCUMENT_PATH` / `MAX_FILE_SIZE`).
- [x] Hogar: crear workspace `household`, invitar/quitar miembros y cambiar rol.
- [x] Auditoría de escrituras: documento (create/update/archive), adjuntos (upload/delete), hogar y miembros; visibles en `/auditoria` con labels ES.

---

## 11. Riesgos

| Riesgo | Mitigación |
|--------|------------|
| Modelar OCR demasiado pronto | Metadatos manuales; OCR = SDD futuro |
| Perder campos útiles de facturas reales | `due_date`, `amount_cents`, `reference_number`, `issuer` desde B |
| Mezclar platform con archive | SDD E separado |
| Float en dinero | `amount_cents` int64 |

---

## 12. Próximo paso

**Iteración E — Platform** (otro SDD): catálogo de módulos + activación por tenant.

**Auditoría:** las escrituras del módulo emiten acciones namespaced (`archive.document.*`, `archive.file.*`, `archive.household.created`, `archive.member.*`) vía `auth.AuditRepository`. Se consultan en `GET /auditoria` (filtros y etiquetas en español).

Iteración D entregada:
- Workspace `household` con owner automático
- Invitar miembros existentes por email (roles `member` / `viewer`)
- API + HTMX: listar workspaces, crear hogar, gestionar miembros
- Documentos y adjuntos filtrables por `workspace_id`

Iteración C entregada:
- `LocalDocumentStorage` en `DOCUMENT_PATH` (límite `MAX_FILE_SIZE`)
- Adjuntos PDF/imagen: API + HTMX en edición de documento
- Tabla `archive_document_files` en uso; máx. 10 adjuntos/documento

Iteración B entregada:
- CRUD metadatos documentos (API + HTMX list/form)
- Campos: categoría, emisor, referencia, fechas, monto (`amount_cents`), notas
- Soft-archive vía API `POST .../documents/:id/archive`
- OpenAPI tag `archive`

Iteración A entregada:
- `internal/archive/` + migración `000008`
- `GET /archivo` y `GET /api/v1/archive/workspaces/me`
- Seed categorías y permisos `archive.*`
