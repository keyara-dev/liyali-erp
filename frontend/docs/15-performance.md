# Performance Optimization

Comprehensive guide to optimizing the performance of the Liyali Gateway Frontend, covering Core Web Vitals, bundle optimization, runtime performance, and monitoring strategies.

## Performance Metrics

### Core Web Vitals

The application is optimized for Google's Core Web Vitals:

```typescript
// src/lib/web-vitals.ts
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

export function reportWebVitals(onPerfEntry?: (metric: any) => void) {
  if (onPerfEntry && onPerfEntry instanceof Function) {
    getCLS(onPerfEntry);
    getFID(onPerfEntry);
    getFCP(onPerfEntry);
    getLCP(onPerfEntry);
    getTTFB(onPerfEntry);
  }
}

// Custom performance tracking
export class PerformanceTracker {
  private static measurements = new Map<string, number>();

  static startMeasurement(name: string) {
    this.measurements.set(name, performance.now());
  }

  static endMeasurement(name: string): number {
    const startTime = this.measurements.get(name);
    if (!startTime) return 0;

    const duration = performance.now() - startTime;
    this.measurements.delete(name);

    // Report to analytics
    if (typeof window !== 'undefined' && window.gtag) {
      window.gtag('event', 'timing_complete', {
        name,
        value: Math.round(duration),
      });
    }

    return duration;
  }

  static measureAsync<T>(name: string, fn: () => Promise<T>): Promise<T> {
    this.startMeasurement(name);
    return fn().finally(() => {
      this.endMeasurement(name);
    });
  }
}
```

### Performance Targets

- **Largest Contentful Paint (LCP)**: < 2.5s
- **First Input Delay (FID)**: < 100ms
- **Cumulative Layout Shift (CLS)**: < 0.1
- **First Contentful Paint (FCP)**: < 1.8s
- **Time to First Byte (TTFB)**: < 600ms

## Bundle Optimization

### Code Splitting

#### Route-based Splitting

Next.js automatically splits code by routes, but we can optimize further:

```typescript
// src/app/layout.tsx
import dynamic from 'next/dynamic';

// Lazy load non-critical components
const NotificationCenter = dynamic(
  () => import('@/components/layout/notification-center'),
  { ssr: false }
);

const HelpWidget = dynamic(
  () => import('@/components/layout/help-widget'),
  { ssr: false }
);

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        {children}
        <NotificationCenter />
        <HelpWidget />
      </body>
    </html>
  );
}
```

#### Component-based Splitting

```typescript
// src/components/charts/performance-chart.tsx
import { lazy, Suspense } from 'react';

// Lazy load heavy chart library
const Chart = lazy(() => import('react-chartjs-2').then(module => ({
  default: module.Chart
})));

const ChartSkeleton = () => (
  <div className="h-64 bg-muted animate-pulse rounded-lg" />
);

export function PerformanceChart({ data }: { data: any[] }) {
  return (
    <Suspense fallback={<ChartSkeleton />}>
      <Chart data={data} />
    </Suspense>
  );
}
```

#### Dynamic Imports with Loading States

```typescript
// src/hooks/use-dynamic-import.ts
import { useState, useEffect } from 'react';

export function useDynamicImport<T>(
  importFn: () => Promise<{ default: T }>,
  deps: any[] = []
) {
  const [component, setComponent] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    
    setLoading(true);
    setError(null);

    importFn()
      .then(module => {
        if (!cancelled) {
          setComponent(module.default);
        }
      })
      .catch(err => {
        if (!cancelled) {
          setError(err);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, deps);

  return { component, loading, error };
}

// Usage
function PDFViewer({ documentId }: { documentId: string }) {
  const { component: PDFRenderer, loading, error } = useDynamicImport(
    () => import('@/components/pdf/pdf-renderer'),
    [documentId]
  );

  if (loading) return <div>Loading PDF viewer...</div>;
  if (error) return <div>Failed to load PDF viewer</div>;
  if (!PDFRenderer) return null;

  return <PDFRenderer documentId={documentId} />;
}
```

