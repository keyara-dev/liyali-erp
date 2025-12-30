# Environment Setup Guide

**Date:** December 28, 2025  
**Status:** ✅ COMPLETE  
**Priority:** DEVELOPMENT SETUP

## Overview

This guide covers the environment configuration setup for the Liyali Gateway application, including both backend and root-level environment variables.

## Files Created

### 1. **Backend Environment** (`backend/.env`)
- Database configuration for PostgreSQL
- JWT secret for authentication
- CORS settings for frontend integration
- Application port and environment settings

### 2. **Root Environment** (`.env`)
- Comprehensive configuration for all application components
- Feature flags for enabling/disabling functionality
- Approval workflow thresholds
- Security and logging configuration

## Environment Variables Explained

### 🗄️ **Database Configuration**
```bash
DB_HOST=localhost          # Database host (localhost for local dev)
DB_PORT=5432              # PostgreSQL default port
DB_USER=postgres          # Database username
DB_PASSWORD=00110011      # Database password
DB_NAME=liyali-dev-db     # Database name
DB_SSL_MODE=disable       # SSL mode (disabled for local dev)
```

### 🔐 **Authentication & Security**
```bash
JWT_SECRET=liyali-gateway-jwt-secret-key-development-2025-secure
JWT_EXPIRATION=24h        # Token expiration time
```

### 🌐 **CORS & API Configuration**
```bash
CORS_ORIGINS=http://localhost:3000,http://localhost:8080,http://localhost:5173
API_VERSION=v1
RATE_LIMIT=100           # Requests per minute
```

### 🚀 **Feature Flags**
```bash
ENABLE_NOTIFICATIONS=true      # Enable notification system
ENABLE_AUDIT_LOG=true         # Enable audit logging
ENABLE_BUDGET_CONSTRAINTS=true # Enable budget validation
```

### 💰 **Approval Workflow Thresholds**
```bash
LOW_AMOUNT_THRESHOLD=1000000    # $10,000 (in cents)
MEDIUM_AMOUNT_THRESHOLD=5000000 # $50,000 (in cents)
HIGH_AMOUNT_THRESHOLD=10000000  # $100,000 (in cents)
```

### 📊 **Budget Constraints**
```bash
VENDOR_SPENDING_LIMIT_PERCENT=30  # Max 30% to single vendor
RESERVE_FUNDS_PERCENT=10          # Keep 10% in reserve
```

## Prerequisites

### 1. **PostgreSQL Database**
Ensure PostgreSQL is installed and running:
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Create database (if needed)
createdb liyali-dev-db
```

### 2. **Redis (Optional)**
For caching and session management:
```bash
# Check if Redis is running
redis-cli ping
```

## Development Setup Steps

### 1. **Database Setup**
```bash
# Navigate to backend directory
cd backend

# Run database migrations (if available)
make migrate-up

# Or manually create tables using SQL files
psql -h localhost -U postgres -d liyali-dev-db -f database/migrations/001_initial.sql
```

### 2. **Backend Setup**
```bash
# Navigate to backend directory
cd backend

# Install dependencies
go mod download

# Run the application
go run main.go
```

### 3. **Frontend Setup**
```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

## Environment-Specific Configurations

### 🔧 **Development Environment**
- **Database:** Local PostgreSQL instance
- **CORS:** Allows localhost origins
- **Logging:** Debug level enabled
- **HTTPS:** Disabled for local development
- **Backups:** Disabled

### 🚀 **Production Environment**
For production deployment, update these values:
```bash
ENVIRONMENT=production
LOG_LEVEL=info
DB_SSL_MODE=require
ENABLE_HTTPS=true
JWT_SECRET=<strong-production-secret>
CORS_ORIGINS=https://yourdomain.com
```

## Security Considerations

### 🔒 **JWT Secret**
- **Development:** Uses a long, descriptive key
- **Production:** Must use a cryptographically secure random key
- **Rotation:** Should be rotated periodically in production

### 🛡️ **Database Security**
- **Development:** SSL disabled for simplicity
- **Production:** SSL should be enabled (`DB_SSL_MODE=require`)
- **Credentials:** Use environment-specific credentials

### 🌐 **CORS Configuration**
- **Development:** Allows multiple localhost ports
- **Production:** Should only allow your production domain(s)

## Troubleshooting

### Common Issues

#### 1. **Database Connection Failed**
```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql

# Check if database exists
psql -h localhost -U postgres -l | grep liyali-dev-db
```

#### 2. **JWT Token Issues**
- Ensure `JWT_SECRET` is set and consistent between restarts
- Check token expiration settings

#### 3. **CORS Errors**
- Verify `CORS_ORIGINS` includes your frontend URL
- Check that frontend is running on the specified port

#### 4. **Port Conflicts**
```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill process using the port (if needed)
kill -9 <PID>
```

## Environment Validation

### Backend Validation
```bash
cd backend
go run main.go
# Should see: "Server starting on port 8080"
```

### Frontend Validation
```bash
cd frontend
npm run dev
# Should see: "Local: http://localhost:3000"
```

### API Health Check
```bash
curl http://localhost:8080/health
# Should return: {"status": "ok"}
```

## Next Steps

1. **✅ Environment files created and configured**
2. **🔄 Start PostgreSQL database**
3. **🚀 Run backend application**
4. **🌐 Start frontend development server**
5. **🧪 Test registration and login functionality**

## File Security

### ⚠️ **Important Notes**
- **Never commit `.env` files to version control**
- **Use `.env.example` files as templates**
- **Rotate secrets regularly in production**
- **Use different secrets for different environments**

The environment is now properly configured for development with secure defaults and comprehensive feature flags.