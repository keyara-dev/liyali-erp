# System Architecture

Comprehensive overview of the Liyali Gateway Backend architecture, design patterns, system components, and the new production-ready bootstrap system.

## Architecture Overview

The Liyali Gateway Backend follows **Clean Architecture** principles with a multi-layered approach that ensures separation of concerns, testability, and maintainability. It now includes an advanced bootstrap system for reliable database initialization.

```
┌─────────────────────────────────────────────────────────────────┐
│                     Bootstrap System                           │
├─────────────────────────────────────────────────────────────────┤
│  Phase Control - Connect → Validate → Migrate → Verify → Seed │
├─────────────────────────────────────────────────────────────────┤
│                        HTTP Layer                               │
├─────────────────────────────────────────────────────────────────┤
│  Handlers (Controllers) - HTTP request/response handling       │
├─────────────────────────────────────────────────────────────────┤
│  Middleware - Auth, CORS, Logging, Rate Limiting              │
├─────────────────────────────────────────────────────────────────┤
│                     Business Layer                             │
├─────────────────────────────────────────────────────────────────┤
│  Services - Business logic, validation, orchestration         │
├─────────────────────────────────────────────────────────────────┤
│                      Data Layer                                │
├─────────────────────────────────────────────────────────────────┤
│  Repositories - Data access abstraction                       │
├─────────────────────────────────────────────────────────────────┤
│  Models - Domain entities and data structures                 │
├─────────────────────────────────────────────────────────────────┤
│                    Database Layer                              │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL - Primary database with triggers + Bootstrap      │
└─────────────────────────────────────────────────────────────────┘
```

## Bootstrap System Architecture

### Overview

The bootstrap system is a production-ready initialization layer that ensures proper database setup and eliminates race conditions between migrations and seeding.

### Bootstrap Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Bootstrapper  │────│    Validator    │────│     Seeder      │
│                 │    │                 │    │                 │
│ • Phase Control │    │ • Schema Check  │    │ • UPSERT Ops    │
│ • Error Handling│    │ • Constraint    │    │ • Transactions  │
│ • Metrics       │    │ • Index Verify  │    │ • Dependency    │
│ • Health Checks │    │ • Trigger Check │    │ • Idempotency   │
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
                    │ • Retry Logic   │
                    └─────────────────┘
```

### Bootstrap Flow

```
Application Start
       │
       ▼
┌─────────────┐
│   Connect   │ ──► Validate DB connection and pool health
└─────────────┘
       │
       ▼
┌─────────────┐
│  Validate   │ ──► Check database readiness and PostgreSQL version
└─────────────┘
       │
       ▼
┌─────────────┐
│   Migrate   │ ──► Verify all required tables exist
└─────────────┘
       │
       ▼
┌─────────────┐
│   Verify    │ ──► Comprehensive schema integrity checks
└─────────────┘
       │
       ▼
┌─────────────┐
│    Seed     │ ──► Idempotent data seeding with UPSERT
└─────────────┘
       │
       ▼
┌─────────────┐
│  Complete   │ ──► Application ready to serve requests
└─────────────┘
```

### Key Bootstrap Features

- **Race Condition Prevention**: Strict phase ordering ensures migrations before seeding
- **Idempotent Operations**: PostgreSQL UPSERT operations prevent duplicate data
- **Circuit Breaker Protection**: Prevents cascading failures during startup
- **Comprehensive Validation**: Table, column, constraint, and index verification
- **Production Observability**: Detailed logging, timing, and metrics collection
- **Zero-Downtime Ready**: Health checks for Kubernetes deployments

## Core Architectural Principles

### 1. Clean Architecture
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Single Responsibility**: Each component has one reason to change
- **Interface Segregation**: Clients depend only on interfaces they use
- **Separation of Concerns**: Clear boundaries between layers

### 2. Multi-Tenant Architecture
- **Organization Isolation**: Complete data separation between organizations
- **Tenant Context**: Automatic tenant resolution from authentication
- **Scalable Design**: Supports unlimited organizations

### 3. Hybrid Database Approach
- **GORM**: Rich ORM for complex business operations
- **sqlc**: Type-safe queries for performance-critical operations
- **Database Triggers**: Automatic data synchronization

## System Components

### HTTP Layer

#### Fiber Web Framework
```go
// High-performance HTTP framework
app := fiber.New(fiber.Config{
    AppName:      "Liyali Gateway Backend",
    ErrorHandler: customErrorHandler,
})
```

**Features:**
- High performance (Express.js-like API)
- Built-in middleware support
- JSON serialization/deserialization
- Route grouping and parameter binding

#### Middleware Stack
```go
// Middleware execution order
app.Use(middleware.ErrorHandlingMiddleware())  // Error handling
app.Use(middleware.LoggerMiddleware())         // Request logging
app.Use(middleware.CORSMiddleware())           // CORS handling
app.Use(middleware.AuthMiddleware())           // Authentication
app.Use(middleware.TenantMiddleware())         // Multi-tenancy
app.Use(middleware.RequirePermission())       // Authorization
```

### Business Layer

#### Service Architecture
```go
type ServiceInterface interface {
    Create(ctx context.Context, data CreateRequest) (*Model, error)
    GetByID(ctx context.Context, id string) (*Model, error)
    Update(ctx context.Context, id string, data UpdateRequest) (*Model, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]*Model, int64, error)
}
```

**Service Responsibilities:**
- Business logic implementation
- Data validation and transformation
- Cross-cutting concerns (audit, notifications)
- Transaction management
- Error handling and logging

#### Handler Registry Pattern
```go
type HandlerRegistry struct {
    Auth     *AuthHandler
    Approval *ApprovalHandler
    Workflow *WorkflowHandler
    Document *DocumentHandler
}

