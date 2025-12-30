# Troubleshooting Guide

Comprehensive troubleshooting guide for common issues, debugging techniques, and solutions for the Liyali Gateway Frontend.

## Common Issues and Solutions

### Development Environment Issues

#### Node.js Version Conflicts

**Problem**: Application fails to start due to Node.js version mismatch.

```bash
Error: The engine "node" is incompatible with this module.
```

**Solution**:
```bash
# Check current Node.js version
node --version

# Install correct version using nvm
nvm install 18.17.0
nvm use 18.17.0

# Verify version
node --version

# Clear npm cache and reinstall
rm -rf node_modules package-lock.json
npm install
```

#### Package Manager Issues

**Problem**: Dependency conflicts or installation failures.

```bash
npm ERR! peer dep missing: react@^18.0.0
```

**Solution**:
```bash
# Clear all caches
npm cache clean --force
rm -rf node_modules package-lock.json

# Use exact package manager
pnpm install --frozen-lockfile

# Or force resolution
pnpm install --force
```

#### Port Already in Use

**Problem**: Development server can't start because port is occupied.

```bash
Error: listen EADDRINUSE: address already in use :::3000
```

**Solution**:
```bash
# Find process using port 3000
lsof -ti:3000

# Kill the process
kill -9 $(lsof -ti:3000)

# Or use different port
npm run dev -- --port 3001
```

### Build and Compilation Issues

#### TypeScript Compilation Errors

**Problem**: TypeScript errors preventing build.

```typescript
// Common TypeScript errors and solutions

// Error: Property 'id' does not exist on type 'unknown'
// Solution: Add proper typing
interface ApiResponse<T> {
  data: T;
  success: boolean;
}

const response: ApiResponse<User> = await fetchUser();
const userId = response.data.id; // Now TypeScript knows the type

// Error: Argument of type 'string | undefined' is not assignable
// Solution: Add null checks or use optional chaining
const user = users.find(u => u.id === userId);
const userName = user?.name ?? 'Unknown'; // Safe access

// Error: Cannot find module or its corresponding type declarations
// Solution: Add type definitions
npm install --save-dev @types/package-name
```

#### Next.js Build Failures

**Problem**: Build fails with various Next.js specific errors.

```bash
# Error: Module not found
Error: Module not found: Can't resolve '@/components/ui/button'

# Solution: Check tsconfig.json paths
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

```bash
# Error: Image optimization error
Error: Invalid src prop on `next/image`

# Solution: Configure image domains in next.config.js
module.exports = {
  images: {
    domains: ['localhost', 'api.liyali.com'],
    unoptimized: process.env.NODE_ENV === 'development',
  },
};
```

#### Memory Issues During Build

**Problem**: Build process runs out of memory.

```bash
FATAL ERROR: Ineffective mark-compacts near heap limit
```

**Solution**:
```bash
# Increase Node.js memory limit
export NODE_OPTIONS="--max-old-space-size=4096"
npm run build

# Or add to package.json scripts
{
  "scripts": {
    "build": "NODE_OPTIONS='--max-old-space-size=4096' next build"
  }
}
```

### Runtime Issues

#### React Hydration Errors

**Problem**: Hydration mismatches between server and client.

```bash
Warning: Text content did not match. Server: "Loading..." Client: "Welcome, John"
```

**Solution**:
```typescript
// Use useEffect for client-only content
import { useEffect, useState } from 'react';

function UserGreeting() {
  const [mounted, setMounted] = useState(false);
  const [user, setUser] = useState(null);

  useEffect(() => {
    setMounted(true);
    // Fetch user data
    fetchUser().then(setUser);
  }, []);

  if (!mounted) {
    return <div>Loading...</div>; // Same as server
  }

  return <div>Welcome, {user?.name}</div>;
}

// Or use dynamic imports with ssr: false
import dynamic from 'next/dynamic';

const ClientOnlyComponent = dynamic(
  () => import('./client-only-component'),
  { ssr: false }
);
```

#### State Management Issues

**Problem**: State not updating or persisting correctly.

```typescript
// Problem: State updates not reflecting
const [count, setCount] = useState(0);

const handleClick = () => {
  setCount(count + 1);
  console.log(count); // Still shows old value
};

