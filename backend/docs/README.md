# Liyali Gateway Backend Documentation

A comprehensive procurement management system built with Go, Fiber, PostgreSQL, and GORM.

## 🏗️ Architecture Overview

The Liyali Gateway Backend is built using **Clean Architecture** principles with a hybrid database approach combining the best of GORM and sqlc for optimal performance and type safety.

### Core Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Handlers      │    │   Services      │    │  Repositories   │
│  (HTTP Layer)   │───▶│ (Business Logic)│───▶│ (Data Access)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                               ┌─────────────────┐
                                               │   Database      │
                                               │  (PostgreSQL)   │
                                               └─────────────────┘
```

### Key Features

- **Multi-Tenant Architecture** - Complete organization isolation
- **Advanced RBAC** - 50+ granular permissions with custom roles
- **Hybrid Database** - GORM + sqlc for optimal performance
- **Generic Document System** - Unified search across all document types
- **Workflow Engine** - Dynamic approval workflows
- **Real-time Sync** - Database triggers ensure data consistency
- **Comprehensive Audit** - Full audit trail for compliance

## 📚 Documentation Structure

### Getting Started
- [Quick Start Guide](./01-quick-start.md) - Get up and running in 5 minutes
- [Installation](./02-installation.md) - Detailed setup instructions
- [Configuration](./03-configuration.md) - Environment and database setup

### Architecture & Design
- [System Architecture](./04-architecture.md) - Detailed architecture overview
- [Database Design](./05-database.md) - Schema and relationships
- [API Design](./06-api-design.md) - RESTful API principles

### Core Features
- [Authentication & Authorization](./07-auth.md) - Multi-tenant auth system
- [Document Management](./08-documents.md) - Document lifecycle and operations
- [Workflow Engine](./09-workflows.md) - Dynamic approval workflows
- [Search & Analytics](./10-search.md) - Cross-document search and analytics

### Development
- [Development Guide](./11-development.md) - Local development setup
- [Testing](./12-testing.md) - Unit and integration testing
- [API Reference](./13-api-reference.md) - Complete API documentation

### Deployment & Operations
- [Deployment Guide](./14-deployment.md) - Production deployment
- [Monitoring](./15-monitoring.md) - Health checks and metrics
- [Troubleshooting](./16-troubleshooting.md) - Common issues and solutions

## 🚀 Quick Start

```bash
# Clone and setup
git clone <repository-url>
cd liyali-gateway/backend

# Install dependencies
go mod download

# Setup database
createdb liyali_gateway
psql -d liyali_gateway -f database/migrations/*.sql

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Run the application
go run main.go
```

## 🔧 Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Fiber v2
- **Database**: PostgreSQL 14+
- **ORM**: GORM v2 + sqlc
- **Authentication**: JWT + Sessions
- **Testing**: Go testing + Testify
- **Documentation**: OpenAPI 3.0

## 📊 System Capabilities

### Document Types Supported
- **Requisitions** - Purchase requests with approval workflow
- **Budgets** - Budget allocation and tracking
- **Purchase Orders** - Vendor purchase orders
- **Payment Vouchers** - Payment processing
- **Goods Received Notes** - Inventory receiving
- **Categories** - Item categorization
- **Vendors** - Supplier management

### Key Metrics
- **50+ API Endpoints** - Comprehensive REST API
- **Multi-tenant** - Unlimited organizations
- **RBAC** - 50+ granular permissions
- **Real-time** - Database triggers for instant sync
- **Scalable** - Handles high-volume operations

## 🤝 Contributing

Please read our [Development Guide](./11-development.md) for details on our code of conduct and the process for submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

- **Documentation**: Check the docs in this folder
- **Issues**: Create an issue in the repository
- **API Reference**: See [API Reference](./13-api-reference.md)
- **Troubleshooting**: See [Troubleshooting Guide](./16-troubleshooting.md)