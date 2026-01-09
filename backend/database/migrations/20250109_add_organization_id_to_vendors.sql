-- Migration: Add OrganizationID to vendors table for multi-tenant security
-- Date: 2025-01-09
-- Purpose: Fix critical security vulnerability where vendors were not organization-scoped

-- Add organization_id column to vendors table
ALTER TABLE vendors 
ADD COLUMN organization_id VARCHAR(255);

-- Create index for performance
CREATE INDEX idx_vendors_organization_id ON vendors(organization_id);

-- Update existing vendors to belong to the first organization (temporary fix)
-- In production, this should be done more carefully based on business logic
UPDATE vendors 
SET organization_id = (
    SELECT id 
    FROM organizations 
    ORDER BY created_at ASC 
    LIMIT 1
) 
WHERE organization_id IS NULL;

-- Make organization_id NOT NULL after updating existing records
ALTER TABLE vendors 
ALTER COLUMN organization_id SET NOT NULL;

-- Add foreign key constraint
ALTER TABLE vendors 
ADD CONSTRAINT fk_vendors_organization 
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

-- Update unique constraint for vendor_code to be organization-scoped
DROP INDEX IF EXISTS idx_vendors_vendor_code;
CREATE UNIQUE INDEX idx_org_vendor_code ON vendors(organization_id, vendor_code);

-- Update email index to be organization-scoped for better performance
DROP INDEX IF EXISTS idx_vendors_email;
CREATE INDEX idx_vendors_email_org ON vendors(email, organization_id);