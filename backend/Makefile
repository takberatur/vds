# Database Variables
DB_USER=postgres
DB_PASSWORD=
DB_NAME=video_downloader
DB_HOST=localhost
DB_PORT=5432
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Go Variables
BINARY_NAME=main
BUILD_DIR=bin
CMD_PATH=cmd/api/main.go
WORKER_PATH=cmd/worker/main.go
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')
TOML_FILE=.air.toml
MIGRATION_DIR=internal/repository/migrations
SEEDER_DIR=cmd/seed/main.go

# Colors for output
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

# Phony targets
.PHONY: all build run clean test dev createdb dropdb migrate-create migrate-up migrate-down

all: build

# Build the application
build:
	@echo "$(GREEN)Building application...$(NC)"
	@CGO_ENABLED=0 go build -ldflags="-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o bin/api cmd/api/main.go
	@echo "$(GREEN)Build application complete! Binaries available in bin/$(NC)"
	@echo "$(GREEN)Building worker...$(NC)"
	@CGO_ENABLED=0 go build -ldflags="-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o bin/worker cmd/worker/main.go
	@echo "$(GREEN)Build worker complete! Binaries available in bin/$(NC)"

# Run the application (build binaries first)
run: build
	@echo "$(GREEN)Starting application...$(NC)"
	@air -c $(TOML_FILE)

# Run Woker
worker:
	@echo "$(GREEN)Starting worker...$(NC)"
	@./bin/worker

# Live Reload using Air (build binaries first)
dev:
	@echo "$(GREEN)Starting live reload with Air...$(NC)"
	@air -c $(TOML_FILE)


## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf bin/ tmp/ *.log
	@echo "$(GREEN)Clean complete!$(NC)"

# Run tests
test:
	go test -v ./...
	@echo "$(GREEN)Test complete!$(NC)"
		
## tidy: Tidy dependencies
tidy:
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	@go mod tidy
	@echo "$(GREEN)Tidy complete!$(NC)"

## vendor: Vendor dependencies
vendor:
	@echo "$(GREEN)Vendoring dependencies...$(NC)"
	@go mod vendor
	@echo "$(GREEN)Vendor complete!$(NC)"

# Database Commands (Requires golang-migrate and local postgres client)
createdb:
	@echo "$(GREEN)Creating database $(DB_NAME)...$(NC)"
	createdb --username=$(DB_USER) --host=$(DB_HOST) --port=$(DB_PORT) $(DB_NAME)

dropdb:
	@echo "$(RED)Dropping database $(DB_NAME)...$(NC)"
	dropdb --username=$(DB_USER) --host=$(DB_HOST) --port=$(DB_PORT) $(DB_NAME)

# Migration Commands
# Usage: make migrate-create name=init_schema
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Please provide migration name (e.g., make migrate-create name=create_users_table)$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating migration: $(name)$(NC)"
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(name)
	@echo "$(GREEN)Migration files created in $(MIGRATION_DIR)$(NC)"

migrate-up:
	@echo "$(GREEN)Running migrations...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "$(GREEN)Migrations complete!$(NC)"

migrate-down:
	@echo "$(GREEN)Running migrations down...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down
	@echo "$(GREEN)Migrations down complete!$(NC)"

## migrate-down-all: Rollback all migrations (DANGER!)
migrate-down-all:
	@echo "$(RED)Rolling back ALL migrations...$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down -all; \
		echo "$(GREEN)All migrations rolled back!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled!$(NC)"; \
	fi

## migrate-force: Force migration version (usage: make migrate-force version=1)	
migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "$(RED)Error: Please provide version (e.g., make migrate-force version=1)$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migration version to $(version)...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $(version)
	@echo "$(GREEN)Version $(version) forced!$(NC)"

## migrate-version: Show current migration version
migrate-version:
	@echo "$(GREEN)Current migration version...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" version

## migrate-reset: Reset database (down all, then up)
migrate-reset: migrate-down-all migrate-up
	@echo "$(GREEN)Database reset complete!$(NC)"

## migrate-force-reset: Force reset with schema cleanup
migrate-hard-reset:
	@echo "$(RED)HARD RESETTING DATABASE...$(NC)"
	@echo "$(RED)THIS WILL DELETE ALL DATA!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(MAKE) dropdb; \
		$(MAKE) createdb; \
		$(MAKE) migrate-up; \
		echo "$(GREEN)Hard reset complete!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled!$(NC)"; \
	fi
	
## migrate-force-reset: Alias for hard reset
migrate-force-reset: migrate-hard-reset

## seed: Seed database with initial data
seed:
	@echo "$(GREEN)Seeding database...$(NC)"
	go run $(SEEDER_DIR)
	@echo "$(GREEN)Seeding complete!$(NC)"
