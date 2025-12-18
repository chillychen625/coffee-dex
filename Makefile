# CoffeeDex Makefile
# Project management and database setup commands

# Variables
MYSQL_USER ?= root
MYSQL_DB ?= coffee_log
MYSQL_HOST ?= localhost
MYSQL_PORT ?= 3306
MYSQL_PASSWORD ?=

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

.PHONY: help install-deps setup-db load-pokemon-data start-server build-desktop run-desktop clean test lint

# Default target
help:
	@echo "$(BLUE)CoffeeDex Project Management$(RESET)"
	@echo ""
	@echo "Available commands:"
	@echo "  $(GREEN)setup-db$(RESET)          - Set up MySQL database with schema"
	@echo "  $(GREEN)load-pokemon-data$(RESET) - Load Pokemon data into database"
	@echo "  $(GREEN)start-server$(RESET)      - Start the Go backend server"
	@echo "  $(GREEN)build-desktop$(RESET)     - Build the Electron desktop app"
	@echo "  $(GREEN)run-desktop$(RESET)       - Run the desktop app"
	@echo "  $(GREEN)install-deps$(RESET)      - Install dependencies"
	@echo "  $(GREEN)test$(RESET)              - Run tests"
	@echo "  $(GREEN)clean$(RESET)             - Clean build artifacts"
	@echo "  $(GREEN)lint$(RESET)              - Run linting"
	@echo "  $(GREEN)full-setup$(RESET)        - Complete setup (db + data + dependencies)"
	@echo ""
	@echo "$(YELLOW)Database Variables:$(RESET)"
	@echo "  MYSQL_USER=$(MYSQL_USER)"
	@echo "  MYSQL_DB=$(MYSQL_DB)"
	@echo "  MYSQL_HOST=$(MYSQL_HOST)"
	@echo "  MYSQL_PORT=$(MYSQL_PORT)"

# Database Setup
setup-db:
	@echo "$(BLUE)Setting up database...$(RESET)"
	@if [ ! -f "setup_pokemon_database.sql" ]; then \
		echo "$(RED)Error: setup_pokemon_database.sql not found$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Creating database and tables...$(RESET)"
	mysql -u $(MYSQL_USER) $(if $(MYSQL_PASSWORD),-p$(MYSQL_PASSWORD)) -h $(MYSQL_HOST) -P $(MYSQL_PORT) < setup_pokemon_database.sql
	@echo "$(GREEN)Database setup complete!$(RESET)"

# Load Pokemon data
load-pokemon-data:
	@echo "$(BLUE)Loading Pokemon data...$(RESET)"
	@if [ ! -f "pokemon_gen1_data.sql" ]; then \
		echo "$(RED)Error: pokemon_gen1_data.sql not found$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Loading Gen 1 Pokemon data...$(RESET)"
	mysql -u $(MYSQL_USER) $(if $(MYSQL_PASSWORD),-p$(MYSQL_PASSWORD)) -h $(MYSQL_HOST) -P $(MYSQL_PORT) $(MYSQL_DB) < pokemon_gen1_data.sql
	@echo "$(GREEN)Pokemon data loaded successfully!$(RESET)"
	@mysql -u $(MYSQL_USER) $(if $(MYSQL_PASSWORD),-p$(MYSQL_PASSWORD)) -h $(MYSQL_HOST) -P $(MYSQL_PORT) $(MYSQL_DB) -e "SELECT COUNT(*) as total_pokemon FROM pokemons;" || echo "Database verification failed"

# Download Pokemon sprites
download-sprites:
	@echo "$(BLUE)Downloading Pokemon sprites...$(RESET)"
	@if [ -f "download_pokemon_sprites.sh" ]; then \
		chmod +x download_pokemon_sprites.sh && ./download_pokemon_sprites.sh; \
		echo "$(GREEN)Pokemon sprites downloaded!$(RESET)"; \
	else \
		echo "$(RED)Sprite download script not found$(RESET)"; \
	fi

# Start the Go server
start-server:
	@echo "$(BLUE)Starting CoffeeDex server...$(RESET)"
	@if [ ! -f "main.go" ]; then \
		echo "$(RED)Error: main.go not found$(RESET)"; \
		exit 1; \
	fi
	go run main.go

# Build the Go server
build-server:
	@echo "$(BLUE)Building CoffeeDex server...$(RESET)"
	mkdir -p bin
	go build -o bin/coffee-dex main.go
	@echo "$(GREEN)Server built successfully!$(RESET)"

