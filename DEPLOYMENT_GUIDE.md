# Liyali Gateway - Deployment Guide

**Last Updated**: February 23, 2026

Complete guide for deploying all Liyali Gateway applications.

---

## Overview

Liyali Gateway consists of 3 applications:

1. **Backend** - Go/Fiber API server
2. **Web Frontend** - Next.js user-facing application
3. **Admin Console** - Next.js admin dashboard

---

## Quick Deployment

### Using Makefile (Recommended)

```bash
# Deploy all apps at once
make deploy

# Or deploy individually
make deploy-backend
make deploy-web
make deploy-admin
```

---

## Pre-Deployment Checklist

### 1. Verify Environment Files

```bash
make check-env
```

Ensure these files exist and are configured:

- `backend/.env`
- `frontend/.env`
- `admin-console/.env`

### 2. Run Pre-Deployment Checks

```bash
make pre-deploy
```

This command will:

- ✅ Check environment files
- ✅ Build all apps
- ✅ Run tests
- ✅ Run database migrations

---

## Detailed Deployment Steps

### Step 1: Database Migration

```bash
make migrate
```

Or manually:

```bash
cd backend
export DATABASE_URL="postgres://..."
go run cmd/migrate/main.go
```

**Verify migrations**:

```sql
SELECT filename FROM schema_migrations ORDER BY filename;
```

### Step 2: Build Applications

```bash
# Build all
make build

# Or individually
make build-backend
make build-web
make build-admin
```

**Verify builds**:

- Backend: `backend/liyali-backend` binary exists
- Frontend: `frontend/.next/` directory exists
- Admin Console: `admin-console/.next/` directory exists

### Step 3: Run Tests

```bash
# Test all
make test

# Or individually
make test-backend
make test-web
```

### Step 4: Deploy

#### Option A: Deploy All (Recommended)

```bash
make deploy
```

#### Option B: Deploy Individually

```bash
# Deploy backend
make deploy-backend

# Deploy web frontend
make deploy-web

# Deploy admin console
make deploy-admin
```

#### Option C: Manual Fly.io Deployment

```bash
cd backend && fly deploy
cd frontend && fly deploy
cd admin-console && fly deploy
```

---

## Environment Variables

### Backend (.env)

```env
# Database
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require

# Application
APP_PORT=8081
APP_ENV=production
JWT_SECRET=your-secret-key-min-32-chars

# CORS
FRONTEND_URL=https://your-frontend.com,https://admin.your-domain.com
```

### Frontend (.env)

```env
# API
NEXT_PUBLIC_API_URL=https://your-backend.com

# ImageKit
NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY=your_public_key
IMAGEKIT_PRIVATE_KEY=your_private_key
NEXT_PUBLIC_IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/your_id
```

### Admin Console (.env)

```env
# API
NEXT_PUBLIC_API_URL=https://your-backend.com

# Optional: ImageKit (if using logo upload)
NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY=your_public_key
IMAGEKIT_PRIVATE_KEY=your_private_key
NEXT_PUBLIC_IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/your_id
```

---

## Deployment Platforms

### Fly.io (Current Platform)

#### Initial Setup

1. **Install Fly CLI**:

   ```bash
   curl -L https://fly.io/install.sh | sh
   ```

2. **Login**:

   ```bash
   fly auth login
   ```

3. **Create Apps** (if not already created):
   ```bash
   cd backend && fly launch
   cd frontend && fly launch
   cd admin-console && fly launch
   ```

#### Deploy

```bash
make deploy
```

Or individually:

```bash
cd backend && fly deploy
cd frontend && fly deploy
cd admin-console && fly deploy
```

#### View Logs

```bash
cd backend && fly logs
cd frontend && fly logs
cd admin-console && fly logs
```

#### SSH Access

```bash
cd backend && fly ssh console
cd frontend && fly ssh console
cd admin-console && fly ssh console
```

### Other Platforms

#### Docker

Each app has a `Dockerfile`:

```bash
# Backend
cd backend
docker build -t liyali-backend .
docker run -p 8081:8081 --env-file .env liyali-backend

# Frontend
cd frontend
docker build -t liyali-frontend .
docker run -p 3000:3000 --env-file .env liyali-frontend

# Admin Console
cd admin-console
docker build -t liyali-admin .
docker run -p 3001:3000 --env-file .env liyali-admin
```

#### Manual Server Deployment

1. **Build**:

   ```bash
   make build
   ```

2. **Transfer Files**:

   ```bash
   # Backend
   scp backend/liyali-backend user@server:/opt/liyali/backend/

   # Frontend
   rsync -avz frontend/.next/ user@server:/opt/liyali/frontend/.next/
   rsync -avz frontend/public/ user@server:/opt/liyali/frontend/public/

   # Admin Console
   rsync -avz admin-console/.next/ user@server:/opt/liyali/admin/.next/
   rsync -avz admin-console/public/ user@server:/opt/liyali/admin/public/
   ```

3. **Set Environment Variables** on server

4. **Restart Services**:
   ```bash
   systemctl restart liyali-backend
   systemctl restart liyali-frontend
   systemctl restart liyali-admin
   ```

---

## Post-Deployment Verification

### 1. Check Application Health

```bash
# Backend
curl https://your-backend.com/health

# Frontend
curl https://your-frontend.com

# Admin Console
curl https://admin.your-domain.com
```

### 2. Verify Database Connection

