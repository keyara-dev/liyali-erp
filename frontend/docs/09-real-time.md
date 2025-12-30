# Real-time Features

The Liyali Gateway frontend implements real-time updates and notifications to keep users informed of changes across the system, even when multiple users are working simultaneously.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Real-time Architecture                   │
├─────────────────────────────────────────────────────────────┤
│ 1. WebSocket Connection (Future)                            │
│    ↕ Real-time bidirectional communication                 │
│                                                             │
│ 2. Server-Sent Events (Current)                            │
│    ↕ Server-to-client notifications                        │
│                                                             │
│ 3. Polling Strategy (Fallback)                             │
│    ↕ Regular API polling for updates                       │
│                                                             │
│ 4. Optimistic Updates                                      │
│    ↕ Immediate UI feedback with rollback                   │
│                                                             │
│ 5. Background Sync                                         │
│    ↕ Automatic data synchronization                        │
└─────────────────────────────────────────────────────────────┘
```

## Optimistic Updates

### Implementation Pattern

All mutations use optimistic updates for immediate user feedback:

```typescript
// src/hooks/use-optimistic-mutations.ts
export function useOptimisticRequisitionUpdate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateRequisition,
    
    // Optimistic update before API call
    onMutate: async (updatedRequisition) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ 
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL] 
      });

      // Snapshot previous value
      const previousRequisitions = queryClient.getQueryData([
        QUERY_KEYS.REQUISITIONS.ALL
      ]);

      // Optimistically update cache
      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.ALL],
        (old: Requisition[]) => 
          old?.map(req => 
            req.id === updatedRequisition.id 
              ? { ...req, ...updatedRequisition }
              : req
          ) || []
      );

      // Update individual requisition cache
      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.BY_ID, updatedRequisition.id],
        updatedRequisition
      );

      return { previousRequisitions };
    },

    // Rollback on error
    onError: (err, updatedRequisition, context) => {
      if (context?.previousRequisitions) {
        queryClient.setQueryData(
          [QUERY_KEYS.REQUISITIONS.ALL],
          context.previousRequisitions
        );
      }
      
      toast.error('Failed to update requisition');
    },

    // Always refetch to ensure consistency
    onSettled: () => {
      queryClient.invalidateQueries({ 
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL] 
      });
    },

    onSuccess: (data) => {
      toast.success('Requisition updated successfully');
      
      // Update related caches
      queryClient.invalidateQueries({ 
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS] 
      });
    },
  });
}
```

### Optimistic UI Components

Components show immediate feedback while operations are pending:

```typescript
// src/components/workflows/optimistic-approval-button.tsx
interface OptimisticApprovalButtonProps {
  requisitionId: string;
  currentStatus: string;
  onApprove: (id: string) => Promise<void>;
}

export function OptimisticApprovalButton({
  requisitionId,
  currentStatus,
  onApprove,
}: OptimisticApprovalButtonProps) {
  const [isOptimistic, setIsOptimistic] = useState(false);
  const [optimisticStatus, setOptimisticStatus] = useState(currentStatus);

  const handleApprove = async () => {
    // Show optimistic state immediately
    setIsOptimistic(true);
    setOptimisticStatus('APPROVED');

    try {
      await onApprove(requisitionId);
      // Success - optimistic state matches reality
      setIsOptimistic(false);
    } catch (error) {
      // Rollback optimistic state
      setIsOptimistic(false);
      setOptimisticStatus(currentStatus);
      toast.error('Failed to approve requisition');
    }
  };

  const displayStatus = isOptimistic ? optimisticStatus : currentStatus;

  return (
    <div className="flex items-center gap-2">
      <Button
        onClick={handleApprove}
        disabled={isOptimistic || displayStatus === 'APPROVED'}
        className={cn(
          isOptimistic && "opacity-70 cursor-wait"
        )}
      >
        {isOptimistic ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Approving...
          </>
        ) : displayStatus === 'APPROVED' ? (
          <>
            <CheckCircle className="w-4 h-4 mr-2" />
            Approved
          </>
        ) : (
          <>
            <ThumbsUp className="w-4 h-4 mr-2" />
            Approve
          </>
        )}
      </Button>

      {isOptimistic && (
        <Badge variant="outline" className="animate-pulse">
          Pending
        </Badge>
      )}
    </div>
  );
}
```

## Background Synchronization

### Auto-Refresh Strategy

Data is automatically refreshed based on user activity and data staleness:

```typescript
// src/hooks/use-background-sync.ts
export function useBackgroundSync() {
  const queryClient = useQueryClient();
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [lastActivity, setLastActivity] = useState(Date.now());

  // Track user activity
  useEffect(() => {
    const updateActivity = () => setLastActivity(Date.now());
    
    const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart'];
    events.forEach(event => {
      document.addEventListener(event, updateActivity, true);
    });

    return () => {
      events.forEach(event => {
        document.removeEventListener(event, updateActivity, true);
      });
    };
  }, []);

  // Online/offline detection
  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      // Refetch all queries when coming back online
      queryClient.refetchQueries();
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
  }, [queryClient]);

  // Background refresh for active users
  useEffect(() => {
    if (!isOnline) return;

    const interval = setInterval(() => {
      const timeSinceActivity = Date.now() - lastActivity;
      
      // Only refresh if user was active in last 5 minutes
      if (timeSinceActivity < 5 * 60 * 1000) {
        queryClient.refetchQueries({
          stale: true,
          type: 'active',
        });
      }
    }, 30000); // Check every 30 seconds

    return () => clearInterval(interval);
  }, [isOnline, lastActivity, queryClient]);

  return { isOnline, lastActivity };
}
```

### Smart Invalidation

Cache invalidation is optimized to minimize unnecessary requests:

```typescript
// src/lib/cache-invalidation.ts
export class SmartInvalidation {
  private static pendingInvalidations = new Set<string>();
  private static invalidationTimer: NodeJS.Timeout | null = null;

