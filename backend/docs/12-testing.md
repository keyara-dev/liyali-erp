# Testing Guide

Comprehensive testing strategy for the Liyali Gateway Backend.

## Testing Overview

The backend implements a multi-layered testing strategy with a modular test suite architecture:

- **Unit Tests** - Test individual components in isolation (Go unit tests)
- **Integration Tests** - Test component interactions and database operations
- **API Tests** - Test HTTP endpoints and request/response handling
- **Performance Tests** - Test system performance under load
- **Security Tests** - Test authentication, authorization, and security measures
- **Workflow Tests** - Comprehensive workflow and approval system testing

## Current Test Structure

### Directory Organization

```
backend/
├── scripts/                    # Test execution scripts (modular test suite)
│   ├── common_tests.sh        # Shared utilities and configurations
│   ├── run_tests.sh           # Main orchestrator script
│   ├── auth_tests.sh          # Authentication & session management
│   ├── rbac_tests.sh          # Role-based access control & multi-tenant
│   ├── document_tests.sh      # Document management (CRUD operations)
│   ├── workflow_test.sh       # Comprehensive workflow & approval system
│   ├── workflow_unit_tests.sh # Workflow-specific unit tests (Go)
│   ├── custom_role_tests.sh   # Custom role workflow API tests
│   ├── department_tests.sh    # Department management
│   ├── analytics_tests.sh     # Analytics, notifications & system ops
│   ├── error_tests.sh         # Error handling & edge cases
│   └── README_TESTS.md        # Detailed test suite documentation
├── tests/                     # Go test files
│   ├── unit/                  # Unit tests
│   │   ├── workflow_concurrency_fixes_test.go
│   │   ├── custom_role_edge_cases_test.go
│   │   ├── custom_role_validation_test.go
│   │   └── concurrent_approval_issues_test.go
│   ├── integration/           # Integration tests
│   │   ├── workflow_api_integration_test.go
│   │   └── custom_role_workflow_integration_test.go
│   └── workflow_execution_service_test.go
└── docs/
    └── 12-testing.md          # This file
```

### Test Coverage Areas

#### 🔐 Authentication & Authorization (`auth_tests.sh`)

- User registration and login
- JWT token management and refresh
- Password change and reset
- Session management
- Multi-factor authentication

#### 🏢 Multi-Tenant & RBAC (`rbac_tests.sh`)

- Organization operations
- Role-based access control
- Custom organization roles
- Permission management
- Multi-tenant isolation

#### 📄 Document Management (`document_tests.sh`)

- CRUD operations for all document types
- Category and vendor management
- Document search and filtering
- File upload and management
- Document status workflows

#### 🔄 Workflow System (`workflow_test.sh`)

- Comprehensive workflow testing including:
  - Unit tests (delegates to `workflow_unit_tests.sh`)
  - Integration tests for API endpoints
  - Custom role workflow functionality
  - Performance and coverage analysis
  - Concurrency and optimistic locking
  - Approval task management

#### 🧪 Workflow Unit Tests (`workflow_unit_tests.sh`)

- Workflow concurrency fixes
- Custom role validation
- Workflow execution service
- Edge case handling
- Integration test support

#### 👥 Custom Role Workflows (`custom_role_tests.sh`)

- Custom role creation and validation
- Role-based workflow assignment
- Permission inheritance
- Role hierarchy testing

#### 🏬 Department Management (`department_tests.sh`)

- Department CRUD operations
- Hierarchical department structures
- Department-based permissions
- Budget allocation by department

#### 📊 Analytics & System (`analytics_tests.sh`)

- Reporting and analytics endpoints
- Notification system
- System health monitoring
- Performance metrics

#### ⚠️ Error Handling (`error_tests.sh`)

- Error response validation
- Edge case handling
- Security vulnerability testing
- Rate limiting and throttling

## Modular Test Architecture

### Design Principles

The Liyali Gateway test suite follows a modular architecture designed for:

1. **Maintainability** - Each test module focuses on a specific domain
2. **Reusability** - Shared utilities in `common_tests.sh`
3. **Scalability** - Easy to add new test modules
4. **Flexibility** - Run individual modules or comprehensive suite
5. **Context Management** - Persistent authentication across modules

