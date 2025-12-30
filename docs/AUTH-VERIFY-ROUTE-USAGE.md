# Authentication Verify Route Usage Guide

## Route: `POST /api/v1/auth/verify`

### Purpose
The `/auth/verify` route is used to validate JWT access tokens without requiring authentication middleware. This is a **public route** that allows clients to verify if their tokens are still valid.

### When to Use This Route

#### 1. **Client-Side Token Validation**
- **Frontend Applications**: Before making authenticated requests, check if the current token is still valid
- **Mobile Apps**: Validate tokens after app resume or network reconnection
- **SPA Applications**: Check token validity on page refresh or route changes

#### 2. **Token Expiry Checking**
- **Proactive Refresh**: Check if token is about to expire before making important requests
- **Background Validation**: Periodically validate tokens in background processes
- **Session Management**: Determine if user needs to re-authenticate

#### 3. **Third-Party Integration**
- **API Gateways**: Validate tokens from external services
- **Microservices**: Verify tokens received from other services
- **Webhook Validation**: Validate tokens in webhook payloads

#### 4. **Security Validation**
- **Token Integrity**: Ensure token hasn't been tampered with
- **Blacklist Checking**: Verify token hasn't been revoked (if blacklisting is implemented)
- **Claims Validation**: Extract and validate token claims without authentication

### Request Format
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Response Format

#### Success Response (200)
```json
{
  "success": true,
  "message": "Token is valid",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "role": "admin",
    "organization_id": "456e7890-e89b-12d3-a456-426614174001",
    "expires_at": 1640995200
  }
}
```

#### Error Response (401)
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

### Use Cases Examples

#### Frontend JavaScript
```javascript
// Check token validity before making requests
async function isTokenValid(token) {
  try {
    const response = await fetch('/api/v1/auth/verify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token })
    });
    
    if (response.ok) {
      const data = await response.json();
      return data.success;
    }
    return false;
  } catch (error) {
    return false;
  }
}

// Use before important operations
if (await isTokenValid(localStorage.getItem('accessToken'))) {
  // Proceed with authenticated request
} else {
  // Refresh token or redirect to login
}
```

#### Mobile App (React Native)
```javascript
// Check token on app resume
useEffect(() => {
  const handleAppStateChange = async (nextAppState) => {
    if (nextAppState === 'active') {
      const token = await AsyncStorage.getItem('accessToken');
      if (token && !(await isTokenValid(token))) {
        // Token expired, refresh or logout
        await refreshTokenOrLogout();
      }
    }
  };

  AppState.addEventListener('change', handleAppStateChange);
  return () => AppState.removeEventListener('change', handleAppStateChange);
}, []);
```

#### API Gateway Integration
```javascript
// Validate incoming tokens
app.use('/protected', async (req, res, next) => {
  const token = req.headers.authorization?.replace('Bearer ', '');
  
  if (!token) {
    return res.status(401).json({ error: 'No token provided' });
  }
  
  const isValid = await verifyTokenWithAuthService(token);
  if (!isValid) {
    return res.status(401).json({ error: 'Invalid token' });
  }
  
  next();
});
```

### Important Notes

1. **Public Route**: This route doesn't require authentication, making it accessible for token validation
2. **No Side Effects**: Only validates the token, doesn't modify any state
3. **Security**: Returns token claims for convenience but doesn't expose sensitive information
4. **Rate Limiting**: Consider implementing rate limiting to prevent abuse
5. **Caching**: Results can be cached briefly to improve performance

### Alternative Approaches

Instead of using this route, you could:
1. **Decode JWT Client-Side**: Check expiry time locally (less secure)
2. **Try Authenticated Request**: Make the actual request and handle 401 errors
3. **Use Refresh Token**: Always try to refresh before making requests

### Best Practices

1. **Use Sparingly**: Don't verify tokens before every request
2. **Cache Results**: Cache validation results for a short period
3. **Handle Errors Gracefully**: Always have fallback logic for invalid tokens
4. **Combine with Refresh**: Use with token refresh logic for seamless UX
5. **Monitor Usage**: Track usage patterns to detect potential security issues

This route provides a clean way to validate tokens without the overhead of full authentication middleware, making it ideal for client-side token management and third-party integrations.