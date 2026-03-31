# Integration Tests Guide

This document explains how to run and write integration tests for the Go Hiring Challenge API.

## Overview

Integration tests verify the full API flow end-to-end, including:
- Database migrations and schema setup
- Data persistence and retrieval
- HTTP request/response cycles
- Full handler chains from routing through repository layers

Integration tests are located in `app/integration_test.go` with the `+build integration` build tag.

## Prerequisites

1. **Docker Compose** - PostgreSQL database
2. **Go 1.24.3+** - Go compiler with GOTOOLCHAIN=auto
3. **.env file** - Database configuration (included in repo)

## Running Integration Tests

### Starting the Database

**Option 1: Using Make target**
```bash
make docker-up
```

**Option 2: Using Docker Compose directly**
```bash
docker compose up -d
```

Verify the database is running:
```bash
docker compose ps
```

### Running the Tests

**Option 1: Using Make (recommended)**

Run only integration tests:
```bash
make test-integration
```

Run all tests (unit + integration):
```bash
make test-all
```

**Option 2: Using Go directly**

From the project root directory:
```bash
GOTOOLCHAIN=auto go test -v -tags integration ./app/...
```

Run specific test:
```bash
GOTOOLCHAIN=auto go test -v -tags integration ./app -run TestIntegration_CatalogEndpoint_ListProducts
```

### Stopping the Database

```bash
make docker-down
```

Or directly:
```bash
docker compose down
```

## Test Coverage

### Integration Test Suites

#### 1. **TestIntegration_CatalogEndpoint_ListProducts**
Tests the `GET /catalog` endpoint with:
- Default pagination
- Custom offset/limit
- Category filtering
- Price filtering
- Combined filters
- Invalid parameters

#### 2. **TestIntegration_CatalogEndpoint_ProductDetails**
Tests the `GET /catalog/{code}` endpoint with:
- Product with variants
- Product without variants
- Non-existent product (404)

#### 3. **TestIntegration_CategoriesEndpoint_ListCategories**
Tests the `GET /categories` endpoint with:
- Retrieving all categories
- Response format validation

#### 4. **TestIntegration_CategoriesEndpoint_CreateCategory**
Tests the `POST /categories` endpoint with:
- Successful creation
- Duplicate code handling (conflict)
- Missing required fields (validation)

#### 5. **TestIntegration_FullWorkflow**
Tests a complete workflow:
- Create categories
- List categories
- Verify empty catalog
- Demonstrates multi-endpoint interactions

## Test Fixtures and Helpers

### Test Data Creation

The integration test suite includes helpers to set up test data:

```go
// Create test categories
categories := createTestCategories(t, suite.DB)

// Create test products with variants
products := createTestProducts(t, suite.DB, categories)
```

Test data includes:
- 3 categories (Clothing, Shoes, Accessories)
- 8 products spanning all categories
- Products with and without variants
- Various price points for filtering tests

### Database Setup/Teardown

Automatic handling:
- **Setup**: Runs migrations from `./sql` directory
- **Teardown**: Truncates all data while preserving schema

```go
suite := SetupIntegrationTest(t)
defer TeardownIntegrationTest(t, suite)
```

### HTTP Request Helpers

Convenience functions for making requests:

```go
// Generic request
resp, body := doRequest(suite, http.MethodGet, "/catalog", nil)

// JSON request
resp, body := doJSONRequest(suite, http.MethodPost, "/categories", payload)
```

## Writing New Integration Tests

### Template

```go
func TestIntegration_YourTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup
    suite := SetupIntegrationTest(t)
    defer TeardownIntegrationTest(t, suite)

    // Create test data
    categories := createTestCategories(t, suite.DB)
    products := createTestProducts(t, suite.DB, categories)

    // Make request
    resp, body := doRequest(suite, http.MethodGet, "/your-endpoint", nil)

    // Assert
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    // ... more assertions
}
```

### Key Points

1. **Build tag**: Always include `// +build integration` at the top
2. **Skip on -short**: Check `testing.Short()` to skip in CI if needed
3. **Setup/Teardown**: Use the helper functions for consistent test environment
4. **Assertions**: Use `testify/assert` for readable assertions
5. **JSON parsing**: Helpers handle JSON marshaling/unmarshaling

## Troubleshooting

### Docker Connection Refused
```
dial error: dial tcp [::1]:5432: connect: connection refused
```
**Solution**: Start Docker and PostgreSQL container
```bash
make docker-up
docker compose logs postgres  # Check if container is healthy
```

### Port Already in Use
```
bind: address already in use
```
**Solution**: Stop existing containers
```bash
docker compose down -v  # -v removes volumes
make docker-up
```

### .env File Not Found
The tests automatically search for `.env` in both the current directory and parent directory.
Ensure `.env` exists in the project root with required variables:
```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=challenge
POSTGRES_PORT=5432
```

### Test Timeout
Increase timeout if needed:
```bash
GOTOOLCHAIN=auto go test -v -tags integration -timeout 30s ./app/...
```

## Database Migrations

Migrations run automatically for each test suite from the `./sql` directory in order:
1. `000-truncate.sql` - Clear existing data
2. `001-products.sql` - Products table
3. `002-variants.sql` - Variants table
4. `003-product-data.sql` - Sample data
5. `004-categories.sql` - Categories table
6. `005-category-data.sql` - Category data
7. `006-add-product-name.sql` - Schema updates

## Performance Considerations

- Integration tests are slower than unit tests (~0.8-2s depending on queries)
- Each test runs migrations (included in timing)
- Tests truncate tables between runs (fresh state for each test)
- Database connections are reused across all tests in a suite

## CI/CD Integration

For CI/CD pipelines, use:

```bash
# Unit tests only (fast, no Docker needed)
make test-unit

# All tests with Docker (slower, needs Docker service)
make test-all
```

Environment variables for CI:
```bash
POSTGRES_DB=challenge_test  # Optional: use separate test database
POSTGRES_HOST=postgres      # May differ in containerized CI
```

## Debugging Failed Tests

### Print HTTP Response Body
```go
t.Logf("Response body: %s", body)
```

### Enable SQL Logging
```go
db, _ := database.New(...)
// GORM logs queries by default in development
```

### Stop After First Failure
```bash
make test-integration -count=1 | head -100
```

## Best Practices

1. ✅ Keep tests independent - don't share state between tests
2. ✅ Use meaningful test names - describe what's being tested
3. ✅ Test both happy path and error cases
4. ✅ Verify response status codes AND response content
5. ✅ Use table-driven tests for multiple scenarios
6. ✅ Clean up resources in defer statements
7. ✅ Document complex test scenarios with comments
8. ❌ Don't make assertions outside of test functions
9. ❌ Don't rely on specific database IDs (use fixtures)
10. ❌ Don't skip error checking in setup/teardown

## Future Enhancements

- [ ] Parallel test execution with isolated databases
- [ ] Performance benchmarks for critical paths
- [ ] Load testing with multiple concurrent requests
- [ ] Database state assertions (cross-check via different queries)
- [ ] Response time assertions
- [ ] Contract testing against API schema

## Resources

- [Testing in Go](https://golang.org/doc/effective_go#testing)
- [testify/assert Documentation](https://github.com/stretchr/testify)
- [httptest Package](https://pkg.go.dev/net/http/httptest)
- [GORM Testing Guide](https://gorm.io/docs/test.html)