### Test Module Structure

Each test module follows a consistent structure:

```bash
#!/bin/bash

# Source common utilities
source "$(dirname "$0")/common_tests.sh"

# Module-specific test functions
test_specific_functionality() {
    print_section_header "FUNCTIONALITY TESTING" "🧪"

    local auth_header="-H 'Authorization: Bearer $ACCESS_TOKEN' -H 'X-Organization-ID: $ORGANIZATION_ID'"

    print_status "TESTING" "Specific Feature"
    make_request "GET" "$API_URL/endpoint" "" "$auth_header" 200
}

# Main execution function
run_module_tests() {
    reset_test_counters

    # Check authentication context
    if [ -z "$ACCESS_TOKEN" ] || [ -z "$ORGANIZATION_ID" ]; then
        print_status "ERROR" "Authentication context required"
        return 1
    fi

    test_specific_functionality

    print_module_summary "MODULE NAME"
    return 0
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    check_server

    # Load or establish auth context
    if [ -z "$ACCESS_TOKEN" ]; then
        source "$(dirname "$0")/auth_tests.sh"
        run_auth_tests
    fi

    run_module_tests
fi
```

### Shared Utilities (`common_tests.sh`)

The common utilities provide:

- **HTTP Request Functions** - Standardized API testing
- **Authentication Management** - Token persistence and validation
- **Output Formatting** - Consistent status messages and colors
- **Test Counters** - Tracking passed/failed tests across modules
- **Context Persistence** - Save/load test context between runs
- **Server Health Checks** - Verify server availability

### Context Management

The test suite maintains persistent context:

```bash
# Context file: ~/.liyali_test_context
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
REFRESH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
ORGANIZATION_ID="org-demo-001"
USER_ID="user-admin-001"
VENDOR_ID="vendor-001"
WORKFLOW_ID="workflow-123"
# ... other test entities
```

This enables:

- **Fast Test Execution** - Skip authentication for subsequent runs
- **Cross-Module Dependencies** - Share created entities between modules
- **Consistent State** - Reliable test environment across runs

## Best Practices

### Test Organization

1. **Separate Concerns** - Each module tests a specific domain
2. **Descriptive Names** - Test names should describe what is being tested
3. **Consistent Structure** - Follow the established module pattern
4. **Error Handling** - Graceful handling of failures and edge cases
5. **Cleanup** - Proper resource cleanup after tests

### Test Quality

1. **Comprehensive Coverage** - Test both happy and error paths
2. **Edge Cases** - Include boundary conditions and edge cases
3. **Performance Awareness** - Include performance considerations
4. **Security Testing** - Validate authentication and authorization
5. **Documentation** - Clear comments and documentation

### Adding New Test Modules

To add a new test module:

1. **Create Script** - Follow the module structure template
2. **Update Orchestrator** - Add module to `run_tests.sh`
3. **Update Documentation** - Add to README_TESTS.md
4. **Test Integration** - Verify module works standalone and integrated

```bash
# Example: Adding a new module
cp backend/scripts/document_tests.sh backend/scripts/new_module_tests.sh

# Edit the new module
vim backend/scripts/new_module_tests.sh

# Add to run_tests.sh
# Add to README_TESTS.md
# Test the module
./scripts/new_module_tests.sh
./scripts/run_tests.sh new_module
```

#### Test Environment Setup

```go
type TestConfig struct {
    DatabaseURL string
    JWTSecret   string
    AppPort     string
}

func LoadTestConfig() *TestConfig {
    return &TestConfig{
        DatabaseURL: getEnv("TEST_DB_URL", "postgres://postgres:password@localhost/liyali_gateway_test?sslmode=disable"),
        JWTSecret:   getEnv("JWT_SECRET", "test-secret-key-minimum-32-characters"),
        AppPort:     getEnv("TEST_APP_PORT", "8081"),
    }
}
```

#### Test Database Setup

