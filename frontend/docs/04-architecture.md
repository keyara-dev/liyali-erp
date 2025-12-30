# System Architecture

Comprehensive overview of the Liyali Gateway Frontend architecture, design patterns, and system components.

## Architecture Overview

The Liyali Gateway Frontend follows **modern React patterns** with Next.js 15 App Router, emphasizing performance, maintainability, and user experience.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Browser Layer                            │
├─────────────────────────────────────────────────────────────────┤
│  Next.js App Router - File-based routing & server components   │
├─────────────────────────────────────────────────────────────────┤
│  React 19 - Concurrent features & server components            │
├─────────────────────────────────────────────────────────────────┤
│  Component Layer - shadcn/ui + custom components               │
├─────────────────────────────────────────────────────────────────┤
│  State Management - Zustand + TanStack Query + Context         │
├─────────────────────────────────────────────────────────────────┤
│  Storage Layer - IndexedDB with offline-first strategy         │
├─────────────────────────────────────────────────────────────────┤
│  Network Layer - Server Actions + HTTP client                  │
└─────────────────────────────────────────────────────────────────┘
```

## Core Architectural Principles

### 1. Offline-First Architecture
- **Local Storage**: IndexedDB as primary data store
- **Background Sync**: Automatic synchronization when online
- **Optimistic Updates**: Immediate UI feedback
- **Conflict Resolution**: Smart merge strategies

### 2. Server-First Rendering
- **Server Components**: Default for static content
- **Client Components**: Only when interactivity is needed
- **Streaming**: Progressive page loading
- **Hydration**: Minimal client-side JavaScript

### 3. Type-Safe Development
- **TypeScript**: Full type coverage
- **Schema Validation**: Zod for runtime validation
- **API Types**: Shared types with backend
- **Component Props**: Strict prop typing

### 4. Performance-Optimized
- **Code Splitting**: Route-based and component-based
- **Lazy Loading**: Dynamic imports for heavy components
- **Caching**: Multi-layer caching strategy
- **Bundle Optimization**: Tree shaking and minification

## Directory Structure

### App Router Structure
```
src/app/
├── (auth)/                 # Authentication group
│   ├── login/
│   ├── register/
│   ├── forgot-password/
│   └── layout.tsx         # Auth layout
├── (private)/             # Protected routes group
│   ├── (main)/           # Main dashboard routes
│   │   ├── dashboard/
│   │   ├── requisitions/
│   │   ├── purchase-orders/
│   │   └── payment-vouchers/
│   ├── admin/            # Admin-only routes
│   ├── settings/         # User settings
│   └── layout.tsx        # Private layout
├── api/                  # API routes
├── _actions/             # Server actions
├── globals.css           # Global styles
├── layout.tsx            # Root layout
├── page.tsx             # Home page
└── providers.tsx        # App providers
```

### Component Architecture
```
src/components/
├── ui/                   # Base UI components (shadcn/ui)
│   ├── button.tsx
│   ├── input.tsx
│   ├── dialog.tsx
│   └── ...
├── layout/              # Layout components
│   ├── header/
│   ├── sidebar/
│   └── dashboard-layout.tsx
├── auth/                # Authentication components
│   └── permission-guard.tsx
├── workflows/           # Workflow components
│   ├── approval-flow-display.tsx
│   ├── bulk-operations-toolbar.tsx
│   └── workflow-selector.tsx
├── base/                # Base utility components
│   ├── empty-state.tsx
│   ├── error-display.tsx
│   └── page-header.tsx
└── [feature]/           # Feature-specific components
```

### State Management Structure
```
src/lib/
├── stores/              # Zustand stores
│   ├── auth-store.ts
│   ├── approval-store.ts
│   └── workflow-stores.ts
├── storage/             # IndexedDB management
│   ├── index.ts
│   ├── storage.ts
│   └── seed-data.ts
├── api/                 # API client
│   └── client.ts
└── utils/               # Utilities
    └── index.ts
