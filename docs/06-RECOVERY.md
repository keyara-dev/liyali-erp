# Disaster Recovery Guide

## Overview

Procedures for recovering from system failures and data loss.

## Backup Strategy

### Database Backups

**Automated Daily Backups:**

```bash
# Backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump liyali_gateway | gzip > /backups/db_$DATE.sql.gz

# Retention: 30 days
find /backups -name "db_*.sql.gz" -mtime +30 -delete
```

**Manual Backup:**

```bash
pg_dump liyali_gateway > backup.sql
```

### File Backups

```bash
# Uploads
tar -czf uploads_backup.tar.gz /var/www/uploads

# Configuration
tar -czf config_backup.tar.gz /etc/liyali
```

## Recovery Procedures

### Database Recovery

#### Full Database Restore

```bash
# Stop services
systemctl stop liyali-backend

# Drop existing database
dropdb liyali_gateway

# Create new database
createdb liyali_gateway

# Restore from backup
gunzip < /backups/db_20260208.sql.gz | psql liyali_gateway

# Restart services
systemctl start liyali-backend
```

#### Point-in-Time Recovery

```bash
# Restore to specific time
pg_restore --dbname=liyali_gateway \
  --clean \
  --if-exists \
  backup.dump
```

### Application Recovery

#### Backend Recovery

```bash
# Pull latest stable version
git checkout main
git pull

# Rebuild
cd backend
go build -o liyali-backend

# Restart
systemctl restart liyali-backend
```

#### Frontend Recovery

```bash
# Pull latest stable version
git checkout main
git pull

# Rebuild
cd frontend
npm run build

# Restart
pm2 restart liyali-frontend
```

### Service Recovery

#### Check Service Status

```bash
# Backend
systemctl status liyali-backend

# Database
systemctl status postgresql

# Nginx
systemctl status nginx
```

#### Restart Services

```bash
# Restart all services
systemctl restart liyali-backend
systemctl restart postgresql
systemctl restart nginx
```

## Common Failure Scenarios

### Scenario 1: Database Corruption

**Symptoms:**

- Database connection errors
- Data inconsistencies
- Query failures

**Recovery:**

1. Stop all services
2. Restore from latest backup
3. Run integrity checks
4. Restart services
5. Verify functionality

### Scenario 2: Disk Space Full

**Symptoms:**

- Write failures
- Service crashes
- Slow performance

**Recovery:**

```bash
# Check disk space
df -h

# Clear logs
find /var/log -name "*.log" -mtime +7 -delete

# Clear old backups
find /backups -mtime +30 -delete

# Restart services
systemctl restart liyali-backend
```

### Scenario 3: Memory Exhaustion

**Symptoms:**

- OOM errors
- Service crashes
- Slow response times

**Recovery:**

```bash
# Check memory
free -h

# Identify memory hogs
ps aux --sort=-%mem | head

# Restart services
systemctl restart liyali-backend

# Consider scaling up
```

### Scenario 4: Network Issues

**Symptoms:**

- Connection timeouts
- API failures
- Slow requests

**Recovery:**

```bash
# Check network
ping api.liyali.com

# Check DNS
nslookup api.liyali.com

# Check firewall
sudo ufw status

# Restart network
systemctl restart networking
```

## Rollback Procedures

### Application Rollback

```bash
# Identify last stable version
git log --oneline

# Rollback to specific commit
git checkout <commit-hash>

# Rebuild and deploy
make deploy
```

### Database Rollback

```bash
# Rollback migrations
cd backend
make migrate-down

# Or restore from backup
psql liyali_gateway < backup.sql
```

## Health Checks

### Automated Monitoring

```bash
# Health check script
#!/bin/bash

# Check backend
curl -f http://localhost:8080/health || exit 1

# Check database
psql -U postgres -c "SELECT 1" || exit 1

# Check disk space
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 90 ]; then
    echo "Disk usage critical: ${DISK_USAGE}%"
    exit 1
fi
```

### Manual Checks

```bash
# Service status
systemctl status liyali-backend
systemctl status postgresql
systemctl status nginx

# Database connections
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Disk space
df -h

# Memory usage
free -h

# CPU usage
top
```

## Emergency Contacts

- **DevOps Lead:** devops@liyali.com
- **Database Admin:** dba@liyali.com
- **On-Call:** +1-XXX-XXX-XXXX

## Post-Recovery

### Verification Checklist

- [ ] All services running
- [ ] Database accessible
- [ ] API responding
- [ ] Frontend loading
- [ ] Admin console accessible
- [ ] Critical features working
- [ ] No error logs
- [ ] Monitoring active

### Incident Report

Document:

1. What happened
2. When it happened
3. Impact duration
4. Root cause
5. Recovery steps taken
6. Prevention measures

## Prevention

### Regular Maintenance

- Daily automated backups
- Weekly backup verification
- Monthly disaster recovery drills
- Quarterly security audits

### Monitoring

- Uptime monitoring
- Error tracking
- Performance monitoring
- Log aggregation

### Documentation

- Keep runbooks updated
- Document all procedures
- Maintain contact list
- Update recovery plans

## Resources

- [Deployment Guide](./04-DEPLOYMENT.md)
- [Monitoring](../backend/docs/15-monitoring.md)
- [Troubleshooting](../backend/docs/16-troubleshooting.md)
