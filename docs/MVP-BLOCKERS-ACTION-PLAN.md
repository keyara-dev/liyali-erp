# MVP Blockers - Action Plan 2025-12-26

**Priority**: CRITICAL - Must fix before MVP testing
**Target Completion**: End of Week 1
**Estimated Effort**: 12-15 developer-days
**Status**: Ready for implementation

---

## BLOCKER #1: Password Reset Non-Functional

### Current State
- **Files Affected**:
  - `frontend/src/app/(auth)/forgot-password/page.tsx`
  - `frontend/src/app/(auth)/reset-password/page.tsx`
  - `frontend/src/app/_actions/auth.ts` (sendResetEmail, resetPassword stubs)

- **Problem**: Both functions return success without performing actions
  ```typescript
  // auth.ts lines 306-314
  export async function sendResetEmail(email: string): Promise<APIResponse> {
    try {
      // This is a stub implementation for password reset
      return successResponse({ email }, "Password reset email sent successfully");
    } catch (error: any) {
      return handleError(error, "POST", "/api/v1/auth/reset-password/send");
    }
  }
  ```

### Impact
- Users cannot recover lost passwords
- Critical security feature blocked
- Violates MVP requirement "fully functional features"

### Fix Plan

#### Step 1: Implement Backend Endpoints (Backend team)
```
POST /api/v1/auth/password-reset/send
  Body: { email: string }
  Response: { token_created: bool, email_sent: bool }

POST /api/v1/auth/password-reset/validate
  Body: { token: string }
  Response: { valid: bool, expires_at: timestamp }

POST /api/v1/auth/password-reset/confirm
  Body: { token: string, new_password: string }
  Response: { success: bool }
```

#### Step 2: Update Server Actions
**File**: `frontend/src/app/_actions/auth.ts`

```typescript
// Replace stub implementations
export async function sendResetEmail(email: string): Promise<APIResponse> {
  const url = `/api/v1/auth/password-reset/send`;

  try {
    const response = await axios.post(url, { email });
    return successResponse(
      response.data,
      "Password reset email sent successfully"
    );
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

export async function resetPassword(
  token: string,
  newPassword: string
): Promise<APIResponse> {
  const url = `/api/v1/auth/password-reset/confirm`;

  try {
    const response = await axios.post(url, {
      token,
      new_password: newPassword
    });
    return successResponse(response.data, "Password reset successfully");
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

export async function validateResetToken(token: string): Promise<APIResponse> {
  const url = `/api/v1/auth/password-reset/validate`;

  try {
    const response = await axios.post(url, { token });
    return successResponse(response.data, "Token valid");
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}
```

#### Step 3: Fix Frontend Pages

**File**: `frontend/src/app/(auth)/forgot-password/page.tsx`

- Remove unused username field
- Remove hardcoded 1-minute delay
- Add email validation
- Make it a server component with client form

**File**: `frontend/src/app/(auth)/reset-password/page.tsx`

- Fix password validation logic (condition is backwards)
- Validate token before showing form
- Remove unused token extraction
- Fix type safety in error handling

#### Step 4: Add Validation

```typescript
// Add to form validation
const validatePassword = (pwd: string): string[] => {
  const errors: string[] = [];
  if (pwd.length < 8) errors.push("At least 8 characters");
  if (!/[A-Z]/.test(pwd)) errors.push("At least 1 uppercase letter");
  if (!/[a-z]/.test(pwd)) errors.push("At least 1 lowercase letter");
  if (!/[0-9]/.test(pwd)) errors.push("At least 1 digit");
  return errors;
};

const validateEmail = (email: string): boolean => {
  const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return regex.test(email);
};
```

### Acceptance Criteria
- [ ] User can request password reset by email
- [ ] Email is actually sent (verify via backend logs)
- [ ] Reset link contains valid token
- [ ] Token is validated before password change
- [ ] New password is updated in database
- [ ] User can login with new password
- [ ] Invalid tokens are rejected
- [ ] Expired tokens are rejected

### Estimated Effort
- Backend: 3-4 days (2 developers)
- Frontend: 2-3 days (1 developer)
- Testing: 1 day
- **Total**: 6-8 days

---

## BLOCKER #2: Hardcoded Demo Credentials

### Current State
- **File**: `frontend/src/app/(auth)/login/page.tsx` (lines 31-92)
- **Problem**: 7 hardcoded email addresses and password displayed on login page

