# Monitoring & Observability

Complete guide to monitoring, logging, and observability for the Liyali Gateway Backend.

## Monitoring Overview

The backend implements comprehensive monitoring and observability:

- **Health Checks** - Application and dependency health monitoring
- **Metrics Collection** - Performance and business metrics
- **Logging** - Structured logging with correlation IDs
- **Distributed Tracing** - Request tracing across services
- **Alerting** - Proactive issue detection and notification
- **Dashboards** - Real-time monitoring dashboards

## Health Checks

### Application Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "liyali-gateway-backend",
  "version": "1.0.0",
  "timestamp": "2024-01-01T10:00:00Z",
  "uptime": "2h 30m 45s",
  "checks": {
    "database": {
      "status": "healthy",
      "responseTime": "5ms",
      "bootstrap_completed": true,
      "last_bootstrap": "2024-01-01T07:30:00Z",
      "bootstrap_duration": "2.3s",
      "details": {
        "host": "localhost:5432",
        "database": "liyali_gateway",
        "connections": {
          "active": 5,
          "idle": 15,
          "max": 100
        }
      }
    },
    "bootstrap_system": {
      "status": "healthy",
      "circuit_breaker_state": "CLOSED",
      "circuit_breaker_failures": 0,
      "last_validation": "2024-01-01T09:55:00Z",
      "validation_duration": "123ms"
    },
    "redis": {
      "status": "healthy",
      "responseTime": "2ms",
      "details": {
        "host": "localhost:6379",
        "memory": "45MB",
        "connected_clients": 3
      }
    },
    "external_apis": {
      "status": "healthy",
      "details": {
        "email_service": "healthy",
        "notification_service": "healthy"
      }
    }
  }
}
```

### Detailed Health Check

```http
GET /health/detailed
Authorization: Bearer jwt-access-token
```

**Response:**
```json
{
  "status": "healthy",
  "service": "liyali-gateway-backend",
  "version": "1.0.0",
  "timestamp": "2024-01-01T10:00:00Z",
  "uptime": "2h 30m 45s",
  "system": {
    "memory": {
      "used": "256MB",
      "available": "1.5GB",
      "usage_percent": 14.5
    },
    "cpu": {
      "usage_percent": 12.3,
      "load_average": [0.5, 0.7, 0.8]
    },
    "disk": {
      "used": "45GB",
      "available": "155GB",
      "usage_percent": 22.5
    }
  },
  "application": {
    "goroutines": 45,
    "gc_stats": {
      "num_gc": 123,
      "pause_total": "45ms",
      "last_gc": "2024-01-01T09:58:30Z"
    },
    "requests": {
      "total": 15420,
      "per_second": 12.5,
      "errors": 23,
      "error_rate": 0.15
    }
  },
  "dependencies": {
    "database": {
      "status": "healthy",
      "response_time": "5ms",
      "connections": {
        "active": 5,
        "idle": 15,
        "max": 100
      },
      "queries": {
        "total": 8945,
        "slow_queries": 2,
        "average_time": "12ms"
      }
    }
  }
}
```

### Readiness Check

```http
GET /ready
```

**Response:**
```json
{
  "status": "ready",
  "checks": {
    "database_migrations": {
      "status": "ready",
      "applied": 9,
      "pending": 0
    },
    "database_connection": {
      "status": "ready",
      "response_time": "3ms"
    },
    "external_dependencies": {
      "status": "ready",
      "services": ["email", "notifications"]
    }
  }
}
```

### Liveness Check

```http
GET /live
```

**Response:**
```json
{
  "status": "alive",
  "timestamp": "2024-01-01T10:00:00Z",
  "uptime": "2h 30m 45s"
}
```

## Metrics Collection

### Application Metrics

The backend exposes Prometheus-compatible metrics:

```http
GET /metrics
Authorization: Bearer jwt-access-token
```

**Key Metrics:**

```prometheus
# HTTP Request Metrics
http_requests_total{method="GET",endpoint="/api/v1/requisitions",status="200"} 1234
http_request_duration_seconds{method="GET",endpoint="/api/v1/requisitions"} 0.045

