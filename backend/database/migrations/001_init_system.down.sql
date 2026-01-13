-- ============================================================================
-- LIYALI GATEWAY - SYSTEM INITIALIZATION ROLLBACK
-- Migration: 001_init_system (DOWN)
-- Description: Drop all tables created in the init system migration
-- Date: January 13, 2026
-- ============================================================================

-- Drop tables in reverse dependency order

-- Drop business document tables
DROP TABLE IF EXISTS goods_received_notes CASCADE;
DROP TABLE IF EXISTS payment_vouchers CASCADE;
DROP TABLE IF EXISTS purchase_orders CASCADE;
DROP TABLE IF EXISTS budgets CASCADE;
DROP TABLE IF EXISTS requisitions CASCADE;

-- Drop unified documents table
DROP TABLE IF EXISTS documents CASCADE;

-- Drop master data tables
DROP TABLE IF EXISTS category_budget_codes CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS vendors CASCADE;

-- Drop legacy compatibility tables
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS approval_tasks CASCADE;

-- Drop workflow system tables
DROP TABLE IF EXISTS task_assignment_history CASCADE;
DROP TABLE IF EXISTS stage_approval_records CASCADE;
DROP TABLE IF EXISTS workflow_defaults CASCADE;
DROP TABLE IF EXISTS workflow_tasks CASCADE;
DROP TABLE IF EXISTS workflow_assignments CASCADE;
DROP TABLE IF EXISTS workflows CASCADE;

-- Drop authentication tables
DROP TABLE IF EXISTS user_organization_roles CASCADE;
DROP TABLE IF EXISTS organization_roles CASCADE;
DROP TABLE IF EXISTS account_lockouts CASCADE;
DROP TABLE IF EXISTS login_attempts CASCADE;
DROP TABLE IF EXISTS email_verifications CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;

-- Drop organization related tables
DROP TABLE IF EXISTS organization_members CASCADE;
DROP TABLE IF EXISTS organization_departments CASCADE;
DROP TABLE IF EXISTS organization_settings CASCADE;

-- Drop core tables
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS update_documents_updated_at() CASCADE;

-- Log completion
DO $
BEGIN
    RAISE NOTICE 'Migration 001_init_system rollback completed successfully';
    RAISE NOTICE 'All tables and functions have been dropped';
END $;