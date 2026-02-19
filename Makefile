# Simple Makefile for a Go project
-include .env

# Build the application
all: build test

# Generate templ files
templ-generate:
	@echo "Generating templ files..."
	@if command -v templ > /dev/null 2>&1; then \
		templ generate; \
	else \
		go run github.com/a-h/templ/cmd/templ@latest generate; \
	fi

# Build Tailwind CSS
tailwind-build:
	@echo "Building Tailwind CSS..."
	@cd frontend-template && npx tailwindcss -i ./public/styles/index.css -o ./public/output.css

# Build the application
build: templ-generate sqlc-generate tailwind-build
	@echo "Building Go binary..."
	@go build -o main cmd/api/main.go

# Run SSR server + SPA frontend
run: templ-generate
	@go run cmd/api/main.go &
	@npm install --prefer-offline --no-fund --prefix ./frontend
	@npm run dev --prefix ./frontend

# Create DB container
docker-run:
	@docker compose up --build

# Shutdown DB container
docker-down:
	@docker compose down


docker-watch:
	@docker compose up watch psql frontend --build

# Shutdown watch container
docker-watch-down:
	@docker compose down

# Test the application
test: templ-generate
	@echo "Testing..."
	@go test ./... -v

# Integration tests
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary and generated files
clean:
	@echo "Cleaning..."
	@rm -f main
	@find internal/views -name "*_templ.go" -delete

# Development with hot reload
watch:
	@if command -v air > /dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		go run github.com/air-verse/air@latest -c .air.toml; \
	fi

# Generate sqlc files
sqlc-generate:
	@echo "Generating sqlc files..."
	@if command -v sqlc > /dev/null 2>&1; then \
		sqlc generate; \
	else \
		go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate; \
	fi

# Run goose migrations
migrate-up:
	@echo "Running migrations..."
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/database/migrations postgres "postgres://$(POSTGRES_USERNAME):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable&search_path=$(POSTGRES_SCHEMA)" up

migrate-down:
	@echo "Rolling back migration..."
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/database/migrations postgres "postgres://$(POSTGRES_USERNAME):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable&search_path=$(POSTGRES_SCHEMA)" down

# Lint
lint:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

lint-fix:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix

# Static analysis
vet:
	@go vet ./...

# Format code
fmt:
	@gofmt -w .	

.PHONY: all build run test clean watch docker-run docker-down docker-watch docker-watch-down itest templ-generate tailwind-build sqlc-generate migrate-up migrate-down lint lint-fix vet fmt