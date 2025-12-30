# Unit Tests

This directory contains unit tests for individual components of the Liyali Gateway Backend.

## Test Files

### Service Layer Tests
- `analytics_service_test.go` - Tests for analytics calculations and data processing
- `approval_rules_test.go` - Tests for approval workflow rules and validation logic
- `budget_validation_test.go` - Tests for budget constraint validation
- `document_linking_test.go` - Tests for document relationship management
- `notification_service_test.go` - Tests for notification delivery and formatting
- `workflow_state_machine_test.go` - Tests for state transition logic

### Handler Layer Tests
- `budget_handler_test.go` - Tests for budget API endpoints
- `category_handler_test.go` - Tests for category API endpoints
- `grn_handler_test.go` - Tests for GRN (Goods Received Note) API endpoints
- `payment_voucher_handler_test.go` - Tests for payment voucher API endpoints
- `purchase_order_handler_test.go` - Tests for purchase order API endpoints
- `requisition_handler_test.go` - Tests for requisition API endpoints
- `roles_test.go` - Tests for role management API endpoints
- `vendor_handler_test.go` - Tests for vendor API endpoints

### Disabled Tests
- `role_management_service_test.go.disabled` - Role management service tests (disabled)

## Running Unit Tests

### Run All Unit Tests
```bash
go test ./tests/unit/...
```

### Run Specific Service Tests
```bash
go test ./tests/unit/analytics_service_test.go
go test ./tests/unit/approval_rules_test.go
```

### Run Specific Handler Tests
```bash
go test ./tests/unit/budget_handler_test.go
go test ./tests/unit/requisition_handler_test.go
```

### Run with Verbose Output
```bash
go test -v ./tests/unit/...
```

### Run with Coverage
```bash
go test -cover ./tests/unit/...
```

## Test Guidelines

### Service Tests
- Test business logic in isolation
- Mock external dependencies (database, APIs, etc.)
- Test both success and error scenarios
- Validate input/output transformations
- Test edge cases and boundary conditions

### Handler Tests
- Test HTTP request/response handling
- Mock service layer dependencies
- Test request validation
- Test response formatting
- Test error handling and status codes
- Test middleware integration

### Best Practices
1. **Isolation**: Each test should be independent
2. **Mocking**: Use mocks for external dependencies
3. **Coverage**: Aim for high test coverage
4. **Clarity**: Write clear, descriptive test names
5. **Documentation**: Include comments for complex test logic
6. **Data**: Use realistic but safe test data

## Test Structure

Each test file should follow this structure:

```go
package unit

import (
    "testing"
    // other imports
)

// Test setup helpers
func setupTestService() *Service {
    // Setup code
}

// Test cases
func TestServiceMethod_Success(t *testing.T) {
    // Test successful operation
}

func TestServiceMethod_Error(t *testing.T) {
    // Test error scenarios
}

func TestServiceMethod_EdgeCases(t *testing.T) {
    // Test edge cases
}
```

## Mocking

Use appropriate mocking libraries:
- Database: Use test database or mock interfaces
- HTTP clients: Mock HTTP responses
- External services: Mock service interfaces
- Time: Mock time-dependent operations

## Test Data

- Use consistent test data patterns
- Avoid hardcoded values where possible
- Use test data generators for unique values
- Clean up test data after tests complete