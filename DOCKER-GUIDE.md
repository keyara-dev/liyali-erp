# Docker Guide - Liyali Gateway

## Overview

This guide explains how to build and run the Liyali Gateway application using Docker and Docker Compose.

## Prerequisites

- Docker 20.10 or later
- Docker Compose 2.0 or later
- Git

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/liyali/liyali-gateway.git
cd liyali-gateway
```

### 2. Set Environment Variables

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```bash
# Server
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info

# Database
DB_USER=postgres
DB_PASSWORD=00110011
DB_NAME=liyali-dev-db
DB_PORT=5432

# JWT
JWT_SECRET=your-secret-key-here
JWT_EXPIRATION=24h

# CORS
CORS_ORIGINS=http://localhost:3000,http://localhost:8080

# Redis
REDIS_PORT=6379

# PgAdmin
PGADMIN_EMAIL=admin@liyali.com
PGADMIN_PASSWORD=admin
```

### 3. Start Docker Containers

```bash
# Start all services
docker-compose up -d

# Or with logs
docker-compose up
```

### 4. Verify Services

Check that all services are running:

```bash
docker-compose ps
```

Expected output:
```
NAME              STATUS              PORTS
liyali-postgres   Up 30 seconds       5432/tcp
liyali-redis      Up 30 seconds       6379/tcp
liyali-backend    Up 20 seconds       0.0.0.0:8080->8080/tcp
liyali-pgadmin    Up 10 seconds       0.0.0.0:5050->80/tcp
```

### 5. Test the Application

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","message":"Liyali Gateway Backend API is running"}
```

## Services

### PostgreSQL Database
- **Image:** postgres:15-alpine
- **Port:** 5432
- **Container:** liyali-postgres
- **Default Credentials:**
  - Username: postgres
  - Password: 00110011 (can be overridden with DB_PASSWORD)
  - Database: liyali-dev-db

### Redis Cache
- **Image:** redis:7-alpine
- **Port:** 6379
- **Container:** liyali-redis
- **Features:** Optional caching layer for future use

### Liyali Backend API
- **Port:** 8080
- **Container:** liyali-backend
- **Environment:** Development (hot reload compatible)
- **Health Check:** GET /health endpoint

### PgAdmin (Development Only)
- **Port:** 5050
- **Container:** liyali-pgadmin
- **URL:** http://localhost:5050
- **Credentials:**
  - Email: admin@liyali.com (or PGADMIN_EMAIL)
  - Password: admin (or PGADMIN_PASSWORD)

## Common Commands

### Start Services

```bash
# Start all services in background
docker-compose up -d

# Start specific service
docker-compose up -d postgres
docker-compose up -d backend

# Start with rebuild
docker-compose up -d --build
```

### Stop Services

```bash
# Stop all services
docker-compose down

# Stop specific service
docker-compose stop backend

# Remove volumes and containers
docker-compose down -v
```

### View Logs

```bash
# View logs from all services
docker-compose logs

# View logs from specific service
docker-compose logs backend
docker-compose logs postgres

# Follow logs (real-time)
docker-compose logs -f backend

# View last 100 lines
docker-compose logs --tail=100 backend
```

### Execute Commands

```bash
# Execute command in backend container
docker-compose exec backend go test ./...

# Run bash in backend container
docker-compose exec backend /bin/sh

# Execute command in postgres container
docker-compose exec postgres psql -U postgres -d liyali-dev-db
```

### Database Management

```bash
# Connect to PostgreSQL using psql
docker-compose exec postgres psql -U postgres -d liyali-dev-db

# Backup database
docker-compose exec postgres pg_dump -U postgres liyali-dev-db > backup.sql

# Restore database
docker-compose exec -T postgres psql -U postgres liyali-dev-db < backup.sql

# Access PgAdmin
# Open http://localhost:5050 in browser
```

### Rebuild and Restart

```bash
# Rebuild backend image
docker-compose build backend

# Rebuild and restart backend
docker-compose up -d --build backend

# Rebuild all images
docker-compose build

# Full restart with fresh build
docker-compose down -v && docker-compose up -d --build
```

## Development Workflow

### Hot Reload (Development)

The backend service mounts the source code as a volume, allowing hot reload:

