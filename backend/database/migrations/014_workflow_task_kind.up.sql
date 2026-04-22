-- Payment execution as a workflow task.
--
-- Previously, PV status could only go APPROVED -> PAID via a side-channel
-- endpoint with no workflow audit trail: no approver recorded, no signature
-- on the payment act itself. After a PV's final approval stage completes,
-- the workflow engine now auto-creates a task with kind='payment_execution'
-- assigned to the finance role. Completing that task (through the same claim
-- + approve + signature flow used for approvals) flips PV to PAID with a
-- signed, attributed ActionHistory entry.
--
-- kind = 'approval' (default) preserves existing behavior for every other
-- task that already exists or will be created.
ALTER TABLE workflow_tasks
    ADD COLUMN IF NOT EXISTS kind VARCHAR(32) NOT NULL DEFAULT 'approval';

-- Index to quickly find the open payment_execution task for a given PV.
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_kind_entity
    ON workflow_tasks (kind, entity_id)
    WHERE kind <> 'approval';
