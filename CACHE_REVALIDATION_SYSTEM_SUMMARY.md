# Cache Revalidation System - Complete Implementation

## Overview

This document describes the comprehensive cache revalidation system implemented to ensure fresh data when switching between organizations/workspaces. The system handles both client-side (React Query) and server-side (Next.js) cache invalidation.

## 🎯 **Problem Solved**

When users switch organizations using the nav switcher, they were seeing stale data from the previous organization because:

1. **React Query cache** retained organization-specific data
2. **Next.js server cache** (fetch cache, route cache) wasn't invalidated
3. **localStorage** contained organization-specific cached data
4. **Server actions** were returning cached results for the previous organization

## 🏗️ **Architecture Overview**

```
Organization Switch Trigger
           ↓
    Cache Manager
           ↓
┌─────────────────────────────────────────┐
│  1. Client-Side Cache Invalidation     │
│     - React Query cache clear          │
│     - localStorage cleanup             │
│     - Force query refetch              │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  2. Server-Side Cache Revalidation     │
│     - Next.js path revalidation        │
│     - Cache tag invalidation           │
│     - Layout revalidation              │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  3. Fresh Data Fetch                   │
│     - New organization context         │
│     - Updated server actions           │
│     - Refreshed components             │
└─────────────────────────────────────────┘
```

## 📁 **Files Created/Modified**

### **New Files:**

1. **`frontend/src/lib/cache-manager.ts`**

   - Central cache management system
   - Handles client + server cache invalidation
   - Configurable cache clearing options

2. **`frontend/src/app/_actions/cache-revalidation.ts`**

   - Server actions for Next.js cache revalidation
   - Path and tag revalidation functions
   - Organization-scoped cache invalidation

3. **`frontend/src/hooks/use-cache-revalidation.ts`**

   - React hook for manual cache revalidation
   - Component-level cache management
   - Cache statistics and debugging

4. **`frontend/src/components/debug/cache-debug-panel.tsx`**
   - Development tool for cache debugging
   - Manual cache revalidation controls
   - Cache statistics visualization

### **Modified Files:**

1. **`frontend/src/stores/organization-store.ts`**
   - Updated `switchWorkspace` to use new cache manager
   - Integrated server-side cache revalidation

## 🔧 **Implementation Details**

### **1. Cache Manager (`cache-manager.ts`)**

```typescript
// Comprehensive cache invalidation
await clearOrganizationCache({
  clearLocalStorage: true,
  preserveKeys: ["theme", "user-preferences"],
  revalidateServerCache: true,
});
```

**Features:**

- **Smart localStorage cleanup**: Only removes organization-scoped keys
- **Preserved keys**: Keeps global app state (theme, user preferences)
- **React Query integration**: Clears all cached queries
- **Server cache integration**: Triggers Next.js revalidation

**Organization-Scoped Data Patterns:**

```typescript
const ORGANIZATION_SCOPED_QUERIES = [
  "requisitions",
  "purchase-orders",
  "grn",
  "budgets",
  "analytics",
  "dashboard",
  "workflows",
  "approvals",
  "categories",
  "vendors",
  "users",
  "members",
  "settings",
];
```

### **2. Server Cache Revalidation (`cache-revalidation.ts`)**

```typescript
// Revalidate all organization-scoped paths and tags
export async function revalidateOrganizationCache(organizationId?: string) {
  // Revalidate paths
  for (const path of ORGANIZATION_SCOPED_PATHS) {
    revalidatePath(path);
    revalidatePath(`${path}/[...slug]`); // Dynamic routes
  }

  // Revalidate cache tags
  for (const tag of ORGANIZATION_SCOPED_TAGS) {
    revalidateTag(tag);
    revalidateTag(`${tag}-${organizationId}`); // Org-specific tags
  }

  // Revalidate layouts
  revalidatePath("/(private)", "layout");
  revalidatePath("/(private)/(main)", "layout");
}
```

**Organization-Scoped Paths:**

```typescript
const ORGANIZATION_SCOPED_PATHS = [
  "/dashboard",
  "/requisitions",
  "/purchase-orders",
  "/grn",
  "/budgets",
  "/analytics",
  "/workflows",
  "/approvals",
  "/settings",
  "/reports",
  "/vendors",
];
```

**Cache Tags:**

```typescript
const ORGANIZATION_SCOPED_TAGS = [
  "requisitions",
  "purchase-orders",
  "grn",
  "budgets",
  "analytics",
  "dashboard",
  "workflows",
  "approvals",
  "organization-data",
];
```

### **3. Organization Store Integration**

```typescript
switchWorkspace: async (orgId: string) => {
  // 1. Switch organization on backend
  await switchOrganization(orgId);

  // 2. Update local state
  setCurrentOrganization(orgId);

  // 3. Clear all caches (client + server)
  await clearOrganizationCache({
    clearLocalStorage: true,
    revalidateServerCache: true,
  });

  // 4. Prefetch critical data
  await prefetchOrganizationData(orgId);
};
```

### **4. Manual Cache Revalidation Hook**

```typescript
const {
  revalidateOrganizationData,
  revalidateQueries,
  revalidatePaths,
  revalidateTags,
  revalidateCurrentPage,
  forceHardRefresh,
} = useCacheRevalidation();

// Revalidate all organization data
await revalidateOrganizationData(organizationId);

// Revalidate specific queries
await revalidateQueries(["requisitions", "dashboard"]);

// Revalidate specific paths
await revalidatePaths(["/dashboard", "/requisitions"]);
```

## 🚀 **Usage Examples**

