# =============================================================================
# Go Hiring Challenge — Makefile
# =============================================================================
# Production-grade Makefile for Go backend catalog API with PostgreSQL.
# Follows clean architecture patterns and comprehensive testing practices.

# =============================================================================
# ⚙️  VARIABLES — Configuration Center
# =============================================================================

GO                := go
GOFLAGS           := 
BINARY_NAME       := server
BINARY_PATH       := $(PWD)/bin/$(BINARY_NAME)
COVERAGE_MIN      := 47
VERSION           := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME        := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
COMMIT            := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Database configuration (from .env, with fallbacks)
DB_HOST           ?= localhost
DB_PORT           ?= 5432
DB_USER           ?= postgres
DB_PASSWORD       ?= postgres
DB_NAME           ?= hiring_challenge_db
DB_SSL_MODE       ?= disable

# Derived database variables
DB_DSN            := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)
MIGRATION_PATH    := sql

# Server configuration
SERVER_PORT       ?= 8080
SERVER_ENV        ?= development

# Docker Compose
DOCKER_COMPOSE    := docker compose

# ldflags for version embedding
LDFLAGS           := -ldflags "\
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.Commit=$(COMMIT)"

# Test configuration
TEST_COVERAGE_FILE := coverage.out
TEST_COVERAGE_HTML := coverage.html

# =============================================================================
# 📋 PHONY TARGETS — Mark all targets as .PHONY
# =============================================================================

.PHONY: \
	help \
	build \
	run run-seed \
	fmt fmt-check \
	vet lint \
	test test-unit test-race test-coverage \
	docker-up docker-down docker-logs \
	db-up db-down db-reset \
	tidy clean \
	.DEFAULT_GOAL

.DEFAULT_GOAL := help

# =============================================================================
# 🆘 HELP — Display all available targets with descriptions
# =============================================================================

help:
	@echo "╔════════════════════════════════════════════════════════════════════╗"
	@echo "║     Go Hiring Challenge — Catalog API — Make Targets               ║"
	@echo "╚════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "┌─ Core Development ─────────────────────────────────────────────────┐"
	@echo "│ make build              Build the API binary (with version info)   │"
	@echo "│ make run                Run the API server locally                 │"
	@echo "│ make run-seed           Run the database seeder                    │"
	@echo "│ make tidy               Run 'go mod tidy' and vendor               │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Code Quality ─────────────────────────────────────────────────────┐"
	@echo "│ make fmt                Format all Go files (gofmt)                │"
	@echo "│ make fmt-check          Check if formatting is needed (CI safe)    │"
	@echo "│ make vet                Run go vet (static analysis)               │"
	@echo "│ make lint               Run golangci-lint (if installed)           │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Testing ──────────────────────────────────────────────────────────┐"
	@echo "│ make test               Run all tests with coverage                │"
	@echo "│ make test-unit          Run unit tests (no docker required)        │"
	@echo "│ make test-race          Run tests with race detector               │"
	@echo "│ make test-coverage      Generate coverage report (fails <47%)      │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Docker / Database ────────────────────────────────────────────────┐"
	@echo "│ make docker-up          Start PostgreSQL via docker compose        │"
	@echo "│ make docker-down        Stop all containers                       │"
	@echo "│ make docker-logs        Tail container logs                       │"
	@echo "│ make db-up              Alias for 'docker-up'                     │"
	@echo "│ make db-down            Alias for 'docker-down'                   │"
	@echo "│ make db-reset           Stop containers + cleanup volume          │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Maintenance ──────────────────────────────────────────────────────┐"
	@echo "│ make clean              Remove build artifacts & coverage reports │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "📝 Configuration:"
	@echo "   DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_NAME=$(DB_NAME)"
	@echo "   SERVER_PORT=$(SERVER_PORT) SERVER_ENV=$(SERVER_ENV)"
	@echo ""

# =============================================================================
# 🎨 FORMATTING — Code quality enforcement
# =============================================================================

fmt:
	@echo "🎨 Formatting all Go files (gofmt)..."
	@$(GO) fmt ./...
	@echo "✅ Formatting complete!"

fmt-check:
	@echo "🔍 Checking if code needs formatting..."
	@if [ -n "$$($(GO) fmt ./... 2>&1)" ]; then \
		echo "❌ Code formatting issues found. Run 'make fmt' to fix."; \
		exit 1; \
	fi
	@echo "✅ All files are properly formatted!"

# =============================================================================
# 🔎 STATIC ANALYSIS — Code quality checks
# =============================================================================

vet:
	@echo "🔍 Running go vet (static analysis)..."
	@$(GO) vet ./...
	@echo "✅ No vet issues found!"

