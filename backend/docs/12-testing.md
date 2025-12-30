# Testing Guide

Comprehensive testing strategy for the Liyali Gateway Backend.

## Testing Overview

The backend implements a multi-layered testing strategy:

- **Unit Tests** - Test individual components in isolation
- **Integration Tests** - Test component interactions and database operations
- **API Tests** - Test HTTP endpoints and request/response handling
- **Performance Tests** - Test system performance under load
- **Security Tests** - Test authentication, authorization, and security measures

## Test Structure

### Directory Organization

```
tests/
├── unit/                  # Unit tests
│   ├── handlers/         # Handler unit tests
│   ├── services/         # Service unit tests
│   ├── repository/       # Repository unit tests
│   └── utils/           # Utility unit tests
├── integration/          # Integration tests
│   ├── api/             # API integration tests
│   ├── database/        # Database integration tests
│   └── helpers.go       # Integration test helpers
└── performance/          # Performance tests
    └── load/            # Load testing
```

### Test Configuration

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
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

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

```bash
# Run all tests (including bootstrap tests)
make test

# Run unit tests only
go test ./tests/unit/...

# Run bootstrap system tests
go test ./tests/integration/bootstrap/...

# Run integration tests only
go test ./tests/integration/...

# Run tests with coverage
go test -cover ./...

# Run performance tests
make test-performance

# Run security tests
go test ./tests/security/...

# Run bootstrap benchmarks
go test -bench=. ./tests/performance/bootstrap/...
```

## Troubleshooting Tests

### Common Issues

**Database Connection Issues**
```bash
# Check database status
psql -h localhost -U postgres -l

# Verify test database exists
createdb liyali_gateway_test

# Reset test database (bootstrap will handle setup)
dropdb liyali_gateway_test
createdb liyali_gateway_test
```

**Bootstrap System Issues**
```bash
# Test bootstrap system directly
go test ./tests/integration/bootstrap/... -v

# Check bootstrap configuration
export TEST_DB_URL="postgres://postgres:postgres@localhost/liyali_gateway_test?sslmode=disable"
go test ./tests/integration/bootstrap/... -v -run TestBootstrapSystem_FullCycle

# Debug bootstrap failures
export LOG_LEVEL=debug
go test ./tests/integration/bootstrap/... -v
```

**Flaky Tests**
```bash
# Run tests multiple times to identify flaky tests
go test -count=10 ./tests/integration/...

# Use race detector
go test -race ./...
```

**Mock Issues**
```bash
# Regenerate mocks
go generate ./...

# Verify mock implementations
go test -v ./tests/unit/...
```

For more troubleshooting, see [Troubleshooting Guide](./16-troubleshooting.md).

## Next Steps

- **API Reference**: Explore [Complete API Documentation](./13-api-reference.md)
- **Deployment**: Prepare for [Production Deployment](./14-deployment.md)
- **Monitoring**: Set up [Performance Monitoring](./15-monitoring.md)
- **Troubleshooting**: Review [Common Issues](./16-troubleshooting.md)