```typescript
<div className="border-t pt-6">
  <h3 className="text-sm font-semibold">Demo Accounts</h3>
  <div className="space-y-2 text-xs">
    <div><span>Requester: requester@liyali.com</span></div>
    <div><span>Manager: manager@liyali.com</span></div>
    {/* ... 5 more hardcoded emails ... */}
    <p>Password: password123</p>
  </div>
</div>
```

### Impact
- Violates MVP requirement "zero mock data"
- Creates confusion: are these real credentials?
- Security concern: exposes test passwords
- Makes app look unfinished

### Fix Plan

#### Option 1: Remove Completely (RECOMMENDED)
- Delete lines 29-92 entirely
- Cleaner, professional appearance
- Users focus on login task

#### Option 2: Environment-Gated (If needed for testing)
```typescript
// Only show in development
{process.env.NODE_ENV === 'development' && (
  <div className="border-t pt-6 bg-yellow-50 p-3 rounded">
    <h3 className="text-sm font-semibold">Demo Accounts (Dev Only)</h3>
    {/* demo content */}
  </div>
)}
```

#### Option 3: Backend-Configured
```typescript
// Fetch demo config from backend
const demoAccounts = process.env.NODE_ENV === 'development'
  ? await fetch('/api/v1/config/demo-accounts').then(r => r.json())
  : null;
```

### Recommendation
**Use Option 1** (Remove completely) for MVP. Demo accounts can be documented in:
- Setup guide (SETUP.md)
- Testing guide (TESTING.md)
- Admin panel (if needed)

### Implementation
```typescript
// frontend/src/app/(auth)/login/page.tsx
// DELETE lines 29-92 entirely

return (
  <div className="w-full max-w-md">
    <div className="bg-card rounded-lg p-8 space-y-6">
      <Logo isFull />
      <LoginForm />
      {/* STOP HERE - remove demo section */}
    </div>
  </div>
);
```

### Acceptance Criteria
- [ ] No hardcoded credentials visible in login UI
- [ ] No reference to "demo accounts" on page
- [ ] Demo accounts documented in setup guide
- [ ] Testing team has access to test credentials (separate from UI)

### Estimated Effort
- Implementation: 0.5 day
- Testing: 0.5 day
- **Total**: 1 day

---

## BLOCKER #3: Mock Data in Admin Pages

### Current State

#### 3A: Monitoring Page (WORST)
- **File**: `frontend/src/app/(private)/admin/monitoring/page.tsx` (lines 27-68)
- **Problem**: 100% fake metrics using `Math.random()`
- **Impact**: Dashboard shows completely fabricated data

```typescript
const generateMetricsData = () => {
  const data = []
  for (let i = 23; i >= 0; i--) {
    data.unshift({
      approvals: Math.floor(Math.random() * 20) + 5,  // RANDOM
      submissions: Math.floor(Math.random() * 30) + 10,  // RANDOM
    })
  }
  return data  // Different values on each page load!
}
```

#### 3B: User Details Page
- **File**: `frontend/src/app/(private)/admin/users/[id]/_components/user-details-client.tsx` (lines 52-100)
- **Problem**: 12 hardcoded metric values
  - Risk scores (riskMetrics)
  - Audit metrics
  - 4 fake recent activities

#### 3C: Reports Page
- **File**: `frontend/src/app/(private)/admin/reports/_components/admin-reports-client.tsx` (lines 31-39)
- **Problem**: CSV export contains hardcoded numbers
  ```typescript
  const csv = `Workflow Analytics Report
  Total Pending,24
  Total Approved,187
  Avg Approval Time,3.2 days`  // HARDCODED VALUES
  ```

### Impact
- Decision-makers get false information
- Audit compliance risk (fake metrics in reports)
- User distrust if they discover data is mock
- Violates "zero mock data" MVP requirement

### Fix Plan

#### Fix 3A: Monitoring Page

**Step 1: Create Backend Endpoint**
```
GET /api/v1/admin/metrics/system-health
  Response: {
    approvals_24h: number,
    submissions_24h: number,
    rejections_24h: number,
    error_rate: number,
    uptime_percent: number,
    response_time_ms: number
  }

GET /api/v1/admin/metrics/hourly?hours=24
  Response: [{ hour: timestamp, approvals: number, submissions: number, ... }]
```

