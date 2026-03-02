-- Revert: remove 'returned_to_draft' from allowed actions
-- WARNING: This will fail if any rows have action='returned_to_draft'

ALTER TABLE stage_approval_records
    DROP CONSTRAINT IF EXISTS stage_approval_records_action_check;

ALTER TABLE stage_approval_records
    ADD CONSTRAINT stage_approval_records_action_check
    CHECK (action IN ('approved', 'rejected'));
