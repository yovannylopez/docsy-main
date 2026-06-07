#!/bin/bash

# Verifica configuración de base de datos: modo local (DB_*) vs DATABASE_URL

set -e

echo "============================================"
echo "Test 1: Verificar configuración LOCAL (DB_*)"
echo "============================================"
echo ""
echo "Variables individuales de ejemplo:"
echo "DB_HOST=localhost"
echo "DB_PORT=5432"
echo "DB_USER=postgres"
echo "DB_NAME=go_boilerplate"
echo "DB_SSLMODE=disable"
echo ""

if [ ! -f .env ]; then
    echo "❌ Error: Archivo .env no encontrado"
    echo "Copia un ejemplo a .env, por ejemplo:"
    echo "  cp .env.example .env"
    echo "  # o: cp env.example .env"
    exit 1
fi

if grep -q "^DATABASE_URL=" .env; then
    echo "⚠️  Advertencia: DATABASE_URL está configurada en .env"
    echo "Para probar solo el modo con variables DB_*, comenta o elimina DATABASE_URL."
    echo ""
fi

echo "✅ Revisión de modo local lista"
echo ""

echo "============================================"
echo "Test 2: Modo DATABASE_URL (producción / PaaS)"
echo "============================================"
echo ""
echo "Con DATABASE_URL definida, la app parsea host, puerto, usuario, contraseña y sslmode"
echo "desde la URL. Configura además JWT, logging y email según tu entorno."
echo ""

echo "============================================"
echo "Test 3: Compilar proyecto"
echo "============================================"
echo ""

if go build -o bin/docsy-main ./cmd/; then
    echo "✅ Compilación correcta"
    echo "Binario: bin/docsy-main"
else
    echo "❌ Error al compilar"
    exit 1
fi

echo ""
echo "============================================"
echo "Resumen"
echo "============================================"
echo ""
echo "• Sin DATABASE_URL: se usan DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE."
echo "• Con DATABASE_URL: tiene prioridad; sslmode por defecto en la URL suele ser require."
echo ""
