# Fly.io Deployment Guide

## Prerequisites

- Fly.io account
- Fly CLI installed
- Docker installed

## Install Fly CLI

```bash
# macOS/Linux
curl -L https://fly.io/install.sh | sh

# Windows
powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
```

## Login

```bash
flyctl auth login
```

## Backend Deployment

### 1. Initialize App

```bash
cd backend
flyctl launch --no-deploy

# Choose app name: liyali-backend
# Choose region: closest to users
# Don't deploy yet
```

### 2. Configure Database

```bash
# Create Postgres database
flyctl postgres create --name liyali-db

# Attach to app
flyctl postgres attach liyali-db
```

### 3. Set Secrets

```bash
flyctl secrets set JWT_SECRET=<your-secret>
flyctl secrets set ENVIRONMENT=production
```

### 4. Deploy

```bash
flyctl deploy
```

### 5. Run Migrations

```bash
flyctl ssh console
./liyali-backend migrate
exit
```

## Frontend Deployment

### 1. Initialize App

```bash
cd frontend
flyctl launch --no-deploy

# Choose app name: liyali-frontend
```

### 2. Set Environment

```bash
flyctl secrets set NEXT_PUBLIC_API_URL=https://liyali-backend.fly.dev
flyctl secrets set NEXT_PUBLIC_APP_URL=https://liyali-frontend.fly.dev
```

### 3. Deploy

```bash
flyctl deploy
```

## Admin Console Deployment

### 1. Initialize App

```bash
cd admin-console
flyctl launch --no-deploy

# Choose app name: liyali-admin
```

### 2. Set Environment

```bash
flyctl secrets set NEXT_PUBLIC_API_URL=https://liyali-backend.fly.dev
```

### 3. Deploy

```bash
flyctl deploy
```

## Custom Domains

### Backend API

```bash
cd backend
flyctl certs create api.liyali.com
```

Add DNS records:

```
A     api.liyali.com    -> <fly-ip>
AAAA  api.liyali.com    -> <fly-ipv6>
```

### Frontend

```bash
cd frontend
flyctl certs create liyali.com
flyctl certs create www.liyali.com
```

Add DNS records:

```
A     liyali.com        -> <fly-ip>
AAAA  liyali.com        -> <fly-ipv6>
CNAME www.liyali.com    -> liyali.com
```

### Admin Console

```bash
cd admin-console
flyctl certs create admin.liyali.com
```

## Scaling

### Vertical Scaling

```bash
# Increase VM size
flyctl scale vm shared-cpu-2x --memory 1024
```

### Horizontal Scaling

```bash
# Add more instances
flyctl scale count 2

# Auto-scaling
flyctl autoscale set min=1 max=3
```

## Monitoring

### View Logs

```bash
# Real-time logs
flyctl logs

# Specific app
flyctl logs -a liyali-backend
```

### Check Status

```bash
flyctl status
flyctl info
```

### Metrics

```bash
flyctl dashboard
```

## Database Management

### Backup

```bash
# Create backup
flyctl postgres backup create -a liyali-db

# List backups
flyctl postgres backup list -a liyali-db
```

### Connect to Database

```bash
# Via proxy
flyctl proxy 5432 -a liyali-db

# Then connect
psql postgresql://user:pass@localhost:5432/liyali_gateway
```

## Troubleshooting

### App Not Starting

```bash
# Check logs
flyctl logs -a liyali-backend

# SSH into instance
flyctl ssh console -a liyali-backend
```

### Database Connection Issues

```bash
# Check database status
flyctl status -a liyali-db

# Restart database
flyctl restart -a liyali-db
```

### SSL Certificate Issues

```bash
# Check certificate status
flyctl certs show api.liyali.com

# Remove and recreate
flyctl certs delete api.liyali.com
flyctl certs create api.liyali.com
```

## CI/CD with GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Fly.io

on:
  push:
    branches: [main]

jobs:
  deploy-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: superfly/flyctl-actions/setup-flyctl@master
      - run: flyctl deploy --remote-only
        working-directory: ./backend
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}

  deploy-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: superfly/flyctl-actions/setup-flyctl@master
      - run: flyctl deploy --remote-only
        working-directory: ./frontend
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

## Cost Optimization

- Use shared CPU for development
- Scale down during off-hours
- Use auto-scaling
- Monitor resource usage
- Optimize database queries

## Resources

- [Fly.io Documentation](https://fly.io/docs/)
- [Fly.io Postgres](https://fly.io/docs/postgres/)
- [Deployment Guide](./04-DEPLOYMENT.md)
