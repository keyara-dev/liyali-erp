# Auth & RBAC Quick Start Guide

**Purpose**: Get started with authentication models and role-based access control
**Approach**: Custom JWT-based authentication with sqlc
**Timeline**: Start here before building APIs
**Last Updated**: December 25, 2025

---

## 🎯 Why Start with Auth & RBAC?

Starting with authentication and RBAC before building APIs is the **right approach** because:

1. ✅ **Foundation First**: All API endpoints need authentication - build it once, use everywhere
2. ✅ **Type Safety Early**: sqlc generates models that handlers will use
3. ✅ **Security by Design**: Permissions baked in from day one
4. ✅ **Avoid Refactoring**: Don't build unprotected APIs then secure them later
5. ✅ **Test Early**: Auth bugs are easier to catch before complex business logic

---

## 📋 What You'll Build

### Phase 1: Database Schema (Day 1)
- Users table with password hashing
- Sessions table for JWT tokens
- Password resets table
- SQL migrations
- sqlc configuration

### Phase 2: Authentication (Days 2-3)
- User registration endpoint
- Login endpoint (returns JWT)
- Password reset flow
- JWT middleware
- Token refresh

### Phase 3: RBAC (Day 4)
- 7 user roles
- Permission matrix
- Permission checking middleware
- Role-based route protection

---

## 🗄️ Step 1: Database Schema with sqlc

### 1.1 Create Migration Files

Create `internal/database/migrations/001_initial_schema.up.sql`:

```sql
-- Users table with password hash
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN (
        'ADMIN',
        'CFO',
        'DIRECTOR',
        'FINANCE_OFFICER',
        'DEPARTMENT_MANAGER',
        'COMPLIANCE_OFFICER',
        'USER'
    )),
    department VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Sessions table for JWT tokens
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    refresh_token_hash VARCHAR(255) UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- Indexes for sessions
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Password resets table
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for password resets
CREATE INDEX idx_password_resets_token_hash ON password_resets(token_hash);
CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
```

Create `internal/database/migrations/001_initial_schema.down.sql`:

```sql
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
```

### 1.2 Configure sqlc

Create `sqlc.yaml` in project root:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/database/queries"
    schema: "internal/database/migrations"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_all_enum_values: true
```

### 1.3 Create sqlc Queries

Create `internal/database/queries/users.sql`:

```sql
-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    name,
    role,
    department,
    email_verified
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login = NOW(),
    failed_login_attempts = 0,
    locked_until = NULL
WHERE id = $1;

