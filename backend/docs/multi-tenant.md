# Multi-Tenant System

Complete organization isolation with proper data separation and super admin access.

## Architecture

- **Middleware**: `middleware/tenant.go` - Organization context validation
- **Seed Data**: `database/migrations/002_consolidated_seed_data.up.sql` - Test data generation
- **Verification**: `cmd/verify-separation/main.go` - Data separation testing

## Organizations

- **Demo Corporation** (`org-demo-001`) - Enterprise tier
- **Default Organization** (`e67fe5b7-dd91-47cb-938b-2b2cd52e10b2`) - Starter tier

## Super Admin Access

The super admin (`admin@liyali.com`) has memberships in all organizations and can switch between them while maintaining complete data isolation.

## Commands

```bash
# Verify data separation
go run cmd/verify-separation/main.go

# Run migrations (includes seed data)
go run main.go -migrate
```

## Data Isolation

Each organization has completely separate:

- Users and memberships
- Requisitions and workflows
- Categories and budgets
- Document numbering and data
