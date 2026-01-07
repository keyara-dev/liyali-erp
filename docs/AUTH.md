# Authentication & Authorization Guide

## Overview

Liyali Gateway uses JWT-based authentication with comprehensive role-based access control (RBAC).

## Authentication Flow

### 1. User Registration
```bash
POST /api/v1/auth/register
{
  "name": "John Doe",
  "email": "john@example.com", 
  "password": "securepassword"
}
```

### 2. User Login
```bash
POST /api/v1/auth/login
{
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "john@example.com",
      "name": "John Doe",
      "role": "requester"
    },
    "token": "jwt-token",
    "refreshToken": "refresh-token"
  }
}
```

### 3. Token Refresh
```bash
POST /api/v1/auth/refresh
{
  "refreshToken": "refresh-token"
}
```

## Authorization System

### System Roles

| Role | Permissions | Description |
|------|-------------|-------------|
| **Admin** | Full access (43 permissions) | System administration |
| **Approver** | Approval workflows (21 permissions) | Document approval |
| **Finance** | Financial operations (21 permissions) | Budget and payment management |
| **Requester** | Create requests (8 permissions) | Submit requisitions |
| **Viewer** | Read-only access (7 permissions) | View documents only |

### Custom Roles

Organizations can create custom roles with specific permission combinations:

```bash
# Create custom role
POST /api/v1/organizations/{orgId}/roles
{
  "name": "Department Manager",
  "description": "Manages department operations",
  "permissions": ["requisitions.create", "requisitions.approve"]
}
```

### Permission Categories

**Requisitions**: `create`, `read`, `update`, `delete`, `approve`, `reject`, `reassign`
**Budgets**: `create`, `read`, `update`, `delete`, `approve`, `reject`
**Purchase Orders**: `create`, `read`, `update`, `delete`, `approve`, `reject`
**Payment Vouchers**: `create`, `read`, `update`, `delete`, `approve`, `reject`
**Users**: `create`, `read`, `update`, `delete`, `manage_roles`
**Organizations**: `create`, `read`, `update`, `delete`, `manage_members`

## Multi-Tenant Security

### Organization Isolation
- All data is scoped to organizations
- Users can only access their organization's data
- Cross-organization access is prevented

### Data Access Control
```go
// Example: Organization context middleware
func OrganizationContext(c *fiber.Ctx) error {
    orgID := c.Get("X-Organization-ID")
    // Validate user belongs to organization
    // Set organization context for request
}
```

## Frontend Integration

### Authentication Hook
```typescript
const { user, login, logout, isAuthenticated } = useAuth();

// Login
await login(email, password);

// Check permissions
const canApprove = usePermissions(['requisitions.approve']);
```

### Permission Guards
```typescript
<PermissionGuard permissions={['requisitions.create']}>
  <CreateRequisitionButton />
</PermissionGuard>
```

## Security Features

### Password Security
- Bcrypt hashing with salt rounds
- Minimum password requirements
- Password change tracking

### Token Security
- JWT with expiration (24 hours)
- Refresh tokens for seamless renewal
- Token blacklisting on logout
- Unique JTI (JWT ID) for revocation

### Session Management
- Automatic token refresh
- Secure logout (token revocation)
- Session timeout handling

## API Protection

### Protected Endpoints
All API endpoints require authentication:
```bash
Authorization: Bearer <jwt-token>
X-Organization-ID: <organization-id>
```

### Permission Checking
```go
// Middleware checks permissions
func RequirePermissions(perms ...string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Check user has required permissions
    }
}
```

## Default Users

After setup, these test accounts are available:

### System Administrator
- **Email**: `admin@liyali.com`
- **Password**: `admin123`
- **Permissions**: Full system access
- **Use Case**: System configuration, user management

### Requester
- **Email**: `requester@demo.com`
- **Password**: `admin123`
- **Permissions**: Create and view requisitions
- **Use Case**: Submit purchase requests

### Approver
- **Email**: `approver@demo.com`
- **Password**: `admin123`
- **Permissions**: Approve/reject documents
- **Use Case**: Review and authorize requests

### Finance Officer
- **Email**: `finance@demo.com`
- **Password**: `admin123`
- **Permissions**: Financial operations
- **Use Case**: Budget management, payments

### Department Manager
- **Email**: `manager@demo.com`
- **Password**: `admin123`
- **Permissions**: Department oversight
- **Use Case**: Team management, first-level approval

## Testing Authentication

### Login Test
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@liyali.com","password":"admin123"}'
```

### Protected Endpoint Test
```bash
curl -X GET http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Organization-ID: YOUR_ORG_ID"
```

## Troubleshooting

### Common Issues

**Invalid Token**
- Check token expiration
- Verify token format
- Ensure proper Authorization header

**Permission Denied**
- Verify user role and permissions
- Check organization membership
- Confirm endpoint requires correct permissions

**Login Failed**
- Verify email and password
- Check user account status
- Ensure database connection

---

**Next**: See [API Reference](API.md) for complete endpoint documentation