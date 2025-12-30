# Development Guide

Complete guide for local development setup and best practices.

## Development Environment Setup

### Prerequisites

- **Go 1.21+** - Latest stable version
- **PostgreSQL 14+** - Database server
- **Git** - Version control
- **VS Code** (recommended) - IDE with Go extensions
- **Docker** (optional) - For containerized development

### IDE Setup

#### VS Code Extensions

Install these essential extensions:

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next",
    "formulahendry.auto-rename-tag",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-eslint"
  ]
}
```

#### VS Code Settings

Create `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v", "-race"],
  "go.buildFlags": ["-race"],
  "go.vetFlags": ["-all"],
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "files.exclude": {
    "**/.git": true,
    "**/node_modules": true,
    "**/*.exe": true
  }
}
```

### Local Development Setup

#### 1. Clone Repository

```bash
git clone <repository-url>
cd liyali-gateway/backend
```

#### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

#### 3. Database Setup with Bootstrap System

The new bootstrap system handles database initialization automatically:

```bash
# Create development database
createdb liyali_gateway_dev

# Option 1: Use migration scripts (Recommended)
cd database && ./migrate.sh up

# Option 2: Manual migration
go run database/run_migration.go database/migrations/001_create_complete_schema.up.sql

# The bootstrap system will automatically:
# - Validate database connection
# - Verify schema integrity  
# - Seed initial data (in development)
```

#### Bootstrap Development Features

- **Automatic Seeding**: Creates test users, organizations, and vendors
- **Idempotent Operations**: Safe to run multiple times
- **Comprehensive Validation**: Ensures database is properly configured
- **Detailed Logging**: Shows progress and timing for each phase

#### Bootstrap Logs in Development

```
[BOOTSTRAP] 🚀 Starting database bootstrap process (env: development)
[BOOTSTRAP] ✅ Phase: connect - Completed in 45ms
[BOOTSTRAP] ✅ Phase: validate - Completed in 123ms
[BOOTSTRAP] ✅ Phase: migrate - Completed in 67ms
[BOOTSTRAP] ✅ Phase: verify - Completed in 89ms
[BOOTSTRAP] 🌱 Seeding users: 4 created, 0 updated, 0 skipped (took 67ms)
[BOOTSTRAP] 🌱 Seeding organizations: 2 created, 0 updated, 0 skipped (took 34ms)
[BOOTSTRAP] 🌱 Seeding vendors: 3 created, 0 updated, 0 skipped (took 45ms)
[BOOTSTRAP] ✅ Database bootstrap completed successfully in 2.3s
```

#### 4. Environment Configuration

```bash
# Copy development environment
cp .env.example .env.development

# Edit configuration
nano .env.development

# Development environment variables
APP_ENV=development
ENABLE_SEEDING=true          # Enable automatic seeding
BOOTSTRAP_TIMEOUT=300        # Bootstrap timeout in seconds
CIRCUIT_BREAKER_ENABLED=true # Enable circuit breaker protection
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=liyali_gateway_dev
DB_SSL_MODE=disable
```

**Development Environment Variables:**
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=liyali_gateway_dev
DB_SSL_MODE=disable

# Application
APP_PORT=8080
APP_ENV=development
LOG_LEVEL=debug

# Security (development only)
JWT_SECRET=development-secret-key-not-for-production-use-minimum-32-chars

# Frontend
FRONTEND_URL=http://localhost:3000

# Development features
ENABLE_CORS=true
RATE_LIMIT_ENABLED=false
AUTO_MIGRATE=true
```

#### 5. Start Development Server

```bash
# Using air for hot reloading (recommended)
air

# Or run directly
go run main.go

# Or using make
make dev
```

## Project Structure

### Directory Layout

