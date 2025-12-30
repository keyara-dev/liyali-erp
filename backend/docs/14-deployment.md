# Deployment Guide

Comprehensive production deployment guide for the Liyali Gateway Backend.

## Deployment Overview

The Liyali Gateway Backend can be deployed using several methods:

- **Docker Compose** (Recommended for small to medium deployments)
- **Kubernetes** (Recommended for large-scale deployments)
- **Traditional Server** (VM or bare metal)
- **Cloud Platforms** (AWS, GCP, Azure)

## Prerequisites

### System Requirements

**Minimum Production Requirements:**
- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Storage**: 20GB SSD
- **Network**: 1Gbps connection
- **OS**: Linux (Ubuntu 20.04+ recommended)

**Recommended Production Requirements:**
- **CPU**: 4+ cores
- **Memory**: 8GB+ RAM
- **Storage**: 50GB+ SSD
- **Network**: 10Gbps connection
- **OS**: Linux (Ubuntu 22.04 LTS)

### Software Dependencies

- **Docker** 20.10+ and Docker Compose 2.0+
- **PostgreSQL** 14+ (managed or self-hosted)
- **Reverse Proxy** (Nginx, Traefik, or cloud load balancer)
- **SSL Certificate** (Let's Encrypt or commercial)

## Environment Configuration

### Production Environment Variables

```env
# Application Configuration
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=warn

# Database Configuration
DB_HOST=your-production-db-host.com
DB_PORT=5432
DB_USER=liyali_prod_user
DB_PASSWORD=very-secure-production-password
DB_NAME=liyali_gateway_prod
DB_SSL_MODE=require

# Security Configuration
JWT_SECRET=production-jwt-secret-minimum-32-characters-very-secure
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# Frontend Configuration
FRONTEND_URL=https://app.company.com
CORS_ALLOWED_ORIGINS=https://app.company.com

# Performance Configuration
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h

# Security Features
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
MAX_LOGIN_ATTEMPTS=3
LOCKOUT_DURATION=30m

# Monitoring
METRICS_ENABLED=true
HEALTH_CHECK_ENABLED=true
```

### Environment-Specific Configurations

**Staging Environment:**
```env
APP_ENV=staging
DB_HOST=staging-db.company.com
DB_NAME=liyali_gateway_staging
FRONTEND_URL=https://staging.company.com
LOG_LEVEL=info
RATE_LIMIT_REQUESTS=200
```

**Production Environment:**
```env
APP_ENV=production
DB_HOST=prod-db.company.com
DB_NAME=liyali_gateway_prod
FRONTEND_URL=https://app.company.com
LOG_LEVEL=warn
RATE_LIMIT_REQUESTS=100
```

## Docker Deployment (Recommended)

### 1. Production Dockerfile

```dockerfile
# Multi-stage build for optimized production image
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy migration files
COPY --from=builder /app/database ./database

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
```

### 2. Production Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: liyali-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: liyali_gateway_prod
      POSTGRES_USER: liyali_prod_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    networks:
      - liyali-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U liyali_prod_user -d liyali_gateway_prod"]
      interval: 30s
      timeout: 10s
      retries: 3

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: liyali-backend
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=liyali_prod_user
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=liyali_gateway_prod
      - JWT_SECRET=${JWT_SECRET}
      - FRONTEND_URL=${FRONTEND_URL}
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - liyali-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    volumes:
      - ./logs:/app/logs

  nginx:
    image: nginx:alpine
    container_name: liyali-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - backend
    networks:
      - liyali-network

volumes:
  postgres_data:
    driver: local

networks:
  liyali-network:
    driver: bridge
```

### 3. Nginx Configuration

```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server backend:8080;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

    server {
        listen 80;
        server_name api.company.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name api.company.com;

        # SSL Configuration
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # Security Headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";

        # API Routes
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Health Check
        location /health {
            proxy_pass http://backend;
            access_log off;
        }
    }
}
```

### 4. Deployment Script

```bash
#!/bin/bash
# deploy.sh

set -e

