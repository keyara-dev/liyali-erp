# Authentication System

The frontend implements a comprehensive authentication system with JWT-based sessions, role-based access control, and multi-tenancy support.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Authentication Flow                      │
├─────────────────────────────────────────────────────────────┤
│ 1. User Login → Server Action                              │
│ 2. JWT Token Generation → HTTP-Only Cookie                 │
│ 3. Session Verification → Middleware                       │
│ 4. Role-Based Access → Route Protection                    │
│ 5. Organization Context → Multi-Tenancy                    │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### JWT Token Management

The system uses server-only JWT operations for security:

```typescript
// src/lib/auth.ts
import "server-only";
import { SignJWT, jwtVerify } from "jose";

const getSecretKey = () => {
  const secretKey = process.env.AUTH_SECRET;
  if (!secretKey || secretKey.length < 32) {
    throw new Error("AUTH_SECRET must be at least 32 characters");
  }
  return secretKey;
};

export async function encrypt(payload: any, expirationTime: string = "1h") {
  const key = new TextEncoder().encode(getSecretKey());
  return new SignJWT(payload)
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(expirationTime)
    .sign(key);
}

export async function decrypt(token: string) {
  try {
    const key = new TextEncoder().encode(getSecretKey());
    const { payload } = await jwtVerify(token, key, {
      algorithms: ["HS256"],
      clockTolerance: 15,
    });
    return payload;
  } catch (error) {
    return {
      success: false,
      message: "Token verification failed",
      status: 500,
    };
  }
}
```

### Session Management

Sessions are managed through HTTP-only cookies with automatic expiration:

```typescript
export async function createAuthSession({
  access_token,
  role,
  user_id,
  organization_id,
}: {
  access_token: string;
  role: UserType;
  user_id?: string;
  organization_id?: string;
}): Promise<void> {
  const expiresAt = new Date(Date.now() + SESSION_CONFIG.SESSION_TTL);

  const newSession: AuthSession = {
    access_token,
    role,
    user_id,
    organization_id,
    expiresAt,
  };

  const token = await encrypt(newSession, "30m");

  (await cookies()).set(AUTH_SESSION, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    expires: expiresAt,
    sameSite: "strict",
    path: "/",
  });
}
```

### Session Verification

All protected routes verify sessions through middleware:

```typescript
export async function verifySession(): Promise<{
  isAuthenticated: boolean;
  session: AuthSession | null;
  permissions?: any[];
}> {
  try {
    const cookieStore = await cookies();
    const cookie = cookieStore.get(AUTH_SESSION)?.value;

    if (!cookie) {
      return { isAuthenticated: false, session: null };
    }

    const decrypted = await decrypt(cookie);

    if (!decrypted || decrypted.success === false) {
      await deleteSession();
      return { isAuthenticated: false, session: null };
    }

    const session = decrypted as unknown as AuthSession;

    // Check expiration
    if (session?.expiresAt) {
      const expiresAt = new Date(session.expiresAt);
      const now = new Date();

      if (expiresAt < now) {
        await deleteSession();
        return { isAuthenticated: false, session: null };
      }
    }

    return {
      isAuthenticated: true,
      session: session,
      role: session.role,
    };
  } catch (error) {
    console.error("[verifySession] Error:", error);
    return { isAuthenticated: false, session: null };
  }
}
```

## User Roles and Permissions

### Role Hierarchy

```typescript
export type UserRole =
  | "REQUESTER"           // Can create and submit requisitions
  | "DEPARTMENT_MANAGER"  // Can approve department requisitions
  | "FINANCE_OFFICER"     // Can handle financial approvals
  | "DIRECTOR"            // Can approve high-value items
  | "CFO"                 // Final approval authority
  | "COMPLIANCE_OFFICER"  // Can review compliance aspects
  | "ADMIN";              // System administration

// Role-based access control
export async function hasRole(
  requiredRole: UserRole | UserRole[]
): Promise<boolean> {
  const user = await getCurrentUser();
  if (!user) return false;

  const roles = Array.isArray(requiredRole) ? requiredRole : [requiredRole];
  return roles.includes(user.role);
}

export async function isAdmin(): Promise<boolean> {
  const user = await getCurrentUser();
  return user?.role === "ADMIN";
}
```

### Permission System

Permissions are derived from roles and organizational context:

```typescript
// Permission checking patterns
const canApproveRequisition = await hasRole([
  "DEPARTMENT_MANAGER",
  "FINANCE_OFFICER", 
  "DIRECTOR",
  "CFO"
]);

const canCreatePurchaseOrder = await hasRole([
  "FINANCE_OFFICER",
  "ADMIN"
]);

const canViewAllDocuments = await hasRole([
  "FINANCE_OFFICER",
  "DIRECTOR",
  "CFO",
  "ADMIN"
]);
```