-- name: IncrementFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE
        WHEN failed_login_attempts + 1 >= 5
        THEN NOW() + INTERVAL '15 minutes'
        ELSE locked_until
    END
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: UpdateUserRole :exec
UPDATE users
SET role = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = true,
    updated_at = NOW()
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users
SET is_active = false,
    updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE
    (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role'))
    AND (sqlc.narg('department')::text IS NULL OR department = sqlc.narg('department'))
    AND (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE
    (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role'))
    AND (sqlc.narg('department')::text IS NULL OR department = sqlc.narg('department'))
    AND (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'));
```

Create `internal/database/queries/sessions.sql`:

```sql
-- name: CreateSession :one
INSERT INTO sessions (
    user_id,
    token_hash,
    refresh_token_hash,
    expires_at,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE token_hash = $1 LIMIT 1;

-- name: UpdateSessionActivity :exec
UPDATE sessions
SET last_activity = NOW()
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < NOW();
```

Create `internal/database/queries/password_resets.sql`:

```sql
-- name: CreatePasswordReset :one
INSERT INTO password_resets (
    user_id,
    token_hash,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPasswordResetByToken :one
SELECT * FROM password_resets
WHERE token_hash = $1
  AND used = false
  AND expires_at > NOW()
LIMIT 1;

-- name: MarkPasswordResetAsUsed :exec
UPDATE password_resets
SET used = true
WHERE id = $1;

-- name: DeleteExpiredPasswordResets :exec
DELETE FROM password_resets WHERE expires_at < NOW();
```

### 1.4 Generate sqlc Code

```bash
# Install sqlc if not installed
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate Go code from SQL
sqlc generate

# Output will be in internal/db/
# - db.go (main interface)
# - models.go (type-safe models)
# - querier.go (query interface)
# - users.sql.go (user queries)
# - sessions.sql.go (session queries)
# - password_resets.sql.go (password reset queries)
```

---

## 🔐 Step 2: Authentication Implementation

### 2.1 Create Auth Service

Create `internal/services/auth.go`:

```go
package services

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "time"

    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
    "yourapp/internal/db"
)

const (
    bcryptCost = 12
    accessTokenDuration = 1 * time.Hour
    refreshTokenDuration = 8 * time.Hour
)

type AuthService struct {
    queries *db.Queries
    jwtSecret []byte
}

func NewAuthService(queries *db.Queries, jwtSecret string) *AuthService {
    return &AuthService{
        queries: queries,
        jwtSecret: []byte(jwtSecret),
    }
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(bytes), err
}

// CheckPassword verifies a password against a hash
func (s *AuthService) CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// GenerateJWT generates access and refresh tokens
func (s *AuthService) GenerateJWT(user *db.User) (accessToken, refreshToken string, err error) {
    // Access token
    accessClaims := jwt.MapClaims{
        "user_id": user.ID,
        "email": user.Email,
        "name": user.Name,
        "role": user.Role,
        "department": user.Department,
        "exp": time.Now().Add(accessTokenDuration).Unix(),
        "iat": time.Now().Unix(),
        "type": "access",
    }

    accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessToken, err = accessTokenObj.SignedString(s.jwtSecret)
    if err != nil {
        return "", "", err
    }

    // Refresh token
    refreshClaims := jwt.MapClaims{
        "user_id": user.ID,
        "exp": time.Now().Add(refreshTokenDuration).Unix(),
        "iat": time.Now().Unix(),
        "type": "refresh",
    }

    refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshToken, err = refreshTokenObj.SignedString(s.jwtSecret)
    if err != nil {
        return "", "", err
    }

    return accessToken, refreshToken, nil
}

// ValidateJWT validates a JWT token
func (s *AuthService) ValidateJWT(tokenString string) (*jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.jwtSecret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &claims, nil
    }

    return nil, errors.New("invalid token")
}

// GenerateRandomToken generates a secure random token
func (s *AuthService) GenerateRandomToken(length int) (string, error) {
    b := make([]byte, length)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, email, password, name, role, department string) (*db.User, error) {
    // Hash password
    hash, err := s.HashPassword(password)
    if err != nil {
        return nil, err
    }

    // Create user
    user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
        Email: email,
        PasswordHash: hash,
        Name: name,
        Role: role,
        Department: department,
        EmailVerified: false,
    })

    return &user, err
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, user *db.User, err error) {
    // Get user by email
    dbUser, err := s.queries.GetUserByEmail(ctx, email)
    if err != nil {
        return "", "", nil, errors.New("invalid credentials")
    }

    // Check if account is locked
    if dbUser.LockedUntil.Valid && dbUser.LockedUntil.Time.After(time.Now()) {
        return "", "", nil, errors.New("account locked due to failed login attempts")
    }

    // Verify password
    if !s.CheckPassword(password, dbUser.PasswordHash) {
        // Increment failed attempts
        s.queries.IncrementFailedLoginAttempts(ctx, dbUser.ID)
        return "", "", nil, errors.New("invalid credentials")
    }

    // Check if account is active
    if !dbUser.IsActive {
        return "", "", nil, errors.New("account is deactivated")
    }

    // Generate JWT tokens
    accessToken, refreshToken, err = s.GenerateJWT(&dbUser)
    if err != nil {
        return "", "", nil, err
    }

    // Update last login
    s.queries.UpdateUserLastLogin(ctx, dbUser.ID)

    return accessToken, refreshToken, &dbUser, nil
}
```

### 2.2 Create Auth Middleware

Create `internal/middleware/auth.go`:

```go
package middleware

import (
    "strings"

    "github.com/gofiber/fiber/v2"
    "yourapp/internal/services"
)

func AuthMiddleware(authService *services.AuthService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get Authorization header
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code": "AUTHENTICATION_ERROR",
                    "message": "Missing authorization header",
                },
            })
        }

        // Extract token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code": "AUTHENTICATION_ERROR",
                    "message": "Invalid authorization format",
                },
            })
        }

        token := parts[1]

        // Validate token
        claims, err := authService.ValidateJWT(token)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code": "AUTHENTICATION_ERROR",
                    "message": "Invalid or expired token",
                },
            })
        }

        // Store user info in context
        c.Locals("user_id", (*claims)["user_id"])
        c.Locals("email", (*claims)["email"])
        c.Locals("role", (*claims)["role"])
        c.Locals("department", (*claims)["department"])

        return c.Next()
    }
}
```

### 2.3 Create Auth Handlers

Create `internal/handlers/auth.go`:

```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "yourapp/internal/services"
)

