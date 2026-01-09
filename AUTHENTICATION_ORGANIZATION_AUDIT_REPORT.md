# Authentication & Organization Selection Audit Report

## Executive Summary

This audit identified **5 critical race conditions** and **3 architectural issues** in the authentication and organization selection system that are causing:

1. **Session verification failures** during organization fetching
2. **Redirect loops** to `/welcome` page
3. **Organization context loss** during switching
4. **Inconsistent state** between frontend and backend

## Critical Issues Identified

### 🚨 **RACE CONDITION #1: Login → Organization Fetch Timing**

**Issue**: After login, user is redirected to `/welcome` before session cookie is fully set and readable.

**Location**: `frontend/src/hooks/use-auth-mutations.ts:32-35`

```typescript
onSuccess: async (data) => {
  if (data.success) {
    setIsRedirecting(true);
    // Only 100ms delay - insufficient for cookie propagation
    await new Promise((resolve) => setTimeout(resolve, 100));
    router.push("/welcome");
  }
};
```

**Impact**: `OrganizationProvider` query fails with "No valid session found" error.

**Root Cause**: 100ms delay is insufficient for httpOnly cookie to be set and readable by subsequent requests.

---

### 🚨 **RACE CONDITION #2: Organization Context Initialization**

**Issue**: Multiple `useEffect` hooks in `OrganizationProvider` create race conditions during initialization.

**Location**: `frontend/src/contexts/organization-context.tsx:45-65`

```typescript
// Effect 1: Client-side check
useEffect(() => {
  setIsClient(true);
}, []);

// Effect 2: Organization selection (depends on Effect 1)
useEffect(() => {
  if (isClient && organizations.length > 0 && !currentOrgId) {
    // Race condition: organizations might change while this runs
    const saved = localStorage.getItem("current-organization-id");
    const validOrgId =
      saved && organizations.some((org) => org.id === saved)
        ? saved
        : organizations[0].id;
    setCurrentOrgId(validOrgId);
  }
}, [isClient, organizations, currentOrgId]);
```

**Impact**:

- Multiple renders with inconsistent state
- localStorage overwrites during rapid state changes
- Organization selection fails intermittently

---

### 🚨 **RACE CONDITION #3: Token Refresh vs Organization Switch**

**Issue**: Automatic token refresh and organization switching can occur simultaneously, causing session conflicts.

**Location**: Multiple files

- `frontend/src/hooks/use-auth-queries.ts` (auto-refresh)
- `frontend/src/hooks/use-organization-mutations.ts` (org switch)

**Impact**:

- Session cookie overwritten with stale organization data
- User loses organization context mid-operation
- Redirect to welcome page during active work

---

### 🚨 **RACE CONDITION #4: Session Verification During API Calls**

**Issue**: `authenticatedApiClient` doesn't wait for session verification before making requests.

**Location**: `frontend/src/app/_actions/api-config.ts:45-85`

```typescript
// Request interceptor doesn't verify session state
axios.interceptors.request.use(async (config) => {
  const { session } = await verifySession();
  // No check if session is actually valid before proceeding
  if (session?.access_token) {
    config.headers.Authorization = `Bearer ${session.access_token}`;
  }
  return config;
});
```

**Impact**: API calls proceed with potentially invalid/expired sessions.

---

### 🚨 **RACE CONDITION #5: Organization Switch State Management**

**Issue**: Frontend and backend organization state can become inconsistent during switching.

**Location**: `frontend/src/app/_actions/organizations.ts:118-135`

```typescript
export async function switchOrganization(orgId: string): Promise<string> {
  // 1. Backend call
  await authenticatedApiClient({
    url: `/api/v1/organizations/${orgId}/switch`,
    method: "POST",
  });

  // 2. Frontend session update (race condition here)
  await updateAuthSession({
    organization_id: orgId,
  });

  return orgId;
}
```

**Impact**: Backend updates user's current org, but frontend session update might fail, causing state mismatch.

---

## Architectural Issues

### ❌ **ISSUE A: Insufficient Session Validation**

**Problem**: Session verification doesn't validate organization membership before allowing access.

**Location**: `frontend/src/lib/auth.ts:decrypt()` function doesn't check org membership.

**Impact**: Users can access resources from organizations they're no longer members of.

---

### ❌ **ISSUE B: Missing Request Deduplication**

**Problem**: Multiple simultaneous organization fetch requests can cause race conditions.

**Location**: `OrganizationProvider` doesn't prevent duplicate API calls.

**Impact**: Unnecessary API load and potential state conflicts.

---

### ❌ **ISSUE C: Inadequate Error Recovery**

**Problem**: When organization fetch fails, user gets stuck in redirect loop.

**Location**: `WorkspaceSelector` doesn't handle persistent errors gracefully.

**Impact**: Users cannot recover from temporary network issues without clearing cookies.

---

## Detailed Fix Implementation

### 🔧 **FIX 1: Increase Login Redirect Delay & Add Session Verification**

