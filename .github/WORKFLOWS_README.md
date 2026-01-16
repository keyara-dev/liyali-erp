# GitHub Actions Workflows

This directory contains automated CI/CD workflows for deploying Liyali Gateway to Google Cloud Run.

## 📁 Workflow Files

### `backend-deploy.yml`

Builds and deploys the Go backend service to Cloud Run.

**Triggers:**

- Push to `main` or `develop` branches
- Changes in `backend/**` directory
- Changes to the workflow file itself

**Steps:**

1. Build Docker image
2. Push to GitHub Container Registry (GHCR)
3. Pull from GHCR and push to GCP Artifact Registry
4. Deploy to Cloud Run
5. Output deployment URL

### `frontend-deploy.yml`

Builds and deploys the Next.js frontend application to Cloud Run.

**Triggers:**

- Push to `main` or `develop` branches
- Changes in `frontend/**` directory
- Changes to the workflow file itself

**Steps:**

1. Build Docker image with Next.js optimizations
2. Push to GitHub Container Registry (GHCR)
3. Pull from GHCR and push to GCP Artifact Registry
4. Deploy to Cloud Run
5. Output deployment URL

## 🔧 How It Works

### Deployment Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Developer pushes code to GitHub                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. GitHub Actions detects changes in backend/ or frontend/ │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Build Docker image using Dockerfile                     │
│    - Multi-stage build for optimization                    │
│    - Layer caching for faster builds                       │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Push image to GHCR (ghcr.io)                           │
│    - Tagged with branch name and commit SHA                │
│    - Automatic cleanup of old images                       │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Pull image from GHCR                                    │
│    - Authenticate with GCP                                 │
│    - Tag for GCP Artifact Registry                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Push to GCP Artifact Registry                           │
│    - Region-specific registry                              │
│    - Versioned with commit SHA                             │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. Deploy to Cloud Run                                     │
│    - Zero-downtime deployment                              │
│    - Automatic traffic migration                           │
│    - Environment variables injected                        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 8. Service is live!                                        │
│    - Health checks pass                                    │
│    - URL available in workflow summary                     │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Usage

### Automatic Deployment

Deployments happen automatically when you push code:

```bash
# Deploy backend only
git add backend/
git commit -m "feat: add new endpoint"
git push origin main

# Deploy frontend only
git add frontend/
git commit -m "feat: update UI"
git push origin main

# Deploy both
git add .
git commit -m "feat: full-stack feature"
git push origin main
```

### Manual Deployment

You can also trigger deployments manually from GitHub:

1. Go to **Actions** tab
2. Select the workflow (Backend or Frontend)
3. Click **Run workflow**
4. Select branch
5. Click **Run workflow** button

### Branch-Based Environments

- **`main` branch** → Production environment
- **`develop` branch** → Staging environment

## 📋 Prerequisites

Before workflows can run successfully, ensure:

### GitHub Settings

- [ ] Repository has Actions enabled
- [ ] Workflow permissions set to "Read and write"
- [ ] All required secrets configured (see below)

### Required Secrets

All secrets must be added to: **Settings → Secrets and variables → Actions**

#### Google Cloud Platform

```
GCP_SA_KEY          - Service account key JSON
GCP_PROJECT_ID      - GCP project ID
GCP_REGION          - Deployment region (e.g., us-central1)
```

#### Service Configuration

```
BACKEND_SERVICE_NAME   - Cloud Run service name for backend
FRONTEND_SERVICE_NAME  - Cloud Run service name for frontend
```

#### Application Secrets

```
DATABASE_URL           - PostgreSQL connection string
JWT_SECRET            - JWT signing secret
CORS_ALLOWED_ORIGINS  - Allowed CORS origins
NEXT_PUBLIC_API_URL   - Backend API URL
NEXTAUTH_SECRET       - NextAuth signing secret
NEXTAUTH_URL          - Frontend URL
```

See [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md) for detailed setup instructions.

## 🔍 Monitoring Workflows

### View Workflow Runs

1. Go to **Actions** tab in GitHub
2. Click on a workflow run to see details
3. Expand steps to view logs
4. Check deployment summary at the bottom

### Workflow Status

- ✅ **Green checkmark** - Deployment successful
- ❌ **Red X** - Deployment failed
- 🟡 **Yellow dot** - Deployment in progress
- ⚪ **Gray circle** - Workflow queued

### Deployment Summary

Each successful deployment shows:

- Environment (Production/Staging)
- Service name
- Docker image tag
- Deployed URL
- Commit SHA

