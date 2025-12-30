# Database Migrations

This directory contains SQL migration files for the Liyali Gateway database schema.

## Migration Files

### Current Migrations

1. **000_drop_all_tables.up.sql** - Emergency script to drop all tables
2. **001_create_complete_schema.up.sql** - Complete database schema creation
3. **001_create_complete_schema.down.sql** - Complete schema rollback

## Migration System Overview

The migration system has been cleaned up and simplified:

- **Single Comprehensive Migration**: Instead of multiple incremental migrations, we now have one comprehensive migration that creates the entire schema
- **Proper Dependency Order**: All tables are created in the correct dependency order
- **Global Vendors**: Vendors table is now global (not tied to organizations)
- **Enhanced Authentication**: Includes session management, password resets, email verification, login attempts, and account lockouts
- **Workflow System**: Complete workflow and approval system with enhanced features
- **Performance Optimized**: Includes all necessary indexes and constraints

## Database Schema Structure

### Core Tables (No Dependencies)
- `users` - System users
- `organizations` - Tenant/workspace organizations

### Organization Related Tables
- `organization_settings` - Per-organization configuration
- `organization_members` - User-organization relationships
- `organization_departments` - Organizational departments

### Enhanced Authentication Tables
- `sessions` - User session management with refresh tokens
- `password_resets` - Password reset tokens
- `email_verifications` - Email verification tokens
- `login_attempts` - Security tracking for login attempts
- `account_lockouts` - Account lockout management
- `organization_roles` - Custom roles within organizations
- `user_organization_roles` - User role assignments

### Workflow System Tables
- `workflows` - Workflow definitions
- `approval_tasks_enhanced` - Enhanced approval tasks with workflow support
- `approval_history` - Audit trail for approval actions
- `notifications_enhanced` - Enhanced notification system

### Master Data Tables
- `vendors` - **Global vendors** (accessible to all organizations)
- `categories` - Organization-specific categories
- `category_budget_codes` - Category-budget code relationships

### Business Document Tables
- `requisitions` - Purchase requisitions
- `budgets` - Budget management
- `purchase_orders` - Purchase orders
- `payment_vouchers` - Payment vouchers
- `goods_received_notes` - Goods received notes

### Legacy Compatibility Tables
- `approval_tasks` - Legacy approval tasks (backward compatibility)
- `audit_logs` - System audit logs
- `notifications` - Legacy notifications (backward compatibility)

## Key Changes Made

### 1. Vendor Table Refactoring
- **Before**: Vendors were tied to organizations (`organization_id` column)
- **After**: Vendors are now global and accessible to all organizations
- **Benefit**: Reduces data duplication and allows vendor sharing across organizations

### 2. Migration Cleanup
- **Removed**: Multiple conflicting migration files
- **Consolidated**: Single comprehensive migration
- **Fixed**: Dependency order issues

### 3. Enhanced Authentication
- Added comprehensive session management
- Implemented security features (login attempts, account lockouts)
- Added email verification system
- Enhanced role-based access control (RBAC)

## Running Migrations

### Option 1: Using Migration Scripts (Recommended)

**Linux/Mac:**
```bash
# Run UP migration (create tables)
./migrate.sh up

# Run DOWN migration (drop tables)
./migrate.sh down

# Reset database (drop + create)
./migrate.sh reset

# Drop all tables
./migrate.sh drop
```

**Windows:**
```cmd
# Run UP migration (create tables)
migrate.bat up

# Run DOWN migration (drop tables)
migrate.bat down

# Reset database (drop + create)
migrate.bat reset

# Drop all tables
migrate.bat drop
```

### Option 2: Manual Migration

```bash
# Create all tables
go run database/run_migration.go database/migrations/001_create_complete_schema.up.sql

# Drop all tables
go run database/run_migration.go database/migrations/001_create_complete_schema.down.sql

# Emergency drop (if needed)
go run database/run_migration.go database/migrations/000_drop_all_tables.up.sql
```

## Environment Setup

Ensure your `.env` file contains the database configuration:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=liyali_gateway
DB_SSL_MODE=disable
```

## Migration Best Practices

1. **Always backup your database** before running migrations in production
2. **Test migrations** in a development environment first
3. **Use the reset option** only in development environments
4. **Review the migration files** before running them
5. **Check the logs** for any errors during migration

## Troubleshooting

### Common Issues

1. **Foreign Key Constraint Errors**
   - Ensure tables are created in the correct dependency order
   - Check that referenced tables exist before creating foreign keys

2. **Duplicate Table Errors**
   - Run the DOWN migration first to clean up existing tables
   - Use the reset option to completely recreate the schema

3. **Permission Errors**
   - Ensure the database user has CREATE, DROP, and ALTER permissions
   - Check that the database exists and is accessible

### Recovery Steps

If migrations fail:

1. Check the error message in the console
2. Verify database connectivity and permissions
3. Run the DOWN migration to clean up partial changes
4. Fix any issues and re-run the UP migration

## Schema Validation

After running migrations, verify the schema:

```sql
-- Check all tables exist
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;

-- Check foreign key constraints
SELECT 
    tc.table_name, 
    tc.constraint_name, 
    tc.constraint_type,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM information_schema.table_constraints AS tc 
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
ORDER BY tc.table_name;
```

## Next Steps

After successful migration:

1. **Seed the database** with initial data using the seeding functions
2. **Test the application** to ensure all functionality works
3. **Run integration tests** to verify database operations
4. **Monitor performance** and optimize queries if needed

## Support

If you encounter issues with migrations:

1. Check this README for troubleshooting steps
2. Review the migration logs for specific error messages
3. Verify your environment configuration
4. Test in a clean development environment first