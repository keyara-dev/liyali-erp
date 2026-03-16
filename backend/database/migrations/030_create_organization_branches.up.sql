-- Migration 030: Create organization_branches table
-- Referenced by the OrganizationBranch model and admin_user_handler branch validation
-- but was never created by any prior migration.

CREATE TABLE IF NOT EXISTS organization_branches (
    id              VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(100),
    province_id     VARCHAR(255),
    town_id         VARCHAR(255),
    address         TEXT,
    manager_id      VARCHAR(255),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_organization_branches_org_id ON organization_branches(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_branches_active  ON organization_branches(organization_id, is_active);
