-- Migration: Fix Workflow Tables for UUID Compatibility
-- Description: Drops and recreates workflow tables with correct UUID types
-- Date: 2025-01-06

-- Drop existing tables in correct order (respecting foreign keys)
DROP TABLE IF EXISTS workflow_tasks CASCADE;
DROP TABLE IF EXISTS workflow_assignments CASCADE;
DROP TABLE IF EXISTS workflow_defaults CASCADE;

-- Add new columns to existing workflows table
ALTER TABLE workflows 
  ADD COLUMN IF NOT EXISTS entity_type VARCHAR(100),
  ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
  ADD COLUMN IF NOT EXISTS conditions JSONB,
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Copy document_type to entity_type for backward compatibility
UPDATE workflows SET entity_type = document_type WHERE entity_type IS NULL;

-- Make entity_type not null after copying data
ALTER TABLE workflows ALTER COLUMN entity_type SET NOT NULL;

-- Create workflow_assignments table with UUID workflow_id
CREATE TABLE workflow_assignments (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    workflow_id UUID NOT NULL,
    workflow_version INTEGER NOT NULL,
    current_stage INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'in_progress',
    stage_history JSONB,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    assigned_by VARCHAR(255) NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create workflow_tasks table
CREATE TABLE workflow_tasks (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    workflow_assignment_id VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    stage_number INTEGER NOT NULL,
    stage_name VARCHAR(255) NOT NULL,
    assignment_type VARCHAR(50) DEFAULT 'role',
    assigned_role VARCHAR(100),
    assigned_user_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(50) DEFAULT 'medium',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    claimed_at TIMESTAMP WITH TIME ZONE,
    claimed_by VARCHAR(255),
    completed_at TIMESTAMP WITH TIME ZONE,
    due_date TIMESTAMP WITH TIME ZONE
);

-- Create workflow_defaults table with UUID workflow_id
CREATE TABLE workflow_defaults (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    default_workflow_id UUID NOT NULL,
    default_workflow_version INTEGER NOT NULL,
    set_by VARCHAR(255) NOT NULL,
    set_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for new columns in workflows
CREATE INDEX IF NOT EXISTS idx_workflows_entity_type ON workflows(entity_type);
CREATE INDEX IF NOT EXISTS idx_workflows_deleted_at ON workflows(deleted_at);

-- Create indexes for workflow_assignments
CREATE INDEX idx_workflow_assignments_organization_id ON workflow_assignments(organization_id);
CREATE INDEX idx_workflow_assignments_entity_id ON workflow_assignments(entity_id);
CREATE INDEX idx_workflow_assignments_entity_type ON workflow_assignments(entity_type);
CREATE INDEX idx_workflow_assignments_workflow_id ON workflow_assignments(workflow_id);
CREATE INDEX idx_workflow_assignments_status ON workflow_assignments(status);
CREATE INDEX idx_workflow_assignments_assigned_by ON workflow_assignments(assigned_by);

-- Create indexes for workflow_tasks
CREATE INDEX idx_workflow_tasks_organization_id ON workflow_tasks(organization_id);
CREATE INDEX idx_workflow_tasks_assignment_id ON workflow_tasks(workflow_assignment_id);
CREATE INDEX idx_workflow_tasks_entity_id ON workflow_tasks(entity_id);
CREATE INDEX idx_workflow_tasks_status ON workflow_tasks(status);
CREATE INDEX idx_workflow_tasks_assigned_role ON workflow_tasks(assigned_role);
CREATE INDEX idx_workflow_tasks_assigned_user_id ON workflow_tasks(assigned_user_id);
CREATE INDEX idx_workflow_tasks_claimed_by ON workflow_tasks(claimed_by);

-- Create indexes for workflow_defaults
CREATE INDEX idx_workflow_defaults_organization_id ON workflow_defaults(organization_id);
CREATE INDEX idx_workflow_defaults_entity_type ON workflow_defaults(entity_type);
CREATE INDEX idx_workflow_defaults_workflow_id ON workflow_defaults(default_workflow_id);

-- Create unique constraint for workflow defaults per organization and entity type
CREATE UNIQUE INDEX idx_workflow_defaults_unique_org_entity 
ON workflow_defaults(organization_id, entity_type);

-- Create composite indexes for better query performance
CREATE INDEX idx_workflows_org_entity_active 
ON workflows(organization_id, entity_type, is_active);

CREATE INDEX idx_workflows_org_entity_default 
ON workflows(organization_id, entity_type, is_default);

CREATE INDEX idx_workflow_assignments_org_entity_status 
ON workflow_assignments(organization_id, entity_id, entity_type, status);

CREATE INDEX idx_workflow_tasks_org_role_status 
ON workflow_tasks(organization_id, assigned_role, status);

-- Add foreign key constraints
ALTER TABLE workflow_assignments 
ADD CONSTRAINT fk_workflow_assignments_organization_id 
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE workflow_assignments 
ADD CONSTRAINT fk_workflow_assignments_workflow_id 
FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE;

ALTER TABLE workflow_tasks 
ADD CONSTRAINT fk_workflow_tasks_organization_id 
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE workflow_tasks 
ADD CONSTRAINT fk_workflow_tasks_assignment_id 
FOREIGN KEY (workflow_assignment_id) REFERENCES workflow_assignments(id) ON DELETE CASCADE;

ALTER TABLE workflow_defaults 
ADD CONSTRAINT fk_workflow_defaults_organization_id 
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE workflow_defaults 
ADD CONSTRAINT fk_workflow_defaults_workflow_id 
FOREIGN KEY (default_workflow_id) REFERENCES workflows(id) ON DELETE CASCADE;

-- Add comments for documentation
COMMENT ON TABLE workflows IS 'Enhanced workflow definitions with frontend compatibility';
COMMENT ON TABLE workflow_assignments IS 'Tracks workflow execution for specific entities';
COMMENT ON TABLE workflow_tasks IS 'Individual approval tasks within workflow assignments';
COMMENT ON TABLE workflow_defaults IS 'Default workflow mappings for entity types per organization';

COMMENT ON COLUMN workflows.entity_type IS 'Type of entity this workflow applies to (requisition, purchase_order, etc.)';
COMMENT ON COLUMN workflows.conditions IS 'JSON conditions for when this workflow should be applied';
COMMENT ON COLUMN workflows.stages IS 'JSON array of workflow stages with approval requirements';
COMMENT ON COLUMN workflow_assignments.stage_history IS 'JSON array of completed stage executions';
COMMENT ON COLUMN workflow_tasks.assignment_type IS 'How the task is assigned: role or specific_user';