echo "🚀 Starting Liyali Gateway Backend Deployment with Bootstrap System"

# Load environment variables
if [ -f .env.production ]; then
    export $(cat .env.production | xargs)
fi

# Validate required environment variables
required_vars=("DB_PASSWORD" "JWT_SECRET" "FRONTEND_URL")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Error: $var is not set"
        exit 1
    fi
done

# Build and deploy
echo "📦 Building Docker images..."
docker-compose -f docker-compose.prod.yml build

echo "🗄️ Starting database..."
docker-compose -f docker-compose.prod.yml up -d postgres

echo "⏳ Waiting for database to be ready..."
sleep 30

echo "🚀 Starting backend with bootstrap system..."
docker-compose -f docker-compose.prod.yml up -d backend

echo "🔄 Waiting for bootstrap to complete..."
sleep 15

# The bootstrap system automatically handles:
# - Database connection validation
# - Schema verification and migration
# - Integrity checks
# - Idempotent seeding (if enabled)

echo "🔍 Checking bootstrap completion..."
curl -f http://localhost:8080/health || exit 1

echo "🌐 Starting reverse proxy..."
docker-compose -f docker-compose.prod.yml up -d nginx

echo "✅ Deployment completed successfully!"
echo "🌐 API available at: https://api.company.com"
echo "📊 Health check: https://api.company.com/health"
echo "🔧 Bootstrap metrics: https://api.company.com/health/detailed"
```

## Kubernetes Deployment

### 1. Kubernetes Manifests

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: liyali-gateway
---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: liyali-config
  namespace: liyali-gateway
data:
  APP_ENV: "production"
  APP_PORT: "8080"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "liyali_gateway_prod"
  FRONTEND_URL: "https://app.company.com"
  LOG_LEVEL: "warn"
---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: liyali-secrets
  namespace: liyali-gateway
type: Opaque
stringData:
  DB_PASSWORD: "secure-production-password"
  JWT_SECRET: "production-jwt-secret-minimum-32-characters"
---
# k8s/postgres.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: liyali-gateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_DB
          value: liyali_gateway_prod
        - name: POSTGRES_USER
          value: liyali_prod_user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: liyali-secrets
              key: DB_PASSWORD
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
# k8s/backend.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: liyali-backend
  namespace: liyali-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: liyali-backend
  template:
    metadata:
      labels:
        app: liyali-backend
    spec:
      containers:
      - name: backend
        image: liyali-gateway-backend:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: liyali-config
        - secretRef:
            name: liyali-secrets
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: liyali-backend-service
  namespace: liyali-gateway
spec:
  selector:
    app: liyali-backend
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: liyali-ingress
  namespace: liyali-gateway
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - api.company.com
    secretName: liyali-tls
  rules:
  - host: api.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: liyali-backend-service
            port:
              number: 80
```

### 2. Kubernetes Deployment Script

```bash
#!/bin/bash
# k8s-deploy.sh

set -e

echo "🚀 Deploying to Kubernetes with Bootstrap System"

# Apply manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/postgres.yaml

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n liyali-gateway --timeout=300s

# Deploy backend with bootstrap system
kubectl apply -f k8s/backend.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# Wait for bootstrap to complete
echo "🔄 Waiting for bootstrap to complete..."
kubectl rollout status deployment/liyali-backend -n liyali-gateway

# Verify bootstrap completion
echo "✅ Verifying bootstrap completion..."
kubectl exec -n liyali-gateway deployment/liyali-backend -- curl -f http://localhost:8080/health

echo "✅ Kubernetes deployment with bootstrap completed!"
```

## Database Setup

### 1. Production Database Configuration

```sql
-- Create production database and user
CREATE DATABASE liyali_gateway_prod;
CREATE USER liyali_prod_user WITH PASSWORD 'secure-production-password';
GRANT ALL PRIVILEGES ON DATABASE liyali_gateway_prod TO liyali_prod_user;

-- Configure connection limits
ALTER USER liyali_prod_user CONNECTION LIMIT 100;

-- Enable required extensions
\c liyali_gateway_prod
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
```

