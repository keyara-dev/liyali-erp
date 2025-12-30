-- Drop triggers
DROP TRIGGER IF EXISTS update_approval_tasks_updated_at ON approval_tasks;
DROP TRIGGER IF EXISTS update_documents_updated_at ON documents;
DROP TRIGGER IF EXISTS update_workflows_updated_at ON workflows;

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS approval_history;
DROP TABLE IF EXISTS approval_tasks;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS workflows;
