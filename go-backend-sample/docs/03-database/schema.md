# Database Schema

**PostgreSQL database design for Liyali Gateway**

---

## Overview

Liyali Gateway uses PostgreSQL 12+ for data persistence with 9 core tables organized into authentication, workflow, and audit domains.

---

## Database Tables

### Phase 1: Authentication Tables (✅ Completed)

#### 1. users
**User accounts and authentication**

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN (
        'ADMIN', 'CFO', 'DIRECTOR', 'FINANCE_OFFICER',
        'DEPARTMENT_MANAGER', 'COMPLIANCE_OFFICER', 'REQUESTER'
    )),
    department VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `UNIQUE` on `email`
- `INDEX` on `role`
- `INDEX` on `is_active`

---

#### 2. sessions
**JWT refresh tokens and session management**

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `UNIQUE` on `refresh_token`
- `INDEX` on `user_id`
- `INDEX` on `expires_at`

---

#### 3. password_resets
**Password reset tokens**

```sql
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `UNIQUE` on `token`
- `INDEX` on `user_id`
- `INDEX` on `expires_at, used`

---

### Phase 2: Workflow Tables (🔄 In Progress)

#### 4. approval_tasks
**Tasks requiring approval**

```sql
CREATE TABLE approval_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    assigned_to UUID NOT NULL REFERENCES users(id),
    assigned_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL CHECK (status IN (
        'PENDING', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'REASSIGNED'
    )),
    current_stage INT NOT NULL DEFAULT 1,
    total_stages INT NOT NULL DEFAULT 3,
    priority VARCHAR(20) CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT')),
    due_date TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `INDEX` on `assigned_to, status`
- `INDEX` on `document_id`
- `INDEX` on `status, current_stage`
- `INDEX` on `created_at DESC`

---

#### 5. approval_history
**Audit trail for approvals**

```sql
CREATE TABLE approval_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES approval_tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL CHECK (action IN (
        'APPROVED', 'REJECTED', 'REASSIGNED', 'COMMENTED'
    )),
    stage INT NOT NULL,
    comment TEXT,
    signature TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `INDEX` on `task_id, created_at DESC`
- `INDEX` on `user_id`

---

#### 6. documents
**Base document storage (all workflow types)**

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type VARCHAR(50) NOT NULL CHECK (document_type IN (
        'REQUISITION', 'BUDGET', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER', 'GRN'
    )),
    document_number VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) NOT NULL CHECK (status IN (
        'DRAFT', 'SUBMITTED', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'COMPLETED'
    )),
    created_by UUID NOT NULL REFERENCES users(id),
    department VARCHAR(100),
    workflow_id UUID REFERENCES workflows(id),
    data JSONB NOT NULL,  -- Type-specific fields
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    submitted_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `UNIQUE` on `document_number`
- `INDEX` on `document_type, status`
- `INDEX` on `created_by`
- `INDEX` on `workflow_id`
- `INDEX` on `created_at DESC`
- `GIN` on `data` (JSONB index)

---

#### 7. workflows
**Workflow definitions and templates**

```sql
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_type VARCHAR(50) NOT NULL,
    stages JSONB NOT NULL,  -- Array of stage definitions
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `INDEX` on `document_type, is_active`
- `INDEX` on `created_by`

---

### Phase 3: System Tables (🔄 In Progress)

#### 8. audit_logs
**System-wide audit trail**

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    changes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `INDEX` on `user_id, created_at DESC`
- `INDEX` on `resource_type, resource_id`
- `INDEX` on `action`
- `INDEX` on `created_at DESC`

---

#### 9. notifications
**Email and in-app notifications**

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN (
        'TASK_ASSIGNED', 'TASK_APPROVED', 'TASK_REJECTED',
        'TASK_REASSIGNED', 'TASK_COMMENTED', 'TASK_DUE_SOON'
    )),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    related_id UUID,
    is_read BOOLEAN DEFAULT false,
    sent_via_email BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `PRIMARY KEY` on `id`
- `INDEX` on `user_id, is_read`
- `INDEX` on `created_at DESC`
- `INDEX` on `related_id`