```
backend/
├── cmd/                    # Application entry points
│   └── server/            # Main server application
├── config/                # Configuration management
├── database/              # Database related files
│   ├── migrations/        # SQL migration files
│   └── queries/          # SQL query files
├── docs/                  # Documentation
├── handlers/              # HTTP handlers (controllers)
├── middleware/            # HTTP middleware
├── models/                # Data models
├── repository/            # Data access layer
├── routes/                # Route definitions
├── services/              # Business logic layer
├── tests/                 # Test files
│   ├── integration/       # Integration tests
│   └── unit/             # Unit tests
├── types/                 # Type definitions
├── utils/                 # Utility functions
├── main.go               # Application entry point
├── Makefile              # Build automation
├── go.mod                # Go module definition
└── go.sum                # Go module checksums
```

### Code Organization

#### Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│              Handlers                    │  ← HTTP Layer
├─────────────────────────────────────────┤
│              Services                    │  ← Business Logic
├─────────────────────────────────────────┤
│             Repository                   │  ← Data Access
├─────────────────────────────────────────┤
│              Models                      │  ← Data Models
└─────────────────────────────────────────┘
```

#### Handler Pattern

```go
// handlers/requisition_handler.go
type RequisitionHandler struct {
    service services.RequisitionService
    logger  *logrus.Logger
}

func NewRequisitionHandler(service services.RequisitionService, logger *logrus.Logger) *RequisitionHandler {
    return &RequisitionHandler{
        service: service,
        logger:  logger,
    }
}

func (h *RequisitionHandler) CreateRequisition(c *fiber.Ctx) error {
    var req types.CreateRequisitionRequest
    if err := c.BodyParser(&req); err != nil {
        return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
    }
    
    userID := c.Locals("userID").(string)
    requisition, err := h.service.CreateRequisition(userID, &req)
    if err != nil {
        h.logger.WithError(err).Error("Failed to create requisition")
        return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create requisition", err)
    }
    
    return utils.SuccessResponse(c, fiber.StatusCreated, "Requisition created successfully", requisition)
}
```

#### Service Pattern

```go
// services/requisition_service.go
type RequisitionService interface {
    CreateRequisition(userID string, req *types.CreateRequisitionRequest) (*models.Requisition, error)
    GetRequisition(userID, reqID string) (*models.Requisition, error)
    UpdateRequisition(userID, reqID string, req *types.UpdateRequisitionRequest) (*models.Requisition, error)
    DeleteRequisition(userID, reqID string) error
    ListRequisitions(userID string, filters *types.RequisitionFilters) (*types.PaginatedResponse, error)
}

type requisitionService struct {
    repo       repository.RequisitionRepository
    userRepo   repository.UserRepository
    authService AuthService
    logger     *logrus.Logger
}

func NewRequisitionService(
    repo repository.RequisitionRepository,
    userRepo repository.UserRepository,
    authService AuthService,
    logger *logrus.Logger,
) RequisitionService {
    return &requisitionService{
        repo:        repo,
        userRepo:    userRepo,
        authService: authService,
        logger:      logger,
    }
}

func (s *requisitionService) CreateRequisition(userID string, req *types.CreateRequisitionRequest) (*models.Requisition, error) {
    // Validate permissions
    if !s.authService.HasPermission(userID, "requisitions.create") {
        return nil, ErrUnauthorized
    }
    
    // Get user for organization context
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        return nil, err
    }
    
    // Create requisition
    requisition := &models.Requisition{
        ID:             uuid.New().String(),
        OrganizationID: user.OrganizationID,
        Title:          req.Title,
        Description:    req.Description,
        CreatedBy:      userID,
        Status:         "draft",
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    
    return s.repo.Create(requisition)
}
```

#### Repository Pattern

```go
// repository/requisition_repository.go
type RequisitionRepository interface {
    Create(req *models.Requisition) (*models.Requisition, error)
    GetByID(id string) (*models.Requisition, error)
    GetByIDAndOrg(id, orgID string) (*models.Requisition, error)
    Update(req *models.Requisition) (*models.Requisition, error)
    Delete(id string) error
    List(orgID string, filters *types.RequisitionFilters) ([]*models.Requisition, int64, error)
}

type requisitionRepository struct {
    db *gorm.DB
}

func NewRequisitionRepository(db *gorm.DB) RequisitionRepository {
    return &requisitionRepository{db: db}
}

