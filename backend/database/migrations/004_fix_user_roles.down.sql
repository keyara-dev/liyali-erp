-- Rollback migration: Convert role names back to role UUIDs
-- This is a best-effort rollback - some data may be lost if roles were deleted

-- This rollback is complex because we're going from names back to UUIDs
-- We'll try to match role names to UUIDs, but this may not be perfect
UPDATE users 
SET role = (
    SELECT or_roles.id::text 
    FROM organization_roles or_roles 
    WHERE or_roles.name = users.role
    AND or_roles.active = true
    LIMIT 1  -- In case of duplicates, take the first one
)
WHERE users.role NOT LIKE '%-%'  -- Only update non-UUID formatted roles
AND EXISTS (
    SELECT 1 
    FROM organization_roles or_roles 
    WHERE or_roles.name = users.role
    AND or_roles.active = true
);

-- Note: This rollback may not be perfect if:
-- 1. Role names were changed after the migration
-- 2. Roles were deleted after the migration
-- 3. There are duplicate role names across organizations