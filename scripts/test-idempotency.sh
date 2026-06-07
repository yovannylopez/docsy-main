#!/bin/bash

# Script para probar la idempotencia de las migraciones
# Este script ejecuta las migraciones dos veces para verificar que no fallen

set -e

echo "=========================================="
echo "Test de Idempotencia de Migraciones"
echo "=========================================="
echo ""

# Colores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Cargar variables de entorno
if [ -f .env ]; then
    echo "Cargando variables de entorno desde .env..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}Error: Archivo .env no encontrado${NC}"
    exit 1
fi

# Construir URL de base de datos
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

echo "Configuración de Base de Datos:"
echo "  Host: ${DB_HOST}"
echo "  Port: ${DB_PORT}"
echo "  Database: ${DB_NAME}"
echo "  User: ${DB_USER}"
echo ""

# Verificar conexión a la base de datos
echo "Verificando conexión a la base de datos..."
if ! psql "${DB_URL}" -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}Error: No se puede conectar a la base de datos${NC}"
    echo "Por favor, asegúrate de que PostgreSQL esté corriendo y las credenciales sean correctas."
    exit 1
fi
echo -e "${GREEN}✓ Conexión exitosa${NC}"
echo ""

# Función para ejecutar migraciones
run_migrations() {
    local attempt=$1
    echo "=========================================="
    echo "Ejecución #${attempt} de Migraciones"
    echo "=========================================="
    
    if migrate -path migrations/core -database "${DB_URL}" up; then
        echo -e "${GREEN}✓ Migraciones ejecutadas exitosamente (intento #${attempt})${NC}"
        return 0
    else
        echo -e "${RED}✗ Error ejecutando migraciones (intento #${attempt})${NC}"
        return 1
    fi
}

# Verificar que migrate esté instalado
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: 'migrate' no está instalado${NC}"
    echo "Instálalo con: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Resetear la base de datos (opcional - comentar si no quieres resetear)
echo "¿Deseas resetear la base de datos antes de probar? (y/N)"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "Reseteando base de datos..."
    migrate -path migrations/core -database "${DB_URL}" drop -f
    echo -e "${GREEN}✓ Base de datos reseteada${NC}"
    echo ""
fi

# Primera ejecución de migraciones
echo ""
if ! run_migrations 1; then
    echo -e "${RED}Error en la primera ejecución de migraciones${NC}"
    exit 1
fi

echo ""
echo "Esperando 2 segundos antes de la segunda ejecución..."
sleep 2
echo ""

# Segunda ejecución de migraciones (prueba de idempotencia)
if ! run_migrations 2; then
    echo ""
    echo -e "${RED}=========================================="
    echo "✗ TEST FALLIDO"
    echo "==========================================${NC}"
    echo "Las migraciones NO son idempotentes."
    echo "La segunda ejecución falló."
    exit 1
fi

# Verificar estado de la base de datos
echo ""
echo "=========================================="
echo "Verificando Estado de la Base de Datos"
echo "=========================================="

# Contar tablas creadas
TABLE_COUNT=$(psql "${DB_URL}" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';")
echo "Tablas creadas: ${TABLE_COUNT}"

# Contar roles
ROLE_COUNT=$(psql "${DB_URL}" -t -c "SELECT COUNT(*) FROM roles;")
echo "Roles creados: ${ROLE_COUNT}"

# Contar permisos
PERMISSION_COUNT=$(psql "${DB_URL}" -t -c "SELECT COUNT(*) FROM permissions;")
echo "Permisos creados: ${PERMISSION_COUNT}"

# Contar usuarios
USER_COUNT=$(psql "${DB_URL}" -t -c "SELECT COUNT(*) FROM users;")
echo "Usuarios creados: ${USER_COUNT}"

# Contar dependencias
DEPENDENCIA_COUNT=$(psql "${DB_URL}" -t -c "SELECT COUNT(*) FROM dependencias;")
echo "Dependencias creadas: ${DEPENDENCIA_COUNT}"

# Contar configuraciones del sistema
CONFIG_COUNT=$(psql "${DB_URL}" -t -c "SELECT COUNT(*) FROM system_config;")
echo "Configuraciones del sistema: ${CONFIG_COUNT}"

echo ""
echo -e "${GREEN}=========================================="
echo "✓ TEST EXITOSO"
echo "==========================================${NC}"
echo "Las migraciones son idempotentes."
echo "Se pueden ejecutar múltiples veces sin errores."
echo ""