func NewHandlerRegistry(services...) *HandlerRegistry {
    return &HandlerRegistry{
        Auth:     NewAuthHandler(authService, rbacService),
        Approval: NewApprovalHandler(),
        Workflow: NewWorkflowHandler(workflowService),
        Document: NewDocumentHandler(documentService),
    }
}
```

### Data Layer

#### Repository Pattern
```go
type RepositoryInterface interface {
    Create(ctx context.Context, model *Model) error
    GetByID(ctx context.Context, id string) (*Model, error)
    Update(ctx context.Context, model *Model) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]*Model, error)
}
```

**Repository Responsibilities:**
- Data access abstraction
- Query optimization
- Database transaction handling
- Error translation

#### Hybrid Database Strategy

**GORM for Business Operations:**
```go
// Rich ORM with relationships
type Requisition struct {
    ID           string        `gorm:"primaryKey"`
    Organization *Organization `gorm:"foreignKey:OrganizationID"`
    Items        datatypes.JSONType[[]RequisitionItem] `gorm:"type:jsonb"`
    // ... other fields
}
```

**sqlc for Performance-Critical Operations:**
```sql
-- name: GetUserSessions :many
SELECT * FROM user_sessions 
WHERE user_id = $1 AND expires_at > NOW();

-- name: CreateSession :one
INSERT INTO user_sessions (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;
```

## Multi-Tenant Architecture

### Tenant Isolation Strategy

#### 1. Organization-Based Isolation
```go
// Every entity belongs to an organization
type BaseModel struct {
    ID             string        `gorm:"primaryKey"`
    OrganizationID string        `gorm:"not null;index"`
    Organization   *Organization `gorm:"foreignKey:OrganizationID"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

#### 2. Automatic Tenant Context
```go
// Middleware injects tenant context
func TenantMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        orgID := getUserOrganization(userID)
        c.Locals("organization_id", orgID)
        return c.Next()
    }
}
```

#### 3. Query-Level Isolation
```go
// All queries automatically scoped to organization
func (r *Repository) List(ctx context.Context, orgID string) ([]*Model, error) {
    var models []*Model
    return models, r.db.Where("organization_id = ?", orgID).Find(&models).Error
}
```

### Multi-Organization User Support
```go
type User struct {
    ID                    string         `gorm:"primaryKey"`
    Email                 string         `gorm:"uniqueIndex"`
    CurrentOrganizationID *string        `json:"currentOrganizationId"`
    Organizations         []Organization `gorm:"many2many:user_organizations"`
}
```

## Authentication & Authorization Architecture

### JWT + Session Hybrid Approach

#### 1. JWT for Stateless Authentication
```go
type JWTClaims struct {
    UserID         string `json:"user_id"`
    OrganizationID string `json:"organization_id"`
    SessionID      string `json:"session_id"`
    jwt.RegisteredClaims
}
```

#### 2. Sessions for Security
```go
type UserSession struct {
    ID        string    `gorm:"primaryKey"`
    UserID    string    `gorm:"not null;index"`
    Token     string    `gorm:"not null;unique"`
    ExpiresAt time.Time `gorm:"not null"`
    IsActive  bool      `gorm:"default:true"`
}
```

### RBAC System Architecture

#### 1. Permission-Based Authorization
```go
type Permission struct {
    ID       string `gorm:"primaryKey"`
    Resource string `gorm:"not null"` // requisition, budget, etc.
    Action   string `gorm:"not null"` // view, create, edit, delete
    Name     string `gorm:"not null"` // requisition:view
}
```

#### 2. Custom Organization Roles
```go
type OrganizationRole struct {
    ID             string                    `gorm:"primaryKey"`
    OrganizationID string                    `gorm:"not null"`
    Name           string                    `gorm:"not null"`
    Permissions    []OrganizationPermission  `gorm:"foreignKey:RoleID"`
}
```

#### 3. Dynamic Permission Checking
```go
func RequirePermission(rbacService *RBACService, resource, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        orgID := c.Locals("organization_id").(string)
        
        hasPermission := rbacService.HasPermission(userID, orgID, resource, action)
        if !hasPermission {
            return c.Status(403).JSON(fiber.Map{"error": "Insufficient permissions"})
        }
        
        return c.Next()
    }
}
```

## Document Management Architecture

### Dual Document System

#### 1. Specific Document Models
```go
// Type-safe, rich domain models
type Requisition struct {
    ID          string                 `gorm:"primaryKey"`
    Title       string                 `gorm:"not null"`
    Items       []RequisitionItem      `gorm:"type:jsonb"`
    Priority    string                 `gorm:"not null"`
    Status      string                 `gorm:"not null"`
    // ... business-specific fields
}
```

#### 2. Generic Document Model
```go
// Unified model for search and analytics
type Document struct {
    ID           uuid.UUID      `gorm:"primaryKey"`
    DocumentType string         `gorm:"not null;index"`
    Title        string         `gorm:"not null"`
    Status       string         `gorm:"not null;index"`
    Data         datatypes.JSON `gorm:"type:jsonb"`
    // ... common fields
}
```

### Automatic Synchronization

#### Database Triggers for Data Consistency
```sql
-- Automatic sync on any change
CREATE TRIGGER trigger_sync_requisition
    AFTER INSERT OR UPDATE ON requisitions
    FOR EACH ROW
    EXECUTE FUNCTION sync_requisition_to_document();
