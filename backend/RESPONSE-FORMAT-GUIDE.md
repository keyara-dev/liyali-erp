# API Response Format Guide

**Phase 12C Enhancement**: Standardized Response Utilities

This guide documents the standardized API response format and response utility helpers for consistency across all endpoints.

---

## Standard Response Format

### Success Response
All successful endpoints return responses in this format:

```json
{
  "success": true,
  "message": "Optional success message",
  "data": {
    /* Response data here */
  },
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 45,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

### Error Response
All error endpoints return responses in this format:

```json
{
  "success": false,
  "message": "Human-readable error message",
  "error": "Technical error details"
}
```

### Pagination Details

**Pagination Object** (present only for list endpoints):
- `page` (integer): Current page number (1-indexed)
- `page_size` (integer): Number of items per page
- `total` (integer): Total number of items
- `total_pages` (integer): Total number of pages
- `has_next` (boolean): Whether there's a next page
- `has_prev` (boolean): Whether there's a previous page

**Pagination is NULL for non-paginated endpoints**

---

## Response Helper Functions

### Location
`backend/utils/response.go`

### Available Helpers

#### 1. SuccessResponse
Creates a success response structure

```go
response := utils.SuccessResponse(
  data interface{},           // Response data
  message string,             // Optional message
  pagination *PaginationMeta  // Optional pagination
)
```

**Example**:
```go
func GetUser(c fiber.Ctx) error {
  user := &User{}
  // ... fetch user ...

  pagination := utils.CalculatePagination(1, 10, 1)
  return utils.SendSuccess(c, fiber.StatusOK, user, "User retrieved", pagination)
}
```

#### 2. ErrorResponse
Creates an error response structure

```go
response := utils.ErrorResponse(errorMsg string)
```

#### 3. ErrorResponseWithMessage
Creates an error response with both message and error details

```go
response := utils.ErrorResponseWithMessage(
  message string,   // User-friendly message
  errorMsg string    // Technical error
)
```

#### 4. CalculatePagination
Calculates pagination metadata from total count

```go
pagination := utils.CalculatePagination(
  page int,        // Current page (1-indexed)
  pageSize int,    // Items per page
  total int64      // Total items
)
```

**Returns** `*PaginationMeta`:
- Automatically validates page >= 1
- Caps pageSize to max 100
- Calculates totalPages and has_next/has_prev

**Example**:
```go
func GetRequisitions(c fiber.Ctx) error {
  page := c.QueryInt("page", 1)
  pageSize := c.QueryInt("page_size", 10)

  // ... fetch total count and items ...

  pagination := utils.CalculatePagination(page, pageSize, total)
  return utils.SendSuccess(c, fiber.StatusOK, items, "", pagination)
}
```

#### 5. SendSuccess
Sends a complete success response with proper HTTP status

```go
utils.SendSuccess(
  c fiber.Ctx,
  statusCode int,
  data interface{},
  message string,
  pagination *PaginationMeta
)
```

**Example**:
```go
// List with pagination
return utils.SendSuccess(c, fiber.StatusOK, items, "Items retrieved", pagination)

// Create (no pagination)
return utils.SendSuccess(c, fiber.StatusCreated, newItem, "Item created", nil)

