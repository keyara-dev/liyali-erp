# Fly.io Deployment Guide - Quick Demo Setup

Deploy Liyali Gateway to Fly.io for quick demos and testing before migrating to GCP for production.

## 🎯 Why Fly.io for Demo?

- ✅ **Free tier**: 3 VMs with 256MB RAM each
- ✅ **No cold starts**: Always-on instances
- ✅ **Global edge network**: Fast worldwide
- ✅ **Simple deployment**: One command deploy
- ✅ **Built-in PostgreSQL**: Free tier available
- ✅ **Easy migration**: Same Docker images work on GCP

## 📋 Quick Start Checklist

```
□ Install Fly.io CLI
□ Create Fly.io account
□ Create PostgreSQL database
□ Set secrets
□ Deploy backend
□ Deploy frontend
□ Test application
```

---

## 1️⃣ Install Fly.io CLI

### macOS/Linux

```bash
curl -L https://fly.io/install.sh | sh
```

### Windows (PowerShell)

```powershell
pwsh -Command "iwr https://fly.io/install.ps1 -useb | iex"
```

### Verify Installation

```bash
flyctl version
```

---

## 2️⃣ Create Fly.io Account

```bash
# Sign up and login
flyctl auth signup

# Or login if you have an account
flyctl auth login
```

**Note**: You'll need to add a credit card, but the free tier is sufficient for demos.

---

## 3️⃣ Create PostgreSQL Database

```bash
# Create a PostgreSQL cluster (free tier)
flyctl postgres create \
  --name liyali-db \
  --region iad \
  --initial-cluster-size 1 \
  --vm-size shared-cpu-1x \
  --volume-size 1

# Save the connection string shown in the output
# Format: postgres://user:password@host:5432/database
```

**Important**: Copy and save the connection string - you'll need it for secrets.

---

## 4️⃣ Deploy Backend

### Initialize Backend App

```bash
cd backend

# Create the app (this reads fly.toml)
flyctl apps create liyali-backend

# Or let Fly.io generate a name
flyctl launch --no-deploy
```

### Set Backend Secrets

```bash
# Set database URL
flyctl secrets set DATABASE_URL="postgresql://user:pass@host:5432/db?sslmode=require"

# Set JWT secret (generate with: openssl rand -base64 32)
flyctl secrets set JWT_SECRET="your-jwt-secret-here"

# Set CORS (will update after frontend deployment)
flyctl secrets set CORS_ALLOWED_ORIGINS="*"

# Set environment
flyctl secrets set ENV="production"
```

### Deploy Backend

```bash
# Deploy
flyctl deploy

# Check status
flyctl status

# View logs
flyctl logs

# Get URL
flyctl info
```

**Save the backend URL**: `https://liyali-backend.fly.dev`

---

## 5️⃣ Run Database Migrations

```bash
# From backend directory
export DATABASE_URL="your-fly-postgres-connection-string"

# Run migrations
make db-migrate

# Or manually
psql $DATABASE_URL < database/migrations/001_init_system.up.sql
psql $DATABASE_URL < database/migrations/002_seed_data.up.sql
```

---

## 6️⃣ Deploy Frontend

### Initialize Frontend App

```bash
cd frontend

# Create the app
flyctl apps create liyali-frontend

# Or let Fly.io generate a name
flyctl launch --no-deploy
```

### Set Frontend Secrets

```bash
# Set backend API URL (use the URL from step 4)
flyctl secrets set NEXT_PUBLIC_API_URL="https://liyali-backend.fly.dev"

# Set NextAuth secret (generate with: openssl rand -base64 32)
flyctl secrets set NEXTAUTH_SECRET="your-nextauth-secret-here"

# Set NextAuth URL (will update after deployment)
flyctl secrets set NEXTAUTH_URL="https://liyali-frontend.fly.dev"

# Set database URL (same as backend)
flyctl secrets set DATABASE_URL="postgresql://user:pass@host:5432/db?sslmode=require"

# Set environment
flyctl secrets set NODE_ENV="production"
```

