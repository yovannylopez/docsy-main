# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

# Download dependencias
deps:
	@echo "Downloading dependencias..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Verify code (lint, vet)
verify:
	@echo "Verifying code..."
	@go vet ./...
	@if command -v golint > /dev/null; then \
		golint ./...; \
	else \
		echo "golint is not installed. Run: go install golang.org/x/lint/golint@latest"; \
	fi