```

## Component Patterns

### 1. Server Components (Default)
```tsx
// Server component for static content
export default async function RequisitionsPage() {
  // Server-side data fetching
  const requisitions = await getRequisitions();
  
  return (
    <div>
      <h1>Requisitions</h1>
      <RequisitionsList data={requisitions} />
    </div>
  );
}
```

### 2. Client Components (Interactive)
```tsx
'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';

export function InteractiveComponent() {
  const [count, setCount] = useState(0);
  
  return (
    <Button onClick={() => setCount(c => c + 1)}>
      Count: {count}
    </Button>
  );
}
```

### 3. Compound Components
```tsx
// Flexible component composition
export function DataTable({ children }) {
  return <div className="data-table">{children}</div>;
}

DataTable.Header = function Header({ children }) {
  return <div className="table-header">{children}</div>;
};

DataTable.Body = function Body({ children }) {
  return <div className="table-body">{children}</div>;
};

// Usage
<DataTable>
  <DataTable.Header>
    <h2>Requisitions</h2>
  </DataTable.Header>
  <DataTable.Body>
    {/* Table content */}
  </DataTable.Body>
</DataTable>
```

### 4. Custom Hooks Pattern
```tsx
// Custom hook for data fetching
export function useRequisitions() {
  return useQuery({
    queryKey: ['requisitions'],
    queryFn: async () => {
      const response = await apiClient.requisitions.getAll();
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Usage in component
function RequisitionsList() {
  const { data, isLoading, error } = useRequisitions();
  
  if (isLoading) return <Skeleton />;
  if (error) return <ErrorDisplay error={error} />;
  
  return (
    <div>
      {data?.map(req => (
        <RequisitionCard key={req.id} requisition={req} />
      ))}
    </div>
  );
}
```

## State Management Architecture

### 1. Server State (TanStack Query)
```tsx
// Query configuration
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,        // 5 minutes fresh
      gcTime: 10 * 60 * 1000,          // 10 minutes in cache
      retry: 3,                         // Retry failed queries
      refetchOnWindowFocus: false,      // Don't refetch on focus
      refetchOnReconnect: true,         // Refetch on reconnect
    },
  },
});

// Query hook
export function useRequisitionQuery(id: string) {
  return useQuery({
    queryKey: ['requisition', id],
    queryFn: () => apiClient.requisitions.getById(id),
    enabled: !!id,
  });
}

// Mutation hook
export function useCreateRequisition() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: apiClient.requisitions.create,
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ['requisitions'] });
    },
    onError: (error) => {
      toast.error('Failed to create requisition');
    },
  });
}
```

### 2. Client State (Zustand)
```tsx
// Store definition
interface ApprovalStore {
  selectedItems: string[];
  bulkAction: 'approve' | 'reject' | null;
  setSelectedItems: (items: string[]) => void;
  setBulkAction: (action: 'approve' | 'reject' | null) => void;
  clearSelection: () => void;
}

export const useApprovalStore = create<ApprovalStore>((set) => ({
  selectedItems: [],
  bulkAction: null,
  setSelectedItems: (items) => set({ selectedItems: items }),
  setBulkAction: (action) => set({ bulkAction: action }),
  clearSelection: () => set({ selectedItems: [], bulkAction: null }),
}));

// Usage in component
function BulkOperationsToolbar() {
  const { selectedItems, bulkAction, setBulkAction } = useApprovalStore();
  
  return (
    <div className="flex gap-2">
      <Button 
        onClick={() => setBulkAction('approve')}
        disabled={selectedItems.length === 0}
      >
        Approve Selected ({selectedItems.length})
      </Button>
    </div>
  );
}
```

### 3. Form State (React Hook Form)
```tsx
// Form with validation
const formSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  description: z.string().optional(),
  totalAmount: z.number().min(0.01, 'Amount must be positive'),
});

type FormData = z.infer<typeof formSchema>;

