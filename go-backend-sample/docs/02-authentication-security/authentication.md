# Authentication Flow

**JWT-based authentication with email and password**

---

## Overview

Liyali Gateway uses JWT (JSON Web Tokens) for stateless authentication. Users log in with email and password, receive access and refresh tokens, and use them to access protected endpoints.

---

## Authentication Flow

```
┌─────────┐                 ┌─────────┐                 ┌──────────┐
│  Client │                 │   API   │                 │ Database │
└────┬────┘                 └────┬────┘                 └────┬─────┘
     │                           │                           │
     │ POST /api/auth/register   │                           │
     │ {email, password, ...}    │                           │
     ├──────────────────────────>│                           │
     │                           │ Hash password (bcrypt)    │
     │                           │                           │
     │                           │ INSERT user               │
     │                           ├──────────────────────────>│
     │                           │                           │
     │                           │<──────────────────────────┤
     │ {user}                    │                           │
     │<──────────────────────────┤                           │
     │                           │                           │
     │ POST /api/auth/login      │                           │
     │ {email, password}         │                           │
     ├──────────────────────────>│                           │
     │                           │ SELECT user BY email      │
     │                           ├──────────────────────────>│
     │                           │                           │
     │                           │<──────────────────────────┤
     │                           │ Verify password (bcrypt)  │
     │                           │                           │
     │                           │ Generate JWT tokens       │
     │                           │ - Access (1 hour)         │
     │                           │ - Refresh (8 hours)       │
     │                           │                           │
     │                           │ INSERT session            │
     │                           ├──────────────────────────>│
     │                           │                           │
     │ {access_token,            │                           │
     │  refresh_token, user}     │                           │
     │<──────────────────────────┤                           │
     │                           │                           │
     │ GET /api/auth/me          │                           │
     │ Authorization: Bearer ... │                           │
     ├──────────────────────────>│                           │
     │                           │ Validate JWT              │
     │                           │ Extract user ID           │
     │                           │                           │
     │ {id, email, role}         │                           │
     │<──────────────────────────┤                           │
```

---

## Token Types

### Access Token
- **Purpose**: Short-lived token for API requests
- **Expiration**: 1 hour
- **Contains**: User ID, email, role
- **Storage**: Memory or sessionStorage (not localStorage)
- **Use**: Every API request in Authorization header

### Refresh Token
- **Purpose**: Long-lived token to get new access tokens
- **Expiration**: 8 hours
- **Contains**: Random secure token
- **Storage**: Database (sessions table)
- **Use**: Refresh endpoint when access token expires

---

## JWT Structure

### Access Token Payload

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "DEPARTMENT_MANAGER",
  "exp": 1703088000,
  "iat": 1703084400
}
```

### Token Claims

- `user_id` (uuid) - User's unique identifier
- `email` (string) - User's email address
- `role` (string) - User's role for RBAC
- `exp` (int64) - Expiration time (Unix timestamp)
- `iat` (int64) - Issued at time (Unix timestamp)

---

## API Endpoints

### Register

**Endpoint**: `POST /api/auth/register`

**Request**:
```json
{
  "email": "john.doe@example.com",
  "password": "SecurePass123!",
  "name": "John Doe",
  "role": "DEPARTMENT_MANAGER",
  "department": "Finance"
}
```

**Response** (201 Created):
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "DEPARTMENT_MANAGER"
  }
}
```

**Validation**:
- Email must be valid format
- Password minimum 8 characters
- Role must be one of 7 valid roles
- Email must be unique

---

### Login

**Endpoint**: `POST /api/auth/login`

**Request**:
```json
{
  "email": "john.doe@example.com",
  "password": "SecurePass123!"
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "DEPARTMENT_MANAGER"
  }
}
```

**Failed Login** (401 Unauthorized):
```json
{
  "error": "Invalid email or password"
}
```

**Account Locked** (403 Forbidden):
```json
{
  "error": "Account is locked due to too many failed login attempts"
}
```

---

### Refresh Token

**Endpoint**: `POST /api/auth/refresh`

