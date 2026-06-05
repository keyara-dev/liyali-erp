-- Revert 021_perf_indexes
DROP INDEX IF EXISTS idx_workflow_tasks_claimed_expiry;
