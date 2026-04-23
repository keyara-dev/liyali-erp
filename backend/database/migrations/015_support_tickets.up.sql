-- ============================================================================
-- SUPPORT TICKETS
-- ============================================================================

CREATE TABLE IF NOT EXISTS support_tickets (
    id                     VARCHAR(255) PRIMARY KEY,
    ticket_number          VARCHAR(64)  UNIQUE NOT NULL,
    organization_id        VARCHAR(255),
    user_id                VARCHAR(255),
    created_by_admin_id    VARCHAR(255),
    assigned_to_admin_id   VARCHAR(255),
    source                 VARCHAR(50)  NOT NULL DEFAULT 'manual',
    category               VARCHAR(100) NOT NULL DEFAULT 'general',
    priority               VARCHAR(50)  NOT NULL DEFAULT 'medium',
    status                 VARCHAR(50)  NOT NULL DEFAULT 'open',
    subject                TEXT         NOT NULL,
    description            TEXT         NOT NULL,
    internal_notes         TEXT         NOT NULL DEFAULT '',
    external_reference     VARCHAR(255)  NOT NULL DEFAULT '',
    resolution_summary     TEXT         NOT NULL DEFAULT '',
    metadata               JSONB,
    resolved_at            TIMESTAMP WITH TIME ZONE,
    closed_at              TIMESTAMP WITH TIME ZONE,
    created_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets (status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_priority ON support_tickets (priority);
CREATE INDEX IF NOT EXISTS idx_support_tickets_source ON support_tickets (source);
CREATE INDEX IF NOT EXISTS idx_support_tickets_org_id ON support_tickets (organization_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_user_id ON support_tickets (user_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assigned_to_admin_id ON support_tickets (assigned_to_admin_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_created_at ON support_tickets (created_at DESC);
