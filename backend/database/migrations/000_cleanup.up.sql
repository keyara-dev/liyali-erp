-- ============================================================================
-- CLEANUP MIGRATION: Drop all tables for full database reset
-- This runs first in --reset mode to wipe the schema clean.
-- ============================================================================

DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO public;