func (r *requisitionRepository) Create(req *models.Requisition) (*models.Requisition, error) {
    if err := r.db.Create(req).Error; err != nil {
        return nil, err
    }
    return req, nil
}

func (r *requisitionRepository) GetByIDAndOrg(id, orgID string) (*models.Requisition, error) {
    var req models.Requisition
    err := r.db.Where("id = ? AND organization_id = ?", id, orgID).First(&req).Error
    if err != nil {
        return nil, err
    }
    return &req, nil
}
```

## Development Workflow

### Git Workflow

#### Branch Naming Convention

```bash
# Feature branches
feature/add-user-authentication
feature/implement-search-api

# Bug fix branches
bugfix/fix-login-validation
bugfix/resolve-database-connection

# Hotfix branches
hotfix/security-patch-jwt
hotfix/critical-data-loss-fix

# Release branches
release/v1.2.0
release/v1.2.1
```

#### Commit Message Convention

```bash
# Format: type(scope): description

# Types:
feat(auth): add JWT token refresh functionality
fix(db): resolve connection pool exhaustion
docs(api): update authentication endpoints documentation
style(handlers): format code according to gofmt
refactor(services): extract common validation logic
test(integration): add workflow approval tests
chore(deps): update Go dependencies

# Examples:
git commit -m "feat(search): implement full-text search with PostgreSQL"
git commit -m "fix(auth): resolve session timeout issue"
git commit -m "docs(readme): update installation instructions"
```

### Development Commands

#### Makefile Commands

```makefile
# Development
.PHONY: dev
dev:
	air

.PHONY: build
build:
	go build -o bin/liyali-gateway-backend .

.PHONY: test
test:
	go test -v -race ./...

.PHONY: test-coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Database
.PHONY: migrate-up
migrate-up:
	psql -d $(DB_NAME) -f database/migrations/001_initial_schema.sql
	psql -d $(DB_NAME) -f database/migrations/002_enhanced_auth.sql
	psql -d $(DB_NAME) -f database/migrations/003_workflows.sql
	psql -d $(DB_NAME) -f database/migrations/008_create_documents_table.sql
	psql -d $(DB_NAME) -f database/migrations/009_add_document_sync_triggers.sql

.PHONY: migrate-down
migrate-down:
	psql -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

.PHONY: seed-data
seed-data:
	go run utils/seeddata.go

# Code Quality
.PHONY: lint
lint:
	golangci-lint run

.PHONY: format
format:
	gofmt -s -w .
	goimports -w .

.PHONY: vet
vet:
	go vet ./...

# Documentation
.PHONY: docs
docs:
	swag init -g main.go -o docs/swagger

# Clean
.PHONY: clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
```

### Code Quality

#### Linting Configuration

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m
  issues-exit-code: 1
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - varcheck
    - deadcode
    - structcheck
    - misspell
    - unconvert
    - gofmt
    - goimports
    - golint
    - gocritic
    - gocyclo
    - dupl

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
```

#### Pre-commit Hooks

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofmt
        language: system
        args: [-w]
        files: \.go$
      
      - id: go-imports
        name: go imports
        entry: goimports
        language: system
        args: [-w]
        files: \.go$
      
      - id: go-vet
        name: go vet
        entry: go vet
        language: system
        files: \.go$
        pass_filenames: false
      
      - id: go-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        files: \.go$
        pass_filenames: false
      
      - id: go-test
        name: go test
        entry: go test
        language: system
        args: [-v, -race, ./...]
        files: \.go$
        pass_filenames: false
```

## Testing

### Test Structure

```
tests/
├── integration/           # Integration tests
│   ├── auth_test.go      # Authentication integration tests
│   ├── document_test.go  # Document management tests
│   └── helpers.go        # Test helpers and utilities
└── unit/                 # Unit tests
    ├── handlers/         # Handler unit tests
    ├── services/         # Service unit tests
    ├── repository/       # Repository unit tests
    └── utils/           # Utility unit tests
