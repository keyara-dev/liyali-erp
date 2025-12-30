# API Design

Comprehensive guide to the RESTful API design patterns, conventions, and best practices used in the Liyali Gateway Backend.

## API Overview

The Liyali Gateway Backend provides a comprehensive RESTful API built with Go Fiber, following OpenAPI 3.0 specifications. The API is designed for multi-tenant operations with robust authentication and authorization.

### Base URL Structure

```
Production:  https://api.company.com/api/v1
Staging:     https://staging-api.company.com/api/v1
Development: http://localhost:8080/api/v1
```

### API Versioning

- **Current Version**: v1
- **Versioning Strategy**: URL path versioning (`/api/v1/`)
- **Backward Compatibility**: Maintained within major versions
- **Deprecation Policy**: 6-month notice for breaking changes

## Authentication & Authorization

### Authentication Flow

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@company.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "user-123",
      "email": "user@company.com",
      "name": "John Doe",
      "currentOrganizationId": "org-456"
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 86400
  }
}
```

### Authorization Header

All protected endpoints require the JWT token in the Authorization header:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Permission-Based Access

Endpoints are protected by granular permissions:

```http
GET /api/v1/requisitions
# Requires: requisition:view permission

POST /api/v1/requisitions
# Requires: requisition:create permission

PUT /api/v1/requisitions/:id
# Requires: requisition:edit permission
```

## Request/Response Patterns

### Standard Response Format

All API responses follow a consistent structure:

#### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data here
  }
}
```

#### Error Response
```json
{
  "success": false,
  "error": "Validation failed",
  "message": "The provided data is invalid",
  "details": {
    "field": "email",
    "issue": "Invalid email format"
  }
}
```