// Update (no pagination)
return utils.SendSuccess(c, fiber.StatusOK, updatedItem, "Item updated", nil)
```

#### 6. SendError
Sends an error response with HTTP status

```go
utils.SendError(
  c fiber.Ctx,
  statusCode int,
  message string,
  err error  // optional
)
```

#### 7. SendValidationError
Sends a 400 Bad Request validation error

```go
return utils.SendValidationError(c, "Email is required")
```

#### 8. SendNotFoundError
Sends a 404 Not Found error

```go
return utils.SendNotFoundError(c, "User")
```

#### 9. SendUnauthorizedError
Sends a 401 Unauthorized error

```go
return utils.SendUnauthorizedError(c, "Invalid token")
```

#### 10. SendForbiddenError
Sends a 403 Forbidden error

```go
return utils.SendForbiddenError(c, "Insufficient permissions")
```

#### 11. SendConflictError
Sends a 409 Conflict error (duplicate entry, etc.)

```go
return utils.SendConflictError(c, "Email already registered")
```

#### 12. SendInternalError
Sends a 500 Internal Server Error

```go
return utils.SendInternalError(c, "Failed to save record", err)
```

#### 13. SendUnprocessableEntityError
Sends a 422 Unprocessable Entity error (business logic violation)

```go
return utils.SendUnprocessableEntityError(c, "Cannot approve already approved document")
```

---

## HTTP Status Codes

| Code | Function | Use Case |
|------|----------|----------|
| 200 | SendSuccess | Successful GET, PUT, DELETE |
| 201 | SendSuccess | Successful POST (creation) |
| 400 | SendValidationError | Invalid input validation |
| 401 | SendUnauthorizedError | Missing/invalid authentication |
| 403 | SendForbiddenError | Insufficient permissions |
| 404 | SendNotFoundError | Resource not found |
| 409 | SendConflictError | Duplicate entry/conflict |
| 422 | SendUnprocessableEntityError | Business logic violation |
| 500 | SendInternalError | Server error |

---

## Usage Examples

### Example 1: List Endpoint with Pagination

```go
func GetRequisitions(c fiber.Ctx) error {
  page := c.QueryInt("page", 1)
  pageSize := c.QueryInt("page_size", 10)
  status := c.Query("status")

  // Build query
  query := db
  if status != "" {
    query = query.Where("status = ?", status)
  }

  // Get total count
  var total int64
  if err := query.Model(&Requisition{}).Count(&total).Error; err != nil {
    return utils.SendInternalError(c, "Failed to count requisitions", err)
  }

  // Fetch items
  var items []Requisition
  if err := query.Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error; err != nil {
    return utils.SendInternalError(c, "Failed to fetch requisitions", err)
  }

  // Calculate pagination
  pagination := utils.CalculatePagination(page, pageSize, total)

  return utils.SendSuccess(c, fiber.StatusOK, items, "Requisitions retrieved", pagination)
}
```

**Response**:
```json
{
  "success": true,
  "message": "Requisitions retrieved",
  "data": [
    { "id": "req-1", "title": "..." },
    { "id": "req-2", "title": "..." }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 45,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

### Example 2: Create Endpoint (No Pagination)

```go
func CreateRequisition(c fiber.Ctx) error {
  var req CreateRequisitionRequest

  if err := c.BindJSON(&req); err != nil {
    return utils.SendValidationError(c, "Invalid request body")
  }

  if req.Title == "" {
    return utils.SendValidationError(c, "Title is required")
  }

  // Create item
  item := &Requisition{Title: req.Title}
  if err := db.Create(item).Error; err != nil {
    return utils.SendInternalError(c, "Failed to create requisition", err)
  }

  return utils.SendSuccess(c, fiber.StatusCreated, item, "Requisition created", nil)
}
```

**Response**:
```json
{
  "success": true,
  "message": "Requisition created",
  "data": {
    "id": "req-123",
    "title": "Office Supplies",
    "status": "draft",
    "createdAt": "2025-12-22T21:30:00Z"
  },
  "pagination": null
}
```

### Example 3: Error Handling

```go
func GetRequisition(c fiber.Ctx) error {
  id := c.Params("id")

  if id == "" {
    return utils.SendValidationError(c, "Requisition ID is required")
  }

  var item Requisition
  if err := db.Where("id = ?", id).First(&item).Error; err != nil {
    return utils.SendNotFoundError(c, "Requisition")
  }

  return utils.SendSuccess(c, fiber.StatusOK, item, "Requisition retrieved", nil)
}
```

**404 Response**:
```json
{
  "success": false,
  "message": "Requisition not found",
  "error": "record not found"
}
```

### Example 4: Business Logic Error

```go
func ApproveRequisition(c fiber.Ctx) error {
  var req ApproveRequest
  if err := c.BindJSON(&req); err != nil {
    return utils.SendValidationError(c, "Invalid request body")
  }

  item := &Requisition{}
  // ... fetch item ...

  if item.Status != "pending" {
    return utils.SendUnprocessableEntityError(
      c,
      fmt.Sprintf("Cannot approve requisition in %s status", item.Status),
    )
  }

  // ... approve logic ...

  return utils.SendSuccess(c, fiber.StatusOK, item, "Requisition approved", nil)
}
```

**422 Response**:
```json
{
  "success": false,
  "message": "Cannot approve requisition in rejected status",
  "error": ""
}
```

---

## Query Parameters

### Pagination Parameters
- `page` (default: 1) - Page number (1-indexed)
- `page_size` (default: 10, max: 100) - Items per page

### Common Filter Parameters
- `status` - Filter by document status (draft, pending, approved, rejected)
- `department` - Filter by department
- `priority` - Filter by priority (low, medium, high)
- `vendorId` - Filter by vendor
- `fiscalYear` - Filter by fiscal year

**Example Query**:
```
GET /api/v1/requisitions?page=2&page_size=20&status=pending&department=IT
```

---

## Pagination Math

The pagination helper automatically calculates:

```
offset = (page - 1) * page_size
total_pages = ceil(total / page_size)
has_next = page < total_pages
has_prev = page > 1
```

---

## Response Type Struct

```go
type APIResponse struct {
  Success    bool            `json:"success"`
  Message    string          `json:"message,omitempty"`
  Data       interface{}     `json:"data,omitempty"`
  Error      string          `json:"error,omitempty"`
  Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type PaginationMeta struct {
  Page       int   `json:"page"`
  PageSize   int   `json:"page_size"`
  Total      int64 `json:"total"`
  TotalPages int64 `json:"total_pages"`
  HasNext    bool  `json:"has_next"`
  HasPrev    bool  `json:"has_prev"`
}
```

---

## Migration from Old Format

If updating existing handlers to use the new format:

**Old Format**:
```json
{
  "success": true,
  "data": [...],
  "total": 100,
  "page": 1,
  "limit": 10
}
```

**New Format**:
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_pages": 10,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## Best Practices

1. **Always use helpers** - Never manually construct response JSON
2. **Consistent pagination** - Use `CalculatePagination` for all list endpoints
3. **Appropriate status codes** - Use the correct HTTP status code helper
4. **Clear messages** - Messages should be user-friendly, errors should be technical
5. **Null pagination** - Set pagination to `nil` for non-paginated endpoints
6. **Error handling** - Use specific error helpers for different scenarios

---

## Testing Pagination

```bash
# First page
curl "http://localhost:8080/api/v1/requisitions?page=1&page_size=10"

# Second page with filters
curl "http://localhost:8080/api/v1/requisitions?page=2&page_size=10&status=pending"

# Last page
curl "http://localhost:8080/api/v1/requisitions?page=5&page_size=10"
```

---

**Last Updated**: December 22, 2025
**Status**: Response Utilities Complete
**Files**: `backend/utils/response.go`
