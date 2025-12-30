# Troubleshooting Guide

Common issues and solutions for the Liyali Gateway Backend.

## Common Issues

### Database Connection Issues

#### Issue: "connection refused" or "database does not exist"

**Symptoms:**
```
Error: failed to connect to database: dial tcp [::1]:5432: connect: connection refused
```

**Solutions:**

1. **Check PostgreSQL is running:**
```bash
# Linux/macOS
sudo systemctl status postgresql
# or
brew services list | grep postgresql

# Windows
net start postgresql-x64-14
```

2. **Verify database exists:**
```bash
psql -l | grep liyali_gateway
```

3. **Create database if missing:**
```bash
createdb liyali_gateway
```

4. **Check connection parameters:**
```bash
# Test connection manually
psql -h localhost -p 5432 -U postgres -d liyali_gateway
```

5. **Verify environment variables:**
```bash
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME
```

#### Issue: "password authentication failed"

**Solutions:**

1. **Reset PostgreSQL password:**
```bash
sudo -u postgres psql
ALTER USER postgres PASSWORD 'newpassword';
\q
```

2. **Check pg_hba.conf configuration:**
```bash
# Find config file
sudo -u postgres psql -c "SHOW hba_file;"

# Edit to allow local connections
sudo nano /etc/postgresql/14/main/pg_hba.conf

# Change this line:
# local   all             postgres                                peer
# to:
# local   all             postgres                                md5
```

3. **Restart PostgreSQL:**
```bash
sudo systemctl restart postgresql
```

#### Issue: "too many connections"

**Solutions:**

1. **Check current connections:**
```sql
SELECT count(*) FROM pg_stat_activity;
```

2. **Kill idle connections:**
```sql
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle' AND state_change < now() - interval '5 minutes';
```

3. **Increase max_connections:**
```bash
# Edit postgresql.conf
sudo nano /etc/postgresql/14/main/postgresql.conf

# Increase max_connections
max_connections = 200

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### Application Startup Issues

#### Issue: "Bootstrap failed at phase connect"

**Symptoms:**
```
Error: Bootstrap failed at phase connect: failed to connect to database
```

**Solutions:**

1. **Check database connectivity:**
```bash
# Test database connection
psql -h localhost -p 5432 -U postgres -d liyali_gateway

# Check if database is running
sudo systemctl status postgresql
```

2. **Verify bootstrap configuration:**
```bash
# Check bootstrap environment variables
echo $BOOTSTRAP_TIMEOUT $VALIDATION_TIMEOUT $MIGRATION_TIMEOUT

# Enable bootstrap debug logging
export LOG_LEVEL=debug
go run main.go
```

3. **Check circuit breaker status:**
```bash
# Check circuit breaker metrics
curl http://localhost:8080/metrics | grep bootstrap_circuit_breaker
```

#### Issue: "Bootstrap failed at phase validate"

**Symptoms:**
```
Error: Bootstrap failed at phase validate: database schema validation failed
```

**Solutions:**

1. **Check database schema:**
```sql
-- Verify required tables exist
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;
```

2. **Run migrations manually:**
```bash
# Run migration script
cd database && ./migrate.sh up

# Or run specific migration
psql -d liyali_gateway -f database/migrations/001_create_complete_schema.up.sql
```

3. **Reset bootstrap state:**
```bash
# Clear any bootstrap locks or state
# The bootstrap system is stateless, so restart the application
pkill -f "go run main.go"
go run main.go
```

#### Issue: "Bootstrap failed at phase seed"

**Symptoms:**
```
Error: Bootstrap failed at phase seed: seeding operation failed
```

**Solutions:**

1. **Check seeding configuration:**
```env
# Verify seeding is enabled (if needed)
ENABLE_SEEDING=true
SEED_RETRY_ATTEMPTS=5
SEED_RETRY_DELAY=5s
```

2. **Check database constraints:**
```sql
-- Check for constraint violations
SELECT conname, contype FROM pg_constraint 
WHERE contype IN ('f', 'u', 'c');
```

3. **Manual seeding recovery:**
```bash
# The bootstrap system uses idempotent UPSERT operations
# Simply restart the application - it will retry seeding
go run main.go
```

#### Issue: "Circuit breaker is open"

**Symptoms:**
```
Error: bootstrap circuit breaker is open, refusing to execute
```

**Solutions:**

1. **Check circuit breaker status:**
```bash
# Check circuit breaker metrics
curl http://localhost:8080/health/detailed | jq '.checks.bootstrap_system'
```

2. **Wait for auto-recovery:**
```bash
# Circuit breaker will automatically attempt to close after timeout
# Default timeout is 2 minutes
echo "Waiting for circuit breaker auto-recovery..."
sleep 120
```

3. **Manual circuit breaker reset:**
```bash
# Restart the application to reset circuit breaker
pkill -f "go run main.go"
go run main.go
```

#### Issue: "JWT_SECRET must be at least 32 characters"

**Solution:**
```bash
# Generate secure JWT secret
openssl rand -base64 32

