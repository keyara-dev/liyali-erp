-- Migration 029: Seed missing feature flags into the feature_flags table
-- The organization_has_feature() function queries feature_flags (not subscription_features).
-- Several Pro-tier features referenced by FeatureGate in the frontend were missing.

INSERT INTO feature_flags (name, description, plan_requirements, is_trial_allowed, is_enterprise_only) VALUES
('audit_logs_90_days',  'Audit trail with 90-day retention',        '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('advanced_workflows',  'Complex multi-stage conditional workflows', '["PRO_PLAN", "ENTERPRISE"]'::jsonb, true,  false),
('email_notifications', 'Automated email notifications',            '["PRO_PLAN", "ENTERPRISE"]'::jsonb, true,  false),
('data_export',         'Export data to CSV/Excel',                 '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('multi_currency',      'Support for multiple currencies',          '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('advanced_reporting',  'Custom report builder',                    '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('workflow_templates',  'Pre-built workflow templates',             '["PRO_PLAN", "ENTERPRISE"]'::jsonb, true,  false),
('webhooks',            'Real-time event webhooks',                 '["ENTERPRISE"]'::jsonb,             false, false),
('custom_fields',       'Add custom fields to documents',           '["ENTERPRISE"]'::jsonb,             false, false),
('bulk_operations',     'Batch process documents',                  '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('audit_logs_unlimited','Unlimited audit log retention',            '["ENTERPRISE"]'::jsonb,             false, true)
ON CONFLICT (name) DO UPDATE SET
    description       = EXCLUDED.description,
    plan_requirements = EXCLUDED.plan_requirements,
    is_trial_allowed  = EXCLUDED.is_trial_allowed,
    is_enterprise_only = EXCLUDED.is_enterprise_only,
    updated_at        = CURRENT_TIMESTAMP;
