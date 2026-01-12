# Test Organization

This directory contains all test files for the Liyali Gateway Backend, organized into unit and integration tests with a clean architecture.

## Directory Structure

```
tests/
├── unit/                    # Unit tests for individual components
│   ├── *_service_test.go   # Service layer tests
│   ├── *_handler_test.go   # Handler layer tests
│   └── *_test.go           # Other unit tests
├── integration/            # Integration tests for complete workflows
│   ├── *_integration_test.go # End-to-end workflow tests
│   └── helpers.go          # Test helper functions
└── README.md               # This file

../scripts/                  # Testing scripts and tools
├── test_requests.http      # HTTP requests for manual API testing
└── run_comprehensive_tests.sh # Comprehensive test suite runner
```

## Test Categories

### Unit Tests (`tests/unit/`)

- **Service Tests**: Test business logic in isolation

  - `analytics_service_test.go` - Analytics calculations and data processing
  - `approval_rules_test.go` - Approval workflow rules and validation
  - `budget_validation_test.go` - Budget constraint validation
  - `document_automation_service_test.go` - Document automation workflows
  - `document_linking_test.go` - Document relationship management
  - `notification_service_test.go` - Notification delivery and formatting
  - `workflow_execution_service_test.go` - Workflow execution logic
  - `workflow_state_machine_test.go` - State transition logic

- **Handler Tests**: Test HTTP request/response handling
  - `*_handler_test.go` - API endpoint behavior and validation

### Integration Tests (`tests/integration/`)

- **Workflow Tests**: Test complete business processes
  - `approval_flow_integration_test.go` - Complete approval workflows
  - `budget_constraint_integration_test.go` - Budget validation across services
  - `document_automation_integration_test.go` - Document automation integration
  - `multi_tenant_analytics_test.go` - Multi-tenant analytics testing
  - `workflow_integration_complete_test.go` - Complete workflow integration
  - `workflows_integration_test.go` - End-to-end workflow execution

## Running Tests

### Comprehensive Test Suite

For complete system testing, use the comprehensive test suite:

```bash
# From backend directory
./scripts/run_comprehensive_tests.sh
```

This runs all 47 API endpoints with automated reporting and covers:

- Authentication and authorization
- CRUD operations for all entities
- Workflow execution and approvals
- Multi-tenant isolation
- Error handling and validation

### Individual Test Commands

### Run All Tests

```bash
go test ./tests/...
```

### Run Unit Tests Only

```bash
go test ./tests/unit/...
```

### Run Integration Tests Only

```bash
go test ./tests/integration/...
```

### Run Specific Test File

```bash
go test ./tests/unit/analytics_service_test.go
```

### Run with Verbose Output

```bash
go test -v ./tests/...
```

### Run with Coverage

```bash
go test -cover ./tests/...
```

## Test Guidelines

1. **Unit Tests**: Should test individual functions/methods in isolation
2. **Integration Tests**: Should test complete workflows and interactions between components
3. **Test Data**: Use realistic but anonymized test data
4. **Cleanup**: Always clean up test data and resources
5. **Isolation**: Tests should not depend on each other
6. **Documentation**: Include clear test descriptions and comments

## Test Helpers

Common test utilities and helpers are located in:

- `tests/integration/helpers.go` - Integration test helpers
- Individual test files may include their own helper functions

## Database Testing

For tests requiring database access:

1. Use test database configuration
2. Run migrations before tests
3. Clean up test data after tests
4. Use transactions for isolation when possible

## Mocking

Use mocks for external dependencies:

- Database connections
- External API calls
- File system operations
- Time-dependent operations