#### Paginated Response
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": [
    // Array of items
  ],
  "pagination": {
    "total": 150,
    "page": 2,
    "totalPages": 15,
    "pageSize": 10,
    "hasNext": true,
    "hasPrev": true
  }
}
```

### HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Successful GET, PUT requests |
| 201 | Created | Successful POST requests |
| 204 | No Content | Successful DELETE requests |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Validation errors |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server errors |

## API Endpoints

### Authentication Endpoints

#### Public Authentication
```http
POST /api/v1/auth/register      # User registration
POST /api/v1/auth/login         # User login
POST /api/v1/auth/refresh       # Refresh access token
POST /api/v1/auth/verify        # Verify token validity
POST /api/v1/auth/password-reset/request  # Request password reset
POST /api/v1/auth/password-reset/confirm  # Confirm password reset
```

#### Protected Authentication
```http
GET  /api/v1/auth/profile       # Get user profile
POST /api/v1/auth/logout        # Logout current session
POST /api/v1/auth/logout-all    # Logout all sessions
POST /api/v1/auth/change-password  # Change password
```

### Organization Management

#### Organization Operations
```http
GET  /api/v1/organizations      # List user's organizations
POST /api/v1/organizations      # Create organization
PUT  /api/v1/organizations/:id  # Update organization
POST /api/v1/organizations/:id/switch  # Switch current organization
```

#### Organization Members
```http
GET    /api/v1/organization/members        # List members
POST   /api/v1/organization/members        # Add member
DELETE /api/v1/organization/members/:userId # Remove member
GET    /api/v1/organization/settings       # Get settings
PUT    /api/v1/organization/settings       # Update settings
```

### Document Management

#### Specific Document APIs

**Requisitions:**
```http
GET    /api/v1/requisitions           # List requisitions
POST   /api/v1/requisitions           # Create requisition
GET    /api/v1/requisitions/:id       # Get requisition
PUT    /api/v1/requisitions/:id       # Update requisition
DELETE /api/v1/requisitions/:id       # Delete requisition
POST   /api/v1/requisitions/:id/approve  # Approve requisition
POST   /api/v1/requisitions/:id/reject   # Reject requisition
```

**Budgets:**
```http
GET    /api/v1/budgets                # List budgets
POST   /api/v1/budgets                # Create budget
GET    /api/v1/budgets/:id            # Get budget
PUT    /api/v1/budgets/:id            # Update budget
DELETE /api/v1/budgets/:id            # Delete budget
POST   /api/v1/budgets/:id/approve    # Approve budget
POST   /api/v1/budgets/:id/reject     # Reject budget
```

**Purchase Orders:**
```http
GET    /api/v1/purchase-orders        # List purchase orders
POST   /api/v1/purchase-orders        # Create purchase order
GET    /api/v1/purchase-orders/:id    # Get purchase order
PUT    /api/v1/purchase-orders/:id    # Update purchase order
DELETE /api/v1/purchase-orders/:id    # Delete purchase order
POST   /api/v1/purchase-orders/:id/approve  # Approve purchase order
POST   /api/v1/purchase-orders/:id/reject   # Reject purchase order
```

#### Generic Document API

```http
GET    /api/v1/documents              # List all documents
GET    /api/v1/documents/my           # Get user's documents
GET    /api/v1/documents/search       # Search documents
GET    /api/v1/documents/stats        # Document statistics
GET    /api/v1/documents/:id          # Get document by ID
GET    /api/v1/documents/number/:number  # Get by document number
POST   /api/v1/documents              # Create generic document
PUT    /api/v1/documents/:id          # Update generic document
POST   /api/v1/documents/:id/submit   # Submit for approval
DELETE /api/v1/documents/:id          # Delete document
```

### Workflow Management

```http
GET    /api/v1/workflows              # List workflows
POST   /api/v1/workflows              # Create workflow
GET    /api/v1/workflows/:id          # Get workflow
PUT    /api/v1/workflows/:id          # Update workflow
POST   /api/v1/workflows/:id/activate    # Activate workflow
POST   /api/v1/workflows/:id/deactivate  # Deactivate workflow
DELETE /api/v1/workflows/:id          # Delete workflow
GET    /api/v1/workflows/default/:documentType  # Get default workflow
```

### Approval Management

#### Approval Tasks
```http
GET  /api/v1/approvals                # Get approval tasks
GET  /api/v1/approvals/:id            # Get approval task
POST /api/v1/approvals/:id/approve    # Approve task
POST /api/v1/approvals/:id/reject     # Reject task
POST /api/v1/approvals/:id/reassign   # Reassign task
GET  /api/v1/approvals/tasks/overdue  # Get overdue tasks
```

#### Bulk Operations
```http
POST /api/v1/approvals/bulk/approve   # Bulk approve
POST /api/v1/approvals/bulk/reject    # Bulk reject
POST /api/v1/approvals/bulk/reassign  # Bulk reassign
```

#### Approval History
```http
GET /api/v1/documents/:documentId/approval-history  # Get approval history
```

### Analytics & Reporting

```http
GET /api/v1/analytics/dashboard           # Dashboard metrics
GET /api/v1/analytics/requisitions/metrics  # Requisition metrics
GET /api/v1/analytics/approvals/metrics     # Approval metrics
```

## Request/Response Examples

### Create Requisition

**Request:**
```http
POST /api/v1/requisitions
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "title": "Office Supplies Q1 2024",
  "description": "Quarterly office supplies procurement",
  "department": "Administration",
  "priority": "medium",
  "items": [
    {
      "name": "Laptop Dell XPS 13",
      "description": "Development laptop",
      "quantity": 2,
      "unitPrice": 1200.00,
      "totalPrice": 2400.00,
      "category": "Electronics"
    },
    {
      "name": "Office Chair",
      "description": "Ergonomic office chair",
      "quantity": 5,
      "unitPrice": 300.00,
      "totalPrice": 1500.00,
      "category": "Furniture"
    }
  ],
  "totalAmount": 3900.00,
  "currency": "USD",
  "categoryId": "cat-electronics-001",
  "preferredVendorId": "vendor-dell-001"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Requisition created successfully",
  "data": {
    "id": "req-20241228-12345678",
    "organizationId": "org-456",
    "reqNumber": "REQ-20241228-12345678",
    "requesterId": "user-123",
    "title": "Office Supplies Q1 2024",
    "description": "Quarterly office supplies procurement",
    "department": "Administration",
    "status": "draft",
    "priority": "medium",
    "items": [
      {
        "name": "Laptop Dell XPS 13",
        "description": "Development laptop",
        "quantity": 2,
        "unitPrice": 1200.00,
        "totalPrice": 2400.00,
        "category": "Electronics"
      }
    ],
    "totalAmount": 3900.00,
    "currency": "USD",
    "approvalStage": 0,
    "categoryId": "cat-electronics-001",
    "preferredVendorId": "vendor-dell-001",
    "isEstimate": false,
    "createdAt": "2024-12-28T10:30:00Z",
    "updatedAt": "2024-12-28T10:30:00Z"
  }
}
```

### Search Documents

**Request:**
```http
GET /api/v1/documents/search?q=laptop&documentTypes=REQUISITION,PURCHASE_ORDER&page=1&limit=10
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response:**
```json
{
  "success": true,
  "message": "Document search completed successfully",
  "data": [
    {
      "id": "doc-uuid-123",
      "organizationId": "org-456",
      "documentType": "REQUISITION",
      "documentNumber": "REQ-20241228-12345678",
      "title": "Office Supplies Q1 2024",
      "description": "Quarterly office supplies procurement",
      "status": "draft",
      "amount": 3900.00,
      "currency": "USD",
      "department": "Administration",
      "createdBy": "user-123",
      "data": {
        "items": [...],
        "priority": "medium"
      },
      "relevance": 2.5,
      "matches": ["title", "description"],
      "createdAt": "2024-12-28T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 25,
    "page": 1,
    "totalPages": 3,
    "pageSize": 10,
    "hasNext": true,
    "hasPrev": false
  }
}
```

### Bulk Approve Tasks

