# Admin Console Deployment Setup - Complete

## ✅ What Was Done

The admin console has been successfully integrated into the Fly.io deployment workflow. Here's what was configured:

### 1. Docker Configuration

**Created: `admin-console/Dockerfile`**

- Multi-stage build for optimal image size
- Node.js 20 Alpine base
- Standalone Next.js output
- Production-ready configuration
- Port 3001 exposed

**Updated: `admin-console/next.config.ts`**

- Added `output: "standalone"` for Docker deployment
- Maintains existing rewrites and optimizations

### 2. Fly.io Configuration

**Created: `admin-console/fly.toml`**

- App name: `liyali-admin-console`
- Region: `jnb` (Johannesburg - same as backend)
- Port: 3001
- Auto-scaling enabled
- Health checks configured
- Shared CPU with 512MB memory

### 3. GitHub Actions Workflow

**Updated: `.github/workflows/fly-deploy.yml`**

Added admin console deployment with:

- Change detection for `admin-console/**` files
- Separate deployment job `deploy-admin`
- Environment variable configuration
- Health check verification
- Deployment summary reporting
- Manual trigger option

**Key Features:**

- Selective deployment (only deploys when admin console changes)
- Depends on backend deployment success
- Automatic secret management
- Build-time API URL injection
- Comprehensive error handling

### 4. Local Development

**Updated: `docker-compose.yml`**

- Added admin console service
- Port 3001 mapped
- Connected to backend
- Volume mounts for hot reload
- Network integration

### 5. Documentation

**Updated: `docs/FLY_IO_DEPLOYMENT_GUIDE.md`**

- Admin console deployment instructions
- Environment variable setup
- Custom domain configuration
- Monitoring and debugging
- CORS configuration updates
- Architecture overview

## 🚀 Deployment URLs

Once deployed, the applications will be available at:

- **Backend API**: https://liyali-gateway-api.fly.dev
- **Frontend**: https://liyali-gateway-frontend.fly.dev
- **Admin Console**: https://liyali-admin-console.fly.dev

## 🔧 Required GitHub Secrets

Ensure these secrets are set in your GitHub repository:

```bash
FLY_API_TOKEN                 # Your Fly.io API token
FLY_DATABASE_URL             # PostgreSQL connection string
JWT_SECRET                   # JWT signing secret
NEXTAUTH_SECRET              # NextAuth signing secret
FLY_CORS_ALLOWED_ORIGINS     # Must include admin console URL
```

**Important**: Update `FLY_CORS_ALLOWED_ORIGINS` to include:

```
https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev
```

## 📋 Deployment Steps

### First-Time Setup

1. **Create the Fly.io app** (one-time):

```bash
cd admin-console
flyctl apps create liyali-admin-console --org your-org
```

2. **Set environment secrets**:

```bash
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  NEXTAUTH_SECRET="your-nextauth-secret" \
  NEXTAUTH_URL="https://liyali-admin-console.fly.dev" \
  --app liyali-admin-console
```

3. **Update backend CORS** to allow admin console:

```bash
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api
```

### Automatic Deployment

The admin console will automatically deploy when:

- Changes are pushed to `develop` branch in `admin-console/**` directory
- Workflow file is modified
- Manual workflow trigger with admin console selected

### Manual Deployment

```bash
cd admin-console
flyctl deploy --remote-only
```

## 🧪 Testing the Deployment

After deployment, verify:

1. **Health Check**:

```bash
curl https://liyali-admin-console.fly.dev/
```

2. **API Connectivity**:

- Open https://liyali-admin-console.fly.dev
- Try logging in
- Check browser console for API errors

3. **Backend CORS**:

```bash
curl -H "Origin: https://liyali-admin-console.fly.dev" \
     -H "Access-Control-Request-Method: POST" \
     -X OPTIONS \
     https://liyali-gateway-api.fly.dev/api/v1/auth/login
```

## 🔍 Monitoring

### View Logs

```bash
flyctl logs --app liyali-admin-console -f
```

### Check Status

```bash
flyctl status --app liyali-admin-console
```

### View Metrics

```bash
flyctl metrics --app liyali-admin-console
```

## 🐛 Troubleshooting

### Issue: Admin console can't connect to backend

**Solution**: Check CORS configuration

```bash
# View current CORS settings
flyctl secrets list --app liyali-gateway-api

# Update if needed
flyctl secrets set CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" --app liyali-gateway-api
```

### Issue: Build fails

**Solution**: Check build logs

```bash
flyctl logs --app liyali-admin-console
```

Common causes:

- Missing dependencies in package.json
- Build errors in TypeScript
- Environment variables not set

### Issue: Health check fails

**Solution**: Verify the app is running

```bash
# Check app status
flyctl status --app liyali-admin-console

# Restart if needed
flyctl restart --app liyali-admin-console

# Check logs
flyctl logs --app liyali-admin-console
```

## 🎯 Workflow Features

### Change Detection

The workflow intelligently detects changes:

- ✅ Only deploys admin console when `admin-console/**` files change
- ✅ Skips deployment if no changes detected
- ✅ Deploys all apps if workflow file changes

### Manual Control

You can manually trigger deployment:

1. Go to GitHub Actions
2. Select "Deploy to Staging Environment - (Fly.io)"
3. Click "Run workflow"
4. Choose which apps to deploy:
   - ☑️ Deploy backend
   - ☑️ Deploy frontend
   - ☑️ Deploy admin console

### Deployment Summary

After each deployment, the workflow provides:

- Change detection results
- Deployment status for each app
- URLs for all deployed apps
- Efficiency metrics
- Next steps

## 📊 Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Fly.io Platform                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────────┐      ┌──────────────────┐       │
│  │   Frontend       │      │  Admin Console   │       │
│  │   (Next.js)      │      │   (Next.js)      │       │
│  │   Port: 3000     │      │   Port: 3001     │       │
│  └────────┬─────────┘      └────────┬─────────┘       │
│           │                         │                  │
│           └─────────┬───────────────┘                  │
│                     │                                  │
│           ┌─────────▼─────────┐                        │
│           │   Backend API     │                        │
│           │   (Go/Fiber)      │                        │
│           │   Port: 8080      │                        │
│           └─────────┬─────────┘                        │
│                     │                                  │
│           ┌─────────▼─────────┐                        │
│           │   PostgreSQL      │                        │
│           │   Database        │                        │
│           └───────────────────┘                        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## 🔐 Security Considerations

1. **CORS Configuration**: Ensure backend CORS includes admin console URL
2. **Authentication**: Admin console uses same auth system as frontend
3. **Secrets Management**: All secrets stored in Fly.io secrets (not in code)
4. **HTTPS Only**: All apps force HTTPS
5. **Access Control**: Consider IP whitelisting for admin console (optional)

## 📝 Next Steps

1. ✅ Push changes to `develop` branch to trigger deployment
2. ✅ Monitor GitHub Actions for deployment progress
3. ✅ Verify all three apps are running
4. ✅ Test admin console functionality
5. ✅ Set up custom domains (optional)
6. ✅ Configure monitoring and alerts

## 🎉 Summary

The admin console is now fully integrated into the deployment pipeline with:

- ✅ Automated Docker builds
- ✅ Fly.io configuration
- ✅ GitHub Actions workflow
- ✅ Change detection
- ✅ Health checks
- ✅ Local development support
- ✅ Comprehensive documentation

The deployment is production-ready and follows the same patterns as the backend and frontend applications.

---

**Created**: February 8, 2026
**Status**: ✅ Complete and Ready for Deployment