### Bundle Analysis

#### Webpack Bundle Analyzer

```javascript
// next.config.js
const withBundleAnalyzer = require('@next/bundle-analyzer')({
  enabled: process.env.ANALYZE === 'true',
});

module.exports = withBundleAnalyzer({
  // ... other config
  webpack: (config, { buildId, dev, isServer, defaultLoaders, webpack }) => {
    // Optimize bundle splitting
    if (!dev && !isServer) {
      config.optimization.splitChunks = {
        chunks: 'all',
        minSize: 20000,
        maxSize: 244000,
        cacheGroups: {
          default: {
            minChunks: 1,
            priority: -20,
            reuseExistingChunk: true,
          },
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            priority: -10,
            chunks: 'all',
          },
          common: {
            name: 'common',
            minChunks: 2,
            priority: -5,
            chunks: 'all',
            enforce: true,
          },
        },
      };
    }

    return config;
  },
});
```

#### Bundle Size Monitoring

```typescript
// scripts/analyze-bundle.js
const fs = require('fs');
const path = require('path');

function analyzeBundleSize() {
  const buildManifest = require('../.next/build-manifest.json');
  const sizeMap = new Map();

  Object.entries(buildManifest.pages).forEach(([page, files]) => {
    const totalSize = files.reduce((acc, file) => {
      const filePath = path.join('.next', file);
      if (fs.existsSync(filePath)) {
        return acc + fs.statSync(filePath).size;
      }
      return acc;
    }, 0);

    sizeMap.set(page, totalSize);
  });

  // Sort by size
  const sortedPages = Array.from(sizeMap.entries())
    .sort(([, a], [, b]) => b - a)
    .map(([page, size]) => ({
      page,
      size: `${(size / 1024).toFixed(2)} KB`,
    }));

  console.table(sortedPages);
}

analyzeBundleSize();
```

## React Performance Optimization

### Memoization Strategies

#### Component Memoization

```typescript
// src/components/data/requisition-card.tsx
import { memo } from 'react';

interface RequisitionCardProps {
  requisition: Requisition;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
}

export const RequisitionCard = memo(function RequisitionCard({
  requisition,
  onEdit,
  onDelete,
}: RequisitionCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{requisition.title}</CardTitle>
        <CardDescription>REQ-{requisition.documentNumber}</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">
          Amount: ${requisition.totalAmount?.toFixed(2)}
        </p>
        <StatusBadge status={requisition.status} />
      </CardContent>
      <CardFooter>
        <Button variant="outline" onClick={() => onEdit(requisition.id)}>
          Edit
        </Button>
        <Button variant="destructive" onClick={() => onDelete(requisition.id)}>
          Delete
        </Button>
      </CardFooter>
    </Card>
  );
});

// Custom comparison function for complex props
export const RequisitionCardWithCustomComparison = memo(
  RequisitionCard,
  (prevProps, nextProps) => {
    return (
      prevProps.requisition.id === nextProps.requisition.id &&
      prevProps.requisition.updatedAt === nextProps.requisition.updatedAt &&
      prevProps.onEdit === nextProps.onEdit &&
      prevProps.onDelete === nextProps.onDelete
    );
  }
);
```

#### Hook Memoization