type AuthHandler struct {
    authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
    Email      string `json:"email" validate:"required,email"`
    Password   string `json:"password" validate:"required,min=8"`
    Name       string `json:"name" validate:"required"`
    Role       string `json:"role" validate:"required,oneof=ADMIN CFO DIRECTOR FINANCE_OFFICER DEPARTMENT_MANAGER COMPLIANCE_OFFICER USER"`
    Department string `json:"department"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
    var req RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "success": false,
            "error": fiber.Map{
                "code": "VALIDATION_ERROR",
                "message": "Invalid request body",
            },
        })
    }

    // Validate request
    // TODO: Add validation using go-playground/validator

    // Register user
    user, err := h.authService.Register(
        c.Context(),
        req.Email,
        req.Password,
        req.Name,
        req.Role,
        req.Department,
    )

    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "success": false,
            "error": fiber.Map{
                "code": "REGISTRATION_ERROR",
                "message": err.Error(),
            },
        })
    }

    return c.Status(201).JSON(fiber.Map{
        "success": true,
        "data": fiber.Map{
            "user": fiber.Map{
                "id": user.ID,
                "email": user.Email,
                "name": user.Name,
                "role": user.Role,
                "department": user.Department,
            },
        },
    })
}

// Login handles user authentication
func (h *AuthHandler) Login(c *fiber.Ctx) error {
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "success": false,
            "error": fiber.Map{
                "code": "VALIDATION_ERROR",
                "message": "Invalid request body",
            },
        })
    }

    // Authenticate user
    accessToken, refreshToken, user, err := h.authService.Login(
        c.Context(),
        req.Email,
        req.Password,
    )

    if err != nil {
        return c.Status(401).JSON(fiber.Map{
            "success": false,
            "error": fiber.Map{
                "code": "AUTHENTICATION_ERROR",
                "message": err.Error(),
            },
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "data": fiber.Map{
            "access_token": accessToken,
            "refresh_token": refreshToken,
            "expires_in": 3600,
            "user": fiber.Map{
                "id": user.ID,
                "email": user.Email,
                "name": user.Name,
                "role": user.Role,
                "department": user.Department,
            },
        },
    })
}
```

---

## 🛡️ Step 3: RBAC Implementation

### 3.1 Define Permissions

Create `internal/rbac/permissions.go`:

```go
package rbac

type Permission string

const (
    // Document permissions
    PermCreateDocument    Permission = "create_document"
    PermViewOwnDocuments  Permission = "view_own_documents"
    PermViewAllDocuments  Permission = "view_all_documents"
    PermEditDocument      Permission = "edit_document"
    PermDeleteDocument    Permission = "delete_document"

    // Approval permissions
    PermApproveStage1     Permission = "approve_stage_1"
    PermApproveStage2     Permission = "approve_stage_2"
    PermApproveStage3     Permission = "approve_stage_3"
    PermRejectTask        Permission = "reject_task"
    PermReassignTask      Permission = "reassign_task"
    PermBulkApprove       Permission = "bulk_approve"

    // Workflow permissions
    PermViewWorkflows     Permission = "view_workflows"
    PermCreateWorkflow    Permission = "create_workflow"
    PermEditWorkflow      Permission = "edit_workflow"
    PermDeleteWorkflow    Permission = "delete_workflow"

    // User permissions
    PermViewUsers         Permission = "view_users"
    PermCreateUser        Permission = "create_user"
    PermEditUser          Permission = "edit_user"
    PermDeleteUser        Permission = "delete_user"
    PermAssignRole        Permission = "assign_role"

    // Analytics permissions
    PermViewOwnAnalytics  Permission = "view_own_analytics"
    PermViewDeptAnalytics Permission = "view_dept_analytics"
    PermViewAllAnalytics  Permission = "view_all_analytics"
    PermExportReports     Permission = "export_reports"

    // Audit permissions
    PermViewAuditLogs     Permission = "view_audit_logs"
    PermExportAuditLogs   Permission = "export_audit_logs"

    // System permissions
    PermSystemConfig      Permission = "system_config"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[string][]Permission{
    "ADMIN": {
        // Full access
        PermCreateDocument, PermViewOwnDocuments, PermViewAllDocuments,
        PermEditDocument, PermDeleteDocument,
        PermApproveStage1, PermApproveStage2, PermApproveStage3,
        PermRejectTask, PermReassignTask, PermBulkApprove,
        PermViewWorkflows, PermCreateWorkflow, PermEditWorkflow, PermDeleteWorkflow,
        PermViewUsers, PermCreateUser, PermEditUser, PermDeleteUser, PermAssignRole,
        PermViewOwnAnalytics, PermViewDeptAnalytics, PermViewAllAnalytics, PermExportReports,
        PermViewAuditLogs, PermExportAuditLogs,
        PermSystemConfig,
    },
    "CFO": {
        PermCreateDocument, PermViewOwnDocuments, PermViewAllDocuments,
        PermApproveStage3, PermRejectTask, PermReassignTask, PermBulkApprove,
        PermViewWorkflows,
        PermViewUsers, PermViewDeptAnalytics, PermViewAllAnalytics, PermExportReports,
    },
    "DIRECTOR": {
        PermCreateDocument, PermViewOwnDocuments, PermViewAllDocuments,
        PermApproveStage3, PermRejectTask, PermReassignTask, PermBulkApprove,
        PermViewWorkflows,
        PermViewUsers, PermViewDeptAnalytics, PermViewAllAnalytics, PermExportReports,
    },
    "FINANCE_OFFICER": {
        PermCreateDocument, PermViewOwnDocuments, PermViewAllDocuments,
        PermApproveStage2, PermRejectTask, PermReassignTask, PermBulkApprove,
        PermViewWorkflows,
        PermViewDeptAnalytics, PermExportReports,
    },
    "DEPARTMENT_MANAGER": {
        PermCreateDocument, PermViewOwnDocuments,
        PermApproveStage1, PermRejectTask, PermReassignTask, PermBulkApprove,
        PermViewWorkflows,
        PermViewDeptAnalytics, PermExportReports,
    },
    "COMPLIANCE_OFFICER": {
        PermViewAllDocuments,
        PermViewWorkflows,
        PermViewUsers,
        PermViewAllAnalytics, PermExportReports,
        PermViewAuditLogs, PermExportAuditLogs,
    },
    "USER": {
        PermCreateDocument, PermViewOwnDocuments,
        PermViewWorkflows,
        PermViewOwnAnalytics,
    },
}

// HasPermission checks if a role has a specific permission
func HasPermission(role string, permission Permission) bool {
    permissions, ok := RolePermissions[role]
    if !ok {
        return false
    }

    for _, p := range permissions {
        if p == permission {
            return true
        }
    }

    return false
}
```

### 3.2 Create RBAC Middleware

Create `internal/middleware/rbac.go`:

```go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "yourapp/internal/rbac"
)

// RequirePermission middleware checks if user has required permission
func RequirePermission(permission rbac.Permission) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get role from context (set by AuthMiddleware)
        role, ok := c.Locals("role").(string)
        if !ok {
            return c.Status(403).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code": "AUTHORIZATION_ERROR",
                    "message": "User role not found",
                },
            })
        }

        // Check permission
        if !rbac.HasPermission(role, permission) {
            return c.Status(403).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code": "AUTHORIZATION_ERROR",
                    "message": "Insufficient permissions",
                },
            })
        }

        return c.Next()
    }
}

// RequireRole middleware checks if user has required role
func RequireRole(allowedRoles ...string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        role, ok := c.Locals("role").(string)
        if !ok {
            return c.Status(403).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code": "AUTHORIZATION_ERROR",
                    "message": "User role not found",
                },
            })
        }

        // Check if role is in allowed roles
        for _, allowedRole := range allowedRoles {
            if role == allowedRole {
                return c.Next()
            }
        }

        return c.Status(403).JSON(fiber.Map{
            "success": false,
            "error": fiber.Map{
                "code": "AUTHORIZATION_ERROR",
                "message": "Insufficient permissions",
            },
        })
    }
}
```

---

## 🚀 Step 4: Wire Everything Together

Create `cmd/api/main.go`:

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/jackc/pgx/v5/pgxpool"

    "yourapp/internal/db"
    "yourapp/internal/handlers"
    "yourapp/internal/middleware"
    "yourapp/internal/rbac"
    "yourapp/internal/services"
)

