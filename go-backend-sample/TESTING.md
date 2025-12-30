# Testing Documentation

This document describes the testing strategy and available tests for the Liyali Gateway API.

## Test Structure

```
tests/
├── unit/                           # Unit tests
│   ├── auth_service_test.go       # ✅ Authentication service tests
│   ├── analytics_service_test.go   # ✅ Analytics service tests (9 test cases)
│   ├── notification_handler_test.go # ✅ Notification handler tests (8 test cases)
│   ├── audit_log_handler_test.go   # ✅ Audit log handler tests (11 test cases)
│   └── README.md                   # Unit test documentation
└── integration/                    # Integration tests
    ├── auth_integration_test.go    # ✅ Auth endpoint integration tests
    ├── analytics_integration_test.go # ✅ Analytics endpoint tests (4 test cases)
    ├── notification_integration_test.go # ✅ Notification endpoint tests (8 test cases)
    ├── audit_log_integration_test.go # ✅ Audit log endpoint tests (9 test cases)
    └── helpers.go                  # ✅ Shared test utilities
```

## Running Tests

### All Tests
```bash
# Run all tests
go test ./tests/...

# Run with coverage
go test -cover ./tests/...

# Run with verbose output
go test -v ./tests/...
```

### Unit Tests Only
```bash
go test ./tests/unit/...
```

### Integration Tests Only
```bash
go test ./tests/integration/...
```

### Specific Test File
```bash
go test ./tests/unit/auth_service_test.go
go test ./tests/integration/analytics_integration_test.go
```

### Run Specific Test
```bash
go test -run TestAnalyticsService_GetDashboardMetrics ./tests/unit/...
```

## Test Coverage Summary

### Unit Tests (28+ test cases)

**Analytics Service** (9 test cases)
- ✅ GetDashboardMetrics - success for admin user
- ✅ GetDashboardMetrics - success for regular user
- ✅ GetDashboardMetrics - detects overdue tasks
- ✅ GetTrendData - success with 7 days
- ✅ GetTrendData - defaults to 7 days if invalid
- ✅ GetTrendData - caps at 90 days
- ✅ GetBottlenecks - identifies bottlenecks
- ✅ GetBottlenecks - ignores stages with less than 3 tasks
- ✅ GetBottlenecks - no bottlenecks with no pending tasks

**Notification Handler** (8 test cases)
- ✅ GetNotifications - success
- ✅ GetNotifications - with pagination
- ✅ GetUnreadNotifications - success
- ✅ GetUnreadCount - success
- ✅ GetNotificationByID - success
- ✅ GetNotificationByID - forbidden for different user
- ✅ MarkAsRead - success
- ✅ MarkAllAsRead - success
- ✅ DeleteNotification - success

**Audit Log Handler** (11 test cases)
- ✅ GetAuditLogs - success for admin user
- ✅ GetAuditLogs - success for manager user
- ✅ GetAuditLogs - forbidden for regular user
- ✅ GetAuditLogs - filter by resource_type
- ✅ GetAuditLogs - filter by action
- ✅ GetAuditLogs - filter by user_id
- ✅ GetMyAuditLogs - success for any user
- ✅ GetMyAuditLogs - with pagination
- ✅ GetAuditLogByID - success for admin
- ✅ GetAuditLogByID - forbidden for regular user
- ✅ GetAuditLogsByResource - success for admin
- ✅ GetAuditLogsByResource - forbidden for regular user

### Integration Tests (21+ test cases)

**Auth Integration** (8 test cases - existing)
- ✅ Register endpoint
- ✅ Register with duplicate email
- ✅ Login success
- ✅ Login with invalid credentials
- ✅ Refresh token
- ✅ Get current user authenticated
- ✅ Get current user unauthenticated
- ✅ Logout success

**Analytics Integration** (4 test cases)
- ✅ GET /api/analytics/metrics - success
- ✅ GET /api/analytics/metrics - unauthorized
- ✅ GET /api/analytics/trends - success
- ✅ GET /api/analytics/trends - custom days
- ✅ GET /api/analytics/bottlenecks - success