# Add to .env file
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env
```

#### Issue: "port already in use"

**Solutions:**

1. **Find process using port:**
```bash
# Linux/macOS
lsof -i :8080

# Windows
netstat -ano | findstr :8080
```

2. **Kill process:**
```bash
# Linux/macOS
lsof -ti:8080 | xargs kill -9

# Windows
taskkill /PID <PID> /F
```

3. **Use different port:**
```bash
export APP_PORT=8081
# or edit .env file
```

#### Issue: "migration failed" or "table already exists"

**Solutions:**

1. **Check bootstrap status:**
```bash
# Check if bootstrap completed successfully
curl http://localhost:8080/health | jq '.checks.database.bootstrap_completed'
```

2. **Check migration status:**
```sql
-- The bootstrap system handles migrations automatically
-- Check if tables exist
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;
```

3. **Reset database (development only):**
```bash
dropdb liyali_gateway
createdb liyali_gateway
# Restart application - bootstrap will handle everything
go run main.go
```

4. **Manual migration (if bootstrap fails):**
```bash
# Use the migration script as fallback
cd database && ./migrate.sh up
```

#### Issue: "Bootstrap timeout exceeded"

**Symptoms:**
```
Error: Bootstrap failed: context deadline exceeded
```

**Solutions:**

1. **Increase bootstrap timeout:**
```env
# In .env file
BOOTSTRAP_TIMEOUT=600s        # 10 minutes
VALIDATION_TIMEOUT=120s       # 2 minutes
MIGRATION_TIMEOUT=1200s       # 20 minutes
```

2. **Check database performance:**
```sql
-- Check for slow queries during bootstrap
SELECT query, mean_time, calls 
FROM pg_stat_statements 
WHERE query LIKE '%CREATE%' OR query LIKE '%ALTER%'
ORDER BY mean_time DESC;
```

3. **Monitor bootstrap progress:**
```bash
# Enable debug logging to see bootstrap phases
export LOG_LEVEL=debug
go run main.go
```

### Authentication Issues

#### Issue: "invalid token" or "token expired"

**Solutions:**

1. **Check JWT secret consistency:**
```bash
# Ensure JWT_SECRET is same across restarts
grep JWT_SECRET .env
```

2. **Verify token format:**
```bash
# Token should be: Bearer <jwt-token>
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8080/api/v1/profile
```

3. **Check token expiry:**
```bash
# Decode JWT to check expiry (use jwt.io)
echo "your-jwt-token" | base64 -d
```

#### Issue: "account locked" or "too many failed attempts"

**Solutions:**

1. **Check lockout status:**
```sql
SELECT email, failed_attempts, locked_until 
FROM users 
WHERE email = 'user@example.com';
```

2. **Unlock account manually:**
```sql
UPDATE users 
SET failed_attempts = 0, locked_until = NULL 
WHERE email = 'user@example.com';
```

3. **Adjust lockout settings:**
```env
# In .env file
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=15m
```

### API Issues

#### Issue: "404 Not Found" for valid endpoints

**Solutions:**

1. **Check route registration:**
```bash
# Enable debug logging
export LOG_LEVEL=debug
go run main.go
```

2. **Verify API prefix:**
```bash
# Correct endpoint format
curl http://localhost:8080/api/v1/health
```

3. **Check CORS configuration:**
```bash
# Add CORS headers for browser requests
curl -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Authorization" \
     -X OPTIONS http://localhost:8080/api/v1/requisitions