```go
func setupTestDB(t *testing.T) *gorm.DB {
    config := LoadTestConfig()

    db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err)

    // Use bootstrap system for test database setup
    bootstrapper := bootstrap.NewBootstrapper(db, &bootstrap.BootstrapConfig{
        Environment:       "test",
        SkipSeeding:      false, // Enable seeding for tests
        SeedRetryAttempts: 3,
        SeedRetryDelay:   time.Second,
    }, logger.Default)

    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
    defer cancel()

    result := bootstrapper.Bootstrap(ctx)
    require.True(t, result.Success, "Bootstrap failed: %v", result.Error)

    // Clean up after test
    t.Cleanup(func() {
        cleanupTestDB(t, db)
    })

    return db
}
```

## Unit Testing

### Handler Unit Tests

```go
// tests/unit/handlers/requisition_handler_test.go
func TestRequisitionHandler_CreateRequisition(t *testing.T) {
    // Setup
    mockService := &mocks.RequisitionService{}
    handler := handlers.NewRequisitionHandler(mockService)

    app := fiber.New()
    app.Post("/requisitions", handler.CreateRequisition)

    // Test data
    reqBody := types.CreateRequisitionRequest{
        Title:       "Test Requisition",
        Description: "Test Description",
        TotalAmount: 1000.00,
    }

    // Mock response
    expectedReq := &models.Requisition{
        ID:          "test-id",
        Title:       reqBody.Title,
        Description: reqBody.Description,
        TotalAmount: reqBody.TotalAmount,
    }
    mockService.On("CreateRequisition", mock.AnythingOfType("string"), mock.AnythingOfType("*types.CreateRequisitionRequest")).Return(expectedReq, nil)

    // Execute request
    reqJSON, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/requisitions", bytes.NewReader(reqJSON))
    req.Header.Set("Content-Type", "application/json")

    // Add user context
    ctx := req.Context()
    ctx = context.WithValue(ctx, "userID", "user-id")
    req = req.WithContext(ctx)

    resp, _ := app.Test(req)

    // Assertions
    assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
    mockService.AssertExpectations(t)
}
```

### Service Unit Tests

```go
// tests/unit/services/requisition_service_test.go
func TestRequisitionService_CreateRequisition(t *testing.T) {
    // Setup
    mockRepo := &mocks.RequisitionRepository{}
    mockUserRepo := &mocks.UserRepository{}
    mockAuthService := &mocks.AuthService{}

    service := services.NewRequisitionService(mockRepo, mockUserRepo, mockAuthService)

    // Test data
    userID := "user-id"
    orgID := "org-id"
    user := &models.User{
        ID:             userID,
        OrganizationID: orgID,
    }

    req := &types.CreateRequisitionRequest{
        Title:       "Test Requisition",
        Description: "Test Description",
        TotalAmount: 1000.00,
    }

    // Mock expectations
    mockAuthService.On("HasPermission", userID, "requisitions.create").Return(true)
    mockUserRepo.On("GetByID", userID).Return(user, nil)
    mockRepo.On("Create", mock.AnythingOfType("*models.Requisition")).Return(&models.Requisition{
        ID:             "new-id",
        OrganizationID: orgID,
        Title:          req.Title,
        Description:    req.Description,
        TotalAmount:    req.TotalAmount,
    }, nil)

    // Execute
    result, err := service.CreateRequisition(userID, req)

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, req.Title, result.Title)
    assert.Equal(t, orgID, result.OrganizationID)
    assert.Equal(t, req.TotalAmount, result.TotalAmount)

    mockAuthService.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
}
```

## Bootstrap System Testing

### Bootstrap Integration Tests