**Request**:
```json
{
  "refresh_token": "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6"
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Invalid Token** (401 Unauthorized):
```json
{
  "error": "Invalid or expired refresh token"
}
```

---

### Logout

**Endpoint**: `POST /api/auth/logout`

**Request**:
```json
{
  "refresh_token": "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6"
}
```

**Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

---

### Get Current User

**Endpoint**: `GET /api/auth/me`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john.doe@example.com",
  "role": "DEPARTMENT_MANAGER"
}
```

**Unauthorized** (401):
```json
{
  "error": "User not authenticated"
}
```

---

## Implementation Details

### Password Hashing

```go
// Hash password with bcrypt (cost 12)
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(password),
    bcrypt.DefaultCost, // Cost 12
)

// Verify password
err := bcrypt.CompareHashAndPassword(
    []byte(user.PasswordHash),
    []byte(password),
)
```

### JWT Generation

```go
// Create access token (1 hour)
token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
    UserID: userID,
    Email:  email,
    Role:   role,
    RegisteredClaims: jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    },
})

tokenString, err := token.SignedString(jwtSecret)
```

### JWT Validation

```go
// Parse and validate token
token, err := jwt.ParseWithClaims(
    tokenString,
    &JWTClaims{},
    func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    },
)

// Extract claims
claims := token.Claims.(*JWTClaims)
userID := claims.UserID
role := claims.Role
```

---

## Security Features

### Account Lockout
- **Trigger**: 5 failed login attempts
- **Duration**: 15 minutes
- **Reset**: Successful login or time expiration

### Password Requirements
- Minimum 8 characters
- Required in registration and password change

### Token Security
- Access tokens short-lived (1 hour)
- Refresh tokens stored in database
- Tokens invalidated on logout
- HMAC-SHA256 algorithm (HS256)

---

## Error Handling

| Status Code | Error | Cause |
|-------------|-------|-------|
| 400 | Invalid request body | Malformed JSON |
| 400 | Validation error | Missing/invalid fields |
| 401 | Invalid email or password | Wrong credentials |
| 401 | Invalid or expired token | Token validation failed |
| 403 | Account is locked | Too many failed attempts |
| 403 | Account is inactive | User deactivated |
| 409 | Email already exists | Duplicate registration |

---

## Testing

### Unit Tests

```go
func TestLogin_Success(t *testing.T) {
    // Create mock user with hashed password
    hashedPassword, _ := bcrypt.GenerateFromPassword(
        []byte("password123"),
        bcrypt.DefaultCost,
    )

    mockUser := &db.User{
        ID:           uuid.New(),
        Email:        "test@example.com",
        PasswordHash: string(hashedPassword),
        IsActive:     true,
    }

    // Mock repository
    userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(mockUser, nil)

    // Execute login
    accessToken, refreshToken, user, err := authService.Login(
        ctx, "test@example.com", "password123", "127.0.0.1", "test-agent",
    )

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, accessToken)
    assert.NotEmpty(t, refreshToken)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Integration Tests

```go
func TestLoginEndpoint_Success(t *testing.T) {
    // Setup test app with database
    app := setupTestApp(t)
    defer teardownTestApp()

    // Register user first
    registerUser(t, app, "test@example.com", "password123")

    // Login
    payload := map[string]interface{}{
        "email":    "test@example.com",
        "password": "password123",
    }
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    require.NoError(t, err)

    // Assert
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var response map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&response)
    assert.NotEmpty(t, response["access_token"])
    assert.NotEmpty(t, response["refresh_token"])
}
```

---

## Related Pages

- [RBAC System](./rbac.md) - Roles and permissions
- [Password Security](./passwords.md) - Reset and change password
- [Session Management](./sessions.md) - Token lifecycle
- [Security Best Practices](./security.md) - Rate limiting, CSRF

---

**Files**:
- `internal/services/auth_service.go` - Auth business logic
- `internal/handlers/auth_handler.go` - HTTP handlers
- `internal/middleware/auth_middleware.go` - JWT validation
- `tests/unit/auth_service_test.go` - Unit tests
- `tests/integration/auth_integration_test.go` - Integration tests

**Last Updated**: December 25, 2025
