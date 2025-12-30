# Configuration Guide

Complete configuration reference for the Liyali Gateway Backend.

## Configuration Overview

The application uses environment variables for configuration, with support for `.env` files in development. Configuration is organized into logical groups for easy management.

## Environment Variables

### Database Configuration

```env
# PostgreSQL Database Settings
DB_HOST=localhost                    # Database host
DB_PORT=5432                        # Database port
DB_USER=postgres                    # Database username
DB_PASSWORD=your_password           # Database password
DB_NAME=liyali_gateway             # Database name
DB_SSL_MODE=disable                # SSL mode (disable/require/verify-full)
```

**SSL Mode Options:**
- `disable` - No SSL (development only)
- `require` - SSL required but no verification
- `verify-full` - SSL required with full verification (production)

### Application Configuration

```env
# Server Settings
APP_PORT=8080                      # HTTP server port
APP_ENV=development                # Environment (development/staging/production)

# Security Settings
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
JWT_EXPIRY=24h                     # JWT token expiry (optional, default: 24h)
REFRESH_TOKEN_EXPIRY=168h          # Refresh token expiry (optional, default: 7 days)

# Frontend Integration
FRONTEND_URL=http://localhost:3000  # Frontend URL for CORS
```

### Optional Configuration

```env
# Logging
LOG_LEVEL=info                     # Log level (debug/info/warn/error)

# CORS
ENABLE_CORS=true                   # Enable CORS (true/false)

# Rate Limiting
RATE_LIMIT_ENABLED=true            # Enable rate limiting
RATE_LIMIT_REQUESTS=100            # Requests per minute per IP
RATE_LIMIT_WINDOW=1m               # Rate limit window

# Session Management
SESSION_TIMEOUT=30m                # Session timeout
MAX_SESSIONS_PER_USER=5            # Maximum concurrent sessions per user

# Account Security
MAX_LOGIN_ATTEMPTS=5               # Maximum failed login attempts
LOCKOUT_DURATION=15m               # Account lockout duration
PASSWORD_RESET_EXPIRY=1h           # Password reset token expiry
```

## Environment-Specific Configurations

### Development Configuration

```env
# .env.development
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=liyali_gateway_dev
DB_SSL_MODE=disable

APP_PORT=8080
APP_ENV=development
LOG_LEVEL=debug

JWT_SECRET=development-secret-key-not-for-production-use
FRONTEND_URL=http://localhost:3000

ENABLE_CORS=true
RATE_LIMIT_ENABLED=false
```

### Staging Configuration

```env
# .env.staging
DB_HOST=staging-db.company.com
DB_PORT=5432
DB_USER=staging_user
DB_PASSWORD=staging_secure_password
DB_NAME=liyali_gateway_staging
DB_SSL_MODE=require

APP_PORT=8080
APP_ENV=staging
LOG_LEVEL=info

JWT_SECRET=staging-secret-key-32-characters-minimum
FRONTEND_URL=https://staging.company.com

ENABLE_CORS=true
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=200
```

### Production Configuration

```env
# .env.production
DB_HOST=prod-db.company.com
DB_PORT=5432
DB_USER=prod_user
DB_PASSWORD=very-secure-production-password
DB_NAME=liyali_gateway_prod
DB_SSL_MODE=verify-full

APP_PORT=8080
APP_ENV=production
LOG_LEVEL=warn

JWT_SECRET=production-secret-key-minimum-32-characters-very-secure
FRONTEND_URL=https://app.company.com

ENABLE_CORS=false
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

SESSION_TIMEOUT=30m
MAX_SESSIONS_PER_USER=3
MAX_LOGIN_ATTEMPTS=3
LOCKOUT_DURATION=30m
```

## Configuration Validation

The application validates configuration on startup:

