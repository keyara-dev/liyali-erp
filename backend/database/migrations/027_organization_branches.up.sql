-- Migration: 027_organization_branches
-- Description: Create organization_branches table and add branch_id to organization_members

CREATE TABLE IF NOT EXISTS organization_branches (
    id              VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(50),
    description     TEXT,
    location        TEXT,
    manager_name    VARCHAR(255),
    active          BOOLEAN DEFAULT true,
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_org_branches_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_org_branches_organization ON organization_branches(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_branches_active ON organization_branches(organization_id, is_active);

ALTER TABLE organization_members
    ADD COLUMN IF NOT EXISTS branch_id VARCHAR(255);

ALTER TABLE organization_members
    DROP CONSTRAINT IF EXISTS fk_org_members_branch;

ALTER TABLE organization_members
    ADD CONSTRAINT fk_org_members_branch
        FOREIGN KEY (branch_id) REFERENCES organization_branches(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_org_members_branch ON organization_members(branch_id);