### Deploy Frontend

```bash
# Deploy
flyctl deploy

# Check status
flyctl status

# View logs
flyctl logs

# Get URL
flyctl info
```

**Save the frontend URL**: `https://liyali-frontend.fly.dev`

---

## 7️⃣ Update CORS Settings

Now that you have the frontend URL, update backend CORS:

```bash
cd backend

# Update CORS to include frontend URL
flyctl secrets set CORS_ALLOWED_ORIGINS="https://liyali-frontend.fly.dev"

# Restart to apply changes
flyctl apps restart liyali-backend
```

---

## 8️⃣ Verify Deployment

### Test Backend

```bash
# Health check
curl https://liyali-backend.fly.dev/health

# Should return: {"status":"ok"}
```

### Test Frontend

```bash
# Health check
curl https://liyali-frontend.fly.dev/api/health

# Open in browser
open https://liyali-frontend.fly.dev
```

### Test Full Application

1. Open frontend URL in browser
2. Try to login
3. Create a test requisition
4. Verify workflow functionality

---

## 🔄 GitHub Actions Auto-Deploy

### Setup

1. Get your Fly.io API token:

   ```bash
   flyctl auth token
   ```

2. Add to GitHub Secrets:

   - Go to repository → Settings → Secrets → Actions
   - Add secret: `FLY_API_TOKEN` = your token

3. Commit with deployment tags:

   ```bash
   # Deploy backend only
   git commit -m "feat: update backend [backend]"

   # Deploy frontend only
   git commit -m "feat: update UI [frontend]"

   # Deploy both
   git commit -m "feat: full update [all]"

   git push origin develop
   ```

### Manual Trigger

1. Go to GitHub → Actions
2. Select "Deploy to Fly.io"
3. Click "Run workflow"
4. Select branch
5. Click "Run workflow"

---

## 📊 Monitoring

### View Logs

```bash
# Backend logs
cd backend
flyctl logs

# Frontend logs
cd frontend
flyctl logs

# Follow logs in real-time
flyctl logs -a liyali-backend
```

### Check Status

```bash
# App status
flyctl status -a liyali-backend

# Machine status
flyctl machine list -a liyali-backend

# Database status
flyctl postgres db list -a liyali-db
```

### View Metrics

```bash
# Open dashboard
flyctl dashboard

# Or visit: https://fly.io/dashboard
```

---

## 🔧 Common Commands

### Deployment

```bash
# Deploy with build logs
flyctl deploy --verbose

# Deploy without cache
flyctl deploy --no-cache

# Deploy specific Dockerfile
flyctl deploy --dockerfile Dockerfile.prod
```

### Scaling

```bash
# Scale to 2 machines
flyctl scale count 2

# Scale memory
flyctl scale memory 512

# Scale CPU
flyctl scale vm shared-cpu-2x
```

### Secrets Management

```bash
# List secrets
flyctl secrets list

# Set secret
flyctl secrets set KEY=VALUE

# Remove secret
flyctl secrets unset KEY

# Import from file
flyctl secrets import < .env.production
```

### Database

```bash
# Connect to database
flyctl postgres connect -a liyali-db

# Create database backup
flyctl postgres backup create -a liyali-db

# List backups
flyctl postgres backup list -a liyali-db
```

---

## 🐛 Troubleshooting

### Issue: App won't start

**Check logs:**

```bash
flyctl logs -a liyali-backend
```

**Common fixes:**

- Verify all secrets are set: `flyctl secrets list`
- Check DATABASE_URL format
- Ensure port 8080 (backend) or 3000 (frontend) is exposed
- Verify Dockerfile builds locally: `docker build -t test .`

### Issue: Database connection fails

**Test connection:**

```bash
flyctl postgres connect -a liyali-db
```

**Common fixes:**

