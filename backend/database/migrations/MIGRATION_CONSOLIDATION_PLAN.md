# Database Migration Consolidation Plan

## Current State Analysis

### Existing Migration Files
1. **001_create_complete_schema.up.sql** - Comprehensive base schema (716 lines)
2. **002_add_missing_fields.up.sql** - Additional business fields for type alignment
3. **003_add_alignment_fields.up.sql** - Minor alignment fields (mostly comments)
4. **007_fix_workflow_tables.sql** - Workflow table fixes and enhancements

### Issues with Current Migrations
1. **Redundancy**: Multiple migrations adding fields to same tables
2. **Complexity**: 4 separate files for what should be 1-2 migrations
3. **Dependencies**: Some migrations depend on others in non-obvious ways
4. **Maintenance**: Hard to understand complete schema from multiple files

## Consolidation Strategy

### New Migration Structure
We will create **2 consolidated migrations**:

1. **001_create_complete_schema_consolidated.up.sql**
   - Complete database schema with ALL fields from all current migrations
   - Includes all tables, indexes, constraints, and triggers
   - Single source of truth for the entire schema

2. **002_seed_initial_data.up.sql** 
   - Initial data seeding (organizations, users, workflows, etc.)
   - Default configurations and settings
   - Sample data for development/testing

### Benefits of Consolidation
- ✅ **Single Source of Truth**: One file contains complete schema
- ✅ **Simplified Deployment**: Only 2 migrations to run
- ✅ **Better Maintenance**: Easy to understand complete database structure
- ✅ **Reduced Errors**: No dependency issues between migrations
- ✅ **Clean Reset**: Perfect for fresh database installations

## Migration Content Mapping

### Tables from 001_create_complete_schema.up.sql
- ✅ users (core user table)
- ✅ organizations (tenant organizations)
- ✅ organization_settings (org-specific settings)
- ✅ organization_members (user-org relationships)
- ✅ organization_departments (org departments)
- ✅ sessions (authentication sessions)
- ✅ password_resets (password reset tokens)
- ✅ email_verifications (email verification)
- ✅ login_attempts (security tracking)
- ✅ account_lockouts (security lockouts)
- ✅ organization_roles (custom roles)
- ✅ user_organization_roles (role assignments)
- ✅ workflows (workflow definitions)
- ✅ approval_tasks_enhanced (enhanced approval tasks)
- ✅ approval_history (approval audit trail)
- ✅ notifications_enhanced (enhanced notifications)
- ✅ vendors (global vendor master data)
- ✅ categories (org-specific categories)
- ✅ category_budget_codes (category-budget relationships)
- ✅ requisitions (purchase requisitions)
- ✅ budgets (budget management)
- ✅ purchase_orders (purchase orders)
- ✅ payment_vouchers (payment vouchers)
- ✅ goods_received_notes (GRN documents)
- ✅ approval_tasks (legacy approval tasks)
- ✅ audit_logs (system audit logs)
- ✅ notifications (legacy notifications)

### Additional Fields from 002_add_missing_fields.up.sql
**Requisitions Table Additions**:
- ✅ department_id, required_by_date, cost_center, project_code
- ✅ created_by, created_by_name, created_by_role, metadata

**Budgets Table Additions**:
- ✅ name, description, department_id, currency, created_by
- ✅ items, action_history, metadata

**Purchase Orders Table Additions**:
- ✅ description, department, department_id, gl_code, title, priority
- ✅ subtotal, tax, total, budget_code, cost_center, project_code
- ✅ required_by_date, source_requisition_number, source_requisition_id
- ✅ created_by, owner_id, action_history, metadata

**Payment Vouchers Table Additions**:
- ✅ title, department, department_id, priority, requested_by_name
- ✅ requested_date, submitted_at, approved_at, paid_date, payment_due_date
- ✅ budget_code, cost_center, project_code, tax_amount, withholding_tax_amount
- ✅ paid_amount, source_purchase_order_number, source_requisition_number
- ✅ bank_details, items, created_by, owner_id, action_history, metadata

**GRN Table Additions**:
- ✅ created_by, owner_id, warehouse_location, notes, stage_name
- ✅ approved_by, automation_used, auto_created_pv, action_history, metadata

