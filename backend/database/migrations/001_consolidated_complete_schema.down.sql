-- ============================================================================
-- LIYALI GATEWAY - CONSOLIDATED COMPLETE SCHEMA ROLLBACK
-- Migration: 001_consolidated_complete_schema.down
-- Description: Complete rollback of consolidated database schema
-- Date: January 11, 2026
-- ============================================================================

-- Drop all triggers first
DO $ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT n.nspname as schemaname, c.relname as tablename, t.tgname as triggername 
              FROM pg_trigger t 
              JOIN pg_class c ON t.tgrelid = c.oid 
              JOIN pg_namespace n ON c.relnamespace = n.oid 
              WHERE n.nspname = 'public' AND NOT t.tgisinternal) 
    LOOP
        EXECUTE 'DROP TRIGGER IF EXISTS ' || quote_ident(r.triggername) || ' ON ' || quote_ident(r.schemaname) || '.' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $;

-- Drop all functions
DO $ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT p.proname, oidvectortypes(p.proargtypes) as argtypes 
              FROM pg_proc p 
              JOIN pg_namespace n ON p.pronamespace = n.oid 
              WHERE n.nspname = 'public') 
    LOOP
        EXECUTE 'DROP FUNCTION IF EXISTS ' || quote_ident(r.proname) || '(' || r.argtypes || ') CASCADE';
    END LOOP;
END $;

-- Drop all views
DO $ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT c.relname as viewname 
              FROM pg_class c 
              JOIN pg_namespace n ON c.relnamespace = n.oid 
              WHERE n.nspname = 'public' AND c.relkind = 'v') 
    LOOP
        EXECUTE 'DROP VIEW IF EXISTS ' || quote_ident(r.viewname) || ' CASCADE';
    END LOOP;
END $;

-- Drop document number unique indexes first
DROP INDEX IF EXISTS idx_requisitions_document_number_org CASCADE;
DROP INDEX IF EXISTS idx_purchase_orders_document_number_org CASCADE;
DROP INDEX IF EXISTS idx_payment_vouchers_document_number_org CASCADE;
DROP INDEX IF EXISTS idx_goods_received_notes_document_number_org CASCADE;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS stage_approval_records CASCADE;
DROP TABLE IF EXISTS task_assignment_history CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS approval_tasks CASCADE;
DROP TABLE IF EXISTS goods_received_notes CASCADE;
DROP TABLE IF EXISTS payment_vouchers CASCADE;
DROP TABLE IF EXISTS purchase_orders CASCADE;
DROP TABLE IF EXISTS budgets CASCADE;
DROP TABLE IF EXISTS requisitions CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
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

-- Drop all sequences
DO $ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT c.relname as sequencename 
              FROM pg_class c 
              JOIN pg_namespace n ON c.relnamespace = n.oid 
              WHERE n.nspname = 'public' AND c.relkind = 'S') 
    LOOP
        EXECUTE 'DROP SEQUENCE IF EXISTS ' || quote_ident(r.sequencename) || ' CASCADE';
    END LOOP;
END $;

-- Drop all types
DO $ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT t.typname 
              FROM pg_type t 
              JOIN pg_namespace n ON t.typnamespace = n.oid 
              WHERE n.nspname = 'public' AND t.typtype = 'e') 
    LOOP
        EXECUTE 'DROP TYPE IF EXISTS ' || quote_ident(r.typname) || ' CASCADE';
    END LOOP;
END $;

-- Log completion
DO $
BEGIN
    RAISE NOTICE 'Migration 001_consolidated_complete_schema rollback completed successfully';
    RAISE NOTICE 'All tables, indexes, triggers, and functions have been dropped';
    RAISE NOTICE 'Database has been completely cleaned';
    RAISE NOTICE 'Document number consolidation changes have been rolled back';
END $;