---

## Entity Relationships

```
┌─────────┐
│  users  │
└────┬────┘
     │
     ├──── sessions (1:N)
     ├──── password_resets (1:N)
     ├──── approval_tasks (assigned_to) (1:N)
     ├──── approval_tasks (assigned_by) (1:N)
     ├──── approval_history (1:N)
     ├──── documents (created_by) (1:N)
     ├──── workflows (created_by) (1:N)
     ├──── audit_logs (1:N)
     └──── notifications (1:N)

┌──────────────┐
│  workflows   │
└──────┬───────┘
       │
       └──── documents (1:N)

┌─────────────┐
│  documents  │
└──────┬──────┘
       │
       └──── approval_tasks (1:N)

┌────────────────┐
│ approval_tasks │
└───────┬────────┘
        │
        └──── approval_history (1:N)
```

---

## Data Types

### UUID
- Primary keys and foreign keys
- Generated with `gen_random_uuid()`

### VARCHAR
- Text with maximum length
- Used for emails, names, roles

### TEXT
- Unlimited text length
- Used for comments, descriptions

### DECIMAL(15,2)
- Precise decimal for money
- 15 digits total, 2 after decimal

### JSONB
- Binary JSON storage
- Efficient indexing and querying
- Used for flexible schema fields

### TIMESTAMP
- Date and time with timezone
- Defaults to `CURRENT_TIMESTAMP`

---

## Constraints

### Primary Keys
- All tables use UUID primary keys
- Automatically generated

### Foreign Keys
- Cascade deletes where appropriate
- Maintain referential integrity

### Check Constraints
- Enforce valid enum values
- Validate role names
- Ensure status transitions

### Unique Constraints
- Email uniqueness
- Document number uniqueness
- Token uniqueness

---

## Triggers

### Auto-update Timestamps

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

## Migrations

### Migration Files

Located in `internal/database/migrations/`:

- `001_initial_schema.up.sql` - Create auth tables (users, sessions, password_resets)
- `001_initial_schema.down.sql` - Drop auth tables
- `002_workflow_tables.up.sql` - Create workflow tables
- `002_workflow_tables.down.sql` - Drop workflow tables
- `003_system_tables.up.sql` - Create system tables
- `003_system_tables.down.sql` - Drop system tables

### Running Migrations

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status
```

---

## Sample Data

### Users

```sql
-- Admin user
INSERT INTO users (email, password_hash, name, role, is_active)
VALUES ('admin@example.com', '$2a$12$...', 'System Admin', 'ADMIN', true);

-- Department Manager
INSERT INTO users (email, password_hash, name, role, department, is_active)
VALUES ('manager@example.com', '$2a$12$...', 'John Manager', 'DEPARTMENT_MANAGER', 'Finance', true);
```

### Workflows

```sql
INSERT INTO workflows (name, document_type, stages)
VALUES (
    'Standard Requisition Approval',
    'REQUISITION',
    '[
        {"stage": 1, "role": "DEPARTMENT_MANAGER", "name": "Department Approval"},
        {"stage": 2, "role": "FINANCE_OFFICER", "name": "Budget Verification"},
        {"stage": 3, "role": "CFO", "name": "Final Approval"}
    ]'::jsonb
);
```

---

## Performance Considerations

### Indexes
- Index foreign keys
- Index commonly queried columns
- Index status and date fields
- Use GIN indexes for JSONB

### Partitioning
- Consider partitioning `audit_logs` by date
- Consider partitioning `approval_history` by date

### Archiving
- Archive old completed documents
- Clean up expired sessions and reset tokens

---

## Related Pages

- [Migrations](./migrations.md) - Managing schema changes
- [sqlc Usage](./sqlc.md) - Type-safe queries
- [Repositories](./repositories.md) - Data access layer

---

**Files**:
- `internal/database/migrations/*.sql` - Migration files
- `internal/database/queries/*.sql` - sqlc query files
- `internal/db/models.go` - Generated models

**Last Updated**: December 25, 2025
