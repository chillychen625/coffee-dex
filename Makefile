# CoffeeDex Makefile
# The app is a single Go binary (Ebitengine game + SQLite storage)

.PHONY: help run build-server install-deps test lint clean

# Default target
help:
	@echo "CoffeeDex"
	@echo ""
	@echo "Available commands:"
	@echo "  run            - Run the app (default database: ./coffee-dex.db)"
	@echo "  build-server   - Build the binary to bin/coffee-dex"
	@echo "  install-deps   - Download and tidy Go dependencies"
	@echo "  test           - Run tests"
	@echo "  lint           - Run golangci-lint (falls back to go vet)"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Set OPENROUTER_API_KEY in a .env file to enable LLM Pokemon selection."

# Run the app
run:
	go run main.go -db=./coffee-dex.db

# Build the binary
build-server:
	@echo "Building CoffeeDex..."
	mkdir -p bin
	go build -o bin/coffee-dex main.go
	@echo "Built bin/coffee-dex"

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Done."

# Run tests
test:
	go test ./...

# Run linting
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet..."; \
		go vet ./...; \
	fi

# Clean build artifacts
clean:
	rm -rf bin/
	go clean
	@echo "Clean complete."