# Database Metrics
database_connections_active 15
database_connections_idle 25
database_query_duration_seconds{operation="SELECT"} 0.012
database_queries_total{operation="SELECT",status="success"} 8945

# Bootstrap System Metrics
bootstrap_phase_duration_seconds{phase="connect"} 0.045
bootstrap_phase_duration_seconds{phase="validate"} 0.123
bootstrap_phase_duration_seconds{phase="migrate"} 1.234
bootstrap_phase_duration_seconds{phase="verify"} 0.567
bootstrap_phase_duration_seconds{phase="seed"} 0.890
bootstrap_total_duration_seconds 2.859
bootstrap_circuit_breaker_state{state="CLOSED"} 1
bootstrap_circuit_breaker_failures_total 0
bootstrap_validation_checks_total{status="success"} 156
bootstrap_seed_operations_total{entity="users",status="created"} 4
bootstrap_seed_operations_total{entity="users",status="updated"} 0

# Business Metrics
requisitions_created_total 456
requisitions_approved_total 234
purchase_orders_generated_total 123
documents_searched_total 2345

# Authentication Metrics
auth_login_attempts_total{status="success"} 1234
auth_login_attempts_total{status="failed"} 45
auth_sessions_active 89
auth_password_resets_total 12

# System Metrics
go_goroutines 45
go_memstats_alloc_bytes 67108864
go_gc_duration_seconds 0.001234
```

### Custom Business Metrics

```go
// Example metrics implementation
var (
    requisitionsCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "requisitions_created_total",
            Help: "Total number of requisitions created",
        },
        []string{"department", "priority"},
    )
    
    approvalDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "approval_duration_seconds",
            Help: "Time taken for approval process",
            Buckets: prometheus.DefBuckets,
        },
        []string{"workflow_type"},
    )
)
```

## Structured Logging

### Log Format

All logs use structured JSON format with correlation IDs:

```json
{
  "timestamp": "2024-01-01T10:00:00Z",
  "level": "info",
  "message": "Requisition created successfully",
  "correlation_id": "req-123e4567-e89b-12d3-a456-426614174000",
  "user_id": "user-789",
  "organization_id": "org-456",
  "request_id": "req-abc123",
  "method": "POST",
  "endpoint": "/api/v1/requisitions",
  "status_code": 201,
  "duration_ms": 45,
  "fields": {
    "requisition_id": "req-new-123",
    "total_amount": 2400.00,
    "department": "IT"
  }
}
```

### Log Levels and Usage

**DEBUG** - Detailed debugging information
```json
{
  "level": "debug",
  "message": "Database query executed",
  "query": "SELECT * FROM requisitions WHERE id = $1",
  "params": ["req-123"],
  "duration_ms": 12
}
```

**INFO** - General application flow
```json
{
  "level": "info",
  "message": "User authenticated successfully",
  "user_id": "user-123",
  "session_id": "sess-456"
}
```

**WARN** - Warning conditions
```json
{
  "level": "warn",
  "message": "Rate limit approaching",
  "user_id": "user-123",
  "requests_count": 95,
  "limit": 100
}
```

**ERROR** - Error conditions
```json
{
  "level": "error",
  "message": "Database connection failed",
  "error": "connection refused",
  "retry_count": 3,
  "max_retries": 5
}
```

### Correlation ID Tracking

Every request gets a unique correlation ID that tracks the request through all services:

```go
// Middleware adds correlation ID to context
func CorrelationIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        correlationID := c.GetHeader("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }
        
        c.Set("correlation_id", correlationID)
        c.Header("X-Correlation-ID", correlationID)
        c.Next()
    }
}
```

## Distributed Tracing

### OpenTelemetry Integration

The backend supports OpenTelemetry for distributed tracing:

```go
// Tracing configuration
func initTracing() {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
    ))
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("liyali-gateway-backend"),
            semconv.ServiceVersionKey.String("1.0.0"),
        )),
    )
    
    otel.SetTracerProvider(tp)
}
```

### Trace Examples

**HTTP Request Trace:**
```json
{
  "trace_id": "123e4567e89b12d3a456426614174000",
  "span_id": "a456426614174000",
  "operation_name": "POST /api/v1/requisitions",
  "start_time": "2024-01-01T10:00:00Z",
  "duration": "45ms",
  "tags": {
    "http.method": "POST",
    "http.url": "/api/v1/requisitions",
    "http.status_code": 201,
    "user.id": "user-123"
  },
  "logs": [
    {
      "timestamp": "2024-01-01T10:00:00.010Z",
      "message": "Validating request data"
    },
    {
      "timestamp": "2024-01-01T10:00:00.025Z",
      "message": "Creating requisition in database"
    }
  ]
}
```

## Alerting

### Alert Rules

**High Error Rate:**
```yaml
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "High error rate detected"
    description: "Error rate is {{ $value }} errors per second"