```

### Unit Testing

#### Handler Tests

```go
// tests/unit/handlers/requisition_handler_test.go
func TestRequisitionHandler_CreateRequisition(t *testing.T) {
    // Setup
    mockService := &mocks.RequisitionService{}
    handler := handlers.NewRequisitionHandler(mockService, logrus.New())
    
    app := fiber.New()
    app.Post("/requisitions", handler.CreateRequisition)
    
    // Test data
    reqBody := types.CreateRequisitionRequest{
        Title:       "Test Requisition",
        Description: "Test Description",
        TotalAmount: 1000.00,
    }
    
    // Mock service response
    expectedReq := &models.Requisition{
        ID:          "test-id",
        Title:       reqBody.Title,
        Description: reqBody.Description,
        TotalAmount: reqBody.TotalAmount,
    }
    mockService.On("CreateRequisition", "user-id", &reqBody).Return(expectedReq, nil)
    
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

#### Service Tests

```go
// tests/unit/services/requisition_service_test.go
func TestRequisitionService_CreateRequisition(t *testing.T) {
    // Setup
    mockRepo := &mocks.RequisitionRepository{}
    mockUserRepo := &mocks.UserRepository{}
    mockAuthService := &mocks.AuthService{}
    
    service := services.NewRequisitionService(mockRepo, mockUserRepo, mockAuthService, logrus.New())
    
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
    }
    
    // Mock expectations
    mockAuthService.On("HasPermission", userID, "requisitions.create").Return(true)
    mockUserRepo.On("GetByID", userID).Return(user, nil)
    mockRepo.On("Create", mock.AnythingOfType("*models.Requisition")).Return(&models.Requisition{
        ID:             "new-id",
        OrganizationID: orgID,
        Title:          req.Title,
        Description:    req.Description,
        CreatedBy:      userID,
    }, nil)
    
    // Execute
    result, err := service.CreateRequisition(userID, req)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, req.Title, result.Title)
    assert.Equal(t, orgID, result.OrganizationID)
    assert.Equal(t, userID, result.CreatedBy)
    
    mockAuthService.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
}
```

### Integration Testing

#### Database Integration Tests

```go
// tests/integration/database_test.go
func TestDatabaseIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create repositories
    userRepo := repository.NewUserRepository(db)
    reqRepo := repository.NewRequisitionRepository(db)
    
    // Create test organization
    org := &models.Organization{
        ID:   "test-org",
        Name: "Test Organization",
    }
    db.Create(org)
    
    // Create test user
    user := &models.User{
        ID:             "test-user",
        OrganizationID: org.ID,
        Email:          "test@example.com",
        Name:           "Test User",
    }
    createdUser, err := userRepo.Create(user)
    assert.NoError(t, err)
    
    // Create test requisition
    req := &models.Requisition{
        ID:             "test-req",
        OrganizationID: org.ID,
        Title:          "Test Requisition",
        CreatedBy:      createdUser.ID,
    }
    createdReq, err := reqRepo.Create(req)
    assert.NoError(t, err)
    
    // Verify relationships
    foundReq, err := reqRepo.GetByIDAndOrg(createdReq.ID, org.ID)
    assert.NoError(t, err)
    assert.Equal(t, req.Title, foundReq.Title)
    assert.Equal(t, createdUser.ID, foundReq.CreatedBy)
}
```

#### API Integration Tests

```go
// tests/integration/api_test.go
func TestRequisitionAPI(t *testing.T) {
    // Setup test server
    app := setupTestApp(t)
    
    // Create test user and get JWT
    jwt := createTestUserAndGetJWT(t, app)
    
    // Test create requisition
    reqBody := map[string]interface{}{
        "title":       "Integration Test Requisition",
        "description": "Test requisition for integration testing",
        "totalAmount": 1500.00,
    }
    
    reqJSON, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/v1/requisitions", bytes.NewReader(reqJSON))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+jwt)
    
    resp, _ := app.Test(req)
    assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
    
    // Parse response
    var createResp struct {
        Success bool                `json:"success"`
        Data    models.Requisition `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&createResp)
    
    assert.True(t, createResp.Success)
    assert.Equal(t, reqBody["title"], createResp.Data.Title)
    
    // Test get requisition
    getReq := httptest.NewRequest("GET", "/api/v1/requisitions/"+createResp.Data.ID, nil)
    getReq.Header.Set("Authorization", "Bearer "+jwt)
    
    getResp, _ := app.Test(getReq)
    assert.Equal(t, fiber.StatusOK, getResp.StatusCode)
}
```

### Test Helpers

```go
// tests/integration/helpers.go
func setupTestDB(t *testing.T) *gorm.DB {
    // Create test database connection
    dsn := "host=localhost user=postgres password=password dbname=liyali_gateway_test port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    require.NoError(t, err)
    
    // Run migrations
    runTestMigrations(t, db)
    
    return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
    // Clean up test data
    db.Exec("TRUNCATE TABLE requisitions, users, organizations CASCADE")
}

