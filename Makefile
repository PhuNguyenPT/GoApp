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

# Format templ files
templ-fmt:
	@echo "Formatting templ files..."
	@if command -v templ > /dev/null 2>&1; then \
		templ fmt ./internal/views/; \
	else \
		go tool templ fmt ./internal/views/; \
	fi

# Build Tailwind CSS
tailwind-build:
	@echo "Building Tailwind CSS..."
	@cd frontend-template && npx tailwindcss -i ./public/styles/index.css -o ./public/output.css --minify

# Build the application
build: templ-generate sqlc-generate tailwind-build
	@echo "Building Go binary..."
	@go build -o main cmd/api/main.go

# Run SSR server + SPA frontend
run: templ-generate
	@go run cmd/api/main.go &
	@npm install --prefer-offline --no-fund --prefix ./frontend
	@npm run dev --prefix ./frontend

# Run with watch profile (dev)
docker-watch:
	@docker compose --profile dev up --build

# Shutdown watch
docker-watch-down:
	@docker compose --profile dev down

# Run with prod profile
docker-prod:
	@docker compose --profile prod up --build

# Shutdown prod
docker-prod-down:
	@docker compose --profile prod down

# Test the application
test: templ-generate sqlc-generate
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
	@go tool sqlc compile

lint-fix:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix

# Static analysis
vet:
	@go vet ./...
	@go tool sqlc vet

# Format code (Go + templ)
fmt: templ-fmt
	@gofmt -w .

.PHONY: all build run test clean watch docker-watch docker-watch-down docker-prod docker-prod-down itest templ-generate templ-fmt tailwind-build sqlc-generate migrate-up migrate-down lint lint-fix vet fmt