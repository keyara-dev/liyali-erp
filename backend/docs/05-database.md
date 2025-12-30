# Database Design

Comprehensive overview of the PostgreSQL database schema, relationships, design patterns, and the new production-ready bootstrap system used in the Liyali Gateway Backend.

## Database Overview

The system uses **PostgreSQL 14+** with a hybrid approach combining GORM models with sqlc for type-safe queries. The database is designed for multi-tenancy with complete organization isolation and includes an advanced bootstrap system for reliable initialization.

## Bootstrap System

### Overview

The Liyali Gateway now includes a production-ready bootstrap system that solves the race condition between database migrations and seeding operations. This system ensures proper initialization order and provides comprehensive observability.

### Bootstrap Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Bootstrapper  │────│    Validator    │────│     Seeder      │
│                 │    │                 │    │                 │
│ • Phase Control │    │ • Schema Check  │    │ • UPSERT Ops    │
│ • Error Handling│    │ • Constraint    │    │ • Transactions  │
│ • Metrics       │    │ • Index Verify  │    │ • Dependency    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ Circuit Breaker │
                    │                 │
                    │ • Failure Track │
                    │ • Auto Recovery │
                    │ • State Machine │
                    └─────────────────┘
```

### Bootstrap Phases

The bootstrap system executes in a strict phase order:

1. **Connect**: Validate database connection and pool health
2. **Validate**: Check database readiness and PostgreSQL version
3. **Migrate**: Verify all required tables exist
4. **Verify**: Perform comprehensive schema integrity checks
5. **Seed**: Create initial data using idempotent UPSERT operations

### Key Features

- **Idempotent Operations**: Uses PostgreSQL `ON CONFLICT DO UPDATE` for safe re-runs
- **Circuit Breaker Protection**: Prevents cascading failures during startup
- **Retry Logic**: Exponential backoff with jitter for transient failures
- **Transaction Safety**: Atomic operations with rollback support
- **Comprehensive Validation**: Table, column, constraint, and index checks
- **Production Observability**: Detailed logging, timing, and metrics

### Bootstrap Configuration

```go
type BootstrapConfig struct {
    Environment          string        // development, staging, production
    SkipSeeding         bool          // Skip seeding in production
    SeedRetryAttempts   int           // Number of retry attempts
    SeedRetryDelay      time.Duration // Base delay between retries
    CircuitBreakerConfig circuit.Config // Circuit breaker settings
    ValidationTimeout   time.Duration // Timeout for validation phase
    MigrationTimeout    time.Duration // Timeout for migration phase
}
```

### Idempotent Seeding

The seeding system uses PostgreSQL's `ON CONFLICT` clause for true idempotency:

```sql
-- Example: User seeding with UPSERT
INSERT INTO users (id, email, name, role, active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (email) 
DO UPDATE SET 
    name = EXCLUDED.name,
    role = EXCLUDED.role,
    active = EXCLUDED.active,
    updated_at = EXCLUDED.updated_at;
```

### Migration Management

The system includes comprehensive migration management:

```bash
# Linux/Mac migration scripts
cd database && ./migrate.sh up      # Run UP migration
cd database && ./migrate.sh down    # Run DOWN migration  
cd database && ./migrate.sh reset   # Drop and recreate everything
cd database && ./migrate.sh drop    # Emergency drop all tables

# Windows migration scripts
cd database && migrate.bat up       # Run UP migration
cd database && migrate.bat down     # Run DOWN migration
cd database && migrate.bat reset    # Drop and recreate everything
cd database && migrate.bat drop     # Emergency drop all tables
```

## Schema Architecture

### Core Design Principles

1. **Multi-Tenant Isolation** - Every table includes `organization_id` for data separation
2. **Audit Trail** - All entities track creation and modification timestamps
3. **Soft Deletes** - Data is marked as deleted rather than physically removed
4. **JSONB Storage** - Flexible data storage for complex structures
5. **Referential Integrity** - Foreign key constraints ensure data consistency

### Database Structure

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Organizations  │    │     Users       │    │   User Sessions │
│                 │◄───┤                 │───►│                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Requisitions  │    │    Budgets      │    │ Purchase Orders │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│Payment Vouchers │    │      GRNs       │    │   Categories    │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    Vendors      │    │   Workflows     │    │   Documents     │
│                 │    │                 │    │   (Generic)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Tables

### Organizations Table

```sql
CREATE TABLE organizations (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_organizations_active ON organizations(is_active);
CREATE INDEX idx_organizations_deleted ON organizations(deleted_at);
```

**Purpose**: Multi-tenant organization management
**Key Features**:
- Unique organization identifier
- JSONB settings for flexible configuration
- Soft delete support
- Activity status tracking

### Users Table

```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'requester',
    active BOOLEAN DEFAULT true,
    last_login TIMESTAMP WITH TIME ZONE,
    current_organization_id VARCHAR(255),
    is_super_admin BOOLEAN DEFAULT false,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (current_organization_id) REFERENCES organizations(id)
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_organization ON users(current_organization_id);
CREATE INDEX idx_users_active ON users(active);
```

**Purpose**: User authentication and profile management
**Key Features**:
- Multi-organization user support
- Role-based access control
- Encrypted password storage
- User preferences in JSONB

### User Sessions Table

```sql
CREATE TABLE user_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(token);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
```

**Purpose**: Session management for enhanced security
**Key Features**:
- JWT session tracking
- Session expiration management
- IP and user agent tracking
- Automatic cleanup of expired sessions

## Business Entity Tables

### Requisitions Table

```sql
CREATE TABLE requisitions (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    req_number VARCHAR(100) UNIQUE NOT NULL,
    requester_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    department VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    items JSONB NOT NULL DEFAULT '[]',
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB DEFAULT '[]',
    category_id VARCHAR(255),
    preferred_vendor_id VARCHAR(255),
    is_estimate BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (requester_id) REFERENCES users(id),
    FOREIGN KEY (category_id) REFERENCES categories(id),
    FOREIGN KEY (preferred_vendor_id) REFERENCES vendors(id)
);

-- Indexes
CREATE INDEX idx_requisitions_organization ON requisitions(organization_id);
CREATE INDEX idx_requisitions_requester ON requisitions(requester_id);
CREATE INDEX idx_requisitions_status ON requisitions(status);
CREATE INDEX idx_requisitions_department ON requisitions(department);
CREATE INDEX idx_requisitions_created ON requisitions(created_at);
```

**Purpose**: Purchase requisition management
**Key Features**:
- JSONB items storage for flexible item structure
- Approval workflow integration
- Multi-currency support
- Category and vendor relationships

### Budgets Table

```sql
CREATE TABLE budgets (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    budget_code VARCHAR(100) NOT NULL,
    department VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    fiscal_year VARCHAR(10) NOT NULL,
    total_budget DECIMAL(15,2) NOT NULL,
    allocated_amount DECIMAL(15,2) DEFAULT 0,
    remaining_amount DECIMAL(15,2) DEFAULT 0,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- Indexes
CREATE INDEX idx_budgets_organization ON budgets(organization_id);
CREATE INDEX idx_budgets_owner ON budgets(owner_id);
CREATE INDEX idx_budgets_code ON budgets(budget_code);
CREATE INDEX idx_budgets_fiscal_year ON budgets(fiscal_year);
```

**Purpose**: Budget allocation and tracking
**Key Features**:
- Budget code management
- Fiscal year tracking
- Allocation and remaining amount calculation
- Department-based budgeting

### Purchase Orders Table

```sql
CREATE TABLE purchase_orders (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    po_number VARCHAR(100) UNIQUE NOT NULL,
    vendor_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    items JSONB NOT NULL DEFAULT '[]',
    total_amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    delivery_date DATE,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB DEFAULT '[]',
    linked_requisition VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (vendor_id) REFERENCES vendors(id),
    FOREIGN KEY (linked_requisition) REFERENCES requisitions(id)
);

-- Indexes
CREATE INDEX idx_purchase_orders_organization ON purchase_orders(organization_id);
CREATE INDEX idx_purchase_orders_vendor ON purchase_orders(vendor_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_delivery ON purchase_orders(delivery_date);
```

**Purpose**: Purchase order management
**Key Features**:
- Vendor relationship management
- Requisition linking
- Delivery date tracking
- JSONB items with flexible structure

## Generic Document System

### Documents Table

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    document_number VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    amount DECIMAL(15,2),
    currency VARCHAR(10) DEFAULT 'USD',
    department VARCHAR(255),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    workflow_id UUID,
    data JSONB NOT NULL DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE SET NULL,
    CHECK (document_type IN ('REQUISITION', 'BUDGET', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER', 'GRN', 'CATEGORY', 'VENDOR')),
    CHECK (status IN ('draft', 'submitted', 'approved', 'rejected', 'cancelled')),
    CHECK (amount IS NULL OR amount >= 0)
);

-- Performance indexes
CREATE INDEX idx_documents_organization_id ON documents(organization_id);
CREATE INDEX idx_documents_document_type ON documents(document_type);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_created_by ON documents(created_by);
CREATE INDEX idx_documents_department ON documents(department);
CREATE INDEX idx_documents_created_at ON documents(created_at);
CREATE INDEX idx_documents_deleted_at ON documents(deleted_at);

-- Composite indexes for common queries
CREATE INDEX idx_documents_org_type ON documents(organization_id, document_type);
CREATE INDEX idx_documents_org_status ON documents(organization_id, status);

-- JSONB indexes for fast searches
CREATE INDEX idx_documents_data_gin ON documents USING GIN(data);
CREATE INDEX idx_documents_metadata_gin ON documents USING GIN(metadata);

-- Full-text search index
CREATE INDEX idx_documents_search ON documents USING GIN(
    to_tsvector('english', 
        COALESCE(title, '') || ' ' || 
        COALESCE(description, '') || ' ' || 
        COALESCE(document_number, '') || ' ' || 
        COALESCE(department, '')
    )
);
```

**Purpose**: Unified document search and analytics
**Key Features**:
- Cross-document type search
- JSONB storage for type-specific data
- Full-text search capabilities
- Automatic document number generation
- Workflow integration

## Workflow System Tables

### Workflows Table

```sql
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_type VARCHAR(50) NOT NULL,
    stages JSONB NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Indexes
CREATE INDEX idx_workflows_organization ON workflows(organization_id);
CREATE INDEX idx_workflows_document_type ON workflows(document_type);
CREATE INDEX idx_workflows_active ON workflows(is_active);
```

**Purpose**: Dynamic workflow configuration
**Key Features**:
- JSONB stages for flexible workflow definition
- Document type association
- Active/inactive workflow management
- Organization-specific workflows

### Approval Tasks Table

```sql
CREATE TABLE approval_tasks (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    document_id VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    approver_id VARCHAR(255),
    assigned_to VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    stage INTEGER NOT NULL,
    comments TEXT,
    signature VARCHAR(255),
    approved_by VARCHAR(255),
    approved_at TIMESTAMP WITH TIME ZONE,
    rejected_by VARCHAR(255),
    rejected_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (approver_id) REFERENCES users(id),
    FOREIGN KEY (assigned_to) REFERENCES users(id),
    FOREIGN KEY (approved_by) REFERENCES users(id),
    FOREIGN KEY (rejected_by) REFERENCES users(id)
);

-- Indexes
CREATE INDEX idx_approval_tasks_organization ON approval_tasks(organization_id);
CREATE INDEX idx_approval_tasks_document ON approval_tasks(document_id);
CREATE INDEX idx_approval_tasks_assigned ON approval_tasks(assigned_to);
CREATE INDEX idx_approval_tasks_status ON approval_tasks(status);
CREATE INDEX idx_approval_tasks_stage ON approval_tasks(stage);
```

**Purpose**: Approval workflow task management
**Key Features**:
- Document-agnostic approval tasks
- Stage-based approval process
- Digital signature support
- Comprehensive audit trail

## RBAC System Tables

### Organization Roles Table

```sql
CREATE TABLE organization_roles (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT false,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    UNIQUE(organization_id, name)
);

-- Indexes
CREATE INDEX idx_organization_roles_org ON organization_roles(organization_id);
CREATE INDEX idx_organization_roles_system ON organization_roles(is_system_role);
```

**Purpose**: Custom organization-specific roles
**Key Features**:
- JSONB permissions for flexible permission assignment
- System vs custom role distinction
- Organization-specific role management

### User Organization Roles Table

```sql
CREATE TABLE user_organization_roles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_by VARCHAR(255),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES organization_roles(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES users(id),
    UNIQUE(user_id, organization_id, role_id)
);

-- Indexes
CREATE INDEX idx_user_org_roles_user ON user_organization_roles(user_id);
CREATE INDEX idx_user_org_roles_org ON user_organization_roles(organization_id);
CREATE INDEX idx_user_org_roles_role ON user_organization_roles(role_id);
```

**Purpose**: User role assignments within organizations
**Key Features**:
- Many-to-many relationship between users and roles
- Organization-scoped role assignments
- Assignment tracking and audit

## Data Synchronization

### Database Triggers

The system uses PostgreSQL triggers to maintain data consistency between specific document tables and the generic documents table:

```sql
-- Example trigger for requisitions
CREATE OR REPLACE FUNCTION sync_requisition_to_document()
RETURNS TRIGGER AS $$
DECLARE
    doc_number TEXT;
    doc_data JSONB;
BEGIN
    -- Generate document number
    doc_number := generate_document_number('REQUISITION', NEW.id);
    
    -- Build document data JSONB
    doc_data := jsonb_build_object(
        'id', NEW.id,
        'reqNumber', NEW.req_number,
        'items', COALESCE(NEW.items, '[]'::jsonb),
        'priority', NEW.priority,
        'categoryId', NEW.category_id,
        'preferredVendorId', NEW.preferred_vendor_id,
        'isEstimate', COALESCE(NEW.is_estimate, false),
        'approvalStage', COALESCE(NEW.approval_stage, 0),
        'approvalHistory', COALESCE(NEW.approval_history, '[]'::jsonb)
    );
    
    -- Insert or update generic document
    INSERT INTO documents (...)
    VALUES (...)
    ON CONFLICT (document_number) 
    DO UPDATE SET ...;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_sync_requisition
    AFTER INSERT OR UPDATE ON requisitions
    FOR EACH ROW
    EXECUTE FUNCTION sync_requisition_to_document();
```

**Benefits**:
- **Automatic Synchronization** - No manual sync required
- **Real-time Consistency** - Changes reflected immediately
- **Zero Application Changes** - Works transparently
- **Performance Optimized** - Minimal overhead

## Database Performance

### Indexing Strategy

#### 1. Primary Indexes
- All primary keys are indexed automatically
- Foreign keys have dedicated indexes
- Frequently queried columns have single-column indexes

#### 2. Composite Indexes
```sql
-- Multi-column indexes for common query patterns
CREATE INDEX idx_requisitions_org_status ON requisitions(organization_id, status);
CREATE INDEX idx_approval_tasks_assigned_status ON approval_tasks(assigned_to, status);
CREATE INDEX idx_documents_org_type_status ON documents(organization_id, document_type, status);
```

#### 3. JSONB Indexes
```sql
-- GIN indexes for JSONB columns
CREATE INDEX idx_requisitions_items_gin ON requisitions USING GIN(items);
CREATE INDEX idx_documents_data_gin ON documents USING GIN(data);
CREATE INDEX idx_workflows_stages_gin ON workflows USING GIN(stages);
```

#### 4. Full-Text Search Indexes
```sql
-- Full-text search across multiple columns
CREATE INDEX idx_documents_fulltext ON documents USING GIN(
    to_tsvector('english', title || ' ' || COALESCE(description, ''))
);
```

### Query Optimization

#### 1. Efficient Pagination
```sql
-- Optimized pagination with offset/limit
SELECT * FROM requisitions 
WHERE organization_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- Count query optimization
SELECT COUNT(*) FROM requisitions 
WHERE organization_id = $1;
```

#### 2. JSONB Query Optimization
```sql
-- Efficient JSONB queries
SELECT * FROM requisitions 
WHERE organization_id = $1 
  AND items @> '[{"category": "electronics"}]';

-- JSONB path queries
SELECT * FROM documents 
WHERE data->>'priority' = 'high' 
  AND document_type = 'REQUISITION';
```

## Data Migration

### Migration Strategy

#### 1. Schema Migrations
```sql
-- Migration files are numbered sequentially
001_initial_schema.sql
002_enhanced_auth.sql
003_workflows.sql
008_create_documents_table.sql
009_add_document_sync_triggers.sql
```

#### 2. Data Migration
```sql
-- One-time data sync function
SELECT sync_existing_documents();

-- Verify migration
SELECT * FROM document_sync_status;
```

### Backup and Recovery

#### 1. Backup Strategy
```bash
# Full database backup
pg_dump -h localhost -U postgres -d liyali_gateway > backup.sql

# Schema-only backup
pg_dump -h localhost -U postgres -d liyali_gateway --schema-only > schema.sql

# Data-only backup
pg_dump -h localhost -U postgres -d liyali_gateway --data-only > data.sql
```

#### 2. Point-in-Time Recovery
```bash
# Enable WAL archiving in postgresql.conf
archive_mode = on
archive_command = 'cp %p /path/to/archive/%f'

# Create base backup
pg_basebackup -D /path/to/backup -Ft -z -P

# Restore to specific point in time
pg_ctl stop -D /path/to/data
rm -rf /path/to/data/*
tar -xzf /path/to/backup/base.tar.gz -C /path/to/data
# Configure recovery.conf and restart
```

## Database Monitoring

### Performance Monitoring

#### 1. Query Performance
```sql
-- Slow query monitoring
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Index usage statistics
SELECT schemaname, tablename, indexname, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_tup_read DESC;
```

#### 2. Connection Monitoring
```sql
-- Active connections
SELECT count(*) as active_connections
FROM pg_stat_activity
WHERE state = 'active';

-- Connection by database
SELECT datname, count(*) as connections
FROM pg_stat_activity
GROUP BY datname;
```

#### 3. Table Statistics
```sql
-- Table sizes
SELECT schemaname, tablename, 
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Table activity
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del
FROM pg_stat_user_tables
ORDER BY n_tup_ins + n_tup_upd + n_tup_del DESC;
```

## Next Steps

- **API Design**: Understand [API Patterns](./06-api-design.md)
- **Authentication**: Review [Auth System](./07-auth.md)
- **Document Management**: Explore [Document Operations](./08-documents.md)
- **Development**: Set up [Development Environment](./11-development.md)