```go
// tests/integration/bootstrap/bootstrap_integration_test.go
func TestBootstrapSystem_FullCycle(t *testing.T) {
    // Setup test database
    db := setupCleanTestDB(t)

    // Create bootstrap configuration
    config := &bootstrap.BootstrapConfig{
        Environment:       "test",
        SkipSeeding:      false,
        SeedRetryAttempts: 3,
        SeedRetryDelay:   time.Millisecond * 100,
        CircuitBreakerConfig: circuit.Config{
            MaxFailures: 2,
            Timeout:     time.Second * 30,
            Interval:    time.Second * 60,
        },
    }

    logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
    bootstrapper := bootstrap.NewBootstrapper(db, config, logger)

    // Test bootstrap execution
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
    defer cancel()

    result := bootstrapper.Bootstrap(ctx)

    // Assertions
    assert.True(t, result.Success, "Bootstrap should succeed")
    assert.NoError(t, result.Error)
    assert.Greater(t, result.Duration, time.Duration(0))
    assert.NotNil(t, result.Metrics)

    // Verify database state
    verifyDatabaseSchema(t, db)
    verifySeededData(t, db)

    // Test idempotency - run bootstrap again
    result2 := bootstrapper.Bootstrap(ctx)
    assert.True(t, result2.Success, "Bootstrap should be idempotent")
}

func TestBootstrapSystem_CircuitBreaker(t *testing.T) {
    // Setup database that will fail
    db := setupFailingTestDB(t)

    config := &bootstrap.BootstrapConfig{
        Environment: "test",
        CircuitBreakerConfig: circuit.Config{
            MaxFailures: 2,
            Timeout:     time.Second * 5,
            Interval:    time.Second * 10,
        },
    }

    logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
    bootstrapper := bootstrap.NewBootstrapper(db, config, logger)

    ctx := context.Background()

    // First few attempts should fail and open circuit breaker
    for i := 0; i < 3; i++ {
        result := bootstrapper.Bootstrap(ctx)
        assert.False(t, result.Success)
    }

    // Circuit breaker should now be open
    metrics := bootstrapper.GetMetrics()
    assert.Equal(t, "OPEN", metrics["circuit_breaker_state"])

    // Immediate retry should fail fast
    start := time.Now()
    result := bootstrapper.Bootstrap(ctx)
    duration := time.Since(start)

    assert.False(t, result.Success)
    assert.Less(t, duration, time.Second, "Should fail fast when circuit is open")
}
```

### Bootstrap Performance Tests

```go
// tests/performance/bootstrap/bootstrap_performance_test.go
func BenchmarkBootstrap_FullCycle(b *testing.B) {
    db := setupBenchmarkDB(b)

    config := bootstrap.DefaultBootstrapConfig()
    config.Environment = "test"

    logger := log.New(io.Discard, "", 0) // Silent logger for benchmarks
    bootstrapper := bootstrap.NewBootstrapper(db, config, logger)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
        result := bootstrapper.Bootstrap(ctx)
        cancel()

        if !result.Success {
            b.Fatalf("Bootstrap failed: %v", result.Error)
        }

        // Clean database for next iteration
        cleanupBenchmarkDB(b, db)
    }
}

func BenchmarkBootstrap_IdempotentRuns(b *testing.B) {
    db := setupBenchmarkDB(b)

    config := bootstrap.DefaultBootstrapConfig()
    config.Environment = "test"

    logger := log.New(io.Discard, "", 0)
    bootstrapper := bootstrap.NewBootstrapper(db, config, logger)

    // Initial bootstrap
    ctx := context.Background()
    result := bootstrapper.Bootstrap(ctx)
    if !result.Success {
        b.Fatalf("Initial bootstrap failed: %v", result.Error)
    }

    b.ResetTimer()

    // Benchmark subsequent idempotent runs
    for i := 0; i < b.N; i++ {
        result := bootstrapper.Bootstrap(ctx)
        if !result.Success {
            b.Fatalf("Bootstrap failed: %v", result.Error)
        }
    }
}
```

### API Integration Tests

