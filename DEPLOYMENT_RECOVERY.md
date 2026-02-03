# Deployment Recovery Guide

## Issue: Fly.io Deployment Timeout

The deployment was failing due to the release command (`go run database/migrate_all.go`) timing out during the deployment process.

## Solution Applied

1. **Removed automatic migration from deployment**: The `release_command` has been removed from `backend/fly.toml` to prevent deployment timeouts.

2. **Separated migration from deployment**: Migrations are now run as a separate step after deployment succeeds.

## Manual Migration (if needed)

If the automatic migration in the GitHub workflow fails, you can run it manually:

```bash
# Connect to the deployed app and run migrations
flyctl ssh console --app liyali-gateway-api -C "go run database/migrate_all.go"
```

## Deployment Status Check

The app should deploy successfully even without migrations. You can verify:

```bash
# Check app status
flyctl status --app liyali-gateway-api

# Test health endpoint (doesn't require database)
curl https://liyali-gateway-api.fly.dev/health
```

## Current Migration Status

The current migration (`010_performance_optimization_minimal.up.sql`) contains only 3 critical indexes:

- `idx_org_members_user_active` - Organization members JOIN optimization
- `idx_requisitions_org_status` - Requisitions status for analytics
- `idx_sessions_expires` - Session cleanup

These are lightweight and should execute quickly.

## Next Steps

1. Deploy the app (should succeed now)
2. Migrations will run automatically via GitHub workflow
3. If migration fails, run manually using the command above
4. Verify the app is working by testing the health endpoint

## Rollback Plan

If issues persist, you can:

1. Skip the current migration by renaming it to `.skip`
2. Deploy without any new migrations
3. Run migrations during a maintenance window
