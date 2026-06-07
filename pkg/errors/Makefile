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

# Generate mocks with mockery
gen-mocks:
	@echo "Generating mocks with mockery..."
	@if ! command -v mockery > /dev/null; then \
		echo "mockery is not installed. Installing mockery..."; \
		go install github.com/vektra/mockery/v2@latest; \
	fi
	@mockery --all --output=./mocks --with-expecter

# Generate mocks for specific interface
gen-mock:
	@if [ -z "$(INTERFACE)" ]; then \
		echo "Usage: make gen-mock INTERFACE=InterfaceName"; \
		echo "Example: make gen-mock INTERFACE=AppConfig"; \
		exit 1; \
	fi
	@echo "Generating mock for interface: $(INTERFACE)"
	@if ! command -v mockery > /dev/null; then \
		echo "mockery is not installed. Installing mockery..."; \
		go install github.com/vektra/mockery/v2@latest; \
	fi
	@mockery --name=$(INTERFACE) --output=./mocks --with-expecter

# Clean mocks
clean-mocks:
	@echo "Cleaning mocks directory..."
	@rm -rf ./mocks/*

# Generate coverage report and open in browser
watch-coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./... -coverpkg=./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Opening coverage report in browser..."
	@if command -v xdg-open > /dev/null; then \
		xdg-open coverage.html; \
	elif command -v open > /dev/null; then \
		open coverage.html; \
	elif command -v start > /dev/null; then \
		start coverage.html; \
	else \
		echo "Please open coverage.html in your browser"; \
	fi
	@echo "Coverage report generated: coverage.html"

# Clean coverage files
clean-coverage:
	@echo "Cleaning coverage files..."
	@rm -f coverage.out coverage.html

# Watch for changes and regenerate coverage
watch-coverage-live:
	@echo "Watching for changes and regenerating coverage..."
	@echo "Press Ctrl+C to stop watching"
	@if command -v fswatch > /dev/null; then \
		fswatch -o . | while read f; do \
			echo "Change detected, regenerating coverage..."; \
			make watch-coverage; \
		done; \
	elif command -v inotifywait > /dev/null; then \
		while inotifywait -r -e modify,create,delete .; do \
			echo "Change detected, regenerating coverage..."; \
			make watch-coverage; \
		done; \
	else \
		echo "No file watcher found. Install fswatch or inotify-tools for live coverage updates."; \
		echo "For Ubuntu/Debian: sudo apt-get install inotify-tools"; \
		echo "For macOS: brew install fswatch"; \
	fi
