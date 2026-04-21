-- Add automation opt-in flags to organization_settings table
-- These flags allow organizations to enable automatic workflow submission

ALTER TABLE organization_settings 
ADD COLUMN IF NOT EXISTS auto_submit_grn_to_workflow BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_submit_pv_to_workflow BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_create_pv_from_po BOOLEAN DEFAULT FALSE;

-- Add comments for documentation
COMMENT ON COLUMN organization_settings.auto_submit_grn_to_workflow IS 
'When enabled, auto-created GRNs are automatically submitted to workflow instead of staying in DRAFT';

COMMENT ON COLUMN organization_settings.auto_submit_pv_to_workflow IS 
'When enabled, created PVs are automatically submitted to workflow instead of staying in DRAFT';

COMMENT ON COLUMN organization_settings.auto_create_pv_from_po IS 
'When enabled, PVs are automatically created from approved POs in payment-first flow';
