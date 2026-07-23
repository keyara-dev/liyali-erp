-- ============================================================================
-- SUBSCRIPTION PLANS QUERIES
-- ============================================================================

-- name: GetAllSubscriptionPlans :many
SELECT * FROM subscription_plans 
WHERE is_active = true 
ORDER BY sort_order ASC;

-- name: GetSubscriptionPlanBySlug :one
SELECT * FROM subscription_plans 
WHERE slug = $1 AND is_active = true;

-- name: GetSubscriptionPlanByID :one
SELECT * FROM subscription_plans 
WHERE id = $1;

-- name: CreateSubscriptionPlan :one
INSERT INTO subscription_plans (
    name, slug, description, price_monthly, price_yearly, 
    features, max_users, is_active, sort_order, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateSubscriptionPlan :one
UPDATE subscription_plans SET
    name = $2,
    description = $3,
    price_monthly = $4,
    price_yearly = $5,
    features = $6,
    max_users = $7,
    is_active = $8,
    sort_order = $9,
    metadata = $10,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteSubscriptionPlan :exec
UPDATE subscription_plans SET 
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- ============================================================================
-- ORGANIZATION SUBSCRIPTION QUERIES
-- ============================================================================

-- name: GetOrganizationSubscriptionDetails :one
SELECT * FROM organization_subscription_details 
WHERE organization_id = $1;

-- name: GetOrganizationSubscription :one
SELECT * FROM organization_subscriptions 
WHERE organization_id = $1;

-- name: CreateOrganizationSubscription :one
INSERT INTO organization_subscriptions (
    organization_id, plan_id, stripe_subscription_id, status,
    current_period_start, current_period_end, cancel_at_period_end
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateOrganizationSubscription :one
UPDATE organization_subscriptions SET
    plan_id = $2,
    stripe_subscription_id = $3,
    status = $4,
    current_period_start = $5,
    current_period_end = $6,
    cancel_at_period_end = $7,
    payment_failed_count = $8,
    last_payment_failed_at = $9,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
RETURNING *;

-- name: UpdateSubscriptionStatus :exec
UPDATE organization_subscriptions SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1;

-- name: UpdateOrganizationSubscriptionStatus :exec
UPDATE organizations SET
    subscription_status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- ============================================================================
-- TRIAL MANAGEMENT QUERIES
-- ============================================================================

-- name: GetOrganizationTrialStatus :one
SELECT 
    o.id as organization_id,
    o.subscription_status,
    o.trial_start_date,
    o.trial_end_date,
    o.grace_period_ends_at,
    sp.slug as plan_slug,
    sp.name as plan_name,
    CASE 
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP <= o.trial_end_date THEN
            EXTRACT(DAYS FROM o.trial_end_date - CURRENT_TIMESTAMP)::INTEGER
        ELSE 0
    END as days_remaining,
    CASE 
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP > o.trial_end_date THEN true
        ELSE false
    END as is_expired,
    CASE 
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP <= o.trial_end_date THEN true
        ELSE false
    END as is_active,
    CASE 
        WHEN o.grace_period_ends_at IS NOT NULL AND CURRENT_TIMESTAMP <= o.grace_period_ends_at THEN true
        ELSE false
    END as in_grace_period
FROM organizations o
LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
WHERE o.id = $1;

-- name: StartOrganizationTrial :exec
SELECT start_organization_trial($1::VARCHAR);

-- name: ExtendOrganizationTrial :exec
SELECT extend_organization_trial($1::VARCHAR, $2::INTEGER, $3::VARCHAR);

-- name: GetTrialsEndingSoon :many
SELECT 
    o.id,
    o.name,
    o.trial_end_date,
    EXTRACT(DAYS FROM o.trial_end_date - CURRENT_TIMESTAMP)::INTEGER as days_remaining
FROM organizations o
WHERE o.subscription_status = 'trial'
  AND o.trial_end_date BETWEEN CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP + INTERVAL '$1 days'
ORDER BY o.trial_end_date ASC;

-- ============================================================================
-- FEATURE FLAGS QUERIES
-- ============================================================================

-- name: GetAllFeatureFlags :many
SELECT * FROM feature_flags 
WHERE is_active = true 
ORDER BY name ASC;

-- name: GetFeatureFlagByName :one
SELECT * FROM feature_flags 
WHERE name = $1 AND is_active = true;

-- name: CheckOrganizationFeatureAccess :one
SELECT organization_has_feature($1, $2) as has_access;

-- name: GetFeatureFlagsForPlan :many
SELECT ff.* FROM subscription_feature_requirements ff
WHERE (
    ff.plan_requirements ? $1
    OR ($2 = true AND ff.is_trial_allowed = true)
  )
ORDER BY ff.name ASC;

-- name: CreateFeatureFlag :one
INSERT INTO subscription_feature_requirements (
    name, description, plan_requirements, is_trial_allowed, is_enterprise_only
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateFeatureFlag :one
UPDATE subscription_feature_requirements SET
    description = $2,
    plan_requirements = $3,
    is_trial_allowed = $4,
    is_enterprise_only = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE name = $1
RETURNING *;

-- ============================================================================
-- PLAN LIMITS AND ANALYTICS QUERIES
-- ============================================================================

-- name: GetOrganizationPlanLimits :one
SELECT 
    o.id as organization_id,
    o.max_users_allowed,
    sp.max_users as plan_max_users,
    sp.metadata as plan_metadata,
    COUNT(u.id) as current_user_count,
    CASE 
        WHEN sp.max_users = -1 THEN true  -- Unlimited
        WHEN COUNT(u.id) < o.max_users_allowed THEN true
        ELSE false
    END as can_add_users
FROM organizations o
LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
LEFT JOIN users u ON u.organization_id = o.id AND u.deleted_at IS NULL
WHERE o.id = $1
GROUP BY o.id, o.max_users_allowed, sp.max_users, sp.metadata;

-- name: CheckOrganizationUserLimit :one
SELECT 
    CASE 
        WHEN sp.max_users = -1 THEN true  -- Unlimited
        WHEN COUNT(u.id) < o.max_users_allowed THEN true
        ELSE false
    END as within_limit
FROM organizations o
LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
LEFT JOIN users u ON u.organization_id = o.id AND u.deleted_at IS NULL
WHERE o.id = $1
GROUP BY o.id, o.max_users_allowed, sp.max_users;

-- name: GetSubscriptionAnalytics :one
SELECT 
    COUNT(*) FILTER (WHERE subscription_status = 'trial') as trial_count,
    COUNT(*) FILTER (WHERE subscription_status = 'active') as active_count,
    COUNT(*) FILTER (WHERE subscription_status = 'past_due') as past_due_count,
    COUNT(*) FILTER (WHERE subscription_status = 'canceled') as canceled_count,
    COUNT(*) FILTER (WHERE subscription_status = 'expired') as expired_count,
    COUNT(*) FILTER (WHERE subscription_status = 'trial' AND trial_end_date < CURRENT_TIMESTAMP) as expired_trials,
    COUNT(*) FILTER (WHERE subscription_status = 'trial' AND trial_end_date BETWEEN CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP + INTERVAL '3 days') as trials_ending_soon,
    ROUND(
        COUNT(*) FILTER (WHERE subscription_status = 'active')::DECIMAL / 
        NULLIF(COUNT(*) FILTER (WHERE subscription_status IN ('trial', 'active', 'canceled', 'expired')), 0) * 100, 
        2
    ) as conversion_rate
FROM organizations;

-- name: GetPlanDistribution :many
SELECT 
    sp.name as plan_name,
    sp.slug as plan_slug,
    COUNT(o.id) as organization_count,
    ROUND(COUNT(o.id)::DECIMAL / (SELECT COUNT(*) FROM organizations) * 100, 2) as percentage
FROM subscription_plans sp
LEFT JOIN organizations o ON o.current_plan_id = sp.id
WHERE sp.is_active = true
GROUP BY sp.id, sp.name, sp.slug, sp.sort_order
ORDER BY sp.sort_order ASC;

-- ============================================================================
-- AUDIT LOG QUERIES
-- ============================================================================

-- name: CreateSubscriptionAuditLog :one
INSERT INTO subscription_audit_logs (
    organization_id, action, old_plan_id, new_plan_id, 
    old_status, new_status, metadata, performed_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetOrganizationAuditLogs :many
SELECT 
    sal.*,
    old_plan.name as old_plan_name,
    old_plan.slug as old_plan_slug,
    new_plan.name as new_plan_name,
    new_plan.slug as new_plan_slug
FROM subscription_audit_logs sal
LEFT JOIN subscription_plans old_plan ON sal.old_plan_id = old_plan.id
LEFT JOIN subscription_plans new_plan ON sal.new_plan_id = new_plan.id
WHERE sal.organization_id = $1
ORDER BY sal.performed_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentAuditLogs :many
SELECT 
    sal.*,
    o.name as organization_name,
    old_plan.name as old_plan_name,
    old_plan.slug as old_plan_slug,
    new_plan.name as new_plan_name,
    new_plan.slug as new_plan_slug
FROM subscription_audit_logs sal
LEFT JOIN organizations o ON sal.organization_id = o.id
LEFT JOIN subscription_plans old_plan ON sal.old_plan_id = old_plan.id
LEFT JOIN subscription_plans new_plan ON sal.new_plan_id = new_plan.id
ORDER BY sal.performed_at DESC
LIMIT $1 OFFSET $2;

-- ============================================================================
-- ADMIN QUERIES
-- ============================================================================

-- name: GetAllOrganizationsWithSubscriptionStatus :many
SELECT 
    o.id,
    o.name,
    o.subscription_status,
    o.trial_start_date,
    o.trial_end_date,
    o.grace_period_ends_at,
    sp.name as plan_name,
    sp.slug as plan_slug,
    os.stripe_subscription_id,
    os.payment_failed_count,
    os.last_payment_failed_at,
    CASE 
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP <= o.trial_end_date THEN
            EXTRACT(DAYS FROM o.trial_end_date - CURRENT_TIMESTAMP)::INTEGER
        ELSE 0
    END as trial_days_remaining
FROM organizations o
LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
LEFT JOIN organization_subscriptions os ON o.id = os.organization_id
ORDER BY o.created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateOrganizationPlan :exec
UPDATE organizations SET
    current_plan_id = $2,
    subscription_status = $3,
    max_users_allowed = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: SetOrganizationGracePeriod :exec
UPDATE organizations SET
    grace_period_ends_at = CURRENT_TIMESTAMP + INTERVAL '$2 days',
    subscription_status = 'past_due',
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;