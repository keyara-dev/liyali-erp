# Database Implementation

## Overview

The Liyali Gateway backend uses PostgreSQL with a 100% database-driven architecture. All features are backed by database tables with proper migrations, queries, and SQLC-generated code.

## Key Features

### ✅ Complete Database Integration

- User management with enhanced authentication
- Organization and subscription management
- Role-based access control (RBAC)
- Workflow system with state machines
- Document management
- Audit logging
- System settings and feature flags

### Database Structure

#### Core Tables

- `users` - User accounts with authentication
- `organizations` - Multi-tenant organizations
- `organization_memberships` - User-organization relationships
- `roles` - System and custom roles
- `permissions` - Granular permissions
- `role_permissions` - Role-permission mappings

#### Authentication & Security

- `sessions` - User sessions
- `password_resets` - Password reset tokens
- `login_attempts` - Failed login tracking
- `account_lockouts` - Account security

#### Subscription System

- `subscription_plans` - Available plans
- `organization_subscriptions` - Active subscriptions
- `subscription_features` - Feature definitions
- `subscription_usage` - Usage tracking

#### Admin Features

- `admin_settings` - System configuration
- `admin_feature_flags` - Feature toggles
- `admin_analytics` - System metrics

## Migrations

All migrations are in `backend/database/migrations/`:

- `001_init_system.up.sql` - Core schema
- `002_seed_data.up.sql` - Initial data
- `011_admin_settings_feature_flags.up.sql` - Admin features
- `012_subscription_management_system.up.sql` - Subscriptions
- `013_complete_database_integration.up.sql` - Final integration

## SQLC Queries

All database queries use SQLC for type-safe database access:

```bash
# Generate SQLC code
sqlc generate
```

Query files in `backend/database/queries/`:

- `users_enhanced.sql` - User operations
- `organization_subscription_management.sql` - Subscriptions
- `sessions.sql` - Session management
- `workflows.sql` - Workflow operations

## Testing

Run database tests:

```bash
# Unit tests
go test ./...

# Integration tests
go test ./tests/integration/...

# Database-specific tests
go test ./repository/...
```

## Seeding Data

```bash
# Run seeder
go run cmd/seed/main.go

# Or use make
make seed
```

## Status

✅ 100% database-driven implementation  
✅ All features backed by database  
✅ No mock data in production code  
✅ Full SQLC integration  
✅ Comprehensive migrations  
✅ Type-safe queries

For detailed API documentation, see [API Reference](./13-api-reference.md).