```typescript
// src/hooks/use-optimized-data.ts
import { useMemo, useCallback } from 'react';

export function useOptimizedRequisitionData(requisitions: Requisition[]) {
  // Memoize expensive calculations
  const statistics = useMemo(() => {
    return {
      total: requisitions.length,
      pending: requisitions.filter(r => r.status === 'PENDING').length,
      approved: requisitions.filter(r => r.status === 'APPROVED').length,
      totalAmount: requisitions.reduce((sum, r) => sum + (r.totalAmount || 0), 0),
      averageAmount: requisitions.length > 0 
        ? requisitions.reduce((sum, r) => sum + (r.totalAmount || 0), 0) / requisitions.length 
        : 0,
    };
  }, [requisitions]);

  // Memoize filtered data
  const groupedByStatus = useMemo(() => {
    return requisitions.reduce((groups, requisition) => {
      const status = requisition.status;
      if (!groups[status]) {
        groups[status] = [];
      }
      groups[status].push(requisition);
      return groups;
    }, {} as Record<string, Requisition[]>);
  }, [requisitions]);

  // Memoize callback functions
  const getRequisitionsByDepartment = useCallback((department: string) => {
    return requisitions.filter(r => r.department === department);
  }, [requisitions]);

  const getRequisitionsInDateRange = useCallback((startDate: Date, endDate: Date) => {
    return requisitions.filter(r => {
      const createdAt = new Date(r.createdAt);
      return createdAt >= startDate && createdAt <= endDate;
    });
  }, [requisitions]);

  return {
    statistics,
    groupedByStatus,
    getRequisitionsByDepartment,
    getRequisitionsInDateRange,
  };
}
```

### Virtual Scrolling

For large lists, implement virtual scrolling to improve performance:

```typescript
// src/components/ui/virtual-list.tsx
import { FixedSizeList as List } from 'react-window';
import { memo } from 'react';

interface VirtualListProps<T> {
  items: T[];
  height: number;
  itemHeight: number;
  renderItem: (props: { index: number; style: React.CSSProperties; data: T[] }) => React.ReactElement;
}

export function VirtualList<T>({
  items,
  height,
  itemHeight,
  renderItem,
}: VirtualListProps<T>) {
  const Row = memo(({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style}>
      {renderItem({ index, style, data: items })}
    </div>
  ));

  return (
    <List
      height={height}
      itemCount={items.length}
      itemSize={itemHeight}
      itemData={items}
      width="100%"
    >
      {Row}
    </List>
  );
}

// Usage
function RequisitionsList({ requisitions }: { requisitions: Requisition[] }) {
  const renderRequisition = ({ index, style, data }: any) => (
    <div style={style} className="p-2 border-b">
      <RequisitionCard requisition={data[index]} />
    </div>
  );

  return (
    <VirtualList
      items={requisitions}
      height={600}
      itemHeight={120}
      renderItem={renderRequisition}
    />
  );
}
```

## Image Optimization

### Next.js Image Component

```typescript
// src/components/ui/optimized-image.tsx
import Image from 'next/image';
import { useState } from 'react';

interface OptimizedImageProps {
  src: string;
  alt: string;
  width?: number;
  height?: number;
  priority?: boolean;
  className?: string;
}

export function OptimizedImage({
  src,
  alt,
  width = 400,
  height = 300,
  priority = false,
  className,
}: OptimizedImageProps) {
  const [isLoading, setIsLoading] = useState(true);

  return (
    <div className={`relative overflow-hidden ${className}`}>
      <Image
        src={src}
        alt={alt}
        width={width}
        height={height}
        priority={priority}
        className={`transition-opacity duration-300 ${
          isLoading ? 'opacity-0' : 'opacity-100'
        }`}
        onLoadingComplete={() => setIsLoading(false)}
        placeholder="blur"
        blurDataURL="data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR8tMC0oMCUoKSj/2wBDAQcHBwoIChMKChMoGhYaKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCj/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k="
      />
      
      {isLoading && (
        <div className="absolute inset-0 bg-muted animate-pulse" />
      )}
    </div>
  );
}
```

### Image Preloading

```typescript
// src/hooks/use-image-preloader.ts
import { useEffect } from 'react';

export function useImagePreloader(imageSources: string[]) {
  useEffect(() => {
    const preloadImages = imageSources.map(src => {
      const img = new Image();
      img.src = src;
      return img;
    });

    return () => {
      preloadImages.forEach(img => {
        img.src = '';
      });
    };
  }, [imageSources]);
}

// Usage
function Dashboard() {
  // Preload critical images
  useImagePreloader([
    '/images/dashboard-hero.jpg',
    '/images/company-logo.png',
  ]);

  return <div>Dashboard content</div>;
}
```

## Caching Strategies

### HTTP Caching