func setupTestApp(t *testing.T) *fiber.App {
    // Setup test application with all middleware and routes
    app := fiber.New()
    
    // Setup test database
    db := setupTestDB(t)
    
    // Initialize services and handlers
    // ... setup code
    
    return app
}

func createTestUserAndGetJWT(t *testing.T, app *fiber.App) string {
    // Create test organization
    org := &models.Organization{
        ID:   uuid.New().String(),
        Name: "Test Organization",
    }
    
    // Create test user
    user := &models.User{
        ID:             uuid.New().String(),
        OrganizationID: org.ID,
        Email:          "test@example.com",
        Name:           "Test User",
        Password:       hashPassword("password123"),
    }
    
    // Register user via API
    registerBody := map[string]interface{}{
        "email":          user.Email,
        "password":       "password123",
        "name":           user.Name,
        "organizationId": org.ID,
    }
    
    // ... registration and login logic
    
    return jwt
}
```

## Debugging

### Logging

#### Structured Logging

```go
// utils/logger.go
func NewLogger() *logrus.Logger {
    logger := logrus.New()
    
    // Set log level based on environment
    if os.Getenv("APP_ENV") == "development" {
        logger.SetLevel(logrus.DebugLevel)
        logger.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
            ForceColors:   true,
        })
    } else {
        logger.SetLevel(logrus.InfoLevel)
        logger.SetFormatter(&logrus.JSONFormatter{})
    }
    
    return logger
}

// Usage in handlers
func (h *RequisitionHandler) CreateRequisition(c *fiber.Ctx) error {
    h.logger.WithFields(logrus.Fields{
        "userID":    c.Locals("userID"),
        "endpoint":  "CreateRequisition",
        "method":    c.Method(),
        "path":      c.Path(),
    }).Info("Creating new requisition")
    
    // ... handler logic
    
    h.logger.WithFields(logrus.Fields{
        "userID":       c.Locals("userID"),
        "requisitionID": result.ID,
    }).Info("Requisition created successfully")
    
    return utils.SuccessResponse(c, fiber.StatusCreated, "Requisition created", result)
}
```

#### Debug Middleware

```go
// middleware/debug.go
func DebugMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        if os.Getenv("APP_ENV") == "development" {
            start := time.Now()
            
            // Log request
            logrus.WithFields(logrus.Fields{
                "method": c.Method(),
                "path":   c.Path(),
                "ip":     c.IP(),
                "user":   c.Locals("userID"),
            }).Debug("Incoming request")
            
            // Continue processing
            err := c.Next()
            
            // Log response
            logrus.WithFields(logrus.Fields{
                "method":   c.Method(),
                "path":     c.Path(),
                "status":   c.Response().StatusCode(),
                "duration": time.Since(start),
            }).Debug("Request completed")
            
            return err
        }
        
        return c.Next()
    }
}
```

### Profiling

#### CPU Profiling

```go
// main.go
import _ "net/http/pprof"