## 🐛 Troubleshooting

### Build Fails

**Check:**

1. Dockerfile syntax
2. Dependencies in package.json or go.mod
3. Build logs in Actions tab
4. Local Docker build: `docker build -t test .`

**Common fixes:**

- Update dependencies
- Fix syntax errors
- Check for missing files
- Verify build arguments

### Deployment Fails

**Check:**

1. GCP service account permissions
2. All secrets are set correctly
3. Cloud Run service exists
4. Artifact Registry repository exists

**Common fixes:**

- Verify GCP_SA_KEY is valid JSON
- Check service account has required roles
- Ensure region matches across configs
- Verify service names are correct

### Image Pull Fails

**Check:**

1. GHCR permissions
2. Image was pushed successfully
3. GCP authentication

**Common fixes:**

- Enable "Read and write" permissions in Actions settings
- Check GITHUB_TOKEN has package:write scope
- Verify GCP credentials are valid

### Environment Variables Not Working

**Check:**

1. Secrets are set in GitHub
2. Secret names match exactly (case-sensitive)
3. Values don't have extra spaces
4. Secrets are passed to Cloud Run

**Common fixes:**

- Re-add secrets with correct names
- Trim whitespace from values
- Check workflow file passes secrets correctly
- Verify in Cloud Run console

## 📊 Workflow Metrics

### Build Times

- **Backend**: ~3-5 minutes
- **Frontend**: ~5-8 minutes

### Optimization Tips

1. **Use layer caching**

   - Dependencies cached between builds
   - Only rebuild changed layers

2. **Parallel builds**

   - Backend and frontend build simultaneously
   - Independent deployments

3. **Minimize image size**
   - Multi-stage builds
   - Alpine base images
   - Remove unnecessary files

## 🔐 Security

### Secrets Management

- ✅ Secrets stored in GitHub Secrets
- ✅ Never logged or exposed
- ✅ Encrypted at rest
- ✅ Only accessible to workflows

### Image Security

- ✅ Non-root user in containers
- ✅ Minimal base images (Alpine)
- ✅ No secrets in images
- ✅ Regular security updates

### Access Control

- ✅ Service account with minimal permissions
- ✅ Workload Identity for GCP
- ✅ GHCR access controlled by GitHub
- ✅ Cloud Run IAM policies

## 🔄 Rollback

### Automatic Rollback

Cloud Run keeps previous revisions. To rollback:

```bash
# List revisions
gcloud run revisions list \
  --service=SERVICE_NAME \
  --region=REGION

# Rollback to previous
gcloud run services update-traffic SERVICE_NAME \
  --region=REGION \
  --to-revisions=REVISION_NAME=100
```

### Git Rollback

```bash
# Revert last commit
git revert HEAD
git push origin main

# Or reset to specific commit
git reset --hard COMMIT_SHA
git push origin main --force
```

## 📈 Monitoring

### View Logs

```bash
# Backend logs
gcloud run services logs read BACKEND_SERVICE_NAME \
  --region=REGION \
  --limit=100

# Frontend logs
gcloud run services logs read FRONTEND_SERVICE_NAME \
  --region=REGION \
  --limit=100

# Follow logs
gcloud run services logs tail SERVICE_NAME \
  --region=REGION
```

### Metrics

View in Google Cloud Console:

- Cloud Run → Service → Metrics tab
- Request count
- Request latency
- Error rate
- Container instances

## 🛠️ Customization

### Modify Deployment

Edit workflow files to customize:

1. **Trigger conditions**

   ```yaml
   on:
     push:
       paths:
         - "backend/**"
       branches:
         - main
         - develop
   ```

2. **Build arguments**

   ```yaml
   build-args: |
     CUSTOM_ARG=value
   ```

3. **Cloud Run configuration**

   ```yaml
   --memory 1Gi
   --cpu 2
   --min-instances 1
   --max-instances 10
   ```

4. **Environment variables**
   ```yaml
   --set-env-vars KEY=VALUE
   ```

### Add New Workflow

1. Create new file in `.github/workflows/`
2. Define triggers and jobs
3. Add required secrets
4. Test with manual trigger
5. Commit and push

## 📚 Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Docker Documentation](https://docs.docker.com)
- [GHCR Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

## 🆘 Support

For issues or questions:

1. Check [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md)
2. Review [DEPLOYMENT_CHECKLIST.md](../DEPLOYMENT_CHECKLIST.md)
3. Check workflow logs in Actions tab
4. Open an issue in the repository

---

**Last Updated:** January 2026
