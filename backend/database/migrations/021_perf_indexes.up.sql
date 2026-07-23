-- ============================================================================
-- Performance indexes
-- Migration: 021_perf_indexes
--
-- The claim-expiry sweep worker (StartClaimExpiryWorker, runs every 60s) filters
--   WHERE UPPER(status) = 'CLAIMED' AND claim_expiry < now()
-- The UPPER(status) call is non-sargable, so the existing plain
-- idx_workflow_tasks_status cannot be used and the sweep falls back to a full
-- sequential scan of workflow_tasks on every tick.
--
-- A partial index keyed on claim_expiry, restricted to claimed rows via the same
-- immutable UPPER(status) predicate, lets the sweep locate expired claims with an
-- index range scan instead. It also stays tiny: only currently-claimed rows are
-- indexed, so writes to non-claimed tasks don't touch it.
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_workflow_tasks_claimed_expiry
    ON workflow_tasks (claim_expiry)
    WHERE UPPER(status) = 'CLAIMED';
