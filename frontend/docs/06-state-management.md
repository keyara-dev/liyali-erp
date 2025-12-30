# State Management

The frontend uses a multi-layered state management approach combining React Query for server state, Zustand for client state, and localStorage for offline persistence.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    State Management Layers                  │
├─────────────────────────────────────────────────────────────┤
│ 1. Server State (TanStack Query)                           │
│    - API data caching                                       │
│    - Background refetching                                  │
│    - Optimistic updates                                     │
│                                                             │
│ 2. Client State (Zustand)                                  │
│    - UI state                                               │
│    - Form state                                             │
│    - Workflow stores                                        │
│                                                             │
│ 3. Persistent State (localStorage)                         │
│    - Offline data                                           │
│    - User preferences                                       │
│    - Draft documents                                        │
│                                                             │
│ 4. Context State (React Context)                           │
│    - Organization context                                   │
│    - Theme context                                          │
│    - Authentication state                                   │
└─────────────────────────────────────────────────────────────┘
```

## TanStack Query (Server State)

### Configuration

```typescript
// src/app/providers.tsx
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,        // 5 minutes - data considered fresh
      gcTime: 10 * 60 * 1000,          // 10 minutes - kept in memory
      retry: 3,                         // Retry failed queries 3 times
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      refetchOnWindowFocus: false,      // Don't auto-refetch on window focus
      refetchOnReconnect: true,         // Refetch when network reconnects
      refetchOnMount: true,             // Refetch on component mount if stale
    },
    mutations: {
      retry: 1,                         // Retry mutations once
      onError: (error) => {
        console.error('Mutation error:', error);
      },
    },
  },
});
```

### Query Keys Structure

```typescript
// src/lib/constants.ts
export const QUERY_KEYS = {
  REQUISITIONS: {
    ALL: 'requisitions',
    BY_ID: 'requisition',
    STATS: 'requisition-stats',
  },
  PURCHASE_ORDERS: {
    ALL: 'purchase-orders',
    BY_ID: 'purchase-order',
    STATS: 'purchase-order-stats',
  },
  PAYMENT_VOUCHERS: {
    ALL: 'payment-vouchers',
    BY_ID: 'payment-voucher',
    STATS: 'payment-voucher-stats',
  },
  DASHBOARD: {
    METRICS: 'dashboard-metrics',
    ACTIVITIES: 'dashboard-activities',
  },
  APPROVALS_PENDING: 'approvals-pending',
} as const;
```

### Custom Query Hooks Pattern

All API interactions use custom hooks that encapsulate query logic:

```typescript
// src/hooks/use-requisition-queries.ts
export const useRequisitions = (
  page: number = 1,
  limit: number = 10,
  filters?: { status?: string; department?: string }
) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.ALL, page, limit, filters],
    queryFn: async () => {
      const response = await getRequisitions(page, limit, filters);
      return response.success ? response.data : [];
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

export const useRequisitionById = (
  requisitionId: string,
  initialData?: Requisition
) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId],
    queryFn: async () => {
      const response = await getRequisitionById(requisitionId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    initialData,
    staleTime: 5 * 60 * 1000,
    enabled: !!requisitionId,
  });
```

### Mutation Patterns

Mutations handle optimistic updates and cache invalidation:

```typescript
export const useSaveRequisition = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateRequisitionRequest | UpdateRequisitionRequest) => {
      const response = "requisitionId" in data && data.requisitionId
        ? await updateRequisition(data as UpdateRequisitionRequest)
        : await createRequisition(data as CreateRequisitionRequest);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      toast.success("Requisition saved successfully");

      // Invalidate related queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
      });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to save requisition");
    },
  });
};
```

## Zustand Stores (Client State)

### Workflow Stores

In-memory stores for workflow document management:

```typescript
// src/lib/workflow-stores.ts
export const documentStore = new Map<string, WorkflowDocument>();
export const approversStore = new Map<string, Approver[]>();
export const approvalLogsStore = new Map<string, ApprovalLogEntry[]>();
export const attachmentsStore = new Map<string, Attachment[]>();

export let isInitialized = false;

export function initializeWorkflowStores() {
  if (isInitialized) return;
  // Initialize with empty stores
  isInitialized = true;
}
```

### Store Patterns

Zustand stores follow a consistent pattern:

```typescript
// Example store structure
interface StoreState {
  // State
  data: any[];
  loading: boolean;
  error: string | null;
  