```

**Benefits:**
- **100% Data Consistency** - No manual sync required
- **Real-time Updates** - Changes reflected immediately
- **Zero Code Changes** - Works with existing application
- **Performance Optimized** - Minimal overhead

## Workflow Engine Architecture

### Dynamic Workflow System

#### 1. Configurable Workflows
```go
type Workflow struct {
    ID           uuid.UUID      `gorm:"primaryKey"`
    Name         string         `gorm:"not null"`
    DocumentType string         `gorm:"not null"`
    Stages       datatypes.JSON `gorm:"type:jsonb"`
    IsActive     bool           `gorm:"default:false"`
}
```

#### 2. Workflow Stages
```go
type WorkflowStage struct {
    ID                string   `json:"id"`
    Name              string   `json:"name"`
    Approvers         []string `json:"approvers"`
    RequiredApprovals int      `json:"requiredApprovals"`
    Order             int      `json:"order"`
}
```

#### 3. Approval Task Management
```go
type ApprovalTask struct {
    ID           string    `gorm:"primaryKey"`
    DocumentID   string    `gorm:"not null;index"`
    DocumentType string    `gorm:"not null"`
    AssignedTo   string    `gorm:"not null"`
    Status       string    `gorm:"not null"`
    Stage        int       `gorm:"not null"`
}
```

## Performance Architecture

### Database Optimization

#### 1. Strategic Indexing
```sql
-- Performance-critical indexes
CREATE INDEX idx_documents_org_type ON documents(organization_id, document_type);
CREATE INDEX idx_documents_search ON documents USING GIN(to_tsvector('english', title || ' ' || description));
CREATE INDEX idx_approval_tasks_assigned ON approval_tasks(assigned_to, status);
```

#### 2. Connection Pooling
```go
// Optimized connection pool
db.SetMaxIdleConns(10)
db.SetMaxOpenConns(100)
db.SetConnMaxLifetime(time.Hour)
```

#### 3. Query Optimization
```go
// Efficient pagination with counting
func (r *Repository) ListWithCount(ctx context.Context, filter Filter) ([]*Model, int64, error) {
    var models []*Model
    var count int64
    
    query := r.db.WithContext(ctx).Where(filter.ToWhere())
    
    // Count and fetch in parallel
    countQuery := query.Model(&Model{})
    dataQuery := query.Offset(filter.Offset).Limit(filter.Limit)
    
    err := countQuery.Count(&count).Error
    if err != nil {
        return nil, 0, err
    }
    
    err = dataQuery.Find(&models).Error
    return models, count, err
}
```

### Caching Strategy

#### 1. Application-Level Caching
```go
// Cache frequently accessed data
type CacheService struct {
    cache map[string]interface{}
    mutex sync.RWMutex
}

