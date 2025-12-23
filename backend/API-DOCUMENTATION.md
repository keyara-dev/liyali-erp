# Liyali Gateway Procurement API Documentation

## Overview

The Liyali Gateway Procurement API is a comprehensive REST API for managing procurement workflows, including requisitions, budgets, purchase orders, goods receipt, and payment processing.

**API Version:** 1.0.0
**Base URL:** `http://localhost:8080/api/v1` (Development)
**Authentication:** JWT Bearer Token

## Table of Contents

1. [Authentication](#authentication)
2. [Base Response Format](#base-response-format)
3. [Error Handling](#error-handling)
4. [Pagination](#pagination)
5. [Endpoints](#endpoints)
6. [Examples](#examples)

## Authentication

### Login Endpoint

All authenticated endpoints require a JWT token obtained from the login endpoint.

**Endpoint:** `POST /auth/login`

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@company.com",
    "password": "securepassword123"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@company.com",
      "name": "John Doe",
      "role": "manager",
      "department": "IT",
      "active": true
    }
  }
}
```

### Using the Token

Include the token in the `Authorization` header:

```bash
curl -X GET http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Token Expiration:** 24 hours

## Base Response Format

All API responses follow a consistent format:

### Success Response (2xx)
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Sample Resource"
  }
}
```

### List Response (2xx)
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Item 1"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 10,
  "hasNext": true,
  "hasPrev": false
}
```

### Error Response (4xx/5xx)
```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error information"
}
```

## Error Handling

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Access denied |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 500 | Server Error | Internal server error |

### Common Error Responses

**Missing Required Field:**
```json
{
  "success": false,
  "message": "Vendor ID is required"
}
```

**Invalid Amount:**
```json
{
  "success": false,
  "message": "Amount must be greater than 0"
}
```

**Unauthorized Access:**
```json
{
  "success": false,
  "message": "Unauthorized"
}
```

**Resource Not Found:**
```json
{
  "success": false,
  "message": "Requisition not found"
}
```

## Pagination

List endpoints support pagination with the following parameters:

- `page` (integer, default: 1): Page number
- `limit` (integer, default: 10, max: 100): Items per page

### Pagination Example

```bash
curl -X GET "http://localhost:8080/api/v1/requisitions?page=2&limit=20" \
  -H "Authorization: Bearer <token>"
```

**Response Pagination Fields:**
```json
{
  "success": true,
  "data": [...],
  "total": 150,
  "page": 2,
  "limit": 20,
  "hasNext": true,
  "hasPrev": true
}
```

## Endpoints

### Requisitions

#### List Requisitions
```
GET /requisitions
```

**Parameters:**
- `page` (query): Page number
- `limit` (query): Items per page
- `status` (query): Filter by status (draft, pending, approved, rejected, fulfilled)
- `department` (query): Filter by department

**Example:**
```bash
curl -X GET "http://localhost:8080/api/v1/requisitions?status=pending&page=1&limit=10" \
  -H "Authorization: Bearer <token>"
```

#### Create Requisition
```
POST /requisitions
```

**Request Body:**
```json
{
  "department": "IT",
  "totalAmount": 50000,
  "currency": "USD",
  "deliveryDate": "2025-12-31T23:59:59Z",
  "description": "Office supplies and equipment",
  "items": [
    {
      "description": "Laptops",
      "quantity": 5,
      "unitPrice": 10000,
      "amount": 50000
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "reqNumber": "REQ-20251223-abc12345",
    "userId": "user-id",
    "department": "IT",
    "status": "draft",
    "totalAmount": 50000,
    "currency": "USD",
    "deliveryDate": "2025-12-31T23:59:59Z",
    "approvalStage": 0,
    "createdAt": "2025-12-23T12:00:00Z",
    "updatedAt": "2025-12-23T12:00:00Z"
  }
}
```

#### Get Requisition
```
GET /requisitions/{id}
```

**Example:**
```bash
curl -X GET http://localhost:8080/api/v1/requisitions/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <token>"
```

#### Update Requisition
```
PUT /requisitions/{id}
```

**Request Body (only draft status can be updated):**
```json
{
  "department": "IT",
  "totalAmount": 60000,
  "currency": "USD",
  "deliveryDate": "2025-01-31T23:59:59Z"
}
```

#### Submit Requisition for Approval
```
POST /requisitions/{id}/submit
```

**Request Body:**
```json
{
  "comments": "Ready for approval"
}
```

#### Delete Requisition
```
DELETE /requisitions/{id}
```

**Note:** Only draft requisitions can be deleted

### Budgets

#### List Budgets
```
GET /budgets
```

**Parameters:**
- `page` (query): Page number
- `limit` (query): Items per page
- `status` (query): Filter by status
- `fiscalYear` (query): Filter by fiscal year (e.g., "2025")

#### Create Budget
```
POST /budgets
```

**Request Body:**
```json
{
  "budgetCode": "IT-2025-Q1",
  "department": "IT",
  "fiscalYear": "2025",
  "totalBudget": 500000,
  "allocatedAmount": 0
}
```

**Constraints:**
- Budget code must be unique
- One budget per department per fiscal year
- Total budget must be > 0
- Allocated amount ≤ total budget
- 10-15% reserve funds must be maintained

#### Get Budget
```
GET /budgets/{id}
```

#### Update Budget
```
PUT /budgets/{id}
```

**Note:** Only draft budgets can be updated

### Purchase Orders

#### List Purchase Orders
```
GET /purchase-orders
```

**Parameters:**
- `page` (query): Page number
- `limit` (query): Items per page
- `status` (query): Filter by status

#### Create Purchase Order
```
POST /purchase-orders
```

**Request Body:**
```json
{
  "vendorId": "vendor-uuid",
  "totalAmount": 50000,
  "currency": "USD",
  "deliveryDate": "2025-12-31T23:59:59Z",
  "linkedRequisition": "requisition-uuid",
  "items": [
    {
      "description": "Office Supplies",
      "quantity": 100,
      "unitPrice": 500,
      "amount": 50000
    }
  ]
}
```

**Constraints:**
- Vendor must exist
- Total amount must be > 0
- Vendor spending limit: 30% of total budget
- Quote required for amounts > $25,000

#### Approve Purchase Order
```
POST /purchase-orders/{id}/approve
```

**Request Body:**
```json
{
  "comments": "Approved for procurement",
  "signature": "FM-abc12345"
}
```

### Goods Received Notes (GRN)

#### Create GRN
```
POST /grn
```

**Request Body:**
```json
{
  "poNumber": "PO-20251223-abc12345",
  "receivedBy": "John Doe",
  "items": [
    {
      "itemNo": 1,
      "description": "Laptops",
      "quantity": 5,
      "receivedQty": 5
    }
  ],
  "qualityIssues": [
    {
      "itemNo": 1,
      "description": "Minor packaging damage",
      "severity": "low"
    }
  ]
}
```

**Constraints:**
- PO must exist
- At least one item required
- Received quantity can vary from ordered quantity
- Quality issues are optional

#### Quantity Variance Tracking

The system automatically tracks variance between ordered and received quantities:

```json
{
  "orderedQuantity": 100,
  "receivedQuantity": 95,
  "variance": -5,
  "variancePercent": -5.0
}
```

### Payment Vouchers

#### Create Payment Voucher
```
POST /payment-vouchers
```

**Request Body:**
```json
{
  "vendorId": "vendor-uuid",
  "invoiceNumber": "INV-2025-001",
  "amount": 50000,
  "currency": "USD",
  "paymentMethod": "bank_transfer",
  "glCode": "4000",
  "description": "Payment for office supplies - Invoice INV-2025-001",
  "linkedPO": "po-uuid"
}
```

**Payment Methods:** bank_transfer, check, cash, credit_card, wire

**Constraints:**
- Vendor must exist
- Invoice number must be unique per vendor
- Amount must be > 0
- Description must be ≥ 10 characters
- GL Code must be valid (4+ digits)

#### Approve Payment Voucher
```
POST /payment-vouchers/{id}/approve
```

**Request Body:**
```json
{
  "comments": "Approved for payment",
  "signature": "FM-def67890"
}
```

### Vendors

#### List Vendors
```
GET /vendors
```

**Parameters:**
- `page` (query): Page number
- `limit` (query): Items per page
- `active` (query): Filter by active status (true/false)
- `country` (query): Filter by country

#### Create Vendor
```
POST /vendors
```

**Request Body:**
```json
{
  "name": "ABC Supplies Ltd",
  "email": "contact@abcsupplies.com",
  "phone": "+263 4 123456",
  "country": "Zimbabwe",
  "city": "Harare",
  "bankAccount": "1234567890",
  "taxID": "TAX123456"
}
```

**Constraints:**
- Name must be ≥ 3 characters
- Email must be unique
- All fields required
- Bank account must be ≥ 8 characters

#### Update Vendor
```
PUT /vendors/{id}
```

#### Delete Vendor (Soft Delete)
```
DELETE /vendors/{id}
```

**Note:** Deletes mark vendor as inactive (soft delete via active flag)

## Examples

### Complete Workflow Example

1. **Create Requisition**
```bash
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "department": "IT",
    "totalAmount": 50000,
    "currency": "USD",
    "deliveryDate": "2025-12-31T23:59:59Z",
    "description": "Office supplies",
    "items": [{"description": "Laptops", "quantity": 5, "unitPrice": 10000, "amount": 50000}]
  }'
```

2. **Submit for Approval**
```bash
curl -X POST http://localhost:8080/api/v1/requisitions/{id}/submit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"comments": "Ready for approval"}'
```

3. **Check Budget Availability**
```bash
curl -X GET "http://localhost:8080/api/v1/budgets?department=IT&fiscalYear=2025" \
  -H "Authorization: Bearer <token>"
```

4. **Create Purchase Order**
```bash
curl -X POST http://localhost:8080/api/v1/purchase-orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "vendorId": "vendor-uuid",
    "totalAmount": 50000,
    "currency": "USD",
    "deliveryDate": "2025-12-31T23:59:59Z",
    "linkedRequisition": "requisition-uuid",
    "items": [{"description": "Laptops", "quantity": 5, "unitPrice": 10000, "amount": 50000}]
  }'
```

5. **Approve Purchase Order**
```bash
curl -X POST http://localhost:8080/api/v1/purchase-orders/{id}/approve \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"comments": "Approved", "signature": "FM-abc12345"}'
```

6. **Create GRN**
```bash
curl -X POST http://localhost:8080/api/v1/grn \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "poNumber": "PO-20251223-abc12345",
    "receivedBy": "John Doe",
    "items": [{"itemNo": 1, "description": "Laptops", "quantity": 5, "receivedQty": 5}]
  }'
```

7. **Create Payment Voucher**
```bash
curl -X POST http://localhost:8080/api/v1/payment-vouchers \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "vendorId": "vendor-uuid",
    "invoiceNumber": "INV-2025-001",
    "amount": 50000,
    "currency": "USD",
    "paymentMethod": "bank_transfer",
    "glCode": "4000",
    "description": "Payment for office supplies",
    "linkedPO": "po-uuid"
  }'
```

## API Versioning

The API uses path-based versioning: `/api/v1/`

Future versions (e.g., `/api/v2/`) can be introduced without breaking existing clients.

## Rate Limiting

Rate limiting is applied per user:
- 1000 requests per hour
- 100 requests per minute

## Security

- All endpoints (except `/health`) require JWT authentication
- Tokens expire after 24 hours
- Passwords are hashed using bcrypt
- HTTPS is enforced in production
- CORS is configured for allowed origins

## Support

For issues or questions about the API:
- Email: api-support@liyali.com
- Documentation: https://docs.liyali.com
- Status Page: https://status.liyali.com
