# Admin Console - Build & Deployment Guide

## ✅ Build Status: SUCCESS

**Date**: February 8, 2026
**Build Time**: ~71 seconds total
**Next.js Version**: 16.1.6
**Build Tool**: Turbopack
**Mode**: Production
**Status**: ✅ Production Ready

---

## 📊 Build Statistics

### Compilation

- **Compile Time**: 27.0 seconds
- **TypeScript Check**: 44 seconds
- **Page Data Collection**: 4.3 seconds (19 workers)
- **Static Generation**: 1.3 seconds (19 workers)
- **Optimization**: 2.3 seconds

### Routes Built (20 total)

#### Static Routes (1)

- `/_not-found` - 404 page

#### Dynamic Routes (14 admin pages)

- `/` - Root redirect
- `/admin/dashboard` - Admin dashboard
- `/admin/organizations` - Organization management
- `/admin/subscriptions` - Subscription management
- `/admin/users` - User management
- `/admin/admin-users` - Admin user management
- `/admin/roles` - Role management
- `/admin/analytics` - Analytics dashboard
- `/admin/audit-logs` - Audit log viewer
- `/admin/api-monitoring` - API monitoring
- `/admin/system-health` - System health
- `/admin/database` - Database management
- `/admin/feature-flags` - Feature flag management
- `/admin/settings` - System settings

#### Auth Routes (5)

- `/login` - Login page
- `/forgot-password` - Password reset request
- `/reset-password` - Password reset
- `/apple-icon` - Apple touch icon
- `/icon` - App icon

#### Middleware

- Proxy middleware for API requests

### Build Output

```
admin-console/.next/
├── standalone/
│   ├── server.js          (6.6 KB) - Production server
│   ├── package.json       - Dependencies
│   └── node_modules/      - Runtime dependencies
├── static/                - Static assets (CSS, JS, images)
├── server/                - Server components
└── [other build artifacts]
```

### Build Size Estimates

- **Standalone Server**: ~6.6 KB (entry point)
- **Node Modules**: ~50-100 MB (runtime dependencies)
- **Static Assets**: ~2-5 MB (CSS, JS, fonts)
- **Total Docker Image**: ~150-200 MB (estimated)

---

## 🚀 Deployment Options

### Option 1: Automatic Deployment (Recommended)

**Prerequisites:**

1. Create Fly.io app (one-time)
2. Set GitHub secrets
3. Update backend CORS

**Steps:**

```bash
# 1. Create Fly.io app
cd admin-console
flyctl apps create liyali-admin-console

# 2. Set secrets
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  NEXTAUTH_SECRET="$(openssl rand -base64 32)" \
  NEXTAUTH_URL="https://liyali-admin-console.fly.dev" \
  --app liyali-admin-console

# 3. Update backend CORS
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api

# 4. Push to GitHub
git add .
git commit -m "Deploy admin console to Fly.io"
git push origin develop

# 5. Monitor deployment
# Go to: https://github.com/your-repo/actions
```

**Time**: ~15 minutes (setup) + ~5 minutes (deployment)

---

### Option 2: Manual Deployment

**Using Deployment Script:**

```bash
# Linux/Mac
./scripts/deploy-admin-console.sh

# Windows
.\scripts\deploy-admin-console.ps1

# Choose option 7: Full setup (secrets + deploy)
```

**Time**: ~10 minutes

---

### Option 3: Direct Fly.io Deployment

```bash
cd admin-console

# Deploy directly
flyctl deploy --remote-only

# Monitor deployment
flyctl logs -f
```

**Time**: ~5 minutes

---

## 📋 Deployment Checklist

### Before Deployment

