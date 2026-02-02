# Fly.io Deployment Guide

## 🚀 Quick Setup

### 1. Required GitHub Secrets

Add these secrets to your GitHub repository: **Settings → Secrets and variables → Actions**

#### Core Secrets

```bash
FLY_API_TOKEN                 # Your Fly.io API token
FLY_DATABASE_URL             # PostgreSQL connection string for Fly.io
JWT_SECRET                   # JWT signing secret
NEXTAUTH_SECRET              # NextAuth signing secret
FLY_CORS_ALLOWED_ORIGINS     # CORS origins (e.g., https://liyali-gateway-frontend.fly.dev)
```

#### How to Get These Values

**FLY_API_TOKEN**:

```bash
# Login to Fly.io
flyctl auth login

# Get your token
flyctl auth token
```

**FLY_DATABASE_URL**:

```bash
# Create a PostgreSQL database on Fly.io
flyctl postgres create --name liyali-db --region jnb

# Get connection string
flyctl postgres connect --app liyali-db
# Copy the DATABASE_URL from the output
```

**JWT_SECRET & NEXTAUTH_SECRET**:

```bash
# Generate secure random secrets
openssl rand -base64 32
```

**FLY_CORS_ALLOWED_ORIGINS**:

```
https://liyali-gateway-frontend.fly.dev,http://localhost:3000
```

### 2. Deploy to Fly.io

#### Option A: Automatic Deployment

Push to `develop` branch:

```bash
git checkout develop
git push origin develop
```

#### Option B: Manual Deployment

```bash
# Go to GitHub Actions
# Select "Deploy to Staging Environment - (Fly.io)"
# Click "Run workflow"
# Select develop branch
# Click "Run workflow"
```

### 3. Verify Deployment

Check the deployment status:

```bash
# Backend
curl https://liyali-gateway-api.fly.dev/health

# Frontend
curl https://liyali-gateway-frontend.fly.dev/api/health
```

## 🔧 Advanced Configuration

### Database Setup

1. **Create PostgreSQL Database**:

```bash
flyctl postgres create --name liyali-db --region jnb --vm-size shared-cpu-1x --volume-size 10
```

2. **Get Connection Details**:

```bash
flyctl postgres connect --app liyali-db
```

3. **Set Database URL**:

```bash
# For backend
flyctl secrets set DATABASE_URL="postgresql://..." --app liyali-gateway-api

# For frontend (for server actions)
flyctl secrets set DATABASE_URL="postgresql://..." --app liyali-gateway-frontend
```

### Environment Variables

#### Backend Secrets

```bash
flyctl secrets set \
  DATABASE_URL="postgresql://user:pass@host:5432/db" \
  JWT_SECRET="your-jwt-secret" \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev" \
  --app liyali-gateway-api
```

#### Frontend Secrets

```bash
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  DATABASE_URL="postgresql://user:pass@host:5432/db" \
  NEXTAUTH_SECRET="your-nextauth-secret" \
  NEXTAUTH_URL="https://liyali-gateway-frontend.fly.dev" \
  --app liyali-gateway-frontend
```

### Custom Domains

1. **Add Custom Domain**:

```bash
flyctl certs create yourdomain.com --app liyali-gateway-frontend
flyctl certs create api.yourdomain.com --app liyali-gateway-api
```

2. **Update DNS**:

```
CNAME yourdomain.com liyali-gateway-frontend.fly.dev
CNAME api.yourdomain.com liyali-gateway-api.fly.dev
```

3. **Update Environment Variables**:

```bash
# Update CORS origins
flyctl secrets set CORS_ALLOWED_ORIGINS="https://yourdomain.com" --app liyali-gateway-api

# Update NextAuth URL
flyctl secrets set NEXTAUTH_URL="https://yourdomain.com" --app liyali-gateway-frontend

# Update API URL
flyctl secrets set NEXT_PUBLIC_API_URL="https://api.yourdomain.com/api/v1" --app liyali-gateway-frontend
```

## 🔍 Monitoring & Debugging

### View Logs

```bash
# Backend logs
flyctl logs --app liyali-gateway-api

# Frontend logs
flyctl logs --app liyali-gateway-frontend

# Follow logs in real-time
flyctl logs --app liyali-gateway-api -f
```

### Check App Status

```bash
# Backend status
flyctl status --app liyali-gateway-api

# Frontend status
flyctl status --app liyali-gateway-frontend
```

### Scale Applications

```bash
# Scale backend
flyctl scale count 2 --app liyali-gateway-api

# Scale frontend
flyctl scale count 2 --app liyali-gateway-frontend

# Scale memory
flyctl scale memory 1024 --app liyali-gateway-api
```

