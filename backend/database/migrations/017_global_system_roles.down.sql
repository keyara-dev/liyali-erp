-- Rollback migration 017: best-effort reversal of global system roles
-- Note: This is a best-effort rollback. Per-org role copies and user assignments
-- cannot be perfectly restored.

-- 1. Drop the partial unique index
DROP INDEX IF EXISTS uk_global_system_role_name;

-- 2. Remove super_admin role
DELETE FROM organization_roles WHERE name = 'super_admin' AND is_system_role = true AND organization_id IS NULL;

-- 3. Re-activate deactivated duplicate system roles
UPDATE organization_roles SET active = true WHERE is_system_role = true AND active = false;

-- 4. Set global system roles back to the first available organization
UPDATE organization_roles
SET organization_id = (SELECT id FROM organizations ORDER BY created_at ASC LIMIT 1)
WHERE organization_id IS NULL AND is_system_role = true;

-- 5. Restore NOT NULL constraint
ALTER TABLE organization_roles ALTER COLUMN organization_id SET NOT NULL;
