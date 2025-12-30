# Database Bootstrap System

A production-ready database bootstrap system that solves the race condition between database migrations and seeding operations. This system ensures proper initialization order, implements defensive programming patterns, and provides comprehensive observability.

## 🚀 Features

- **Proper Phase Ordering**: Connect → Validate → Migrate → Verify → Seed
- **Idempotent Operations**: Safe to run multiple times using PostgreSQL UPSERT
- **Circuit Breaker Pattern**: Prevents cascading failures during bootstrap
- **Retry Logic**: Exponential backoff with jitter for transient failures
- **Transaction Safety**: Atomic seeding operations with rollback support
- **Comprehensive Validation**: Schema integrity checks and constraint verification
- **Zero-Downtime Ready**: Health checks and metrics for deployment orchestration
- **Production Observability**: Detailed logging, timing, and metrics collection

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Bootstrapper  │────│    Validator    │────│     Seeder      │
│                 │    │                 │    │                 │
│ • Phase Control │    │ • Schema Check  │    │ • UPSERT Ops    │
│ • Error Handling│    │ • Constraint    │    │ • Transactions  │
│ • Metrics       │    │ • Index Verify  │    │ • Dependency    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ Circuit Breaker │
                    │                 │
                    │ • Failure Track │
                    │ • Auto Recovery │
                    │ • State Machine │
                    └─────────────────┘
```

## 📋 Bootstrap Phases

### Phase 1: Connect
- Validates database connection health
- Checks connection pool statistics
- Verifies PostgreSQL version compatibility

### Phase 2: Validate
- Ensures database is accessible
- Checks basic schema readiness
- Validates PostgreSQL extensions

### Phase 3: Migrate
- Verifies all required tables exist
- Checks table structures and columns
- Validates migration completeness

### Phase 4: Verify
- Comprehensive schema integrity check
- Foreign key constraint validation
- Index existence verification
- Trigger function validation

### Phase 5: Seed
- Idempotent data seeding with UPSERT
- Transactional safety with rollback
- Dependency-ordered seeding
- Circuit breaker protection

## 🔧 Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/liyali/liyali-gateway/bootstrap"
    "gorm.io/gorm"
)

func main() {
    // Initialize database connection
    db := initializeDatabase() // Your DB setup
    
    // Create bootstrap configuration
    config := bootstrap.DefaultBootstrapConfig()
    config.Environment = "production"
    config.SkipSeeding = false // Enable seeding
    
    // Create bootstrapper
    logger := log.Default()
    bootstrapper := bootstrap.NewBootstrapper(db, config, logger)
    
    // Run bootstrap
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
    defer cancel()
    
    result := bootstrapper.Bootstrap(ctx)
    
    if !result.Success {
        log.Fatalf("Bootstrap failed at phase %s: %v", result.Phase, result.Error)
    }
    
    log.Printf("Bootstrap completed in %v", result.Duration)
}
```

### Advanced Configuration

```go
config := &bootstrap.BootstrapConfig{
    Environment:        "production",
    SkipSeeding:       false,
    SeedRetryAttempts: 5,
    SeedRetryDelay:    time.Second * 5,
    CircuitBreakerConfig: circuit.Config{
        MaxFailures: 3,
        Timeout:     time.Minute,
        Interval:    time.Minute * 2,
    },
    ValidationTimeout: time.Minute,
    MigrationTimeout:  time.Minute * 10,
}
```

### Health Checks

