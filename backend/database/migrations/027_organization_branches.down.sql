-- Migration: 027_organization_branches (rollback)

DROP INDEX IF EXISTS idx_org_members_branch;

ALTER TABLE organization_members
    DROP CONSTRAINT IF EXISTS fk_org_members_branch;

ALTER TABLE organization_members
    DROP COLUMN IF EXISTS branch_id;

DROP INDEX IF EXISTS idx_org_branches_active;
DROP INDEX IF EXISTS idx_org_branches_organization;

DROP TABLE IF EXISTS organization_branches;
