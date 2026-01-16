# Deployment Checklist - Quick Reference

Use this checklist for deploying Liyali Gateway to Google Cloud Run.

## 🚀 Pre-Deployment Setup (One-Time)

### Google Cloud Platform

```
□ Create GCP project
□ Enable billing
□ Enable Cloud Run API
□ Enable Artifact Registry API
□ Enable Cloud Build API
□ Create Artifact Registry repository (name: liyali)
□ Create service account (github-actions)
□ Grant roles: run.admin, artifactregistry.writer, iam.serviceAccountUser
□ Download service account key (gcp-key.json)
```

### Database (Prisma.io)

```
□ Create Prisma.io account
□ Create new project
□ Select PostgreSQL
□ Choose region (close to Cloud Run)
□ Copy connection string
□ Test connection: psql "CONNECTION_STRING"
```

### GitHub Repository

```
□ Enable GitHub Actions
□ Set workflow permissions to "Read and write"
□ Allow Actions to create/approve PRs
```

### GitHub Secrets (Settings → Secrets → Actions)

```
Google Cloud:
□ GCP_SA_KEY (entire gcp-key.json content)
□ GCP_PROJECT_ID (e.g., liyali-gateway-123456)
□ GCP_REGION (e.g., us-central1)

Services:
□ BACKEND_SERVICE_NAME (e.g., liyali-backend)
□ FRONTEND_SERVICE_NAME (e.g., liyali-frontend)

Database:
□ DATABASE_URL (postgresql://...)

Backend:
□ JWT_SECRET (generate: openssl rand -base64 32)
□ CORS_ALLOWED_ORIGINS (will update after frontend deploy)

Frontend:
□ NEXT_PUBLIC_API_URL (will update after backend deploy)
□ NEXTAUTH_SECRET (generate: openssl rand -base64 32)
□ NEXTAUTH_URL (will update after frontend deploy)
```

---

## 📦 First Deployment

### Step 1: Deploy Backend

```
□ Commit backend code
□ Push to main branch: git push origin main
□ Go to GitHub → Actions tab
□ Wait for "Backend - Build and Deploy" to complete
□ Go to GCP Console → Cloud Run
□ Click liyali-backend service
□ Copy service URL (e.g., https://liyali-backend-xyz.run.app)
□ Update GitHub Secret: NEXT_PUBLIC_API_URL = backend URL
```

### Step 2: Run Database Migrations

```
□ cd backend
□ export DATABASE_URL="your-connection-string"
□ make db-migrate
   OR
□ psql $DATABASE_URL < database/migrations/001_init_system.up.sql
□ psql $DATABASE_URL < database/migrations/002_seed_data.up.sql
□ Verify tables created
```

### Step 3: Deploy Frontend

```
□ Ensure NEXT_PUBLIC_API_URL is set in GitHub Secrets
□ Commit frontend code
□ Push to main branch: git push origin main
□ Go to GitHub → Actions tab
□ Wait for "Frontend - Build and Deploy" to complete
□ Go to GCP Console → Cloud Run
□ Click liyali-frontend service
□ Copy service URL (e.g., https://liyali-frontend-xyz.run.app)
□ Update GitHub Secret: NEXTAUTH_URL = frontend URL
```

### Step 4: Update CORS

```
□ Update GitHub Secret: CORS_ALLOWED_ORIGINS = frontend URL
□ Trigger backend redeploy:
  - Make small change to backend code, OR
  - git commit --allow-empty -m "chore: update CORS"
  - git push origin main
```

---

## ✅ Verification

### Backend Health Check

```bash
export BACKEND_URL="https://liyali-backend-xyz.run.app"

□ curl $BACKEND_URL/health
  Expected: {"status":"ok"}

□ curl $BACKEND_URL/api/v1/health
  Expected: {"status":"healthy"}

□ Check logs:
  gcloud run services logs read liyali-backend --region=us-central1 --limit=50
```

### Frontend Health Check

```bash
export FRONTEND_URL="https://liyali-frontend-xyz.run.app"

□ curl -I $FRONTEND_URL
  Expected: HTTP/2 200

□ Open in browser
□ Test login page loads
□ Test authentication
□ Test API calls work
□ Check browser console for errors
```

### Integration Tests

```
□ Login with test user
□ Create requisition
□ Test workflow approval
□ Check notifications
□ Test search functionality
□ Verify all features working
```

---

## 🔄 Subsequent Deployments

### Backend Update

```
□ Make changes to backend code
□ git add backend/
□ git commit -m "feat: description"
□ git push origin main
□ GitHub Actions automatically deploys
□ Verify deployment in Actions tab
```

### Frontend Update