```

**Database Connection Issues:**
```yaml
- alert: DatabaseConnectionFailed
  expr: database_connections_active == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Database connection failed"
    description: "No active database connections"
```

**Bootstrap System Failures:**
```yaml
- alert: BootstrapCircuitBreakerOpen
  expr: bootstrap_circuit_breaker_state{state="OPEN"} == 1
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Bootstrap circuit breaker is open"
    description: "Bootstrap system circuit breaker has opened due to failures"

- alert: BootstrapValidationFailed
  expr: increase(bootstrap_validation_checks_total{status="failed"}[5m]) > 0
  for: 1m
  labels:
    severity: warning
  annotations:
    summary: "Bootstrap validation failures detected"
    description: "{{ $value }} bootstrap validation failures in the last 5 minutes"
```

**High Response Time:**
```yaml
- alert: HighResponseTime
  expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High response time"
    description: "95th percentile response time is {{ $value }}s"
```

### Alert Channels

Configure alert notifications:

```yaml
# alertmanager.yml
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  slack_configs:
  - api_url: 'YOUR_SLACK_WEBHOOK_URL'
    channel: '#alerts'
    title: 'Liyali Gateway Alert'
    text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

## Dashboards

### Grafana Dashboard Configuration

**System Overview Dashboard:**
```json
{
  "dashboard": {
    "title": "Liyali Gateway - System Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "5xx errors"
          }
        ]
      }
    ]
  }
}
```

**Business Metrics Dashboard:**
```json
{
  "dashboard": {
    "title": "Liyali Gateway - Business Metrics",
    "panels": [
      {
        "title": "Requisitions Created",
        "type": "stat",
        "targets": [
          {
            "expr": "increase(requisitions_created_total[24h])",
            "legendFormat": "Last 24h"
          }
        ]
      },
      {
        "title": "Approval Rate",
        "type": "gauge",
        "targets": [
          {
            "expr": "rate(requisitions_approved_total[1h]) / rate(requisitions_created_total[1h]) * 100",
            "legendFormat": "Approval Rate %"
          }
        ]
      }
    ]
  }
}
```

## Performance Monitoring

### Database Performance

Monitor database performance with these queries:

```sql
-- Slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Connection usage
SELECT count(*) as active_connections,
       max_conn,
       max_conn - count(*) as available_connections
FROM pg_stat_activity, 
     (SELECT setting::int as max_conn FROM pg_settings WHERE name = 'max_connections') mc;

-- Table sizes
SELECT schemaname, tablename, 
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Application Performance

Monitor Go application performance:

```go
// Memory usage monitoring
func getMemoryStats() map[string]interface{} {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return map[string]interface{}{
        "alloc_mb":      bToMb(m.Alloc),
        "total_alloc_mb": bToMb(m.TotalAlloc),
        "sys_mb":        bToMb(m.Sys),
        "num_gc":        m.NumGC,
        "goroutines":    runtime.NumGoroutine(),
    }
}
```

## Log Aggregation

### ELK Stack Integration

**Logstash Configuration:**
```ruby
input {
  beats {
    port => 5044
  }
}