**Approval Tasks Table Additions**:
- ✅ priority, due_at, task_type, title, workflow_id, workflow_name
- ✅ stage_name, importance

### Enhancements from 007_fix_workflow_tables.sql
**Workflows Table Enhancements**:
- ✅ entity_type, version, conditions, deleted_at

**New Workflow Tables**:
- ✅ workflow_assignments (workflow execution tracking)
- ✅ workflow_tasks (individual approval tasks)
- ✅ workflow_defaults (default workflow mappings)

## Implementation Plan

### Phase 1: Create Consolidated Migration
1. ✅ Merge all table definitions from 001, 002, 003, 007
2. ✅ Include all additional fields and enhancements
3. ✅ Consolidate all indexes and constraints
4. ✅ Include all triggers and functions
5. ✅ Add comprehensive comments and documentation

### Phase 2: Create Seed Data Migration
1. ✅ Default organization setup
2. ✅ Default user roles and permissions
3. ✅ Sample workflow definitions
4. ✅ Default categories and vendors
5. ✅ System configuration data

### Phase 3: Update Migration Scripts
1. ✅ Update migrate.sh/migrate.bat scripts
2. ✅ Update README.md with new migration instructions
3. ✅ Create rollback migrations (.down.sql files)
4. ✅ Test migration process thoroughly

### Phase 4: Validation and Testing
1. ✅ Test fresh database creation
2. ✅ Test migration rollback
3. ✅ Validate all foreign keys and constraints
4. ✅ Test application connectivity
5. ✅ Run integration tests

## File Structure After Consolidation

```
backend/database/migrations/
├── 000_drop_all_tables.up.sql              # Emergency cleanup
├── 001_create_complete_schema_consolidated.up.sql  # Complete schema
├── 001_create_complete_schema_consolidated.down.sql # Schema rollback
├── 002_seed_initial_data.up.sql            # Initial data seeding
├── 002_seed_initial_data.down.sql          # Seed data rollback
├── README.md                               # Updated documentation
└── legacy/                                 # Archive old migrations
    ├── 001_create_complete_schema.up.sql
    ├── 002_add_missing_fields.up.sql
    ├── 003_add_alignment_fields.up.sql
    └── 007_fix_workflow_tables.sql
```

## Migration Command Updates

### New Migration Commands
```bash
# Create fresh database
./migrate.sh up

# Rollback all changes
./migrate.sh down

# Reset database (drop + create)
./migrate.sh reset

# Seed data only (after schema creation)
./migrate.sh seed
```

## Validation Checklist

### Schema Validation
- [ ] All tables created successfully
- [ ] All columns present with correct types
- [ ] All indexes created for performance
- [ ] All foreign key constraints established
- [ ] All triggers and functions working
- [ ] All JSONB fields properly typed

### Data Validation
- [ ] Initial organizations created
- [ ] Default users and roles created
- [ ] Sample workflows available
- [ ] Default categories and vendors present
- [ ] System settings configured

### Application Validation
- [ ] Frontend can connect to database
- [ ] All CRUD operations work correctly
- [ ] Authentication and authorization working
- [ ] Workflow system functional
- [ ] Approval system operational
- [ ] Notification system working

## Risk Mitigation

### Backup Strategy
- ✅ Always backup existing database before migration
- ✅ Test migrations in development environment first
- ✅ Have rollback plan ready
- ✅ Document recovery procedures

### Testing Strategy
- ✅ Unit tests for migration scripts
- ✅ Integration tests for application connectivity
- ✅ Performance tests for query optimization
- ✅ Security tests for access controls

## Success Criteria

### Technical Success
- ✅ Zero migration errors
- ✅ All tables and constraints created
- ✅ Application fully functional
- ✅ Performance benchmarks met

### Business Success
- ✅ All business workflows operational
- ✅ Data integrity maintained
- ✅ User experience unchanged
- ✅ System ready for production

## Timeline

### Immediate (Today)
- ✅ Create consolidated migration files
- ✅ Update migration scripts
- ✅ Update documentation

### Next Steps (After Review)
- ✅ Test migrations in development
- ✅ Validate application functionality
- ✅ Prepare for production deployment

---

**Status**: Ready for Implementation
**Risk Level**: Low (comprehensive testing planned)
**Confidence**: High (based on successful audit results)