export function RequisitionForm() {
  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      title: '',
      description: '',
      totalAmount: 0,
    },
  });
  
  const createMutation = useCreateRequisition();
  
  const onSubmit = (data: FormData) => {
    createMutation.mutate(data);
  };
  
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Title</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        {/* More fields */}
        <Button type="submit" disabled={createMutation.isPending}>
          {createMutation.isPending ? 'Creating...' : 'Create Requisition'}
        </Button>
      </form>
    </Form>
  );
}
```

## Offline-First Architecture

### 1. Storage Layer
```tsx
// IndexedDB wrapper
class DocumentStore {
  private db: IDBDatabase;
  
  async init() {
    this.db = await openDB('liyali-gateway', 1, {
      upgrade(db) {
        // Create object stores
        db.createObjectStore('requisitions', { keyPath: 'id' });
        db.createObjectStore('purchase_orders', { keyPath: 'id' });
        db.createObjectStore('offline_queue', { keyPath: 'id' });
      },
    });
  }
  
  async getRequisitions(): Promise<Requisition[]> {
    const tx = this.db.transaction('requisitions', 'readonly');
    return await tx.objectStore('requisitions').getAll();
  }
  
  async saveRequisition(requisition: Requisition): Promise<void> {
    const tx = this.db.transaction('requisitions', 'readwrite');
    await tx.objectStore('requisitions').put(requisition);
  }
}
```

### 2. Offline Queue
```tsx
// Queue offline actions
interface QueuedAction {
  id: string;
  type: 'create' | 'update' | 'delete';
  entity: 'requisition' | 'purchase_order';
  data: any;
  timestamp: number;
}

class OfflineQueue {
  async addAction(action: Omit<QueuedAction, 'id' | 'timestamp'>) {
    const queuedAction: QueuedAction = {
      ...action,
      id: crypto.randomUUID(),
      timestamp: Date.now(),
    };
    
    await documentStore.saveToQueue(queuedAction);
  }
  
  async processQueue() {
    const actions = await documentStore.getQueuedActions();
    
    for (const action of actions) {
      try {
        await this.executeAction(action);
        await documentStore.removeFromQueue(action.id);
      } catch (error) {
        console.error('Failed to process queued action:', error);
      }
    }
  }
  
  private async executeAction(action: QueuedAction) {
    switch (action.type) {
      case 'create':
        await apiClient[action.entity].create(action.data);
        break;
      case 'update':
        await apiClient[action.entity].update(action.data.id, action.data);
        break;
      case 'delete':
        await apiClient[action.entity].delete(action.data.id);
        break;
    }
  }
}
```

### 3. Sync Strategy
```tsx
// Background sync hook
export function useOfflineSync() {
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const offlineQueue = useRef(new OfflineQueue());
  
  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      // Process queued actions when back online
      offlineQueue.current.processQueue();
    };
    
    const handleOffline = () => {
      setIsOnline(false);
    };
    
    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);
    
    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);
  
  return { isOnline };
}
```

## Authentication Architecture

### 1. JWT-based Authentication
```tsx
// Auth configuration
export interface AuthSession {
  access_token: string;
  role: UserType;
  user_id?: string;
  organization_id?: string;
  expiresAt: Date;
}

// Server-side auth functions
export async function createAuthSession(session: AuthSession) {
  const token = await encrypt(session, "30m");
  
  cookies().set(AUTH_SESSION, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    expires: session.expiresAt,
    sameSite: "strict",
    path: "/",
  });
}