**Notification Integration** (8 test cases)
- ✅ GET /api/notifications - success
- ✅ GET /api/notifications/unread - success
- ✅ GET /api/notifications/unread/count - success
- ✅ GET /api/notifications/:id - success
- ✅ POST /api/notifications/:id/read - success
- ✅ POST /api/notifications/read-all - success
- ✅ DELETE /api/notifications/:id - success
- ✅ GET /api/notifications - unauthorized

**Audit Log Integration** (9 test cases)
- ✅ GET /api/audit-logs/my - success
- ✅ GET /api/audit-logs - admin access
- ✅ GET /api/audit-logs - regular user forbidden
- ✅ GET /api/audit-logs/:id - admin success
- ✅ GET /api/audit-logs/:id - regular user forbidden
- ✅ GET /api/audit-logs with filters
- ✅ GET /api/audit-logs/resource/:resource_type/:resource_id - admin success
- ✅ GET /api/audit-logs - unauthorized

## Manual Testing

### Test Scripts

The `backend/` directory contains shell scripts for manual API testing:

**Authentication & Core Features:**
```bash
# Test authentication and basic endpoints
./test_api.sh

# Test all workflow-related endpoints
./test_all_endpoints.sh

# Test newly added endpoints
./test_new_endpoints.sh
```

### Sample Data

To populate the database with test data:
```bash
# Build the seed script
go build -o bin/seed scripts/seed.go

# Run the seed script
./bin/seed
```

This creates:
- 3 sample workflows (Requisition, Budget, Purchase Order)
- 3 sample documents
- Approval tasks for submitted documents
- Notifications for assigned tasks

## Testing Best Practices

### 1. Arrange-Act-Assert Pattern
All tests follow the AAA pattern:
```go
func TestSomething(t *testing.T) {
    // Arrange - set up test data and mocks
    mockRepo := new(MockRepository)
    mockRepo.On("Method", args).Return(result, nil)

    // Act - execute the function being tested
    result, err := service.DoSomething()

    // Assert - verify the results
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertExpectations(t)
}
```

### 2. Table-Driven Tests
Use table-driven tests for testing multiple scenarios:
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "test@example.com", false},
        {"invalid", "not-an-email", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 3. Test Isolation
- Each test should be independent
- Use setup/teardown functions for database state
- Clean up resources after tests

### 4. Meaningful Test Names
Test names should describe what they test:
- ✅ `TestGetDashboardMetrics_AdminUser`
- ✅ `TestMarkAsRead_ForbiddenForDifferentUser`
- ❌ `TestFunction1`
- ❌ `TestCase2`

## Continuous Integration

### GitHub Actions (Example)
```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: liyali_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run migrations
        run: go run scripts/migrate.go

      - name: Run tests
        run: go test -v -cover ./tests/...
```

## Test Data

### Test Users
The integration tests use the following test users:

- **Admin User**
  - Email: `admin@test.com`
  - Role: `ADMIN`
  - Can access all endpoints

- **Manager User**
  - Email: `manager@test.com`
  - Role: `MANAGER`
  - Can access most endpoints

- **Regular User**
  - Email: `user@test.com`
  - Role: `USER`
  - Limited access

### Test Documents
Sample documents created by seed script:
- Office Supplies Requisition ($395.00)
- Q1 2025 Operating Budget ($270,000.00)
- IT Equipment Purchase Order ($9,500.00)

## Coverage Goals

Current test coverage targets:
- **Unit Tests**: >80% coverage of business logic
- **Integration Tests**: >90% coverage of API endpoints
- **Critical Paths**: 100% coverage (authentication, authorization, approvals)

## Future Testing Enhancements

- [ ] Performance/load testing with k6 or similar tools
- [ ] End-to-end tests with frontend integration
- [ ] Security testing (SQL injection, XSS, etc.)
- [ ] API contract testing with Pact
- [ ] Mutation testing to verify test quality
- [ ] Benchmark tests for critical operations
