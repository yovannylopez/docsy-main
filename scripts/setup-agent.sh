#!/usr/bin/env bash
# Verificación de entorno para desarrollo + requisitos habituales de MCP en Cursor.
# No imprime valores de tokens ni contenido sensible de mcp.json.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "== docsy-main: setup agent (dev + MCP) =="
echo ""

fail=0

ok() { echo "✓ $*"; }
warn() { echo "⚠ $*"; }
bad() { echo "✗ $*"; fail=1; }

# --- Go / Makefile / repo ---
if command -v go >/dev/null 2>&1; then
	ok "go: $(go version)"
else
	bad "go no encontrado"
fi

if command -v make >/dev/null 2>&1; then
	ok "make"
else
	bad "make no encontrado"
fi

if command -v psql >/dev/null 2>&1; then
	ok "psql disponible"
else
	warn "psql no encontrado (opcional si usas otro cliente SQL)"
fi

if [[ -f .env ]]; then
	ok ".env presente"
else
	warn ".env no encontrado — copia desde .env.example (README)"
fi

if make -n verify >/dev/null 2>&1; then
	ok "Makefile: target verify disponible"
else
	warn "no se pudo comprobar verify (revisar Makefile)"
fi

echo ""
echo "-- Runtime que suelen usar los MCP de Cursor --"
if command -v node >/dev/null 2>&1; then
	ok "node: $(node --version)"
else
	warn "node no encontrado — muchos servidores MCP se ejecutan con npx/node; instálalo si usas MCPs basados en JS"
fi

if command -v npx >/dev/null 2>&1; then
	ok "npx disponible ($(command -v npx))"
else
	warn "npx no encontrado — suele venir con Node.js (npm)"
fi

# --- Cursor MCP: rutas habituales (documentación Cursor / comunidad) ---
MCP_USER_CONFIG="${CURSOR_MCP_JSON:-${HOME}/.cursor/mcp.json}"
MCP_PROJECT_CONFIG="${ROOT}/.cursor/mcp.json"

echo ""
echo "-- Configuración MCP en Cursor (sin mostrar secretos) --"

check_mcp_file() {
	local f="$1"
	local label="$2"
	if [[ ! -f "$f" ]]; then
		return 1
	fi
	ok "${label}: ${f}"
	if command -v jq >/dev/null 2>&1; then
		local servers
		servers="$(jq -r '.mcpServers // {} | keys[]?' "$f" 2>/dev/null || true)"
		if [[ -n "${servers// }" ]]; then
			echo "  Servidores declarados: $(echo "$servers" | tr '\n' ' ')"
		else
			warn "  El archivo existe pero no hay claves en .mcpServers (o JSON inválido)"
		fi
		# Avisar si alguna variable de entorno del servidor está vacía (solo nombres)
		local empty_env
		empty_env="$(jq -r '
			.mcpServers // {} | to_entries[] |
			.key as $k |
			(.value.env // {}) | to_entries[] |
			select(.value == "" or .value == null) |
			$k + ": env vacío → " + .key
		' "$f" 2>/dev/null || true)"
		if [[ -n "${empty_env// }" ]]; then
			while IFS= read -r line; do
				[[ -z "$line" ]] && continue
				warn "  $line — revisa Cursor Settings → MCP o rellena el token"
			done <<<"$empty_env"
		fi
	else
		warn "  jq no instalado — instálalo para listar servidores MCP y comprobar env vacíos (brew install jq)"
	fi
	return 0
}

found_any=0
if check_mcp_file "$MCP_USER_CONFIG" "MCP usuario (~/.cursor/mcp.json)"; then
	found_any=1
fi
if [[ -f "$MCP_PROJECT_CONFIG" ]] && [[ "$MCP_PROJECT_CONFIG" != "$MCP_USER_CONFIG" ]]; then
	if check_mcp_file "$MCP_PROJECT_CONFIG" "MCP proyecto (.cursor/mcp.json)"; then
		found_any=1
	fi
fi

if [[ "$found_any" -eq 0 ]]; then
	warn "No se encontró mcp.json en ${MCP_USER_CONFIG} ni en ${MCP_PROJECT_CONFIG}"
	echo "  Añade servidores en Cursor: Settings → Cursor Settings → MCP (o crea el JSON manualmente)."
fi

echo ""
echo "-- Recordatorios para el equipo (tokens / integraciones) --"
echo "  • GitHub: si usáis un MCP de GitHub, el token (PAT fine-grained u OAuth) va en la config MCP de Cursor,"
echo "    no en el repo. Revisa que el servidor aparezca arriba y que las variables no estén vacías."
echo "  • ClickUp: el API token de ClickUp se configura en el servidor MCP de ClickUp (env en mcp.json o UI de Cursor)."
echo "  • Nunca commitees mcp.json con secretos; preferir variables de entorno del sistema o campos de Cursor."
echo ""
echo "Variable opcional: CURSOR_MCP_JSON=/ruta/custom/mcp.json ${0##*/}"

exit "$fail"