  // Actions
  setData: (data: any[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const useStore = create<StoreState>((set) => ({
  // Initial state
  data: [],
  loading: false,
  error: null,
  
  // Actions
  setData: (data) => set({ data }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  reset: () => set({ data: [], loading: false, error: null }),
}));
```

## localStorage (Persistent State)

### Storage Management

Centralized storage operations through a unified API:

```typescript
// src/lib/storage/storage.ts
export const STORAGE_KEYS = {
  PURCHASE_ORDERS: 'liyali_purchase_orders',
  REQUISITIONS: 'liyali_requisitions',
  PAYMENT_VOUCHERS: 'liyali_payment_vouchers',
  GOODS_RECEIVED_NOTES: 'liyali_goods_received_notes',
} as const;

// Generic document operations
export function getDocuments<T>(storageKey: string): T[] {
  if (typeof window === 'undefined') return [];
  
  try {
    const stored = localStorage.getItem(storageKey);
    if (!stored) return [];
    const parsed = JSON.parse(stored);
    return Array.isArray(parsed) ? parsed : [];
  } catch (error) {
    console.error(`Failed to load documents from ${storageKey}:`, error);
    return [];
  }
}

export function saveDocument<T extends { id: string }>(
  storageKey: string,
  document: T
): T {
  try {
    if (typeof window === 'undefined') return document;

    const documents = getDocuments<T>(storageKey);
    const index = documents.findIndex((doc) => doc.id === document.id);

    if (index >= 0) {
      documents[index] = document;
    } else {
      documents.push(document);
    }

    localStorage.setItem(storageKey, JSON.stringify(documents));
    return document;
  } catch (error) {
    console.error(`Failed to save document to ${storageKey}:`, error);
    throw error;
  }
}
```

### Storage Initialization

Storage is initialized on app startup:

```typescript
// src/hooks/use-initialize-storage.ts
export function useInitializeStorage(): void {
  useEffect(() => {
    // Initialize in-memory workflow stores
    initializeWorkflowStores();

    // Initialize localStorage with seed data
    initializeStorage();
  }, []);
}
```

## React Context (Global State)

### Organization Context

Multi-tenancy support through organization context:

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

## State Flow Patterns

### Data Flow Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Action   │───▶│  Custom Hook    │───▶│   Server API    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   UI Update     │◀───│  Query Cache    │◀───│   API Response  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  localStorage   │
                       │   (Offline)     │
                       └─────────────────┘
```

### Optimistic Updates

For better UX, mutations use optimistic updates:

```typescript
const mutation = useMutation({
  mutationFn: updateDocument,
  onMutate: async (newData) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries({ queryKey: ['documents'] });

    // Snapshot previous value
    const previousData = queryClient.getQueryData(['documents']);

    // Optimistically update
    queryClient.setQueryData(['documents'], (old: any[]) => 
      old.map(item => item.id === newData.id ? { ...item, ...newData } : item)
    );

    return { previousData };
  },
  onError: (err, newData, context) => {
    // Rollback on error
    queryClient.setQueryData(['documents'], context?.previousData);
  },
  onSettled: () => {
    // Always refetch after error or success
    queryClient.invalidateQueries({ queryKey: ['documents'] });
  },
});
```

### Cache Invalidation Strategy

Strategic cache invalidation ensures data consistency:

```typescript
// After creating a requisition
queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.REQUISITIONS.ALL] });
queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.REQUISITIONS.STATS] });
queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.DASHBOARD.METRICS] });

// After approving a requisition (creates PO)
queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL] });
queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS] });
```

## Offline-First Patterns

### Background Sync

The app queues operations when offline and syncs when online:

```typescript
// Offline queue pattern
const offlineQueue = [];

const executeWithOfflineSupport = async (operation) => {
  if (navigator.onLine) {
    try {
      return await operation();
    } catch (error) {
      // Queue for later if network error
      offlineQueue.push(operation);
      throw error;
    }
  } else {
    // Queue immediately if offline
    offlineQueue.push(operation);
    return { success: false, queued: true };
  }
};

// Process queue when online
window.addEventListener('online', () => {
  processOfflineQueue();
});
```

### Data Synchronization

Data flows between storage layers:

1. **Online**: API → Query Cache → localStorage
2. **Offline**: localStorage → Query Cache → UI
3. **Sync**: localStorage → API → Query Cache

## Best Practices

### Query Organization

- Use consistent query key patterns
- Group related queries by feature
- Include all variables that affect the query in the key
- Use query key factories for complex keys

### Mutation Handling

- Always handle loading and error states
- Provide user feedback with toast notifications
- Implement optimistic updates for better UX
- Invalidate related queries after mutations

### Storage Management

- Use TypeScript for type safety
- Handle storage errors gracefully
- Implement data migration strategies
- Clear stale data periodically

### Context Usage

- Keep contexts focused and minimal
- Avoid prop drilling with strategic context placement
- Use multiple contexts instead of one large context
- Provide default values and error boundaries

## Migration Strategy

When transitioning to real backend APIs:

1. **Replace Server Actions**: Update API client to use HTTP calls
2. **Update Query Functions**: Replace mock data with real API calls
3. **Remove Storage Layer**: Delete localStorage operations
4. **Update Error Handling**: Handle network errors appropriately
5. **Test Offline Behavior**: Ensure graceful degradation

The current architecture is designed to make this transition seamless while providing a fully functional offline-first experience during development.