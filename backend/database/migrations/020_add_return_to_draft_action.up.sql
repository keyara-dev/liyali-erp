-- Add 'returned_to_draft' and 'returned_for_revision' to the allowed actions
-- in stage_approval_records. This supports the new rejection flow where a
-- document can be returned to draft, returned to a previous stage for revision,
-- or fully rejected.

ALTER TABLE stage_approval_records
    DROP CONSTRAINT IF EXISTS stage_approval_records_action_check;

ALTER TABLE stage_approval_records
    ADD CONSTRAINT stage_approval_records_action_check
    CHECK (action IN ('approved', 'rejected', 'returned_to_draft', 'returned_for_revision'));