```typescript
// src/lib/cache-headers.ts
export const CACHE_HEADERS = {
  // Static assets - cache for 1 year
  STATIC: {
    'Cache-Control': 'public, max-age=31536000, immutable',
  },
  
  // API responses - cache for 5 minutes with stale-while-revalidate
  API: {
    'Cache-Control': 'public, max-age=300, stale-while-revalidate=60',
  },
  
  // User-specific data - no cache
  PRIVATE: {
    'Cache-Control': 'private, no-cache, no-store, must-revalidate',
    'Pragma': 'no-cache',
    'Expires': '0',
  },
  
  // HTML pages - cache for 1 hour
  HTML: {
    'Cache-Control': 'public, max-age=3600, stale-while-revalidate=300',
  },
};

// Apply cache headers in API routes
export function withCacheHeaders(handler: any, cacheType: keyof typeof CACHE_HEADERS) {
  return async (req: any, res: any) => {
    const headers = CACHE_HEADERS[cacheType];
    Object.entries(headers).forEach(([key, value]) => {
      res.setHeader(key, value);
    });
    
    return handler(req, res);
  };
}
```

### Service Worker Caching

```typescript
// public/sw.js
const CACHE_NAME = 'liyali-gateway-v1';
const STATIC_ASSETS = [
  '/',
  '/dashboard',
  '/requisitions',
  '/manifest.json',
  '/offline.html',
];

// Install event - cache static assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(STATIC_ASSETS))
  );
});

// Fetch event - serve from cache with network fallback
self.addEventListener('fetch', (event) => {
  if (event.request.method !== 'GET') return;

  event.respondWith(
    caches.match(event.request)
      .then((response) => {
        if (response) {
          return response;
        }

        return fetch(event.request)
          .then((response) => {
            // Don't cache non-successful responses
            if (!response || response.status !== 200 || response.type !== 'basic') {
              return response;
            }

            // Clone the response
            const responseToCache = response.clone();

            caches.open(CACHE_NAME)
              .then((cache) => {
                cache.put(event.request, responseToCache);
              });

            return response;
          })
          .catch(() => {
            // Return offline page for navigation requests
            if (event.request.mode === 'navigate') {
              return caches.match('/offline.html');
            }
          });
      })
  );
});
```

## Database Query Optimization

### Query Optimization with TanStack Query

```typescript
// src/hooks/use-optimized-queries.ts
import { useQuery, useQueries } from '@tanstack/react-query';

// Parallel queries for dashboard data
export function useDashboardData() {
  const queries = useQueries({
    queries: [
      {
        queryKey: ['requisitions', 'summary'],
        queryFn: () => apiClient.requisitions.getSummary(),
        staleTime: 5 * 60 * 1000, // 5 minutes
      },
      {
        queryKey: ['purchase-orders', 'summary'],
        queryFn: () => apiClient.purchaseOrders.getSummary(),
        staleTime: 5 * 60 * 1000,
      },
      {
        queryKey: ['budget', 'utilization'],
        queryFn: () => apiClient.budget.getUtilization(),
        staleTime: 10 * 60 * 1000, // 10 minutes
      },
    ],
  });

  return {
    requisitionsSummary: queries[0],
    purchaseOrdersSummary: queries[1],
    budgetUtilization: queries[2],
    isLoading: queries.some(query => query.isLoading),
    isError: queries.some(query => query.isError),
  };
}

// Prefetch related data
export function usePrefetchRequisitionDetails() {
  const queryClient = useQueryClient();

  const prefetchRequisition = useCallback((id: string) => {
    queryClient.prefetchQuery({
      queryKey: ['requisition', id],
      queryFn: () => apiClient.requisitions.getById(id),
      staleTime: 5 * 60 * 1000,
    });
  }, [queryClient]);

  return { prefetchRequisition };
}
```

### Pagination and Infinite Queries