```go
// tests/integration/api/requisition_api_test.go
func TestRequisitionAPI_CreateRequisition(t *testing.T) {
    // Setup test server
    app := setupTestApp(t)

    // Create test user and get JWT
    jwt := createTestUserAndGetJWT(t, app)

    // Test create requisition
    reqBody := map[string]interface{}{
        "title":       "Integration Test Requisition",
        "description": "Integration Testing",
        "totalAmount": 1000.00,
    }

    reqJSON, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/v1/requisitions", bytes.NewReader(reqJSON))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+jwt)

    resp, _ := app.Test(req)
    assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

    // Parse response
    var response struct {
        Success bool                   `json:"success"`
        Data    models.Requisition     `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&response)

    assert.True(t, response.Success)
    assert.Equal(t, reqBody["title"], response.Data.Title)
}
}
```

### Database Integration Tests

```go
// tests/integration/database/document_sync_test.go
func TestRequisitionToDocumentSync(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create repositories
    reqRepo := repository.NewRequisitionRepository(db)
    documentRepo := repository.NewDocumentRepository(db)

    // Create requisition
    req := &models.Requisition{
        OrganizationID: "test-org",
        Title:          "Test Requisition",
        Description:    "Test Description",
        Status:         "pending",
        TotalAmount:    1000.00,
        CreatedBy:      "user-123",
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }

    createdReq, err := reqRepo.Create(req)
    assert.NoError(t, err)

    // Wait for trigger to execute
    time.Sleep(100 * time.Millisecond)

    // Verify document was created
    doc, err := documentRepo.GetByID(createdReq.ID)
    assert.NoError(t, err)
    assert.NotNil(t, doc)
    assert.Equal(t, "requisition", doc.DocumentType)
    assert.Equal(t, req.Title, doc.Title)
    assert.Equal(t, req.Description, doc.Description)
    assert.Equal(t, req.Status, doc.Status)
    assert.Equal(t, req.TotalAmount, *doc.TotalAmount)
}
```

## Performance Testing

### Load Testing

```go
// tests/performance/load/api_load_test.go
func TestRequisitionAPI_LoadTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }

    // Test configuration
    baseURL := "http://localhost:8080"
    concurrency := 10
    requestsPerWorker := 100

    // Setup test data
    jwt := setupTestUserAndGetJWT(t)
    reqBody := map[string]interface{}{
        "title":       "Load Test Requisition",
        "description": "Generated for load testing",
        "totalAmount": 1000.00,
    }

    // Metrics
    var (
        totalRequests   int64
        successRequests int64
        failedRequests  int64
        totalDuration   time.Duration
        mu              sync.Mutex
    )

    // Worker function
    worker := func(workerID int) {
        client := &http.Client{Timeout: 30 * time.Second}

        for i := 0; i < requestsPerWorker; i++ {
            start := time.Now()

            // Create request
            reqJSON, _ := json.Marshal(reqBody)
            req, _ := http.NewRequest("POST", baseURL+"/api/v1/requisitions", bytes.NewReader(reqJSON))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("Authorization", "Bearer "+jwt)

            // Execute request
            resp, err := client.Do(req)
            duration := time.Since(start)

            // Update metrics
            mu.Lock()
            totalRequests++
            totalDuration += duration

            if err != nil || resp.StatusCode >= 400 {
                failedRequests++
            } else {
                successRequests++
            }

            if resp != nil {
                resp.Body.Close()
            }
            mu.Unlock()
        }
    }

    // Start load test
    start := time.Now()
    var wg sync.WaitGroup

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            worker(workerID)
        }(i)
    }

    wg.Wait()
    totalTestDuration := time.Since(start)

    // Calculate metrics
    averageResponseTime := totalDuration / time.Duration(totalRequests)
    requestsPerSecond := float64(totalRequests) / totalTestDuration.Seconds()
    successRate := float64(successRequests) / float64(totalRequests) * 100

    // Print results
    t.Logf("Load Test Results:")
    t.Logf("  Total Requests: %d", totalRequests)
    t.Logf("  Success Rate: %.2f%%", successRate)
    t.Logf("  Average Response Time: %v", averageResponseTime)
    t.Logf("  Requests Per Second: %.2f", requestsPerSecond)

    // Assertions
    assert.Greater(t, successRate, 95.0, "Success rate should be > 95%")
    assert.Less(t, averageResponseTime, 500*time.Millisecond, "Average response time should be < 500ms")
}
```

## Security Testing

### Authentication Security Tests

```go
// tests/security/auth/auth_security_test.go
func TestAuthenticationSecurity(t *testing.T) {
    app := integration.SetupTestApp(t)

    tests := []struct {
        name           string
        endpoint       string
        method         string
        headers        map[string]string
        expectedStatus int
        description    string
    }{
        {
            name:           "missing authorization header",
            endpoint:       "/api/v1/requisitions",
            method:         "GET",
            expectedStatus: 401,
            description:    "Should return 401 for missing auth header",
        },
        {
            name:     "invalid JWT token",
            endpoint: "/api/v1/requisitions",
            method:   "GET",
            headers: map[string]string{
                "Authorization": "Bearer invalid-token",
            },
            expectedStatus: 401,
            description:    "Should return 401 for invalid token",
        },
        {
            name:     "expired JWT token",
            endpoint: "/api/v1/requisitions",
            method:   "GET",
            headers: map[string]string{
                "Authorization": "Bearer " + generateExpiredToken(),
            },
            expectedStatus: 401,
            description:    "Should return 401 for expired token",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, tt.endpoint, nil)

            // Set headers
            for key, value := range tt.headers {
                req.Header.Set(key, value)
            }

            resp, _ := app.Test(req)

            assert.Equal(t, tt.expectedStatus, resp.StatusCode, tt.description)
        })
    }
}
```

## Continuous Integration

### CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: liyali_gateway_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.21

      - name: Install dependencies
        run: go mod download

      - name: Run database migrations
        run: |
          # Bootstrap system handles migrations automatically
          # Just verify database is accessible
          PGPASSWORD=postgres psql -h localhost -U postgres -d liyali_gateway_test -c "SELECT 1;"

      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./tests/unit/...
        env:
          TEST_DB_URL: postgres://postgres:postgres@localhost/liyali_gateway_test?sslmode=disable

      - name: Run bootstrap system tests
        run: go test -v -race ./tests/integration/bootstrap/...
        env:
          TEST_DB_URL: postgres://postgres:postgres@localhost/liyali_gateway_test?sslmode=disable

      - name: Run integration tests
        run: go test -v -race ./tests/integration/...
        env:
          TEST_DB_URL: postgres://postgres:postgres@localhost/liyali_gateway_test?sslmode=disable
```

