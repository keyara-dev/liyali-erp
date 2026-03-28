-- Add organization scoping and richer actor/details columns to audit_logs
-- Also adds: man_number + position to stage_approval_records (Group 6)

-- audit_logs: org scope + actor name/role + structured details column
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS organization_id VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_name     VARCHAR(255)  NOT NULL DEFAULT '';
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_role     VARCHAR(100)  NOT NULL DEFAULT '';
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS details        JSONB;

CREATE INDEX IF NOT EXISTS idx_audit_logs_org_id      ON audit_logs (organization_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_doc_type     ON audit_logs (document_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at   ON audit_logs (created_at DESC);

-- stage_approval_records: approver man number + position
ALTER TABLE stage_approval_records ADD COLUMN IF NOT EXISTS man_number VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE stage_approval_records ADD COLUMN IF NOT EXISTS position   VARCHAR(255) NOT NULL DEFAULT '';