**File**: `frontend/src/hooks/use-auth-mutations.ts`

```typescript
onSuccess: async (data) => {
  if (data.success) {
    setIsRedirecting(true);

    // Wait longer for cookie to be set and verify session
    let sessionReady = false;
    let attempts = 0;
    const maxAttempts = 10;

    while (!sessionReady && attempts < maxAttempts) {
      await new Promise((resolve) => setTimeout(resolve, 200));

      try {
        const { isAuthenticated } = await verifySession();
        if (isAuthenticated) {
          sessionReady = true;
        }
      } catch (error) {
        console.log(`Session verification attempt ${attempts + 1} failed`);
      }

      attempts++;
    }

    if (sessionReady) {
      router.push("/welcome");
    } else {
      console.error("Session not ready after login");
      setIsRedirecting(false);
    }
  }
};
```

---

### 🔧 **FIX 2: Refactor Organization Context Initialization**

**File**: `frontend/src/contexts/organization-context.tsx`

```typescript
export function OrganizationProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [currentOrgId, setCurrentOrgId] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);

  // Single initialization effect to prevent race conditions
  useEffect(() => {
    if (typeof window !== "undefined" && !isInitialized) {
      setIsInitialized(true);
    }
  }, [isInitialized]);

  // Fetch organizations with request deduplication
  const {
    data: organizations = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ["organizations"],
    queryFn: async () => {
      // Verify session before fetching
      const { isAuthenticated } = await verifySession();
      if (!isAuthenticated) {
        throw new Error("No valid session found");
      }
      return fetchUserOrganizations();
    },
    enabled: isInitialized,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    retry: (failureCount, error: any) => {
      if (error?.message?.includes("No valid session found")) {
        return failureCount < 2;
      }
      return failureCount < 3;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });

  // Organization selection with atomic updates
  useEffect(() => {
    if (isInitialized && organizations.length > 0 && !currentOrgId) {
      const selectInitialOrg = async () => {
        const saved = localStorage.getItem("current-organization-id");
        const validOrgId =
          saved && organizations.some((org) => org.id === saved)
            ? saved
            : organizations[0].id;

        // Atomic update to prevent race conditions
        setCurrentOrgId(validOrgId);
        localStorage.setItem("current-organization-id", validOrgId);
      };

      selectInitialOrg();
    }
  }, [isInitialized, organizations, currentOrgId]);

  // Rest of component...
}
```

---

### 🔧 **FIX 3: Add Session Verification to API Client**

**File**: `frontend/src/app/_actions/api-config.ts`

```typescript
// Enhanced request interceptor with session verification
axios.interceptors.request.use(async (config) => {
  try {
    const { isAuthenticated, session } = await verifySession();

    if (!isAuthenticated || !session?.access_token) {
      throw new Error("No valid session found");
    }

    config.headers.Authorization = `Bearer ${session.access_token}`;

    // Add organization context if available
    if (session.organization_id) {
      config.headers["X-Organization-ID"] = session.organization_id;
    }

    return config;
  } catch (error) {
    console.error("Session verification failed:", error);
    throw error;
  }
});
```

---

### 🔧 **FIX 4: Improve Organization Switch with Atomic Updates**

**File**: `frontend/src/app/_actions/organizations.ts`

```typescript
export async function switchOrganization(orgId: string): Promise<string> {
  const url = `/api/v1/organizations/${orgId}/switch`;

  try {
    // 1. Verify current session
    const { isAuthenticated, session } = await verifySession();
    if (!isAuthenticated) {
      throw new Error("No valid session found");
    }

    // 2. Backend switch (this updates user's current_organization_id)
    await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    // 3. Update frontend session atomically
    await updateAuthSession({
      organization_id: orgId,
    });

    // 4. Verify the update was successful
    const { session: updatedSession } = await verifySession();
    if (updatedSession?.organization_id !== orgId) {
      throw new Error("Organization switch verification failed");
    }

    return orgId;
  } catch (error: any) {
    console.error("Failed to switch organization:", error);
    throw error;
  }
}
```

---

### 🔧 **FIX 5: Add Error Recovery to Workspace Selector**

**File**: `frontend/src/app/(private)/welcome/_components/workpace-selector.tsx`