```

#### Issue: "500 Internal Server Error"

**Solutions:**

1. **Check application logs:**
```bash
# Enable debug logging
export LOG_LEVEL=debug
go run main.go
```

2. **Check database connectivity:**
```bash
curl http://localhost:8080/health
```

3. **Verify request format:**
```bash
# Ensure Content-Type header
curl -X POST http://localhost:8080/api/v1/requisitions \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer token" \
     -d '{"title": "Test"}'
```

### Performance Issues

#### Issue: Slow database queries

**Solutions:**

1. **Identify slow queries:**
```sql
-- Enable query logging
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();

-- Check slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

2. **Add missing indexes:**
```sql
-- Common indexes
CREATE INDEX CONCURRENTLY idx_requisitions_organization_id ON requisitions(organization_id);
CREATE INDEX CONCURRENTLY idx_requisitions_status ON requisitions(status);
CREATE INDEX CONCURRENTLY idx_documents_type_org ON documents(document_type, organization_id);
```

3. **Optimize connection pool:**
```env
# In .env file
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h
```

#### Issue: High memory usage

**Solutions:**

1. **Check memory stats:**
```bash
curl http://localhost:8080/health/detailed | jq '.system.memory'
```

2. **Force garbage collection:**
```bash
curl -X POST http://localhost:8080/debug/gc
```

3. **Monitor goroutines:**
```bash
curl http://localhost:8080/metrics | grep go_goroutines
```

### Development Issues

#### Issue: "go mod" errors or dependency issues

**Solutions:**

1. **Clean module cache:**
```bash
go clean -modcache
go mod download
```

2. **Update dependencies:**
```bash
go mod tidy
go mod verify
```

3. **Check Go version:**
```bash
go version
# Should be 1.21 or higher
```

#### Issue: Hot reload not working

**Solutions:**

1. **Install air:**
```bash
go install github.com/cosmtrek/air@latest
```

2. **Check air configuration:**
```bash
# Create .air.toml if missing
air init
```

3. **Run with air:**
```bash
air
```

### Testing Issues

#### Issue: Tests failing with database errors

**Solutions:**

1. **Create test database:**
```bash
createdb liyali_gateway_test
```

2. **Set test environment:**
```bash
export DB_NAME=liyali_gateway_test
export APP_ENV=test
```

3. **Run tests with verbose output:**
```bash
go test -v ./tests/...
```

#### Issue: Integration tests timing out

**Solutions:**

1. **Increase test timeout:**
```bash
go test -timeout 30s ./tests/integration/...
```

2. **Check test database performance:**
```sql
-- Test database connection
SELECT 1;
```

3. **Run tests individually:**
```bash
go test -run TestSpecificTest ./tests/integration/
```

## Debugging Tools

### Enable Debug Logging

```bash
export LOG_LEVEL=debug
go run main.go
```

### Bootstrap System Debugging

```bash
# Check bootstrap health and metrics
curl http://localhost:8080/health/detailed | jq '.checks.bootstrap_system'

# Check bootstrap metrics
curl http://localhost:8080/metrics | grep bootstrap

# Monitor bootstrap phases in real-time
tail -f app.log | grep -i bootstrap

# Check circuit breaker status
curl http://localhost:8080/metrics | grep bootstrap_circuit_breaker

# Force bootstrap validation (development only)
curl -X POST http://localhost:8080/debug/bootstrap/validate

# Check bootstrap timing metrics
curl http://localhost:8080/metrics | grep bootstrap_phase_duration
```

### Database Debugging

```sql
-- Check active connections
SELECT pid, usename, application_name, state, query_start, query
FROM pg_stat_activity
WHERE state != 'idle';

-- Check table sizes
SELECT schemaname, tablename, 
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Check bootstrap-related queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
WHERE query LIKE '%UPSERT%' OR query LIKE '%ON CONFLICT%'
ORDER BY mean_time DESC;

-- Check seeding operations
SELECT query, mean_time, calls 
FROM pg_stat_statements 
WHERE query LIKE '%INSERT%' AND query LIKE '%ON CONFLICT%'
ORDER BY calls DESC;
```

### Application Debugging

```bash
# Check health status
curl http://localhost:8080/health

# Check detailed health
curl -H "Authorization: Bearer token" http://localhost:8080/health/detailed

# Check metrics
curl -H "Authorization: Bearer token" http://localhost:8080/metrics

# Test specific endpoint
curl -v -X GET http://localhost:8080/api/v1/organizations \
     -H "Authorization: Bearer token"
```

