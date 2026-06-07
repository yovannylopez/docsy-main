# Makefile para el paquete de paginación

.PHONY: test test-coverage format lint clean install-tools

# Variables
GO_VERSION := 1.26
PACKAGE_NAME := github.com/yovannylopez/docsy-main/pkg/pagination

# Herramientas de desarrollo
GOLANGCI_LINT_VERSION := v1.55.2
MOCKERY_VERSION := v2.43.2

# Colores para output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Información del paquete
info:
	@echo "$(BLUE)=== Paquete de Paginación ===$(NC)"
	@echo "$(GREEN)Paquete:$(NC) $(PACKAGE_NAME)"
	@echo "$(GREEN)Go Version:$(NC) $(GO_VERSION)"
	@echo "$(GREEN)Directorio:$(NC) $(PWD)"
	@echo ""

# Instalar herramientas básicas
install-basic-tools:
	@echo "$(YELLOW)Instalando herramientas básicas...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@echo "$(GREEN)✓ golangci-lint instalado$(NC)"

# Instalar herramientas de testing
install-test-tools:
	@echo "$(YELLOW)Instalando herramientas de testing...$(NC)"
	@go install github.com/stretchr/testify@latest
	@echo "$(GREEN)✓ testify instalado$(NC)"

# Instalar todas las herramientas
install-tools: install-basic-tools install-test-tools
	@echo "$(GREEN)✓ Todas las herramientas instaladas$(NC)"

# Formatear código
format:
	@echo "$(YELLOW)Formateando código...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Código formateado$(NC)"

# Ejecutar tests
test:
	@echo "$(YELLOW)Ejecutando tests...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)✓ Tests completados$(NC)"

# Ejecutar tests con coverage
test-coverage:
	@echo "$(YELLOW)Ejecutando tests con coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Tests con coverage completados$(NC)"
	@echo "$(BLUE)Coverage report: coverage.html$(NC)"

# Ver coverage en terminal
coverage:
	@echo "$(YELLOW)Mostrando coverage...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Linting
lint:
	@echo "$(YELLOW)Ejecutando linting...$(NC)"
	@golangci-lint run
	@echo "$(GREEN)✓ Linting completado$(NC)"

# Linting con configuración de desarrollo (más permisivo)
lint-dev:
	@echo "$(YELLOW)Ejecutando linting (desarrollo)...$(NC)"
	@golangci-lint run --config .golangci-dev.yml
	@echo "$(GREEN)✓ Linting (desarrollo) completado$(NC)"

# Verificar dependencias
verify-deps:
	@echo "$(YELLOW)Verificando dependencias...$(NC)"
	@go mod verify
	@go mod tidy
	@echo "$(GREEN)✓ Dependencias verificadas$(NC)"

# Limpiar archivos generados
clean:
	@echo "$(YELLOW)Limpiando archivos generados...$(NC)"
	@rm -f coverage.out coverage.html
	@go clean
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

# Build del paquete
build:
	@echo "$(YELLOW)Construyendo paquete...$(NC)"
	@go build ./...
	@echo "$(GREEN)✓ Paquete construido$(NC)"

# Verificar que el paquete compila
check:
	@echo "$(YELLOW)Verificando compilación...$(NC)"
	@go build ./...
	@echo "$(GREEN)✓ Paquete compila correctamente$(NC)"

# Workflow completo de desarrollo
dev: format lint-dev test
	@echo "$(GREEN)✓ Workflow de desarrollo completado$(NC)"

# Workflow completo de producción
prod: format lint test-coverage build
	@echo "$(GREEN)✓ Workflow de producción completado$(NC)"

# Verificar todo
verify: format lint test build
	@echo "$(GREEN)✓ Verificación completa exitosa$(NC)"

# Mostrar ayuda
help:
	@echo "$(BLUE)=== Comandos Disponibles ===$(NC)"
	@echo "$(GREEN)info$(NC)              - Mostrar información del paquete"
	@echo "$(GREEN)install-tools$(NC)     - Instalar herramientas de desarrollo"
	@echo "$(GREEN)format$(NC)            - Formatear código"
	@echo "$(GREEN)test$(NC)              - Ejecutar tests"
	@echo "$(GREEN)test-coverage$(NC)     - Ejecutar tests con coverage"
	@echo "$(GREEN)coverage$(NC)          - Mostrar coverage en terminal"
	@echo "$(GREEN)lint$(NC)              - Ejecutar linting"
	@echo "$(GREEN)lint-dev$(NC)          - Ejecutar linting (desarrollo)"
	@echo "$(GREEN)verify-deps$(NC)       - Verificar dependencias"
	@echo "$(GREEN)clean$(NC)             - Limpiar archivos generados"
	@echo "$(GREEN)build$(NC)             - Construir paquete"
	@echo "$(GREEN)check$(NC)             - Verificar compilación"
	@echo "$(GREEN)dev$(NC)               - Workflow de desarrollo"
	@echo "$(GREEN)prod$(NC)              - Workflow de producción"
	@echo "$(GREEN)verify$(NC)            - Verificar todo"
	@echo "$(GREEN)help$(NC)              - Mostrar esta ayuda"

# Comando por defecto
.DEFAULT_GOAL := help
