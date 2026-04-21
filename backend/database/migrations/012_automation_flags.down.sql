-- Rollback automation flags from organization_settings table

ALTER TABLE organization_settings 
DROP COLUMN IF EXISTS auto_submit_grn_to_workflow,
DROP COLUMN IF EXISTS auto_submit_pv_to_workflow,
DROP COLUMN IF EXISTS auto_create_pv_from_po;