```go
// Use in Kubernetes readiness probe
func healthHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
    defer cancel()
    
    if err := bootstrapper.HealthCheck(ctx); err != nil {
        http.Error(w, "Database not ready", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### Metrics Collection

```go
// Expose metrics for monitoring
func metricsHandler(w http.ResponseWriter, r *http.Request) {
    metrics := bootstrapper.GetMetrics()
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "database": metrics,
        "timestamp": time.Now(),
    })
}
```

## 🔄 Idempotent Seeding

The seeding system uses PostgreSQL's `ON CONFLICT` clause for true idempotency:

```go
// Example: User seeding with UPSERT
err := tx.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "email"}},
    DoUpdates: clause.AssignmentColumns([]string{"name", "role", "updated_at"}),
}).Create(&user).Error
```

### Seeding Features

- **Dependency Ordering**: Seeds in correct dependency order
- **Transaction Safety**: Each entity type seeded in its own transaction
- **Conflict Resolution**: Updates existing records, creates new ones
- **Comprehensive Logging**: Detailed statistics for each seeding operation
- **Error Recovery**: Continues seeding other entities if one fails

## 🛡️ Circuit Breaker

Protects against cascading failures during bootstrap:

```go
type Config struct {
    MaxFailures int           // Failures before opening circuit
    Timeout     time.Duration // Time before attempting to close
    Interval    time.Duration // Check interval for closing
}
```

### Circuit States

- **CLOSED**: Normal operation, requests pass through
- **OPEN**: Circuit is open, requests fail immediately
- **HALF_OPEN**: Testing if service has recovered

## 🔄 Retry Logic

Implements exponential backoff with jitter:

```go
// Retry with exponential backoff
err := retry.WithExponentialBackoff(
    ctx,
    maxAttempts,
    baseDelay,
    func() error {
        return seedOperation()
    },
)
```

### Retry Features

- **Exponential Backoff**: Delays increase exponentially
- **Jitter**: Random variation to prevent thundering herd
- **Context Cancellation**: Respects context timeouts
- **Error Classification**: Distinguishes retryable vs non-retryable errors

## 📊 Observability

### Logging

Structured logging with phase timing:

```
[BOOTSTRAP] 🚀 Starting database bootstrap process (env: production)
[BOOTSTRAP] 📋 Phase: connect - Starting
[BOOTSTRAP] ✅ Phase: connect - Completed in 45ms
[BOOTSTRAP] 📋 Phase: validate - Starting
[BOOTSTRAP] ✅ Phase: validate - Completed in 123ms
[BOOTSTRAP] 🌱 Seeding users: 4 created, 0 updated, 0 skipped (took 67ms)
[BOOTSTRAP] ✅ Database bootstrap completed successfully in 2.3s
```

### Metrics

Comprehensive metrics collection:

```json
{
  "db_connections_open": 5,
  "db_connections_in_use": 2,
  "db_connections_idle": 3,
  "circuit_breaker_state": "CLOSED",
  "circuit_breaker_failures": 0,
  "total_duration_ms": 2300
}
```

## 🧪 Testing

### Integration Tests

```bash
# Run integration tests
go test ./bootstrap/... -v

# Run with coverage
go test ./bootstrap/... -cover -coverprofile=coverage.out

# Run benchmarks
go test ./bootstrap/... -bench=. -benchmem
```

### Test Database Setup

```bash
# Create test database
createdb liyali_gateway_test

# Set test environment variables
export TEST_DB_NAME=liyali_gateway_test
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
```

## 🚀 Production Deployment

### Docker Integration

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bootstrap ./bootstrap/example

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bootstrap .
CMD ["./bootstrap"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: liyali-gateway
spec:
  template:
    spec:
      containers:
      - name: app
        image: liyali-gateway:latest
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

### Environment Configuration

```bash
# Production environment variables
APP_ENV=production
ENABLE_SEEDING=false  # Disable seeding in production
DB_HOST=postgres.example.com
DB_PORT=5432
DB_USER=app_user
DB_PASSWORD=secure_password
DB_NAME=liyali_gateway
DB_SSL_MODE=require
```

## 🔧 Troubleshooting

### Common Issues

1. **Tables Don't Exist**
   ```
   Error: missing required tables: [users, organizations]
   Solution: Run migrations first using cd database && ./migrate.sh up
   ```

2. **Connection Pool Exhausted**
   ```
   Error: connection pool exhausted
   Solution: Increase MaxOpenConns or check for connection leaks
   ```

3. **Circuit Breaker Open**
   ```
   Error: circuit breaker is open
   Solution: Check database connectivity and wait for auto-recovery
   ```

### Debug Mode

Enable debug logging:

```go
config := bootstrap.DefaultBootstrapConfig()
config.Environment = "development" // Enables verbose logging
```

### Manual Recovery

```go
// Reset circuit breaker manually
bootstrapper.GetCircuitBreaker().Reset()

// Force re-seeding
config.SkipSeeding = false
result := bootstrapper.Bootstrap(ctx)
```

## 📚 API Reference

### Bootstrapper

```go
type Bootstrapper struct {
    // Methods
    Bootstrap(ctx context.Context) *BootstrapResult
    HealthCheck(ctx context.Context) error
    GetMetrics() map[string]interface{}
}
```

### Configuration

```go
type BootstrapConfig struct {
    Environment          string
    SkipSeeding         bool
    SeedRetryAttempts   int
    SeedRetryDelay      time.Duration
    CircuitBreakerConfig circuit.Config
    ValidationTimeout   time.Duration
    MigrationTimeout    time.Duration
}
```

### Results

```go
type BootstrapResult struct {
    Success  bool
    Phase    BootstrapPhase
    Duration time.Duration
    Error    error
    Metrics  map[string]interface{}
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.