**Request:**
```http
POST /api/v1/approvals/bulk/approve
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "taskIds": [
    "task-123",
    "task-456",
    "task-789"
  ],
  "signature": "John Doe - Approved via bulk operation",
  "comment": "All items reviewed and approved for Q1 procurement"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Bulk approval completed",
  "data": {
    "successCount": 2,
    "failureCount": 1,
    "successIds": ["task-123", "task-456"],
    "errors": [
      "Task task-789: not in pending status"
    ]
  }
}
```

## Query Parameters

### Pagination Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (1-based) |
| `limit` | integer | 20 | Items per page (max 100) |

### Filtering Parameters

#### Common Filters
| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status |
| `department` | string | Filter by department |
| `dateFrom` | date | Start date (YYYY-MM-DD) |
| `dateTo` | date | End date (YYYY-MM-DD) |
| `createdBy` | string | Filter by creator |

#### Document-Specific Filters
| Parameter | Type | Description |
|-----------|------|-------------|
| `documentTypes` | string | Comma-separated document types |
| `statuses` | string | Comma-separated statuses |
| `amountMin` | number | Minimum amount |
| `amountMax` | number | Maximum amount |

#### Search Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Search query |
| `search` | string | Alternative search parameter |

### Example Query Strings

```http
# Pagination
GET /api/v1/requisitions?page=2&limit=25

# Filtering
GET /api/v1/requisitions?status=approved&department=IT&dateFrom=2024-01-01

# Search with filters
GET /api/v1/documents/search?q=laptop&documentTypes=REQUISITION,PURCHASE_ORDER&amountMin=1000

# Complex filtering
GET /api/v1/documents?statuses=approved,pending&dateFrom=2024-01-01&dateTo=2024-12-31&page=1&limit=50
```

## Error Handling

### Validation Errors

**Request:**
```http
POST /api/v1/requisitions
{
  "title": "",
  "items": [],
  "totalAmount": -100
}
```

**Response:**
```json
{
  "success": false,
  "error": "Validation failed",
  "message": "The provided data contains validation errors",
  "details": {
    "title": "Title is required",
    "items": "At least one item is required",
    "totalAmount": "Total amount must be greater than 0"
  }
}
```

### Permission Errors

**Response:**
```json
{
  "success": false,
  "error": "Insufficient permissions",
  "message": "You don't have permission to perform this action",
  "details": {
    "required": "requisition:create",
    "action": "create_requisition"
  }
}
```

### Not Found Errors

**Response:**
```json
{
  "success": false,
  "error": "Resource not found",
  "message": "The requested requisition was not found",
  "details": {
    "resource": "requisition",
    "id": "req-nonexistent"
  }
}
```

## Rate Limiting

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

### Rate Limit Response

```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "details": {
    "limit": 100,
    "window": "1 minute",
    "retryAfter": 60
  }
}
```

## API Versioning & Deprecation

### Version Headers

```http
API-Version: v1
Supported-Versions: v1
```

### Deprecation Warnings

```http
Deprecation: true
Sunset: Sat, 01 Jan 2025 00:00:00 GMT
Link: </api/v2/requisitions>; rel="successor-version"
```

## Content Types

### Supported Content Types

- **Request**: `application/json`
- **Response**: `application/json`
- **File Upload**: `multipart/form-data`

### Character Encoding

- **Default**: UTF-8
- **Header**: `Content-Type: application/json; charset=utf-8`

## CORS Configuration

### Allowed Origins
- Development: `http://localhost:3000`
- Staging: `https://staging.company.com`
- Production: `https://app.company.com`

### Allowed Methods
- `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`

### Allowed Headers
- `Origin`, `Content-Type`, `Accept`, `Authorization`

## API Documentation

### OpenAPI Specification

The complete API is documented using OpenAPI 3.0:

- **Specification File**: `backend/openapi.yaml`
- **Interactive Docs**: Available via Swagger UI
- **Postman Collection**: Generated from OpenAPI spec

### Testing

- **HTTP File**: `backend/API.http` - REST Client compatible
- **Test Suite**: Comprehensive integration tests
- **Mock Server**: Available for frontend development

## Best Practices

### Request Design

1. **Use appropriate HTTP methods**
2. **Include proper Content-Type headers**
3. **Validate all input data**
4. **Use consistent naming conventions**
5. **Implement proper error handling**

### Response Design

1. **Return consistent response formats**
2. **Include appropriate HTTP status codes**
3. **Provide meaningful error messages**
4. **Use pagination for large datasets**
5. **Include relevant metadata**

### Security

1. **Always validate authentication**
2. **Check permissions for all operations**
3. **Sanitize input data**
4. **Use HTTPS in production**
5. **Implement rate limiting**

## Next Steps

- **Authentication**: Deep dive into [Auth System](./07-auth.md)
- **Document Management**: Explore [Document Operations](./08-documents.md)
- **API Reference**: Complete [API Documentation](./13-api-reference.md)
- **Testing**: Set up [API Testing](./12-testing.md)