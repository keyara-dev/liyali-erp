# Local Docker Setup Guide

Complete guide to run Liyali Gateway locally using Docker.

## 📋 Prerequisites

- [ ] Windows 10/11 (64-bit) or macOS or Linux
- [ ] At least 4GB RAM available
- [ ] 10GB free disk space
- [ ] Internet connection

---

## Step 1: Install Docker Desktop (10 minutes)

### Windows

1. **Download Docker Desktop**

   - Go to: https://www.docker.com/products/docker-desktop/
   - Click "Download for Windows"
   - Wait for download to complete (~500MB)

2. **Install Docker Desktop**

   - Double-click `Docker Desktop Installer.exe`
   - Follow the installation wizard
   - Check "Use WSL 2 instead of Hyper-V" (recommended)
   - Click "Ok" and wait for installation

3. **Start Docker Desktop**

   - Docker Desktop will start automatically
   - You'll see the Docker icon in your system tray
   - Wait for "Docker Desktop is running" message

4. **Verify Installation**

   ```powershell
   docker --version
   docker-compose --version
   ```

   You should see:

   ```
   Docker version 24.x.x
   Docker Compose version v2.x.x
   ```

### macOS

1. **Download Docker Desktop**

   - Go to: https://www.docker.com/products/docker-desktop/
   - Choose your Mac chip:
     - **Apple Silicon (M1/M2/M3)**: Download for Apple Silicon
     - **Intel**: Download for Intel

2. **Install Docker Desktop**

   - Open the downloaded `.dmg` file
   - Drag Docker to Applications folder
   - Open Docker from Applications
   - Grant permissions when prompted

3. **Verify Installation**
   ```bash
   docker --version
   docker-compose --version
   ```

### Linux (Ubuntu/Debian)

```bash
# Update package index
sudo apt-get update

# Install dependencies
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and back in, then verify
docker --version
docker compose version
```

✅ **Checkpoint**: Docker is installed and running

---

## Step 2: Prepare Environment Files (2 minutes)

### Backend Environment

Create `backend/.env`:

```bash
# Database
DATABASE_URL=postgresql://liyali:liyali_dev_password@postgres:5432/liyali_gateway?sslmode=disable

# Server
PORT=8080
ENV=development

# Security
JWT_SECRET=dev_jwt_secret_change_in_production_32chars
ENCRYPTION_KEY=dev_encryption_key_change_in_production

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000

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

### Frontend Environment

Create `frontend/.env.local`:

```bash
# API
NEXT_PUBLIC_API_URL=http://localhost:8080

# NextAuth
NEXTAUTH_SECRET=dev_nextauth_secret_change_in_production_32chars
NEXTAUTH_URL=http://localhost:3000

# Database (for NextAuth)
DATABASE_URL=postgresql://liyali:liyali_dev_password@postgres:5432/liyali_gateway?sslmode=disable

# App
NEXT_PUBLIC_APP_NAME=Liyali Gateway (Dev)
NEXT_PUBLIC_APP_VERSION=1.0.0-dev

# Node
NODE_ENV=development
NEXT_TELEMETRY_DISABLED=1
```

✅ **Checkpoint**: Environment files created

---

## Step 3: Build and Start Containers (5 minutes)

### Option A: Start Everything (Recommended)

```bash
# From project root
docker-compose up --build
```

**What this does:**

- Builds backend Docker image
- Builds frontend Docker image
- Starts PostgreSQL database
- Starts Redis cache
- Starts backend API
- Starts frontend app

**What you'll see:**

```
[+] Building 45.2s (23/23) FINISHED
[+] Running 5/5
 ✔ Network liyali-network       Created
 ✔ Container liyali-postgres    Started
 ✔ Container liyali-redis       Started
 ✔ Container liyali-backend     Started
 ✔ Container liyali-frontend    Started
```

### Option B: Start in Background

```bash
# Start in detached mode
docker-compose up -d --build

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend
```

### Option C: Start Individual Services

```bash
# Start only database
docker-compose up postgres

# Start backend only
docker-compose up backend

# Start frontend only
docker-compose up frontend
```

✅ **Checkpoint**: All containers are running

---

## Step 4: Run Database Migrations (2 minutes)

### Wait for Database to be Ready

```bash
# Check if postgres is healthy
docker-compose ps
```

You should see:

```
NAME                STATUS              PORTS
liyali-postgres     Up (healthy)        0.0.0.0:5432->5432/tcp
```

### Run Migrations

```bash
# Connect to backend container
docker-compose exec backend sh

# Inside container, run migrations
cd /app
make db-migrate

# Or manually
psql $DATABASE_URL < database/migrations/001_init_system.up.sql
psql $DATABASE_URL < database/migrations/002_seed_data.up.sql

# Exit container
exit
```

**Alternative (from host machine):**

```bash
# If you have psql installed locally
psql "postgresql://liyali:liyali_dev_password@localhost:5432/liyali_gateway" < backend/database/migrations/001_init_system.up.sql
psql "postgresql://liyali:liyali_dev_password@localhost:5432/liyali_gateway" < backend/database/migrations/002_seed_data.up.sql
```

✅ **Checkpoint**: Database tables created and seeded

---

## Step 5: Verify Everything is Working (3 minutes)

### Check Container Status

```bash
docker-compose ps
```

Expected output:

```
NAME                IMAGE                    STATUS              PORTS
liyali-backend      liyali-gateway-backend   Up                  0.0.0.0:8080->8080/tcp
liyali-frontend     liyali-gateway-frontend  Up                  0.0.0.0:3000->3000/tcp
liyali-postgres     postgres:15-alpine       Up (healthy)        0.0.0.0:5432->5432/tcp
liyali-redis        redis:7-alpine           Up                  0.0.0.0:6379->6379/tcp
```

### Test Backend API

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","timestamp":"2024-01-16T..."}
```