```typescript
export function WorkspaceSelector({
  onCreateWorkspace,
  showLogo = true,
  showSignOut = true,
}: WorkspaceSelectorProps) {
  const { user } = useSession();
  const {
    userOrganizations,
    currentOrganization,
    isLoading,
    error,
    retryFetch,
  } = useOrganizationContext();
  const { selectOrganization, isPending: isNavigating } =
    useSelectOrganization();
  const { logout } = useLogout();
  const [selectedOrgId, setSelectedOrgId] = useState<string | null>(
    currentOrganization?.id ?? null
  );
  const [retryCount, setRetryCount] = useState(0);
  const [showSkeleton, setShowSkeleton] = useState(true);

  // Enhanced error recovery
  useEffect(() => {
    if (error && error.includes("No valid session found")) {
      if (retryCount < 3) {
        const delay = Math.min(1000 * Math.pow(2, retryCount), 5000);
        setTimeout(() => {
          setRetryCount((prev) => prev + 1);
          retryFetch();
        }, delay);
      } else {
        // After 3 retries, force logout to clear invalid session
        console.error("Persistent session error, forcing logout");
        logout();
      }
    }
  }, [error, retryCount, retryFetch, logout]);

  // Reset retry count on successful load
  useEffect(() => {
    if (!error && userOrganizations.length > 0) {
      setRetryCount(0);
    }
  }, [error, userOrganizations.length]);

  // Enhanced organization selection with validation
  const handleSelectOrganization = async (orgId: string) => {
    if (isNavigating) return;

    // Validate organization exists in current list
    const orgExists = userOrganizations.some((org) => org.id === orgId);
    if (!orgExists) {
      console.error("Selected organization not found in user organizations");
      return;
    }

    setSelectedOrgId(orgId);

    try {
      await selectOrganization(orgId);
    } catch (error) {
      console.error("Organization selection failed:", error);
      setSelectedOrgId(currentOrganization?.id ?? null);
    }
  };

  // Rest of component with enhanced error display...
}
```

---

### 🔧 **FIX 6: Add Backend Session Validation Enhancement**

**File**: `backend/middleware/middleware.go`

```go
// Enhanced auth middleware with organization validation
func EnhancedAuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authorization header required",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]
		claims, err := authService.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired token",
				"error":   err.Error(),
			})
		}

		// Enhanced: Validate organization membership if org context exists
		if claims.OrganizationID != nil {
			// Verify user is still a member of this organization
			var membership models.OrganizationMember
			if err := config.DB.Where(
				"organization_id = ? AND user_id = ? AND active = ?",
				*claims.OrganizationID, claims.UserID, true,
			).First(&membership).Error; err != nil {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"message": "User is no longer a member of this organization",
				})
			}
		}

		// Store user information in context
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userName", claims.Name)
		c.Locals("userRole", claims.Role)
		c.Locals("organizationID", claims.OrganizationID)
		c.Locals("sessionID", claims.SessionID)

		return c.Next()
	}
}
```

---

## Implementation Priority

### **Phase 1: Critical Fixes (Immediate)**

1. ✅ Fix login redirect delay and session verification
2. ✅ Add session verification to API client
3. ✅ Improve organization switch atomic updates

### **Phase 2: Stability Improvements (Week 1)**

4. ✅ Refactor organization context initialization
5. ✅ Add error recovery to workspace selector
6. ✅ Enhance backend session validation

### **Phase 3: Monitoring & Testing (Week 2)**

7. Add comprehensive logging for race condition detection
8. Implement automated tests for authentication flows
9. Add performance monitoring for session operations

---

## Testing Strategy

### **Race Condition Testing**

```typescript
// Test rapid login → organization fetch
describe("Authentication Race Conditions", () => {
  it("should handle rapid login to organization fetch", async () => {
    // Login
    await loginAction("user@test.com", "password");

    // Immediately try to fetch organizations (should not fail)
    const orgs = await fetchUserOrganizations();
    expect(orgs).toBeDefined();
  });

  it("should handle concurrent organization switches", async () => {
    // Simulate multiple rapid org switches
    const promises = [
      switchOrganization("org1"),
      switchOrganization("org2"),
      switchOrganization("org3"),
    ];

    // Only the last one should succeed
    const results = await Promise.allSettled(promises);
    const successful = results.filter((r) => r.status === "fulfilled");
    expect(successful).toHaveLength(1);
  });
});
```

### **Session Validation Testing**

```typescript
describe("Session Validation", () => {
  it("should reject requests with invalid sessions", async () => {
    // Corrupt session cookie
    document.cookie = "AUTH_SESSION=invalid";

    // Should fail gracefully
    await expect(fetchUserOrganizations()).rejects.toThrow(
      "No valid session found"
    );
  });
});
```

---

## Monitoring & Alerting

### **Key Metrics to Track**

1. **Session Verification Failures**: Rate of "No valid session found" errors
2. **Organization Fetch Failures**: Failed organization API calls after login
3. **Redirect Loop Incidents**: Users hitting `/welcome` multiple times
4. **Token Refresh Conflicts**: Concurrent refresh attempts
5. **Organization Switch Failures**: Failed organization switches

### **Recommended Alerts**

- Session verification failure rate > 5%
- Organization fetch failure rate > 2%
- Average login-to-organization-access time > 3 seconds
- Redirect loop detection (>3 welcome page visits in 30 seconds)

---

## Conclusion

The identified race conditions are causing significant user experience issues. The proposed fixes address the root causes through:

1. **Proper session verification timing**
2. **Atomic state updates**
3. **Request deduplication**
4. **Enhanced error recovery**
5. **Backend validation improvements**

Implementing these fixes in the suggested phases will eliminate the redirect loops and session verification failures while maintaining system performance and reliability.
