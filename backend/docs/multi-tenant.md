# Multi-Tenant System

Complete organization isolation with proper data separation and super admin access.

## Architecture

- **Middleware**: `middleware/tenant.go` - Organization context validation
- **Seeder**: `database/seeders/multi_tenant_seeder.go` - Test data generation
- **Verification**: `cmd/verify-separation/main.go` - Data separation testing

## Organizations

- **Demo Corporation** (`org-demo-001`) - Enterprise tier
- **ACME Corporation** (`org-acme-001`) - Enterprise tier
- **Default Organization** (`org-default-001`) - Starter tier

## Super Admin Access

The super admin (`admin@liyali.com`) has memberships in all organizations and can switch between them while maintaining complete data isolation.

## Commands

```bash
# Seed multi-tenant data
go run cmd/seed/main.go --multi-tenant

# Verify data separation
go run cmd/verify-separation/main.go

# Cleanup test data
go run cmd/seed/main.go --cleanup
```

## Data Isolation

Each organization has completely separate:

- Users and memberships
- Requisitions and workflows
- Categories and budgets
- Document numbering (DEMO-REQ-xxx vs ACME-REQ-xxx)
