#!/usr/bin/env bash
# scaffold_module.sh — Genera un nuevo módulo vertical a partir de la plantilla.
#
# Uso:
#   scripts/scaffold_module.sh <nombre_modulo>
#
# Ejemplo:
#   scripts/scaffold_module.sh products
#   scripts/scaffold_module.sh sales_orders
#
# El script:
#   1. Copia _template/TEMPLATE_MODULE → internal/<nombre_modulo>
#   2. Reemplaza todos los identificadores de plantilla por los del nuevo módulo
#   3. Renombra archivos .go
#   4. Genera .mockery.yaml del módulo
#   5. Imprime el checklist de pasos manuales restantes

set -euo pipefail

# ─── Helpers ──────────────────────────────────────────────────────────────────

info()    { printf '[scaffold] %s\n' "$*"; }
success() { printf '[scaffold] ✓ %s\n' "$*"; }
warn()    { printf '[scaffold] ⚠ %s\n' "$*"; }
error()   { printf '[scaffold] ✗ %s\n' "$*" >&2; exit 1; }

# ─── Validaciones ─────────────────────────────────────────────────────────────

if [ $# -ne 1 ]; then
  error "Uso: $0 <nombre_modulo>   (ej: $0 products)"
fi

MODULE_SNAKE="$1"

if ! echo "$MODULE_SNAKE" | grep -Eq '^[a-z][a-z0-9_]*$'; then
  error "El nombre debe ser snake_case: letras minúsculas, números y guión bajo. Ej: sales_orders"
fi

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TEMPLATE_DIR="${REPO_ROOT}/_template/TEMPLATE_MODULE"
TARGET_DIR="${REPO_ROOT}/internal/${MODULE_SNAKE}"

if [ ! -d "$TEMPLATE_DIR" ]; then
  error "No se encontró la plantilla en: ${TEMPLATE_DIR}"
fi

if [ -d "$TARGET_DIR" ]; then
  error "El módulo ya existe en: ${TARGET_DIR}"
fi

# ─── Convertir nombre a PascalCase ────────────────────────────────────────────
# sales_orders → SalesOrders  (compatible macOS BSD awk + GNU awk)

MODULE_PASCAL=$(echo "$MODULE_SNAKE" | awk -F_ '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) substr($i,2)} 1' OFS="")

info "Creando módulo '${MODULE_PASCAL}' en internal/${MODULE_SNAKE}/"
info "  snake_case : ${MODULE_SNAKE}"
info "  PascalCase : ${MODULE_PASCAL}"

# ─── Copiar plantilla ──────────────────────────────────────────────────────────

cp -r "$TEMPLATE_DIR" "$TARGET_DIR"

# ─── Reemplazar contenido en archivos .go y .yaml ─────────────────────────────
# Orden importa: primero el path completo del módulo, luego PascalCase, luego snake

find "$TARGET_DIR" -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.md" \) | while IFS= read -r file; do
  # GNU sed (Linux) and BSD sed (macOS)
  if sed --version >/dev/null 2>&1; then
    sed -i "s|TEMPLATE_MODULE|${MODULE_SNAKE}|g" "$file"
    sed -i "s|Template|${MODULE_PASCAL}|g" "$file"
    sed -i "s|template|${MODULE_SNAKE}|g" "$file"
  else
    sed -i '' "s|TEMPLATE_MODULE|${MODULE_SNAKE}|g" "$file"
    sed -i '' "s|Template|${MODULE_PASCAL}|g" "$file"
    sed -i '' "s|template|${MODULE_SNAKE}|g" "$file"
  fi
done

# ─── Renombrar archivos .go (template_* → <module>_*) ─────────────────────────

find "$TARGET_DIR" -type f -name "template_*.go" | while IFS= read -r file; do
  dir="$(dirname "$file")"
  base="$(basename "$file")"
  newbase="${base#template_}"
  mv "$file" "${dir}/${MODULE_SNAKE}_${newbase}"
done

# ─── Generar .mockery.yaml en la raíz del repo ────────────────────────────────
# Solo si no existe uno ya (no sobreescribir configuración existente)

MOCKERY_YAML="${REPO_ROOT}/.mockery.yaml"

if [ ! -f "$MOCKERY_YAML" ]; then
  cat > "$MOCKERY_YAML" <<YAML
with-expecter: true
packages:
  github.com/yovannylopez/docsy-main/internal/${MODULE_SNAKE}/domain/ports:
    config:
      dir: "internal/${MODULE_SNAKE}/mocks"
      outpkg: "mocks"
      filename: "{{.InterfaceName | snakecase}}_mock.go"
YAML
  info "Creado .mockery.yaml raíz (para mockery v2+)"
else
  warn ".mockery.yaml ya existe. Agrega manualmente la entrada del módulo '${MODULE_SNAKE}'."
fi

# ─── Resumen ──────────────────────────────────────────────────────────────────

echo ""
success "Módulo '${MODULE_PASCAL}' creado en internal/${MODULE_SNAKE}/"
echo ""
printf 'Archivos generados:\n'
find "$TARGET_DIR" -type f | sort | sed "s|${REPO_ROOT}/||" | while IFS= read -r f; do
  printf '  %s\n' "$f"
done

echo ""
printf 'Pasos manuales restantes:\n'
echo ""
echo "  1. Ajustar entidad:"
echo "       internal/${MODULE_SNAKE}/domain/entities/${MODULE_SNAKE}_entity.go"
echo ""
echo "  2. Crear migración SQL (usa el siguiente número disponible):"
echo "       migrations/core/XXXXXX_create_${MODULE_SNAKE}.up.sql"
echo "       migrations/core/XXXXXX_create_${MODULE_SNAKE}.down.sql"
echo ""
echo "  3. Registrar en cmd/composition/container.go:"
echo "       ${MODULE_PASCAL}  *${MODULE_SNAKE}container.${MODULE_PASCAL}Container"
echo "       // en NewContainer():"
echo "       ${MODULE_SNAKE}Container, err := ${MODULE_SNAKE}container.New${MODULE_PASCAL}Container(db)"
echo ""
echo "  4. Registrar rutas en cmd/composition/router.go:"
echo "       ${MODULE_SNAKE}routes.New${MODULE_PASCAL}Routes(r.e, r.container.${MODULE_PASCAL}).Register(jwtMiddleware)"
echo ""
echo "  5. Agregar tag OpenAPI en cmd/composition/bootstrap.go:"
echo "       openapiGen.AddTag(\"${MODULE_SNAKE}\", \"${MODULE_PASCAL} operations\")"
echo ""
echo "  6. Agregar target en Makefile:"
printf "       generate-%s-mocks:\n" "${MODULE_SNAKE}"
printf "           mockery --dir=internal/%s/domain/ports \\\\\n" "${MODULE_SNAKE}"
printf "                   --output=internal/%s/mocks \\\\\n" "${MODULE_SNAKE}"
printf "                   --outpkg=mocks --all\n"
echo ""
echo "  7. Generar mocks y verificar:"
echo "       make generate-${MODULE_SNAKE}-mocks"
echo "       make format && make verify && make test"
echo ""
success "¡Listo! Implementa la lógica en usecases/ y repositories/."