## Best Practices

### Test Organization

1. **Separate test types** - Keep unit, integration, and performance tests separate
2. **Use descriptive test names** - Test names should describe what is being tested
3. **Mock external dependencies** - Use mocks for external services and databases in unit tests
4. **Test data management** - Use factories and builders for test data creation
5. **Cleanup after tests** - Always clean up resources and test data

### Test Quality

1. **Test both happy and error paths** - Include positive and negative test cases
2. **Test edge cases** - Test boundary conditions and edge cases
3. **Comprehensive assertions** - Verify all important aspects of the response
4. **Error testing** - Test error handling and error responses
5. **Performance awareness** - Include performance considerations in critical paths

### Running Tests

### Quick Start

```bash
# Navigate to backend directory
cd backend

# Run all tests (comprehensive test suite)
./scripts/run_tests.sh

# Run specific test modules
./scripts/run_tests.sh auth          # Authentication tests only
./scripts/run_tests.sh workflows     # Workflow tests only
./scripts/run_tests.sh documents     # Document tests only
./scripts/run_tests.sh unit          # Workflow unit tests only

# Run individual test scripts
./scripts/auth_tests.sh              # Authentication & session management
./scripts/rbac_tests.sh              # Role-based access control
./scripts/document_tests.sh          # Document management
./scripts/workflow_test.sh           # Comprehensive workflow testing
./scripts/workflow_unit_tests.sh     # Workflow unit tests (Go)
./scripts/custom_role_tests.sh       # Custom role workflows
./scripts/department_tests.sh        # Department management
./scripts/analytics_tests.sh         # Analytics & system operations
./scripts/error_tests.sh             # Error handling & edge cases
```

### Test Options

```bash
# Run workflow tests with specific options
./scripts/workflow_test.sh --unit-only       # Unit tests only
./scripts/workflow_test.sh --integration-only # Integration tests only
./scripts/workflow_test.sh --api-only        # API endpoint tests only
./scripts/workflow_test.sh --no-coverage     # Skip coverage analysis
./scripts/workflow_test.sh --no-performance  # Skip performance tests

# Run with verbose output
./scripts/run_tests.sh --verbose

# Get help for any script
./scripts/run_tests.sh --help
./scripts/workflow_test.sh --help
```

### Go Unit Tests

```bash
# Run Go unit tests directly
go test ./tests/unit/...
go test ./tests/integration/...
go test ./tests/...

# Run with coverage
go test -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out -o coverage.html

# Run specific test files
go test -v ./tests/unit/workflow_concurrency_fixes_test.go
go test -v ./tests/integration/workflow_api_integration_test.go

# Run with race detection
go test -race ./tests/...

# Run benchmarks
go test -bench=. ./tests/...
```

### Test Environment Setup