**Step 2: Create React Query Hooks**
```typescript
// frontend/src/hooks/use-admin-metrics.ts
export function useSystemMetrics() {
  return useQuery({
    queryKey: ['admin', 'metrics', 'system'],
    queryFn: async () => {
      const response = await fetch('/api/v1/admin/metrics/system-health');
      return response.json();
    },
    staleTime: 30 * 1000, // 30 seconds for real-time feel
    refetchInterval: 30 * 1000,
  });
}

export function useHourlyMetrics(hours = 24) {
  return useQuery({
    queryKey: ['admin', 'metrics', 'hourly', hours],
    queryFn: async () => {
      const response = await fetch(`/api/v1/admin/metrics/hourly?hours=${hours}`);
      return response.json();
    },
    staleTime: 60 * 1000,
    refetchInterval: 60 * 1000,
  });
}
```

**Step 3: Update Monitoring Page**
```typescript
// frontend/src/app/(private)/admin/monitoring/page.tsx
export default function MonitoringPage() {
  const { data: systemMetrics, isLoading } = useSystemMetrics();
  const { data: hourlyData, isLoading: isLoadingChart } = useHourlyMetrics();

  if (isLoading) return <MonitoringSkeleton />;

  return (
    <div>
      <SystemHealthCard metrics={systemMetrics} />
      <MetricsChart data={hourlyData} isLoading={isLoadingChart} />
    </div>
  );
}
```

#### Fix 3B: User Details Page

**Step 1: Backend Endpoint**
```
GET /api/v1/admin/users/{id}/metrics
  Response: {
    total_documents: number,
    risk_score: number,
    recent_activities: Activity[],
    compliance_status: string,
    audit_metrics: object
  }
```

**Step 2: Create Hook**
```typescript
// frontend/src/hooks/use-user-metrics.ts
export function useUserMetrics(userId: string) {
  return useQuery({
    queryKey: ['admin', 'users', userId, 'metrics'],
    queryFn: async () => {
      const response = await fetch(`/api/v1/admin/users/${userId}/metrics`);
      return response.json();
    },
    enabled: !!userId,
  });
}
```

**Step 3: Update Component**
```typescript
// frontend/src/app/(private)/admin/users/[id]/_components/user-details-client.tsx
export function UserDetailsClient({ userId }) {
  const { data: metrics, isLoading, isError } = useUserMetrics(userId);

  if (isLoading) return <UserDetailsSkeleton />;
  if (isError) return <ErrorState />;

  return (
    <div>
      <RiskMetrics data={metrics.riskMetrics} />
      <RecentActivities data={metrics.recentActivities} />
    </div>
  );
}
```

#### Fix 3C: Reports Page

**Step 1: Backend Endpoint**
```
GET /api/v1/admin/reports/analytics?format=json
  Response: {
    total_pending: number,
    total_approved: number,
    total_rejected: number,
    avg_approval_time_hours: number,
    sla_compliance_percent: number
  }

GET /api/v1/admin/reports/download?format=csv
  Response: CSV file blob
```

**Step 2: Update Export Function**
```typescript
const handleExport = async () => {
  const response = await fetch('/api/v1/admin/reports/download?format=csv');
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `report-${new Date().toISOString()}.csv`;
  a.click();
};
```

### Acceptance Criteria
- [ ] Monitoring page displays real system metrics
- [ ] Metrics update automatically (refresh every 30 sec)
- [ ] User details show actual user metrics
- [ ] Reports export real data
- [ ] Loading states shown while fetching
- [ ] Error states handled gracefully
- [ ] No hardcoded values visible
- [ ] Metrics match backend source of truth

### Estimated Effort
- Backend: 4-5 days (2 developers)
- Frontend: 2-3 days (1 developer)
- Testing: 1 day
- **Total**: 7-9 days

---

## BLOCKER #4: Hardcoded Admin User Context

### Current State
- **Files Affected**:
  - `frontend/src/app/(private)/admin/users/page.tsx` (lines 79-80)
  - `frontend/src/app/(private)/admin/workflows/page.tsx` (line 13)
  - Workflow create/edit pages

- **Problem**: User ID and role hardcoded instead of from session

```typescript
// users/page.tsx
<UserManagementClient userId="system" userRole="ADMIN" />

// workflows/page.tsx
<WorkflowsClient userId="system" userRole="ADMIN" />
```