```bash
# Check backend logs for database connection
cd backend && fly logs | grep "database"
```

### 3. Test Authentication

1. Navigate to frontend: `https://your-frontend.com/login`
2. Login with test credentials
3. Verify JWT token is issued
4. Check protected routes work

### 4. Test Admin Console

1. Navigate to admin console: `https://admin.your-domain.com/login`
2. Login with admin credentials
3. Verify admin features work
4. Check reports load correctly

### 5. Monitor Logs

```bash
# Backend
cd backend && fly logs --tail

# Frontend
cd frontend && fly logs --tail

# Admin Console
cd admin-console && fly logs --tail
```

---

## Rollback Procedure

If deployment fails or issues occur:

### Quick Rollback

```bash
# Rollback to previous version
cd backend && fly releases rollback
cd frontend && fly releases rollback
cd admin-console && fly releases rollback
```

### Manual Rollback

1. **Checkout previous commit**:

   ```bash
   git log --oneline -5
   git checkout <previous-commit-hash>
   ```

2. **Rebuild and redeploy**:

   ```bash
   make build
   make deploy
   ```

3. **Rollback database** (if needed):
   ```bash
   cd backend
   psql $DATABASE_URL -f database/migrations/XXX_migration.down.sql
   ```

---

## Troubleshooting

### Backend Issues

**Issue**: Backend won't start

```bash
# Check logs
cd backend && fly logs

# Common fixes:
- Verify DATABASE_URL is correct
- Check JWT_SECRET is set
- Ensure migrations ran successfully
```

**Issue**: Database connection fails

```bash
# Test connection
psql $DATABASE_URL -c "SELECT 1"

# Check firewall rules
# Verify SSL mode (require/disable)
```

### Frontend Issues

**Issue**: Build fails

```bash
# Check TypeScript errors
cd frontend && npm run build

# Clear cache and rebuild
rm -rf .next node_modules
npm install
npm run build
```

**Issue**: API calls fail

```bash
# Verify NEXT_PUBLIC_API_URL
echo $NEXT_PUBLIC_API_URL

# Check CORS settings in backend
# Verify authentication token
```

### Admin Console Issues

**Issue**: Can't access admin features

```bash
# Verify user has admin role
psql $DATABASE_URL -c "SELECT id, email, role FROM users WHERE role = 'admin';"

# Check JWT token claims
# Verify admin middleware is applied
```

---

## Monitoring

### Application Metrics

```bash
# Fly.io metrics
fly dashboard

# View resource usage
cd backend && fly status
cd frontend && fly status
cd admin-console && fly status
```

### Database Monitoring

```bash
# Connection count
psql $DATABASE_URL -c "SELECT count(*) FROM pg_stat_activity;"

# Slow queries
psql $DATABASE_URL -c "SELECT query, calls, total_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"

# Database size
psql $DATABASE_URL -c "SELECT pg_size_pretty(pg_database_size('postgres'));"
```

### Log Aggregation

```bash
# Tail all logs
make logs  # If you add this to Makefile

# Or individually
cd backend && fly logs --tail &
cd frontend && fly logs --tail &
cd admin-console && fly logs --tail &
```

---

## Scaling

### Horizontal Scaling (Fly.io)

```bash
# Scale backend
cd backend && fly scale count 2

# Scale frontend
cd frontend && fly scale count 2

# Scale admin console
cd admin-console && fly scale count 2
```

### Vertical Scaling (Fly.io)

```bash
# Increase resources
cd backend && fly scale vm shared-cpu-2x

# Check current scale
cd backend && fly status
```

---

## Maintenance

### Database Backups

```bash
# Manual backup
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore from backup
psql $DATABASE_URL < backup_file.sql
```

### Update Dependencies

```bash
# Backend
cd backend && go get -u ./...
cd backend && go mod tidy

# Frontend
cd frontend && npm update
cd frontend && npm audit fix

# Admin Console
cd admin-console && npm update
cd admin-console && npm audit fix
```

---

## Security Checklist

- [ ] All environment variables are set correctly
- [ ] JWT_SECRET is strong (min 32 characters)
- [ ] DATABASE_URL uses SSL (sslmode=require)
- [ ] CORS is configured correctly (FRONTEND_URL)
- [ ] Admin routes require admin role
- [ ] All queries filter by organization_id
- [ ] No sensitive data in logs
- [ ] HTTPS is enforced
- [ ] Rate limiting is enabled
- [ ] Database backups are automated

---

## Performance Checklist

- [ ] Database indexes are created (migration 014)
- [ ] React Query caching is enabled (5-min stale time)
- [ ] Images are optimized (Next.js Image component)
- [ ] API responses are compressed
- [ ] Static assets are cached
- [ ] Database connection pooling is configured
- [ ] Slow queries are monitored
- [ ] CDN is used for static assets (ImageKit)

---

## Support

### Getting Help

1. Check logs: `make logs` or `fly logs`
2. Review documentation: `DEVELOPER_GUIDE.md`, `QUICK_REFERENCE.md`
3. Check `.kiro/specs/` for feature details
4. Review existing code for examples

### Common Commands

```bash
make help          # Show all available commands
make deploy        # Deploy all apps
make build         # Build all apps
make test          # Run all tests
make migrate       # Run migrations
make clean         # Clean artifacts
make pre-deploy    # Pre-deployment checks
```

---

**Last Updated**: February 23, 2026  
**Version**: 1.0