The test suite automatically handles environment setup through `common_tests.sh`:

```bash
# Environment variables (automatically configured)
export GO_ENV=test
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=liyali_test
export DB_USER=postgres
export DB_PASSWORD=password
export API_URL=http://localhost:8080/api/v1
```

### Authentication Context

The modular test suite maintains authentication context across test modules:

1. **Automatic Setup**: `run_tests.sh` runs authentication first if no context exists
2. **Context Persistence**: Authentication tokens and IDs are saved and reused
3. **Manual Authentication**: Run `./scripts/auth_tests.sh` to establish context
4. **Context Cleanup**: Authentication context is automatically managed

## Test Quality Metrics

### Current Coverage Areas

- ✅ **Authentication & Authorization** - Complete coverage
- ✅ **Multi-tenant Operations** - Complete coverage
- ✅ **Document Management** - Complete CRUD coverage
- ✅ **Workflow System** - Comprehensive coverage including:
  - Unit tests for concurrency fixes
  - Integration tests for API endpoints
  - Custom role validation
  - Performance benchmarks
  - Coverage analysis
- ✅ **Custom Role Workflows** - Complete coverage
- ✅ **Department Management** - Complete coverage
- ✅ **Analytics & Reporting** - Complete coverage
- ✅ **Error Handling** - Comprehensive edge case coverage

### Test Metrics

The test suite provides detailed metrics:

```bash
# Example output from workflow_test.sh
📊 Test Results:
   Total Tests Run: 45
   Unit Tests: 12
   Integration Tests: 18
   API Endpoint Tests: 15
   Passed: 43
   Failed: 2
   Test Coverage: 87.3%
   Performance: All benchmarks passed
```

## Troubleshooting Tests

### Common Issues

**Server Not Running**

```bash
# Start the development server
go run main.go

# Or use the Makefile
make dev

# Check if server is responding
curl http://localhost:8080/health
```

**Authentication Context Issues**

```bash
# Clear and re-establish authentication context
./scripts/auth_tests.sh

# Check current context
cat ~/.liyali_test_context

# Manual context cleanup
rm ~/.liyali_test_context
```

**Database Connection Issues**

```bash
# Check database status
psql -h localhost -U postgres -l

# Verify test database exists
createdb liyali_test

# Reset test database
dropdb liyali_test
createdb liyali_test
```

**Go Test Issues**

```bash
# Clean Go module cache
go clean -modcache
go mod download

# Run tests with verbose output
go test -v ./tests/...

# Check for race conditions
go test -race ./tests/...
```

**Permission Issues**

```bash
# Make scripts executable
chmod +x backend/scripts/*.sh

# Fix common permission issues
find backend/scripts -name "*.sh" -exec chmod +x {} \;
```

### Test Debugging

**Enable Debug Mode**

```bash
# Set debug environment variables
export DEBUG=true
export LOG_LEVEL=debug

# Run tests with debug output
./scripts/workflow_test.sh --verbose
```

**Check Test Logs**

```bash
# View test output logs
cat backend/scripts/test_output.log

# View coverage reports
open backend/scripts/coverage.html
```

**Isolate Failing Tests**

```bash
# Run specific test categories
./scripts/workflow_test.sh --unit-only
./scripts/workflow_test.sh --integration-only

# Run individual Go tests
go test -v -run TestSpecificFunction ./tests/unit/...
```

### Performance Issues

**Slow Tests**

```bash
# Skip performance tests
./scripts/workflow_test.sh --no-performance

# Run with timeout
go test -timeout 30s ./tests/...

# Profile test execution
go test -cpuprofile cpu.prof ./tests/...
```

**Memory Issues**

```bash
# Run with memory profiling
go test -memprofile mem.prof ./tests/...

# Check for memory leaks
go test -race -count=10 ./tests/...
```

For more troubleshooting, see [Troubleshooting Guide](./16-troubleshooting.md).

## Next Steps

- **API Reference**: Explore [Complete API Documentation](./13-api-reference.md)
- **Deployment**: Prepare for [Production Deployment](./14-deployment.md)
- **Monitoring**: Set up [Performance Monitoring](./15-monitoring.md)
- **Troubleshooting**: Review [Common Issues](./16-troubleshooting.md)
