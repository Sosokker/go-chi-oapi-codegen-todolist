.PHONY: help generate build run test migrate-up migrate-down clean docker-db docker-db-stop dev

BINARY_NAME=todolist-server
CMD_PATH=./cmd/server
OUTPUT_DIR=./bin
CONFIG_PATH=.

# Tool Paths (adjust if needed)
OAPI_CODEGEN=oapi-codegen
SQLC=sqlc
MIGRATE=migrate

OPENAPI=openapi.yaml
PKG=api

DB_URL?=postgresql://postgres:@localhost:5433/postgres?sslmode=disable
MIGRATIONS_PATH=./backend/migrations

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  help             Show this help message."
	@echo "  generate         Generate Go code from OpenAPI spec and SQL queries."
	@echo "  build            Build the Go application binary."
	@echo "  run              Build and run the Go application."
	@echo "  test             Run Go tests."
	@echo "  migrate-up       Apply all up database migrations."
	@echo "  migrate-down     Roll back the last database migration."
	@echo "  migrate-force    Force set migration version (e.g., make migrate-force VERSION=1)."
	@echo "  clean            Remove build artifacts."
	@echo "  docker-db        Start a PostgreSQL container using Docker."
	@echo "  docker-db-stop   Stop and remove the PostgreSQL container."
	@echo ""
	@echo "Environment Variables:"
	@echo "  DB_URL           Database connection URL (used for migrations)."
	@echo "                   Default: Attempts to get from running 'todolist-db' container."
	@echo "                   Example: export DB_URL='postgres://user:password@localhost:5432/todolist?sslmode=disable'"

dev:
	@echo ">> Starting development server with Air live reload..."
	@air -c .air.toml

generate-types:
	oapi-codegen --package $(PKG) --generate types -o internal/api/openapi_types.go $(OPENAPI)

generate-chi:
	oapi-codegen --package $(PKG) --generate chi-server -o internal/api/openapi_generated.go $(OPENAPI)

generate-models:
	oapi-codegen --package models --generate models -o internal/api/models/openapi_models_generated.go $(OPENAPI)

# generate-strict:
# 	oapi-codegen --package $(PKG) --generate strict-server -o internal/api/openapi_strict_server_generated.go $(OPENAPI)

generate: generate-types generate-chi generate-models
	@echo ">> Generating SQLC code..."
	$(SQLC) generate
	@echo ">> Tidying modules..."
	go mod tidy

build:
	@echo ">> Building binary..."
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) $(CMD_PATH)/main.go

run: build
	@echo ">> Running application..."
	$(OUTPUT_DIR)/$(BINARY_NAME) -config=$(CONFIG_PATH)

test:
	@echo ">> Running tests..."
	go test ./... -v -cover

migrate-up:
	@echo ">> Applying migrations..."
	@if [ -z "$(DB_URL)" ]; then echo "Error: DB_URL is not set. Run 'make docker-db' or set it manually."; exit 1; fi
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_PATH) up

migrate-down:
	@echo ">> Rolling back last migration..."
	@if [ -z "$(DB_URL)" ]; then echo "Error: DB_URL is not set. Run 'make docker-db' or set it manually."; exit 1; fi
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_PATH) down 1

migrate-force:
	@echo ">> Forcing migration version $(VERSION)..."
	@if [ -z "$(DB_URL)" ]; then echo "Error: DB_URL is not set. Run 'make docker-db' or set it manually."; exit 1; fi
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION is not set. Usage: make migrate-force VERSION=<version_number>"; exit 1; fi
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_PATH) force $(VERSION)

clean:
	@echo ">> Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)