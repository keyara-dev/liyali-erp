# Unit Tests

This directory contains unit tests for the Liyali Gateway API.

## Running Tests

To run all unit tests:
```bash
go test ./tests/unit/...
```

To run a specific test file:
```bash
go test ./tests/unit/auth_service_test.go
```

To run with coverage:
```bash
go test -cover ./tests/unit/...
```

To run with verbose output:
```bash
go test -v ./tests/unit/...
```

## Test Structure

- `auth_service_test.go` - Tests for authentication service
- `analytics_service_test.go` - Tests for analytics service (requires mocking)
- `notification_handler_test.go` - Tests for notification handlers (requires mocking)
- `audit_log_handler_test.go` - Tests for audit log handlers (requires mocking)

## Note on Mocking

The analytics, notification, and audit log tests use mock repositories. To properly run these tests, you need to:

1. Install testify for mocking:
```bash
go get github.com/stretchr/testify/mock
```

2. Ensure your mock structs implement the full repository interfaces

## Integration Tests

For full end-to-end testing with a real database, see the integration tests in `tests/integration/`.
