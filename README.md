# 🛍️ Go Product Catalog API

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Latest-2496ED?style=flat-square&logo=docker&logoColor=white)](https://www.docker.com)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

Production-grade REST API for product catalog management with clean architecture, comprehensive filtering, and pagination.

---

## 📋 Overview

Go Product Catalog API is a backend service for managing product catalogs with categories, variants, and advanced filtering capabilities. Built as a hiring challenge solution, the API demonstrates clean architecture principles, comprehensive test coverage, and production-ready practices including structured error handling, pagination, and database optimization.

The system features:
- **Multi-tier architecture** with clear separation of concerns (HTTP, Service, Repository, Model layers)
- **Advanced querying** with pagination (offset/limit), category filtering, and price range searches
- **Product variants** with inherited pricing and detailed product composition
- **Category management** with deterministic ordering for consistent API responses
- **Centralized response formatting** with structured error handling
- **Comprehensive testing** with 40+ unit tests and integration tests with real PostgreSQL

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 📦 **Product Catalog** | Create, retrieve, and list products with detailed information |
| 🏷️ **Category System** | Hierarchical product categorization with CRUD operations |
| 📊 **Pagination** | Offset and limit-based pagination (constraints: 1–100 items) |
| 🔍 **Advanced Filtering** | Filter by category (case-insensitive) and maximum price |
| 🎨 **Product Variants** | Support for product sizes, colors with inherited pricing |
| 💰 **Price Management** | Decimal precision for accurate financial calculations |
| 📝 **Centralized Responses** | Unified JSON response format with error codes and request IDs |
| 🧪 **Comprehensive Testing** | 40+ unit tests + integration tests with PostgreSQL |
| 🔐 **Database Migrations** | Progressive constraint hardening with safe migration practices |
| 📈 **Performance** | Indexed queries and optimized repository patterns |

---

## 🏗️ Architecture

### Layer Architecture

```
┌─────────────────────────────────────────────────────┐
│                   HTTP Client                       │
└─────────────────────┬───────────────────────────────┘
                      │ REST
┌─────────────────────▼───────────────────────────────┐
│           Transport Layer (Handlers)                │
│      HTTP routing · DTOs · Response formatting      │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│             Repository Layer                        │
│        Database queries · Filtering · Pagination    │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                Model Layer (GORM)                   │
│         Entity definitions · Relationships           │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│             Infrastructure Layer                    │
│        PostgreSQL · Migrations · Constraints        │
└─────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Location | Responsibilities |
|-------|----------|------------------|
| **Transport** | `app/catalog/`, `app/categories/` | HTTP handlers, request/response serialization, error mapping |
| **Repository** | `models/*_repository.go` | Database queries, filtering logic, pagination |
| **Model** | `models/*.go` | GORM entities, relationships, business object definitions |
| **Infrastructure** | `app/database/`, `sql/` | PostgreSQL connection, migrations, schema management |
| **API Response** | `app/api/response.go` | Centralized JSON formatting, error wrapping, status codes |

---

## 🛠️ Tech Stack

| Category | Technology | Version |
|----------|-----------|---------|
| **Language** | Go | 1.21+ |
| **Web Framework** | Standard Library (net/http) | Built-in |
| **ORM** | GORM | Latest |
| **Database** | PostgreSQL | 15-alpine |
| **Decimal** | shopspring/decimal | Latest |
| **Testing** | testify (assert/require) | Latest |
| **Database Driver** | pgx | Latest |
| **Containers** | Docker + Docker Compose | Latest |

---

## 📁 Project Structure

```
go-hiring-challenge-1.2.0/
├── cmd/
│   ├── server/
│   │   └── main.go                         # API server entry point
│   └── seed/
│       └── main.go                         # Database seeding utility
├── app/
│   ├── api/
│   │   ├── response.go                     # Centralized response formatting
│   │   └── response_test.go                # Response tests
│   ├── catalog/
│   │   ├── handler.go                      # Product catalog endpoints
│   │   └── handler_test.go                 # Catalog handler tests
│   ├── categories/
│   │   ├── handler.go                      # Category management endpoints
│   │   └── handler_test.go                 # Category handler tests
│   ├── database/
│   │   └── pg.go                           # PostgreSQL connection setup
│   └── integration_test.go                 # End-to-end integration tests
├── models/
│   ├── db.go                               # Database interface abstraction
│   ├── products.go                         # Product entity
│   ├── variants.go                         # Product variant entity
│   ├── categories.go                       # Category entity
│   ├── products_repository.go              # Product repository
│   ├── products_repository_test.go         # Product repository tests
│   ├── categories_repository.go            # Category repository
│   └── categories_repository_test.go       # Category repository tests
├── sql/
│   ├── 000-truncate.sql                    # Test data cleanup
│   ├── 001-products.sql                    # Products table
│   ├── 002-variants.sql                    # Variants table
│   ├── 003-product-data.sql                # Sample product data
│   ├── 004-categories.sql                  # Categories table
│   ├── 005-category-data.sql               # Sample category data
│   ├── 006-add-product-name.sql            # Name column (migration)
│   ├── 007-product-code-constraints.sql    # Code NOT NULL + UNIQUE
│   └── 008-category-id-constraints.sql     # Category NOT NULL + index
├── Dockerfile                              # Multi-stage build
├── docker-compose.yml                      # PostgreSQL + API orchestration
├── Makefile                                # Build targets and automation
├── go.mod                                  # Go module definition
├── go.sum                                  # Dependency lock file
└── README.md                               # This file
```

---

## 🚀 Getting Started

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.21+ | Build and run locally |
| **Docker** | 24+ | Run containerized services |
| **Docker Compose** | 2.x+ | Service orchestration |
| **make** | any | Build automation |

### Step 1 — Clone the Repository

```bash
git clone https://github.com/alex-necsoiu/go-hiring-challenge-1.2.0.git
cd go-hiring-challenge-1.2.0
```

### Step 2 — Configure Environment

```bash
cp .env.example .env
# Optional: Edit .env if needed (defaults work with docker-compose)
```

### Step 3 — Start the Full Stack

```bash
make dev-up
```

This starts:
- PostgreSQL 15 on `localhost:5432`
- API on `localhost:8000`

### Step 4 — Verify Everything is Running

```bash
# Health check
curl http://localhost:8000/health

# List products
curl http://localhost:8000/catalog
```

---

## 📡 API Reference

### Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| **GET** | `/catalog` | List products with pagination and filters |
| **GET** | `/catalog/:code` | Get product details including variants |
| **GET** | `/categories` | List all categories |
| **POST** | `/categories` | Create a new category |

### Example 1 — List Products

**Request:**
```bash
curl -X GET "http://localhost:8000/catalog?offset=0&limit=10&category=CLOTHING&priceLessThan=100" \
  -H "Content-Type: application/json"
```

**Response (200 OK):**
```json
{
  "data": {
    "products": [
      {
        "id": 1,
        "code": "PROD001",
        "name": "T-Shirt",
        "price": "29.99",
        "category": {
          "id": 1,
          "code": "CLOTHING",
          "name": "Clothing & Apparel"
        },
        "variants": [
          {
            "id": 1,
            "sku": "PROD001-S-BLACK",
            "size": "S",
            "color": "Black",
            "price": null
          }
        ]
      }
    ],
    "total": 42,
    "limit": 10,
    "offset": 0
  },
  "meta": {
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Example 2 — Get Product by Code

**Request:**
```bash
curl -X GET "http://localhost:8000/catalog/PROD001" \
  -H "Content-Type: application/json"
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "code": "PROD001",
    "name": "T-Shirt",
    "price": "29.99",
    "category": {
      "id": 1,
      "code": "CLOTHING",
      "name": "Clothing & Apparel"
    },
    "variants": [
      {
        "id": 1,
        "sku": "PROD001-S-BLACK",
        "size": "S",
        "color": "Black",
        "price": null
      },
      {
        "id": 2,
        "sku": "PROD001-M-WHITE",
        "size": "M",
        "color": "White",
        "price": "34.99"
      }
    ]
  },
  "meta": {
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Example 3 — Create Category

**Request:**
```bash
curl -X POST "http://localhost:8000/categories" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "ELECTRONICS",
    "name": "Electronics & Gadgets"
  }'
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 4,
    "code": "ELECTRONICS",
    "name": "Electronics & Gadgets"
  },
  "meta": {
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "Invalid pagination: limit must be between 1 and 100",
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Error Codes

| Status | Code | Description |
|--------|------|-------------|
| 400 | `INVALID_INPUT` | Request validation failed (invalid filters, pagination) |
| 404 | `NOT_FOUND` | Resource does not exist |
| 409 | `CONFLICT` | Duplicate key (e.g., duplicate category code) |
| 500 | `INTERNAL_ERROR` | Unexpected server error |

---

## 🗄️ Database Design

### Entity Relationships

```
categories (id, code, name)
    ↑
    │ (one-to-many)
    │
products (id, code, name, price, category_id)
    ↓
    │ (one-to-many)
    │
product_variants (id, product_code, sku, size, color, price)
```

### Key Constraints

- **Products**: `code` is UNIQUE and NOT NULL
- **Categories**: `code` is UNIQUE and NOT NULL
- **Variants**: Inherit price from product if not explicitly set
- **Ordering**: Categories are ordered by `code ASC` for deterministic API responses
- **Foreign Keys**: Category references maintained with ON DELETE RESTRICT

### Migrations

| Migration | Purpose |
|-----------|---------|
| `001-products.sql` | Create products table with basic schema |
| `002-variants.sql` | Create product_variants table |
| `003-product-data.sql` | Seed sample product data |
| `004-categories.sql` | Create categories table and add foreign key |
| `005-category-data.sql` | Seed sample category data |
| `006-add-product-name.sql` | Add name column to products |
| `007-product-code-constraints.sql` | Add NOT NULL and UNIQUE constraints on code |
| `008-category-id-constraints.sql` | Add NOT NULL constraint on category_id + index |

---

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL hostname |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `products_db` | Database name |
| `DB_SSL_MODE` | `disable` | SSL mode: `disable` or `require` |
| `SERVER_PORT` | `8000` | API listen port |

---

## 🧪 Testing

### Test Targets

| Command | Purpose |
|---------|---------|
| `make test` | Run all unit tests |
| `make test-race` | Run tests with race detector |
| `make test-coverage` | Generate coverage report (HTML) |

### Coverage

- **app/api**: Response formatting and error handling
- **app/catalog**: Product listing, filtering, details endpoints
- **app/categories**: Category CRUD operations
- **models**: Repository patterns, database queries, filtering logic

### Running Tests

```bash
# All unit tests
make test

# With race detector
make test-race

# Coverage report
make test-coverage
```

---

## 🔧 Development

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make help` | Show all available targets |
| `make build` | Compile the API binary |
| `make run` | Build and run locally |
| `make dev-up` | Start Docker stack (PostgreSQL + API) |
| `make dev-down` | Stop all containers |
| `make test` | Run unit tests |
| `make test-race` | Run tests with race detector |
| `make test-coverage` | Generate HTML coverage report |
| `make fmt` | Format code |
| `make vet` | Run go vet static analysis |
| `make docker-build` | Build Docker image |
| `make clean` | Remove build artifacts |

### Code Formatting

```bash
# Format all Go files
make fmt

# Check formatting (CI-safe, exits 1 if issues found)
make fmt-check

# Run static analysis
make vet
```

---

## 🐳 Docker

### Services

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| **api** | Built from Dockerfile | 8000 | REST API |
| **postgres** | `postgres:15-alpine` | 5432 | Primary database |

### Multi-Stage Dockerfile

The Dockerfile uses a two-stage build for efficiency:

**Stage 1: Build**
- Base: `golang:1.21-alpine`
- Installs dependencies
- Compiles binary with version info
- Result: ~18MB binary

**Stage 2: Runtime**
- Base: `alpine:3.18`
- Copies optimized binary
- Minimal attack surface
- Final image: ~20MB

### Running Services

```bash
# Start all services
make dev-up

# View logs
make logs

# Stop services
make dev-down
```

---

## 📄 License

MIT License — Copyright (c) 2026 Alex Necsoiu

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software.

---

## 👤 Author

**Alex Necsoiu**

- GitHub: [@alex-necsoiu](https://github.com/alex-necsoiu)
- Email: [axel.necsoiu@gmail.com](mailto:axel.necsoiu@gmail.com)

💡 For questions or contributions, please open an issue or submit a pull request on GitHub.