## Multi-Tenancy Support

### Organization Context

The system supports multiple organizations through React Context:

```typescript
// src/contexts/organization-context.tsx
export interface OrganizationContextType {
  currentOrganization: Organization | null;
  userOrganizations: Organization[];
  switchWorkspace: (orgId: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
  refreshOrganizations: () => void;
}

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [currentOrgId, setCurrentOrgId] = useState<string | null>(null);

  // Fetch user's organizations
  const { data: organizations = [], isLoading, error, refetch } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => fetchUserOrganizations(),
  });

  const switchWorkspace = async (orgId: string) => {
    await switchMutation.mutateAsync(orgId);
    // Invalidate all queries to refetch with new org context
    queryClient.invalidateQueries();
  };

  return (
    <OrganizationContext.Provider value={{
      currentOrganization,
      userOrganizations: organizations,
      switchWorkspace,
      isLoading,
      error: error?.message || null,
      refreshOrganizations: () => refetch(),
    }}>
      {children}
    </OrganizationContext.Provider>
  );
}
```

### Organization Switching

Users can switch between organizations they have access to:

```typescript
// Usage in components
const { currentOrganization, switchWorkspace } = useOrganizationContext();

const handleOrgSwitch = async (orgId: string) => {
  await switchWorkspace(orgId);
  // All data is automatically refetched for the new organization
};
```

## Screen Lock System

### Idle Detection

The system implements automatic screen locking for security:

```typescript
// src/lib/auth.ts
export async function setScreenLockCookie(isLocked: boolean): Promise<void> {
  const expiresAt = new Date(Date.now() + SESSION_CONFIG.SCREEN_LOCK_COUNTDOWN);

  const lockState = {
    locked: isLocked,
    timestamp: new Date().toISOString(),
  };

  const token = await encrypt(lockState, "90s");

  if (token) {
    const cookieStore = await cookies();
    cookieStore.set(SCREEN_LOCK_SESSION, token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      expires: expiresAt,
      sameSite: "strict",
      path: "/",
    });
  }
}
```

### Idle Timer Component

```typescript
// src/components/base/screen-lock.tsx
export function IdleTimerContainer({ children }: { children: ReactNode }) {
  const [isLocked, setIsLocked] = useState(false);
  const [countdown, setCountdown] = useState(0);

  useEffect(() => {
    const timer = new IdleTimer({
      timeout: SESSION_CONFIG.IDLE_TIMEOUT,
      onIdle: () => {
        setCountdown(SESSION_CONFIG.SCREEN_LOCK_COUNTDOWN / 1000);
        // Start countdown before locking
      },
      onActive: () => {
        setCountdown(0);
        setIsLocked(false);
      },
    });

    return () => timer.destroy();
  }, []);

  if (isLocked) {
    return <ScreenLockModal onUnlock={() => setIsLocked(false)} />;
  }

  return <>{children}</>;
}
```

## Route Protection

### Middleware Protection

```typescript
// middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { verifySession } from '@/lib/auth';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Public routes that don't require authentication
  const publicRoutes = ['/login', '/signup', '/forgot-password'];
  
  if (publicRoutes.includes(pathname)) {
    return NextResponse.next();
  }

  // Verify session for protected routes
  const { isAuthenticated, session } = await verifySession();

  if (!isAuthenticated) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  // Role-based route protection
  if (pathname.startsWith('/admin') && session?.role !== 'ADMIN') {
    return NextResponse.redirect(new URL('/unauthorized', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};
```

### Component-Level Protection

```typescript
// Higher-order component for role protection
export function withRoleProtection<P extends object>(
  Component: React.ComponentType<P>,
  requiredRoles: UserRole[]
) {
  return function ProtectedComponent(props: P) {
    const [hasAccess, setHasAccess] = useState(false);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
      const checkAccess = async () => {
        const access = await hasRole(requiredRoles);
        setHasAccess(access);
        setLoading(false);
      };

      checkAccess();
    }, []);

    if (loading) {
      return <div>Loading...</div>;
    }

    if (!hasAccess) {
      return <div>Access denied</div>;
    }

    return <Component {...props} />;
  };
}

// Usage
const AdminPanel = withRoleProtection(AdminPanelComponent, ['ADMIN']);
```

## Authentication Hooks

### useAuth Hook

```typescript
// src/hooks/use-auth.ts
export function useAuth() {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const currentUser = await getCurrentUser();
        setUser(currentUser);
      } catch (error) {
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = async (credentials: LoginCredentials) => {
    // Handle login logic
  };

  const logout = async () => {
    await deleteSession();
    setUser(null);
    window.location.href = '/login';
  };

  return {
    user,
    loading,
    isAuthenticated: !!user,
    login,
    logout,
  };
}
```

### usePermissions Hook

