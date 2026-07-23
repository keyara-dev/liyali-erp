ALTER TABLE organization_settings
  DROP COLUMN IF EXISTS auto_create_grn_from_po,
  DROP COLUMN IF EXISTS auto_create_pv_from_grn,
  DROP COLUMN IF EXISTS grn_automation_level,
  DROP COLUMN IF EXISTS pv_automation_level,
  DROP COLUMN IF EXISTS auto_approve_max_amount;