lint:
	@echo "🔍 Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "⚠️  golangci-lint not found. Install with:"; \
		echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	}
	@golangci-lint run ./...
	@echo "✅ No lint issues found!"

# =============================================================================
# 🏗️  BUILD — Compile the binary with version info
# =============================================================================

build:
	@echo "🏗️  Building binary: $(BINARY_NAME)"
	@echo "   Version: $(VERSION)"
	@echo "   Build Time: $(BUILD_TIME)"
	@echo "   Commit: $(COMMIT)"
	@mkdir -p bin
	@$(GO) build $(LDFLAGS) -o $(BINARY_PATH) ./cmd/server
	@echo "✅ Binary built: $(BINARY_PATH)"
	@ls -lh $(BINARY_PATH)

# =============================================================================
# 🧪 TESTING — Multiple test modes for different scenarios
# =============================================================================

test: test-unit test-coverage
	@echo "✅ All tests passed!"

test-unit:
	@echo "🧪 Running unit tests..."
	@$(GO) test -v -short -count=1 ./...
	@echo "✅ Unit tests passed!"

test-race:
	@echo "🧪 Running tests with race detector..."
	@$(GO) test -race -v -count=1 ./...
	@echo "✅ No data races detected!"

test-coverage:
	@echo "🧪 Running tests with coverage analysis..."
	@$(GO) test -v -count=1 -race ./... -coverprofile=$(TEST_COVERAGE_FILE) -covermode=atomic
	@echo "📊 Generating HTML coverage report..."
	@$(GO) tool cover -html=$(TEST_COVERAGE_FILE) -o $(TEST_COVERAGE_HTML)
	@echo "📈 Coverage report: file://$(PWD)/$(TEST_COVERAGE_HTML)"
	@echo ""
	@COVERAGE=$$($(GO) tool cover -func=$(TEST_COVERAGE_FILE) | grep ^total | awk '{print $$3}' | sed 's/%//'); \
	echo "📌 Total Coverage: $$COVERAGE%"; \
	if [ "$$(echo "$$COVERAGE < $(COVERAGE_MIN)" | bc)" -eq 1 ]; then \
		echo "❌ Coverage below $(COVERAGE_MIN)% threshold!"; \
		exit 1; \
	fi
	@echo "✅ Coverage meets $(COVERAGE_MIN)% threshold!"

# =============================================================================
# ▶️  RUN — Execute locally
# =============================================================================

run: build
	@echo "▶️  Running API server..."
	@echo "   Server: http://localhost:$(SERVER_PORT)"
	@echo "   Database: $(DB_DSN)"
	@$(BINARY_PATH)

run-seed:
	@echo "▶️  Running database seeder..."
	@$(GO) run cmd/seed/main.go

# =============================================================================
# 🐳 DOCKER — Container management with docker compose
# =============================================================================

docker-up:
	@echo "🐳 Starting PostgreSQL container..."
	@$(DOCKER_COMPOSE) up -d
	@echo "⏳ Waiting for services to start..."
	@sleep 2
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "✅ PostgreSQL running!"
	@echo "   Host: $(DB_HOST):$(DB_PORT)"
	@echo "   Database: $(DB_NAME)"

docker-down:
	@echo "⛔ Stopping containers..."
	@$(DOCKER_COMPOSE) down
	@echo "✅ Containers stopped!"

docker-logs:
	@echo "📋 Tailing container logs (press Ctrl+C to stop)..."
	@$(DOCKER_COMPOSE) logs -f

db-up: docker-up

db-down: docker-down

db-reset: docker-down
	@echo "🗑️  Removing PostgreSQL volume..."
	@$(DOCKER_COMPOSE) down -v
	@echo "✅ Database reset complete!"

# =============================================================================
# 📦 DEPENDENCY MANAGEMENT
# =============================================================================

tidy:
	@echo "📦 Running 'go mod tidy' and 'go mod vendor'..."
	@$(GO) mod tidy
	@$(GO) mod vendor
	@echo "✅ Dependencies tidied!"

# =============================================================================
# 🧹 CLEANUP — Remove build artifacts
# =============================================================================

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/ $(TEST_COVERAGE_FILE) $(TEST_COVERAGE_HTML)
	@echo "✅ Cleanup complete!"

# =============================================================================
# 📌 END OF MAKEFILE
# =============================================================================
# Principles:
# • Version info embedded at build time via ldflags
# • Separate unit/race test targets for flexible CI/CD
# • Coverage enforcement with configurable threshold
# • All code quality checks (fmt, vet, lint) integrated
# • Docker Compose for container orchestration
# • Comprehensive help text with current settings displayed
# • Simple, maintainable structure aligned with project needs