```typescript
export function usePermissions() {
  const { user } = useAuth();

  const can = useCallback((permission: string) => {
    if (!user) return false;
    
    // Check role-based permissions
    const rolePermissions = getRolePermissions(user.role);
    return rolePermissions.includes(permission);
  }, [user]);

  const canApprove = useCallback((documentType: string, amount?: number) => {
    if (!user) return false;
    
    // Complex approval logic based on role and amount
    return checkApprovalPermission(user.role, documentType, amount);
  }, [user]);

  return {
    can,
    canApprove,
    role: user?.role,
    isAdmin: user?.role === 'ADMIN',
  };
}
```

## Security Features

### CSRF Protection

```typescript
// CSRF token generation and validation
export function generateCSRFToken(): string {
  return crypto.randomUUID();
}

export function validateCSRFToken(token: string, sessionToken: string): boolean {
  // Validate CSRF token against session
  return token === sessionToken;
}
```

### Rate Limiting

```typescript
// Rate limiting for authentication attempts
const loginAttempts = new Map<string, { count: number; lastAttempt: Date }>();

export function checkRateLimit(identifier: string): boolean {
  const attempts = loginAttempts.get(identifier);
  const now = new Date();

  if (!attempts) {
    loginAttempts.set(identifier, { count: 1, lastAttempt: now });
    return true;
  }

  // Reset if more than 15 minutes have passed
  if (now.getTime() - attempts.lastAttempt.getTime() > 15 * 60 * 1000) {
    loginAttempts.set(identifier, { count: 1, lastAttempt: now });
    return true;
  }

  // Block if more than 5 attempts
  if (attempts.count >= 5) {
    return false;
  }

  attempts.count++;
  attempts.lastAttempt = now;
  return true;
}
```

### Password Security

```typescript
// Password validation and hashing
export function validatePassword(password: string): {
  isValid: boolean;
  errors: string[];
} {
  const errors: string[] = [];

  if (password.length < 8) {
    errors.push('Password must be at least 8 characters long');
  }

  if (!/[A-Z]/.test(password)) {
    errors.push('Password must contain at least one uppercase letter');
  }

  if (!/[a-z]/.test(password)) {
    errors.push('Password must contain at least one lowercase letter');
  }

  if (!/\d/.test(password)) {
    errors.push('Password must contain at least one number');
  }

  if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
    errors.push('Password must contain at least one special character');
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}
```

## Session Configuration

### Session Settings

```typescript
// src/lib/session-config.ts
export const SESSION_CONFIG = {
  SESSION_TTL: 30 * 60 * 1000,        // 30 minutes
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,   // 90 seconds
  IDLE_TIMEOUT: 25 * 60 * 1000,       // 25 minutes
  REFRESH_THRESHOLD: 5 * 60 * 1000,   // 5 minutes before expiry
} as const;
```

### Cookie Configuration

```typescript
const COOKIE_OPTIONS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "strict" as const,
  path: "/",
  maxAge: SESSION_CONFIG.SESSION_TTL / 1000, // Convert to seconds
};
```

## Error Handling

### Authentication Errors

```typescript
export enum AuthError {
  INVALID_CREDENTIALS = 'INVALID_CREDENTIALS',
  SESSION_EXPIRED = 'SESSION_EXPIRED',
  INSUFFICIENT_PERMISSIONS = 'INSUFFICIENT_PERMISSIONS',
  ACCOUNT_LOCKED = 'ACCOUNT_LOCKED',
  MFA_REQUIRED = 'MFA_REQUIRED',
}

export function handleAuthError(error: AuthError): string {
  switch (error) {
    case AuthError.INVALID_CREDENTIALS:
      return 'Invalid email or password';
    case AuthError.SESSION_EXPIRED:
      return 'Your session has expired. Please log in again.';
    case AuthError.INSUFFICIENT_PERMISSIONS:
      return 'You do not have permission to access this resource';
    case AuthError.ACCOUNT_LOCKED:
      return 'Your account has been locked due to too many failed attempts';
    case AuthError.MFA_REQUIRED:
      return 'Multi-factor authentication is required';
    default:
      return 'An authentication error occurred';
  }
}
```

## Best Practices

### Security Guidelines

1. **Never store sensitive data in localStorage**
2. **Use HTTP-only cookies for session tokens**
3. **Implement proper CSRF protection**
4. **Validate all user inputs**
5. **Use secure password policies**
6. **Implement rate limiting**
7. **Log security events**
8. **Regular security audits**

### Development Guidelines

1. **Always verify sessions server-side**
2. **Use TypeScript for type safety**
3. **Implement proper error handling**
4. **Test authentication flows thoroughly**
5. **Document permission requirements**
6. **Use consistent naming conventions**
7. **Implement proper logging**

The authentication system provides a secure, scalable foundation for the application while maintaining good user experience through features like automatic session management and organization switching.