```
□ Make changes to frontend code
□ git add frontend/
□ git commit -m "feat: description"
□ git push origin main
□ GitHub Actions automatically deploys
□ Verify deployment in Actions tab
```

### Full-Stack Update

```
□ Make changes to both backend and frontend
□ git add .
□ git commit -m "feat: description"
□ git push origin main
□ Both workflows run in parallel
□ Verify both deployments
```

---

## 🐛 Troubleshooting Quick Fixes

### Build Fails

```
□ Check GitHub Actions logs
□ Test Docker build locally: docker build -t test .
□ Verify all dependencies in package.json/go.mod
□ Check Dockerfile syntax
□ Ensure all files committed
```

### Deployment Fails

```
□ Verify GCP_SA_KEY is valid JSON
□ Check service account permissions
□ Verify all GitHub Secrets are set
□ Check Cloud Run logs:
  gcloud run services logs read SERVICE_NAME --region=us-central1
```

### Database Connection Fails

```
□ Test connection: psql "DATABASE_URL"
□ Verify format: postgresql://user:pass@host:port/db?sslmode=require
□ Check Prisma.io IP whitelist (allow 0.0.0.0/0)
□ Verify SSL is enabled
□ Check DATABASE_URL in GitHub Secrets
```

### CORS Errors

```
□ Update CORS_ALLOWED_ORIGINS with frontend URL
□ Redeploy backend
□ Clear browser cache
□ Check backend logs for CORS messages
```

### Environment Variables Not Working

```
□ List current env vars:
  gcloud run services describe SERVICE_NAME --region=us-central1
□ Verify secret names match exactly (case-sensitive)
□ Check for typos in GitHub Secrets
□ Redeploy service after updating secrets
```

---

## 🔙 Rollback Procedure

### Quick Rollback

```
□ List revisions:
  gcloud run revisions list --service=SERVICE_NAME --region=us-central1

□ Rollback to previous:
  gcloud run services update-traffic SERVICE_NAME \
    --region=us-central1 \
    --to-revisions=REVISION_NAME=100
```

### Git Rollback

```
□ Revert last commit:
  git revert HEAD
  git push origin main

□ Or reset to specific commit:
  git reset --hard COMMIT_HASH
  git push origin main --force
```

---

## 📊 Monitoring

### View Logs

```bash
# Backend logs
gcloud run services logs read liyali-backend --region=us-central1 --limit=100

# Frontend logs
gcloud run services logs read liyali-frontend --region=us-central1 --limit=100

# Follow logs in real-time
gcloud run services logs tail liyali-backend --region=us-central1
```

### Check Service Status

```bash
# Service details
gcloud run services describe liyali-backend --region=us-central1

# List all services
gcloud run services list --region=us-central1
```

---

## 🔐 Security Checklist

```
□ Never commit secrets to Git
□ Rotate JWT_SECRET monthly
□ Rotate NEXTAUTH_SECRET monthly
□ Use strong database passwords
□ Enable Cloud Armor (optional)
□ Set up monitoring alerts
□ Regular security audits
□ Keep dependencies updated
□ Review service account permissions
□ Enable audit logging
```

---

## 💰 Cost Optimization

```
□ Set min-instances=0 for dev/staging
□ Set appropriate max-instances
□ Monitor request patterns
□ Optimize container size
□ Delete unused revisions
□ Review monthly billing
□ Set up budget alerts
```

---

## 📝 Environment Variables Quick Reference

### Backend Required

```
DATABASE_URL          - PostgreSQL connection string
JWT_SECRET           - JWT signing secret
CORS_ALLOWED_ORIGINS - Comma-separated allowed origins
PORT                 - Server port (default: 8080)
ENV                  - Environment (production/staging)
```

### Frontend Required

```
NEXT_PUBLIC_API_URL  - Backend API URL
NEXTAUTH_SECRET      - NextAuth signing secret
NEXTAUTH_URL         - Frontend URL
DATABASE_URL         - PostgreSQL connection (for NextAuth)
NODE_ENV             - Node environment
```

---

## 🆘 Emergency Contacts

```
□ GCP Support: https://cloud.google.com/support
□ GitHub Support: https://support.github.com
□ Prisma Support: https://www.prisma.io/support
□ Team Lead: [Add contact]
□ DevOps: [Add contact]
```

---

## 📚 Quick Links

- [Full Deployment Guide](./DEPLOYMENT_GUIDE.md)
- [GitHub Actions Workflows](./.github/workflows/)
- [Backend Dockerfile](./backend/Dockerfile)
- [Frontend Dockerfile](./frontend/Dockerfile)
- [API Documentation](./docs/API.md)
- [Database Migrations](./backend/database/migrations/)

---

**Print this checklist and keep it handy for deployments!**

**Last Updated:** January 2026