// Solution: Use functional updates or useEffect
const handleClick = () => {
  setCount(prev => {
    const newCount = prev + 1;
    console.log(newCount); // Shows correct value
    return newCount;
  });
};

// Or use useEffect to react to state changes
useEffect(() => {
  console.log('Count updated:', count);
}, [count]);
```

#### API Integration Issues

**Problem**: API calls failing or returning unexpected data.

```typescript
// Debug API issues with proper error handling
async function fetchData() {
  try {
    const response = await fetch('/api/data');
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('API Error:', error);
    
    // Log additional context
    console.error('Request details:', {
      url: '/api/data',
      method: 'GET',
      headers: response?.headers,
      status: response?.status,
    });
    
    throw error;
  }
}

// Use React Query for better error handling
function useDataQuery() {
  return useQuery({
    queryKey: ['data'],
    queryFn: fetchData,
    retry: (failureCount, error) => {
      // Don't retry on 4xx errors
      if (error.status >= 400 && error.status < 500) {
        return false;
      }
      return failureCount < 3;
    },
    onError: (error) => {
      console.error('Query failed:', error);
      toast.error('Failed to load data');
    },
  });
}
```

### Performance Issues

#### Slow Page Loading

**Problem**: Pages take too long to load.

**Debugging Steps**:
```typescript
// 1. Measure performance
import { PerformanceTracker } from '@/lib/performance';

function MyPage() {
  useEffect(() => {
    PerformanceTracker.startMeasurement('page-load');
    
    return () => {
      const duration = PerformanceTracker.endMeasurement('page-load');
      console.log(`Page loaded in ${duration}ms`);
    };
  }, []);

  // Component content
}

// 2. Identify heavy components
const HeavyComponent = React.memo(function HeavyComponent({ data }) {
  const processedData = useMemo(() => {
    console.time('data-processing');
    const result = expensiveOperation(data);
    console.timeEnd('data-processing');
    return result;
  }, [data]);

  return <div>{/* Render processed data */}</div>;
});

// 3. Use React DevTools Profiler
// Enable profiler in development
if (process.env.NODE_ENV === 'development') {
  const { Profiler } = require('react');
  
  function onRenderCallback(id, phase, actualDuration) {
    console.log('Component render:', { id, phase, actualDuration });
  }
  
  // Wrap components with Profiler
  <Profiler id="MyComponent" onRender={onRenderCallback}>
    <MyComponent />
  </Profiler>
}
```

#### Memory Leaks

**Problem**: Application memory usage keeps increasing.

**Solution**:
```typescript
// 1. Clean up event listeners
useEffect(() => {
  const handleResize = () => {
    // Handle resize
  };

  window.addEventListener('resize', handleResize);
  
  return () => {
    window.removeEventListener('resize', handleResize);
  };
}, []);

// 2. Cancel async operations
useEffect(() => {
  let cancelled = false;
  
  async function fetchData() {
    const data = await api.getData();
    if (!cancelled) {
      setData(data);
    }
  }
  
  fetchData();
  
  return () => {
    cancelled = true;
  };
}, []);

// 3. Clear timers and intervals
useEffect(() => {
  const interval = setInterval(() => {
    // Do something
  }, 1000);
  
  return () => {
    clearInterval(interval);
  };
}, []);

// 4. Abort fetch requests
useEffect(() => {
  const abortController = new AbortController();
  
  fetch('/api/data', {
    signal: abortController.signal,
  })
    .then(response => response.json())
    .then(data => setData(data))
    .catch(error => {
      if (error.name !== 'AbortError') {
        console.error('Fetch error:', error);
      }
    });
  
  return () => {
    abortController.abort();
  };
}, []);
```

### Authentication Issues

#### Session Management Problems

**Problem**: Users getting logged out unexpectedly or sessions not persisting.

**Debugging**:
```typescript
// 1. Check session expiration
async function checkSession() {
  try {
    const response = await fetch('/api/auth/session');
    const session = await response.json();
    
    console.log('Session status:', {
      isAuthenticated: session.isAuthenticated,
      expiresAt: session.expiresAt,
      timeUntilExpiry: session.expiresAt ? 
        new Date(session.expiresAt).getTime() - Date.now() : null,
    });
    
    return session;
  } catch (error) {
    console.error('Session check failed:', error);
    return null;
  }
}

