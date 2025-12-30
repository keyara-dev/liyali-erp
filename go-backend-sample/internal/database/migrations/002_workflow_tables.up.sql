-- Create workflows table
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_type VARCHAR(50) NOT NULL CHECK (document_type IN (
        'REQUISITION', 'BUDGET', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER', 'GRN'
    )),
    stages JSONB NOT NULL,  -- Array of stage definitions
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create documents table
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type VARCHAR(50) NOT NULL CHECK (document_type IN (
        'REQUISITION', 'BUDGET', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER', 'GRN'
    )),
    document_number VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) NOT NULL CHECK (status IN (
        'DRAFT', 'SUBMITTED', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'COMPLETED'
    )),
    created_by UUID NOT NULL REFERENCES users(id),
    department VARCHAR(100),
    workflow_id UUID REFERENCES workflows(id),
    data JSONB NOT NULL,  -- Type-specific fields
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    submitted_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Create approval_tasks table
CREATE TABLE approval_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    assigned_to UUID NOT NULL REFERENCES users(id),
    assigned_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL CHECK (status IN (
        'PENDING', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'REASSIGNED'
    )),
    current_stage INT NOT NULL DEFAULT 1,
    total_stages INT NOT NULL DEFAULT 3,
    priority VARCHAR(20) CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT')),
    due_date TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create approval_history table
CREATE TABLE approval_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES approval_tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL CHECK (action IN (
        'APPROVED', 'REJECTED', 'REASSIGNED', 'COMMENTED'
    )),
    stage INT NOT NULL,
    comment TEXT,
    signature TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create audit_logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    changes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN (
        'TASK_ASSIGNED', 'TASK_APPROVED', 'TASK_REJECTED',
        'TASK_REASSIGNED', 'TASK_COMMENTED', 'TASK_DUE_SOON'
    )),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    related_id UUID,
    is_read BOOLEAN DEFAULT false,
    sent_via_email BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for workflows
CREATE INDEX idx_workflows_document_type_active ON workflows(document_type, is_active);
CREATE INDEX idx_workflows_created_by ON workflows(created_by);

-- Create indexes for documents
CREATE INDEX idx_documents_type_status ON documents(document_type, status);
CREATE INDEX idx_documents_created_by ON documents(created_by);
CREATE INDEX idx_documents_workflow_id ON documents(workflow_id);
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);
CREATE INDEX idx_documents_data ON documents USING GIN (data);

-- Create indexes for approval_tasks
CREATE INDEX idx_approval_tasks_assigned_to_status ON approval_tasks(assigned_to, status);
CREATE INDEX idx_approval_tasks_document_id ON approval_tasks(document_id);
CREATE INDEX idx_approval_tasks_status_stage ON approval_tasks(status, current_stage);
CREATE INDEX idx_approval_tasks_created_at ON approval_tasks(created_at DESC);

-- Create indexes for approval_history
CREATE INDEX idx_approval_history_task_id_created_at ON approval_history(task_id, created_at DESC);
CREATE INDEX idx_approval_history_user_id ON approval_history(user_id);

-- Create indexes for audit_logs
CREATE INDEX idx_audit_logs_user_id_created_at ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Create indexes for notifications
CREATE INDEX idx_notifications_user_id_read ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_related_id ON notifications(related_id);

-- Add triggers for updated_at columns
CREATE TRIGGER update_workflows_updated_at BEFORE UPDATE ON workflows
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_approval_tasks_updated_at BEFORE UPDATE ON approval_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
