# Environment Variables Reference

Complete reference for all environment variables used in Liyali Gateway.

## 📋 Table of Contents

- [Backend Environment Variables](#backend-environment-variables)
- [Frontend Environment Variables](#frontend-environment-variables)
- [GitHub Secrets](#github-secrets)
- [Local Development](#local-development)
- [Production Configuration](#production-configuration)

---

## Backend Environment Variables

### Required Variables

| Variable               | Description                                 | Example                                               | Required |
| ---------------------- | ------------------------------------------- | ----------------------------------------------------- | -------- |
| `DATABASE_URL`         | PostgreSQL connection string from Prisma.io | `postgresql://user:pass@host:5432/db?sslmode=require` | ✅ Yes   |
| `JWT_SECRET`           | Secret key for JWT token signing            | `your-super-secret-jwt-key-here`                      | ✅ Yes   |
| `PORT`                 | Server port                                 | `8080`                                                | ✅ Yes   |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed origins     | `https://frontend.com,https://app.com`                | ✅ Yes   |

### Optional Variables

| Variable              | Description                  | Default       | Example                                |
| --------------------- | ---------------------------- | ------------- | -------------------------------------- |
| `ENV`                 | Environment name             | `development` | `production`, `staging`, `development` |
| `LOG_LEVEL`           | Logging level                | `info`        | `debug`, `info`, `warn`, `error`       |
| `RATE_LIMIT_REQUESTS` | Max requests per window      | `100`         | `100`                                  |
| `RATE_LIMIT_WINDOW`   | Rate limit window in seconds | `60`          | `60`                                   |
| `SESSION_TIMEOUT`     | Session timeout in hours     | `24`          | `24`                                   |
| `MAX_UPLOAD_SIZE`     | Max file upload size in MB   | `10`          | `10`                                   |

### Database Configuration

```bash
# Format
DATABASE_URL="postgresql://[user]:[password]@[host]:[port]/[database]?sslmode=require"

# Example (Prisma.io)
DATABASE_URL="postgresql://user:password@aws-0-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require"

# Example (Cloud SQL)
DATABASE_URL="postgresql://user:password@/database?host=/cloudsql/project:region:instance"
```

### JWT Configuration

```bash
# Generate secure JWT secret
openssl rand -base64 32

# Example
JWT_SECRET="xK9mP2nQ5rT8wY1zA4bC7dE0fG3hJ6kL"
```

### CORS Configuration

```bash
# Single origin
CORS_ALLOWED_ORIGINS="https://app.example.com"

# Multiple origins
CORS_ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com,http://localhost:3000"

# Development (allow all - NOT for production)
CORS_ALLOWED_ORIGINS="*"
```

---

## Frontend Environment Variables

### Required Variables

| Variable              | Description                                   | Example                               | Required |
| --------------------- | --------------------------------------------- | ------------------------------------- | -------- |
| `NEXT_PUBLIC_API_URL` | Backend API base URL                          | `https://api.example.com`             | ✅ Yes   |
| `NEXTAUTH_SECRET`     | NextAuth.js secret for session encryption     | `your-nextauth-secret-here`           | ✅ Yes   |
| `NEXTAUTH_URL`        | Frontend application URL                      | `https://app.example.com`             | ✅ Yes   |
| `DATABASE_URL`        | PostgreSQL connection (for NextAuth sessions) | `postgresql://user:pass@host:5432/db` | ✅ Yes   |

### Optional Variables

| Variable                  | Description               | Default          | Example                     |
| ------------------------- | ------------------------- | ---------------- | --------------------------- |
| `NODE_ENV`                | Node environment          | `development`    | `production`, `development` |
| `NEXT_PUBLIC_APP_NAME`    | Application name          | `Liyali Gateway` | `Liyali Gateway`            |
| `NEXT_PUBLIC_APP_VERSION` | Application version       | `1.0.0`          | `1.0.0`                     |
| `NEXT_TELEMETRY_DISABLED` | Disable Next.js telemetry | `1`              | `1`                         |

### NextAuth Configuration

```bash
# Generate secure NextAuth secret
openssl rand -base64 32

# Example
NEXTAUTH_SECRET="aB2cD4eF6gH8iJ0kL1mN3oP5qR7sT9uV"

# NextAuth URL (must match deployed URL)
NEXTAUTH_URL="https://liyali-frontend-xyz.run.app"
```

### API Configuration

```bash
# Production
NEXT_PUBLIC_API_URL="https://liyali-backend-xyz.run.app"

# Staging
NEXT_PUBLIC_API_URL="https://liyali-backend-staging-xyz.run.app"

# Local development
NEXT_PUBLIC_API_URL="http://localhost:8080"
```

---

## GitHub Secrets

All secrets must be added to: **Repository Settings → Secrets and variables → Actions**

### Google Cloud Platform Secrets

```yaml
GCP_SA_KEY:
  Description: Service account key JSON (entire file content)
  Format: JSON
  Example: |
    {
      "type": "service_account",
      "project_id": "liyali-gateway-123456",
      "private_key_id": "abc123...",
      "private_key": "-----BEGIN PRIVATE KEY-----\n...",
      "client_email": "github-actions@project.iam.gserviceaccount.com",
      ...
    }
  How to get: gcloud iam service-accounts keys create gcp-key.json --iam-account=...

GCP_PROJECT_ID:
  Description: Google Cloud project ID
  Format: String
  Example: liyali-gateway-123456
  How to get: gcloud config get-value project

GCP_REGION:
  Description: Google Cloud region for deployment
  Format: String
  Example: us-central1
  Options: us-central1, us-east1, europe-west1, asia-southeast1
```

### Service Configuration Secrets

```yaml
BACKEND_SERVICE_NAME:
  Description: Cloud Run service name for backend
  Format: String (lowercase, hyphens only)
  Example: liyali-backend
  Note: Must be unique within project

FRONTEND_SERVICE_NAME:
  Description: Cloud Run service name for frontend
  Format: String (lowercase, hyphens only)
  Example: liyali-frontend
  Note: Must be unique within project
```

### Database Secrets

```yaml
DATABASE_URL:
  Description: PostgreSQL connection string
  Format: postgresql://[user]:[password]@[host]:[port]/[database]?sslmode=require
  Example: postgresql://user:pass@host.pooler.supabase.com:5432/postgres?sslmode=require
  How to get: Copy from Prisma.io dashboard or your PostgreSQL provider
  Security: Never commit this to Git!
```

### Backend Application Secrets

```yaml
JWT_SECRET:
  Description: Secret key for JWT token signing
  Format: Base64 string (32+ characters)
  Example: xK9mP2nQ5rT8wY1zA4bC7dE0fG3hJ6kL
  How to generate: openssl rand -base64 32
  Security: Rotate monthly in production

CORS_ALLOWED_ORIGINS:
  Description: Comma-separated list of allowed origins
  Format: https://domain1.com,https://domain2.com
  Example: https://liyali-frontend-xyz.run.app,https://yourdomain.com
  Note: Update after frontend deployment
```

### Frontend Application Secrets

```yaml
NEXT_PUBLIC_API_URL:
  Description: Backend API base URL
  Format: https://domain.com (no trailing slash)
  Example: https://liyali-backend-xyz.run.app
  Note: Update after backend deployment

NEXTAUTH_SECRET:
  Description: NextAuth.js secret for session encryption
  Format: Base64 string (32+ characters)
  Example: aB2cD4eF6gH8iJ0kL1mN3oP5qR7sT9uV
  How to generate: openssl rand -base64 32
  Security: Rotate monthly in production

NEXTAUTH_URL:
  Description: Frontend application URL
  Format: https://domain.com (no trailing slash)
  Example: https://liyali-frontend-xyz.run.app
  Note: Update after frontend deployment
```

---

## Local Development

### Backend `.env` File

Create `backend/.env`:

```bash
# Database
DATABASE_URL="postgresql://user:password@localhost:5432/liyali_gateway?sslmode=disable"

# Server
PORT=8080
ENV=development

# Security
JWT_SECRET="local-dev-jwt-secret-change-in-production"

# CORS
CORS_ALLOWED_ORIGINS="http://localhost:3000,http://localhost:3001"

# Logging
LOG_LEVEL=debug

# Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=60

# Session
SESSION_TIMEOUT=24

# File Upload
MAX_UPLOAD_SIZE=10
```

### Frontend `.env.local` File

Create `frontend/.env.local`:

```bash
# API
NEXT_PUBLIC_API_URL="http://localhost:8080"

# NextAuth
NEXTAUTH_SECRET="local-dev-nextauth-secret-change-in-production"
NEXTAUTH_URL="http://localhost:3000"

# Database (for NextAuth)
DATABASE_URL="postgresql://user:password@localhost:5432/liyali_gateway?sslmode=disable"

# App
NEXT_PUBLIC_APP_NAME="Liyali Gateway (Dev)"
NEXT_PUBLIC_APP_VERSION="1.0.0-dev"

# Node
NODE_ENV=development
NEXT_TELEMETRY_DISABLED=1
```

### Docker Compose `.env` File

Create `.env` in root:

```bash
# PostgreSQL
POSTGRES_USER=liyali
POSTGRES_PASSWORD=liyali_password
POSTGRES_DB=liyali_gateway
POSTGRES_PORT=5432

# Backend
BACKEND_PORT=8080
JWT_SECRET="docker-dev-jwt-secret"

# Frontend
FRONTEND_PORT=3000
NEXTAUTH_SECRET="docker-dev-nextauth-secret"
```

---

## Production Configuration

### Backend Production Environment

```bash
# Database (Prisma.io)
DATABASE_URL="postgresql://user:password@aws-0-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require"

# Server
PORT=8080
ENV=production

# Security
JWT_SECRET="<STRONG_RANDOM_SECRET_32_CHARS>"

# CORS (update with actual frontend URL)
CORS_ALLOWED_ORIGINS="https://liyali-frontend-xyz.run.app,https://yourdomain.com"

# Logging
LOG_LEVEL=info

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# Session
SESSION_TIMEOUT=24

# File Upload
MAX_UPLOAD_SIZE=10
```

### Frontend Production Environment

```bash
# API (update with actual backend URL)
NEXT_PUBLIC_API_URL="https://liyali-backend-xyz.run.app"

# NextAuth
NEXTAUTH_SECRET="<STRONG_RANDOM_SECRET_32_CHARS>"
NEXTAUTH_URL="https://liyali-frontend-xyz.run.app"

# Database (Prisma.io - same as backend)
DATABASE_URL="postgresql://user:password@aws-0-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require"

# App
NEXT_PUBLIC_APP_NAME="Liyali Gateway"
NEXT_PUBLIC_APP_VERSION="1.0.0"

# Node
NODE_ENV=production
NEXT_TELEMETRY_DISABLED=1
```

---

## Environment-Specific Configurations

### Development

```bash
# Relaxed security for easier development
JWT_SECRET="dev-secret"
CORS_ALLOWED_ORIGINS="*"
LOG_LEVEL=debug
RATE_LIMIT_REQUESTS=10000
```

### Staging

```bash
# Similar to production but with staging URLs
JWT_SECRET="<STAGING_SECRET>"
CORS_ALLOWED_ORIGINS="https://staging-frontend.com"
LOG_LEVEL=info
RATE_LIMIT_REQUESTS=500
ENV=staging
```

### Production

```bash
# Strict security settings
JWT_SECRET="<PRODUCTION_SECRET>"
CORS_ALLOWED_ORIGINS="https://app.yourdomain.com"
LOG_LEVEL=warn
RATE_LIMIT_REQUESTS=100
ENV=production
```

---

## Security Best Practices

### Secret Generation

```bash
# Generate JWT_SECRET
openssl rand -base64 32

# Generate NEXTAUTH_SECRET
openssl rand -base64 32

# Generate random password
openssl rand -base64 24

# Generate UUID
uuidgen
```

### Secret Rotation

1. **Monthly Rotation** (Recommended):

   - JWT_SECRET
   - NEXTAUTH_SECRET

2. **Quarterly Rotation**:

   - Database passwords
   - Service account keys

3. **Rotation Process**:

   ```bash
   # 1. Generate new secret
   NEW_SECRET=$(openssl rand -base64 32)

   # 2. Update GitHub Secret
   # Go to Settings → Secrets → Edit

   # 3. Trigger redeployment
   git commit --allow-empty -m "chore: rotate secrets"
   git push origin main

   # 4. Verify deployment
   # 5. Update documentation
   ```

### Secret Storage

- ✅ **DO**: Store in GitHub Secrets
- ✅ **DO**: Use environment variables
- ✅ **DO**: Use secret management services (GCP Secret Manager)
- ❌ **DON'T**: Commit to Git
- ❌ **DON'T**: Share in plain text
- ❌ **DON'T**: Log secrets
- ❌ **DON'T**: Include in error messages

---

## Validation

### Backend Environment Validation

```bash
# Check required variables
cd backend

# Test database connection
psql "$DATABASE_URL" -c "SELECT 1"

# Validate JWT secret length
echo -n "$JWT_SECRET" | wc -c  # Should be 32+

# Test CORS origins format
echo "$CORS_ALLOWED_ORIGINS" | grep -E '^https?://'
```

### Frontend Environment Validation

```bash
# Check required variables
cd frontend

# Test API connectivity
curl "$NEXT_PUBLIC_API_URL/health"

# Validate NextAuth secret length
echo -n "$NEXTAUTH_SECRET" | wc -c  # Should be 32+

# Test NextAuth URL format
echo "$NEXTAUTH_URL" | grep -E '^https?://'
```

---

## Troubleshooting

### Common Issues

#### Issue: "DATABASE_URL is not defined"

**Solution:**

```bash
# Check if variable is set
echo $DATABASE_URL

# Set in current shell
export DATABASE_URL="postgresql://..."

# Add to .env file
echo 'DATABASE_URL="postgresql://..."' >> .env
```

#### Issue: "CORS error in browser"

**Solution:**

```bash
# Check CORS_ALLOWED_ORIGINS includes frontend URL
echo $CORS_ALLOWED_ORIGINS

# Update to include frontend
export CORS_ALLOWED_ORIGINS="https://frontend.com"

# Redeploy backend
```

#### Issue: "NextAuth session not persisting"

**Solution:**

```bash
# Check NEXTAUTH_SECRET is set
echo $NEXTAUTH_SECRET

# Check NEXTAUTH_URL matches deployed URL
echo $NEXTAUTH_URL

# Regenerate secret if needed
export NEXTAUTH_SECRET=$(openssl rand -base64 32)
```

---

## Quick Reference

### Generate All Secrets at Once

```bash
#!/bin/bash
# generate-secrets.sh

echo "=== Liyali Gateway Secrets Generator ==="
echo ""
echo "JWT_SECRET=$(openssl rand -base64 32)"
echo "NEXTAUTH_SECRET=$(openssl rand -base64 32)"
echo "ADMIN_PASSWORD=$(openssl rand -base64 24)"
echo ""
echo "Copy these to your GitHub Secrets!"
```

### Environment Variables Checklist

```
Backend:
□ DATABASE_URL
□ JWT_SECRET
□ PORT
□ CORS_ALLOWED_ORIGINS
□ ENV

Frontend:
□ NEXT_PUBLIC_API_URL
□ NEXTAUTH_SECRET
□ NEXTAUTH_URL
□ DATABASE_URL
□ NODE_ENV

GitHub Secrets:
□ GCP_SA_KEY
□ GCP_PROJECT_ID
□ GCP_REGION
□ BACKEND_SERVICE_NAME
□ FRONTEND_SERVICE_NAME
□ All backend variables
□ All frontend variables
```

---

**Last Updated:** January 2026
**Version:** 1.0.0
