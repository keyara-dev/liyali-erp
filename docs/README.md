# Liyali Gateway Documentation

**Enterprise Workflow Management Platform**

A production-ready multi-tenant platform for managing financial workflows with comprehensive authentication, role-based access control, and approval processes.

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

### Setup
```bash
# Backend
cd backend
go mod tidy
go run main.go

# Frontend
cd frontend
npm install
npm run dev
```

## 📚 Documentation Structure

### Essential Guides
- **[Setup Guide](SETUP.md)** - Complete installation and configuration
- **[Authentication Guide](AUTH.md)** - User accounts, roles, and permissions  
- **[API Reference](API.md)** - Complete API documentation
- **[Development Guide](DEVELOPMENT.md)** - Development workflow and testing

### Project Information
- **[Project Status](STATUS.md)** - Current completion status and metrics
- **[Roadmap](ROADMAP.md)** - Project timeline and future plans

## 🏗️ System Overview

### Core Features
- ✅ **Multi-tenant Organizations** - Complete isolation and management
- ✅ **Role-Based Access Control** - 5 system roles + custom roles
- ✅ **Workflow Management** - Requisitions, POs, Payments, Budgets, GRNs
- ✅ **Authentication & Security** - JWT, password hashing, session management
- ✅ **Analytics & Reporting** - Dashboard metrics and insights

### Technology Stack
- **Backend**: Go 1.21+, Fiber, PostgreSQL, SQLC
- **Frontend**: Next.js 14, React 18, TypeScript, TanStack Query
- **Database**: PostgreSQL with comprehensive migrations
- **Testing**: 100+ unit tests, 50+ integration tests

## 📊 Current Status

**Overall Completion**: ~95% of core features complete
- **Backend**: 20,000+ lines of Go code
- **Frontend**: 15,000+ lines of TypeScript/React
- **API Endpoints**: 60+ endpoints
- **Test Coverage**: 85%+ of critical paths

## 🔧 Quick Commands

```bash
# Backend
cd backend && go run main.go

# Frontend  
cd frontend && npm run dev

# Tests
cd backend && go test ./...
cd frontend && npm test

# Database
cd backend/database && ./migrate.sh up
```

## 📞 Support

For detailed information, see the individual guide files in this directory.

---

**Status**: Production Ready | **Last Updated**: January 8, 2025