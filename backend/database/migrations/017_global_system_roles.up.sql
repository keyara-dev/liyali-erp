-- Migration 017: Convert per-org system roles to global system roles + add super_admin
-- System roles (admin, approver, requester, finance, viewer) are currently duplicated per-organization.
-- This migration consolidates them into single global rows with organization_id = NULL.

-- 1. Make organization_id nullable for global system roles
ALTER TABLE organization_roles ALTER COLUMN organization_id DROP NOT NULL;

-- 2. Deduplicate system roles: keep oldest per name, re-point user assignments, deactivate dupes
DO $$
DECLARE
    role_name_var VARCHAR(100);
    canonical_id UUID;
BEGIN
    FOR role_name_var IN
        SELECT DISTINCT name FROM organization_roles
        WHERE is_system_role = true AND active = true
    LOOP
        -- Pick the oldest row as the canonical global role
        SELECT id INTO canonical_id FROM organization_roles
        WHERE name = role_name_var AND is_system_role = true AND active = true
        ORDER BY created_at ASC LIMIT 1;

        -- Re-point user role assignments from duplicate roles to the canonical one
        -- First delete any assignments that would create duplicates (same user+org+canonical_id)
        DELETE FROM user_organization_roles
        WHERE id IN (
            SELECT uor.id
            FROM user_organization_roles uor
            WHERE uor.role_id IN (
                SELECT id FROM organization_roles
                WHERE name = role_name_var AND is_system_role = true AND active = true AND id != canonical_id
            )
            AND EXISTS (
                SELECT 1 FROM user_organization_roles uor2
                WHERE uor2.user_id = uor.user_id
                AND uor2.organization_id = uor.organization_id
                AND uor2.role_id = canonical_id
            )
        );

        -- Now safely re-point remaining assignments
        UPDATE user_organization_roles SET role_id = canonical_id
        WHERE role_id IN (
            SELECT id FROM organization_roles
            WHERE name = role_name_var AND is_system_role = true AND active = true AND id != canonical_id
        );

        -- Make the canonical role global (organization_id = NULL)
        UPDATE organization_roles SET organization_id = NULL WHERE id = canonical_id;

        -- Deactivate the duplicate rows
        UPDATE organization_roles SET active = false, updated_at = NOW()
        WHERE name = role_name_var AND is_system_role = true AND id != canonical_id;
    END LOOP;
END $$;

-- 3. Add super_admin global role with ALL org-level permissions
INSERT INTO organization_roles (name, description, is_system_role, permissions, active)
SELECT 'super_admin', 'Full platform access with all permissions', true,
    '["requisition:view","requisition:create","requisition:edit","requisition:delete","requisition:approve","requisition:reject","budget:view","budget:create","budget:edit","budget:delete","budget:approve","budget:reject","purchase_order:view","purchase_order:create","purchase_order:edit","purchase_order:delete","purchase_order:approve","purchase_order:reject","payment_voucher:view","payment_voucher:create","payment_voucher:edit","payment_voucher:delete","payment_voucher:approve","payment_voucher:reject","grn:view","grn:create","grn:edit","grn:delete","vendor:view","vendor:create","vendor:edit","vendor:delete","category:view","category:create","category:edit","category:delete","organization:view","organization:edit","organization:manage_users","organization:manage_workflows","analytics:view","audit_log:view"]'::jsonb,
    true
WHERE NOT EXISTS (
    SELECT 1 FROM organization_roles WHERE name = 'super_admin' AND is_system_role = true AND organization_id IS NULL
);

-- 4. Partial unique index: prevent duplicate global system roles
CREATE UNIQUE INDEX IF NOT EXISTS uk_global_system_role_name
ON organization_roles (name) WHERE organization_id IS NULL AND is_system_role = true;
