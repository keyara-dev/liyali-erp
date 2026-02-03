-- ============================================================================
-- SETUP DEMO ORGANIZATION WITH TRIAL DATA
-- ============================================================================

-- This migration sets up the demo organization with proper trial data
-- so that the trial banners and subscription system work correctly

DO $$
DECLARE
    starter_plan_id UUID;
    demo_org_id VARCHAR(255) := 'org-demo-001';
    trial_start TIMESTAMP := CURRENT_TIMESTAMP;
    trial_end TIMESTAMP := CURRENT_TIMESTAMP + INTERVAL '14 days';
BEGIN
    -- Get STARTER_PLAN ID
    SELECT id INTO starter_plan_id FROM subscription_plans WHERE slug = 'STARTER_PLAN' AND is_active = true;
    
    IF starter_plan_id IS NULL THEN
        RAISE NOTICE 'STARTER_PLAN not found, creating subscription plans first...';
        
        -- Insert subscription plans if they don't exist
        INSERT INTO subscription_plans (name, slug, description, price_monthly, price_yearly, features, max_users, sort_order, metadata) VALUES
        (
            'Starter Plan',
            'STARTER_PLAN',
            'Perfect for small teams getting started with procurement workflows',
            0.00,
            0.00,
            '[
                "Core procurement workflows",
                "Up to 50 users",
                "Single Workspace",
                "Document Verification (QR Codes and Doc Numbers)",
                "Standard analytics",
                "Notifications (Email & In-App)"
            ]'::jsonb,
            50,
            1,
            '{
                "offline_capabilities": false,
                "api_access": false,
                "custom_roles": false,
                "priority_support": false,
                "dedicated_instance": false,
                "sla_guarantees": false
            }'::jsonb
        ),
        (
            'Pro Plan',
            'PRO_PLAN',
            'Advanced features for growing organizations',
            99.00,
            990.00,
            '[
                "Everything in Starter Plan",
                "Up to 200 users",
                "Custom Role management",
                "Offline capabilities",
                "Priority support",
                "Advanced analytics",
                "API Access"
            ]'::jsonb,
            200,
            2,
            '{
                "offline_capabilities": true,
                "api_access": true,
                "custom_roles": true,
                "priority_support": true,
                "dedicated_instance": false,
                "sla_guarantees": false
            }'::jsonb
        ),
        (
            'Enterprise',
            'ENTERPRISE',
            'Complete solution for large organizations',
            0.00,
            0.00,
            '[
                "Everything in Pro Plan",
                "Unlimited users",
                "Dedicated instance",
                "Custom integrations",
                "SLA guarantees",
                "Dedicated success manager",
                "Models Creation/Modifications"
            ]'::jsonb,
            -1,
            3,
            '{
                "offline_capabilities": true,
                "api_access": true,
                "custom_roles": true,
                "priority_support": true,
                "dedicated_instance": true,
                "sla_guarantees": true,
                "custom_pricing": true
            }'::jsonb
        )
        ON CONFLICT (slug) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            price_monthly = EXCLUDED.price_monthly,
            price_yearly = EXCLUDED.price_yearly,
            features = EXCLUDED.features,
            max_users = EXCLUDED.max_users,
            sort_order = EXCLUDED.sort_order,
            metadata = EXCLUDED.metadata,
            updated_at = CURRENT_TIMESTAMP;
        
        -- Get the starter plan ID again
        SELECT id INTO starter_plan_id FROM subscription_plans WHERE slug = 'STARTER_PLAN' AND is_active = true;
    END IF;
    
    -- Update demo organization with trial information
    UPDATE organizations SET
        trial_start_date = trial_start,
        trial_end_date = trial_end,
        current_plan_id = starter_plan_id,
        subscription_status = 'trial',
        max_users_allowed = 50,
        tier = 'starter',  -- Set tier to starter for consistency
        updated_at = CURRENT_TIMESTAMP
    WHERE id = demo_org_id;
    
    -- Create organization subscription record
    INSERT INTO organization_subscriptions (
        organization_id,
        plan_id,
        status,
        current_period_start,
        current_period_end
    ) VALUES (
        demo_org_id,
        starter_plan_id,
        'trial',
        trial_start,
        trial_end
    ) ON CONFLICT (organization_id) DO UPDATE SET
        plan_id = EXCLUDED.plan_id,
        status = EXCLUDED.status,
        current_period_start = EXCLUDED.current_period_start,
        current_period_end = EXCLUDED.current_period_end,
        updated_at = CURRENT_TIMESTAMP;
    
    -- Create audit log
    INSERT INTO subscription_audit_logs (
        organization_id,
        action,
        new_plan_id,
        new_status,
        performed_by,
        metadata
    ) VALUES (
        demo_org_id,
        'trial_started',
        starter_plan_id,
        'trial',
        'system',
        jsonb_build_object(
            'trial_duration_days', 14,
            'trial_start', trial_start,
            'trial_end', trial_end,
            'setup_type', 'demo_migration'
        )
    );
    
    RAISE NOTICE 'Demo organization % set up with 14-day trial (ends: %)', demo_org_id, trial_end;
END;
$$;