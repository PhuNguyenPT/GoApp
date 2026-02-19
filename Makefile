# Simple Makefile for a Go project
-include .env

# Build the application
all: build test

# Generate templ files
templ-generate:
	@echo "Generating templ files..."
	@go run github.com/a-h/templ/cmd/templ@latest generate

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
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi


docker-watch:
	@if docker compose up watch psql --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up watch psql --build; \
	fi

# Shutdown watch container
docker-watch-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

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
	@go run github.com/air-verse/air@latest -c .air.toml

# Generate sqlc files
sqlc-generate:
	@echo "Generating sqlc files..."
	@go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

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