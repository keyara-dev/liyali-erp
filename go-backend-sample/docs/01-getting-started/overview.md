# Project Overview

**What we're building and why**

---

## What is Liyali Gateway?

Liyali Gateway is a high-performance Go backend API for a government-compliant approval workflow system. It replaces localStorage-based storage with a robust PostgreSQL database and provides REST APIs for all workflow operations.

---

## Key Features

### 🔐 Authentication & Security
- Custom JWT-based authentication (email/password)
- Role-Based Access Control (RBAC)
- 7 user roles with 35+ granular permissions
- Password reset flow
- Account lockout after failed login attempts

### 📊 Workflow Management
- Multi-stage approval workflows
- Digital signature support
- 5 workflow types: Requisition, Budget, Purchase Order, Payment Voucher, GRN
- Bulk approval operations
- Task reassignment and delegation

### 📧 Notifications
- Email notifications via SendGrid
- Task assignment alerts
- Approval/rejection notifications

### 📝 Audit & Compliance
- Complete audit trail
- Action logging
- Government-compliant record keeping

### 📈 Analytics
- Real-time dashboard metrics
- 7-day trend analysis
- Bottleneck identification

---

## Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.21+ | Backend language |
| **Framework** | Fiber v3 | HTTP routing |
| **Database** | PostgreSQL 12+ | Data persistence |
| **Query Generator** | sqlc | Type-safe SQL |
| **Authentication** | JWT (HS256) | Token-based auth |
| **Password Hashing** | bcrypt | Secure passwords |
| **Email** | SendGrid | Notifications |
| **Testing** | Testify | Test framework |

---

## Architecture

```
┌─────────────┐
│   Frontend  │ Next.js 16
│  (React)    │
└──────┬──────┘
       │ REST API
       │
┌──────▼───────┐
│  Go Backend  │ Fiber Framework
│   (API)      │
└──────┬───────┘
       │ pgx driver
       │
┌──────▼────────┐
│  PostgreSQL   │ Database
│   Database    │
└───────────────┘
```

### Request Flow

```
HTTP Request
    ↓
[Fiber Router]
    ↓
[Auth Middleware] → Validate JWT → Extract user claims
    ↓
[RBAC Middleware] → Check permissions
    ↓
[Handler] → Validate request → Call service
    ↓
[Service] → Business logic → Call repository
    ↓
[Repository] → Execute SQL via sqlc
    ↓
[PostgreSQL] → Return data
    ↓
[Response] → JSON response
```

---

## Project Goals

### Technical Goals
- ✅ Type-safe database queries with sqlc
- ✅ < 100ms API response time
- ✅ 80%+ test coverage
- ✅ Zero SQL injection vulnerabilities
- ⏳ Horizontal scalability
- ⏳ Production-ready error handling

### Business Goals
- Replace localStorage with persistent database
- Support 1000+ concurrent users
- Ensure government compliance
- Enable email notifications
- Provide complete audit trail
- Deliver real-time analytics

---

## Success Metrics

### Technical
- [x] API response time < 100ms (95th percentile)
- [x] Database query time < 50ms
- [x] Test coverage > 80% (unit tests)
- [x] 0 critical security vulnerabilities
- [ ] 0 high-priority bugs

### Business
- [ ] 100% of frontend server actions migrated
- [ ] All 5 workflow types working
- [ ] Email notifications sending
- [ ] Audit logs complete
- [ ] User acceptance achieved

---

## Current Status

**Phase**: Phase 2 (Database Schema)

**Completed**:
- ✅ Phase 1: Project Setup (Go modules, Fiber, PostgreSQL, sqlc, config)
- ✅ Phase 3: Authentication System (JWT, login, registration, password reset)
- ✅ Phase 5: RBAC Implementation (7 roles, 35+ permissions, middleware)

**In Progress**:
- 🔄 Phase 2: Database Schema (3/9 tables created)

**Next Up**:
- Phase 2: Complete remaining 6 tables
- Phase 4: Core API endpoints
- Phase 6: Email notifications
- Phase 7: Audit logging

---

## Related Pages

- [Setup Guide](./setup.md) - Install and configure
- [Project Structure](./structure.md) - Understand the codebase
- [Development Planner](../05-development/planner.md) - Full roadmap

---

**Last Updated**: December 25, 2025
