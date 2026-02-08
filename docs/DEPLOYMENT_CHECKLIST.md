# Deployment Checklist - Admin Console Integration

## ✅ Pre-Deployment Checklist

### 1. GitHub Secrets Configuration

Ensure these secrets are set in GitHub repository settings:

- [ ] `FLY_API_TOKEN` - Your Fly.io API token
- [ ] `FLY_DATABASE_URL` - PostgreSQL connection string
- [ ] `JWT_SECRET` - JWT signing secret
- [ ] `NEXTAUTH_SECRET` - NextAuth signing secret
- [ ] `FLY_CORS_ALLOWED_ORIGINS` - Updated to include admin console URL

**CORS Origins should include:**

```
https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev
```

### 2. Fly.io App Creation

- [ ] Backend app exists: `liyali-gateway-api`
- [ ] Frontend app exists: `liyali-gateway-frontend`
- [ ] Admin console app created: `liyali-admin-console`

**Create admin console app:**

```bash
cd admin-console
flyctl apps create liyali-admin-console --org your-org
```

### 3. Environment Variables

#### Backend

- [ ] `DATABASE_URL` set
- [ ] `JWT_SECRET` set
- [ ] `CORS_ALLOWED_ORIGINS` includes admin console

#### Frontend

- [ ] `NEXT_PUBLIC_API_URL` set
- [ ] `DATABASE_URL` set
- [ ] `NEXTAUTH_SECRET` set
- [ ] `NEXTAUTH_URL` set

#### Admin Console

- [ ] `NEXT_PUBLIC_API_URL` set
- [ ] `NEXTAUTH_SECRET` set
- [ ] `NEXTAUTH_URL` set

### 4. Code Changes

- [x] `admin-console/Dockerfile` created
- [x] `admin-console/fly.toml` created
- [x] `admin-console/next.config.ts` updated with standalone output
- [x] `.github/workflows/fly-deploy.yml` updated
- [x] `docker-compose.yml` updated
- [x] Documentation updated

## 🚀 Deployment Steps

### Option 1: Automatic Deployment (Recommended)

1. [ ] Commit all changes
2. [ ] Push to `develop` branch
3. [ ] Monitor GitHub Actions workflow
4. [ ] Verify deployment in GitHub Actions summary

```bash
git add .
git commit -m "Add admin console to deployment workflow"
git push origin develop
```

### Option 2: Manual Deployment

#### Step 1: Deploy Backend (if needed)

```bash
cd backend
flyctl deploy --remote-only
```

#### Step 2: Deploy Frontend (if needed)

```bash
cd frontend
flyctl deploy --remote-only
```

#### Step 3: Deploy Admin Console

```bash
cd admin-console

# Set secrets first
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  NEXTAUTH_SECRET="your-secret" \
  NEXTAUTH_URL="https://liyali-admin-console.fly.dev" \
  --app liyali-admin-console

# Deploy
flyctl deploy --remote-only
```

#### Step 4: Update Backend CORS

```bash
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api
```

### Option 3: Using Deployment Script

```bash
# Linux/Mac
./scripts/deploy-admin-console.sh

# Windows
.\scripts\deploy-admin-console.ps1
```

## 🧪 Post-Deployment Verification

### 1. Health Checks

- [ ] Backend health check passes

```bash
curl https://liyali-gateway-api.fly.dev/health
```

- [ ] Frontend loads successfully

```bash
curl https://liyali-gateway-frontend.fly.dev/
```

- [ ] Admin console loads successfully

```bash
curl https://liyali-admin-console.fly.dev/
```

### 2. Connectivity Tests

- [ ] Admin console can reach backend API
- [ ] Login works on admin console
- [ ] No CORS errors in browser console
- [ ] API requests succeed

### 3. Functionality Tests

- [ ] Dashboard loads
- [ ] Organizations list displays
- [ ] Trial management works
- [ ] Subscription management accessible
- [ ] User management functional

### 4. Monitoring

- [ ] Check logs for errors

```bash
flyctl logs --app liyali-admin-console
```

- [ ] Verify app status

```bash
flyctl status --app liyali-admin-console
```

- [ ] Check metrics

```bash
flyctl metrics --app liyali-admin-console
```

## 🔍 Troubleshooting

### Issue: CORS Errors

**Symptoms:** Browser console shows CORS errors when admin console tries to reach backend

**Solution:**

```bash
# Check current CORS settings
flyctl secrets list --app liyali-gateway-api

# Update CORS to include admin console
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api

# Restart backend
flyctl restart --app liyali-gateway-api
```

### Issue: Build Fails

**Symptoms:** Deployment fails during build phase

**Solution:**

```bash
# Check build logs
flyctl logs --app liyali-admin-console

# Common fixes:
# 1. Ensure package.json has all dependencies
# 2. Check TypeScript errors
# 3. Verify next.config.ts is valid
# 4. Try building locally first: npm run build
```

### Issue: App Won't Start

**Symptoms:** Deployment succeeds but app doesn't start

**Solution:**

```bash
# Check logs
flyctl logs --app liyali-admin-console

# Check status
flyctl status --app liyali-admin-console

# Restart app
flyctl restart --app liyali-admin-console

# Verify secrets are set
flyctl secrets list --app liyali-admin-console
```

### Issue: Health Check Fails

**Symptoms:** Deployment reports health check failure

**Solution:**

```bash
# Test health endpoint directly
curl -v https://liyali-admin-console.fly.dev/

# Check if app is running
flyctl status --app liyali-admin-console

# View recent logs
flyctl logs --app liyali-admin-console --lines 100
```

## 📊 Monitoring Setup

### Set Up Alerts

```bash
# Create alert for app down
flyctl alerts create \
  --app liyali-admin-console \
  --type app_down \
  --email your-email@example.com
```

### Regular Checks

- [ ] Set up uptime monitoring (e.g., UptimeRobot, Pingdom)
- [ ] Configure log aggregation
- [ ] Set up error tracking (e.g., Sentry)
- [ ] Monitor resource usage

## 🎯 Success Criteria

Deployment is successful when:

- [x] All three apps are deployed and running
- [x] Health checks pass for all apps
- [x] Admin console can communicate with backend
- [x] No CORS errors
- [x] Login works on admin console
- [x] All admin features are accessible
- [x] Logs show no critical errors
- [x] GitHub Actions workflow completes successfully

## 📝 Notes

- Admin console runs on port 3001
- Uses same authentication system as frontend
- Shares backend API with frontend
- Requires admin-level permissions to access
- Auto-scales based on traffic
- Deployed to same region as backend (jnb)

## 🔄 Rollback Plan

If deployment fails:

1. Check GitHub Actions for error details
2. Review logs: `flyctl logs --app liyali-admin-console`
3. Rollback if needed: `flyctl releases rollback --app liyali-admin-console`
4. Fix issues and redeploy

## 📞 Support

- Fly.io Docs: https://fly.io/docs/
- GitHub Actions Logs: Check repository Actions tab
- Project Issues: Open issue in repository

---

**Last Updated:** February 8, 2026
**Status:** Ready for Deployment
