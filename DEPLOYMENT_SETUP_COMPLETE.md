# ✅ Deployment Setup Complete

Your Liyali Gateway project is now configured for automated deployment to Google Cloud Run using GitHub Actions and GitHub Container Registry (GHCR).

## 📦 What's Been Configured

### GitHub Actions Workflows

✅ **Backend Deployment** (`.github/workflows/backend-deploy.yml`)

- Triggers on changes to `backend/**`
- Builds Go application with multi-stage Docker
- Pushes to GHCR → GCP Artifact Registry → Cloud Run
- Supports `main` (production) and `develop` (staging) branches

✅ **Frontend Deployment** (`.github/workflows/frontend-deploy.yml`)

- Triggers on changes to `frontend/**`
- Builds Next.js with standalone output
- Optimized with layer caching and npm cache mounts
- Pushes to GHCR → GCP Artifact Registry → Cloud Run
- Supports `main` (production) and `develop` (staging) branches

### Docker Configuration

✅ **Backend Dockerfile** (`backend/Dockerfile`)

- Multi-stage build for minimal image size
- Alpine Linux base (~50MB final image)
- Non-root user for security
- Health check endpoint
- Optimized Go build with stripped binaries

✅ **Frontend Dockerfile** (`frontend/Dockerfile`)

- Multi-stage build with deps caching
- Next.js standalone output
- OpenSSL for Prisma support
- Non-root user for security
- Health check endpoint
- Build-time argument support

✅ **Docker Ignore Files**

- `backend/.dockerignore` - Excludes test files, docs, IDE configs
- `frontend/.dockerignore` - Excludes node_modules, .next, env files

### Next.js Configuration

✅ **Standalone Output** (`frontend/next.config.ts`)

- Enabled `output: "standalone"` for Docker deployment
- Added Cloud Run image domain support
- Optimized for production builds

✅ **Health Check Endpoint** (`frontend/src/app/api/health/route.ts`)

- Simple health check for Docker/Cloud Run
- Returns JSON with status and timestamp

### Documentation

✅ **Comprehensive Guides**

- `DEPLOYMENT_GUIDE.md` - Complete step-by-step deployment guide (100+ pages)
- `DEPLOYMENT_CHECKLIST.md` - Quick reference checklist
- `ENVIRONMENT_VARIABLES.md` - All environment variables documented
- `.github/WORKFLOWS_README.md` - GitHub Actions workflows explained

## 🚀 Next Steps

### 1. Set Up Google Cloud Platform (15 minutes)

```bash
# Install gcloud CLI
# macOS: brew install --cask google-cloud-sdk
# Windows: Download from https://cloud.google.com/sdk/docs/install

# Login and set project
gcloud auth login
gcloud config set project YOUR_PROJECT_ID

# Enable required APIs
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com

# Create Artifact Registry repository
gcloud artifacts repositories create liyali \
  --repository-format=docker \
  --location=us-central1 \
  --description="Liyali Gateway containers"

# Create service account
gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions Deployment"

# Grant permissions
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:github-actions@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:github-actions@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:github-actions@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Create and download key
gcloud iam service-accounts keys create gcp-key.json \
  --iam-account=github-actions@YOUR_PROJECT_ID.iam.gserviceaccount.com
```

### 2. Set Up Database on Prisma.io (5 minutes)

1. Go to https://cloud.prisma.io
2. Sign up/Login
3. Create new project
4. Select PostgreSQL
5. Choose region (close to your Cloud Run region)
6. Copy connection string

### 3. Configure GitHub Secrets (10 minutes)

Go to your repository → Settings → Secrets and variables → Actions

Add these secrets:

**Google Cloud:**

```
GCP_SA_KEY          = <entire content of gcp-key.json>
GCP_PROJECT_ID      = your-project-id
GCP_REGION          = us-central1
```

**Services:**

```
BACKEND_SERVICE_NAME  = liyali-backend
FRONTEND_SERVICE_NAME = liyali-frontend
```

**Database:**

```
DATABASE_URL = postgresql://user:pass@host:5432/db?sslmode=require
```

**Backend:**

```
JWT_SECRET           = <generate: openssl rand -base64 32>
CORS_ALLOWED_ORIGINS = https://your-frontend-url.run.app
```

**Frontend:**

```
NEXT_PUBLIC_API_URL = https://your-backend-url.run.app
NEXTAUTH_SECRET     = <generate: openssl rand -base64 32>
NEXTAUTH_URL        = https://your-frontend-url.run.app
```

### 4. First Deployment (20 minutes)

#### Deploy Backend First:

```bash
# Commit and push backend
git add backend/
git commit -m "feat: initial backend deployment"
git push origin main

# Wait for deployment (check Actions tab)
# Copy backend URL from Cloud Run console
```

#### Run Database Migrations:

```bash
cd backend
export DATABASE_URL="your-prisma-connection-string"
make db-migrate
```

#### Update Secrets:

```bash
# Update these GitHub Secrets with actual URLs:
# - NEXT_PUBLIC_API_URL (backend URL)
# - CORS_ALLOWED_ORIGINS (will add frontend URL after next step)
```