### Network Debugging

```bash
# Check port availability
netstat -tulpn | grep :8080

# Test connectivity
telnet localhost 8080

# Check DNS resolution
nslookup localhost
```

## Log Analysis

### Common Log Patterns

**Successful Request:**
```json
{
  "level": "info",
  "message": "Request completed",
  "method": "GET",
  "endpoint": "/api/v1/requisitions",
  "status_code": 200,
  "duration_ms": 45
}
```

**Authentication Error:**
```json
{
  "level": "error",
  "message": "Authentication failed",
  "error": "invalid token",
  "endpoint": "/api/v1/requisitions"
}
```

**Database Error:**
```json
{
  "level": "error",
  "message": "Database query failed",
  "error": "connection refused",
  "query": "SELECT * FROM requisitions"
}
```

### Log Analysis Commands

```bash
# Filter error logs
grep '"level":"error"' app.log | jq .

# Count requests by endpoint
grep '"endpoint"' app.log | jq -r .endpoint | sort | uniq -c

# Find slow requests
grep '"duration_ms"' app.log | jq 'select(.duration_ms > 1000)'

# Monitor real-time logs
tail -f app.log | jq .
```

## Performance Monitoring

### Key Metrics to Monitor

1. **Response Time**: 95th percentile < 500ms
2. **Error Rate**: < 1% of total requests
3. **Database Connections**: < 80% of max_connections
4. **Memory Usage**: < 80% of available memory
5. **CPU Usage**: < 70% average

### Monitoring Commands

```bash
# Check system resources
top
htop
free -h
df -h

# Monitor database
psql -d liyali_gateway -c "SELECT count(*) FROM pg_stat_activity;"

# Monitor application
curl http://localhost:8080/metrics | grep -E "(http_requests|database_connections)"
```

## Recovery Procedures

### Database Recovery

1. **Backup current state:**
```bash
pg_dump liyali_gateway > backup_$(date +%Y%m%d_%H%M%S).sql
```

2. **Restore from backup:**
```bash
dropdb liyali_gateway
createdb liyali_gateway
psql -d liyali_gateway < backup_20240101_120000.sql
```

3. **Verify data integrity:**
```sql
SELECT count(*) FROM requisitions;
SELECT count(*) FROM users;
SELECT count(*) FROM organizations;
```

### Application Recovery

1. **Restart application:**
```bash
# Kill existing process
pkill -f "go run main.go"

# Start fresh
go run main.go
```

2. **Clear cache/sessions:**
```sql
DELETE FROM user_sessions WHERE expires_at < NOW();
```

3. **Verify functionality:**
```bash
curl http://localhost:8080/health
```

## Getting Help

### Before Asking for Help

1. **Check logs** for error messages
2. **Verify configuration** settings
3. **Test basic connectivity** (database, network)
4. **Review recent changes** that might have caused issues
5. **Check system resources** (memory, disk, CPU)

### Information to Include

When reporting issues, include:

- **Error messages** (full stack trace)
- **Configuration** (sanitized .env file)
- **System information** (OS, Go version, PostgreSQL version)
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Recent changes** made to the system

### Useful Commands for Diagnostics

```bash
# System information
uname -a
go version
psql --version

# Application status
ps aux | grep go
netstat -tulpn | grep :8080

# Database status
sudo systemctl status postgresql
psql -l

# Disk space
df -h
du -sh /var/log/

# Memory usage
free -h
cat /proc/meminfo
```

## Prevention

### Best Practices

1. **Regular backups** of database and configuration
2. **Monitor system resources** proactively
3. **Keep dependencies updated** regularly
4. **Use version control** for all configuration changes
5. **Test changes** in development environment first
6. **Document customizations** and configurations
7. **Set up monitoring** and alerting
8. **Regular security updates** for OS and dependencies

### Maintenance Schedule

**Daily:**
- Check application logs for errors
- Monitor system resources
- Verify backup completion

**Weekly:**
- Review slow query logs
- Check disk space usage
- Update dependencies (in development)

**Monthly:**
- Full system backup
- Security updates
- Performance review
- Documentation updates

This troubleshooting guide covers the most common issues you might encounter. For specific problems not covered here, check the application logs and system status first, then consult the relevant documentation sections.