### Impact
- Wrong user context shown in admin pages
- Any user appears as "system" with "ADMIN" role
- Contradicts audit trail (can't track real user)
- Security issue: context bypass

### Fix Plan

#### Step 1: Get Authenticated User Context
```typescript
// All admin pages should follow this pattern:
export default async function AdminUsersPage() {
  const { session, isAuthenticated } = await verifySession();

  if (!isAuthenticated || !session?.user) {
    redirect("/login");
  }

  // Verify admin role
  if (session.user.role !== "ADMIN") {
    redirect("/unauthorized");
  }

  // Pass actual user context
  return (
    <UserManagementClient
      userId={session.user.id}
      userRole={session.user.role}
    />
  );
}
```

#### Step 2: Update All Admin Pages

**Files to update:**
1. `frontend/src/app/(private)/admin/users/page.tsx`
2. `frontend/src/app/admin/roles/page.tsx`
3. `frontend/src/app/(private)/admin/workflows/page.tsx`
4. `frontend/src/app/(private)/admin/workflows/create/page.tsx`
5. `frontend/src/app/(private)/admin/workflows/[id]/edit/page.tsx`

**Template:**
```typescript
import { verifySession } from '@/lib/auth';
import { redirect } from 'next/navigation';

export default async function AdminPage() {
  const { session, isAuthenticated } = await verifySession();

  if (!isAuthenticated) {
    redirect("/login");
  }

  const isAdmin = session?.user?.role === "ADMIN" ||
                  session?.user?.role === "SUPERADMIN";

  if (!isAdmin) {
    redirect("/unauthorized");
  }

  return (
    <AdminContent
      userId={session.user.id}
      userRole={session.user.role}
      userName={session.user.name}
    />
  );
}
```

### Acceptance Criteria
- [ ] All admin pages verify admin role at server level
- [ ] No hardcoded "system" user context
- [ ] Actual authenticated user passed to components
- [ ] Non-admin users redirected to /unauthorized
- [ ] Audit logs show real user ID for all actions

### Estimated Effort
- Implementation: 2-3 days (1 developer)
- Testing: 1 day
- **Total**: 3-4 days

---

## BLOCKER #5: Missing Admin Permission Verification

### Current State
- **Files Affected**: All 10 admin pages
- **Problem**: No server-level role verification before rendering

### Fix Plan

Implement role guard for all admin pages:

```typescript
// Create shared utility
// frontend/src/lib/admin-guard.ts
export async function requireAdminRole() {
  const { session, isAuthenticated } = await verifySession();

  if (!isAuthenticated) {
    redirect("/login");
  }

  const isAdmin = ["ADMIN", "SUPERADMIN", "COMPLIANCE_OFFICER"].includes(
    session.user?.role
  );

  if (!isAdmin) {
    redirect("/unauthorized");
  }

  return session;
}
```

Use in all admin pages:
```typescript
import { requireAdminRole } from '@/lib/admin-guard';

export default async function AdminPage() {
  const session = await requireAdminRole();

  return (
    <AdminContent
      userId={session.user.id}
      userRole={session.user.role}
    />
  );
}
```

### Acceptance Criteria
- [ ] Non-admin users cannot access /admin/* routes
- [ ] Proper redirect to /unauthorized shown
- [ ] Role is verified at server level (before client code runs)
- [ ] Audit log captures unauthorized attempts

### Estimated Effort
- Implementation: 1-2 days
- **Total**: 1-2 days

---

## BLOCKER #6: Purchase Orders Using Mock Data

### Current State
- **File**: `frontend/src/app/(private)/(main)/purchase-orders/[id]/_components/po-detail-client.tsx` (lines 52-65)
- **Problem**: Generates fake PO data instead of fetching from backend

```typescript
function generateMockPO(poId: string): PurchaseOrder {
  const poNumber = `PO-2024-${String(Math.floor(Math.random() * 9000) + 1000).padStart(4, "0")}`;
  return {
    id: poId,
    poNumber,
    vendor: { name: "Global Supplies Inc.", email: "...", phone: "..." },
    // ... more hardcoded data
  };
}
```

### Impact
- Users see fake purchase orders
- Amounts and vendor info incorrect
- No connection to actual procurement
- Violates "zero mock data" requirement

### Fix Plan

#### Step 1: Create Hook
```typescript
// frontend/src/hooks/use-purchase-order-queries.ts
export const PURCHASE_ORDER_KEYS = {
  all: ['purchase-orders'] as const,
  detail: (id: string) => [...PURCHASE_ORDER_KEYS.all, 'detail', id] as const,
};

export function usePurchaseOrderDetail(poId: string) {
  return useQuery({
    queryKey: PURCHASE_ORDER_KEYS.detail(poId),
    queryFn: async () => {
      const response = await getPurchaseOrderById(poId);
      if (!response.success) {
        throw new Error(response.message || 'Failed to fetch PO');
      }
      return response.data;
    },
    enabled: !!poId,
  });
}
```

#### Step 2: Update Component
```typescript
// po-detail-client.tsx
export function PODetailClient({ poId }) {
  const { data: po, isLoading, isError, error } = usePurchaseOrderDetail(poId);

  if (isLoading) return <PODetailSkeleton />;
  if (isError) return <ErrorUI message={error?.message} />;
  if (!po) return <NotFoundUI />;

  return <PODetailContent po={po} />;
}
```

#### Step 3: Ensure Server Action Exists
```typescript
// frontend/src/app/_actions/purchase-orders.ts
export async function getPurchaseOrderById(
  poId: string
): Promise<APIResponse<PurchaseOrder>> {
  const { session } = await verifySession();

  if (!session?.user) {
    return unauthorizedResponse();
  }

  try {
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    const response = await fetch(
      `${backendUrl}/api/v1/purchase-orders/${poId}`,
      {
        headers: {
          'Authorization': `Bearer ${session.user.token}`,
        },
      }
    );

    if (!response.ok) {
      if (response.status === 404) return notFoundResponse();
      return errorResponse('Failed to fetch PO', response.status);
    }

    const data = await response.json();
    return successResponse(data, 'PO loaded');
  } catch (error) {
    return handleError(error, 'GET', `/purchase-orders/${poId}`);
  }
}
```

### Acceptance Criteria
- [ ] PO detail fetches from backend API
- [ ] Real PO data displayed (not generated)
- [ ] Loading state shown while fetching
- [ ] Error state handled gracefully
- [ ] Missing POs show 404 error
- [ ] No hardcoded vendor data

### Estimated Effort
- Implementation: 2-3 days
- **Total**: 2-3 days

---

## Summary Table

| Blocker | Effort | Priority | Days |
|---------|--------|----------|------|
| #1: Password Reset | 6-8 | Critical | 8 |
| #2: Demo Credentials | 1 | Critical | 1 |
| #3: Mock Data (3 pages) | 7-9 | Critical | 9 |
| #4: Hardcoded Contexts | 3-4 | Critical | 4 |
| #5: Permission Verification | 1-2 | Critical | 2 |
| #6: PO Mock Data | 2-3 | High | 3 |
| **TOTAL** | **20-27** | - | **27 days** |

**Parallel Work Possible**: Yes
- Backend team: Blockers #1, #3 (8-10 days)
- Frontend team: Blockers #2, #4, #5, #6 (4-5 days)
- Overlap: Blockers #1, #3 (coordinate API contracts)

**Suggested Timeline**:
- **Week 1**: All blockers in parallel
- **Week 2**: Testing & fixes
- **Week 3**: Final validation & MVP ready

---

## Implementation Order

### Day 1-2
- [x] Blocker #2: Remove demo credentials (1 dev)
- [x] Blocker #5: Add permission guards (1 dev)

### Day 2-3
- [x] Blocker #4: Fix hardcoded user contexts (1 dev)
- [ ] Blocker #1: Backend endpoints (2 devs)
- [ ] Blocker #3: Backend endpoints (2 devs)

### Day 4-5
- [x] Blocker #6: PO detail backend integration (1 dev)
- [ ] Blocker #1: Frontend form fixes (1 dev)
- [ ] Blocker #3: Frontend hooks & components (1 dev)

### Day 6-7
- [ ] Integration testing across all blockers
- [ ] End-to-end testing
- [ ] Performance verification

### Day 8+
- [ ] Bug fixes
- [ ] Final validation
- [ ] Ready for MVP testing

---

**Status**: Ready for Implementation
**Approved by**: Comprehensive Audit
**Date**: 2025-12-26
