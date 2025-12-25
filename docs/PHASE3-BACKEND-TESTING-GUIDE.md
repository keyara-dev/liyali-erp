# Phase 3A Backend Testing Guide

## Running Tests

### Unit Tests for Permission Service

Run the permission service tests:

```bash
cd backend
go test ./services -v -run TestPermissionService
```

This will run all permission service tests:
- `TestHasPermission` - Tests permission checking for all roles
- `TestGetRolePermissions` - Tests retrieving permissions for a role
- `TestGetAllRoles` - Tests getting all available roles
- `TestGetResources` - Tests getting all available resources
- `TestGetActionsForResource` - Tests getting actions for a resource

### Expected Test Output

```
=== RUN   TestHasPermission
=== RUN   TestHasPermission/Admin_-_view_requisition
--- PASS: TestHasPermission/Admin_-_view_requisition (0.00s)
=== RUN   TestHasPermission/Admin_-_approve_requisition
--- PASS: TestHasPermission/Admin_-_approve_requisition (0.00s)
=== RUN   TestHasPermission/Approver_-_view_requisition
--- PASS: TestHasPermission/Approver_-_view_requisition (0.00s)
...
ok      github.com/liyali/liyali-gateway/services   0.025s
```

---

## API Testing with cURL

### 1. Setup: Get Authentication Token

First, you need to register and login to get a JWT token:

```bash
# Register a new user
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "approver@example.com",
    "password": "password123",
    "name": "John Approver"
  }'

# Login to get JWT token
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "approver@example.com",
    "password": "password123"
  }' | jq '.data.token'
```

Save the token in a variable for easier testing:

```bash
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 2. Test Permission Granted Scenarios

#### Test 1: Approver CAN view requisitions

```bash
curl -X GET http://localhost:3000/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

Expected response: `200 OK` with requisitions list

```json
{
  "success": true,
  "message": "Requisitions retrieved successfully",
  "data": [
    {
      "id": "req1",
      "title": "Office Supplies",
      "status": "draft"
    }
  ]
}
```

#### Test 2: Approver CAN approve requisitions

```bash
curl -X POST http://localhost:3000/api/v1/requisitions/{id}/approve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "comment": "Approved"
  }'
```

Expected response: `200 OK` with updated requisition

#### Test 3: Approver CAN view budgets

```bash
curl -X GET http://localhost:3000/api/v1/budgets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

Expected response: `200 OK` with budgets list

### 3. Test Permission Denied Scenarios

#### Test 4: Requester CANNOT approve requisitions

Login as a requester first:

```bash
# Register requester
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "requester@example.com",
    "password": "password123",
    "name": "Jane Requester"
  }'

# Login as requester
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "requester@example.com",
    "password": "password123"
  }' | jq '.data.token'

export REQUESTER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

Try to approve a requisition:

```bash
curl -X POST http://localhost:3000/api/v1/requisitions/{id}/approve \
  -H "Authorization: Bearer $REQUESTER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "comment": "Approved"
  }'
```

Expected response: `403 Forbidden`

```json
{
  "error": "Insufficient permissions for this action"
}
```

#### Test 5: Viewer CANNOT create requisitions

Login as a viewer and try to create:

```bash
curl -X POST http://localhost:3000/api/v1/requisitions \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Requisition",
    "description": "Test requisition",
    "items": [],
    "totalAmount": 1000
  }'
```

Expected response: `403 Forbidden`

#### Test 6: Requester CANNOT delete requisitions

```bash
curl -X DELETE http://localhost:3000/api/v1/requisitions/{id} \
  -H "Authorization: Bearer $REQUESTER_TOKEN" \
  -H "Content-Type: application/json"
```

Expected response: `403 Forbidden`

### 4. Test Missing Organization Context

#### Test 7: Request without organization context fails

Try to access a protected route without being in an organization:

```bash
curl -X GET http://localhost:3000/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

Expected response: `400 Bad Request` (if organization context is required)

```json
{
  "error": "Organization ID not found in context"
}
```

### 5. Test Missing Authorization Header

#### Test 8: Request without token fails

```bash
curl -X GET http://localhost:3000/api/v1/requisitions \
  -H "Content-Type: application/json"
```

Expected response: `401 Unauthorized`

```json
{
  "error": "Authorization header required"
}
```

### 6. Complete Test Sequence

Here's a complete test sequence to validate the entire permission system:

```bash
#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:3000/api/v1"

echo "=== Phase 3A Permission System Test ==="
echo ""

# 1. Register users with different roles
echo "1. Registering test users..."
curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"test123","name":"Admin User"}' > /dev/null
echo -e "${GREEN}✓ Admin registered${NC}"

curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"approver@test.com","password":"test123","name":"Approver User"}' > /dev/null
echo -e "${GREEN}✓ Approver registered${NC}"

curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"requester@test.com","password":"test123","name":"Requester User"}' > /dev/null
echo -e "${GREEN}✓ Requester registered${NC}"

