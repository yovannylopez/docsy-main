#!/bin/bash

# Script para migrar de WSL a herramientas básicas de linting
# Autor: Asistente AI
# Fecha: $(date)

echo "🚀 Migración a herramientas básicas de linting"
echo "=============================================="

# Verificar si las herramientas están instaladas
check_tool() {
    if command -v $1 > /dev/null; then
        echo "✅ $1 está instalado"
        return 0
    else
        echo "❌ $1 no está instalado"
        return 1
    fi
}

echo ""
echo "📋 Verificando herramientas instaladas:"
check_tool "goimports"
check_tool "golint"
check_tool "golangci-lint"

echo ""
echo "🔧 Instalando herramientas básicas si no están disponibles..."
make install-basic-tools

echo ""
echo "🧹 Ejecutando limpieza básica del código..."
make format

echo ""
echo "🔍 Ejecutando linting básico..."
make lint-basic

echo ""
echo "✅ Migración completada!"
echo ""
echo "📝 Comandos disponibles:"
echo "  make format        - Formatear código y organizar imports"
echo "  make lint-basic    - Linting básico (gofmt + goimports + golint)"
echo "  make verify-simple - Verificación sin WSL"
echo "  make verify        - Verificación completa (con WSL)"
echo ""
echo "💡 Recomendación: Usa 'make lint-basic' para desarrollo diario"
echo "   y 'make verify-simple' para verificaciones más completas." 