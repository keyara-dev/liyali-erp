# Liyali Gateway - Go Backend Development Planner

**Project**: Liyali Gateway Approval Workflow System
**Backend**: Go with Fiber Framework
**Database**: PostgreSQL with sqlc
**Authentication**: Custom JWT-based authentication
**Status**: Ready for Phase 13 Implementation
**Last Updated**: December 25, 2025

---

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Development Phases](#development-phases)
3. [Tech Stack](#tech-stack)
4. [Prerequisites](#prerequisites)
5. [Project Structure](#project-structure)
6. [Development Workflow](#development-workflow)
7. [Phase-by-Phase Implementation](#phase-by-phase-implementation)
8. [Testing Strategy](#testing-strategy)
9. [Deployment Plan](#deployment-plan)
10. [Team & Resources](#team--resources)

---

## 📖 Project Overview

### What We're Building

A high-performance Go backend API for the Liyali Gateway workflow approval system that will:
- Replace localStorage with PostgreSQL database
- Provide REST APIs for all workflow operations
- Implement custom JWT-based authentication (email/password)
- Handle email notifications via SendGrid
- Enforce role-based access control (RBAC)
- Maintain comprehensive audit logging
- Use sqlc for type-safe SQL queries

### Frontend Integration

The backend will integrate with the existing Next.js 16 frontend that has:
- ✅ 5 Workflow Types (Requisition, Budget, PO, Payment Voucher, GRN)
- ✅ Multi-stage Approvals with digital signatures
- ✅ Government-Compliant PDF exports
- ✅ Real-time Analytics Dashboard
- ✅ Bulk Operations
- ✅ 18+ Server Actions ready for API migration

### Success Criteria

- [ ] All 18+ frontend server actions connected to backend APIs
- [ ] PostgreSQL database fully operational
- [ ] Custom JWT authentication working (email/password login)
- [ ] Password reset flow functional
- [ ] Email notifications sending
- [ ] Audit logging recording all actions
- [ ] RBAC enforced at API level
- [ ] API response time < 100ms
- [ ] 0 security vulnerabilities
- [ ] 100% API test coverage
- [ ] sqlc generating type-safe queries

---

## 🚀 Development Phases

### Phase 1: Project Setup & Foundation (Days 1-3) ✅ COMPLETED
**Duration**: 3 days
**Goal**: Get development environment ready

- [x] Initialize Go project with modules
- [x] Set up Fiber web framework
- [x] Install and configure sqlc
- [x] Set up PostgreSQL connection (pgx driver)
- [x] Create database schema (SQL migrations)
- [x] Configure sqlc.yaml for query generation
- [x] Set up project folder structure
- [x] Configure environment variables
- [ ] Set up logging and monitoring (zerolog) - using Fiber's built-in logger
- [x] Create basic health check endpoint

**Deliverables**:
- ✅ Working Go project with modules
- ✅ Database connected with pgx in main.go
- ✅ sqlc configured and generating type-safe code
- ✅ Basic API responding at `/health`

---

### Phase 2: Database Schema & sqlc Setup (Days 4-6) 🔄 IN PROGRESS
**Duration**: 3 days
**Goal**: Complete database schema and sqlc query generation

#### Tables to Create (9 total)

1. ✅ **users** - User accounts and roles (with password hash)
2. ✅ **sessions** - Authentication sessions (JWT tokens)
3. ✅ **password_resets** - Password reset tokens
4. ⏳ **approval_tasks** - Tasks to approve
5. ⏳ **approval_history** - Approval records
6. ⏳ **documents** - Base document storage
7. ⏳ **workflows** - Workflow definitions
8. ⏳ **audit_logs** - Action tracking
9. ⏳ **notifications** - System notifications

**Tasks**:
- [x] Write SQL migration files for auth tables (users, sessions, password_resets)
- [x] Define foreign keys and indexes in SQL
- [x] Create sqlc queries for auth CRUD operations
- [x] Configure sqlc.yaml (package name, output paths)
- [x] Run sqlc generate to create Go code
- [ ] Create seed data SQL scripts
- [x] Run migrations for auth tables
- [x] Verify generated sqlc code compiles

**Deliverables**:
- ✅ 3/9 tables created via migrations (auth tables)
- ✅ sqlc generating type-safe Go code
- ✅ Auth CRUD queries defined in SQL
- ⏳ Sample data loaded (pending)
- ⏳ Remaining 6 tables (approval_tasks, approval_history, documents, workflows, audit_logs, notifications)

---

### Phase 3: Custom Authentication System (Days 7-11) ✅ COMPLETED
**Duration**: 5 days
**Goal**: Implement secure custom authentication

> 📚 **Reference**: See [user-management-plan.md](./user-management-plan.md) for complete user roles, permissions, and authentication flow details.

#### Custom Authentication Components

**Password-based Login**:
- [x] User registration with email/password
- [x] Password hashing with bcrypt (cost 12)
- [x] Email/password login endpoint
- [x] JWT token generation (HS256 algorithm)
- [x] Refresh token mechanism
- [x] Password strength validation (8 chars minimum)

**Password Reset Flow**:
- [x] Request password reset endpoint
- [x] Generate secure reset token (service implemented)
- [ ] Send reset email via SendGrid (deferred to Phase 6)
- [x] Verify reset token endpoint
- [x] Update password endpoint (service implemented)
- [x] Token expiration (1 hour)

**Session Management**:
- [x] JWT token validation middleware
- [x] 1-hour access token expiration
- [x] 8-hour refresh token expiration
- [x] Token revocation on logout (service implemented)
- [x] Concurrent session handling (service implemented)
- [ ] Session cleanup job (deferred)

**Security Features**:
- [ ] Rate limiting on login endpoint (5 attempts/15 min) - deferred
- [x] Account lockout after failed attempts (5 attempts, 15 min lockout)
- [ ] Email verification for new accounts - deferred to later
- [ ] Secure cookie settings (httpOnly, secure, sameSite) - deferred
- [ ] CSRF protection tokens - deferred

**Tasks**:
- [x] Create user registration service with validation
- [x] Implement bcrypt password hashing
- [x] Create login endpoint handler (`/api/auth/login`)
- [x] Implement JWT token generation service
- [x] Create JWT validation middleware
- [x] Add role extraction from JWT claims
- [x] Create password reset service (4 methods)
- [ ] Implement email verification flow (deferred)
- [ ] Add rate limiting middleware (deferred)
- [x] Create logout service with token revocation
- [x] Write authentication tests (7 unit tests, 10 integration tests)

**Deliverables**:
- ✅ Auth service layer complete (registration, login, JWT, password reset, logout)
- ✅ Repository layer complete with interfaces (users, sessions, password_resets)
- ✅ JWT middleware protecting routes
- ✅ Auth handlers complete (8 endpoints)
- ✅ Auth routes wired up in main.go
- ✅ All auth tests passing (7/7 unit tests, 10 integration test cases)
- ✅ Server running on localhost:3001
- ⏳ Email verification functional (deferred)
- ⏳ Rate limiting preventing brute force (deferred)

---

### Phase 4: Core API Endpoints (Days 12-17)
**Duration**: 6 days
**Goal**: Implement all workflow APIs using sqlc

> 📚 **Reference**: See [api-specification.md](./api-specification.md) for complete API documentation with request/response examples.

#### API Endpoints to Create

**Approval Tasks** (`/api/approvals`):
- [ ] `GET /tasks` - Fetch approval tasks
- [ ] `GET /tasks/:id` - Get single task detail
- [ ] `POST /tasks/:id/approve` - Approve task
- [ ] `POST /tasks/:id/reject` - Reject task
- [ ] `POST /tasks/:id/reassign` - Reassign task

**Bulk Operations** (`/api/approvals/bulk`):
- [ ] `POST /approve` - Bulk approve
- [ ] `POST /reject` - Bulk reject
- [ ] `POST /reassign` - Bulk reassign

**Analytics** (`/api/analytics`):
- [ ] `GET /metrics` - Dashboard metrics
- [ ] `GET /trends` - 7-day trends
- [ ] `GET /bottleneck` - Bottleneck analysis

**Workflows** (`/api/workflows`):
- [ ] `GET /` - List workflows
- [ ] `GET /:id` - Get workflow detail
- [ ] `POST /` - Create workflow
- [ ] `PUT /:id` - Update workflow
- [ ] `DELETE /:id` - Delete workflow

**Documents** (per type: requisitions, purchase-orders, payment-vouchers, budgets, grn):
- [ ] `GET /` - List documents
- [ ] `GET /:id` - Get document detail
- [ ] `POST /` - Create document
- [ ] `PUT /:id` - Update document
- [ ] `DELETE /:id` - Delete document

**Tasks**:
- [ ] Write sqlc queries for all data operations
- [ ] Generate Go code with sqlc
- [ ] Create handler functions for all endpoints
- [ ] Implement request validation (go-playground/validator)
- [ ] Add error handling middleware
- [ ] Implement pagination with sqlc LIMIT/OFFSET
- [ ] Add filtering and sorting in SQL queries
- [ ] Test all endpoints with Postman/Insomnia

**Deliverables**:
- 40+ API endpoints working
- All queries type-safe via sqlc
- Request/response validation
- Error handling implemented
- Pagination working efficiently

---

### Phase 5: RBAC Implementation (Days 18-20) ✅ COMPLETED
**Duration**: 3 days
**Goal**: Enforce role-based permissions

> 📚 **Reference**: See [user-management-plan.md](./user-management-plan.md) for complete permission matrix and role capabilities.

#### 7 User Roles (Aligned with Frontend)

1. **DEPARTMENT_MANAGER** - Can approve at stage 1
2. **FINANCE_OFFICER** - Can approve at stage 2
3. **DIRECTOR** - Can approve at stage 3
4. **CFO** - Can approve at stage 3
5. **COMPLIANCE_OFFICER** - Can view all, no approve
6. **ADMIN** - Full access
7. **REQUESTER** - Can submit, view own (formerly USER)

#### Permission Matrix

```
                Submit  Approve  Reject  Reassign  View-All  Delete
Manager           ✓       ✓        ✓       ✓         ✗        ✗
Finance           ✓       ✓        ✓       ✓         ✗        ✗
Director          ✓       ✓        ✓       ✓         ✓        ✗
CFO               ✓       ✓        ✓       ✓         ✓        ✗
Compliance        ✓       ✗        ✗       ✗         ✓        ✗
Admin             ✓       ✓        ✓       ✓         ✓        ✓
Requester         ✓       ✗        ✗       ✗         ✗        ✗
```

**Tasks**:
- [x] Create permission checking middleware (RequirePermission, RequireRole, etc.)
- [x] Define permission constants (35+ permissions defined)
- [x] Map roles to permissions
- [x] Role validation in user registration
- [ ] Add role checks to all endpoints (deferred to Phase 4)
- [ ] Test permission enforcement (deferred to Phase 4)
- [x] Document permission requirements

**Deliverables**:
- ✅ RBAC middleware functional (5 middleware functions)
- ✅ Permission system complete (35+ permissions, 7 roles)
- ✅ Roles aligned with frontend (REQUESTER instead of USER)
- ✅ Role validation in registration endpoint
- ⏳ All endpoints protected (deferred to Phase 4)
- ⏳ Permission tests passing (deferred to Phase 4)

---

### Phase 6: Email Notifications (Days 21-22)
**Duration**: 2 days
**Goal**: Send email notifications

#### Email Templates (3 types)

1. **Task Assigned** - New approval waiting
2. **Task Approved** - Task was approved
3. **Task Rejected** - Task was rejected

**Tasks**:
- [ ] Set up SendGrid account and API key
- [ ] Create email service module
- [ ] Design HTML email templates
- [ ] Implement send functions
- [ ] Add email triggers to approval actions
- [ ] Test email delivery
- [ ] Handle bounces and failures

**Deliverables**:
- Email service functional
- 3 templates created
- Emails sending on actions

---

### Phase 7: Audit Logging (Days 23-24)
**Duration**: 2 days
**Goal**: Log all system actions

#### What to Log

- Every approval action
- Every rejection
- Every reassignment
- System configuration changes
- Permission changes
- Login/logout events

**Tasks**:
- [ ] Create audit logging service
- [ ] Implement `logAudit()` function
- [ ] Add logging to all mutations
- [ ] Create audit query endpoints
- [ ] Add audit log filtering
- [ ] Test audit trail completeness

**Deliverables**:
- Audit logging functional
- All actions logged
- Audit query API working

---

### Phase 8: Testing & Quality Assurance (Days 25-29)
**Duration**: 5 days
**Goal**: Ensure system reliability

#### Testing Layers

**Unit Tests**:
- [ ] Test all models
- [ ] Test all services
- [ ] Test utility functions
- [ ] Test middleware

**Integration Tests**:
- [ ] Test API endpoints
- [ ] Test database transactions
- [ ] Test email delivery
- [ ] Test permission enforcement

**E2E Tests**:
- [ ] Test complete approval workflow
- [ ] Test bulk operations
- [ ] Test authentication flow
- [ ] Test error scenarios

**Performance Tests**:
- [ ] Load testing (1000+ concurrent users)
- [ ] Database query optimization
- [ ] API response time verification
- [ ] Memory leak detection

**Security Tests**:
- [ ] SQL injection protection
- [ ] XSS prevention
- [ ] CSRF protection
- [ ] Authentication bypass attempts
- [ ] Authorization bypass attempts

**Tasks**:
- [ ] Write unit tests (target: 80%+ coverage)
- [ ] Write integration tests
- [ ] Write E2E tests
- [ ] Run performance tests
- [ ] Run security scan
- [ ] Fix all issues found

**Deliverables**:
- 80%+ test coverage
- All tests passing
- Performance benchmarks met
- 0 security vulnerabilities

---

### Phase 9: Deployment & DevOps (Days 30-34)
**Duration**: 5 days
**Goal**: Deploy to production

#### Infrastructure Setup

**Staging Environment**:
- [ ] Set up PostgreSQL database
- [ ] Deploy Go application
- [ ] Configure environment variables
- [ ] Set up monitoring
- [ ] Run smoke tests

**Production Environment**:
- [ ] Set up production database (AWS RDS, Azure, etc.)
- [ ] Deploy application
- [ ] Configure load balancer
- [ ] Set up SSL/TLS
- [ ] Configure monitoring and alerts
- [ ] Set up backup system

#### Deployment Strategy

**4-Phase Rollout**:
1. **Phase 1 (10%)**: Deploy to 10% of users
2. **Phase 2 (25%)**: Expand to 25% of users
3. **Phase 3 (50%)**: Expand to 50% of users
4. **Phase 4 (100%)**: Full rollout

**Tasks**:
- [ ] Create deployment scripts
- [ ] Set up CI/CD pipeline (GitHub Actions/GitLab CI)
- [ ] Configure environment variables
- [ ] Run staging deployment
- [ ] Verify all features work
- [ ] Get stakeholder approval
- [ ] Execute production deployment
- [ ] Monitor error rates
- [ ] Verify performance

**Deliverables**:
- Staging environment live
- Production environment live
- Monitoring active
- Rollback plan ready

---

## 🛠️ Tech Stack

### Backend Technologies

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Language** | Go | 1.21+ | Backend language |
| **Web Framework** | Fiber | 3.x | HTTP routing |
| **Database** | PostgreSQL | 12+ | Data persistence |
| **Database Driver** | pgx | 5.x | PostgreSQL driver |
| **Query Builder** | sqlc | 1.25+ | Type-safe SQL queries |
| **Migrations** | golang-migrate | 4.x | Database migrations |
| **Authentication** | Custom JWT | - | Token-based auth |
| **Password Hashing** | bcrypt | - | Secure password storage |
| **Email** | SendGrid | - | Email delivery |
| **Logging** | zerolog | - | Structured logging |
| **Testing** | Testify | - | Test framework |
| **Validation** | go-playground/validator | 10.x | Request validation |
| **Rate Limiting** | fiber-limiter | - | API rate limiting |

### Infrastructure

- **Cloud**: AWS / Azure / GCP
- **Database Hosting**: AWS RDS / Azure Database
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack / CloudWatch
- **CI/CD**: GitHub Actions / GitLab CI

---

## 📋 Prerequisites

### Required Software

- [ ] Go 1.21 or higher
- [ ] PostgreSQL 12 or higher
- [ ] Git
- [ ] Docker (optional, for local development)
- [ ] Postman/Insomnia (API testing)

### Required Accounts

- [ ] SendGrid account (email delivery)
- [ ] SMTP server credentials (email sending)
- [ ] Cloud provider account (AWS/Azure/GCP for hosting)
- [ ] Domain for email (verified with SendGrid)

### Development Tools

- [ ] VS Code with Go extension
- [ ] Database management tool (pgAdmin, DBeaver)
- [ ] API testing tool (Postman, Insomnia)
- [ ] Git client

---

## 📁 Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── database/
│   │   ├── db.go                  # PostgreSQL connection (pgx)
│   │   ├── migrations/            # SQL migration files
│   │   └── queries/               # sqlc query definitions (.sql files)
│   ├── db/                        # Generated sqlc code (auto-generated)
│   │   ├── db.go                  # Database interface
│   │   ├── models.go              # Type-safe models
│   │   ├── querier.go             # Query interface
│   │   ├── users.sql.go           # User queries
│   │   ├── approval_tasks.sql.go  # Approval task queries
│   │   ├── workflows.sql.go       # Workflow queries
│   │   └── ...                    # Other generated files
│   ├── handlers/
│   │   ├── auth.go                # Auth handlers
│   │   ├── approvals.go           # Approval handlers
│   │   ├── workflows.go           # Workflow handlers
│   │   ├── analytics.go           # Analytics handlers
│   │   └── documents.go           # Document handlers
│   ├── middleware/
│   │   ├── auth.go                # Auth middleware
│   │   ├── rbac.go                # RBAC middleware
│   │   ├── logger.go              # Logging middleware
│   │   └── error.go               # Error handling middleware
│   ├── services/
│   │   ├── auth.go                # Auth service
│   │   ├── approval.go            # Approval service
│   │   ├── email.go               # Email service
│   │   └── audit.go               # Audit service
│   ├── routes/
│   │   └── routes.go              # Route definitions
│   └── utils/
│       ├── response.go            # API response helpers
│       └── validation.go          # Validation helpers
├── tests/
│   ├── unit/                      # Unit tests
│   ├── integration/               # Integration tests
│   └── e2e/                       # End-to-end tests
├── docs/
│   ├── development-planner.md     # This file
│   └── implementation-guide.md    # Detailed implementation steps
├── scripts/
│   ├── migrate.sh                 # Migration script
│   ├── seed.sh                    # Seed data script
│   └── deploy.sh                  # Deployment script
├── .env.example                   # Environment variables template
├── .gitignore                     # Git ignore file
├── go.mod                         # Go modules
├── go.sum                         # Go modules checksum
├── Dockerfile                     # Docker configuration
├── docker-compose.yml             # Docker compose for local dev
└── README.md                      # Project README
```

---

## 🔄 Development Workflow

### Daily Workflow

1. **Morning Standup** (15 min)
   - What did I do yesterday?
   - What will I do today?
   - Any blockers?

2. **Development** (6 hours)
   - Pick a task from current phase
   - Write code following best practices
   - Write tests for new code
   - Run tests locally

3. **Code Review** (1 hour)
   - Submit PR for review
   - Review others' PRs
   - Address feedback

4. **Documentation** (30 min)
   - Update API docs
   - Update README if needed
   - Document any decisions made

5. **End of Day** (15 min)
   - Commit and push changes
   - Update task status
   - Plan tomorrow's work

### Git Workflow

```bash
# Start new feature
git checkout -b feature/approval-endpoints

# Make changes
git add .
git commit -m "feat: add approval endpoints"

# Push to remote
git push origin feature/approval-endpoints

# Create Pull Request
# Get approval
# Merge to main
```

### Testing Workflow

```bash
# Run unit tests
go test ./... -v

# Run integration tests
go test ./tests/integration/... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test -run TestApproveTask ./internal/handlers/...
```

---

## 🧪 Testing Strategy

### Unit Tests (Target: 80% Coverage)

**What to Test**:
- All model methods
- All service functions
- All utility functions
- Middleware logic

**Example**:
```go
func TestApproveTask(t *testing.T) {
    // Setup
    db := setupTestDB()
    handler := NewApprovalHandler(db)

    // Execute
    result := handler.ApproveTask(mockRequest)

    // Assert
    assert.Equal(t, "approved", result.Status)
    assert.NotNil(t, result.Signature)
}
```

### Integration Tests

**What to Test**:
- API endpoints
- Database transactions
- Email delivery
- Permission enforcement

**Example**:
```go
func TestApprovalWorkflowE2E(t *testing.T) {
    // Create task
    task := createTestTask()

    // Approve at stage 1
    approveTask(task.ID, stage1User)

    // Verify status change
    updated := getTask(task.ID)
    assert.Equal(t, "IN_REVIEW", updated.Status)
}
```

### Performance Tests

**Benchmarks**:
- API response time < 100ms
- Database queries < 50ms
- Concurrent user handling: 1000+
- Memory usage stable

---

## 🚀 Deployment Plan

### Staging Deployment

**Steps**:
1. Deploy database with migrations
2. Deploy application
3. Run smoke tests
4. Verify all endpoints work
5. Get QA sign-off

### Production Deployment

**Pre-deployment Checklist**:
- [ ] All tests passing
- [ ] Code reviewed and approved
- [ ] Documentation updated
- [ ] Database backup completed
- [ ] Rollback plan ready
- [ ] Monitoring configured
- [ ] Team notified

**Deployment Steps**:
1. Enable maintenance mode
2. Backup database
3. Run migrations
4. Deploy new code
5. Verify health checks
6. Disable maintenance mode
7. Monitor for errors

**Post-deployment**:
- [ ] Monitor error rates (target: <0.1%)
- [ ] Monitor performance (API < 100ms)
- [ ] Check logs for issues
- [ ] Verify all features working
- [ ] Collect user feedback

---

## 👥 Team & Resources

### Team Composition

**Required Roles**:
- **1 Backend Developer** (Go) - 32 days full-time
- **1 DevOps Engineer** - 10 days
- **1 QA Engineer** - 10 days
- **1 Technical Lead** (part-time oversight)

### Time Estimate

| Phase | Duration | Developer Days |
|-------|----------|----------------|
| Phase 1: Setup | 3 days | 3 |
| Phase 2: Database & sqlc | 3 days | 3 |
| Phase 3: Custom Auth | 5 days | 5 |
| Phase 4: APIs | 6 days | 6 |
| Phase 5: RBAC | 3 days | 3 |
| Phase 6: Email | 2 days | 2 |
| Phase 7: Audit | 2 days | 2 |
| Phase 8: Testing | 5 days | 5 |
| Phase 9: Deployment | 5 days | 5 |
| **Total** | **34 days** | **34** |

> 💡 **Recommended Start**: Begin with Phases 2-3 (Auth Models & RBAC) before APIs. This establishes the foundation for all protected endpoints.

### Budget Estimate

**Developer Costs**:
- Backend Developer: 32 days × $X/day
- DevOps Engineer: 10 days × $Y/day
- QA Engineer: 10 days × $Z/day

**Infrastructure Costs** (monthly):
- Database hosting: $100-300
- Application hosting: $50-150
- SendGrid: $15-100
- Monitoring tools: $50-100

---

## 📊 Progress Tracking

### Phase 1: Setup ⏳
- [ ] 0/8 tasks complete

### Phase 2: Database ⏳
- [ ] 0/7 tasks complete

### Phase 3: Authentication ⏳
- [ ] 0/6 tasks complete

### Phase 4: APIs ⏳
- [ ] 0/40 endpoints complete

### Phase 5: RBAC ⏳
- [ ] 0/5 tasks complete

### Phase 6: Email ⏳
- [ ] 0/7 tasks complete

### Phase 7: Audit ⏳
- [ ] 0/6 tasks complete

### Phase 8: Testing ⏳
- [ ] 0/10 test suites complete

### Phase 9: Deployment ⏳
- [ ] 0/9 deployment steps complete

---

## 🎯 Success Metrics

### Technical Metrics
- [ ] API response time < 100ms (95th percentile)
- [ ] Database query time < 50ms
- [ ] Test coverage > 80%
- [ ] 0 critical security vulnerabilities
- [ ] 0 high-priority bugs

### Business Metrics
- [ ] 100% of server actions migrated
- [ ] All 5 workflow types working
- [ ] Email notifications sending
- [ ] Audit logs complete
- [ ] User acceptance achieved

---

## 📞 Support & Resources

### Documentation
- [User Management Plan](./user-management-plan.md) - Complete user roles & permissions guide
- [API Specification](./api-specification.md) - REST API reference (40+ endpoints)
- [Go Backend Technical Guide](../../docs/BACKEND-GUIDE-GO.md) - Detailed Go implementation guide
- [Frontend Documentation](../../docs/README.md) - Frontend system overview

### External Resources
- [Fiber Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

### Communication
- Daily standups: 9:00 AM
- Weekly planning: Mondays
- Code reviews: Ongoing
- Deployment windows: Fridays 6 PM - 8 PM

---

**Last Updated**: December 23, 2025
**Version**: 1.0
**Status**: Ready for Implementation
**Next Step**: Begin Phase 1 - Project Setup
