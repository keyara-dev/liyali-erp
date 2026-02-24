-- Migration: 013_complete_database_integration
-- Description: Complete database integration for 100% data-driven admin console
-- Date: 2026-02-07

-- ============================================================================
-- SYSTEM MONITORING TABLES
-- ============================================================================

-- System metrics table for real-time monitoring
CREATE TABLE IF NOT EXISTS system_metrics (
    id TEXT PRIMARY KEY,
    metric_type TEXT NOT NULL, -- 'cpu', 'memory', 'disk', 'network'
    value REAL NOT NULL,
    unit TEXT NOT NULL, -- 'percent', 'bytes', 'count'
    metadata TEXT, -- JSON for additional data
    recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- System alerts table
CREATE TABLE IF NOT EXISTS system_alerts (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- 'info', 'warning', 'error', 'critical'
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    severity TEXT NOT NULL, -- 'low', 'medium', 'high', 'critical'
    category TEXT NOT NULL, -- 'database', 'security', 'performance', 'maintenance'
    status TEXT NOT NULL DEFAULT 'active', -- 'active', 'acknowledged', 'resolved'
    source TEXT, -- Source of the alert
    metadata TEXT, -- JSON for additional data
    acknowledged_by TEXT,
    acknowledged_at DATETIME,
    resolved_by TEXT,
    resolved_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- System logs table
CREATE TABLE IF NOT EXISTS system_logs (
    id TEXT PRIMARY KEY,
    level TEXT NOT NULL, -- 'debug', 'info', 'warn', 'error', 'fatal'
    service TEXT NOT NULL, -- 'api-server', 'database', 'auth-service', etc.
    message TEXT NOT NULL,
    metadata TEXT, -- JSON for additional data
    user_id TEXT,
    organization_id TEXT,
    ip_address TEXT,
    request_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- System services status table
CREATE TABLE IF NOT EXISTS system_services (
    id TEXT PRIMARY KEY,
    service_name TEXT UNIQUE NOT NULL, -- 'database', 'redis', 'api_server', 'file_system'
    status TEXT NOT NULL DEFAULT 'unknown', -- 'healthy', 'degraded', 'unhealthy', 'unknown'
    response_time_ms REAL,
    last_check_at DATETIME,
    error_message TEXT,
    metadata TEXT, -- JSON for additional data
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- BILLING & REVENUE TABLES
-- ============================================================================

-- Payments table for revenue tracking
CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    subscription_tier TEXT NOT NULL,
    amount REAL NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    payment_method TEXT, -- 'card', 'bank_transfer', 'paypal', etc.
    payment_status TEXT NOT NULL, -- 'pending', 'completed', 'failed', 'refunded'
    billing_period_start DATETIME NOT NULL,
    billing_period_end DATETIME NOT NULL,
    invoice_id TEXT,
    transaction_id TEXT,
    metadata TEXT, -- JSON for additional data
    paid_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    invoice_number TEXT UNIQUE NOT NULL,
    amount REAL NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    status TEXT NOT NULL, -- 'draft', 'sent', 'paid', 'overdue', 'cancelled'
    due_date DATETIME NOT NULL,
    paid_date DATETIME,
    items TEXT NOT NULL, -- JSON array of line items
    metadata TEXT, -- JSON for additional data
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Subscription events for conversion tracking
CREATE TABLE IF NOT EXISTS subscription_events (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    event_type TEXT NOT NULL, -- 'trial_started', 'trial_extended', 'trial_converted', 'trial_expired', 'subscription_upgraded', 'subscription_downgraded', 'subscription_cancelled'
    from_tier TEXT,
    to_tier TEXT,
    metadata TEXT, -- JSON for additional data
    created_by TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- ============================================================================
-- PERFORMANCE MONITORING TABLES
-- ============================================================================

-- API request logs for performance tracking
CREATE TABLE IF NOT EXISTS api_request_logs (
    id TEXT PRIMARY KEY,
    method TEXT NOT NULL, -- 'GET', 'POST', 'PUT', 'DELETE'
    endpoint TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms REAL NOT NULL,
    user_id TEXT,
    organization_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    error_message TEXT,
    metadata TEXT, -- JSON for additional data
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Feature flag evaluations for tracking
CREATE TABLE IF NOT EXISTS feature_flag_evaluations (
    id TEXT PRIMARY KEY,
    flag_key TEXT NOT NULL,
    user_id TEXT,
    organization_id TEXT,
    result BOOLEAN NOT NULL,
    evaluation_time_ms REAL,
    context TEXT, -- JSON for evaluation context
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

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

-- Feature flag evaluations indexes
CREATE INDEX IF NOT EXISTS idx_flag_evals_key ON feature_flag_evaluations(flag_key);
CREATE INDEX IF NOT EXISTS idx_flag_evals_created ON feature_flag_evaluations(created_at DESC);

-- ============================================================================
-- SEED INITIAL DATA
-- ============================================================================

-- Insert default system services
INSERT OR IGNORE INTO system_services (id, service_name, status, last_check_at) VALUES
('service-db', 'database', 'healthy', datetime('now')),
('service-redis', 'redis', 'healthy', datetime('now')),
('service-api', 'api_server', 'healthy', datetime('now')),
('service-fs', 'file_system', 'healthy', datetime('now'));

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
);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Trigger to update system_alerts updated_at
CREATE TRIGGER IF NOT EXISTS update_system_alerts_updated_at
    AFTER UPDATE ON system_alerts
    FOR EACH ROW
BEGIN
    UPDATE system_alerts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to update system_services updated_at
CREATE TRIGGER IF NOT EXISTS update_system_services_updated_at
    AFTER UPDATE ON system_services
    FOR EACH ROW
BEGIN
    UPDATE system_services SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to update payments updated_at
CREATE TRIGGER IF NOT EXISTS update_payments_updated_at
    AFTER UPDATE ON payments
    FOR EACH ROW
BEGIN
    UPDATE payments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to update invoices updated_at
CREATE TRIGGER IF NOT EXISTS update_invoices_updated_at
    AFTER UPDATE ON invoices
    FOR EACH ROW
BEGIN
    UPDATE invoices SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to track subscription changes
CREATE TRIGGER IF NOT EXISTS track_subscription_tier_changes
    AFTER UPDATE OF subscription_tier ON organizations
    FOR EACH ROW
    WHEN OLD.subscription_tier != NEW.subscription_tier
BEGIN
    INSERT INTO subscription_events (id, organization_id, event_type, from_tier, to_tier, created_at)
    VALUES (
        'event-' || hex(randomblob(16)),
        NEW.id,
        CASE 
            WHEN NEW.subscription_tier > OLD.subscription_tier THEN 'subscription_upgraded'
            ELSE 'subscription_downgraded'
        END,
        OLD.subscription_tier,
        NEW.subscription_tier,
        CURRENT_TIMESTAMP
    );
END;

-- Trigger to track trial conversions
CREATE TRIGGER IF NOT EXISTS track_trial_conversions
    AFTER UPDATE OF subscription_status ON organizations
    FOR EACH ROW
    WHEN OLD.subscription_status = 'trial' AND NEW.subscription_status = 'active'
BEGIN
    INSERT INTO subscription_events (id, organization_id, event_type, from_tier, to_tier, created_at)
    VALUES (
        'event-' || hex(randomblob(16)),
        NEW.id,
        'trial_converted',
        OLD.subscription_tier,
        NEW.subscription_tier,
        CURRENT_TIMESTAMP
    );
END;