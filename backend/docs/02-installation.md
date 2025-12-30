# Installation Guide

Detailed installation instructions for the Liyali Gateway Backend.

## System Requirements

### Minimum Requirements
- **OS**: Linux, macOS, or Windows
- **Go**: Version 1.21 or higher
- **PostgreSQL**: Version 14 or higher
- **Memory**: 2GB RAM minimum
- **Storage**: 10GB available space

### Recommended Requirements
- **OS**: Ubuntu 20.04+ or macOS 12+
- **Go**: Version 1.21+
- **PostgreSQL**: Version 15+
- **Memory**: 4GB RAM
- **Storage**: 20GB available space

## Installation Methods

### Method 1: Development Setup (Recommended)

#### 1. Install Go

**Ubuntu/Debian:**
```bash
# Remove old Go installation
sudo rm -rf /usr/local/go

# Download and install Go 1.21
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

**macOS:**
```bash
# Using Homebrew
brew install go

# Or download from https://golang.org/dl/
```

**Windows:**
```bash
# Download installer from https://golang.org/dl/
# Run the installer and follow instructions
```

#### 2. Install PostgreSQL

**Ubuntu/Debian:**
```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# Start PostgreSQL service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database user
sudo -u postgres createuser --interactive
```

**macOS:**
```bash
# Using Homebrew
brew install postgresql
brew services start postgresql

# Create database
createdb liyali_gateway
```

**Windows:**
```bash
# Download installer from https://www.postgresql.org/download/windows/
# Run installer and follow instructions
```

#### 3. Clone Repository

```bash
# Clone the repository
git clone <repository-url>
cd liyali-gateway/backend

# Verify Go modules
go mod download
go mod verify
```

### Method 2: Docker Setup

#### 1. Create Docker Compose File

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: liyali_gateway
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/migrations:/docker-entrypoint-initdb.d

  backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: password
      DB_NAME: liyali_gateway
      JWT_SECRET: your-secret-key
    depends_on:
      - postgres
    volumes:
      - .:/app

volumes:
  postgres_data:
```

#### 2. Create Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

EXPOSE 8080
CMD ["./main"]
```

#### 3. Run with Docker

```bash
# Build and start services
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f backend
```

## Database Setup

### 1. Create Database

```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Create database and user
CREATE DATABASE liyali_gateway;
CREATE USER liyali_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE liyali_gateway TO liyali_user;

# Exit PostgreSQL
\q
```

### 2. Run Migrations with Bootstrap System

The application now uses an advanced bootstrap system that handles database initialization automatically:

#### Option A: Using Migration Scripts (Recommended)

```bash
# Navigate to backend directory
cd backend

# Linux/Mac - Run complete migration
cd database && ./migrate.sh up

# Windows - Run complete migration  
cd database && migrate.bat up

# Reset database (drops and recreates everything)
cd database && ./migrate.sh reset
```

#### Option B: Manual Migration

```bash
# Run the comprehensive schema migration
go run database/run_migration.go database/migrations/001_create_complete_schema.up.sql
```

### 3. Bootstrap System Features

The new bootstrap system provides:

- **Phase-Ordered Initialization**: Connect → Validate → Migrate → Verify → Seed
- **Idempotent Operations**: Safe to run multiple times using PostgreSQL UPSERT
- **Circuit Breaker Protection**: Prevents cascading failures during startup
- **Comprehensive Validation**: Schema integrity and constraint verification
- **Production Observability**: Detailed logging and metrics collection

### 4. Verify Database Setup

The bootstrap system automatically verifies:

```bash
# Start the application - bootstrap runs automatically
go run main.go
```

You'll see structured bootstrap logging:
```
[BOOTSTRAP] 🚀 Starting database bootstrap process (env: development)
[BOOTSTRAP] ✅ Phase: connect - Completed in 45ms
[BOOTSTRAP] ✅ Phase: validate - Completed in 123ms
[BOOTSTRAP] ✅ Phase: migrate - Completed in 67ms
[BOOTSTRAP] ✅ Phase: verify - Completed in 89ms
[BOOTSTRAP] 🌱 Seeding users: 4 created, 0 updated, 0 skipped (took 67ms)
[BOOTSTRAP] ✅ Database bootstrap completed successfully in 2.3s
```

### 5. Manual Verification (Optional)

```bash
# Connect to database
psql -d liyali_gateway

