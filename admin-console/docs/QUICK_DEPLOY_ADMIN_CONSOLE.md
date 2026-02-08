# Quick Deploy - Admin Console

## 🚀 One-Time Setup (5 minutes)

### 1. Create Fly.io App

```bash
cd admin-console
flyctl apps create liyali-admin-console
```

### 2. Set Secrets

```bash
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  NEXTAUTH_SECRET="$(openssl rand -base64 32)" \
  NEXTAUTH_URL="https://liyali-admin-console.fly.dev" \
  --app liyali-admin-console
```

### 3. Update Backend CORS

```bash
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api
```

### 4. Deploy

```bash
flyctl deploy --remote-only
```

## 🔄 Regular Deployments

### Automatic (Recommended)

```bash
git add .
git commit -m "Update admin console"
git push origin develop
```

### Manual

```bash
cd admin-console
flyctl deploy --remote-only
```

### Using Script

```bash
# Linux/Mac
./scripts/deploy-admin-console.sh

# Windows
.\scripts\deploy-admin-console.ps1
```

## ✅ Verify Deployment

```bash
# Check status
flyctl status --app liyali-admin-console

# Test URL
curl https://liyali-admin-console.fly.dev/

# View logs
flyctl logs --app liyali-admin-console -f
```

## 🔧 Common Commands

```bash
# View logs
flyctl logs --app liyali-admin-console

# Restart app
flyctl restart --app liyali-admin-console

# Check secrets
flyctl secrets list --app liyali-admin-console

# Scale app
flyctl scale count 1 --app liyali-admin-console

# SSH into container
flyctl ssh console --app liyali-admin-console
```

## 🐛 Quick Fixes

### CORS Error

```bash
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api
```

### App Won't Start

```bash
flyctl logs --app liyali-admin-console
flyctl restart --app liyali-admin-console
```

### Build Fails

```bash
# Test locally first
cd admin-console
npm run build
```

## 📊 URLs

- **Admin Console**: https://liyali-admin-console.fly.dev
- **Backend API**: https://liyali-gateway-api.fly.dev
- **Frontend**: https://liyali-gateway-frontend.fly.dev

## 📚 Full Documentation

- [Complete Setup Guide](./ADMIN_CONSOLE_DEPLOYMENT_SETUP.md)
- [Deployment Checklist](./DEPLOYMENT_CHECKLIST.md)
- [Fly.io Guide](./docs/FLY_IO_DEPLOYMENT_GUIDE.md)

---

**Need Help?** Check the full documentation or run the deployment script.
