DROP INDEX IF EXISTS idx_workflow_tasks_kind_entity;
ALTER TABLE workflow_tasks DROP COLUMN IF EXISTS kind;
