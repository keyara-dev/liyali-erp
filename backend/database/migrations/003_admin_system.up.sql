-- ============================================================================
-- LIYALI GATEWAY — CONSOLIDATED ADMIN SYSTEM
-- Migration: 003_admin_system
-- Replaces: 010, 012
-- ============================================================================

-- ============================================================================
-- SYSTEM SETTINGS
-- ============================================================================
CREATE TABLE IF NOT EXISTS system_settings (
    id            VARCHAR(255) PRIMARY KEY,
    key           VARCHAR(255) UNIQUE NOT NULL,
    value         TEXT,
    type          VARCHAR(50)  NOT NULL DEFAULT 'string',
    category      VARCHAR(100) NOT NULL DEFAULT 'general',
    description   TEXT,
    default_value TEXT,
    is_required   BOOLEAN      DEFAULT FALSE,
    is_secret     BOOLEAN      DEFAULT FALSE,
    environment   VARCHAR(50)  DEFAULT 'all',
    validation    JSONB,
    created_by    VARCHAR(255),
    updated_by    VARCHAR(255),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- ENVIRONMENT VARIABLES
-- ============================================================================
CREATE TABLE IF NOT EXISTS environment_variables (
    id          VARCHAR(255) PRIMARY KEY,
    key         VARCHAR(255) UNIQUE NOT NULL,
    value       TEXT,
    environment VARCHAR(50)  NOT NULL,
    is_secret   BOOLEAN      DEFAULT FALSE,
    description TEXT,
    is_required BOOLEAN      DEFAULT FALSE,
    category    VARCHAR(100),
    created_by  VARCHAR(255),
    updated_by  VARCHAR(255),
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- FEATURE FLAGS  (admin-UI schema from 010 — NOT the subscription-style schema)
-- ============================================================================
CREATE TABLE IF NOT EXISTS feature_flags (
    id               VARCHAR(255) PRIMARY KEY,
    key              VARCHAR(255) UNIQUE NOT NULL,
    name             VARCHAR(255) NOT NULL,
    description      TEXT,
    type             VARCHAR(50)  NOT NULL DEFAULT 'boolean',
    default_value    TEXT,
    enabled          BOOLEAN      DEFAULT FALSE,
    environment      VARCHAR(50)  DEFAULT 'all',
    category         VARCHAR(100) DEFAULT 'feature',
    tags             JSONB        DEFAULT '[]',
    targeting        JSONB        DEFAULT '{}',
    variations       JSONB        DEFAULT '[]',
    last_evaluated   TIMESTAMP WITH TIME ZONE,
    evaluation_count BIGINT       DEFAULT 0,
    is_archived      BOOLEAN      DEFAULT FALSE,
    expires_at       TIMESTAMP WITH TIME ZONE,
    created_by       VARCHAR(255),
    updated_by       VARCHAR(255),
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- FEATURE FLAG EVALUATIONS
-- ============================================================================
CREATE TABLE IF NOT EXISTS feature_flag_evaluations (
    id               VARCHAR(255) PRIMARY KEY,
    flag_key         VARCHAR(255) NOT NULL,
    user_id          VARCHAR(255),
    user_attributes  JSONB        DEFAULT '{}',
    variation        VARCHAR(255),
    value            TEXT,
    reason           VARCHAR(100),
    timestamp        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- SYSTEM MONITORING TABLES  (012)
-- ============================================================================

CREATE TABLE IF NOT EXISTS system_metrics (
    id          VARCHAR(255) PRIMARY KEY,
    metric_type VARCHAR(100) NOT NULL,
    value       NUMERIC(10,2) NOT NULL,
    unit        VARCHAR(50)  NOT NULL,
    metadata    JSONB,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_alerts (
    id               VARCHAR(255) PRIMARY KEY,
    type             VARCHAR(50)  NOT NULL,
    title            VARCHAR(255) NOT NULL,
    message          TEXT         NOT NULL,
    severity         VARCHAR(50)  NOT NULL,
    category         VARCHAR(100) NOT NULL,
    status           VARCHAR(50)  NOT NULL DEFAULT 'active',
    source           VARCHAR(255),
    metadata         JSONB,
    acknowledged_by  VARCHAR(255),
    acknowledged_at  TIMESTAMP WITH TIME ZONE,
    resolved_by      VARCHAR(255),
    resolved_at      TIMESTAMP WITH TIME ZONE,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_logs (
    id              VARCHAR(255) PRIMARY KEY,
    level           VARCHAR(50)  NOT NULL,
    service         VARCHAR(100) NOT NULL,
    message         TEXT         NOT NULL,
    metadata        JSONB,
    user_id         VARCHAR(255),
    organization_id VARCHAR(255),
    ip_address      VARCHAR(50),
    request_id      VARCHAR(255),
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_services (
    id               VARCHAR(255) PRIMARY KEY,
    service_name     VARCHAR(100) UNIQUE NOT NULL,
    status           VARCHAR(50)  NOT NULL DEFAULT 'unknown',
    response_time_ms NUMERIC(10,2),
    last_check_at    TIMESTAMP WITH TIME ZONE,
    error_message    TEXT,
    metadata         JSONB,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_request_logs (
    id                   VARCHAR(255) PRIMARY KEY,
    method               VARCHAR(10)  NOT NULL,
    endpoint             VARCHAR(500) NOT NULL,
    status_code          INTEGER      NOT NULL,
    response_time_ms     NUMERIC(10,2) NOT NULL,
    user_id              VARCHAR(255),
    organization_id      VARCHAR(255),
    ip_address           VARCHAR(50),
    user_agent           TEXT,
    request_size_bytes   INTEGER,
    response_size_bytes  INTEGER,
    error_message        TEXT,
    metadata             JSONB,
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- System settings
CREATE INDEX IF NOT EXISTS idx_system_settings_category   ON system_settings(category);
CREATE INDEX IF NOT EXISTS idx_system_settings_environment ON system_settings(environment);
CREATE INDEX IF NOT EXISTS idx_system_settings_is_required ON system_settings(is_required);
CREATE INDEX IF NOT EXISTS idx_system_settings_is_secret   ON system_settings(is_secret);
CREATE INDEX IF NOT EXISTS idx_system_settings_updated_at  ON system_settings(updated_at);

-- Environment variables
CREATE INDEX IF NOT EXISTS idx_environment_variables_environment ON environment_variables(environment);
CREATE INDEX IF NOT EXISTS idx_environment_variables_category    ON environment_variables(category);
CREATE INDEX IF NOT EXISTS idx_environment_variables_is_secret   ON environment_variables(is_secret);

-- Feature flags
CREATE INDEX IF NOT EXISTS idx_feature_flags_category    ON feature_flags(category);
CREATE INDEX IF NOT EXISTS idx_feature_flags_environment ON feature_flags(environment);
CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled     ON feature_flags(enabled);
CREATE INDEX IF NOT EXISTS idx_feature_flags_is_archived ON feature_flags(is_archived);
CREATE INDEX IF NOT EXISTS idx_feature_flags_expires_at  ON feature_flags(expires_at);
CREATE INDEX IF NOT EXISTS idx_feature_flags_updated_at  ON feature_flags(updated_at);

-- Feature flag evaluations
CREATE INDEX IF NOT EXISTS idx_feature_flag_evaluations_flag_key  ON feature_flag_evaluations(flag_key);
CREATE INDEX IF NOT EXISTS idx_feature_flag_evaluations_user_id   ON feature_flag_evaluations(user_id);
CREATE INDEX IF NOT EXISTS idx_feature_flag_evaluations_timestamp ON feature_flag_evaluations(timestamp);

-- System metrics
CREATE INDEX IF NOT EXISTS idx_system_metrics_type     ON system_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_system_metrics_recorded ON system_metrics(recorded_at DESC);

-- System alerts
CREATE INDEX IF NOT EXISTS idx_system_alerts_status   ON system_alerts(status);
CREATE INDEX IF NOT EXISTS idx_system_alerts_severity ON system_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_system_alerts_category ON system_alerts(category);
CREATE INDEX IF NOT EXISTS idx_system_alerts_created  ON system_alerts(created_at DESC);

-- System logs
CREATE INDEX IF NOT EXISTS idx_system_logs_level   ON system_logs(level);
CREATE INDEX IF NOT EXISTS idx_system_logs_service  ON system_logs(service);
CREATE INDEX IF NOT EXISTS idx_system_logs_created  ON system_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_logs_user     ON system_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_system_logs_org      ON system_logs(organization_id);

-- API request logs
CREATE INDEX IF NOT EXISTS idx_api_logs_endpoint ON api_request_logs(endpoint);
CREATE INDEX IF NOT EXISTS idx_api_logs_status   ON api_request_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_api_logs_created  ON api_request_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_logs_org      ON api_request_logs(organization_id);

-- ============================================================================
-- TRIGGERS
-- ============================================================================
DROP TRIGGER IF EXISTS update_system_alerts_updated_at ON system_alerts;
CREATE TRIGGER update_system_alerts_updated_at
    BEFORE UPDATE ON system_alerts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_system_services_updated_at ON system_services;
CREATE TRIGGER update_system_services_updated_at
    BEFORE UPDATE ON system_services FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- System settings
INSERT INTO system_settings (id, key, value, type, category, description, default_value, is_required, is_secret, environment, created_by, updated_by) VALUES
('setting_001', 'app.name',                  'Liyali Gateway', 'string',  'general',      'Application name displayed in the UI',     'Liyali Gateway', true,  false, 'all', 'system', 'system'),
('setting_002', 'security.session_timeout',  '3600',           'number',  'security',     'Session timeout in seconds',               '3600',           true,  false, 'all', 'system', 'system'),
('setting_003', 'performance.cache_enabled', 'true',           'boolean', 'performance',  'Enable application-level caching',         'true',           false, false, 'all', 'system', 'system'),
('setting_004', 'notification.email_enabled','true',           'boolean', 'notification', 'Enable email notifications',               'true',           false, false, 'all', 'system', 'system'),
('setting_005', 'ui.theme',                  'light',          'string',  'ui',           'Default UI theme',                         'light',          false, false, 'all', 'system', 'system')
ON CONFLICT (id) DO NOTHING;

-- Feature flags
INSERT INTO feature_flags (id, key, name, description, type, default_value, enabled, environment, category, tags, variations, created_by, updated_by) VALUES
(
    'flag_001', 'new_dashboard', 'New Dashboard UI',
    'Enable the redesigned dashboard interface',
    'boolean', 'false', false, 'all', 'feature',
    '["ui","dashboard","redesign"]'::jsonb,
    '[{"id":"enabled","name":"Enabled","value":"true","weight":50,"is_control":false},{"id":"disabled","name":"Disabled","value":"false","weight":50,"is_control":true}]'::jsonb,
    'system', 'system'
),
(
    'flag_002', 'maintenance_mode', 'Maintenance Mode',
    'Enable maintenance mode to block user access',
    'boolean', 'false', false, 'all', 'killswitch',
    '["maintenance","emergency","killswitch"]'::jsonb,
    '[{"id":"active","name":"Active","value":"true","weight":100,"is_control":false},{"id":"inactive","name":"Inactive","value":"false","weight":0,"is_control":true}]'::jsonb,
    'system', 'system'
),
(
    'flag_003', 'enhanced_security', 'Enhanced Security Features',
    'Enable additional security features',
    'boolean', 'false', true, 'production', 'feature',
    '["security","authentication"]'::jsonb,
    '[{"id":"enabled","name":"Enabled","value":"true","weight":100,"is_control":false},{"id":"disabled","name":"Disabled","value":"false","weight":0,"is_control":true}]'::jsonb,
    'system', 'system'
)
ON CONFLICT (id) DO NOTHING;

-- System services
INSERT INTO system_services (id, service_name, status, last_check_at) VALUES
('service-db',    'database',    'healthy', CURRENT_TIMESTAMP),
('service-redis', 'redis',       'healthy', CURRENT_TIMESTAMP),
('service-api',   'api_server',  'healthy', CURRENT_TIMESTAMP),
('service-fs',    'file_system', 'healthy', CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;
