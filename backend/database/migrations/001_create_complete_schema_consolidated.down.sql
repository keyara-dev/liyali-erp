-- ============================================================================
-- LIYALI GATEWAY - ROLLBACK COMPLETE DATABASE SCHEMA
-- Migration: 001_create_complete_schema_consolidated.down.sql
-- Description: Drops all tables and functions created by the consolidated schema
-- Version: Rollback for consolidated migration
-- Date: 2025-01-07
-- ============================================================================

-- Drop all triggers first
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;
DROP TRIGGER IF EXISTS update_approval_tasks_updated_at ON approval_tasks;
DROP TRIGGER IF EXISTS update_vendors_updated_at ON vendors;
DROP TRIGGER IF EXISTS update_category_budget_codes_updated_at ON category_budget_codes;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_goods_received_notes_updated_at ON goods_received_notes;
DROP TRIGGER IF EXISTS update_payment_vouchers_updated_at ON payment_vouchers;
DROP TRIGGER IF EXISTS update_purchase_orders_updated_at ON purchase_orders;
DROP TRIGGER IF EXISTS update_budgets_updated_at ON budgets;
DROP TRIGGER IF EXISTS update_requisitions_updated_at ON requisitions;
DROP TRIGGER IF EXISTS update_approval_tasks_enh_updated_at ON approval_tasks_enhanced;
DROP TRIGGER IF EXISTS update_workflow_assignments_updated_at ON workflow_assignments;
DROP TRIGGER IF EXISTS update_workflows_updated_at ON workflows;
DROP TRIGGER IF EXISTS update_organization_roles_updated_at ON organization_roles;
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;
DROP TRIGGER IF EXISTS update_organization_departments_updated_at ON organization_departments;
DROP TRIGGER IF EXISTS update_organization_members_updated_at ON organization_members;
DROP TRIGGER IF EXISTS update_organization_settings_updated_at ON organization_settings;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop all tables in reverse dependency order
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS approval_tasks CASCADE;
DROP TABLE IF EXISTS goods_received_notes CASCADE;
DROP TABLE IF EXISTS payment_vouchers CASCADE;
DROP TABLE IF EXISTS purchase_orders CASCADE;
DROP TABLE IF EXISTS budgets CASCADE;
DROP TABLE IF EXISTS requisitions CASCADE;
DROP TABLE IF EXISTS category_budget_codes CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS vendors CASCADE;
DROP TABLE IF EXISTS notifications_enhanced CASCADE;
DROP TABLE IF EXISTS approval_history CASCADE;
DROP TABLE IF EXISTS approval_tasks_enhanced CASCADE;
DROP TABLE IF EXISTS workflow_defaults CASCADE;
DROP TABLE IF EXISTS workflow_tasks CASCADE;
DROP TABLE IF EXISTS workflow_assignments CASCADE;
DROP TABLE IF EXISTS workflows CASCADE;
DROP TABLE IF EXISTS user_organization_roles CASCADE;
DROP TABLE IF EXISTS organization_roles CASCADE;
DROP TABLE IF EXISTS account_lockouts CASCADE;
DROP TABLE IF EXISTS login_attempts CASCADE;
DROP TABLE IF EXISTS email_verifications CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS organization_departments CASCADE;
DROP TABLE IF EXISTS organization_members CASCADE;
DROP TABLE IF EXISTS organization_settings CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 001_create_complete_schema_consolidated rollback completed successfully';
    RAISE NOTICE 'All tables, triggers, and functions have been dropped';
    RAISE NOTICE 'Database is now in clean state';
END $$;