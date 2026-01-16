# Fly.io Deployment - Complete Step-by-Step Guide

**Time Required**: 20-30 minutes  
**Cost**: FREE (within free tier)  
**Difficulty**: Beginner-friendly

This guide walks you through deploying Liyali Gateway to Fly.io with screenshots and detailed explanations.

---

## 📋 What You'll Need

- [ ] Computer with internet connection
- [ ] GitHub account (for code)
- [ ] Credit/debit card (for Fly.io verification - won't be charged on free tier)
- [ ] 30 minutes of time

---

## 🎯 Overview - What We'll Do

```
Step 1: Create Accounts (5 min)
   ↓
Step 2: Create Database (5 min)
   ↓
Step 3: Install Fly.io CLI (2 min)
   ↓
Step 4: Deploy Backend (8 min)
   ↓
Step 5: Setup Database (3 min)
   ↓
Step 6: Deploy Frontend (7 min)
   ↓
Step 7: Test Application (5 min)
```

---

# STEP 1: Create Accounts (5 minutes)

## 1.1 Create Fly.io Account

### Option A: Sign up with GitHub (Recommended)

1. Go to https://fly.io/app/sign-up
2. Click **"Sign up with GitHub"**
3. Authorize Fly.io to access your GitHub account
4. You'll be redirected to Fly.io dashboard

### Option B: Sign up with Email

1. Go to https://fly.io/app/sign-up
2. Enter your email address
3. Create a password
4. Verify your email
5. Complete the signup

### 1.2 Add Payment Method

**Important**: This is required but you won't be charged on the free tier.

1. Go to https://fly.io/dashboard/personal/billing
2. Click **"Add Payment Method"**
3. Enter your credit/debit card details
4. Click **"Add Card"**

✅ **Checkpoint**: You should see "Payment method added" confirmation

---

## 1.3 Create Prisma.io Account (For Database)

1. Go to https://cloud.prisma.io
2. Click **"Sign up"**
3. Choose **"Continue with GitHub"** (easiest)
4. Authorize Prisma Data Platform
5. Complete your profile

✅ **Checkpoint**: You should see the Prisma dashboard

---

# STEP 2: Create Database (5 minutes)

## 2.1 Create PostgreSQL Database on Prisma.io

1. In Prisma dashboard, click **"New Project"**

2. Fill in the form:

   - **Project name**: `liyali-gateway`
   - **Description**: `Liyali Gateway Database` (optional)

3. Click **"Create Project"**

4. Select **"PostgreSQL"** as database type

5. Choose your region:

   - **US East**: If you're in Americas
   - **EU West**: If you're in Europe
   - **Asia Southeast**: If you're in Asia

6. Click **"Create Database"**

7. Wait 30-60 seconds for database to be created

## 2.2 Get Connection String

1. Once created, you'll see the database dashboard

2. Click **"Connection String"** or **"Connect"**

3. Copy the connection string - it looks like:

   ```
   postgresql://user:password@aws-0-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require
   ```

4. **IMPORTANT**: Save this in a secure place (like a password manager)
   - You'll need it multiple times
   - Don't share it publicly
   - Don't commit it to Git

✅ **Checkpoint**: You have a connection string saved

---

# STEP 3: Install Fly.io CLI (2 minutes)

## 3.1 Install Based on Your Operating System

### macOS

Open Terminal and run:

```bash
curl -L https://fly.io/install.sh | sh
```

### Windows

Open PowerShell as Administrator and run:

```powershell
pwsh -Command "iwr https://fly.io/install.ps1 -useb | iex"
```

### Linux

Open Terminal and run:

```bash
curl -L https://fly.io/install.sh | sh
```

## 3.2 Verify Installation

```bash
flyctl version
```

You should see something like:

```
flyctl v0.x.xxx linux/amd64 Commit: xxxxx BuildDate: 2024-xx-xx
```

## 3.3 Login to Fly.io

```bash
flyctl auth login
```

This will:

1. Open your browser
2. Ask you to confirm login
3. Show "Successfully logged in" message

✅ **Checkpoint**: You're logged into Fly.io CLI

---

# STEP 4: Deploy Backend (8 minutes)

## 4.1 Navigate to Backend Directory

```bash
cd backend
```

## 4.2 Create Fly.io App

```bash
flyctl launch --no-deploy
```

You'll be asked several questions:

**Question 1**: "Choose an app name (leave blank to generate one)"

- **Answer**: Press Enter (let Fly.io generate a name)
- Or type: `liyali-backend-yourname`

**Question 2**: "Choose a region for deployment"

- **Answer**: Select the region closest to you
- Use arrow keys to navigate, Enter to select
- Recommended: `iad` (US East) or `lhr` (London)

**Question 3**: "Would you like to set up a Postgresql database now?"

- **Answer**: `N` (No - we're using Prisma.io)

**Question 4**: "Would you like to set up an Upstash Redis database now?"

- **Answer**: `N` (No - not needed)

**Question 5**: "Would you like to deploy now?"

- **Answer**: `N` (No - we need to set secrets first)

✅ **Checkpoint**: You should see "Your app is ready!" message

## 4.3 Set Backend Secrets

### Generate JWT Secret

First, generate a secure JWT secret:

**macOS/Linux**:

```bash
openssl rand -base64 32
```

**Windows PowerShell**:

```powershell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | % {[char]$_})
```

Copy the output (you'll need it in the next step).

### Set All Secrets

Replace the values with your actual values:

```bash
# Set database URL (use your Prisma.io connection string)
flyctl secrets set DATABASE_URL="postgresql://user:pass@host:5432/db?sslmode=require"

# Set JWT secret (use the one you just generated)
flyctl secrets set JWT_SECRET="your-generated-jwt-secret-here"

# Set CORS (we'll update this after frontend deployment)
flyctl secrets set CORS_ALLOWED_ORIGINS="*"

# Set environment
flyctl secrets set ENV="production"

# Set port
flyctl secrets set PORT="8080"
```

**Important**:

- Replace `DATABASE_URL` with your actual Prisma.io connection string
- Replace `JWT_SECRET` with the secret you generated
- Don't include the quotes in the actual values

✅ **Checkpoint**: You should see "Secrets are staged for the first deployment"

## 4.4 Deploy Backend

```bash
flyctl deploy
```

This will:

1. Build your Docker image (takes 2-5 minutes)
2. Push to Fly.io
3. Deploy to your region
4. Start the application

**What you'll see**:

```
==> Building image
==> Pushing image to fly
==> Deploying
==> Monitoring deployment
```

Wait for: `✓ Deployment successful`

## 4.5 Get Backend URL

```bash
flyctl info
```

Look for the **Hostname** line:

```
Hostname = liyali-backend-xyz.fly.dev
```

Your backend URL is: `https://liyali-backend-xyz.fly.dev`

**SAVE THIS URL** - you'll need it for the frontend!

## 4.6 Test Backend

```bash
curl https://your-backend-url.fly.dev/health
```

You should see:

```json
{ "status": "ok", "timestamp": "2024-01-16T..." }
```

✅ **Checkpoint**: Backend is deployed and responding!

---

# STEP 5: Setup Database (3 minutes)

## 5.1 Run Database Migrations

From the `backend` directory:

```bash
# Set the database URL as environment variable
export DATABASE_URL="your-prisma-connection-string"

# Run migrations
make db-migrate
```

**If you don't have `make` installed**, run manually:

```bash
# Run first migration
psql "$DATABASE_URL" < database/migrations/001_init_system.up.sql

# Run seed data
psql "$DATABASE_URL" < database/migrations/002_seed_data.up.sql
```

**What this does**:

- Creates all database tables
- Sets up relationships
- Inserts seed data (test users, roles, etc.)

✅ **Checkpoint**: You should see "CREATE TABLE" messages without errors

---

# STEP 6: Deploy Frontend (7 minutes)

## 6.1 Navigate to Frontend Directory

```bash
cd ../frontend
```

## 6.2 Create Fly.io App

```bash
flyctl launch --no-deploy
```

Answer the questions:

**Question 1**: "Choose an app name"

- **Answer**: Press Enter or type `liyali-frontend-yourname`

**Question 2**: "Choose a region"

- **Answer**: Select the SAME region as your backend

**Question 3**: "Would you like to set up a Postgresql database?"

- **Answer**: `N` (No)

**Question 4**: "Would you like to set up an Upstash Redis database?"

- **Answer**: `N` (No)

**Question 5**: "Would you like to deploy now?"

- **Answer**: `N` (No - need to set secrets first)

✅ **Checkpoint**: "Your app is ready!" message

## 6.3 Set Frontend Secrets

### Generate NextAuth Secret

**macOS/Linux**:

```bash
openssl rand -base64 32
```

**Windows PowerShell**:

```powershell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | % {[char]$_})
```

### Set All Secrets

Replace with your actual values:

```bash
# Set backend API URL (use the URL from Step 4.5)
flyctl secrets set NEXT_PUBLIC_API_URL="https://your-backend-url.fly.dev"

# Set NextAuth secret (use the one you just generated)
flyctl secrets set NEXTAUTH_SECRET="your-generated-nextauth-secret"

# Set NextAuth URL (we'll update this after deployment)
flyctl secrets set NEXTAUTH_URL="https://liyali-frontend-xyz.fly.dev"

# Set database URL (same as backend)
flyctl secrets set DATABASE_URL="your-prisma-connection-string"

# Set environment
flyctl secrets set NODE_ENV="production"
```

✅ **Checkpoint**: "Secrets are staged for the first deployment"

## 6.4 Deploy Frontend

```bash
flyctl deploy
```

This will:

1. Build Next.js application (takes 3-7 minutes)
2. Create optimized production build
3. Push to Fly.io
4. Deploy and start

**Note**: Frontend build takes longer than backend (this is normal!)

Wait for: `✓ Deployment successful`

## 6.5 Get Frontend URL

```bash
flyctl info
```

Look for the **Hostname**:

```
Hostname = liyali-frontend-xyz.fly.dev
```

Your frontend URL is: `https://liyali-frontend-xyz.fly.dev`

**SAVE THIS URL!**

✅ **Checkpoint**: Frontend is deployed!

---

# STEP 7: Update CORS Settings (2 minutes)

Now that we have the frontend URL, update backend CORS:

## 7.1 Update Backend CORS

```bash
# Go back to backend directory
cd ../backend

# Update CORS with actual frontend URL
flyctl secrets set CORS_ALLOWED_ORIGINS="https://your-frontend-url.fly.dev"

# Restart backend to apply changes
flyctl apps restart
```

Wait 10-20 seconds for restart to complete.

✅ **Checkpoint**: CORS updated

---

# STEP 8: Test Your Application (5 minutes)

## 8.1 Test Backend Health

```bash
curl https://your-backend-url.fly.dev/health
```

Expected response:

```json
{ "status": "ok", "timestamp": "..." }
```

## 8.2 Test Frontend Health

```bash
curl https://your-frontend-url.fly.dev/api/health
```

Expected response:

```json
{ "status": "ok", "timestamp": "...", "service": "liyali-frontend" }
```

## 8.3 Open Application in Browser

1. Open your browser
2. Go to: `https://your-frontend-url.fly.dev`
3. You should see the Liyali Gateway login page

## 8.4 Test Login

Use the seeded test account:

- **Email**: `admin@liyali.com`
- **Password**: `Admin@123`

(Check your seed data file for actual credentials)

## 8.5 Test Features

Try these to verify everything works:

- [ ] Login successfully
- [ ] View dashboard
- [ ] Create a requisition
- [ ] View tasks
- [ ] Test workflow approval
- [ ] Check notifications

✅ **Checkpoint**: Application is fully functional!

---

# 🎉 Congratulations!

Your Liyali Gateway is now live on Fly.io!

## 📝 Save These URLs

**Backend**: `https://your-backend-url.fly.dev`  
**Frontend**: `https://your-frontend-url.fly.dev`  
**Database**: `postgresql://...` (your Prisma.io connection string)

---

# 📊 What's Next?

## Share Your Demo

Share the frontend URL with:

- Team members
- Stakeholders
- Clients
- Testers

## Monitor Your Application

### View Logs

```bash
# Backend logs
cd backend
flyctl logs

# Frontend logs
cd frontend
flyctl logs
```

### Check Status

```bash
# Backend status
cd backend
flyctl status

# Frontend status
cd frontend
flyctl status
```

### View Dashboard

Go to: https://fly.io/dashboard

You can see:

- Application status
- Resource usage
- Deployment history
- Logs

## Make Updates

### Update Backend

```bash
cd backend
# Make your changes
git add .
git commit -m "feat: update backend"
flyctl deploy
```

### Update Frontend

```bash
cd frontend
# Make your changes
git add .
git commit -m "feat: update frontend"
flyctl deploy
```

---

# 🔧 Common Issues & Solutions

## Issue 1: "flyctl: command not found"

**Solution**:

```bash
# Add to PATH (macOS/Linux)
export PATH="$HOME/.fly/bin:$PATH"

# Or reinstall
curl -L https://fly.io/install.sh | sh
```

## Issue 2: Build Fails - "go.mod requires go >= 1.24.0"

**Solution**: Update your `go.mod` file:

```bash
cd backend
# Edit go.mod and change the go version to match your installed version
go mod tidy
flyctl deploy
```

## Issue 3: Database Connection Fails

**Solution**:

1. Verify DATABASE_URL format:
   ```
   postgresql://user:pass@host:5432/db?sslmode=require
   ```
2. Ensure `?sslmode=require` is at the end
3. Test connection:
   ```bash
   psql "your-connection-string"
   ```

## Issue 4: CORS Errors in Browser

**Solution**:

```bash
cd backend
flyctl secrets set CORS_ALLOWED_ORIGINS="https://your-frontend-url.fly.dev"
flyctl apps restart
```

## Issue 5: Frontend Shows 500 Error

**Solution**:

1. Check frontend logs:
   ```bash
   cd frontend
   flyctl logs
   ```
2. Verify NEXT_PUBLIC_API_URL is set correctly
3. Ensure backend is running

## Issue 6: "Out of Memory" Error

**Solution**:

```bash
# Scale up memory
flyctl scale memory 512

# Or edit fly.toml and change memory_mb
```

## Issue 7: Deployment Takes Too Long

**Solution**:

- This is normal for first deployment (5-10 minutes)
- Subsequent deployments are faster (2-3 minutes)
- Use `--verbose` flag to see progress:
  ```bash
  flyctl deploy --verbose
  ```

---

# 💰 Cost Tracking

## Free Tier Limits

You get for FREE:

- 3 shared-cpu-1x VMs (256MB RAM each)
- 3GB persistent volume storage
- 160GB outbound data transfer

## Check Your Usage

1. Go to https://fly.io/dashboard/personal/billing
2. View current usage
3. Set up billing alerts (recommended)

## Stay Within Free Tier

Your current setup uses:

- Backend: 1 VM (256MB) ✅
- Frontend: 1 VM (512MB) ✅
- Total: 2 VMs, 768MB ✅ Within free tier!

---

# 🔐 Security Checklist

- [ ] Never commit secrets to Git
- [ ] Use strong JWT_SECRET (32+ characters)
- [ ] Use strong NEXTAUTH_SECRET (32+ characters)
- [ ] Keep DATABASE_URL private
- [ ] Update CORS to specific frontend URL (not "\*")
- [ ] Enable 2FA on Fly.io account
- [ ] Enable 2FA on Prisma.io account
- [ ] Regularly update dependencies
- [ ] Monitor logs for suspicious activity

---

# 📚 Additional Resources

## Documentation

- [Fly.io Docs](https://fly.io/docs/)
- [Fly.io Go Guide](https://fly.io/docs/languages-and-frameworks/golang/)
- [Prisma Docs](https://www.prisma.io/docs)

## Support

- [Fly.io Community](https://community.fly.io/)
- [Fly.io Discord](https://fly.io/discord)
- [Prisma Discord](https://pris.ly/discord)

## Useful Commands

```bash
# View all apps
flyctl apps list

# View app info
flyctl info

# View logs
flyctl logs

# SSH into machine
flyctl ssh console

# View secrets
flyctl secrets list

# Scale app
flyctl scale count 2
flyctl scale memory 512

# Restart app
flyctl apps restart

# Destroy app (careful!)
flyctl apps destroy app-name
```

---

# ✅ Final Checklist

```
Setup:
□ Fly.io account created
□ Prisma.io account created
□ Payment method added to Fly.io
□ Fly.io CLI installed and logged in

Database:
□ PostgreSQL database created on Prisma.io
□ Connection string saved securely
□ Migrations run successfully
□ Seed data inserted

Backend:
□ App created on Fly.io
□ Secrets configured
□ Deployed successfully
□ Health check passes
□ URL saved

Frontend:
□ App created on Fly.io
□ Secrets configured
□ Deployed successfully
□ Health check passes
□ URL saved

Final Steps:
□ CORS updated with frontend URL
□ Backend restarted
□ Login tested
□ Features verified
□ URLs shared with team
```

---

**Total Time**: ~30 minutes  
**Status**: ✅ DEPLOYED  
**Environment**: Production-ready demo

**Questions?** Check the troubleshooting section or open an issue on GitHub.

**Last Updated**: January 2026
