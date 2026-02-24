-- Migration: 013_complete_database_integration_pg
-- Description: Complete database integration for 100% data-driven admin console (PostgreSQL version)
-- Date: 2026-02-07
-- Fixed: 2026-02-24 - Converted from SQLite to PostgreSQL syntax

-- ============================================================================
-- SYSTEM MONITORING TABLES
-- ============================================================================

-- System metrics table for real-time monitoring
CREATE TABLE IF NOT EXISTS system_metrics (
    id VARCHAR(255) PRIMARY KEY,
    metric_type VARCHAR(100) NOT NULL, -- 'cpu', 'memory', 'disk', 'network'
    value NUMERIC(10,2) NOT NULL,
    unit VARCHAR(50) NOT NULL, -- 'percent', 'bytes', 'count'
    metadata JSONB, -- JSON for additional data
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- System alerts table
CREATE TABLE IF NOT EXISTS system_alerts (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- 'info', 'warning', 'error', 'critical'
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL, -- 'low', 'medium', 'high', 'critical'
    category VARCHAR(100) NOT NULL, -- 'database', 'security', 'performance', 'maintenance'
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'acknowledged', 'resolved'
    source VARCHAR(255), -- Source of the alert
    metadata JSONB, -- JSON for additional data
    acknowledged_by VARCHAR(255),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(255),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- System logs table
CREATE TABLE IF NOT EXISTS system_logs (
    id VARCHAR(255) PRIMARY KEY,
    level VARCHAR(50) NOT NULL, -- 'debug', 'info', 'warn', 'error', 'fatal'
    service VARCHAR(100) NOT NULL, -- 'api-server', 'database', 'auth-service', etc.
    message TEXT NOT NULL,
    metadata JSONB, -- JSON for additional data
    user_id VARCHAR(255),
    organization_id VARCHAR(255),
    ip_address VARCHAR(50),
    request_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- System services status table
CREATE TABLE IF NOT EXISTS system_services (
    id VARCHAR(255) PRIMARY KEY,
    service_name VARCHAR(100) UNIQUE NOT NULL, -- 'database', 'redis', 'api_server', 'file_system'
    status VARCHAR(50) NOT NULL DEFAULT 'unknown', -- 'healthy', 'degraded', 'unhealthy', 'unknown'
    response_time_ms NUMERIC(10,2),
    last_check_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    metadata JSONB, -- JSON for additional data
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- BILLING & REVENUE TABLES
-- ============================================================================

-- Payments table for revenue tracking
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    subscription_tier VARCHAR(100) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50), -- 'card', 'bank_transfer', 'paypal', etc.
    payment_status VARCHAR(50) NOT NULL, -- 'pending', 'completed', 'failed', 'refunded'
    billing_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    billing_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    invoice_id VARCHAR(255),
    transaction_id VARCHAR(255),
    metadata JSONB, -- JSON for additional data
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    invoice_number VARCHAR(100) UNIQUE NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL, -- 'draft', 'sent', 'paid', 'overdue', 'cancelled'
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    paid_date TIMESTAMP WITH TIME ZONE,
    items JSONB NOT NULL, -- JSON array of line items
    metadata JSONB, -- JSON for additional data
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Subscription events for conversion tracking
CREATE TABLE IF NOT EXISTS subscription_events (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL, -- 'trial_started', 'trial_extended', 'trial_converted', 'trial_expired', 'subscription_upgraded', 'subscription_downgraded', 'subscription_cancelled'
    from_tier VARCHAR(100),
    to_tier VARCHAR(100),
    metadata JSONB, -- JSON for additional data
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- ============================================================================
-- PERFORMANCE MONITORING TABLES
-- ============================================================================

-- API request logs for performance tracking
CREATE TABLE IF NOT EXISTS api_request_logs (
    id VARCHAR(255) PRIMARY KEY,
    method VARCHAR(10) NOT NULL, -- 'GET', 'POST', 'PUT', 'DELETE'
    endpoint VARCHAR(500) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms NUMERIC(10,2) NOT NULL,
    user_id VARCHAR(255),
    organization_id VARCHAR(255),
    ip_address VARCHAR(50),
    user_agent TEXT,
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    error_message TEXT,
    metadata JSONB, -- JSON for additional data
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Feature flag evaluations for tracking (note: this table already exists from migration 011, so we'll skip it)
-- CREATE TABLE IF NOT EXISTS feature_flag_evaluations (...)

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- System metrics indexes
CREATE INDEX IF NOT EXISTS idx_system_metrics_type ON system_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_system_metrics_recorded ON system_metrics(recorded_at DESC);

-- System alerts indexes
CREATE INDEX IF NOT EXISTS idx_system_alerts_status ON system_alerts(status);
CREATE INDEX IF NOT EXISTS idx_system_alerts_severity ON system_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_system_alerts_category ON system_alerts(category);
CREATE INDEX IF NOT EXISTS idx_system_alerts_created ON system_alerts(created_at DESC);

-- System logs indexes
CREATE INDEX IF NOT EXISTS idx_system_logs_level ON system_logs(level);
CREATE INDEX IF NOT EXISTS idx_system_logs_service ON system_logs(service);
CREATE INDEX IF NOT EXISTS idx_system_logs_created ON system_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_logs_user ON system_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_system_logs_org ON system_logs(organization_id);

-- Payments indexes
CREATE INDEX IF NOT EXISTS idx_payments_org ON payments(organization_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(payment_status);
CREATE INDEX IF NOT EXISTS idx_payments_created ON payments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_period ON payments(billing_period_start, billing_period_end);

-- Invoices indexes
CREATE INDEX IF NOT EXISTS idx_invoices_org ON invoices(organization_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_due ON invoices(due_date);

-- Subscription events indexes
CREATE INDEX IF NOT EXISTS idx_subscription_events_org ON subscription_events(organization_id);
CREATE INDEX IF NOT EXISTS idx_subscription_events_type ON subscription_events(event_type);
CREATE INDEX IF NOT EXISTS idx_subscription_events_created ON subscription_events(created_at DESC);

-- API request logs indexes
CREATE INDEX IF NOT EXISTS idx_api_logs_endpoint ON api_request_logs(endpoint);
CREATE INDEX IF NOT EXISTS idx_api_logs_status ON api_request_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_api_logs_created ON api_request_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_logs_org ON api_request_logs(organization_id);

-- ============================================================================
-- SEED INITIAL DATA
-- ============================================================================

-- Insert default system services
INSERT INTO system_services (id, service_name, status, last_check_at) VALUES
('service-db', 'database', 'healthy', CURRENT_TIMESTAMP),
('service-redis', 'redis', 'healthy', CURRENT_TIMESTAMP),
('service-api', 'api_server', 'healthy', CURRENT_TIMESTAMP),
('service-fs', 'file_system', 'healthy', CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Create subscription events for existing trial organizations
INSERT INTO subscription_events (id, organization_id, event_type, from_tier, to_tier, created_at)
SELECT 
    'event-trial-' || o.id,
    o.id,
    'trial_started',
    NULL,
    COALESCE(o.subscription_tier, 'basic'),
    COALESCE(o.trial_start_date, o.created_at)
FROM organizations o
WHERE o.subscription_status = 'trial'
AND NOT EXISTS (
    SELECT 1 FROM subscription_events se 
    WHERE se.organization_id = o.id 
    AND se.event_type = 'trial_started'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Reuse the update_updated_at_column function from migration 012

-- Trigger to update system_alerts updated_at
DROP TRIGGER IF EXISTS update_system_alerts_updated_at ON system_alerts;
CREATE TRIGGER update_system_alerts_updated_at
    BEFORE UPDATE ON system_alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to update system_services updated_at
DROP TRIGGER IF EXISTS update_system_services_updated_at ON system_services;
CREATE TRIGGER update_system_services_updated_at
    BEFORE UPDATE ON system_services
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to update payments updated_at
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to update invoices updated_at
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
CREATE TRIGGER update_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to track subscription changes
CREATE OR REPLACE FUNCTION track_subscription_tier_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.subscription_tier IS DISTINCT FROM NEW.subscription_tier THEN
        INSERT INTO subscription_events (id, organization_id, event_type, from_tier, to_tier, created_at)
        VALUES (
            'event-' || gen_random_uuid()::text,
            NEW.id,
            CASE 
                WHEN NEW.subscription_tier > OLD.subscription_tier THEN 'subscription_upgraded'
                ELSE 'subscription_downgraded'
            END,
            OLD.subscription_tier,
            NEW.subscription_tier,
            CURRENT_TIMESTAMP
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS track_subscription_tier_changes ON organizations;
CREATE TRIGGER track_subscription_tier_changes
    AFTER UPDATE OF subscription_tier ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION track_subscription_tier_changes();

-- Trigger to track trial conversions
CREATE OR REPLACE FUNCTION track_trial_conversions()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.subscription_status = 'trial' AND NEW.subscription_status = 'active' THEN
        INSERT INTO subscription_events (id, organization_id, event_type, from_tier, to_tier, created_at)
        VALUES (
            'event-' || gen_random_uuid()::text,
            NEW.id,
            'trial_converted',
            OLD.subscription_tier,
            NEW.subscription_tier,
            CURRENT_TIMESTAMP
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS track_trial_conversions ON organizations;
CREATE TRIGGER track_trial_conversions
    AFTER UPDATE OF subscription_status ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION track_trial_conversions();