func main() {
    if os.Getenv("APP_ENV") == "development" {
        // Start pprof server
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
    
    // ... rest of application
}
```

Access profiling endpoints:
- `http://localhost:6060/debug/pprof/` - Profile index
- `http://localhost:6060/debug/pprof/goroutine` - Goroutine profile
- `http://localhost:6060/debug/pprof/heap` - Memory heap profile
- `http://localhost:6060/debug/pprof/profile` - CPU profile

#### Memory Profiling

```bash
# Generate CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Generate memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Analyze profile
(pprof) top10
(pprof) list functionName
(pprof) web
```

## Performance Optimization

### Database Optimization

#### Connection Pooling

```go
// config/database.go
func setupDatabase() *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatal("Failed to get database instance:", err)
    }
    
    // Configure connection pool
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db
}
```

#### Query Optimization

```go
// Optimize queries with proper indexing and preloading
func (r *requisitionRepository) ListWithDetails(orgID string, filters *types.RequisitionFilters) ([]*models.Requisition, error) {
    query := r.db.Where("organization_id = ?", orgID)
    
    // Use indexes for filtering
    if filters.Status != "" {
        query = query.Where("status = ?", filters.Status)
    }
    
    if filters.CreatedBy != "" {
        query = query.Where("created_by = ?", filters.CreatedBy)
    }
    
    // Preload related data to avoid N+1 queries
    var requisitions []*models.Requisition
    err := query.Preload("CreatedByUser").
        Preload("Items").
        Preload("Approvals").
        Find(&requisitions).Error
    
    return requisitions, err
}
```

### Caching

#### Redis Caching

```go
// services/cache_service.go
type CacheService interface {
    Get(key string, dest interface{}) error
    Set(key string, value interface{}, expiration time.Duration) error
    Delete(key string) error
}

type redisCacheService struct {
    client *redis.Client
}

func NewRedisCacheService(addr, password string, db int) CacheService {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    
    return &redisCacheService{client: client}
}

func (c *redisCacheService) Get(key string, dest interface{}) error {
    val, err := c.client.Get(context.Background(), key).Result()
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(val), dest)
}

func (c *redisCacheService) Set(key string, value interface{}, expiration time.Duration) error {
    json, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return c.client.Set(context.Background(), key, json, expiration).Err()
}
```

#### Application-Level Caching

```go
// services/requisition_service.go
func (s *requisitionService) GetRequisition(userID, reqID string) (*models.Requisition, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("requisition:%s", reqID)
    var cached models.Requisition
    if err := s.cache.Get(cacheKey, &cached); err == nil {
        // Verify user has access
        if s.canAccessRequisition(userID, &cached) {
            return &cached, nil
        }
    }
    
    // Fetch from database
    req, err := s.repo.GetByID(reqID)
    if err != nil {
        return nil, err
    }
    
    // Verify access
    if !s.canAccessRequisition(userID, req) {
        return nil, ErrUnauthorized
    }
    
    // Cache for 5 minutes
    s.cache.Set(cacheKey, req, 5*time.Minute)
    
    return req, nil
}
```

## Security Best Practices

### Input Validation

```go
// utils/validation.go
func ValidateCreateRequisitionRequest(req *types.CreateRequisitionRequest) error {
    if strings.TrimSpace(req.Title) == "" {
        return errors.New("title is required")
    }
    
    if len(req.Title) > 255 {
        return errors.New("title must be less than 255 characters")
    }
    
    if req.TotalAmount < 0 {
        return errors.New("total amount must be positive")
    }
    
    if req.TotalAmount > 1000000 {
        return errors.New("total amount exceeds maximum limit")
    }
    
    // Validate items
    if len(req.Items) == 0 {
        return errors.New("at least one item is required")
    }
    
    for i, item := range req.Items {
        if err := validateRequisitionItem(&item); err != nil {
            return fmt.Errorf("item %d: %w", i+1, err)
        }
    }
    
    return nil
}
```

### SQL Injection Prevention

```go
// Always use parameterized queries
func (r *requisitionRepository) Search(orgID, query string) ([]*models.Requisition, error) {
    var requisitions []*models.Requisition
    
    // Safe parameterized query
    err := r.db.Where("organization_id = ? AND (title ILIKE ? OR description ILIKE ?)", 
        orgID, "%"+query+"%", "%"+query+"%").Find(&requisitions).Error
    
    return requisitions, err
}

// Use GORM's built-in protection
func (r *requisitionRepository) GetByStatus(orgID, status string) ([]*models.Requisition, error) {
    var requisitions []*models.Requisition
    
    // GORM automatically escapes parameters
    err := r.db.Where("organization_id = ? AND status = ?", orgID, status).Find(&requisitions).Error
    
    return requisitions, err
}
```

## Deployment Preparation

### Build Optimization

```bash
# Build for production with optimizations
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bin/liyali-gateway-backend .

# Build with version information
VERSION=$(git describe --tags --always --dirty)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
go build -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" -o bin/liyali-gateway-backend .
```

### Docker Development

```dockerfile
# Dockerfile.dev
FROM golang:1.21-alpine AS development

WORKDIR /app

# Install air for hot reloading
RUN go install github.com/cosmtrek/air@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Run with air for hot reloading
CMD ["air"]
```

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: liyali_gateway_dev
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data

  backend:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: password
      DB_NAME: liyali_gateway_dev
      APP_ENV: development
    volumes:
      - .:/app
    depends_on:
      - postgres

volumes:
  postgres_dev_data:
```

## Troubleshooting

### Common Development Issues

**Module Not Found**
```bash
# Clean module cache and re-download
go clean -modcache
go mod download
go mod tidy
```

**Database Connection Issues**
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Test connection
psql -h localhost -U postgres -d liyali_gateway_dev

# Check environment variables
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME
```

**Port Already in Use**
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
export APP_PORT=8081
```

**Hot Reload Not Working**
```bash
# Reinstall air
go install github.com/cosmtrek/air@latest

# Check .air.toml configuration
# Ensure file watching is properly configured
```

For more troubleshooting, see [Troubleshooting Guide](./16-troubleshooting.md).

## Next Steps

- **Testing**: Set up comprehensive [Testing Environment](./12-testing.md)
- **API Reference**: Explore [Complete API Documentation](./13-api-reference.md)
- **Deployment**: Prepare for [Production Deployment](./14-deployment.md)
- **Monitoring**: Implement [Performance Monitoring](./15-monitoring.md)
est

# Check air configuration
cat .air.toml

# Create air config if missing
air init
```

**Build Errors**
```bash
# Check Go version
go version

# Verify dependencies
go mod verify

# Clean build cache
go clean -cache
```

**Test Failures**
```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestSpecificFunction ./tests/unit/

# Check test database
psql -h localhost -U postgres -d liyali_gateway_test
```

### Debugging Tips

1. **Use structured logging** with correlation IDs
2. **Enable debug mode** in development environment
3. **Use pprof** for performance profiling
4. **Check database queries** with GORM debug mode
5. **Monitor goroutines** for potential leaks
6. **Use race detector** during testing

### Performance Monitoring

```bash
# Monitor application performance
go tool pprof http://localhost:6060/debug/pprof/profile

# Check memory usage
go tool pprof http://localhost:6060/debug/pprof/heap

# Monitor goroutines
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Best Practices Summary

### Code Quality
- Follow Go conventions and idioms
- Use meaningful variable and function names
- Write comprehensive tests
- Document public APIs
- Use linting tools consistently

### Security
- Validate all inputs
- Use parameterized queries
- Implement proper authentication and authorization
- Sanitize sensitive data in logs
- Keep dependencies updated

### Performance
- Use connection pooling
- Implement caching where appropriate
- Optimize database queries
- Monitor application metrics
- Profile critical paths

### Maintainability
- Follow clean architecture principles
- Keep functions small and focused
- Use dependency injection
- Write clear documentation
- Maintain consistent code style

## Next Steps

After setting up your development environment:

1. **Explore the codebase** - Familiarize yourself with the project structure
2. **Run tests** - Ensure all tests pass in your environment
3. **Make a small change** - Test the development workflow
4. **Read the documentation** - Review other documentation files
5. **Join the team** - Connect with other developers

For more information:
- [Testing Guide](./12-testing.md) - Learn about testing strategies
- [API Reference](./13-api-reference.md) - Explore the API endpoints
- [Deployment Guide](./14-deployment.md) - Understand deployment processes
- [Troubleshooting](./16-troubleshooting.md) - Common issues and solutions

Happy coding! 🚀