func (c *CacheService) GetOrganizationPermissions(orgID string) []Permission {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    if cached, exists := c.cache[orgID]; exists {
        return cached.([]Permission)
    }
    
    // Fetch from database and cache
    permissions := c.fetchFromDB(orgID)
    c.cache[orgID] = permissions
    return permissions
}
```

## Security Architecture

### Defense in Depth

#### 1. Input Validation
```go
// Request validation at handler level
func (h *Handler) CreateRequisition(c *fiber.Ctx) error {
    var req CreateRequisitionRequest
    if err := c.BodyParser(&req); err != nil {
        return utils.SendBadRequestError(c, "Invalid request body")
    }
    
    if err := h.validate.Struct(req); err != nil {
        return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
    }
    
    // Process request...
}
```

#### 2. SQL Injection Prevention
```go
// Parameterized queries with GORM
db.Where("organization_id = ? AND status = ?", orgID, status).Find(&models)

// Type-safe queries with sqlc
queries.GetUserByEmail(ctx, email)
```

#### 3. Authentication Layers
```go
// Multiple authentication checks
func AuthMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Extract JWT token
        token := extractToken(c)
        
        // 2. Validate JWT signature
        claims, err := validateJWT(token)
        if err != nil {
            return unauthorized(c)
        }
        
        // 3. Check session validity
        session, err := getSession(claims.SessionID)
        if err != nil || !session.IsActive {
            return unauthorized(c)
        }
        
        // 4. Set user context
        c.Locals("user_id", claims.UserID)
        c.Locals("organization_id", claims.OrganizationID)
        
        return c.Next()
    }
}
```

## Error Handling Architecture

### Centralized Error Management

#### 1. Custom Error Types
```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}
```

#### 2. Global Error Handler
```go
func customErrorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    message := "Internal Server Error"
    
    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
        message = e.Message
    }
    
    if e, ok := err.(*AppError); ok {
        return c.Status(code).JSON(fiber.Map{
            "error": e.Code,
            "message": e.Message,
            "details": e.Details,
        })
    }
    
    return c.Status(code).JSON(fiber.Map{
        "error": message,
    })
}
```

#### 3. Consistent Response Format
```go
// Standardized response helpers
func SendSuccess(c *fiber.Ctx, data interface{}, message string) error {
    return c.JSON(types.SuccessResponse{
        Success: true,
        Message: message,
        Data:    data,
    })
}

func SendError(c *fiber.Ctx, code int, message string) error {
    return c.Status(code).JSON(types.ErrorResponse{
        Success: false,
        Error:   message,
    })
}
```

## Monitoring & Observability

### Logging Architecture
```go
// Structured logging
logger := log.WithFields(log.Fields{
    "user_id":         userID,
    "organization_id": orgID,
    "operation":       "create_requisition",
    "request_id":      requestID,
})

logger.Info("Processing requisition creation")
```

### Health Check System
```go
func HealthCheck(c *fiber.Ctx) error {
    // Check database connectivity
    if err := db.Ping(); err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy",
            "error":  "database connection failed",
        })
    }
    
    return c.JSON(fiber.Map{
        "status":  "healthy",
        "service": "liyali-gateway-backend",
        "version": version,
        "uptime":  time.Since(startTime).String(),
    })
}
```

## Deployment Architecture

### Container Strategy
```dockerfile
# Multi-stage build for optimization
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-w -s" -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Scalability Considerations

#### 1. Stateless Design
- No server-side session storage
- JWT-based authentication
- Database-backed sessions for security

#### 2. Horizontal Scaling
- Load balancer compatible
- Database connection pooling
- Shared nothing architecture

#### 3. Database Scaling
- Read replicas support
- Connection pooling
- Query optimization

## Next Steps

- **Database Design**: Review [Database Schema](./05-database.md)
- **API Design**: Understand [API Patterns](./06-api-design.md)
- **Authentication**: Deep dive into [Auth System](./07-auth.md)
- **Development**: Set up [Development Environment](./11-development.md)