# Admin Console Troubleshooting Guide

## 🔍 Common Issues and Solutions

### Issue 1: CORS Errors in Browser Console

**Symptoms:**

- Browser console shows: `Access to fetch at 'https://liyali-gateway-api.fly.dev/api/v1/...' from origin 'https://liyali-admin-console.fly.dev' has been blocked by CORS policy`
- API requests fail with CORS errors
- Login doesn't work

**Diagnosis:**

```bash
# Check current CORS settings
flyctl secrets list --app liyali-gateway-api | grep CORS
```

**Solution:**

```bash
# Update CORS to include admin console
flyctl secrets set \
  CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" \
  --app liyali-gateway-api

# Restart backend to apply changes
flyctl restart --app liyali-gateway-api

# Wait 30 seconds and test
sleep 30
curl -H "Origin: https://liyali-admin-console.fly.dev" \
     -H "Access-Control-Request-Method: POST" \
     -X OPTIONS \
     https://liyali-gateway-api.fly.dev/api/v1/auth/login
```

**Prevention:**

- Always update CORS when adding new frontend applications
- Include all origins in a comma-separated list
- Test CORS after deployment

---

### Issue 2: Build Fails During Deployment

**Symptoms:**

- GitHub Actions shows build failure
- Error: `npm ERR! code ELIFECYCLE`
- Error: `TypeScript compilation failed`

**Diagnosis:**

```bash
# Test build locally
cd admin-console
npm run build

# Check for TypeScript errors
npm run lint
```

**Common Causes:**

1. Missing dependencies in package.json
2. TypeScript compilation errors
3. Environment variables not set during build
4. Out of memory during build

**Solutions:**

**For Missing Dependencies:**

```bash
cd admin-console
npm install
npm run build
```

**For TypeScript Errors:**

```bash
# Check for errors
npm run lint

# Fix errors in code
# Then commit and push
```

**For Memory Issues:**

```bash
# Increase memory in fly.toml
# Change memory_mb from 512 to 1024
flyctl scale memory 1024 --app liyali-admin-console
```

**For Environment Variables:**

```bash
# Ensure secrets are set
flyctl secrets list --app liyali-admin-console

# Set missing secrets
flyctl secrets set NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" --app liyali-admin-console
```

---

### Issue 3: App Won't Start After Deployment

**Symptoms:**

- Deployment succeeds but app doesn't start
- Health checks fail
- Status shows "stopped" or "crashed"

**Diagnosis:**

```bash
# Check app status
flyctl status --app liyali-admin-console

# View recent logs
flyctl logs --app liyali-admin-console --lines 100

# Check for common errors
flyctl logs --app liyali-admin-console | grep -i error
```

**Common Causes:**

1. Missing environment variables
2. Port mismatch
3. Startup script issues
4. Memory limits

**Solutions:**

**Check Environment Variables:**

```bash
# List all secrets
flyctl secrets list --app liyali-admin-console

# Set required secrets
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  NEXTAUTH_SECRET="your-secret" \
  NEXTAUTH_URL="https://liyali-admin-console.fly.dev" \
  --app liyali-admin-console
```

**Check Port Configuration:**

```bash
# Verify fly.toml has correct port
cat admin-console/fly.toml | grep internal_port
# Should show: internal_port = 3001

# Verify Dockerfile exposes correct port
cat admin-console/Dockerfile | grep EXPOSE
# Should show: EXPOSE 3001
```

**Restart App:**

```bash
flyctl restart --app liyali-admin-console
```

**Increase Memory:**

```bash
flyctl scale memory 1024 --app liyali-admin-console
```

---

### Issue 4: Health Check Fails

**Symptoms:**

- Deployment reports: "Health check failed"
- App shows as unhealthy in Fly.io dashboard
- Cannot access admin console URL

**Diagnosis:**

```bash
# Test health endpoint directly
curl -v https://liyali-admin-console.fly.dev/

# Check app status
flyctl status --app liyali-admin-console

# View logs
flyctl logs --app liyali-admin-console -f
```

**Solutions:**

**If App is Not Running:**

```bash
# Restart app
flyctl restart --app liyali-admin-console

# Check logs for startup errors
flyctl logs --app liyali-admin-console
```