### Test Frontend

Open your browser and go to:

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/health

### Test Login

1. Go to http://localhost:3000
2. Login with seeded credentials:
   - **Email**: `admin@liyali.com`
   - **Password**: `Admin@123`

✅ **Checkpoint**: Application is fully functional!

---

## 🎯 Common Docker Commands

### Container Management

```bash
# Start all services
docker-compose up

# Start in background
docker-compose up -d

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v

# Restart a service
docker-compose restart backend

# View logs
docker-compose logs -f

# View logs for specific service
docker-compose logs -f backend
```

### Building

```bash
# Rebuild all images
docker-compose build

# Rebuild specific service
docker-compose build backend

# Rebuild without cache
docker-compose build --no-cache

# Build and start
docker-compose up --build
```

### Accessing Containers

```bash
# Execute command in running container
docker-compose exec backend sh

# Run one-off command
docker-compose run backend go version

# Access database
docker-compose exec postgres psql -U liyali -d liyali_gateway
```

### Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove volumes (deletes database data!)
docker-compose down -v

# Remove images
docker-compose down --rmi all

# Clean up everything
docker system prune -a
```

---

## 🔧 Development Workflow

### Making Code Changes

**Backend changes:**

1. Edit code in `backend/` directory
2. Restart backend container:
   ```bash
   docker-compose restart backend
   ```

**Frontend changes:**

1. Edit code in `frontend/` directory
2. Changes auto-reload (hot reload enabled)
3. If not working, restart:
   ```bash
   docker-compose restart frontend
   ```

### Viewing Logs

```bash
# All logs
docker-compose logs -f

# Backend only
docker-compose logs -f backend

# Frontend only
docker-compose logs -f frontend

# Last 100 lines
docker-compose logs --tail=100 backend
```

### Database Access

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U liyali -d liyali_gateway

# Run SQL query
docker-compose exec postgres psql -U liyali -d liyali_gateway -c "SELECT * FROM users;"

# Backup database
docker-compose exec postgres pg_dump -U liyali liyali_gateway > backup.sql

# Restore database
docker-compose exec -T postgres psql -U liyali liyali_gateway < backup.sql
```

---

## 🐛 Troubleshooting

### Issue 1: Port Already in Use

**Error:**

```
Error: bind: address already in use
```

**Solution:**

```bash
# Find what's using the port
# Windows
netstat -ano | findstr :8080

# macOS/Linux
lsof -i :8080

# Kill the process or change port in docker-compose.yml
```

### Issue 2: Database Connection Failed

**Error:**

```
connection refused
```

**Solution:**

```bash
# Check if postgres is running
docker-compose ps postgres

# Check postgres logs
docker-compose logs postgres

# Restart postgres
docker-compose restart postgres

# Wait for health check
docker-compose ps
```

### Issue 3: Frontend Not Loading

**Error:**

```
Module not found
```

**Solution:**

```bash
# Rebuild frontend
docker-compose build frontend

# Clear node_modules volume
docker-compose down -v
docker-compose up --build
```

### Issue 4: Backend Build Fails

**Error:**

```
go: module requires go >= 1.24.0
```

**Solution:**

```bash
# Already fixed in go.mod
# Just rebuild
docker-compose build --no-cache backend
```

### Issue 5: Out of Disk Space

**Solution:**

```bash
# Clean up Docker
docker system prune -a

# Remove unused volumes
docker volume prune

# Check disk usage
docker system df
```

### Issue 6: Slow Performance

**Solution:**

1. Increase Docker Desktop resources:

   - Open Docker Desktop
   - Settings → Resources
   - Increase CPU and Memory
   - Apply & Restart

2. Use volumes for node_modules:
   ```yaml
   volumes:
     - ./frontend:/app
     - /app/node_modules # Don't sync node_modules
   ```

---

## 📊 Monitoring

### View Resource Usage

```bash
# Container stats
docker stats

# Specific container
docker stats liyali-backend
```

### Check Logs

```bash
# Real-time logs
docker-compose logs -f

# Search logs
docker-compose logs | grep ERROR

# Export logs
docker-compose logs > logs.txt
```

---

## 🔐 Security Notes

### For Development:

- ✅ Using simple passwords (fine for local dev)
- ✅ Debug logging enabled
- ✅ CORS set to localhost

### Before Production:

- ❌ Change all passwords
- ❌ Use strong JWT secrets
- ❌ Disable debug logging
- ❌ Update CORS settings
- ❌ Use environment-specific configs

---

## 🚀 Next Steps

### After Local Setup:

1. **Test Features**

   - Create requisitions
   - Test workflows
   - Try approvals
   - Check notifications

2. **Make Changes**

   - Edit code
   - Test locally
   - Commit changes

3. **Deploy to Cloud**
   - Push to GitHub
   - Deploy to Fly.io (staging)
   - Deploy to GCP (production)

---

## 📚 Useful Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Desktop Manual](https://docs.docker.com/desktop/)

---

## ✅ Quick Reference

### Start Development

```bash
# 1. Start everything
docker-compose up -d

# 2. Run migrations (first time only)
docker-compose exec backend sh -c "cd /app && make db-migrate"

# 3. Open browser
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
```

### Stop Development

```bash
# Stop containers (keep data)
docker-compose down

# Stop and remove data
docker-compose down -v
```

### Reset Everything

```bash
# Nuclear option - clean slate
docker-compose down -v
docker system prune -a
docker-compose up --build
```

---

**Estimated Setup Time**: 20 minutes  
**Status**: Ready for local development! 🎉

**Last Updated**: January 2026
