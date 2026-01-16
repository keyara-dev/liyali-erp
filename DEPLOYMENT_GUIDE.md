# Liyali Gateway - Deployment Guide

Complete step-by-step guide for deploying the Liyali Gateway application to Google Cloud Run using GitHub Actions and GitHub Container Registry (GHCR).

## 📋 Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Initial Setup Checklist](#initial-setup-checklist)
- [GitHub Secrets Configuration](#github-secrets-configuration)
- [Google Cloud Setup](#google-cloud-setup)
- [Database Setup (Prisma)](#database-setup-prisma)
- [Deployment Process](#deployment-process)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

---

## Overview

### Architecture

```
┌─────────────────┐
│   GitHub Repo   │
│   (Push Code)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ GitHub Actions  │
│  (CI/CD Build)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│      GHCR       │
│ (Container Reg) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  GCP Artifact   │
│    Registry     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Cloud Run      │
│  (Production)   │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Prisma.io DB   │
│  (PostgreSQL)   │
└─────────────────┘
```

### Deployment Triggers

- **Backend**: Deploys when changes are pushed to `backend/**` folder
- **Frontend**: Deploys when changes are pushed to `frontend/**` folder
- **Branches**: `main` (production) and `develop` (staging)

---

## Prerequisites

### Required Accounts

- [ ] GitHub account with repository access
- [ ] Google Cloud Platform account with billing enabled
- [ ] Prisma.io account (or any PostgreSQL database provider)

### Required Tools (for local setup)

- [ ] Git
- [ ] Docker Desktop
- [ ] gcloud CLI (Google Cloud SDK)
- [ ] Node.js 20+ (for local development)
- [ ] Go 1.21+ (for local development)

---

## Initial Setup Checklist

### Phase 1: Google Cloud Platform Setup

#### 1.1 Create GCP Project

- [ ] Go to [Google Cloud Console](https://console.cloud.google.com)
- [ ] Click "Select a project" → "New Project"
- [ ] Enter project name: `liyali-gateway` (or your preferred name)
- [ ] Note down the **Project ID** (e.g., `liyali-gateway-123456`)
- [ ] Enable billing for the project

#### 1.2 Enable Required APIs

```bash
# Set your project ID
export PROJECT_ID="your-project-id"
gcloud config set project $PROJECT_ID

# Enable required APIs
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  containerregistry.googleapis.com \
  iam.googleapis.com
```

- [ ] Cloud Run API enabled
- [ ] Artifact Registry API enabled
- [ ] Cloud Build API enabled
- [ ] Container Registry API enabled
- [ ] IAM API enabled

#### 1.3 Create Artifact Registry Repository

```bash
# Create repository for Docker images
gcloud artifacts repositories create liyali \
  --repository-format=docker \
  --location=us-central1 \
  --description="Liyali Gateway container images"
```

- [ ] Artifact Registry repository created
- [ ] Note down the region (e.g., `us-central1`)

#### 1.4 Create Service Account

```bash
# Create service account
gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions Deployment"

# Grant necessary permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Create and download key
gcloud iam service-accounts keys create gcp-key.json \
  --iam-account=github-actions@${PROJECT_ID}.iam.gserviceaccount.com
```

- [ ] Service account created
- [ ] Permissions granted
- [ ] Service account key downloaded (`gcp-key.json`)
- [ ] **IMPORTANT**: Keep this file secure and never commit it to Git

---

### Phase 2: Database Setup (Prisma.io)

#### 2.1 Create Prisma Database

- [ ] Go to [Prisma Data Platform](https://cloud.prisma.io)
- [ ] Sign up or log in
- [ ] Click "New Project"
- [ ] Select "PostgreSQL" as database type
- [ ] Choose a region close to your Cloud Run region
- [ ] Note down the **Connection String** (format: `postgresql://user:password@host:port/database`)

#### 2.2 Alternative: Use Your Own PostgreSQL

If not using Prisma.io:

- [ ] Set up PostgreSQL database (Cloud SQL, AWS RDS, etc.)
- [ ] Create database: `liyali_gateway`
- [ ] Create user with appropriate permissions
- [ ] Get connection string in format: `postgresql://user:password@host:port/database?sslmode=require`

#### 2.3 Test Database Connection

```bash
# Test connection (replace with your connection string)
psql "postgresql://user:password@host:port/database"
```

- [ ] Database connection successful
- [ ] Connection string saved securely

---

### Phase 3: GitHub Repository Setup

#### 3.1 Enable GitHub Container Registry

- [ ] Go to your GitHub repository
- [ ] Navigate to Settings → Actions → General
- [ ] Under "Workflow permissions", select:
  - [x] Read and write permissions
  - [x] Allow GitHub Actions to create and approve pull requests
- [ ] Click "Save"

#### 3.2 Configure GitHub Secrets

Go to your repository → Settings → Secrets and variables → Actions → New repository secret

Add the following secrets:

##### Google Cloud Secrets

- [ ] **GCP_SA_KEY**: Content of `gcp-key.json` file (entire JSON)

  ```
  Copy the entire content of gcp-key.json file
  ```

- [ ] **GCP_PROJECT_ID**: Your GCP project ID

  ```
  Example: liyali-gateway-123456
  ```

- [ ] **GCP_REGION**: Your GCP region
  ```
  Example: us-central1
  ```

##### Service Names

- [ ] **BACKEND_SERVICE_NAME**: Cloud Run service name for backend

  ```
  Example: liyali-backend
  ```

- [ ] **FRONTEND_SERVICE_NAME**: Cloud Run service name for frontend
  ```
  Example: liyali-frontend
  ```

##### Database Secrets

- [ ] **DATABASE_URL**: PostgreSQL connection string from Prisma.io
  ```
  Example: postgresql://user:password@host:port/database?sslmode=require
  ```

##### Backend Secrets

- [ ] **JWT_SECRET**: Random secure string for JWT signing

  ```bash
  # Generate with:
  openssl rand -base64 32
  ```

- [ ] **CORS_ALLOWED_ORIGINS**: Comma-separated list of allowed origins
  ```
  Example: https://liyali-frontend-xyz.run.app,https://yourdomain.com
  ```

##### Frontend Secrets

- [ ] **NEXT_PUBLIC_API_URL**: Backend API URL (will be updated after first backend deployment)

  ```
  Example: https://liyali-backend-xyz.run.app
  ```

- [ ] **NEXTAUTH_SECRET**: Random secure string for NextAuth

  ```bash
  # Generate with:
  openssl rand -base64 32
  ```

- [ ] **NEXTAUTH_URL**: Frontend URL (will be updated after first frontend deployment)
  ```
  Example: https://liyali-frontend-xyz.run.app
  ```

---

## GitHub Secrets Configuration

### Complete Secrets Checklist

Copy this checklist and mark each secret as you add it:

```
Google Cloud Platform:
□ GCP_SA_KEY
□ GCP_PROJECT_ID
□ GCP_REGION

Service Configuration:
□ BACKEND_SERVICE_NAME
□ FRONTEND_SERVICE_NAME

Database:
□ DATABASE_URL

Backend Application:
□ JWT_SECRET
□ CORS_ALLOWED_ORIGINS

Frontend Application:
□ NEXT_PUBLIC_API_URL
□ NEXTAUTH_SECRET
□ NEXTAUTH_URL
```

### How to Add Secrets

1. Go to your GitHub repository
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Enter the secret name (exactly as shown above)
5. Paste the secret value
6. Click **Add secret**
7. Repeat for all secrets

---

## Google Cloud Setup

### Detailed GCP Configuration

#### Step 1: Install gcloud CLI

**macOS:**

```bash
brew install --cask google-cloud-sdk
```

**Windows:**
Download from: https://cloud.google.com/sdk/docs/install

**Linux:**

```bash
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
```

#### Step 2: Initialize gcloud

```bash
# Login to Google Cloud
gcloud auth login

# Set your project
gcloud config set project YOUR_PROJECT_ID

# Verify configuration
gcloud config list
```

#### Step 3: Create Cloud Run Services (Optional - will be created automatically)

The services will be created automatically on first deployment, but you can pre-create them:

```bash
# Create backend service
gcloud run services create liyali-backend \
  --region=us-central1 \
  --platform=managed \
  --allow-unauthenticated

# Create frontend service
gcloud run services create liyali-frontend \
  --region=us-central1 \
  --platform=managed \
  --allow-unauthenticated
```

---

## Database Setup (Prisma)

### Option 1: Prisma.io (Recommended)

#### Step 1: Create Account

- [ ] Visit https://cloud.prisma.io
- [ ] Sign up with GitHub or email
- [ ] Verify your email

#### Step 2: Create Project

- [ ] Click "New Project"
- [ ] Name: `liyali-gateway`
- [ ] Select region: Choose closest to your Cloud Run region
- [ ] Click "Create Project"

#### Step 3: Get Connection String

- [ ] Go to project settings
- [ ] Copy the connection string
- [ ] Format: `postgresql://user:password@host:port/database?sslmode=require`
- [ ] Add to GitHub Secrets as `DATABASE_URL`

#### Step 4: Run Migrations (After First Deployment)

```bash
# From backend directory
cd backend

# Set DATABASE_URL
export DATABASE_URL="your-connection-string"

# Run migrations
make db-migrate

# Or manually:
psql $DATABASE_URL < database/migrations/001_init_system.up.sql
psql $DATABASE_URL < database/migrations/002_seed_data.up.sql
```

### Option 2: Google Cloud SQL

If you prefer Cloud SQL:

```bash
# Create Cloud SQL instance
gcloud sql instances create liyali-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=us-central1

# Create database
gcloud sql databases create liyali_gateway \
  --instance=liyali-db

# Create user
gcloud sql users create liyali_user \
  --instance=liyali-db \
  --password=YOUR_SECURE_PASSWORD

# Get connection string
gcloud sql instances describe liyali-db \
  --format="value(connectionName)"
```

---

## Deployment Process

### First-Time Deployment

#### Step 1: Deploy Backend First

1. **Commit and push backend changes:**

   ```bash
   git add backend/
   git commit -m "feat: initial backend deployment"
   git push origin main
   ```

2. **Monitor deployment:**

   - [ ] Go to GitHub repository → Actions tab
   - [ ] Watch "Backend - Build and Deploy to Cloud Run" workflow
   - [ ] Wait for all steps to complete (usually 5-10 minutes)

3. **Get backend URL:**

   - [ ] Go to [Google Cloud Console](https://console.cloud.google.com)
   - [ ] Navigate to Cloud Run
   - [ ] Click on `liyali-backend` service
   - [ ] Copy the service URL (e.g., `https://liyali-backend-xyz.run.app`)

4. **Update GitHub Secrets:**
   - [ ] Update `NEXT_PUBLIC_API_URL` with backend URL
   - [ ] Update `CORS_ALLOWED_ORIGINS` to include frontend URL (after frontend deployment)

#### Step 2: Run Database Migrations

```bash
# Set the backend URL
export BACKEND_URL="https://liyali-backend-xyz.run.app"

# Test health endpoint
curl $BACKEND_URL/health

# Run migrations (from local machine)
cd backend
export DATABASE_URL="your-prisma-connection-string"
make db-migrate
```

- [ ] Migrations completed successfully
- [ ] Database tables created
- [ ] Seed data inserted

#### Step 3: Deploy Frontend

1. **Update frontend environment variables:**

   - [ ] Ensure `NEXT_PUBLIC_API_URL` is set in GitHub Secrets
   - [ ] Ensure `NEXTAUTH_SECRET` is set
   - [ ] Ensure `DATABASE_URL` is set (for NextAuth)

2. **Commit and push frontend changes:**

   ```bash
   git add frontend/
   git commit -m "feat: initial frontend deployment"
   git push origin main
   ```

3. **Monitor deployment:**

   - [ ] Go to GitHub repository → Actions tab
   - [ ] Watch "Frontend - Build and Deploy to Cloud Run" workflow
   - [ ] Wait for completion (usually 5-10 minutes)

4. **Get frontend URL:**
   - [ ] Go to Cloud Run in GCP Console
   - [ ] Click on `liyali-frontend` service
   - [ ] Copy the service URL (e.g., `https://liyali-frontend-xyz.run.app`)

#### Step 4: Update CORS and NextAuth

1. **Update backend CORS:**

   - [ ] Go to GitHub Secrets
   - [ ] Update `CORS_ALLOWED_ORIGINS` to include frontend URL
   - [ ] Trigger backend redeployment (push a small change or redeploy manually)

2. **Update NextAuth URL:**
   - [ ] Update `NEXTAUTH_URL` with frontend URL
   - [ ] Trigger frontend redeployment

---

### Subsequent Deployments

After initial setup, deployments are automatic:

#### Backend Updates

```bash
# Make changes to backend code
git add backend/
git commit -m "feat: add new feature"
git push origin main
```

- [ ] GitHub Actions automatically builds and deploys
- [ ] No manual intervention needed

#### Frontend Updates

```bash
# Make changes to frontend code
git add frontend/
git commit -m "feat: update UI"
git push origin main
```

- [ ] GitHub Actions automatically builds and deploys
- [ ] No manual intervention needed

#### Both Backend and Frontend

```bash
# Make changes to both
git add .
git commit -m "feat: full-stack feature"
git push origin main
```

- [ ] Both workflows run in parallel
- [ ] Independent deployments

---

## Verification

### Post-Deployment Checklist

#### Backend Verification

```bash
# Set backend URL
export BACKEND_URL="https://liyali-backend-xyz.run.app"

# Test health endpoint
curl $BACKEND_URL/health
# Expected: {"status":"ok","timestamp":"..."}

# Test API endpoint
curl $BACKEND_URL/api/v1/health
# Expected: {"status":"healthy"}

# Check logs
gcloud run services logs read liyali-backend \
  --region=us-central1 \
  --limit=50
```

- [ ] Health endpoint returns 200 OK
- [ ] API responds correctly
- [ ] No errors in logs
- [ ] Database connection successful

#### Frontend Verification

```bash
# Set frontend URL
export FRONTEND_URL="https://liyali-frontend-xyz.run.app"

# Test homepage
curl -I $FRONTEND_URL
# Expected: HTTP/2 200

# Check logs
gcloud run services logs read liyali-frontend \
  --region=us-central1 \
  --limit=50
```

- [ ] Homepage loads successfully
- [ ] Login page accessible
- [ ] API calls working
- [ ] No console errors
- [ ] Authentication working

#### Integration Testing

- [ ] Open frontend URL in browser
- [ ] Test login functionality
- [ ] Create a test requisition
- [ ] Verify workflow approval process
- [ ] Check notifications
- [ ] Test search functionality

---

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: Build Fails in GitHub Actions

**Symptoms:**

- GitHub Actions workflow shows red X
- Build step fails

**Solutions:**

1. **Check Docker build:**

   ```bash
   # Test locally
   cd backend  # or frontend
   docker build -t test-build .
   ```

2. **Check logs:**

   - Go to Actions tab
   - Click on failed workflow
   - Expand failed step
   - Read error messages

3. **Common fixes:**
   - [ ] Verify Dockerfile syntax
   - [ ] Check for missing dependencies in package.json or go.mod
   - [ ] Ensure all files are committed

#### Issue 2: Deployment Fails

**Symptoms:**

- Build succeeds but deployment fails
- Cloud Run shows error

**Solutions:**

1. **Check service account permissions:**

   ```bash
   gcloud projects get-iam-policy $PROJECT_ID \
     --flatten="bindings[].members" \
     --filter="bindings.members:serviceAccount:github-actions@*"
   ```

2. **Verify secrets:**

   - [ ] All required secrets are set in GitHub
   - [ ] Secret values are correct (no extra spaces)
   - [ ] GCP_SA_KEY is valid JSON

3. **Check Cloud Run logs:**
   ```bash
   gcloud run services logs read SERVICE_NAME \
     --region=us-central1 \
     --limit=100
   ```

#### Issue 3: Database Connection Fails

**Symptoms:**

- Application starts but can't connect to database
- "Connection refused" or "Authentication failed" errors

**Solutions:**

1. **Verify DATABASE_URL:**

   ```bash
   # Test connection
   psql "YOUR_DATABASE_URL"
   ```

2. **Check format:**

   ```
   Correct: postgresql://user:password@host:port/database?sslmode=require
   Wrong: postgres://user:password@host:port/database (missing 'ql')
   ```

3. **Verify Prisma.io settings:**
   - [ ] Database is running
   - [ ] IP whitelist includes 0.0.0.0/0 (for Cloud Run)
   - [ ] SSL is enabled

#### Issue 4: CORS Errors

**Symptoms:**

- Frontend can't call backend API
- Browser console shows CORS errors

**Solutions:**

1. **Update CORS_ALLOWED_ORIGINS:**

   ```bash
   # Should include frontend URL
   https://liyali-frontend-xyz.run.app,https://yourdomain.com
   ```

2. **Redeploy backend:**

   ```bash
   git commit --allow-empty -m "chore: trigger redeploy"
   git push origin main
   ```

3. **Verify in backend logs:**
   ```bash
   gcloud run services logs read liyali-backend \
     --region=us-central1 \
     | grep CORS
   ```

#### Issue 5: Environment Variables Not Working

**Symptoms:**

- Application can't read environment variables
- Features not working as expected

**Solutions:**

1. **List current env vars:**

   ```bash
   gcloud run services describe liyali-backend \
     --region=us-central1 \
     --format="value(spec.template.spec.containers[0].env)"
   ```

2. **Update env vars:**

   ```bash
   gcloud run services update liyali-backend \
     --region=us-central1 \
     --set-env-vars KEY=VALUE
   ```

3. **Verify in GitHub Secrets:**
   - [ ] Secret names match exactly (case-sensitive)
   - [ ] No typos in secret names
   - [ ] Values are properly formatted

#### Issue 6: Image Pull Errors

**Symptoms:**

- "Failed to pull image" error
- "Authentication required" error

**Solutions:**

1. **Verify GHCR permissions:**

   - Go to GitHub → Settings → Actions → General
   - Ensure "Read and write permissions" is enabled

2. **Check image exists:**

   ```bash
   # List images in GHCR
   gh api /user/packages/container/liyali-gateway-backend/versions
   ```

3. **Manually pull and push:**

   ```bash
   # Login to GHCR
   echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

   # Pull from GHCR
   docker pull ghcr.io/YOUR_ORG/liyali-gateway-backend:latest

   # Tag for GCP
   docker tag ghcr.io/YOUR_ORG/liyali-gateway-backend:latest \
     us-central1-docker.pkg.dev/PROJECT_ID/liyali/backend:latest

   # Push to GCP
   docker push us-central1-docker.pkg.dev/PROJECT_ID/liyali/backend:latest
   ```

---

## Monitoring and Maintenance

### View Logs

```bash
# Backend logs
gcloud run services logs read liyali-backend \
  --region=us-central1 \
  --limit=100 \
  --format=json

# Frontend logs
gcloud run services logs read liyali-frontend \
  --region=us-central1 \
  --limit=100 \
  --format=json

# Follow logs in real-time
gcloud run services logs tail liyali-backend \
  --region=us-central1
```

### Monitor Performance

```bash
# Get service details
gcloud run services describe liyali-backend \
  --region=us-central1

# View metrics in Cloud Console
# Go to: Cloud Run → Service → Metrics tab
```

### Update Service Configuration

```bash
# Update memory
gcloud run services update liyali-backend \
  --region=us-central1 \
  --memory=1Gi

# Update CPU
gcloud run services update liyali-backend \
  --region=us-central1 \
  --cpu=2

# Update scaling
gcloud run services update liyali-backend \
  --region=us-central1 \
  --min-instances=1 \
  --max-instances=20
```

---

## Rollback Procedure

If a deployment causes issues:

### Quick Rollback

```bash
# List revisions
gcloud run revisions list \
  --service=liyali-backend \
  --region=us-central1

# Rollback to previous revision
gcloud run services update-traffic liyali-backend \
  --region=us-central1 \
  --to-revisions=REVISION_NAME=100
```

### Git Rollback

```bash
# Revert last commit
git revert HEAD
git push origin main

# Or reset to specific commit
git reset --hard COMMIT_HASH
git push origin main --force
```

---

## Security Best Practices

- [ ] Never commit secrets to Git
- [ ] Rotate JWT_SECRET and NEXTAUTH_SECRET regularly
- [ ] Use strong passwords for database
- [ ] Enable Cloud Armor for DDoS protection
- [ ] Set up Cloud Monitoring alerts
- [ ] Regular security audits
- [ ] Keep dependencies updated
- [ ] Use least privilege for service accounts
- [ ] Enable audit logging
- [ ] Implement rate limiting

---

## Cost Optimization

### Cloud Run Pricing Tips

- [ ] Set appropriate min/max instances
- [ ] Use `--min-instances=0` for development
- [ ] Monitor request patterns
- [ ] Optimize container size
- [ ] Use caching where possible
- [ ] Set appropriate timeout values
- [ ] Review and delete unused revisions

### Estimated Monthly Costs

**Development/Staging:**

- Cloud Run: $5-20/month
- Prisma.io: Free tier or $25/month
- Total: ~$5-45/month

**Production (moderate traffic):**

- Cloud Run: $50-200/month
- Prisma.io: $25-100/month
- Total: ~$75-300/month

---

## Support and Resources

### Documentation Links

- [Google Cloud Run Docs](https://cloud.google.com/run/docs)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Prisma Docs](https://www.prisma.io/docs)
- [Docker Docs](https://docs.docker.com)

### Getting Help

- GitHub Issues: [Your Repo Issues]
- Cloud Run Support: [GCP Support](https://cloud.google.com/support)
- Community: [Stack Overflow](https://stackoverflow.com/questions/tagged/google-cloud-run)

---

## Appendix

### A. Environment Variables Reference

See [ENVIRONMENT_VARIABLES.md](./ENVIRONMENT_VARIABLES.md) for complete list.

### B. API Endpoints

See [docs/API.md](./docs/API.md) for API documentation.

### C. Database Schema

See [backend/database/migrations/](./backend/database/migrations/) for schema.

---

**Last Updated:** January 2026
**Version:** 1.0.0