# 2. Login and get tokens
echo ""
echo "2. Getting authentication tokens..."
ADMIN_TOKEN=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"test123"}' | jq -r '.data.token')
echo -e "${GREEN}✓ Admin token obtained${NC}"

APPROVER_TOKEN=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"approver@test.com","password":"test123"}' | jq -r '.data.token')
echo -e "${GREEN}✓ Approver token obtained${NC}"

REQUESTER_TOKEN=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"requester@test.com","password":"test123"}' | jq -r '.data.token')
echo -e "${GREEN}✓ Requester token obtained${NC}"

# 3. Test permission checks
echo ""
echo "3. Testing permission checks..."

# Test admin can view requisitions
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/requisitions \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json")
if [ $STATUS -eq 200 ]; then
  echo -e "${GREEN}✓ Admin can view requisitions (200)${NC}"
else
  echo -e "${RED}✗ Admin view requisitions failed ($STATUS)${NC}"
fi

# Test approver can view requisitions
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/requisitions \
  -H "Authorization: Bearer $APPROVER_TOKEN" \
  -H "Content-Type: application/json")
if [ $STATUS -eq 200 ]; then
  echo -e "${GREEN}✓ Approver can view requisitions (200)${NC}"
else
  echo -e "${RED}✗ Approver view requisitions failed ($STATUS)${NC}"
fi

# Test requester cannot approve requisitions
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/requisitions/test-id/approve \
  -H "Authorization: Bearer $REQUESTER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"comment":"Test"}')
if [ $STATUS -eq 403 ]; then
  echo -e "${GREEN}✓ Requester cannot approve requisitions (403)${NC}"
else
  echo -e "${RED}✗ Requester approval check failed ($STATUS)${NC}"
fi

# Test missing authorization header
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/requisitions \
  -H "Content-Type: application/json")
if [ $STATUS -eq 401 ]; then
  echo -e "${GREEN}✓ Request without token rejected (401)${NC}"
else
  echo -e "${RED}✗ Authorization check failed ($STATUS)${NC}"
fi

echo ""
echo "=== Test Complete ==="
```

Save this as `test-permissions.sh` and run:

```bash
chmod +x test-permissions.sh
./test-permissions.sh
```

---

## Permission Matrix Reference

### Admin Role
- **Requisition**: view, create, edit, delete, approve, reject
- **Budget**: view, create, edit, delete, approve, reject
- **Purchase Order**: view, create, edit, delete, approve, reject
- **Payment Voucher**: view, create, edit, delete, approve, reject
- **GRN**: view, create, edit, delete
- **Vendor**: view, create, edit, delete
- **Category**: view, create, edit, delete
- **Organization**: view, edit, manage_users, manage_workflows
- **Analytics**: view
- **Audit Log**: view

### Approver Role
- **Requisition**: view, create, edit, approve, reject
- **Budget**: view, approve, reject
- **Purchase Order**: view, approve, reject
- **Payment Voucher**: view, approve, reject
- **GRN**: view
- **Vendor**: view
- **Category**: view
- **Analytics**: view

### Requester Role
- **Requisition**: view, create, edit
- **Budget**: view
- **Purchase Order**: view
- **Payment Voucher**: view
- **GRN**: view
- **Vendor**: view
- **Category**: view

### Finance Role
- **Requisition**: view, approve, reject
- **Budget**: view, create, edit, approve, reject
- **Purchase Order**: view, approve, reject
- **Payment Voucher**: view, create, edit, approve, reject
- **GRN**: view
- **Vendor**: view
- **Category**: view
- **Analytics**: view
- **Audit Log**: view

### Viewer Role
- **Requisition**: view
- **Budget**: view
- **Purchase Order**: view
- **Payment Voucher**: view
- **GRN**: view
- **Vendor**: view
- **Category**: view
- **Analytics**: view

---

## HTTP Status Codes

- **200 OK**: Permission granted, request successful
- **400 Bad Request**: Missing organization context or invalid request
- **401 Unauthorized**: Missing or invalid authentication token
- **403 Forbidden**: User lacks required permission
- **404 Not Found**: Resource not found
- **500 Internal Server Error**: Server error

---

## Debugging Permission Issues

### Enable Debug Logging

Add this to your main.go to enable permission service debug logging:

```go
log.SetFlags(log.LstdFlags | log.Llongfile)
```

### Check User Context

In your handler, you can inspect what's in the context:

```go
userID := c.Locals("userID").(string)
userRole := c.Locals("userRole").(string)
organizationID := c.Locals("organizationID").(string)

log.Printf("User: %s, Role: %s, Org: %s", userID, userRole, organizationID)
```

### Common Issues

1. **"Organization ID not found in context"**
   - Ensure TenantMiddleware is applied before the permission middleware
   - Check that the organization context is being set correctly

2. **"User role not found in context"**
   - Ensure AuthMiddleware is applied and parsing the JWT correctly
   - Verify JWT token contains the `role` claim

3. **"Insufficient permissions for this action"**
   - Check the permission matrix above
   - Verify the resource and action names match exactly
   - Ensure the user's role is set correctly in the database