  static scheduleInvalidation(queryKey: string, delay = 100) {
    this.pendingInvalidations.add(queryKey);

    // Debounce invalidations
    if (this.invalidationTimer) {
      clearTimeout(this.invalidationTimer);
    }

    this.invalidationTimer = setTimeout(() => {
      this.executePendingInvalidations();
    }, delay);
  }

  private static executePendingInvalidations() {
    const queryClient = getQueryClient();
    
    // Group related invalidations
    const groups = this.groupInvalidations([...this.pendingInvalidations]);
    
    groups.forEach(group => {
      queryClient.invalidateQueries({ queryKey: group });
    });

    this.pendingInvalidations.clear();
    this.invalidationTimer = null;
  }

  private static groupInvalidations(keys: string[]): string[][] {
    const groups: { [key: string]: string[] } = {};

    keys.forEach(key => {
      const baseKey = key.split('.')[0];
      if (!groups[baseKey]) {
        groups[baseKey] = [];
      }
      groups[baseKey].push(key);
    });

    return Object.values(groups);
  }
}

// Usage in mutations
export function useSmartMutation() {
  return useMutation({
    mutationFn: updateDocument,
    onSuccess: () => {
      // Schedule smart invalidation instead of immediate
      SmartInvalidation.scheduleInvalidation('requisitions');
      SmartInvalidation.scheduleInvalidation('dashboard.metrics');
    },
  });
}
```

## Live Notifications

### Toast Notification System

Real-time feedback through toast notifications:

```typescript
// src/components/ui/toast-provider.tsx
import { Toaster } from 'sonner';

export function ToastProvider() {
  return (
    <Toaster
      position="top-right"
      expand={true}
      richColors={true}
      closeButton={true}
      toastOptions={{
        duration: 4000,
        style: {
          background: 'hsl(var(--background))',
          color: 'hsl(var(--foreground))',
          border: '1px solid hsl(var(--border))',
        },
      }}
    />
  );
}

// Enhanced toast functions
export const toast = {
  success: (message: string, options?: any) => {
    return sonnerToast.success(message, {
      icon: '✅',
      ...options,
    });
  },

  error: (message: string, options?: any) => {
    return sonnerToast.error(message, {
      icon: '❌',
      ...options,
    });
  },

  loading: (message: string, options?: any) => {
    return sonnerToast.loading(message, {
      icon: '⏳',
      ...options,
    });
  },

  promise: <T>(
    promise: Promise<T>,
    {
      loading,
      success,
      error,
    }: {
      loading: string;
      success: string | ((data: T) => string);
      error: string | ((error: any) => string);
    }
  ) => {
    return sonnerToast.promise(promise, {
      loading,
      success,
      error,
    });
  },
};
```

### Activity Feed

Real-time activity feed showing system-wide changes:

```typescript
// src/components/layout/activity-feed.tsx
interface ActivityItem {
  id: string;
  type: 'requisition_created' | 'requisition_approved' | 'po_generated';
  user: string;
  message: string;
  timestamp: Date;
  metadata?: any;
}