# Install dependencies
install-deps:
	@echo "$(BLUE)Installing Go dependencies...$(RESET)"
	go mod download
	go mod tidy
	@echo "$(GREEN)Go dependencies installed!$(RESET)"
	@if [ -d "coffee-dex-desktop" ]; then \
		echo "$(BLUE)Installing Electron app dependencies...$(RESET)"; \
		cd coffee-dex-desktop && npm install; \
		echo "$(GREEN)Electron dependencies installed!$(RESET)"; \
	fi

# Build desktop app
build-desktop:
	@echo "$(BLUE)Building Electron desktop app...$(RESET)"
	@if [ ! -d "coffee-dex-desktop" ]; then \
		echo "$(RED)Error: coffee-dex-desktop directory not found$(RESET)"; \
		exit 1; \
	fi
	cd coffee-dex-desktop && npm run build
	@echo "$(GREEN)Desktop app built successfully!$(RESET)"

# Run desktop app
run-desktop:
	@echo "$(BLUE)Running CoffeeDex desktop app...$(RESET)"
	@if [ ! -d "coffee-dex-desktop" ]; then \
		echo "$(RED)Error: coffee-dex-desktop directory not found$(RESET)"; \
		exit 1; \
	fi
	cd coffee-dex-desktop && npm run dev

# Test database connection
test-db:
	@echo "$(BLUE)Testing database connection...$(RESET)"
	@mysql -u $(MYSQL_USER) $(if $(MYSQL_PASSWORD),-p$(MYSQL_PASSWORD)) -h $(MYSQL_HOST) -P $(MYSQL_PORT) -e "SELECT 'Database connection successful' as status;" && \
	echo "$(GREEN)Database connection successful!$(RESET)" || \
	echo "$(RED)Database connection failed!$(RESET)"

# Run tests
test:
	@echo "$(BLUE)Running tests...$(RESET)"
	go test ./...

# Run linting
lint:
	@echo "$(BLUE)Running linting...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed, running go vet...$(RESET)"; \
		go vet ./...; \
	fi

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	rm -rf dist/
	@if [ -d "coffee-dex-desktop" ]; then \
		cd coffee-dex-desktop && rm -rf node_modules/ dist/ release/; \
	fi
	go clean
	@echo "$(GREEN)Clean complete!$(RESET)"

# Complete setup - database + data + dependencies
full-setup: install-deps setup-db load-pokemon-data download-sprites
	@echo "$(GREEN)Complete setup finished!$(RESET)"
	@echo "$(BLUE)You can now run 'make start-server' and 'make run-desktop'$(RESET)"

# Check if MySQL is running
check-mysql:
	@echo "$(BLUE)Checking MySQL status...$(RESET)"
	@if command -v mysql >/dev/null 2>&1; then \
		mysql -u $(MYSQL_USER) $(if $(MYSQL_PASSWORD),-p$(MYSQL_PASSWORD)) -h $(MYSQL_HOST) -P $(MYSQL_PORT) -e "SELECT VERSION();" 2>/dev/null && \
		echo "$(GREEN)MySQL is running and accessible$(RESET)" || \
		echo "$(RED)MySQL is not accessible$(RESET)"; \
	else \
		echo "$(RED)MySQL client not found$(RESET)"; \
	fi

# Database reset (WARNING: This drops and recreates the database)
reset-db:
	@echo "$(RED)WARNING: This will drop and recreate the database!$(RESET)"
	@read -p "Are you sure? (y/N) " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		mysql -u $(MYSQL_USER) $(if $(MYSQL_PASSWORD),-p$(MYSQL_PASSWORD)) -h $(MYSQL_HOST) -P $(MYSQL_PORT) -e "DROP DATABASE IF EXISTS $(MYSQL_DB); CREATE DATABASE $(MYSQL_DB);"; \
		echo "$(GREEN)Database reset complete!$(RESET)"; \
	else \
		echo "$(YELLOW)Database reset cancelled$(RESET)"; \
	fi

# Show database info
db-info:
	@echo "$(BLUE)Database Information:$(RESET)"
	@mysql -u $(MYSQL_USER) $(if $(MYSQL_PASSWORD),-p$(MYSQL_PASSWORD)) -h $(MYSQL_HOST) -P $(MYSQL_PORT) $(MYSQL_DB) -e \
		"SELECT 'Pokemons' as table_name, COUNT(*) as count FROM pokemons \
		UNION ALL \
		SELECT 'Coffee-Pokemon Mappings' as table_name, COUNT(*) as count FROM coffee_pokemon;"

run:
	go run main.go -storage=mysql \
		-mysql-user=coffee_user \
		-mysql-password=coffee_pass123 \
		-mysql-db=coffee_log