### **Automatic Organization Switch**

```typescript
// Triggered automatically when user switches organization
const handleSelectWorkspace = async (orgId: string) => {
  await switchWorkspace(orgId); // Handles all cache invalidation
};
```

### **Manual Cache Refresh**

```typescript
// In a component that needs fresh data
const { revalidateCurrentPage } = useCacheRevalidation();

const handleRefresh = async () => {
  await revalidateCurrentPage();
};
```

### **Debugging Cache Issues**

```typescript
// Development component for cache debugging
<CacheDebugPanel />
```

## 📊 **Cache Invalidation Scope**

### **Client-Side (React Query)**

- ✅ All cached queries cleared
- ✅ Active queries refetched
- ✅ Stale queries invalidated
- ✅ Error queries reset

### **Server-Side (Next.js)**

- ✅ Route cache revalidated
- ✅ Data cache (fetch) revalidated
- ✅ Layout cache revalidated
- ✅ Dynamic routes revalidated

### **Local Storage**

- ✅ Organization-specific data cleared
- ✅ Global preferences preserved
- ✅ Authentication data preserved
- ✅ Theme settings preserved

## 🔍 **Cache Patterns**

### **Organization-Scoped Keys to Clear:**

```typescript
// localStorage patterns that get cleared
const patterns = [
  "requisitions-*",
  "purchase-orders-*",
  "grn-*",
  "budget-*",
  "analytics-*",
  "workflow-*",
  "approval-*",
  "dashboard-*",
  "filters-*",
  "preferences-*",
  "table-state-*",
  "form-data-*",
];
```

### **Preserved Keys:**

```typescript
// Keys that are never cleared
const preservedKeys = [
  "current-organization-id",
  "theme",
  "sidebar-state",
  "user-preferences",
  "auth-session",
  "language",
];
```

## 🛠️ **Development Tools**

### **Cache Debug Panel**

- **Cache Statistics**: View active, stale, error queries
- **Manual Actions**: Trigger specific cache invalidations
- **Organization Info**: Current organization context
- **Development Only**: Automatically hidden in production

### **Console Logging**

```typescript
// Detailed logging for debugging
console.log("[CacheManager] Clearing organization-scoped cache...");
console.log("[Server Cache] Revalidating cache for organization: org123");
console.log("[CacheManager] Organization cache cleared successfully");
```

## 📈 **Performance Considerations**

### **Optimizations:**

1. **Selective Invalidation**: Only clear organization-scoped data
2. **Preserved Keys**: Keep global app state intact
3. **Batch Operations**: Group cache operations together
4. **Background Prefetch**: Load critical data after switch
5. **Error Handling**: Graceful fallbacks for cache failures

### **Trade-offs:**

- **Thoroughness vs Speed**: Complete cache clear ensures fresh data but takes longer
- **Network Requests**: More requests after switch but guaranteed fresh data
- **User Experience**: Brief loading state but no stale data confusion

## 🔧 **Configuration Options**

### **Cache Manager Options:**

```typescript
interface CacheManagerOptions {
  queryClient?: QueryClient; // Custom query client
  clearLocalStorage?: boolean; // Clear localStorage (default: true)
  preserveKeys?: string[]; // Additional keys to preserve
  revalidateServerCache?: boolean; // Revalidate server cache (default: true)
}
```

### **Customization:**

```typescript
// Custom cache clearing
await clearOrganizationCache({
  clearLocalStorage: false, // Skip localStorage
  preserveKeys: ["custom-key"], // Preserve additional keys
  revalidateServerCache: false, // Skip server revalidation
});
```

## 🚨 **Error Handling**

### **Graceful Degradation:**

```typescript
try {
  await clearOrganizationCache();
} catch (error) {
  console.error("Cache clearing failed:", error);
  // Fallback: Force page reload
  window.location.reload();
}
```

### **Retry Logic:**

- **Client Cache**: Automatic retry with exponential backoff
- **Server Cache**: Individual path/tag failure doesn't block others
- **Network Errors**: Graceful handling with user feedback

## 📋 **Testing Strategy**

### **Unit Tests:**

- Cache manager functions
- Server action responses
- Hook behavior

### **Integration Tests:**

- Organization switch flow
- Cache invalidation verification
- Data freshness validation

### **Manual Testing:**

1. Switch organizations multiple times
2. Verify no stale data appears
3. Check cache statistics
4. Test error scenarios

## 🎯 **Success Metrics**

### **Before Implementation:**

- ❌ Stale data visible after organization switch
- ❌ Server actions returning cached results
- ❌ Inconsistent UI state
- ❌ Manual page refresh required

### **After Implementation:**

- ✅ Fresh data immediately after switch
- ✅ Server actions return current organization data
- ✅ Consistent UI state across all components
- ✅ No manual refresh required
- ✅ Comprehensive cache debugging tools

## 🔮 **Future Enhancements**

1. **Smart Prefetching**: Predict likely organization switches
2. **Partial Invalidation**: More granular cache control
3. **Cache Warming**: Background data loading
4. **Performance Metrics**: Cache hit/miss tracking
5. **User Preferences**: Configurable cache behavior

## 📚 **Related Documentation**

- [Next.js Caching Documentation](https://nextjs.org/docs/app/building-your-application/caching)
- [React Query Cache Management](https://tanstack.com/query/latest/docs/react/guides/caching)
- [Organization Store Architecture](./frontend/src/stores/README.md)

---

This cache revalidation system ensures that users always see fresh, organization-specific data when switching workspaces, eliminating the confusion and data inconsistencies that were occurring before.
