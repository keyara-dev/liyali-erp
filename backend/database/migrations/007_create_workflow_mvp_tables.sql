-- Migration: Create MVP Workflow Tables
-- Description: Creates tables for the MVP workflow system that integrates with the existing frontend UI
-- Date: 2025-01-01

-- Create workflows table
CREATE TABLE IF NOT EXISTS workflows (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    entity_type VARCHAR(100) NOT NULL,
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    conditions JSONB,
    stages JSONB NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for workflows
CREATE INDEX IF NOT EXISTS idx_workflows_entity_type ON workflows(entity_type);
CREATE INDEX IF NOT EXISTS idx_workflows_is_active ON workflows(is_active);
CREATE INDEX IF NOT EXISTS idx_workflows_is_default ON workflows(is_default);
CREATE INDEX IF NOT EXISTS idx_workflows_created_by ON workflows(created_by);
CREATE INDEX IF NOT EXISTS idx_workflows_deleted_at ON workflows(deleted_at);

-- Create workflow_assignments table
CREATE TABLE IF NOT EXISTS workflow_assignments (
    id VARCHAR(255) PRIMARY KEY,
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    workflow_id VARCHAR(255) NOT NULL,
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

-- Create indexes for workflow_assignments
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_entity_id ON workflow_assignments(entity_id);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_entity_type ON workflow_assignments(entity_type);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_workflow_id ON workflow_assignments(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_status ON workflow_assignments(status);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_assigned_by ON workflow_assignments(assigned_by);

-- Create workflow_tasks table
CREATE TABLE IF NOT EXISTS workflow_tasks (
    id VARCHAR(255) PRIMARY KEY,
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

-- Create indexes for workflow_tasks
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assignment_id ON workflow_tasks(workflow_assignment_id);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_entity_id ON workflow_tasks(entity_id);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_status ON workflow_tasks(status);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assigned_role ON workflow_tasks(assigned_role);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assigned_user_id ON workflow_tasks(assigned_user_id);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_claimed_by ON workflow_tasks(claimed_by);

-- Create workflow_defaults table
CREATE TABLE IF NOT EXISTS workflow_defaults (
    id VARCHAR(255) PRIMARY KEY,
    entity_type VARCHAR(100) UNIQUE NOT NULL,
    default_workflow_id VARCHAR(255) NOT NULL,
    default_workflow_version INTEGER NOT NULL,
    set_by VARCHAR(255) NOT NULL,
    set_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for workflow_defaults
CREATE INDEX IF NOT EXISTS idx_workflow_defaults_entity_type ON workflow_defaults(entity_type);
CREATE INDEX IF NOT EXISTS idx_workflow_defaults_workflow_id ON workflow_defaults(default_workflow_id);

-- Add foreign key constraints
ALTER TABLE workflow_assignments 
ADD CONSTRAINT fk_workflow_assignments_workflow_id 
FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE;

ALTER TABLE workflow_tasks 
ADD CONSTRAINT fk_workflow_tasks_assignment_id 
FOREIGN KEY (workflow_assignment_id) REFERENCES workflow_assignments(id) ON DELETE CASCADE;

ALTER TABLE workflow_defaults 
ADD CONSTRAINT fk_workflow_defaults_workflow_id 
FOREIGN KEY (default_workflow_id) REFERENCES workflows(id) ON DELETE CASCADE;

-- Create unique constraint to ensure only one default per entity type
CREATE UNIQUE INDEX IF NOT EXISTS idx_workflow_defaults_unique_entity_type 
ON workflow_defaults(entity_type);

-- Create composite indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_workflows_entity_type_active 
ON workflows(entity_type, is_active);

CREATE INDEX IF NOT EXISTS idx_workflows_entity_type_default 
ON workflows(entity_type, is_default);

CREATE INDEX IF NOT EXISTS idx_workflow_assignments_entity_status 
ON workflow_assignments(entity_id, entity_type, status);

CREATE INDEX IF NOT EXISTS idx_workflow_tasks_role_status 
ON workflow_tasks(assigned_role, status);

-- Add comments for documentation
COMMENT ON TABLE workflows IS 'MVP workflow definitions with configurable stages and conditions';
COMMENT ON TABLE workflow_assignments IS 'Tracks workflow execution for specific entities';
COMMENT ON TABLE workflow_tasks IS 'Individual approval tasks within workflow assignments';
COMMENT ON TABLE workflow_defaults IS 'Default workflow mappings for entity types';

COMMENT ON COLUMN workflows.entity_type IS 'Type of entity this workflow applies to (requisition, purchase_order, etc.)';
COMMENT ON COLUMN workflows.conditions IS 'JSON conditions for when this workflow should be applied';
COMMENT ON COLUMN workflows.stages IS 'JSON array of workflow stages with approval requirements';
COMMENT ON COLUMN workflow_assignments.stage_history IS 'JSON array of completed stage executions';
COMMENT ON COLUMN workflow_tasks.assignment_type IS 'How the task is assigned: role or specific_user';