```typescript
// src/hooks/use-infinite-requisitions.ts
import { useInfiniteQuery } from '@tanstack/react-query';

export function useInfiniteRequisitions(filters?: any) {
  return useInfiniteQuery({
    queryKey: ['requisitions', 'infinite', filters],
    queryFn: ({ pageParam = 1 }) => 
      apiClient.requisitions.getAll({
        page: pageParam,
        limit: 20,
        ...filters,
      }),
    getNextPageParam: (lastPage, pages) => {
      if (lastPage.data.length < 20) return undefined;
      return pages.length + 1;
    },
    staleTime: 5 * 60 * 1000,
  });
}

// Usage with virtual scrolling
function InfiniteRequisitionsList() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteRequisitions();

  const allRequisitions = data?.pages.flatMap(page => page.data) ?? [];

  return (
    <div>
      <VirtualList
        items={allRequisitions}
        height={600}
        itemHeight={120}
        renderItem={({ index, style, data }) => (
          <div style={style}>
            <RequisitionCard requisition={data[index]} />
            {/* Load more when near the end */}
            {index === data.length - 5 && hasNextPage && !isFetchingNextPage && (
              <div ref={() => fetchNextPage()} />
            )}
          </div>
        )}
      />
      
      {isFetchingNextPage && (
        <div className="text-center py-4">Loading more...</div>
      )}
    </div>
  );
}
```

## Runtime Performance Monitoring

### Performance Observer

```typescript
// src/lib/performance-observer.ts
export class PerformanceObserver {
  private static observer: PerformanceObserver | null = null;

  static init() {
    if (typeof window === 'undefined' || this.observer) return;

    this.observer = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        this.handlePerformanceEntry(entry);
      }
    });

    // Observe different types of performance entries
    this.observer.observe({ entryTypes: ['navigation', 'paint', 'largest-contentful-paint'] });
  }

  private static handlePerformanceEntry(entry: PerformanceEntry) {
    switch (entry.entryType) {
      case 'navigation':
        this.trackNavigationTiming(entry as PerformanceNavigationTiming);
        break;
      case 'paint':
        this.trackPaintTiming(entry as PerformancePaintTiming);
        break;
      case 'largest-contentful-paint':
        this.trackLCP(entry);
        break;
    }
  }

  private static trackNavigationTiming(entry: PerformanceNavigationTiming) {
    const metrics = {
      dns: entry.domainLookupEnd - entry.domainLookupStart,
      tcp: entry.connectEnd - entry.connectStart,
      ttfb: entry.responseStart - entry.requestStart,
      download: entry.responseEnd - entry.responseStart,
      domParse: entry.domContentLoadedEventStart - entry.responseEnd,
      domReady: entry.domContentLoadedEventEnd - entry.navigationStart,
      loadComplete: entry.loadEventEnd - entry.navigationStart,
    };

    // Send to analytics
    this.sendMetrics('navigation', metrics);
  }

  private static trackPaintTiming(entry: PerformancePaintTiming) {
    const metrics = {
      [entry.name]: entry.startTime,
    };

    this.sendMetrics('paint', metrics);
  }

  private static trackLCP(entry: PerformanceEntry) {
    const metrics = {
      lcp: entry.startTime,
    };

    this.sendMetrics('lcp', metrics);
  }

  private static sendMetrics(type: string, metrics: Record<string, number>) {
    if (process.env.NODE_ENV === 'production') {
      // Send to your analytics service
      console.log(`Performance ${type}:`, metrics);
    }
  }
}

// Initialize in app
if (typeof window !== 'undefined') {
  PerformanceObserver.init();
}
```

### Memory Usage Monitoring