# List tables (should show all required tables)
\dt

# Check seeded data
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM organizations;
SELECT COUNT(*) FROM vendors;

# Exit
\q
```

## Configuration

### 1. Environment Variables

```bash
# Copy template
cp .env.example .env

# Edit configuration
nano .env
```

### 2. Required Configuration

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=liyali_user
DB_PASSWORD=secure_password
DB_NAME=liyali_gateway
DB_SSL_MODE=disable

# Application Configuration
APP_PORT=8080
APP_ENV=development

# Security Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars

# Frontend Configuration
FRONTEND_URL=http://localhost:3000

# Optional Configuration
LOG_LEVEL=info
ENABLE_CORS=true

# Bootstrap System Configuration
ENABLE_SEEDING=true          # Enable automatic seeding in development
BOOTSTRAP_TIMEOUT=300        # Bootstrap timeout in seconds
CIRCUIT_BREAKER_ENABLED=true # Enable circuit breaker protection
```

### 3. Production Configuration

```env
# Production Database
DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=production_user
DB_PASSWORD=very-secure-production-password
DB_NAME=liyali_gateway_prod
DB_SSL_MODE=require

# Production Application
APP_PORT=8080
APP_ENV=production

# Production Security
JWT_SECRET=very-long-random-secret-key-for-production-use-at-least-32-characters

# Production Frontend
FRONTEND_URL=https://your-production-domain.com

# Production Logging
LOG_LEVEL=warn
ENABLE_CORS=false

# Production Bootstrap Configuration
ENABLE_SEEDING=false         # Disable automatic seeding in production
BOOTSTRAP_TIMEOUT=600        # Longer timeout for production
CIRCUIT_BREAKER_ENABLED=true # Always enable circuit breaker in production
```

## Build and Run

### Development Mode

```bash
# Install air for hot reloading (optional)
go install github.com/cosmtrek/air@latest

# Run with hot reloading
air

# Or run directly
go run main.go
```

### Production Build

```bash
# Build binary
go build -o liyali-gateway-backend .

# Run binary
./liyali-gateway-backend

# Or build with optimizations
go build -ldflags="-w -s" -o liyali-gateway-backend .
```

### Cross-Platform Build

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o liyali-gateway-linux .

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o liyali-gateway-windows.exe .

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o liyali-gateway-macos .
```

## Verification

### 1. Health Check

```bash
# Start the application
go run main.go

# In another terminal, test health endpoint
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "liyali-gateway-backend"
}
```

### 2. Database Connection Test

```bash
# Test database connectivity
curl http://localhost:8080/api/v1/organizations
```

### 3. Feature Test

```bash
# Test document search (should return empty array initially)
curl -X GET "http://localhost:8080/api/v1/documents/search?q=test"
```

## Troubleshooting Installation

### Go Installation Issues

```bash
# Check Go installation
go version
go env GOPATH
go env GOROOT

# Fix PATH issues
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### PostgreSQL Issues

```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*.log

# Test connection
psql -h localhost -U postgres -d liyali_gateway
```

### Permission Issues

```bash
# Fix PostgreSQL permissions
sudo -u postgres psql
ALTER USER liyali_user CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE liyali_gateway TO liyali_user;
```

### Port Issues

```bash
# Check if port is in use
lsof -i :8080

# Kill process using port
lsof -ti:8080 | xargs kill -9

# Use different port
export APP_PORT=8081
```

### Module Issues

```bash
# Clean module cache
go clean -modcache

# Re-download modules
go mod download

# Verify modules
go mod verify

# Update modules
go mod tidy
```

## Next Steps

After successful installation:

1. **Configuration**: Review [Configuration Guide](./03-configuration.md)
2. **Development**: Set up [Development Environment](./11-development.md)
3. **Testing**: Run [Test Suite](./12-testing.md)
4. **API**: Explore [API Reference](./13-api-reference.md)

## Production Deployment

For production deployment, see:
- [Deployment Guide](./14-deployment.md)
- [Monitoring Setup](./15-monitoring.md)
- [Security Considerations](./07-auth.md)

The installation is now complete! Your Liyali Gateway Backend is ready for development or production use.