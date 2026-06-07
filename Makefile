.PHONY: build run clean test

# Variables
BINARY_NAME=docsy-main
GO_FILES=$(shell find . -name '*.go')
BUILD_DIR=bin

# Build the application
build:
	@echo "Building Core..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/

# Run the application
run:
	@go run ./cmd/

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running internal tests..."
	@go test -v -timeout 30s ./internal/...; \
	echo "Running pkg tests..."; \
	go test -v -timeout 30s ./pkg/... 2>&1 | grep -v "does not contain package" || true

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover -timeout 30s ./internal/... || true
	@go test -v -cover -timeout 30s ./pkg/... || true

# Generate HTML coverage report
coverage-html:
	@echo "Generating HTML coverage report..."
	@go test -v -coverprofile=coverage.out -timeout 30s ./internal/... || true
	@go test -v -coverprofile=coverage.out -timeout 30s ./pkg/... || true
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Open coverage.html in your browser to view the report"

# Watch coverage - Generate HTML report and open in browser
watch-coverage:
	@echo "Generating HTML coverage report and opening in browser..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@if command -v open > /dev/null; then \
		open coverage.html; \
		echo "Report opened in browser"; \
	elif command -v xdg-open > /dev/null; then \
		xdg-open coverage.html; \
		echo "Report opened in browser"; \
	else \
		echo "Please open coverage.html manually in your browser"; \
	fi

# Download dependencias
deps:
	@echo "Downloading dependencias..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Format code with gofumpt (strict formatting including line length)
fmt-strict:
	@echo "Formatting code with gofumpt..."
	@if command -v gofumpt > /dev/null; then \
		gofumpt -w .; \
		echo "✓ gofumpt completed"; \
	else \
		echo "gofumpt is not installed. Run: go install mvdan.cc/gofumpt@latest"; \
		echo "Falling back to gofmt..."; \
		go fmt ./...; \
	fi

# Verify code (lint, vet)
check-linter:
	@echo "Verifying code..."
	@go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format and organize imports
format:
	@echo "Formatting code and organizing imports..."
	@if command -v gofumpt > /dev/null; then \
		gofumpt -w .; \
		echo "✓ gofumpt completed (strict formatting)"; \
	else \
		echo "gofumpt not found, using gofmt..."; \
		go fmt ./...; \
	fi
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
		echo "✓ goimports completed"; \
	else \
		echo "goimports is not installed. Run: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Basic linting with gofmt + goimports + golint
lint-basic:
	@echo "Running basic linting..."
	@go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
		echo "✓ goimports completed"; \
	else \
		echo "goimports is not installed. Run: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi
	@if command -v golint > /dev/null; then \
		golint ./...; \
		echo "✓ golint completed"; \
	else \
		echo "golint is not installed. Run: go install golang.org/x/lint/golint@latest"; \
	fi

# Build and run
dev: build run

# Default target
all: clean build

# Generate mocks for auth module MFA interfaces
generate-auth-mfa-mocks:
	@echo "Generating auth MFA mocks..."
	@if command -v mockery > /dev/null; then \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=mfa_setup_service_mock.go --name=MFASetupService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=mfa_confirm_service_mock.go --name=MFAConfirmService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=mfa_verify_service_mock.go --name=MFAVerifyService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=mfa_disable_service_mock.go --name=MFADisableService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=totp_provider_mock.go --name=TOTPProvider; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=mfa_secret_encryptor_mock.go --name=MFASecretEncryptor; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=verification_token_repository_mock.go --name=VerificationTokenRepository; \
	else \
		echo "mockery is not installed. Run: go install github.com/vektra/mockery/v2@latest"; \
	fi

# Generate mocks for auth module
generate-auth-mocks:
	@echo "Generating auth module mocks..."
	@if command -v mockery > /dev/null; then \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=user_repository_mock.go --name=UserRepository; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=authentication_service_mock.go --name=AuthenticationService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=authorization_service_mock.go --name=AuthorizationService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=login_service_mock.go --name=LoginService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=signup_service_mock.go --name=SignupService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=change_password_service_mock.go --name=ChangePasswordService; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=password_hasher_mock.go --name=PasswordHasher; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=token_generator_mock.go --name=TokenGenerator; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=session_repository_mock.go --name=SessionRepository; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=audit_repository_mock.go --name=AuditRepository; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=list_audit_logs_usecase_mock.go --name=ListAuditLogsUseCase; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=system_config_repository_mock.go --name=SystemConfigRepository; \
		mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=password_history_repository_mock.go --name=PasswordHistoryRepository; \
	else \
		echo "mockery is not installed. Run: go install github.com/vektra/mockery/v2@latest"; \
	fi

# Generate mocks for users module
generate-users-mocks:
	@echo "Generating users module mocks..."
	@if command -v mockery > /dev/null; then \
		mockery --dir=internal/users/domain/ports --output=internal/users/mocks --outpkg=mocks --all; \
	else \
		echo "mockery is not installed. Run: go install github.com/vektra/mockery/v2@latest"; \
	fi

# Generate all mocks (base boilerplate: auth + users)
# Para agregar un nuevo módulo: make generate-<modulo>-mocks
generate-mocks: generate-auth-mocks generate-users-mocks

# Scaffold a new module (usage: make scaffold MODULE=products)
scaffold:
	@if [ -z "$(MODULE)" ]; then echo "Usage: make scaffold MODULE=<name>"; exit 1; fi
	@scripts/scaffold_module.sh $(MODULE)

# Install mockery
install-mockery:
	@echo "Installing mockery..."
	@go install github.com/vektra/mockery/v2@latest

# Install golangci-lint
install-lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install basic linting tools
install-basic-tools:
	@echo "Installing basic linting tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/lint/golint@latest

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  coverage-html - Generate HTML coverage report"
	@echo "  watch-coverage - Generate HTML coverage report and open in browser"
	@echo "  deps          - Download dependencias"
	@echo "  fmt           - Format code"
	@echo "  fmt-strict    - Format code with gofumpt (strict formatting)"
	@echo "  format        - Format code and organize imports (uses gofumpt if available)"
	@echo "  verify        - Run full lint (CI/production) — .golangci.yml"
	@echo "  verify-dev    - Run relaxed lint (daily dev) — .golangci-dev.yml"
	@echo "  lint-basic    - Run basic linting (gofmt + goimports + golint)"
	@echo "  dev           - Build and run the application"
	@echo "  all           - Clean and build"
	@echo "  generate-mocks - Generate all mocks for testing (auth + users)"
	@echo "  generate-auth-mocks - Generate auth module mocks"
	@echo "  generate-users-mocks - Generate users module mocks"
	@echo "  generate-auth-mfa-mocks - Generate auth MFA-related mocks"
	@echo "  install-mockery - Install mockery tool"
	@echo "  install-lint - Install golangci-lint tool"
	@echo "  install-basic-tools - Install basic linting tools"
	@echo "  help          - Show this help message"
