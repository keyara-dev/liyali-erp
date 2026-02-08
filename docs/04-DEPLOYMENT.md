# Deployment Guide

## Pre-Deployment Checklist

- [ ] All tests passing
- [ ] Environment variables configured
- [ ] Database migrations ready
- [ ] SSL certificates obtained
- [ ] Domain DNS configured
- [ ] Monitoring setup
- [ ] Backup strategy in place

## Environment Setup

### Production Environment Variables

**Backend:**

```env
DATABASE_URL=postgresql://user:pass@host:5432/liyali_gateway
JWT_SECRET=<strong-secret>
PORT=8080
ENVIRONMENT=production
CORS_ORIGINS=https://liyali.com,https://admin.liyali.com
```

**Frontend:**

```env
NEXT_PUBLIC_API_URL=https://api.liyali.com
NEXT_PUBLIC_APP_URL=https://liyali.com
```

**Admin Console:**

```env
NEXT_PUBLIC_API_URL=https://api.liyali.com
```

## Deployment Options

### Option 1: Docker Compose

```bash
# Build images
docker-compose build

# Deploy
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### Option 2: Fly.io

See [Fly.io Deployment Guide](./05-FLY-IO.md)

### Option 3: Manual Deployment

#### Backend

```bash
cd backend
go build -o liyali-backend
./liyali-backend
```

#### Frontend

```bash
cd frontend
npm run build
npm start
```

#### Admin Console

```bash
cd admin-console
npm run build
npm start
```

## Database Migration

### Production Migration

```bash
# Backup database first
pg_dump liyali_gateway > backup.sql

# Run migrations
cd backend
make migrate

# Verify
psql liyali_gateway -c "SELECT * FROM schema_migrations;"
```

### Rollback Plan

```bash
# Restore from backup
psql liyali_gateway < backup.sql
```

## SSL/TLS Setup

### Using Let's Encrypt

```bash
# Install certbot
sudo apt-get install certbot

# Get certificate
sudo certbot certonly --standalone -d liyali.com -d api.liyali.com

# Auto-renewal
sudo certbot renew --dry-run
```

## Reverse Proxy (Nginx)

```nginx
# /etc/nginx/sites-available/liyali

# Backend API
server {
    listen 443 ssl;
    server_name api.liyali.com;

    ssl_certificate /etc/letsencrypt/live/api.liyali.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.liyali.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Frontend
server {
    listen 443 ssl;
    server_name liyali.com;

    ssl_certificate /etc/letsencrypt/live/liyali.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/liyali.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Admin Console
server {
    listen 443 ssl;
    server_name admin.liyali.com;

    ssl_certificate /etc/letsencrypt/live/admin.liyali.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.liyali.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Monitoring

### Health Checks

- Backend: `https://api.liyali.com/health`
- Frontend: `https://liyali.com`
- Admin: `https://admin.liyali.com`

### Logging

```bash
# Backend logs
tail -f /var/log/liyali/backend.log

# Frontend logs
tail -f /var/log/liyali/frontend.log

# Nginx logs
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
```

### Monitoring Tools

- Uptime monitoring: UptimeRobot, Pingdom
- Error tracking: Sentry
- Performance: New Relic, DataDog
- Logs: Papertrail, Loggly

## Backup Strategy

### Database Backups

```bash
# Daily backup script
#!/bin/bash
DATE=$(date +%Y%m%d)
pg_dump liyali_gateway | gzip > /backups/db_$DATE.sql.gz

# Keep last 30 days
find /backups -name "db_*.sql.gz" -mtime +30 -delete
```

### File Backups

```bash
# Backup uploads
tar -czf /backups/uploads_$DATE.tar.gz /var/www/uploads
```

## Scaling

### Horizontal Scaling

- Use load balancer (Nginx, HAProxy)
- Deploy multiple backend instances
- Use Redis for session storage
- Implement database read replicas

### Vertical Scaling

- Increase server resources
- Optimize database queries
- Implement caching
- Use CDN for static assets

## Security

- [ ] HTTPS enabled
- [ ] Firewall configured
- [ ] Database access restricted
- [ ] Secrets in environment variables
- [ ] Rate limiting enabled
- [ ] CORS configured
- [ ] Security headers set
- [ ] Regular security updates

## Troubleshooting

### Service Not Starting

```bash
# Check logs
journalctl -u liyali-backend -f

# Check ports
netstat -tulpn | grep :8080

# Check process
ps aux | grep liyali
```

### Database Connection Issues

```bash
# Test connection
psql -h host -U user -d liyali_gateway

# Check connections
SELECT * FROM pg_stat_activity;
```

### High Memory Usage

```bash
# Check memory
free -h

# Check processes
top

# Restart services
systemctl restart liyali-backend
```

## Rollback Procedure

1. Stop services
2. Restore database backup
3. Deploy previous version
4. Restart services
5. Verify functionality

## Post-Deployment

- [ ] Verify all services running
- [ ] Test critical functionality
- [ ] Check error logs
- [ ] Monitor performance
- [ ] Update documentation
- [ ] Notify team

## Resources

- [Fly.io Guide](./05-FLY-IO.md)
- [Recovery Guide](./06-RECOVERY.md)
- [Monitoring Setup](../backend/docs/15-monitoring.md)