### 2. Bootstrap System Integration

The backend uses a production-ready bootstrap system that handles database initialization with proper phase ordering and error recovery:

**Bootstrap Phases:**
1. **Connect** - Validates database connection health
2. **Validate** - Ensures database accessibility and schema readiness
3. **Migrate** - Verifies all required tables exist and structures are correct
4. **Verify** - Comprehensive schema integrity and constraint validation
5. **Seed** - Idempotent data seeding with UPSERT operations

**Production Bootstrap Configuration:**
```env
# Bootstrap system settings
ENABLE_SEEDING=false          # Disable seeding in production
BOOTSTRAP_TIMEOUT=300s        # 5 minute timeout for bootstrap
VALIDATION_TIMEOUT=60s        # 1 minute for validation phase
MIGRATION_TIMEOUT=600s        # 10 minutes for migration phase
CIRCUIT_BREAKER_ENABLED=true  # Enable circuit breaker protection
```

### 3. Run Migrations with Bootstrap

```bash
# Production migration script with bootstrap
#!/bin/bash

DB_URL="postgres://liyali_prod_user:secure-password@prod-db.company.com/liyali_gateway_prod?sslmode=require"

echo "🚀 Starting database bootstrap process..."

# The application automatically handles bootstrap on startup
# No manual migration commands needed - bootstrap system handles:
# - Connection validation
# - Schema verification
# - Migration execution
# - Integrity checks
# - Idempotent seeding (if enabled)

# For manual migration (if needed):
cd database && ./migrate.sh up

echo "✅ Database bootstrap completed!"
```

### 4. Bootstrap Health Checks

The bootstrap system provides health check endpoints for deployment orchestration:

```bash
# Check bootstrap status
curl http://localhost:8080/health

# Detailed bootstrap metrics
curl http://localhost:8080/health/detailed
```

**Health Check Response:**
```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "status": "healthy",
      "bootstrap_completed": true,
      "last_bootstrap": "2024-01-01T10:00:00Z",
      "bootstrap_duration": "2.3s"
    }
  }
}
```

### 3. Database Backup Strategy

```bash
#!/bin/bash
# backup.sh

DB_URL="postgres://liyali_prod_user:password@prod-db.company.com/liyali_gateway_prod"
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
pg_dump $DB_URL > $BACKUP_DIR/liyali_gateway_$DATE.sql

# Compress backup
gzip $BACKUP_DIR/liyali_gateway_$DATE.sql

# Remove backups older than 30 days
find $BACKUP_DIR -name "liyali_gateway_*.sql.gz" -mtime +30 -delete

echo "✅ Backup completed: liyali_gateway_$DATE.sql.gz"
```

## Monitoring and Logging

### 1. Production Logging

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  filebeat:
    image: docker.elastic.co/beats/filebeat:8.5.0
    volumes:
      - ./filebeat.yml:/usr/share/filebeat/filebeat.yml
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    depends_on:
      - backend
```

### 2. Health Monitoring

```bash
#!/bin/bash
# health-check.sh

API_URL="https://api.company.com"

# Check API health
response=$(curl -s -o /dev/null -w "%{http_code}" $API_URL/health)

if [ $response -eq 200 ]; then
    echo "✅ API is healthy"
    exit 0
else
    echo "❌ API is unhealthy (HTTP $response)"
    exit 1
fi
```

## Security Considerations

### 1. SSL/TLS Configuration

```bash
# Generate SSL certificate with Let's Encrypt
certbot certonly --webroot -w /var/www/html -d api.company.com

# Auto-renewal
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
```

### 2. Firewall Configuration

```bash
# UFW firewall rules
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw deny 8080/tcp   # Block direct backend access
ufw enable
```

### 3. Security Headers

Already configured in the Nginx configuration above.

## Performance Optimization

### 1. Database Optimization

```sql
-- Create indexes for better performance
CREATE INDEX CONCURRENTLY idx_requisitions_organization_id ON requisitions(organization_id);
CREATE INDEX CONCURRENTLY idx_requisitions_status ON requisitions(status);
CREATE INDEX CONCURRENTLY idx_requisitions_created_at ON requisitions(created_at);
CREATE INDEX CONCURRENTLY idx_documents_type_org ON documents(document_type, organization_id);
CREATE INDEX CONCURRENTLY idx_documents_search ON documents USING gin(to_tsvector('english', title || ' ' || description));