#### Deploy Frontend:

```bash
# Commit and push frontend
git add frontend/
git commit -m "feat: initial frontend deployment"
git push origin main

# Wait for deployment (check Actions tab)
# Copy frontend URL from Cloud Run console
```

#### Final Updates:

```bash
# Update these GitHub Secrets:
# - NEXTAUTH_URL (frontend URL)
# - CORS_ALLOWED_ORIGINS (add frontend URL)

# Trigger backend redeploy to pick up new CORS settings
git commit --allow-empty -m "chore: update CORS"
git push origin main
```

### 5. Verify Deployment (5 minutes)

```bash
# Test backend
curl https://your-backend-url.run.app/health

# Test frontend
curl https://your-frontend-url.run.app/api/health

# Open frontend in browser
# Test login and basic functionality
```

## 📚 Documentation Reference

| Document                      | Purpose                   | When to Use                        |
| ----------------------------- | ------------------------- | ---------------------------------- |
| `DEPLOYMENT_GUIDE.md`         | Complete deployment guide | First-time setup, troubleshooting  |
| `DEPLOYMENT_CHECKLIST.md`     | Quick reference checklist | Every deployment, verification     |
| `ENVIRONMENT_VARIABLES.md`    | All env vars documented   | Setting up secrets, debugging      |
| `.github/WORKFLOWS_README.md` | GitHub Actions explained  | Understanding CI/CD, customization |

## 🔧 Common Commands

### View Logs

```bash
# Backend logs
gcloud run services logs read liyali-backend --region=us-central1 --limit=50

# Frontend logs
gcloud run services logs read liyali-frontend --region=us-central1 --limit=50

# Follow logs in real-time
gcloud run services logs tail liyali-backend --region=us-central1
```

### Manual Deployment

```bash
# Trigger backend deployment
git add backend/
git commit -m "feat: update backend"
git push origin main

# Trigger frontend deployment
git add frontend/
git commit -m "feat: update frontend"
git push origin main
```

### Rollback

```bash
# List revisions
gcloud run revisions list --service=liyali-backend --region=us-central1

# Rollback to previous
gcloud run services update-traffic liyali-backend \
  --region=us-central1 \
  --to-revisions=REVISION_NAME=100
```

## 🎯 Deployment Workflow

```
Developer → Git Push → GitHub Actions → GHCR → GCP → Cloud Run → Live!
```

**Automatic triggers:**

- Push to `main` → Production deployment
- Push to `develop` → Staging deployment
- Changes in `backend/**` → Backend deployment only
- Changes in `frontend/**` → Frontend deployment only

## ✅ Pre-Deployment Checklist

Before your first deployment, ensure:

```
□ GCP project created and billing enabled
□ Required APIs enabled (Cloud Run, Artifact Registry, Cloud Build)
□ Artifact Registry repository created
□ Service account created with proper permissions
□ Service account key downloaded (gcp-key.json)
□ Prisma.io database created
□ Database connection string obtained
□ All GitHub Secrets configured
□ GitHub Actions enabled in repository
□ Workflow permissions set to "Read and write"
```

## 🔐 Security Checklist

```
□ Never commit secrets to Git
□ Service account has minimal required permissions
□ Database uses SSL/TLS connections
□ JWT_SECRET is strong and random (32+ characters)
□ NEXTAUTH_SECRET is strong and random (32+ characters)
□ CORS_ALLOWED_ORIGINS is properly configured
□ Containers run as non-root users
□ Health checks are configured
□ Secrets are stored in GitHub Secrets only
```

## 💰 Cost Estimate

**Development/Staging:**

- Cloud Run: $5-20/month (with min-instances=0)
- Prisma.io: Free tier or $25/month
- **Total: ~$5-45/month**

**Production (moderate traffic):**

- Cloud Run: $50-200/month
- Prisma.io: $25-100/month
- **Total: ~$75-300/month**

## 🆘 Getting Help

1. **Check Documentation:**

   - Read `DEPLOYMENT_GUIDE.md` for detailed instructions
   - Check `DEPLOYMENT_CHECKLIST.md` for quick reference
   - Review workflow logs in GitHub Actions tab

2. **Common Issues:**

   - Build fails → Check Dockerfile syntax and dependencies
   - Deployment fails → Verify GCP permissions and secrets
   - Database connection fails → Check DATABASE_URL format
   - CORS errors → Update CORS_ALLOWED_ORIGINS

3. **Support Resources:**
   - [Google Cloud Run Docs](https://cloud.google.com/run/docs)
   - [GitHub Actions Docs](https://docs.github.com/en/actions)
   - [Prisma Docs](https://www.prisma.io/docs)

## 🎉 You're Ready!

Your deployment infrastructure is fully configured. Follow the "Next Steps" above to complete your first deployment.

**Estimated time to first deployment: ~1 hour**

---

**Questions?** Check the documentation files or open an issue in the repository.

**Last Updated:** January 2026
