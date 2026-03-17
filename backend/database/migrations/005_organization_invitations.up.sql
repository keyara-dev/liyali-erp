-- ============================================================================
-- 005_organization_invitations
-- Adds the organization_invitations table used by the invitation system.
-- Admins invite existing platform users to join their org; invitees accept or
-- decline in-app and via email link.  A background worker expires stale rows.
-- ============================================================================

CREATE TABLE IF NOT EXISTS organization_invitations (
    id                  VARCHAR(255)             PRIMARY KEY,
    organization_id     VARCHAR(255)             NOT NULL
                            REFERENCES organizations(id) ON DELETE CASCADE,
    -- NULL when the invitee has no platform account yet (future: email-only invite)
    invited_user_id     VARCHAR(255)
                            REFERENCES users(id) ON DELETE SET NULL,
    invited_email       VARCHAR(255)             NOT NULL,
    invited_by          VARCHAR(255)             NOT NULL
                            REFERENCES users(id),
    role                VARCHAR(255)             NOT NULL DEFAULT 'requester',
    department_id       VARCHAR(255)
                            REFERENCES organization_departments(id) ON DELETE SET NULL,
    branch_id           VARCHAR(255)
                            REFERENCES organization_branches(id) ON DELETE SET NULL,
    status              VARCHAR(50)              NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending','accepted','declined','expired','cancelled')),
    -- Secure token embedded in accept/decline links (never exposed in list endpoints)
    token               VARCHAR(255)             UNIQUE,
    expires_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at         TIMESTAMP WITH TIME ZONE,
    declined_at         TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Fast lookup: list pending invitations for an org or check duplicates before sending
CREATE INDEX IF NOT EXISTS idx_org_invitations_org_status
    ON organization_invitations(organization_id, status);

-- Fast lookup: all pending invitations for a given email address within an org
CREATE INDEX IF NOT EXISTS idx_org_invitations_org_email
    ON organization_invitations(organization_id, invited_email);

-- Fast lookup: in-app notification feed for an invitee
CREATE INDEX IF NOT EXISTS idx_org_invitations_invited_user
    ON organization_invitations(invited_user_id, status);

-- Token lookups for accept/decline flows (already covered by UNIQUE, belt-and-suspenders)
CREATE INDEX IF NOT EXISTS idx_org_invitations_token
    ON organization_invitations(token);

-- updated_at trigger (reuses the function defined in 001_core_schema)
CREATE TRIGGER trg_organization_invitations_updated_at
    BEFORE UPDATE ON organization_invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
