-- Migration for admin settings and feature flags tables
-- Created: 2026-02-05
-- Fixed: 2026-02-24 - Added DROP TABLE IF EXISTS to handle existing tables

-- Drop existing tables if they exist (to handle schema conflicts)
DROP TABLE IF EXISTS feature_flag_evaluations CASCADE;
DROP TABLE IF EXISTS feature_flags CASCADE;
DROP TABLE IF EXISTS environment_variables CASCADE;
DROP TABLE IF EXISTS system_settings CASCADE;

-- System Settings table
CREATE TABLE system_settings (
    id VARCHAR(255) PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT,
    type VARCHAR(50) NOT NULL DEFAULT 'string',
    category VARCHAR(100) NOT NULL DEFAULT 'general',
    description TEXT,
    default_value TEXT,
    is_required BOOLEAN DEFAULT FALSE,
    is_secret BOOLEAN DEFAULT FALSE,
    environment VARCHAR(50) DEFAULT 'all',
    validation JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255)
);

-- Environment Variables table
CREATE TABLE environment_variables (
    id VARCHAR(255) PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT,
    environment VARCHAR(50) NOT NULL,
    is_secret BOOLEAN DEFAULT FALSE,
    description TEXT,
    is_required BOOLEAN DEFAULT FALSE,
    category VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255)
);

-- Feature Flags table
CREATE TABLE feature_flags (
    id VARCHAR(255) PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL DEFAULT 'boolean',
    default_value TEXT,
    enabled BOOLEAN DEFAULT FALSE,
    environment VARCHAR(50) DEFAULT 'all',
    category VARCHAR(100) DEFAULT 'feature',
    tags JSONB DEFAULT '[]',
    targeting JSONB DEFAULT '{}',
    variations JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    last_evaluated TIMESTAMP WITH TIME ZONE,
    evaluation_count BIGINT DEFAULT 0,
    is_archived BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Feature Flag Evaluations table
CREATE TABLE feature_flag_evaluations (
    id VARCHAR(255) PRIMARY KEY,
    flag_key VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    user_attributes JSONB DEFAULT '{}',
    variation VARCHAR(255),
    value TEXT,
    reason VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX idx_system_settings_category ON system_settings(category);
CREATE INDEX idx_system_settings_environment ON system_settings(environment);
CREATE INDEX idx_system_settings_is_required ON system_settings(is_required);
CREATE INDEX idx_system_settings_is_secret ON system_settings(is_secret);
CREATE INDEX idx_system_settings_updated_at ON system_settings(updated_at);

CREATE INDEX idx_environment_variables_environment ON environment_variables(environment);
CREATE INDEX idx_environment_variables_category ON environment_variables(category);
CREATE INDEX idx_environment_variables_is_secret ON environment_variables(is_secret);

CREATE INDEX idx_feature_flags_category ON feature_flags(category);
CREATE INDEX idx_feature_flags_environment ON feature_flags(environment);
CREATE INDEX idx_feature_flags_enabled ON feature_flags(enabled);
CREATE INDEX idx_feature_flags_is_archived ON feature_flags(is_archived);
CREATE INDEX idx_feature_flags_expires_at ON feature_flags(expires_at);
CREATE INDEX idx_feature_flags_updated_at ON feature_flags(updated_at);

CREATE INDEX idx_feature_flag_evaluations_flag_key ON feature_flag_evaluations(flag_key);
CREATE INDEX idx_feature_flag_evaluations_user_id ON feature_flag_evaluations(user_id);
CREATE INDEX idx_feature_flag_evaluations_timestamp ON feature_flag_evaluations(timestamp);

-- Insert some default system settings
INSERT INTO system_settings (id, key, value, type, category, description, default_value, is_required, is_secret, environment, created_by, updated_by) VALUES
('setting_001', 'app.name', 'Liyali Gateway', 'string', 'general', 'Application name displayed in the UI', 'Liyali Gateway', true, false, 'all', 'system', 'system'),
('setting_002', 'security.session_timeout', '3600', 'number', 'security', 'Session timeout in seconds', '3600', true, false, 'all', 'system', 'system'),
('setting_003', 'performance.cache_enabled', 'true', 'boolean', 'performance', 'Enable application-level caching', 'true', false, false, 'all', 'system', 'system'),
('setting_004', 'notification.email_enabled', 'true', 'boolean', 'notification', 'Enable email notifications', 'true', false, false, 'all', 'system', 'system'),
('setting_005', 'ui.theme', 'light', 'string', 'ui', 'Default UI theme', 'light', false, false, 'all', 'system', 'system');

-- Insert some default feature flags
INSERT INTO feature_flags (id, key, name, description, type, default_value, enabled, environment, category, tags, variations, created_by, updated_by) VALUES
('flag_001', 'new_dashboard', 'New Dashboard UI', 'Enable the redesigned dashboard interface', 'boolean', 'false', false, 'all', 'feature', '["ui", "dashboard", "redesign"]', '[{"id": "enabled", "name": "Enabled", "value": "true", "weight": 50, "is_control": false}, {"id": "disabled", "name": "Disabled", "value": "false", "weight": 50, "is_control": true}]', 'system', 'system'),
('flag_002', 'maintenance_mode', 'Maintenance Mode', 'Enable maintenance mode to block user access', 'boolean', 'false', false, 'all', 'killswitch', '["maintenance", "emergency", "killswitch"]', '[{"id": "active", "name": "Active", "value": "true", "weight": 100, "is_control": false}, {"id": "inactive", "name": "Inactive", "value": "false", "weight": 0, "is_control": true}]', 'system', 'system'),
('flag_003', 'enhanced_security', 'Enhanced Security Features', 'Enable additional security features', 'boolean', 'false', true, 'production', 'feature', '["security", "authentication"]', '[{"id": "enabled", "name": "Enabled", "value": "true", "weight": 100, "is_control": false}, {"id": "disabled", "name": "Disabled", "value": "false", "weight": 0, "is_control": true}]', 'system', 'system');