### Health Checks

```bash
# Test backend health
curl https://liyali-gateway-api.fly.dev/health

# Test frontend health
curl https://liyali-gateway-frontend.fly.dev/api/health

# Test full application
curl https://liyali-gateway-frontend.fly.dev/
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Database Connection Failed

```bash
# Check database status
flyctl status --app liyali-db

# Test connection
flyctl postgres connect --app liyali-db

# Check secrets
flyctl secrets list --app liyali-gateway-api
```

#### 2. Frontend Can't Connect to Backend

```bash
# Check CORS settings
flyctl secrets list --app liyali-gateway-api

# Update CORS origins
flyctl secrets set CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev" --app liyali-gateway-api

# Check API URL
flyctl secrets list --app liyali-gateway-frontend
```

#### 3. Build Failures

```bash
# Check build logs
flyctl logs --app liyali-gateway-api

# Deploy with verbose output
flyctl deploy --verbose --app liyali-gateway-api
```

#### 4. Health Check Failures

```bash
# Check health endpoint directly
curl -v https://liyali-gateway-api.fly.dev/health

# Check application logs
flyctl logs --app liyali-gateway-api

# Restart application
flyctl restart --app liyali-gateway-api
```

### Rollback Deployment

```bash
# List releases
flyctl releases --app liyali-gateway-api

# Rollback to previous release
flyctl releases rollback --app liyali-gateway-api
```

### Debug Mode

```bash
# SSH into running container
flyctl ssh console --app liyali-gateway-api

# Check environment variables
flyctl ssh console --app liyali-gateway-api -C "env"
```

## 📊 Performance Optimization

### Resource Allocation

```bash
# Optimize for cost (shared CPU)
flyctl scale vm shared-cpu-1x --memory 512 --app liyali-gateway-api

# Optimize for performance (dedicated CPU)
flyctl scale vm dedicated-cpu-1x --memory 1024 --app liyali-gateway-api
```

### Auto-scaling

```toml
# In fly.toml
[http_service.concurrency]
  type = "connections"
  hard_limit = 50
  soft_limit = 40

[[vm]]
  cpu_kind = "shared"
  cpus = 1
  memory_mb = 512
```

### Caching

```bash
# Enable Redis for caching
flyctl redis create --name liyali-cache --region jnb

# Get Redis URL
flyctl redis status liyali-cache
```

## 🔐 Security Best Practices

### 1. Secrets Management

- ✅ Use `flyctl secrets set` for sensitive data
- ✅ Never commit secrets to git
- ✅ Rotate secrets regularly
- ✅ Use different secrets for staging/production

### 2. Network Security

- ✅ Enable HTTPS only (`force_https = true`)
- ✅ Configure proper CORS origins
- ✅ Use private networking for database
- ✅ Implement rate limiting

### 3. Access Control

- ✅ Use least privilege principle
- ✅ Enable audit logging
- ✅ Monitor access patterns
- ✅ Regular security updates

## 📈 Monitoring Setup

### Application Metrics

```bash
# View metrics
flyctl metrics --app liyali-gateway-api

# Set up alerts
flyctl alerts create --app liyali-gateway-api
```

### Log Aggregation

```bash
# Export logs to external service
flyctl logs --app liyali-gateway-api --format json > logs.json
```

### Uptime Monitoring

```bash
# Set up external monitoring
curl -f https://liyali-gateway-api.fly.dev/health || echo "Service down"
```

## 🔄 CI/CD Pipeline

The deployment pipeline automatically:

1. ✅ **Builds** Docker images
2. ✅ **Sets** environment variables
3. ✅ **Deploys** to Fly.io
4. ✅ **Verifies** health checks
5. ✅ **Reports** deployment status

### Pipeline Triggers

- Push to `develop` branch
- Manual workflow dispatch
- Path-based triggers (`backend/**`, `frontend/**`)

### Pipeline Features

- 🔄 Automatic rollback on failure
- 📊 Deployment verification
- 📝 Detailed logging
- 🚨 Slack/email notifications (configurable)

## 📞 Support

### Getting Help

1. Check [Fly.io Documentation](https://fly.io/docs/)
2. Review deployment logs in GitHub Actions
3. Check application logs with `flyctl logs`
4. Open issue in repository

### Useful Commands

```bash
# Quick status check
flyctl status --app liyali-gateway-api

# Quick restart
flyctl restart --app liyali-gateway-api

# Quick scale
flyctl scale count 1 --app liyali-gateway-api

# Quick logs
flyctl logs --app liyali-gateway-api -f
```

---

**Last Updated**: February 2, 2026
**Version**: 1.0
**Status**: Production Ready