// 2. Monitor session in development
if (process.env.NODE_ENV === 'development') {
  setInterval(checkSession, 30000); // Check every 30 seconds
}

// 3. Handle session expiration gracefully
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      console.log('Session expired, redirecting to login');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

#### Permission Errors

**Problem**: Users seeing content they shouldn't have access to.

**Solution**:
```typescript
// 1. Debug permission checks
function usePermissionDebug() {
  const { user, permissions } = useAuth();
  
  useEffect(() => {
    if (process.env.NODE_ENV === 'development') {
      console.log('User permissions:', {
        user: user?.email,
        role: user?.role,
        permissions: permissions,
      });
    }
  }, [user, permissions]);
}

// 2. Add permission guards with logging
function PermissionGuard({ 
  permissions, 
  children, 
  fallback = <AccessDenied /> 
}) {
  const { hasPermissions } = usePermissions();
  const hasAccess = hasPermissions(permissions);
  
  if (process.env.NODE_ENV === 'development') {
    console.log('Permission check:', {
      required: permissions,
      hasAccess,
      userPermissions: useAuth().permissions,
    });
  }
  
  return hasAccess ? children : fallback;
}
```

## Debugging Tools and Techniques

### Browser Developer Tools

#### React Developer Tools

```typescript
// Enable React DevTools in production for debugging
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  window.__REACT_DEVTOOLS_GLOBAL_HOOK__ = window.__REACT_DEVTOOLS_GLOBAL_HOOK__ || {};
}

// Add component names for better debugging
function MyComponent() {
  // Component logic
}

MyComponent.displayName = 'MyComponent';
```

#### Network Tab Debugging

```typescript
// Add request/response logging
const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
});

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    console.log('API Request:', {
      method: config.method?.toUpperCase(),
      url: config.url,
      data: config.data,
      headers: config.headers,
    });
    return config;
  },
  (error) => {
    console.error('Request Error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    console.log('API Response:', {
      status: response.status,
      url: response.config.url,
      data: response.data,
    });
    return response;
  },
  (error) => {
    console.error('Response Error:', {
      status: error.response?.status,
      url: error.config?.url,
      data: error.response?.data,
      message: error.message,
    });
    return Promise.reject(error);
  }
);
```

### Custom Debugging Utilities

#### Debug Component

```typescript
// src/components/debug/debug-panel.tsx
import { useState } from 'react';

interface DebugPanelProps {
  data: any;
  title?: string;
}

export function DebugPanel({ data, title = 'Debug Info' }: DebugPanelProps) {
  const [isOpen, setIsOpen] = useState(false);

  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  return (
    <div className="fixed bottom-4 left-4 z-50">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="bg-red-500 text-white px-3 py-1 rounded text-sm"
      >
        {title}
      </button>
      
      {isOpen && (
        <div className="mt-2 bg-black text-green-400 p-4 rounded max-w-md max-h-96 overflow-auto text-xs font-mono">
          <pre>{JSON.stringify(data, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}

// Usage
function MyComponent() {
  const { data, isLoading, error } = useQuery(['data'], fetchData);
  
  return (
    <div>
      {/* Component content */}
      <DebugPanel 
        data={{ data, isLoading, error }} 
        title="Query State" 
      />
    </div>
  );
}
```

#### Performance Monitor

```typescript
// src/components/debug/performance-monitor.tsx
import { useEffect, useState } from 'react';

export function PerformanceMonitor() {
  const [metrics, setMetrics] = useState<any>({});

  useEffect(() => {
    if (process.env.NODE_ENV !== 'development') return;

    const updateMetrics = () => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      const paint = performance.getEntriesByType('paint');
      
      setMetrics({
        domContentLoaded: navigation.domContentLoadedEventEnd - navigation.navigationStart,
        loadComplete: navigation.loadEventEnd - navigation.navigationStart,
        firstPaint: paint.find(p => p.name === 'first-paint')?.startTime,
        firstContentfulPaint: paint.find(p => p.name === 'first-contentful-paint')?.startTime,
        memory: (performance as any).memory ? {
          used: Math.round((performance as any).memory.usedJSHeapSize / 1024 / 1024),
          total: Math.round((performance as any).memory.totalJSHeapSize / 1024 / 1024),
        } : null,
      });
    };

    updateMetrics();
    const interval = setInterval(updateMetrics, 5000);

    return () => clearInterval(interval);
  }, []);

  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  return (
    <div className="fixed top-4 right-4 bg-black text-white p-2 rounded text-xs font-mono">
      <div>DOM: {metrics.domContentLoaded}ms</div>
      <div>Load: {metrics.loadComplete}ms</div>
      <div>FP: {metrics.firstPaint}ms</div>
      <div>FCP: {metrics.firstContentfulPaint}ms</div>
      {metrics.memory && (
        <div>Memory: {metrics.memory.used}/{metrics.memory.total}MB</div>
      )}
    </div>
  );
}
```

