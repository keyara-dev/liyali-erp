# Testing Guide

## Overview

Comprehensive testing documentation for the Liyali Gateway backend.

## Test Structure

```
backend/tests/
├── unit/              # Unit tests
├── integration/       # Integration tests
└── helpers/           # Test utilities
```

## Running Tests

### All Tests

```bash
go test ./...
```

### Unit Tests Only

```bash
go test ./tests/unit/...
```

### Integration Tests

```bash
go test ./tests/integration/...
```

### Specific Test File

```bash
go test ./tests/unit/auth_service_test.go -v
```

### With Coverage

```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Scripts

### Admin Endpoints

```bash
# Windows
.\backend\scripts\admin_tests.ps1

# Linux/Mac
./backend/scripts/admin_tests.sh
```

### API Coverage

```bash
./backend/scripts/run_tests.sh
```

### Workflow Tests

```bash
./backend/scripts/workflow_test.sh
```

## API Testing

### Using HTTP Files

Test files in `backend/scripts/test_requests.http`:

```http
### Login
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "password"
}
```

### Using cURL

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'

# Get users (with auth)
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Test Coverage

Current coverage status:

- ✅ Authentication & Authorization
- ✅ User Management
- ✅ Organization Management
- ✅ Subscription System
- ✅ Workflow Engine
- ✅ Document Management
- ✅ Admin Endpoints
- ✅ RBAC System

See `backend/scripts/API_ENDPOINT_COVERAGE_REPORT.md` for detailed coverage.

## Writing Tests

### Unit Test Example

```go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()

    service := NewUserService(db)

    // Test
    user, err := service.CreateUser(context.Background(), CreateUserParams{
        Email: "test@example.com",
        Name: "Test User",
    })

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Integration Test Example

```go
func TestAuthFlow_Integration(t *testing.T) {
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()

    // Register user
    resp := registerUser(t, server, "test@example.com", "password")
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // Login
    token := loginUser(t, server, "test@example.com", "password")
    assert.NotEmpty(t, token)

    // Access protected endpoint
    user := getProfile(t, server, token)
    assert.Equal(t, "test@example.com", user.Email)
}
```

## Quick Test Guide

For quick testing of specific features, see `QUICK_TEST_GUIDE.md` in the root directory.

## Troubleshooting

### Database Connection Issues

```bash
# Check database is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d liyali_gateway
```

### Test Failures

1. Ensure database is clean: `make reset-db`
2. Run migrations: `make migrate`
3. Seed test data: `make seed`
4. Clear cache: `go clean -testcache`

## CI/CD Integration

Tests run automatically on:

- Pull requests
- Commits to develop/main
- Before deployment

See `.github/workflows/` for CI configuration.