func main() {
    // Load environment
    dbURL := os.Getenv("DATABASE_URL")
    jwtSecret := os.Getenv("JWT_SECRET")

    // Connect to database
    ctx := context.Background()
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        log.Fatal("Unable to connect to database:", err)
    }
    defer pool.Close()

    // Create queries
    queries := db.New(pool)

    // Create services
    authService := services.NewAuthService(queries, jwtSecret)

    // Create handlers
    authHandler := handlers.NewAuthHandler(authService)

    // Create Fiber app
    app := fiber.New()

    // Global middleware
    app.Use(logger.New())
    app.Use(recover.New())

    // Health check
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "ok",
            "timestamp": time.Now(),
        })
    })

    // Public routes (no auth required)
    auth := app.Group("/api/auth")
    auth.Post("/register", authHandler.Register)
    auth.Post("/login", authHandler.Login)

    // Protected routes (auth required)
    api := app.Group("/api", middleware.AuthMiddleware(authService))

    // Admin-only routes
    admin := api.Group("/admin", middleware.RequireRole("ADMIN"))
    admin.Get("/users", func(c *fiber.Ctx) error {
        // List users endpoint
        return c.JSON(fiber.Map{"message": "Admin users endpoint"})
    })

    // Approval routes with permission checks
    approvals := api.Group("/approvals")
    approvals.Post("/tasks/:id/approve",
        middleware.RequirePermission(rbac.PermApproveStage1),
        func(c *fiber.Ctx) error {
            // Approve task endpoint
            return c.JSON(fiber.Map{"message": "Approve task"})
        },
    )

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "3001"
    }

    log.Fatal(app.Listen(":" + port))
}
```

---

## ✅ Testing Your Implementation

### Test 1: Register a User

```bash
curl -X POST http://localhost:3001/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@company.com",
    "password": "SecurePass123",
    "name": "Admin User",
    "role": "ADMIN",
    "department": "IT"
  }'