- Verify DATABASE_URL includes `?sslmode=require`
- Check database is running: `flyctl status -a liyali-db`
- Ensure app is in same region as database

### Issue: CORS errors

**Update CORS:**

```bash
cd backend
flyctl secrets set CORS_ALLOWED_ORIGINS="https://liyali-frontend.fly.dev"
flyctl apps restart liyali-backend
```

### Issue: Out of memory

**Scale up:**

```bash
flyctl scale memory 512 -a liyali-backend
```

---

## 💰 Cost Management

### Free Tier Limits

- **Compute**: 3 shared-cpu-1x VMs (256MB RAM each)
- **Storage**: 3GB total
- **Bandwidth**: 160GB outbound/month
- **PostgreSQL**: 1 cluster (shared-cpu-1x, 1GB storage)

### Stay Within Free Tier

```bash
# Use minimal resources
flyctl scale count 1
flyctl scale memory 256
flyctl scale vm shared-cpu-1x

# Enable auto-stop (already in fly.toml)
# Machines stop when idle, start on request
```

### Monitor Usage

```bash
# View current usage
flyctl dashboard

# Check billing
# Visit: https://fly.io/dashboard/personal/billing
```

---

## 🚀 Migration to GCP

When ready to move to production on GCP:

### 1. Export Configuration

```bash
# Export secrets
flyctl secrets list -a liyali-backend > backend-secrets.txt
flyctl secrets list -a liyali-frontend > frontend-secrets.txt
```

### 2. Backup Database

```bash
# Create backup
flyctl postgres backup create -a liyali-db

# Export data
pg_dump $FLY_DATABASE_URL > backup.sql
```

### 3. Deploy to GCP

```bash
# Push to main branch (triggers GCP deployment)
git checkout main
git merge develop
git push origin main
```

### 4. Migrate Database

```bash
# Import to GCP database
psql $GCP_DATABASE_URL < backup.sql
```

### 5. Update DNS

Point your domain to GCP Cloud Run URLs instead of Fly.io.

---

## 📚 Useful Resources

- [Fly.io Documentation](https://fly.io/docs/)
- [Fly.io Go Guide](https://fly.io/docs/languages-and-frameworks/golang/)
- [Fly.io PostgreSQL](https://fly.io/docs/postgres/)
- [Fly.io Pricing](https://fly.io/docs/about/pricing/)

---

## 🎯 Quick Reference

### Deploy Commands

```bash
# Backend
cd backend && flyctl deploy

# Frontend
cd frontend && flyctl deploy

# Both (from root)
cd backend && flyctl deploy && cd ../frontend && flyctl deploy
```

### URLs

```bash
# Get backend URL
cd backend && flyctl info | grep Hostname

# Get frontend URL
cd frontend && flyctl info | grep Hostname
```

### Logs

```bash
# Backend logs
flyctl logs -a liyali-backend

# Frontend logs
flyctl logs -a liyali-frontend
```

---

## ✅ Deployment Checklist

```
Setup:
□ Fly.io CLI installed
□ Account created and logged in
□ PostgreSQL database created
□ Connection string saved

Backend:
□ App created (liyali-backend)
□ Secrets set (DATABASE_URL, JWT_SECRET, CORS_ALLOWED_ORIGINS)
□ Deployed successfully
□ Health check passes
□ URL saved

Database:
□ Migrations run
□ Seed data inserted
□ Connection verified

Frontend:
□ App created (liyali-frontend)
□ Secrets set (NEXT_PUBLIC_API_URL, NEXTAUTH_SECRET, etc.)
□ Deployed successfully
□ Health check passes
□ URL saved

Final:
□ CORS updated with frontend URL
□ Backend restarted
□ Full application tested
□ GitHub Actions configured (optional)
```

---

**Estimated Setup Time**: 15-20 minutes

**Ready to deploy?** Start with step 1 and follow the checklist!

**Last Updated:** January 2026