```bash
# Changes to Go files trigger automatic restart (if using air or similar)
nano backend/main.go
# Save changes - backend will automatically restart
```

### Running Tests

```bash
# Run all tests
docker-compose exec backend make test

# Run unit tests
docker-compose exec backend make test-unit

# Run integration tests
docker-compose exec backend make test-integration

# Run with coverage
docker-compose exec backend make test-coverage
```

### Building Binary

```bash
# Build binary inside container
docker-compose exec backend make build

# Binary will be at: backend/bin/liyali-gateway
```

## Production Deployment

### Environment Setup for Production

Create `.env.production`:

```bash
# Server
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=warn

# Database (use production database)
DB_HOST=prod-db.example.com
DB_USER=prod_user
DB_PASSWORD=secure_password_here
DB_NAME=liyali_prod_db
DB_SSL_MODE=require

# JWT
JWT_SECRET=your-production-secret-key-change-this

# CORS
CORS_ORIGINS=https://app.liyali.com

# Redis
REDIS_HOST=prod-redis.example.com
REDIS_PASSWORD=redis_password
```

### Building Production Image

```bash
# Build optimized image
docker build -t liyali-gateway:1.0.0 .

# Tag for registry
docker tag liyali-gateway:1.0.0 registry.example.com/liyali-gateway:1.0.0

# Push to registry
docker push registry.example.com/liyali-gateway:1.0.0
```

### Running in Production

```bash
# Use docker-compose.prod.yml (without pgadmin, redis volumes)
docker-compose -f docker-compose.prod.yml up -d

# Or use Kubernetes/Docker Swarm for orchestration
```

## Troubleshooting

### Container Not Starting

```bash
# Check logs
docker-compose logs backend

# Rebuild image
docker-compose build --no-cache backend

# Restart container
docker-compose restart backend
```

### Database Connection Issues

```bash
# Check database is running
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U postgres -c "SELECT 1"

# Check network connectivity
docker-compose exec backend ping postgres
```

### Port Already in Use

```bash
# Change port in .env
PORT=9000

# Or find process using port 8080
lsof -i :8080
# Kill process
kill -9 <PID>
```

### Rebuild Everything from Scratch

```bash
# Remove all containers and volumes
docker-compose down -v

# Remove images
docker-compose rm

# Remove dangling images
docker image prune -f

# Start fresh
docker-compose up -d --build
```

### View Resource Usage

```bash
# Check container stats
docker stats

# Check disk usage
docker system df

# Clean up unused resources
docker system prune -a
```

## Health Checks

All services include health checks:

```bash
# Check health status
docker-compose ps

# Manually test endpoint
curl http://localhost:8080/health

# Check PostgreSQL health
docker-compose exec postgres pg_isready
```

## Networking

Services communicate via the `liyali-network` bridge network:

- **Backend → PostgreSQL:** `postgres:5432`
- **Backend → Redis:** `redis:6379`
- **PgAdmin → PostgreSQL:** `postgres:5432`

### Custom Network Configuration

Edit `docker-compose.yml` to modify network settings:

```yaml
networks:
  liyali-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## Volumes

Three volumes manage data persistence:

- **postgres_data:** PostgreSQL database files
- **redis_data:** Redis data
- **pgadmin_data:** PgAdmin configuration

### Backup Volumes

```bash
# Backup PostgreSQL volume
docker run --rm -v postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres_backup.tar.gz -C /data .

# Restore volume
docker run --rm -v postgres_data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/postgres_backup.tar.gz -C /data
```

## Security Considerations

1. **Change Default Passwords:**
   - Always change `DB_PASSWORD` and `PGADMIN_PASSWORD`
   - Use strong, random passwords

2. **JWT Secret:**
   - Change `JWT_SECRET` to a unique, long string
   - Store in secure environment (not git)

3. **CORS Origins:**
   - Only allow trusted origins in `CORS_ORIGINS`
   - Use HTTPS in production

4. **Database Access:**
   - Don't expose database port 5432 to internet
   - Use VPC or private networks
   - Enable SSL/TLS for connections

5. **Container Security:**
   - Run containers as non-root user (already configured)
   - Use read-only filesystem where possible
   - Scan images for vulnerabilities

## References

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go Fiber Framework](https://gofiber.io/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
