-- ============================================================================
-- ROLLBACK: 003_admin_system
-- ============================================================================

DROP TABLE IF EXISTS api_request_logs CASCADE;
DROP TABLE IF EXISTS system_services CASCADE;
DROP TABLE IF EXISTS system_logs CASCADE;
DROP TABLE IF EXISTS system_alerts CASCADE;
DROP TABLE IF EXISTS system_metrics CASCADE;
DROP TABLE IF EXISTS feature_flag_evaluations CASCADE;
DROP TABLE IF EXISTS feature_flags CASCADE;
DROP TABLE IF EXISTS environment_variables CASCADE;
DROP TABLE IF EXISTS system_settings CASCADE;