export async function verifySession() {
  const cookie = cookies().get(AUTH_SESSION)?.value;
  if (!cookie) return { isAuthenticated: false, session: null };
  
  const session = await decrypt(cookie);
  return { isAuthenticated: true, session };
}
```

### 2. Permission-based Guards
```tsx
// Permission guard component
interface PermissionGuardProps {
  permissions: string[];
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function PermissionGuard({ 
  permissions, 
  fallback, 
  children 
}: PermissionGuardProps) {
  const { hasPermissions } = usePermissions();
  
  if (!hasPermissions(permissions)) {
    return fallback || <AccessDenied />;
  }
  
  return <>{children}</>;
}

// Usage
<PermissionGuard permissions={['requisitions.create']}>
  <CreateRequisitionButton />
</PermissionGuard>
```

## Performance Optimizations

### 1. Code Splitting
```tsx
// Route-based splitting (automatic with App Router)
// Component-based splitting
const HeavyComponent = lazy(() => import('./heavy-component'));

function MyPage() {
  return (
    <Suspense fallback={<Skeleton />}>
      <HeavyComponent />
    </Suspense>
  );
}
```

### 2. Image Optimization
```tsx
// Next.js Image component
import Image from 'next/image';

<Image
  src="/images/logo.png"
  alt="Liyali Gateway"
  width={200}
  height={100}
  priority // For above-the-fold images
  placeholder="blur" // Show blur while loading
/>
```

### 3. Bundle Analysis
```bash
# Analyze bundle size
pnpm analyze

# Output shows:
# - Largest modules
# - Duplicate dependencies
# - Optimization opportunities
```

## Error Handling Architecture

### 1. Error Boundaries
```tsx
// Global error boundary
export class ErrorBoundary extends Component<
  { children: ReactNode; fallback?: ComponentType<{ error: Error }> },
  { hasError: boolean; error?: Error }
> {
  constructor(props: any) {
    super(props);
    this.state = { hasError: false };
  }
  
  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }
  
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
    // Send to error reporting service
  }
  
  render() {
    if (this.state.hasError) {
      const Fallback = this.props.fallback || DefaultErrorFallback;
      return <Fallback error={this.state.error!} />;
    }
    
    return this.props.children;
  }
}
```

### 2. API Error Handling
```tsx
// Centralized error handling
export function useApiError() {
  return useMutation({
    mutationFn: apiCall,
    onError: (error: ApiError) => {
      switch (error.status) {
        case 401:
          // Redirect to login
          router.push('/login');
          break;
        case 403:
          toast.error('You do not have permission to perform this action');
          break;
        case 500:
          toast.error('Server error. Please try again later.');
          break;
        default:
          toast.error(error.message || 'An unexpected error occurred');
      }
    },
  });
}
```

## Testing Architecture

### 1. Component Testing
```tsx
// Component test example
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RequisitionForm } from './requisition-form';

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  
  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  );
}

test('should create requisition when form is submitted', async () => {
  renderWithProviders(<RequisitionForm />);
  
  fireEvent.change(screen.getByLabelText(/title/i), {
    target: { value: 'Test Requisition' },
  });
  
  fireEvent.click(screen.getByRole('button', { name: /create/i }));
  
  expect(await screen.findByText(/requisition created/i)).toBeInTheDocument();
});
```

### 2. Hook Testing
```tsx
// Hook test example
import { renderHook, waitFor } from '@testing-library/react';
import { useRequisitions } from './use-requisitions';

test('should fetch requisitions', async () => {
  const { result } = renderHook(() => useRequisitions(), {
    wrapper: QueryClientProvider,
  });
  
  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });
  
  expect(result.current.data).toHaveLength(3);
});
```

## Security Considerations

### 1. XSS Prevention
- All user input is sanitized
- Content Security Policy (CSP) headers
- Trusted types for DOM manipulation

### 2. CSRF Protection
- SameSite cookies
- CSRF tokens for state-changing operations
- Origin validation

### 3. Data Validation
- Client-side validation with Zod
- Server-side validation
- Type-safe API contracts

## Deployment Architecture

### 1. Build Process
```bash
# Production build
pnpm build

# Outputs:
# - Static assets in .next/static/
# - Server-side code in .next/server/
# - Optimized bundles with code splitting
```

### 2. Environment Configuration
```env
# Production environment
NODE_ENV=production
NEXT_PUBLIC_API_URL=https://api.liyali.com
AUTH_SECRET=production-secret-key
NEXTAUTH_URL=https://app.liyali.com
```

### 3. Performance Monitoring
- Core Web Vitals tracking
- Bundle size monitoring
- Error tracking and reporting
- User analytics (optional)

This architecture provides a solid foundation for a modern, scalable, and maintainable React application with excellent user experience and developer productivity.