### Required Variables
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET` (minimum 32 characters)
- `APP_PORT`

### Validation Rules
- JWT_SECRET must be at least 32 characters
- APP_PORT must be a valid port number (1-65535)
- DB_PORT must be a valid port number
- LOG_LEVEL must be one of: debug, info, warn, error
- APP_ENV must be one of: development, staging, production

## Database Configuration

### Connection Pool Settings

The application uses GORM with connection pooling:

```go
// Configured automatically based on environment
MaxIdleConns:    10              // Maximum idle connections
MaxOpenConns:    100             // Maximum open connections
ConnMaxLifetime: time.Hour       // Connection maximum lifetime
```

### Migration Configuration

Migrations are run automatically on startup in development mode:

```env
# Migration Settings (optional)
AUTO_MIGRATE=true                 # Auto-run migrations (development only)
MIGRATION_PATH=./database/migrations  # Migration files path
```

## Security Configuration

### JWT Configuration

```env
# JWT Settings
JWT_SECRET=your-secret-key-minimum-32-characters
JWT_EXPIRY=24h                    # Token expiry time
JWT_ISSUER=liyali-gateway         # Token issuer
JWT_AUDIENCE=liyali-users         # Token audience
```

### Password Policy

```env
# Password Requirements (enforced by application)
MIN_PASSWORD_LENGTH=8             # Minimum password length
REQUIRE_UPPERCASE=true            # Require uppercase letters
REQUIRE_LOWERCASE=true            # Require lowercase letters
REQUIRE_NUMBERS=true              # Require numbers
REQUIRE_SPECIAL_CHARS=true        # Require special characters
```

### CORS Configuration

```env
# CORS Settings
ENABLE_CORS=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.company.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,PATCH,OPTIONS
CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization
CORS_ALLOW_CREDENTIALS=true
```

## Logging Configuration

### Log Levels

- `debug` - Detailed debugging information
- `info` - General information messages
- `warn` - Warning messages
- `error` - Error messages only

### Log Format

```env
# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json                   # json or text
LOG_OUTPUT=stdout                 # stdout or file path
```

## Performance Configuration

### Database Performance

```env
# Database Performance
DB_MAX_IDLE_CONNS=10             # Maximum idle connections
DB_MAX_OPEN_CONNS=100            # Maximum open connections
DB_CONN_MAX_LIFETIME=1h          # Connection maximum lifetime
```

### Application Performance

```env
# Server Performance
READ_TIMEOUT=30s                 # HTTP read timeout
WRITE_TIMEOUT=30s                # HTTP write timeout
IDLE_TIMEOUT=120s                # HTTP idle timeout
MAX_HEADER_SIZE=1MB              # Maximum header size
```

## Monitoring Configuration

### Health Check Configuration

```env
# Health Check Settings
HEALTH_CHECK_ENABLED=true        # Enable health check endpoint
HEALTH_CHECK_PATH=/health        # Health check endpoint path
```

### Metrics Configuration

```env
# Metrics Settings
METRICS_ENABLED=true             # Enable metrics collection
METRICS_PATH=/metrics            # Metrics endpoint path
```

## Configuration Loading

### Loading Order

1. Default values (hardcoded)
2. Environment variables
3. `.env` file (if present)
4. Command line flags (if implemented)

### Configuration Validation

```go
// Example configuration validation
func validateConfig() error {
    if len(os.Getenv("JWT_SECRET")) < 32 {
        return errors.New("JWT_SECRET must be at least 32 characters")
    }
    
    if os.Getenv("DB_HOST") == "" {
        return errors.New("DB_HOST is required")
    }
    
    // Additional validations...
    return nil
}
```

## Configuration Examples

### Docker Configuration

```yaml
# docker-compose.yml
version: '3.8'
services:
  backend:
    image: liyali-gateway-backend
    environment:
      DB_HOST: postgres
      DB_USER: postgres
      DB_PASSWORD: password
      DB_NAME: liyali_gateway
      JWT_SECRET: docker-secret-key-minimum-32-characters
      FRONTEND_URL: http://localhost:3000
    ports:
      - "8080:8080"
```

### Kubernetes Configuration

```yaml
# k8s-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: liyali-gateway-config
data:
  APP_ENV: "production"
  APP_PORT: "8080"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "liyali_gateway"
  FRONTEND_URL: "https://app.company.com"
---
apiVersion: v1
kind: Secret
metadata:
  name: liyali-gateway-secrets
type: Opaque
stringData:
  DB_PASSWORD: "secure-password"
  JWT_SECRET: "production-secret-key-minimum-32-characters"
```

## Configuration Best Practices

### Security Best Practices

1. **Never commit secrets** to version control
2. **Use strong JWT secrets** (minimum 32 characters)
3. **Enable SSL** in production (`DB_SSL_MODE=verify-full`)
4. **Restrict CORS origins** in production
5. **Use environment-specific configurations**

### Performance Best Practices

1. **Tune connection pools** based on load
2. **Set appropriate timeouts** for your use case
3. **Enable rate limiting** in production
4. **Configure logging levels** appropriately

### Operational Best Practices

1. **Use configuration management** tools
2. **Validate configuration** on startup
3. **Document all configuration** options
4. **Use consistent naming** conventions
5. **Provide sensible defaults**

## Troubleshooting Configuration

### Common Issues

**Database Connection Failed**
```bash
# Check database configuration
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME

# Test connection
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
```

**JWT Secret Too Short**
```bash
# Generate secure JWT secret
openssl rand -base64 32
```

**CORS Issues**
```bash
# Check CORS configuration
echo $FRONTEND_URL
echo $CORS_ALLOWED_ORIGINS
```

**Port Already in Use**
```bash
# Check what's using the port
lsof -i :8080

# Use different port
export APP_PORT=8081
```

### Configuration Debugging

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Check configuration loading
go run main.go --config-debug
```

## Next Steps

- **Development**: Set up [Development Environment](./11-development.md)
- **Security**: Review [Authentication & Authorization](./07-auth.md)
- **Deployment**: Configure for [Production Deployment](./14-deployment.md)
- **Monitoring**: Set up [Monitoring & Logging](./15-monitoring.md)