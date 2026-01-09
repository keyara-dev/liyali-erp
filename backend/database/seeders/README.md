# Multi-Tenant Database Seeding

This directory contains seeders for creating properly separated test data across multiple organizations/workspaces.

## Overview

The multi-tenant seeder ensures that each organization has its own isolated dataset, preventing data leakage between workspaces and providing realistic test scenarios for the multi-tenant functionality.

## Files

- `multi_tenant_seeder.go` - Main multi-tenant seeder with workspace separation
- `approval_test_seeder.go` - Legacy approval test seeder (single organization)
- `README.md` - This documentation file

## Usage

### Quick Start

```bash
# Navigate to backend directory
cd backend

# Seed multi-tenant data with proper workspace separation
make seed-multi-tenant

# Clean up existing multi-tenant test data
make seed-cleanup

# Show seeding help
make seed-help
```

### Manual Commands

```bash
# Seed multi-tenant data
go run cmd/seed/main.go --multi-tenant

# Clean up test data
go run cmd/seed/main.go --cleanup

# Show help
go run cmd/seed/main.go --help
```

## What Gets Created

### Organizations

- **Demo Organization** (`org-demo-001`)
  - Tier: Pro
  - Focus: General business operations
- **ACME Corporation** (`org-acme-001`)
  - Tier: Enterprise
  - Focus: Manufacturing operations

### Users (Per Organization)

Each organization gets its own set of users, plus the super admin has access to both:

**Super Admin (Cross-Organization):**

- `admin@liyali.com` - System Administrator (admin role, super admin)
  - Has memberships in both organizations
  - Can switch between workspaces
  - Default organization: Demo Organization

**Demo Organization:**

- `admin@demo-org.com` - Demo Admin (admin role)
- `manager@demo-org.com` - Demo Manager (manager role)
- `requester@demo-org.com` - Demo Requester (requester role)
- `finance@demo-org.com` - Demo Finance Officer (finance_manager role)

**ACME Corporation:**

- `admin@acme-corp.com` - ACME Admin (admin role)
- `manager@acme-corp.com` - ACME Manager (manager role)
- `requester@acme-corp.com` - ACME Requester (requester role)
- `finance@acme-corp.com` - ACME Finance Officer (finance_manager role)

### Categories (Organization-Specific)

**Demo Organization:**

- Office Supplies
- IT Equipment
- Marketing Materials

**ACME Corporation:**

- Manufacturing Equipment
- Raw Materials
- Safety Equipment

### Sample Documents

**Demo Organization Requisitions:**

- Demo Office Furniture Purchase (pending, $12,000)
- Demo Software Licenses (approved, $6,500)
- Demo Training Materials (draft, $3,500)

**ACME Corporation Requisitions:**

- ACME Production Line Upgrade (pending, $45,000)
- ACME Safety Equipment Renewal (approved, $18,000)
- ACME Raw Materials Stock (rejected, $32,000)
- ACME Quality Control Equipment (draft, $25,000)

**Budgets (Per Organization):**

- Demo: Marketing ($50k), IT ($75k), HR ($30k)
- ACME: Production ($200k), Safety ($80k), Quality ($60k)

## Data Separation Verification

After seeding, you can verify proper data separation by:

1. **Switching Organizations**: Log in and switch between Demo Organization and ACME Corporation
2. **Dashboard Statistics**: Each organization should show different metrics
3. **Document Lists**: Requisitions, budgets, etc. should be different for each organization
4. **User Lists**: Each organization should only see its own users

## Testing Multi-Tenancy

### Login Credentials

All test users use the password: `password`

**Super Admin (Cross-Organization):**

```
admin@liyali.com / password
```

**Demo Organization Users:**

```
admin@demo-org.com / password
manager@demo-org.com / password
requester@demo-org.com / password
finance@demo-org.com / password
```

**ACME Corporation Users:**

```
admin@acme-corp.com / password
manager@acme-corp.com / password
requester@acme-corp.com / password
finance@acme-corp.com / password
```

### Verification Steps

1. **Login as Super Admin (admin@liyali.com)**

   - Should see Demo Organization as default workspace (or last selected)
   - Can switch between both Demo Organization and ACME Corporation
   - Dashboard metrics should change when switching organizations
   - Should see different data sets for each organization

2. **Login as Demo Admin**

   - Should see Demo Organization as current workspace
   - Dashboard should show Demo-specific metrics
   - Requisitions should show DEMO-REQ-001, DEMO-REQ-002, etc.

3. **Switch to ACME Corporation**

   - Dashboard metrics should change
   - Requisitions should show ACME-REQ-001, ACME-REQ-002, etc.
   - Categories should show Manufacturing Equipment, Raw Materials, etc.

4. **Login as ACME Admin**
   - Should see ACME Corporation as current workspace
   - Should not see any Demo Organization data

## Database Schema

The seeder creates data that respects the multi-tenant architecture:

- All entities have `organization_id` fields
- Users belong to organizations via `organization_members` table
- Categories, requisitions, budgets are scoped to organizations
- Workflows (when implemented) will be organization-specific

## Cleanup

To remove all test data and start fresh:

```bash
make seed-cleanup
```

This will remove:

- All test requisitions and budgets
- Organization-specific categories
- Organization memberships
- Test users (but not the organizations themselves)

## Troubleshooting

### Common Issues

1. **"Organization not found" errors**

   - Ensure organizations exist before running seeder
   - Check database connection

2. **Duplicate key errors**

   - Run cleanup first: `make seed-cleanup`
   - Check for existing test data

3. **Permission errors**
   - Ensure database user has CREATE/INSERT permissions
   - Check database connection string

### Logs

The seeder provides detailed logging:

- ✅ Created items
- 📋 Already existing items
- ❌ Errors with details

## Integration with Application

The seeded data integrates with:

- **Authentication System**: Test users can log in
- **Organization Switching**: Users can switch between workspaces
- **Dashboard Analytics**: Each organization shows different metrics
- **Approval Workflows**: Sample requisitions have different approval states
- **RBAC System**: Users have appropriate roles and permissions

## Development Notes

- The seeder is idempotent - safe to run multiple times
- Existing data is updated, not duplicated
- UUIDs are used for all primary keys
- Timestamps reflect realistic creation dates
- Approval histories show realistic workflow progression

## Future Enhancements

- [ ] Add Purchase Orders for each organization
- [ ] Add Payment Vouchers with organization separation
- [ ] Add GRNs (Goods Received Notes) per organization
- [ ] Add organization-specific workflows
- [ ] Add department-specific data
- [ ] Add vendor relationships per organization