```typescript
// src/hooks/use-memory-monitor.ts
import { useEffect, useState } from 'react';

export function useMemoryMonitor() {
  const [memoryInfo, setMemoryInfo] = useState<any>(null);

  useEffect(() => {
    if (typeof window === 'undefined' || !('memory' in performance)) return;

    const updateMemoryInfo = () => {
      const memory = (performance as any).memory;
      setMemoryInfo({
        usedJSHeapSize: memory.usedJSHeapSize,
        totalJSHeapSize: memory.totalJSHeapSize,
        jsHeapSizeLimit: memory.jsHeapSizeLimit,
        usedPercentage: (memory.usedJSHeapSize / memory.jsHeapSizeLimit) * 100,
      });
    };

    updateMemoryInfo();
    const interval = setInterval(updateMemoryInfo, 10000); // Every 10 seconds

    return () => clearInterval(interval);
  }, []);

  return memoryInfo;
}

// Usage in development
function MemoryMonitor() {
  const memoryInfo = useMemoryMonitor();

  if (process.env.NODE_ENV !== 'development' || !memoryInfo) return null;

  return (
    <div className="fixed bottom-4 right-4 bg-black text-white p-2 rounded text-xs">
      <div>Used: {(memoryInfo.usedJSHeapSize / 1024 / 1024).toFixed(2)} MB</div>
      <div>Total: {(memoryInfo.totalJSHeapSize / 1024 / 1024).toFixed(2)} MB</div>
      <div>Usage: {memoryInfo.usedPercentage.toFixed(1)}%</div>
    </div>
  );
}
```

## Performance Testing

### Lighthouse CI

```yaml
# .github/workflows/lighthouse.yml
name: Lighthouse CI

on:
  pull_request:
    branches: [main]

jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'pnpm'
      
      - name: Install dependencies
        run: pnpm install
      
      - name: Build application
        run: pnpm build
      
      - name: Start server
        run: pnpm start &
        
      - name: Wait for server
        run: npx wait-on http://localhost:3000
      
      - name: Run Lighthouse CI
        run: |
          npm install -g @lhci/cli@0.12.x
          lhci autorun
        env:
          LHCI_GITHUB_APP_TOKEN: ${{ secrets.LHCI_GITHUB_APP_TOKEN }}
```

```javascript
// lighthouserc.js
module.exports = {
  ci: {
    collect: {
      url: [
        'http://localhost:3000/',
        'http://localhost:3000/dashboard',
        'http://localhost:3000/requisitions',
      ],
      numberOfRuns: 3,
    },
    assert: {
      assertions: {
        'categories:performance': ['error', { minScore: 0.9 }],
        'categories:accessibility': ['error', { minScore: 0.9 }],
        'categories:best-practices': ['error', { minScore: 0.9 }],
        'categories:seo': ['error', { minScore: 0.9 }],
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
};
```

### Performance Budget

```javascript
// performance-budget.js
module.exports = {
  budgets: [
    {
      resourceSizes: [
        { resourceType: 'script', budget: 400 }, // 400KB for JS
        { resourceType: 'stylesheet', budget: 100 }, // 100KB for CSS
        { resourceType: 'image', budget: 500 }, // 500KB for images
        { resourceType: 'total', budget: 1000 }, // 1MB total
      ],
      resourceCounts: [
        { resourceType: 'script', budget: 10 },
        { resourceType: 'stylesheet', budget: 5 },
        { resourceType: 'third-party', budget: 5 },
      ],
    },
  ],
};
```

## Best Practices Summary

### Development Guidelines

1. **Measure First**: Always measure performance before optimizing
2. **Progressive Enhancement**: Build for the slowest devices first
3. **Critical Path**: Optimize the critical rendering path
4. **Lazy Loading**: Load non-critical resources on demand
5. **Code Splitting**: Split code at logical boundaries

### Runtime Optimizations

1. **Memoization**: Use React.memo, useMemo, and useCallback appropriately
2. **Virtual Scrolling**: For large lists and tables
3. **Image Optimization**: Use Next.js Image component with proper sizing
4. **Bundle Splitting**: Optimize chunk sizes and loading strategies
5. **Caching**: Implement proper caching at all levels

### Monitoring and Maintenance

1. **Core Web Vitals**: Monitor and maintain good scores
2. **Bundle Analysis**: Regular bundle size analysis
3. **Performance Budgets**: Set and enforce performance budgets
4. **Real User Monitoring**: Track real user performance metrics
5. **Regular Audits**: Conduct regular performance audits

This performance optimization guide provides a comprehensive approach to building and maintaining a fast, efficient React application that delivers excellent user experience across all devices and network conditions.