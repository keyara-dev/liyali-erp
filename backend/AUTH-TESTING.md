# Authentication Testing Guide

**Phase 12B: Authentication Implementation**

## Overview

This guide provides detailed instructions for testing the authentication endpoints with cURL or Postman.

## API Endpoints

### 1. Register a New User

**Endpoint**: `POST /api/v1/auth/register`
**Authentication**: Public (no token required)

**Request Body**:
```json
{
  "email": "newuser@liyali.com",
  "password": "SecurePass123",
  "name": "New User",
  "role": "requester"
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@liyali.com",
    "password": "SecurePass123",
    "name": "New User",
    "role": "requester"
  }'
```

**Success Response (201)**:
```json
{
  "success": true,
  "message": "Registration successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user-uuid-here",
    "email": "newuser@liyali.com",
    "name": "New User",
    "role": "requester",
    "active": true,
    "createdAt": "2025-12-22T20:45:00Z"
  }
}
```

**Error Response (400)**:
```json
{
  "success": false,
  "message": "Password does not meet requirements",
  "error": "password must contain at least one uppercase letter"
}
```

### 2. Login with Existing User

**Endpoint**: `POST /api/v1/auth/login`
**Authentication**: Public (no token required)

**Test Users** (seeded automatically):
- **Admin**: admin@liyali.com
- **Approver**: approver@liyali.com
- **Requester**: requester@liyali.com
- **Finance**: finance@liyali.com
- **Viewer**: viewer@liyali.com

**Request Body**:
```json
{
  "email": "admin@liyali.com",
  "password": "any_password"
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@liyali.com",
    "password": "any_password"
  }'
```

**Success Response (200)**:
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user-admin-001",
    "email": "admin@liyali.com",
    "name": "Admin User",
    "role": "admin",
    "active": true,
    "createdAt": "2025-12-22T20:45:00Z"
  }
}
```

**Error Response (401)**:
```json
{
  "success": false,
  "message": "Invalid email or password"
}
```

### 3. Get User Profile

**Endpoint**: `GET /api/v1/auth/profile`
**Authentication**: Required (JWT token)

**cURL Example**:
```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer <your_token_here>"
```

**Success Response (200)**:
```json
{
  "success": true,
  "user": {
    "id": "user-admin-001",
    "email": "admin@liyali.com",
    "name": "Admin User",
    "role": "admin",
    "active": true,
    "createdAt": "2025-12-22T20:45:00Z"
  }
}
```

**Error Response (401)**:
```json
{
  "error": "Authorization header required"
}
```

### 4. Verify Token

**Endpoint**: `POST /api/v1/auth/verify`
**Authentication**: Public (no token required)

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Success Response (200)**:
```json
{
  "valid": true,
  "user": {
    "id": "user-admin-001",
    "email": "admin@liyali.com",
    "name": "Admin User",
    "role": "admin",
    "active": true,
    "createdAt": "2025-12-22T20:45:00Z"
  }
}
```

**Error Response (401)**:
```json
{
  "valid": false,
  "error": "invalid token"
}
```

### 5. Refresh Token

**Endpoint**: `POST /api/v1/auth/refresh`
**Authentication**: Public (no token required)

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Success Response (200)**:
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## Testing Workflow

### Step 1: Register New User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123",
    "name": "Test User",
    "role": "requester"
  }'
```

Save the returned `token` for subsequent requests.

### Step 2: Login with Credentials
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123"
  }'
```

### Step 3: Access Protected Endpoint
```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Step 4: Verify Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Step 5: Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

---

## Test Users (Pre-seeded)

| Email | Password | Role | Purpose |
|-------|----------|------|---------|
| admin@liyali.com | any | admin | Full system access |
| approver@liyali.com | any | approver | Can approve documents |
| requester@liyali.com | any | requester | Can create requisitions |
| finance@liyali.com | any | finance | Finance operations |
| viewer@liyali.com | any | viewer | Read-only access |

> **Note**: Pre-seeded users accept any password for demo purposes.

---

## Password Requirements

New user passwords must meet these criteria:
- **Minimum 8 characters**
- **At least one uppercase letter** (A-Z)
- **At least one lowercase letter** (a-z)
- **At least one digit** (0-9)

**Valid Examples**:
- `SecurePass123`
- `MyPassword456`
- `TestPass789`

**Invalid Examples**:
- `password` (no uppercase, no digits)
- `UPPERCASE` (no lowercase, no digits)
- `Pass` (too short)
- `Pass123` (length = 7, need 8)

---

## JWT Token Structure

The generated JWT token contains the following claims:

```json
{
  "sub": "user-id-here",
  "email": "user@example.com",
  "name": "User Name",
  "role": "requester",
  "exp": 1703350800,
  "iat": 1703264400,
  "nbf": 1703264400,
  "iss": "liyali-gateway"
}
```

**Token Expiration**: 24 hours from creation

---

## Common Issues & Solutions

### Issue: "Invalid authorization header format"
**Problem**: Authorization header not in correct format
**Solution**: Use format: `Authorization: Bearer <token>`

### Issue: "User not found"
**Problem**: Email doesn't exist in database
**Solution**: Register first or use pre-seeded test users

### Issue: "Password does not meet requirements"
**Problem**: New password too weak
**Solution**: Use password with 8+ chars, uppercase, lowercase, digits

### Issue: "Email already registered"
**Problem**: Trying to register with existing email
**Solution**: Use different email or login instead

### Issue: "User account is inactive"
**Problem**: User account disabled
**Solution**: Check user status in database

---

## Postman Collection

### Setup in Postman

1. **Create Environment Variables**:
   - `base_url`: `http://localhost:8080`
   - `token`: (leave blank, will be set by tests)

2. **Create Login Request**:
   ```
   POST {{base_url}}/api/v1/auth/login

   Body (JSON):
   {
     "email": "admin@liyali.com",
     "password": "any"
   }

   Tests:
   pm.environment.set("token", pm.response.json().token);
   ```

3. **Create Profile Request**:
   ```
   GET {{base_url}}/api/v1/auth/profile

   Headers:
   Authorization: Bearer {{token}}
   ```

---

## Security Notes

### Development Only
- Pre-seeded users accept any password
- JWT_SECRET is basic (change in production)
- No rate limiting on endpoints
- CORS allows all origins (configure in production)

### Production Requirements
- Use strong JWT_SECRET (min 32 characters)
- Implement rate limiting (prevent brute force)
- Use HTTPS only
- Configure CORS properly
- Store passwords hashed in database
- Implement refresh token rotation
- Add email verification
- Implement 2FA
- Monitor failed login attempts

---

## Next Steps

1. ✅ Test all authentication endpoints
2. ⬜ Implement CRUD handlers for documents (Phase 12C)
3. ⬜ Add audit logging for authentication
4. ⬜ Implement role-based access control
5. ⬜ Add email notifications

---

**Last Updated**: December 22, 2025
**Status**: Authentication Implementation Complete
