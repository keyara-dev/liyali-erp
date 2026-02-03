-- ============================================================================
-- LIYALI GATEWAY - MINIMAL PERFORMANCE OPTIMIZATION
-- Migration: 010_performance_optimization_minimal
-- Description: Essential indexes only to avoid deployment timeout
-- Date: February 3, 2026
-- ============================================================================

-- Only the most critical indexes that should create quickly
-- Full optimization will be done in a separate maintenance window

-- Most critical: Organization members JOIN optimization
CREATE INDEX IF NOT EXISTS idx_org_members_user_active ON organization_members(user_id, active);

-- Most critical: Requisitions status for analytics
CREATE INDEX IF NOT EXISTS idx_requisitions_org_status ON requisitions(organization_id, status);

-- Session cleanup (small table, fast creation)
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Performance note: Additional indexes will be created in maintenance window
-- This minimal set addresses the most critical slow queries without timeout risk