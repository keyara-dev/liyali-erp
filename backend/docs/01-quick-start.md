# Quick Start Guide

Get the Liyali Gateway Backend up and running in 5 minutes.

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- Git

## 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd liyali-gateway/backend

# Install Go dependencies
go mod download
```

## 2. Database Setup

```bash
# Create database
createdb liyali_gateway

# Run migrations
psql -d liyali_gateway -f database/migrations/001_initial_schema.sql
psql -d liyali_gateway -f database/migrations/002_enhanced_auth.sql
psql -d liyali_gateway -f database/migrations/003_workflows.sql
psql -d liyali_gateway -f database/migrations/008_create_documents_table.sql
psql -d liyali_gateway -f database/migrations/009_add_document_sync_triggers.sql

# Initialize data sync (one-time)
psql -d liyali_gateway -c "SELECT sync_existing_documents();"
```

## 3. Environment Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

**Required Environment Variables:**
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=liyali_gateway
DB_SSL_MODE=disable

# Application
APP_PORT=8080
JWT_SECRET=your-super-secret-jwt-key-change-in-production
FRONTEND_URL=http://localhost:3000

# Optional
APP_ENV=development
```

## 4. Run the Application

```bash
# Start the server
go run main.go
```

You should see:
```
🚀 Starting Enhanced Liyali Gateway Backend on port 8080
📊 Features: Enhanced Auth, Session Management, Custom RBAC, Workflow Engine
🔐 Security: Account Lockout, Password Reset, Audit Logging
🏗️  Architecture: Clean Architecture (Repository → Service → Handler)
💾 Database: GORM + pgx with sqlc for type-safe queries
🔄 Workflows: Dynamic workflow management with bulk operations
❤️  Health check: http://localhost:8080/health
```

## 5. Verify Installation

### Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "liyali-gateway-backend"
}
```

### Create First Organization
```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "description": "Test organization"
  }'
```

### Register First User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@company.com",
    "password": "SecurePassword123!",
    "name": "Admin User",
    "organizationId": "org-id-from-previous-step"
  }'
```

## 6. Test Key Features

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@company.com",
    "password": "SecurePassword123!"
  }'
```

### Create a Requisition
```bash
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Office Supplies",
    "description": "Monthly office supplies order",
    "items": [
      {
        "name": "Laptop",
        "quantity": 2,
        "unitPrice": 1200.00,
        "totalPrice": 2400.00
      }
    ],
    "totalAmount": 2400.00,
    "priority": "medium",
    "department": "IT"
  }'
```

### Search Documents
```bash
curl -X GET "http://localhost:8080/api/v1/documents/search?q=laptop" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 7. Access API Documentation

The API is documented using OpenAPI 3.0. You can:

1. **View OpenAPI Spec**: `backend/openapi.yaml`
2. **Test Endpoints**: Use the provided `backend/API.http` file with REST Client
3. **Read API Reference**: See [API Reference](./13-api-reference.md)

## Next Steps

- **Development**: Read the [Development Guide](./11-development.md)
- **Architecture**: Understand the [System Architecture](./04-architecture.md)
- **API Reference**: Explore the [Complete API Documentation](./13-api-reference.md)
- **Testing**: Set up [Testing Environment](./12-testing.md)

## Troubleshooting

### Common Issues

**Database Connection Error**
```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql -h localhost -U postgres -d liyali_gateway
```

**Port Already in Use**
```bash
# Change port in .env file
APP_PORT=8081

# Or kill process using port 8080
lsof -ti:8080 | xargs kill -9
```

**Migration Errors**
```bash
# Check database exists
psql -l | grep liyali_gateway

# Run migrations individually
psql -d liyali_gateway -f database/migrations/001_initial_schema.sql
```

For more troubleshooting, see [Troubleshooting Guide](./16-troubleshooting.md).

## What's Next?

You now have a fully functional Liyali Gateway Backend! The system includes:

- ✅ **Multi-tenant authentication** with JWT and sessions
- ✅ **Complete document management** with approval workflows
- ✅ **Advanced RBAC** with 50+ permissions
- ✅ **Generic document search** across all document types
- ✅ **Real-time data synchronization** with database triggers
- ✅ **Comprehensive audit logging** for compliance

Explore the other documentation files to learn more about specific features and advanced configuration options.