### Error Tracking and Logging

#### Error Boundary with Logging

```typescript
// src/components/error/error-boundary.tsx
import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: React.ComponentType<{ error: Error; resetError: () => void }>;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error Boundary caught an error:', error, errorInfo);
    
    // Log to external service
    if (process.env.NODE_ENV === 'production') {
      // Send to Sentry, LogRocket, etc.
      this.logError(error, errorInfo);
    }
  }

  private logError(error: Error, errorInfo: ErrorInfo) {
    const errorData = {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString(),
      url: window.location.href,
      userAgent: navigator.userAgent,
    };

    // Send to logging service
    fetch('/api/errors', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(errorData),
    }).catch(err => {
      console.error('Failed to log error:', err);
    });
  }

  resetError = () => {
    this.setState({ hasError: false, error: undefined });
  };

  render() {
    if (this.state.hasError) {
      const Fallback = this.props.fallback || DefaultErrorFallback;
      return <Fallback error={this.state.error!} resetError={this.resetError} />;
    }

    return this.props.children;
  }
}

function DefaultErrorFallback({ error, resetError }: { error: Error; resetError: () => void }) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white shadow-lg rounded-lg p-6">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <svg className="h-8 w-8 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-gray-800">
              Something went wrong
            </h3>
            <div className="mt-2 text-sm text-gray-500">
              <p>{error.message}</p>
            </div>
          </div>
        </div>
        
        <div className="mt-4">
          <button
            onClick={resetError}
            className="w-full bg-red-600 text-white py-2 px-4 rounded hover:bg-red-700"
          >
            Try again
          </button>
        </div>
        
        {process.env.NODE_ENV === 'development' && (
          <details className="mt-4">
            <summary className="cursor-pointer text-sm text-gray-600">
              Error details (development only)
            </summary>
            <pre className="mt-2 text-xs bg-gray-100 p-2 rounded overflow-auto">
              {error.stack}
            </pre>
          </details>
        )}
      </div>
    </div>
  );
}
```

## Testing and Debugging

### Unit Test Debugging

```typescript
// Debug failing tests
describe('MyComponent', () => {
  it('should render correctly', () => {
    const { container, debug } = render(<MyComponent />);
    
    // Print DOM structure
    debug();
    
    // Or print specific element
    debug(container.firstChild);
    
    // Check what's actually rendered
    screen.logTestingPlaygroundURL();
    
    expect(screen.getByText('Expected text')).toBeInTheDocument();
  });
});

// Mock debugging
jest.mock('../api/client', () => ({
  fetchData: jest.fn(),
}));

it('should handle API errors', async () => {
  const mockFetchData = require('../api/client').fetchData;
  
  // Debug mock calls
  mockFetchData.mockRejectedValue(new Error('API Error'));
  
  render(<MyComponent />);
  
  await waitFor(() => {
    expect(screen.getByText('Error occurred')).toBeInTheDocument();
  });
  
  // Check mock was called correctly
  console.log('Mock calls:', mockFetchData.mock.calls);
  expect(mockFetchData).toHaveBeenCalledTimes(1);
});
```

### Integration Test Debugging

```typescript
// Debug API integration
test('should create requisition', async () => {
  // Enable MSW logging
  server.use(
    rest.post('/api/requisitions', (req, res, ctx) => {
      console.log('MSW intercepted request:', req.body);
      return res(ctx.json({ id: '123', title: 'Test' }));
    })
  );
  
  render(<CreateRequisitionForm />);
  
  // Fill form and submit
  await user.type(screen.getByLabelText(/title/i), 'Test Requisition');
  await user.click(screen.getByRole('button', { name: /submit/i }));
  
  // Debug network requests
  await waitFor(() => {
    expect(screen.getByText(/success/i)).toBeInTheDocument();
  });
});
```