**If Health Check Path is Wrong:**

```bash
# Verify fly.toml health check path
cat admin-console/fly.toml | grep path
# Should show: path = "/"

# Test the path manually
curl https://liyali-admin-console.fly.dev/
```

**If App is Slow to Start:**

```bash
# Increase grace period in fly.toml
# Change grace_period from "10s" to "30s"
# Then redeploy
flyctl deploy --remote-only --app liyali-admin-console
```

---

### Issue 5: Cannot Connect to Backend API

**Symptoms:**

- Admin console loads but shows "Cannot connect to server"
- API requests timeout
- Network errors in browser console

**Diagnosis:**

```bash
# Test backend health
curl https://liyali-gateway-api.fly.dev/health

# Check admin console API URL
flyctl secrets list --app liyali-admin-console | grep API_URL

# Test from admin console
curl -H "Origin: https://liyali-admin-console.fly.dev" \
     https://liyali-gateway-api.fly.dev/api/v1/health
```

**Solutions:**

**Check API URL:**

```bash
# Verify API URL is correct
flyctl secrets list --app liyali-admin-console

# Update if wrong
flyctl secrets set \
  NEXT_PUBLIC_API_URL="https://liyali-gateway-api.fly.dev/api/v1" \
  --app liyali-admin-console
```

**Check Backend Status:**

```bash
# Ensure backend is running
flyctl status --app liyali-gateway-api

# Restart if needed
flyctl restart --app liyali-gateway-api
```

**Check CORS (see Issue 1)**

---

### Issue 6: Deployment Stuck or Timeout

**Symptoms:**

- GitHub Actions shows "Waiting for deployment..."
- Deployment times out after 10 minutes
- No progress in logs

**Diagnosis:**

```bash
# Check Fly.io status
flyctl status --app liyali-admin-console

# Check recent deployments
flyctl releases --app liyali-admin-console

# View build logs
flyctl logs --app liyali-admin-console
```

**Solutions:**

**Cancel and Retry:**

```bash
# Cancel stuck deployment
# (In GitHub Actions, click "Cancel workflow")

# Retry deployment
cd admin-console
flyctl deploy --remote-only
```

**Increase Timeout:**

```bash
# In GitHub Actions workflow, increase wait-timeout
# Change from 300 to 600 seconds
# Edit .github/workflows/fly-deploy.yml
```

**Check Fly.io Status:**

- Visit https://status.fly.io/
- Check for ongoing incidents

---

### Issue 7: Authentication Not Working

**Symptoms:**

- Cannot log in to admin console
- "Invalid credentials" error
- Session not persisting

**Diagnosis:**

```bash
# Check NEXTAUTH_SECRET is set
flyctl secrets list --app liyali-admin-console | grep NEXTAUTH

# Check NEXTAUTH_URL is correct
flyctl secrets list --app liyali-admin-console | grep NEXTAUTH_URL

# Test backend auth endpoint
curl -X POST https://liyali-gateway-api.fly.dev/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'
```

**Solutions:**

**Set NEXTAUTH_SECRET:**

```bash
# Generate new secret
NEXTAUTH_SECRET=$(openssl rand -base64 32)

# Set secret
flyctl secrets set \
  NEXTAUTH_SECRET="$NEXTAUTH_SECRET" \
  --app liyali-admin-console
```

**Set NEXTAUTH_URL:**

```bash
flyctl secrets set \
  NEXTAUTH_URL="https://liyali-admin-console.fly.dev" \
  --app liyali-admin-console
```

**Check Backend Auth:**

```bash
# Ensure backend is working
curl https://liyali-gateway-api.fly.dev/health

# Check backend logs
flyctl logs --app liyali-gateway-api
```

---

### Issue 8: Changes Not Reflected After Deployment

**Symptoms:**

- Deployed new code but changes not visible
- Old version still showing
- Cache issues

**Diagnosis:**

```bash
# Check deployment history
flyctl releases --app liyali-admin-console

# Verify latest deployment
flyctl status --app liyali-admin-console
```

**Solutions:**

**Hard Refresh Browser:**

- Chrome/Firefox: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)
- Clear browser cache