-- Analyze tables
ANALYZE requisitions;
ANALYZE documents;
ANALYZE users;
```

### 2. Application Optimization

```env
# Production performance settings
DB_MAX_IDLE_CONNS=25
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h

# Enable connection pooling
DB_POOL_ENABLED=true

# Cache settings
CACHE_ENABLED=true
CACHE_TTL=300
```

## Troubleshooting Deployment

### Common Issues

**Container Won't Start:**
```bash
# Check container logs
docker logs liyali-backend

# Check container status
docker ps -a
```

**Database Connection Issues:**
```bash
# Test database connection
docker exec -it liyali-postgres psql -U liyali_prod_user -d liyali_gateway_prod

# Check database logs
docker logs liyali-postgres
```

**SSL Certificate Issues:**
```bash
# Check certificate validity
openssl x509 -in /etc/nginx/ssl/cert.pem -text -noout

# Renew certificate
certbot renew --force-renewal
```

## Rollback Procedures

### 1. Application Rollback

```bash
#!/bin/bash
# rollback.sh

PREVIOUS_VERSION="v1.0.0"

echo "🔄 Rolling back to version $PREVIOUS_VERSION"

# Pull previous image
docker pull liyali-gateway-backend:$PREVIOUS_VERSION

# Update docker-compose to use previous version
sed -i "s/image: liyali-gateway-backend:latest/image: liyali-gateway-backend:$PREVIOUS_VERSION/" docker-compose.prod.yml

# Restart services
docker-compose -f docker-compose.prod.yml up -d backend

echo "✅ Rollback completed!"
```

### 2. Database Rollback

```bash
#!/bin/bash
# db-rollback.sh

BACKUP_FILE="/backups/liyali_gateway_20240101_120000.sql.gz"

echo "🔄 Rolling back database..."

# Stop application
docker-compose -f docker-compose.prod.yml stop backend

# Restore database
gunzip -c $BACKUP_FILE | psql $DB_URL

# Start application
docker-compose -f docker-compose.prod.yml start backend

echo "✅ Database rollback completed!"
```

## Maintenance

### 1. Regular Maintenance Tasks

```bash
#!/bin/bash
# maintenance.sh

echo "🔧 Running maintenance tasks..."

# Update system packages
apt update && apt upgrade -y

# Clean Docker images
docker system prune -f

# Vacuum database
psql $DB_URL -c "VACUUM ANALYZE;"

# Rotate logs
logrotate /etc/logrotate.conf

echo "✅ Maintenance completed!"
```

### 2. Monitoring Script

```bash
#!/bin/bash
# monitor.sh

# Check disk space
df -h | awk '$5 > 80 {print "⚠️  Disk usage high: " $0}'

# Check memory usage
free -m | awk 'NR==2{printf "Memory Usage: %s/%sMB (%.2f%%)\n", $3,$2,$3*100/$2 }'

# Check API health
curl -f https://api.company.com/health || echo "❌ API health check failed"

# Check database connections
psql $DB_URL -c "SELECT count(*) as active_connections FROM pg_stat_activity;" | tail -n 1
```

## Next Steps

After successful deployment:

1. **Set up monitoring** - Configure [Monitoring & Observability](./15-monitoring.md)
2. **Configure backups** - Set up automated database backups
3. **Set up CI/CD** - Automate deployments with GitHub Actions or similar
4. **Performance testing** - Run load tests against production environment
5. **Security audit** - Perform security assessment and penetration testing

## Support

For deployment issues:
- Check [Troubleshooting Guide](./16-troubleshooting.md)
- Review application logs
- Verify environment configuration
- Test database connectivity

The deployment is now complete! Your Liyali Gateway Backend is ready for production use.