- [ ] **Fly.io CLI installed** (`flyctl version`)
- [ ] **Logged in to Fly.io** (`flyctl auth whoami`)
- [ ] **Backend is deployed** (https://liyali-gateway-api.fly.dev/health)
- [ ] **Database is running** (PostgreSQL on Fly.io)
- [ ] **Production build completed** (`npm run build`)

### During Deployment

- [ ] **Create Fly.io app** (`flyctl apps create liyali-admin-console`)
- [ ] **Set environment secrets** (API_URL, NEXTAUTH_SECRET, NEXTAUTH_URL)
- [ ] **Update backend CORS** (include admin console URL)
- [ ] **Deploy application** (push to GitHub or use flyctl)
- [ ] **Monitor deployment** (GitHub Actions or flyctl logs)

### After Deployment

- [ ] **Verify health check** (`curl https://liyali-admin-console.fly.dev/`)
- [ ] **Test login** (open in browser)
- [ ] **Check API connectivity** (no CORS errors)
- [ ] **Verify admin features** (dashboard, organizations, etc.)
- [ ] **Monitor logs** (`flyctl logs --app liyali-admin-console`)

---

## 🧪 Testing

### Local Testing

```bash
# Test the built app locally
cd admin-console
PORT=3001 node .next/standalone/server.js

# Or use npm
npm start

# Access at: http://localhost:3001
```

### Docker Testing

```bash
# Build Docker image
cd admin-console
docker build -t liyali-admin-console .

# Run container
docker run -p 3001:3001 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 \
  liyali-admin-console

# Access at: http://localhost:3001
```

---

## 📊 Deployment URLs

After deployment, your applications will be available at:

| Application       | URL                                      | Status              |
| ----------------- | ---------------------------------------- | ------------------- |
| Backend API       | https://liyali-gateway-api.fly.dev       | Should be running   |
| Frontend          | https://liyali-gateway-frontend.fly.dev  | Should be running   |
| **Admin Console** | **https://liyali-admin-console.fly.dev** | **Ready to deploy** |

---

## 🔐 Required Secrets

### Admin Console Secrets

```bash
NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1"
NEXTAUTH_SECRET="<generate-with-openssl-rand-base64-32>"
NEXTAUTH_URL="https://liyali-admin-console.fly.dev"
```

### Backend CORS Update

```bash
CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev"
```

---

## 🎯 Quick Deploy Commands

### Full Setup (First Time)

```bash
# 1. Create app
cd admin-console
flyctl apps create liyali-admin-console

# 2. Set secrets
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  NEXTAUTH_SECRET="$(openssl rand -base64 32)" \
  NEXTAUTH_URL="https://liyali-admin-console.fly.dev" \
  --app liyali-admin-console

# 3. Update backend CORS
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api

# 4. Deploy
flyctl deploy --remote-only

# 5. Verify
flyctl status --app liyali-admin-console
curl https://liyali-admin-console.fly.dev/
```

### Update Deployment (After First Time)

```bash
# Just push to GitHub
git add .
git commit -m "Update admin console"
git push origin develop

# Or deploy manually
cd admin-console
flyctl deploy --remote-only
```

---

## 🔍 Verification Commands

```bash
# Check app status
flyctl status --app liyali-admin-console

# Test health endpoint
curl https://liyali-admin-console.fly.dev/

# View logs
flyctl logs --app liyali-admin-console -f

# Check secrets
flyctl secrets list --app liyali-admin-console

# View metrics
flyctl metrics --app liyali-admin-console
```

---

## ⚠️ Important Notes

### CORS Configuration

**Critical**: Backend CORS must include admin console URL, otherwise API requests will fail.

```bash
# Verify CORS includes admin console
flyctl secrets list --app liyali-gateway-api | grep CORS

# Should show:
# CORS_ALLOWED_ORIGINS = https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev
```

### First Deployment

The first deployment will take longer (~5-8 minutes) because:

- Docker image needs to be built
- Dependencies need to be installed
- Application needs to start

Subsequent deployments will be faster (~3-5 minutes).

### Auto-Scaling

The app is configured to auto-scale:

- **Min instances**: 0 (stops when idle)
- **Max instances**: Unlimited (scales based on traffic)
- **Memory**: 512MB per instance

---

## 🎉 Success Criteria

Deployment is successful when:

- ✅ App status shows "running"
- ✅ Health check returns 200 OK
- ✅ Can access https://liyali-admin-console.fly.dev
- ✅ Can log in successfully
- ✅ No CORS errors in browser console
- ✅ API requests work correctly
- ✅ All admin features are accessible

---

## 🚨 Troubleshooting

### If deployment fails:

1. Check logs: `flyctl logs --app liyali-admin-console`
2. Verify secrets: `flyctl secrets list --app liyali-admin-console`
3. Check app status: `flyctl status --app liyali-admin-console`
4. Review: [ADMIN_CONSOLE_TROUBLESHOOTING.md](../ADMIN_CONSOLE_TROUBLESHOOTING.md)

### If CORS errors occur:

1. Update backend CORS to include admin console URL
2. Restart backend: `flyctl restart --app liyali-gateway-api`
3. Clear browser cache and retry

### If app won't start:

1. Check if secrets are set
2. Verify Dockerfile is correct
3. Test build locally first
4. Check memory limits

---

## 📚 Related Documentation

- **Quick Deploy**: [../QUICK_DEPLOY_ADMIN_CONSOLE.md](../QUICK_DEPLOY_ADMIN_CONSOLE.md)
- **Full Setup Guide**: [../ADMIN_CONSOLE_DEPLOYMENT_SETUP.md](../ADMIN_CONSOLE_DEPLOYMENT_SETUP.md)
- **Troubleshooting**: [../ADMIN_CONSOLE_TROUBLESHOOTING.md](../ADMIN_CONSOLE_TROUBLESHOOTING.md)
- **Documentation Index**: [../ADMIN_CONSOLE_DEPLOYMENT_INDEX.md](../ADMIN_CONSOLE_DEPLOYMENT_INDEX.md)
- **Fly.io Guide**: [./FLY_IO_DEPLOYMENT_GUIDE.md](./FLY_IO_DEPLOYMENT_GUIDE.md)

---

## 📊 Performance Metrics

### Build Performance

- **Workers Used**: 19 parallel workers
- **Compilation**: Fast (Turbopack)
- **TypeScript**: Incremental checking
- **Optimization**: Automatic

### Runtime Performance

- **Server**: Node.js standalone
- **Rendering**: Server-side + Static
- **Assets**: Optimized and minified
- **Caching**: Built-in Next.js caching

---

## ✨ Features Enabled

### Next.js Features

- ✅ App Router
- ✅ Server Components
- ✅ API Routes (via proxy)
- ✅ Middleware
- ✅ Static Generation
- ✅ Dynamic Rendering
- ✅ Image Optimization
- ✅ Font Optimization

### Production Optimizations

- ✅ Code splitting
- ✅ Tree shaking
- ✅ Minification
- ✅ Compression
- ✅ Asset optimization
- ✅ Bundle analysis

---

## 🎯 Summary

**Build Status**: ✅ Complete
**Configuration**: ✅ Ready
**Documentation**: ✅ Complete
**Deployment**: ⏳ Ready to deploy

The admin console has been successfully built and is ready for deployment to Fly.io. All configuration files, documentation, and deployment scripts are in place.

**Next Step**: Choose a deployment option above and deploy!

---

**Build Date**: February 8, 2026
**Build Time**: 71 seconds
**Routes**: 20 pages
**Status**: ✅ Production Ready
**Deployment**: Ready to go! 🚀