**Verify Deployment:**

```bash
# Check latest release
flyctl releases --app liyali-admin-console

# View deployment logs
flyctl logs --app liyali-admin-console
```

**Force Redeploy:**

```bash
cd admin-console
flyctl deploy --remote-only --no-cache
```

---

### Issue 9: Out of Memory Errors

**Symptoms:**

- App crashes randomly
- Logs show "JavaScript heap out of memory"
- Build fails with memory errors

**Diagnosis:**

```bash
# Check current memory allocation
flyctl status --app liyali-admin-console

# View memory usage
flyctl metrics --app liyali-admin-console
```

**Solutions:**

**Increase Memory:**

```bash
# Scale to 1GB
flyctl scale memory 1024 --app liyali-admin-console

# Or scale to 2GB for build-heavy apps
flyctl scale memory 2048 --app liyali-admin-console
```

**Optimize Build:**

```bash
# In next.config.ts, add:
# experimental: {
#   workerThreads: false,
#   cpus: 1
# }
```

---

### Issue 10: SSL/TLS Certificate Errors

**Symptoms:**

- Browser shows "Your connection is not private"
- SSL certificate warnings
- HTTPS not working

**Diagnosis:**

```bash
# Check certificate status
flyctl certs list --app liyali-admin-console

# Test HTTPS
curl -v https://liyali-admin-console.fly.dev/
```

**Solutions:**

**Wait for Certificate:**

- Fly.io automatically provisions certificates
- Can take 1-5 minutes after first deployment
- Check status: `flyctl certs show liyali-admin-console.fly.dev`

**Force Certificate Renewal:**

```bash
# Remove and re-add certificate
flyctl certs remove liyali-admin-console.fly.dev --app liyali-admin-console
flyctl certs create liyali-admin-console.fly.dev --app liyali-admin-console
```

---

## 🛠️ Diagnostic Commands

### Quick Health Check

```bash
# All-in-one health check
echo "Backend:" && curl -s https://liyali-gateway-api.fly.dev/health && \
echo "Frontend:" && curl -s https://liyali-gateway-frontend.fly.dev/ > /dev/null && echo "OK" && \
echo "Admin:" && curl -s https://liyali-admin-console.fly.dev/ > /dev/null && echo "OK"
```

### View All Logs

```bash
# View logs from all apps
flyctl logs --app liyali-gateway-api &
flyctl logs --app liyali-gateway-frontend &
flyctl logs --app liyali-admin-console &
```

### Check All Secrets

```bash
# List secrets for all apps
echo "Backend secrets:" && flyctl secrets list --app liyali-gateway-api
echo "Frontend secrets:" && flyctl secrets list --app liyali-gateway-frontend
echo "Admin secrets:" && flyctl secrets list --app liyali-admin-console
```

### Full Status Report

```bash
# Generate status report
echo "=== Backend ===" && flyctl status --app liyali-gateway-api
echo "=== Frontend ===" && flyctl status --app liyali-gateway-frontend
echo "=== Admin Console ===" && flyctl status --app liyali-admin-console
```

---

## 📞 Getting Help

### 1. Check Documentation

- [Deployment Guide](./docs/FLY_IO_DEPLOYMENT_GUIDE.md)
- [Setup Instructions](./ADMIN_CONSOLE_DEPLOYMENT_SETUP.md)
- [Deployment Checklist](./DEPLOYMENT_CHECKLIST.md)

### 2. Check Logs

```bash
flyctl logs --app liyali-admin-console -f
```

### 3. Check Fly.io Status

- Visit: https://status.fly.io/

### 4. Fly.io Community

- Forum: https://community.fly.io/
- Discord: https://fly.io/discord

### 5. GitHub Issues

- Open an issue in the repository with:
  - Error message
  - Deployment logs
  - Steps to reproduce

---

## 🔄 Emergency Rollback

If deployment causes critical issues:

```bash
# List recent releases
flyctl releases --app liyali-admin-console

# Rollback to previous version
flyctl releases rollback --app liyali-admin-console

# Verify rollback
flyctl status --app liyali-admin-console
```

---

**Last Updated**: February 8, 2026
**Version**: 1.0