```

### Test 2: Login

```bash
curl -X POST http://localhost:3001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@company.com",
    "password": "SecurePass123"
  }'
```

Response:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "email": "admin@company.com",
      "name": "Admin User",
      "role": "ADMIN"
    }
  }
}
```

### Test 3: Protected Route

```bash
curl -X GET http://localhost:3001/api/admin/users \
  -H "Authorization: Bearer eyJhbGc..."
```

---

## 📊 What You've Built

After completing this guide, you'll have:

✅ **Database Schema**
- 3 tables (users, sessions, password_resets)
- sqlc generating type-safe Go code
- Migration system ready

✅ **Authentication**
- User registration
- Email/password login
- JWT token generation
- Password hashing with bcrypt
- Token validation middleware

✅ **RBAC System**
- 7 user roles defined
- 25+ permissions mapped
- Permission checking middleware
- Role-based route protection

✅ **Security**
- Bcrypt password hashing (cost 12)
- JWT tokens with expiration
- Account lockout after failed attempts
- Protected API routes

---

## 🎯 Next Steps

Now that auth and RBAC are in place, you can:

1. **Build Protected APIs** - All endpoints can use your auth middleware
2. **Add Password Reset** - Implement the full password reset flow
3. **Email Verification** - Add email verification for new accounts
4. **Session Management** - Add logout and session cleanup
5. **Audit Logging** - Log all authentication events

---

**Last Updated**: December 25, 2025
**Status**: Ready to implement
**Estimated Time**: 4 days