## Production Debugging

### Error Monitoring Setup

```typescript
// src/lib/error-monitoring.ts
import * as Sentry from '@sentry/nextjs';

export function initErrorMonitoring() {
  if (process.env.NEXT_PUBLIC_SENTRY_DSN) {
    Sentry.init({
      dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
      environment: process.env.NODE_ENV,
      tracesSampleRate: 0.1,
      beforeSend(event, hint) {
        // Filter out known issues
        if (event.exception) {
          const error = event.exception.values?.[0];
          if (error?.type === 'ChunkLoadError') {
            return null;
          }
        }
        
        // Add additional context
        event.extra = {
          ...event.extra,
          timestamp: new Date().toISOString(),
          userAgent: navigator.userAgent,
          url: window.location.href,
        };
        
        return event;
      },
    });
  }
}

// Custom error tracking
export function trackError(error: Error, context?: Record<string, any>) {
  console.error('Application Error:', error);
  
  if (process.env.NODE_ENV === 'production') {
    Sentry.withScope((scope) => {
      if (context) {
        Object.entries(context).forEach(([key, value]) => {
          scope.setContext(key, value);
        });
      }
      Sentry.captureException(error);
    });
  }
}
```

### Remote Debugging

```typescript
// src/lib/remote-debug.ts
export class RemoteDebugger {
  private static logs: any[] = [];
  private static maxLogs = 100;

  static log(level: 'info' | 'warn' | 'error', message: string, data?: any) {
    const logEntry = {
      level,
      message,
      data,
      timestamp: new Date().toISOString(),
      url: window.location.href,
    };

    this.logs.push(logEntry);
    
    // Keep only recent logs
    if (this.logs.length > this.maxLogs) {
      this.logs = this.logs.slice(-this.maxLogs);
    }

    // Send critical errors immediately
    if (level === 'error') {
      this.sendLogs([logEntry]);
    }
  }

  static async sendLogs(logs = this.logs) {
    if (process.env.NODE_ENV !== 'production') return;

    try {
      await fetch('/api/debug/logs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ logs }),
      });
    } catch (error) {
      console.error('Failed to send debug logs:', error);
    }
  }

  static getLogs() {
    return this.logs;
  }

  static clearLogs() {
    this.logs = [];
  }
}

// Usage
RemoteDebugger.log('info', 'User logged in', { userId: '123' });
RemoteDebugger.log('error', 'API call failed', { endpoint: '/api/data', error: 'Network error' });
```

## Quick Reference

### Common Commands

```bash
# Development
npm run dev                 # Start development server
npm run build              # Build for production
npm run start              # Start production server
npm run lint               # Run ESLint
npm run type-check         # Run TypeScript check

# Debugging
npm run analyze            # Analyze bundle size
npm run test -- --verbose  # Run tests with detailed output
npm run test -- --watch    # Run tests in watch mode

# Cleanup
rm -rf node_modules .next  # Clean install
npm cache clean --force    # Clear npm cache
```

### Environment Variables Checklist

```bash
# Required variables
NEXT_PUBLIC_APP_URL        # Application URL
NEXT_PUBLIC_API_URL        # API base URL
AUTH_SECRET               # Authentication secret
NEXTAUTH_URL              # NextAuth URL

# Optional variables
NEXT_PUBLIC_SENTRY_DSN    # Error tracking
NEXT_PUBLIC_GA_ID         # Google Analytics
NODE_ENV                  # Environment
```

### Browser Console Commands

```javascript
// Check React version
React.version

// Access query client (if using React Query)
window.__REACT_QUERY_CLIENT__

// Check performance
performance.getEntriesByType('navigation')
performance.getEntriesByType('paint')

// Memory usage (Chrome only)
performance.memory

// Clear all storage
localStorage.clear()
sessionStorage.clear()
```

This troubleshooting guide provides comprehensive solutions for common issues and debugging techniques to help maintain and debug the Liyali Gateway Frontend effectively.