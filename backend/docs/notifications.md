# Notification System

Real-time notification system for approval workflows and document status updates.

## Architecture

- **Handler**: `handlers/notification_handler.go` - HTTP endpoints
- **Model**: `models/models.go` - Notification data structure
- **Routes**: Registered in `routes/routes.go` under `/api/v1/notifications`

## Endpoints

```
GET    /api/v1/notifications          # List notifications (paginated)
GET    /api/v1/notifications/recent   # Recent notifications for header
GET    /api/v1/notifications/stats    # Notification statistics
POST   /api/v1/notifications/mark-as-read
POST   /api/v1/notifications/mark-all-as-read
DELETE /api/v1/notifications/{id}
```

## Database Schema

```sql
CREATE TABLE notifications (
    id              VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    recipient_id    VARCHAR(255) NOT NULL,
    type           VARCHAR(50) NOT NULL,
    subject        VARCHAR(255) NOT NULL,
    body           TEXT NOT NULL,
    sent           BOOLEAN DEFAULT false,
    sent_at        TIMESTAMP,
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Usage

All endpoints require authentication and organization context via middleware.

```bash
curl -H "Authorization: Bearer {token}" \
     -H "X-Organization-ID: org-demo-001" \
     http://localhost:8080/api/v1/notifications/stats
```
