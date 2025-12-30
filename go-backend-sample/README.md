# Liyali Gateway - Go Backend

High-performance Go backend API for the Liyali Gateway approval workflow system.

## Tech Stack

- **Go 1.21+** - Backend language
- **Fiber v3** - Web framework
- **PostgreSQL 12+** - Database
- **pgx v5** - PostgreSQL driver
- **sqlc** - Type-safe SQL query generation
- **JWT (HS256)** - Authentication
- **bcrypt** - Password hashing

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Environment configuration
│   ├── database/
│   │   ├── migrations/            # SQL migration files
│   │   └── queries/               # sqlc query definitions
│   ├── db/                        # Generated sqlc code
│   │   ├── models.go              # Type-safe models
│   │   ├── querier.go             # Query interface
│   │   └── *.sql.go               # Generated query code
│   ├── handlers/
│   │   └── auth_handler.go        # HTTP request handlers
│   ├── middleware/
│   │   ├── auth_middleware.go     # JWT authentication
│   │   └── rbac_middleware.go     # Permission checking
│   ├── repository/
│   │   ├── interfaces.go          # Repository interfaces
│   │   ├── user_repository.go     # User data access
│   │   ├── session_repository.go  # Session data access
│   │   └── password_reset_repository.go
│   ├── services/
│   │   └── auth_service.go        # Business logic
│   ├── rbac/
│   │   └── permissions.go         # RBAC system (7 roles, 35+ permissions)
│   └── utils/
│       └── pgtype_utils.go        # Type conversion helpers
├── tests/
│   ├── unit/                      # Unit tests with mocks
│   └── integration/               # Integration tests
├── docs/                          # Documentation
├── .env.example                   # Environment template
├── go.mod                         # Go modules
├── Makefile                       # Build commands
└── sqlc.yaml                      # sqlc configuration
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Make (optional)

### Setup

1. **Clone and navigate**:
```bash
cd backend
```

2. **Install dependencies**:
```bash
go mod download
```

3. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your database URL and JWT secret
```

4. **Run migrations**:
```bash
make migrate-up
# or
go run cmd/migrate/main.go up
```

5. **Start server**:
```bash
go run cmd/api/main.go
```

Server runs on `http://localhost:3001`

## Available Commands

```bash
# Development
go run cmd/api/main.go          # Start server
go test ./...                   # Run all tests
go test ./tests/unit/...        # Run unit tests
go test ./tests/integration/... # Run integration tests

# Build
go build -o bin/api cmd/api/main.go

# Database
make migrate-up                 # Run migrations
make migrate-down              # Rollback migrations
make sqlc-generate             # Generate sqlc code

# Testing
go test ./... -v               # Verbose tests
go test ./... -coverprofile=coverage.out  # With coverage
```

## API Endpoints

### Authentication (`/api/auth`)

- `POST /register` - Register new user
- `POST /login` - Login with email/password
- `POST /refresh` - Refresh access token
- `POST /logout` - Logout user
- `POST /password-reset/request` - Request password reset
- `POST /password-reset/confirm` - Reset password
- `POST /change-password` - Change password (protected)
- `GET /me` - Get current user (protected)

### Health Check

- `GET /health` - Health check endpoint

## Authentication

JWT-based authentication with:
- 1-hour access tokens
- 8-hour refresh tokens
- bcrypt password hashing (cost 12)
- Account lockout after 5 failed attempts (15 min)

## RBAC System

**7 Roles**:
1. **ADMIN** - Full system access
2. **CFO** - Executive approvals
3. **DIRECTOR** - Senior approvals
4. **FINANCE_OFFICER** - Financial approvals
5. **DEPARTMENT_MANAGER** - Department approvals
6. **COMPLIANCE_OFFICER** - View-only compliance
7. **REQUESTER** - Submit and view own requests

**35+ Permissions** across:
- Approvals (approve, reject, reassign)
- Workflows (create, update, delete)
- Users (manage, deactivate)
- Documents (view, export)
- Analytics (view metrics)

## Database Schema

### Current Tables (Phase 1)
- `users` - User accounts with roles
- `sessions` - JWT refresh tokens
- `password_resets` - Password reset tokens

### Upcoming Tables (Phase 2)
- `approval_tasks` - Approval workflow tasks
- `approval_history` - Approval audit trail
- `documents` - Document storage
- `workflows` - Workflow definitions
- `audit_logs` - System audit logs
- `notifications` - Email notifications

## Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test -run TestLogin ./tests/unit/...
```

**Current Coverage**:
- ✅ 7/7 unit tests passing
- ✅ 10 integration test cases ready
- Repository pattern with mock interfaces

## Development Status

### ✅ Completed (Phase 1, 3, 5)

- [x] Project setup with Go modules
- [x] PostgreSQL connection with pgx
- [x] sqlc configuration and code generation
- [x] Database migrations (3 auth tables)
- [x] Config management with environment variables
- [x] Complete authentication system
  - [x] User registration with validation
  - [x] Login with JWT tokens
  - [x] Password reset flow
  - [x] Change password
  - [x] Logout with token revocation
- [x] Repository layer with interfaces
- [x] Service layer (auth business logic)
- [x] HTTP handlers for all auth endpoints
- [x] JWT middleware for protected routes
- [x] RBAC middleware (7 roles, 35+ permissions)
- [x] Fiber server with CORS and logging
- [x] Unit tests with mock repositories
- [x] Integration tests for auth flow
- [x] Account lockout after failed login attempts

### 🔄 In Progress (Phase 2)

- [ ] Remaining 6 database tables
- [ ] sqlc queries for new tables
- [ ] Repositories for new tables

### 📋 Upcoming

- [ ] Core API endpoints (approvals, workflows, documents)
- [ ] Email notifications with SendGrid
- [ ] Audit logging service
- [ ] Rate limiting middleware
- [ ] Email verification flow
- [ ] Performance optimization
- [ ] Production deployment

## Environment Variables

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/liyali_gateway

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Server
PORT=3001
ENVIRONMENT=development

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

# Email (Future)
SENDGRID_API_KEY=your-sendgrid-api-key
```

## Documentation

See [docs/](./docs/) folder for detailed documentation organized by topic:

- **Getting Started** - Setup and configuration guides
- **Authentication & Security** - Auth flow, RBAC, security features
- **Database** - Schema, migrations, sqlc usage
- **API Reference** - Complete endpoint documentation
- **Development** - Contribution guidelines, workflow

## Contributing

1. Create feature branch: `git checkout -b feature/your-feature`
2. Make changes with tests
3. Run tests: `go test ./...`
4. Commit: `git commit -m "feat: add feature"`
5. Push: `git push origin feature/your-feature`
6. Create Pull Request

## License

Proprietary - All rights reserved

## Support

For issues and questions, contact the development team.

---

**Built with** ❤️ **using Go + Fiber + PostgreSQL + sqlc**
