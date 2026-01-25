-- Fix user roles: Convert role UUIDs to role names
-- This migration updates the users.role field to contain role names instead of role IDs

-- Update users table to use role names instead of role IDs
UPDATE users 
SET role = (
    SELECT or_roles.name 
    FROM organization_roles or_roles 
    WHERE or_roles.id::text = users.role
    AND or_roles.active = true
)
WHERE users.role ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'  -- Only update UUID-formatted roles
AND EXISTS (
    SELECT 1 
    FROM organization_roles or_roles 
    WHERE or_roles.id::text = users.role
    AND or_roles.active = true
);

-- Set default role for any users that couldn't be matched
UPDATE users 
SET role = 'requester'
WHERE users.role ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'  -- Still UUID format
AND users.role IS NOT NULL;