export function ActivityFeed() {
  const [activities, setActivities] = useState<ActivityItem[]>([]);
  const [isOpen, setIsOpen] = useState(false);

  // Simulate real-time updates (replace with actual WebSocket/SSE)
  useEffect(() => {
    const interval = setInterval(() => {
      // In real implementation, this would come from WebSocket
      fetchRecentActivities().then(newActivities => {
        setActivities(prev => {
          const combined = [...newActivities, ...prev];
          return combined.slice(0, 50); // Keep last 50 activities
        });
      });
    }, 30000); // Check every 30 seconds

    return () => clearInterval(interval);
  }, []);

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="w-5 h-5" />
          {activities.length > 0 && (
            <Badge 
              variant="destructive" 
              className="absolute -top-1 -right-1 h-5 w-5 p-0 flex items-center justify-center text-xs"
            >
              {activities.length > 9 ? '9+' : activities.length}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>

      <PopoverContent className="w-80 p-0" align="end">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Recent Activity</h3>
        </div>

        <ScrollArea className="h-96">
          <div className="p-2">
            {activities.length === 0 ? (
              <div className="text-center text-muted-foreground py-8">
                No recent activity
              </div>
            ) : (
              activities.map(activity => (
                <ActivityItem key={activity.id} activity={activity} />
              ))
            )}
          </div>
        </ScrollArea>

        <div className="p-2 border-t">
          <Button variant="ghost" size="sm" className="w-full">
            View All Activity
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

function ActivityItem({ activity }: { activity: ActivityItem }) {
  const getIcon = (type: string) => {
    switch (type) {
      case 'requisition_created':
        return <FileText className="w-4 h-4 text-blue-500" />;
      case 'requisition_approved':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'po_generated':
        return <ShoppingCart className="w-4 h-4 text-purple-500" />;
      default:
        return <Activity className="w-4 h-4 text-gray-500" />;
    }
  };

  return (
    <div className="flex items-start gap-3 p-2 rounded-lg hover:bg-muted/50 transition-colors">
      <div className="flex-shrink-0 mt-0.5">
        {getIcon(activity.type)}
      </div>
      
      <div className="flex-1 min-w-0">
        <p className="text-sm text-foreground">
          <span className="font-medium">{activity.user}</span>{' '}
          {activity.message}
        </p>
        <p className="text-xs text-muted-foreground mt-1">
          {formatDistanceToNow(activity.timestamp, { addSuffix: true })}
        </p>
      </div>
    </div>
  );
}
```

## Connection Status

### Network Status Indicator

Visual feedback about connection status:

```typescript
// src/components/layout/connection-status.tsx
export function ConnectionStatus() {
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [lastOnline, setLastOnline] = useState<Date | null>(null);

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      setLastOnline(null);
    };

    const handleOffline = () => {
      setIsOnline(false);
      setLastOnline(new Date());
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  if (isOnline) {
    return null; // Don't show anything when online
  }

  return (
    <div className="fixed bottom-4 left-4 right-4 z-50">
      <Alert variant="destructive" className="max-w-md mx-auto">
        <WifiOff className="h-4 w-4" />
        <AlertTitle>You're offline</AlertTitle>
        <AlertDescription>
          {lastOnline ? (
            <>
              Connection lost {formatDistanceToNow(lastOnline, { addSuffix: true })}.
              Your changes are being saved locally.
            </>
          ) : (
            'Your changes are being saved locally and will sync when you reconnect.'
          )}
        </AlertDescription>
      </Alert>
    </div>
  );
}
```

## Conflict Resolution

### Merge Strategies

When conflicts occur during sync, the system uses smart merge strategies:

```typescript
// src/lib/conflict-resolution.ts
export interface ConflictResolution {
  strategy: 'client_wins' | 'server_wins' | 'merge' | 'manual';
  resolvedData?: any;
  requiresUserInput?: boolean;
}

export class ConflictResolver {
  static resolveDocumentConflict(
    clientVersion: any,
    serverVersion: any,
    lastSyncVersion?: any
  ): ConflictResolution {
    // If we have the last sync version, we can do a three-way merge
    if (lastSyncVersion) {
      return this.threeWayMerge(clientVersion, serverVersion, lastSyncVersion);
    }

    // Simple conflict resolution rules
    const clientModified = new Date(clientVersion.updatedAt);
    const serverModified = new Date(serverVersion.updatedAt);

    // If server version is newer, prefer server
    if (serverModified > clientModified) {
      return {
        strategy: 'server_wins',
        resolvedData: serverVersion,
      };
    }

    // If client has unsaved changes, prefer client
    if (clientVersion.hasUnsavedChanges) {
      return {
        strategy: 'client_wins',
        resolvedData: clientVersion,
      };
    }

    // Default to server version
    return {
      strategy: 'server_wins',
      resolvedData: serverVersion,
    };
  }

  private static threeWayMerge(
    client: any,
    server: any,
    base: any
  ): ConflictResolution {
    const merged = { ...base };
    let hasConflicts = false;

    // Merge non-conflicting changes
    Object.keys(client).forEach(key => {
      const clientValue = client[key];
      const serverValue = server[key];
      const baseValue = base[key];

      if (clientValue === serverValue) {
        // No conflict
        merged[key] = clientValue;
      } else if (clientValue === baseValue) {
        // Only server changed
        merged[key] = serverValue;
      } else if (serverValue === baseValue) {
        // Only client changed
        merged[key] = clientValue;
      } else {
        // Both changed - conflict
        hasConflicts = true;
        merged[key] = clientValue; // Prefer client for now
      }
    });

    return {
      strategy: hasConflicts ? 'manual' : 'merge',
      resolvedData: merged,
      requiresUserInput: hasConflicts,
    };
  }
}
```

### Conflict Resolution UI

When manual resolution is needed, show a conflict resolution dialog:

```typescript
// src/components/base/conflict-resolution-dialog.tsx
interface ConflictResolutionDialogProps {
  isOpen: boolean;
  onClose: () => void;
  clientVersion: any;
  serverVersion: any;
  onResolve: (resolution: 'client' | 'server' | 'merge') => void;
}

export function ConflictResolutionDialog({
  isOpen,
  onClose,
  clientVersion,
  serverVersion,
  onResolve,
}: ConflictResolutionDialogProps) {
  const [selectedResolution, setSelectedResolution] = useState<'client' | 'server' | 'merge'>('client');

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl">
        <DialogHeader>
          <DialogTitle>Resolve Conflict</DialogTitle>
          <DialogDescription>
            This document was modified both locally and on the server. 
            Choose how to resolve the conflict.
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label className="text-sm font-medium">Your Version (Local)</Label>
            <Card className="p-4">
              <pre className="text-xs overflow-auto max-h-64">
                {JSON.stringify(clientVersion, null, 2)}
              </pre>
            </Card>
            <Button
              variant={selectedResolution === 'client' ? 'default' : 'outline'}
              onClick={() => setSelectedResolution('client')}
              className="w-full"
            >
              Use My Version
            </Button>
          </div>

          <div className="space-y-2">
            <Label className="text-sm font-medium">Server Version</Label>
            <Card className="p-4">
              <pre className="text-xs overflow-auto max-h-64">
                {JSON.stringify(serverVersion, null, 2)}
              </pre>
            </Card>
            <Button
              variant={selectedResolution === 'server' ? 'default' : 'outline'}
              onClick={() => setSelectedResolution('server')}
              className="w-full"
            >
              Use Server Version
            </Button>
          </div>
        </div>

        <div className="flex justify-between">
          <Button
            variant={selectedResolution === 'merge' ? 'default' : 'outline'}
            onClick={() => setSelectedResolution('merge')}
          >
            Smart Merge
          </Button>

          <div className="space-x-2">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button onClick={() => onResolve(selectedResolution)}>
              Resolve Conflict
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

## Performance Optimization

### Debounced Updates

Prevent excessive API calls with debounced updates:

```typescript
// src/hooks/use-debounced-mutation.ts
export function useDebouncedMutation<T, V>(
  mutationFn: (variables: V) => Promise<T>,
  delay = 500
) {
  const [debouncedMutate] = useDebouncedCallback(
    mutationFn,
    delay,
    { leading: false, trailing: true }
  );

  return useMutation({
    mutationFn: debouncedMutate,
  });
}

// Usage for auto-save
export function useAutoSave<T>(
  data: T,
  saveFn: (data: T) => Promise<void>,
  delay = 2000
) {
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  const debouncedSave = useDebouncedCallback(
    async (dataToSave: T) => {
      try {
        await saveFn(dataToSave);
        setLastSaved(new Date());
        setHasUnsavedChanges(false);
      } catch (error) {
        console.error('Auto-save failed:', error);
      }
    },
    delay,
    { leading: false, trailing: true }
  );

  useEffect(() => {
    if (data) {
      setHasUnsavedChanges(true);
      debouncedSave(data);
    }
  }, [data, debouncedSave]);

  return {
    lastSaved,
    hasUnsavedChanges,
    forceSave: () => debouncedSave.flush(),
  };
}
```

## Future Enhancements

### WebSocket Integration

Planned WebSocket implementation for true real-time updates:

```typescript
// src/lib/websocket.ts (Future implementation)
export class WebSocketManager {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;

  connect() {
    try {
      this.ws = new WebSocket(process.env.NEXT_PUBLIC_WS_URL!);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
      };

      this.ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.attemptReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
    }
  }

  private handleMessage(message: any) {
    switch (message.type) {
      case 'document_updated':
        // Invalidate relevant queries
        break;
      case 'approval_required':
        // Show notification
        break;
      case 'user_activity':
        // Update activity feed
        break;
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        this.connect();
      }, Math.pow(2, this.reconnectAttempts) * 1000);
    }
  }
}
```

The real-time features provide immediate feedback and keep users synchronized across the application, creating a responsive and collaborative experience even in offline scenarios.