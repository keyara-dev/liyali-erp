# Liyali Gateway Backend Documentation

Complete documentation for the Go backend API.

---

## 📖 How to Use This Documentation

This documentation is organized like a book with chapters (folders) and pages (markdown files). Each page focuses on a single topic.

---

## 📚 Chapters

### [Chapter 1: Getting Started](./01-getting-started/)
Setting up your development environment and understanding the project.

- [**Project Overview**](./01-getting-started/overview.md) - What we're building
- [**Setup Guide**](./01-getting-started/setup.md) - Installation and configuration
- [**Project Structure**](./01-getting-started/structure.md) - Directory organization
- [**Development Workflow**](./01-getting-started/workflow.md) - Git, testing, deployment

### [Chapter 2: Authentication & Security](./02-authentication-security/)
User authentication, authorization, and security features.

- [**Authentication Flow**](./02-authentication-security/authentication.md) - JWT login and registration
- [**RBAC System**](./02-authentication-security/rbac.md) - Roles and permissions
- [**User Management**](./02-authentication-security/user-management.md) - User roles and capabilities
- [**Password Security**](./02-authentication-security/passwords.md) - Hashing, reset, change password
- [**Session Management**](./02-authentication-security/sessions.md) - Token lifecycle and revocation
- [**Security Best Practices**](./02-authentication-security/security.md) - Rate limiting, account lockout, CSRF

### [Chapter 3: Database](./03-database/)
PostgreSQL schema, migrations, and sqlc usage.

- [**Schema Design**](./03-database/schema.md) - Database tables and relationships
- [**Migrations**](./03-database/migrations.md) - Managing schema changes
- [**sqlc Usage**](./03-database/sqlc.md) - Type-safe query generation
- [**Repositories**](./03-database/repositories.md) - Data access layer pattern

### [Chapter 4: API Reference](./04-api-reference/)
Complete REST API documentation with request/response examples.

- [**API Overview**](./04-api-reference/overview.md) - Base URL, authentication, error handling
- [**Complete API Reference**](./04-api-reference/complete-reference.md) - All 84 endpoints with examples
  - Authentication (8 endpoints)
  - Approval Tasks (10 endpoints)
  - Workflows (8 endpoints)
  - Documents (7 endpoints)
  - Analytics (3 endpoints)
  - Notifications (7 endpoints)
  - Audit Logs (4 endpoints)

### [Chapter 5: Development](./05-development/)
Development guidelines, roadmap, and best practices.

- [**Development Planner**](./05-development/planner.md) - Implementation phases and progress tracking
- [**Implementation Roadmap**](./05-development/implementation-roadmap.md) - 34-day development plan with 9 phases

### [Chapter 6: Testing](./06-testing/)
Comprehensive testing documentation for unit, integration, and manual tests.

- [**Testing Overview**](./06-testing/overview.md) - Testing strategy, structure, and best practices
  - Unit Tests (28+ test cases)
  - Integration Tests (21+ test cases)
  - Manual Testing Scripts
  - Coverage Goals and CI/CD

### [Chapter 7: Workflows](./07-workflows/)
Workflow system architecture and implementation.

- [**Workflow Overview**](./07-workflows/overview.md) - Workflow types, architecture, and approval process
  - 5 Document Types (Requisition, Budget, PO, Payment Voucher, GRN)
  - Custom Workflow Builder
  - Multi-stage Approval Chains
  - Bulk Operations

---

## 🚀 Quick Navigation

### I want to...

**Set up the project** → [Setup Guide](./01-getting-started/setup.md)

**Understand authentication** → [Authentication Flow](./02-authentication-security/authentication.md)

**Learn about roles and permissions** → [RBAC System](./02-authentication-security/rbac.md)

**Work with the database** → [Schema Design](./03-database/schema.md)

**Build API endpoints** → [Complete API Reference](./04-api-reference/complete-reference.md)

**Understand workflows** → [Workflow Overview](./07-workflows/overview.md)

**Write tests** → [Testing Overview](./06-testing/overview.md)

**Follow the development roadmap** → [Implementation Roadmap](./05-development/implementation-roadmap.md)

---

## 📋 Documentation Standards

Each page in this documentation follows these principles:

1. **Single Focus** - One topic per page
2. **Clear Structure** - Introduction, content, examples
3. **Practical Examples** - Real code snippets
4. **Cross References** - Links to related pages
5. **Up to Date** - Reflects current implementation

---

## 📊 Project Status

### Completed Phases
- ✅ **Phase 1**: Project Setup (Go, PostgreSQL, Fiber, sqlc)
- ✅ **Phase 2**: Database Schema (10 tables with migrations)
- ✅ **Phase 3**: Authentication System (JWT with HS256)
- ✅ **Phase 4**: Core API Endpoints (84 handlers implemented)
- ✅ **Phase 5**: RBAC Implementation (7 roles with middleware)
- ✅ **Phase 6**: Analytics & Reporting (Metrics, trends, bottlenecks)
- ✅ **Phase 7**: Notifications (6 event types with handlers)
- ✅ **Phase 8**: Audit Logging (Complete change tracking)
- ✅ **Phase 9**: Testing (49+ unit & integration tests)

### Current Capabilities
- **84 API Endpoints**: Full REST API for workflow management
- **7 User Roles**: ADMIN, MANAGER, DEPARTMENT_MANAGER, FINANCE_MANAGER, APPROVER, REQUESTER, VIEWER
- **5 Workflow Types**: Multi-stage approval processes
- **Analytics Dashboard**: Real-time metrics and bottleneck analysis
- **Comprehensive Testing**: 28+ unit tests, 21+ integration tests
- **Complete Audit Trail**: All actions logged with JSONB change tracking

---

## 📦 Archive

Historical documentation and deprecated files are available in the [archive](./archive/) folder for reference.

---

## 📞 Support

For detailed information about specific features, navigate to the appropriate chapter above. Each chapter contains focused pages covering specific topics with examples and best practices.

---

**Last Updated**: December 26, 2025
**Version**: 2.0
**Status**: Production Ready
