-- Migration 018 rollback: Remove tagline column from organizations
ALTER TABLE organizations DROP COLUMN IF EXISTS tagline;