filter {
  if [fields][service] == "liyali-gateway-backend" {
    json {
      source => "message"
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
    
    mutate {
      add_field => { "service" => "liyali-gateway-backend" }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "liyali-gateway-%{+YYYY.MM.dd}"
  }
}
```

**Filebeat Configuration:**
```yaml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/liyali-gateway/*.log
  fields:
    service: liyali-gateway-backend
  fields_under_root: true

output.logstash:
  hosts: ["logstash:5044"]
```

## Security Monitoring

### Audit Logging

All security-relevant events are logged:

```json
{
  "timestamp": "2024-01-01T10:00:00Z",
  "level": "info",
  "event_type": "security_audit",
  "action": "login_attempt",
  "user_id": "user-123",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "result": "success",
  "session_id": "sess-456"
}
```

### Failed Authentication Monitoring

```json
{
  "timestamp": "2024-01-01T10:00:00Z",
  "level": "warn",
  "event_type": "security_audit",
  "action": "login_failed",
  "email": "user@example.com",
  "ip_address": "192.168.1.100",
  "reason": "invalid_password",
  "attempt_count": 3
}
```

## Monitoring Setup

### Docker Compose Monitoring Stack

```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

  jaeger:
    image: jaegertracing/all-in-one
    ports:
      - "16686:16686"
      - "14268:14268"

volumes:
  grafana-storage:
```

### Kubernetes Monitoring

```yaml
apiVersion: v1
kind: Service
metadata:
  name: liyali-gateway-metrics
  labels:
    app: liyali-gateway
spec:
  ports:
  - port: 8080
    name: metrics
  selector:
    app: liyali-gateway
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: liyali-gateway
spec:
  selector:
    matchLabels:
      app: liyali-gateway
  endpoints:
  - port: metrics
    path: /metrics
```

## Troubleshooting Monitoring

### Common Issues

**Metrics Not Appearing:**
```bash
# Check metrics endpoint
curl http://localhost:8080/metrics

# Verify Prometheus configuration
curl http://localhost:9090/api/v1/targets
```

**High Memory Usage:**
```bash
# Check Go memory stats
curl http://localhost:8080/health/detailed | jq '.system.memory'

# Force garbage collection
curl -X POST http://localhost:8080/debug/gc
```

**Database Connection Issues:**
```bash
# Check database health
curl http://localhost:8080/health | jq '.checks.database'

# Monitor connection pool
curl http://localhost:8080/metrics | grep database_connections
```

## Best Practices

### Monitoring Best Practices

1. **Monitor business metrics** alongside technical metrics
2. **Set up proactive alerts** for critical issues
3. **Use correlation IDs** for request tracing
4. **Implement health checks** for all dependencies
5. **Monitor resource usage** trends over time

### Logging Best Practices

1. **Use structured logging** (JSON format)
2. **Include correlation IDs** in all log entries
3. **Log at appropriate levels** (avoid debug in production)
4. **Include relevant context** in log messages
5. **Sanitize sensitive data** before logging

### Performance Best Practices

1. **Monitor database query performance**
2. **Track response time percentiles**
3. **Set up resource usage alerts**
4. **Monitor garbage collection** performance
5. **Track business KPIs** alongside technical metrics

## Next Steps

- **Deployment**: Set up [Production Deployment](./14-deployment.md)
- **Security**: Review [Security Monitoring](./07-auth.md)
- **Testing**: Implement [Performance Testing](./12-testing.md)
- **Operations**: Create [Runbook Procedures](./16-troubleshooting.md)

The monitoring system provides comprehensive observability into the Liyali Gateway Backend, enabling proactive issue detection and performance optimization.