# API Documentation

## Base URL

- Development: `http://localhost:8080`
- Production: `https://api.liyali.com`

## Authentication

All protected endpoints require JWT token:

```http
Authorization: Bearer <token>
```

### Get Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}
```

Response:

```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "User Name"
  }
}
```

## Core Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/forgot-password` - Request reset
- `POST /api/v1/auth/reset-password` - Reset password

### Users

- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Organizations

- `GET /api/v1/organizations` - List organizations
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations/:id` - Get organization
- `PUT /api/v1/organizations/:id` - Update organization

### Workflows

- `GET /api/v1/workflows` - List workflows
- `POST /api/v1/workflows` - Create workflow
- `GET /api/v1/workflows/:id` - Get workflow
- `PUT /api/v1/workflows/:id` - Update workflow
- `POST /api/v1/workflows/:id/execute` - Execute workflow

### Documents

- `GET /api/v1/documents` - List documents
- `POST /api/v1/documents` - Upload document
- `GET /api/v1/documents/:id` - Get document
- `DELETE /api/v1/documents/:id` - Delete document

## Admin Endpoints

Require admin role.

### User Management

- `GET /api/v1/admin/users` - List all users
- `PUT /api/v1/admin/users/:id` - Update user
- `POST /api/v1/admin/users/:id/suspend` - Suspend user
- `POST /api/v1/admin/users/:id/activate` - Activate user

### Organization Management

- `GET /api/v1/admin/organizations` - List all organizations
- `PUT /api/v1/admin/organizations/:id` - Update organization
- `DELETE /api/v1/admin/organizations/:id` - Delete organization

### Subscription Management

- `GET /api/v1/admin/subscriptions` - List subscriptions
- `POST /api/v1/admin/subscriptions` - Create subscription
- `PUT /api/v1/admin/subscriptions/:id` - Update subscription
- `POST /api/v1/admin/subscriptions/:id/cancel` - Cancel subscription

### System Settings

- `GET /api/v1/admin/settings` - List settings
- `POST /api/v1/admin/settings` - Create setting
- `PUT /api/v1/admin/settings/:id` - Update setting
- `DELETE /api/v1/admin/settings/:id` - Delete setting

### Feature Flags

- `GET /api/v1/admin/feature-flags` - List flags
- `POST /api/v1/admin/feature-flags` - Create flag
- `PUT /api/v1/admin/feature-flags/:id` - Update flag
- `DELETE /api/v1/admin/feature-flags/:id` - Delete flag

### Analytics

- `GET /api/v1/admin/analytics/overview` - System overview
- `GET /api/v1/admin/analytics/users` - User analytics
- `GET /api/v1/admin/analytics/subscriptions` - Subscription analytics

## Response Format

### Success Response

```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful"
}
```

### Error Response

```json
{
  "success": false,
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

## Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `422` - Validation Error
- `500` - Server Error

## Rate Limiting

- 100 requests per minute per IP
- 1000 requests per hour per user

## Pagination

```http
GET /api/v1/users?page=1&limit=20
```

Response:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "pages": 5
  }
}
```

## Filtering & Sorting

```http
GET /api/v1/users?status=active&sort=created_at&order=desc
```

## Testing

Use the HTTP files in `backend/scripts/test_requests.http` with REST Client extension.

## OpenAPI Spec

Full OpenAPI specification: `backend/openapi.yaml`

## Resources

- [Backend Documentation](../backend/docs/)
- [Authentication Guide](../backend/docs/07-auth.md)
- [API Reference](../backend/docs/13-api-reference.md)
