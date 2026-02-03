export default function Loading() {
  return (
    <div className="space-y-6">
      {/* Page Header Skeleton */}
      <div>
        <div className="h-8 bg-muted rounded-lg w-48 mb-2 animate-pulse"></div>
        <div className="h-5 bg-muted rounded-lg w-64 animate-pulse"></div>
      </div>

      {/* System Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-card rounded-lg border p-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="h-4 bg-muted rounded w-20 mb-2 animate-pulse"></div>
                <div className="h-6 bg-muted rounded-full w-16 animate-pulse"></div>
              </div>
              <div className="w-8 h-8 bg-muted rounded-full animate-pulse"></div>
            </div>
          </div>
        ))}
      </div>

      {/* Monitoring Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {[1, 2].map((i) => (
          <div key={i} className="bg-card rounded-lg border p-6">
            <div className="flex items-center justify-between mb-6">
              <div className="h-6 bg-muted rounded w-48 animate-pulse"></div>
              <div className="h-4 bg-muted rounded w-24 animate-pulse"></div>
            </div>
            <div className="h-[300px] bg-muted rounded-lg animate-pulse flex items-center justify-center">
              <div className="text-muted-foreground">Loading chart...</div>
            </div>
          </div>
        ))}
      </div>

      {/* System Logs */}
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center justify-between mb-6">
          <div className="h-6 bg-muted rounded w-32 mb-4 animate-pulse"></div>
          <div className="flex space-x-2">
            <div className="h-8 bg-muted rounded w-20 animate-pulse"></div>
            <div className="h-8 bg-muted rounded w-24 animate-pulse"></div>
          </div>
        </div>

        <div className="space-y-3">
          {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
            <div
              key={i}
              className="flex items-center space-x-4 p-3 bg-muted/20 rounded"
            >
              <div className="h-4 bg-muted rounded w-20 animate-pulse"></div>
              <div className="h-6 bg-muted rounded-full w-16 animate-pulse"></div>
              <div className="flex-1 h-4 bg-muted rounded animate-pulse"></div>
              <div className="h-4 bg-muted rounded w-24 animate-pulse"></div>
            </div>
          ))}
        </div>
      </div>

      {/* Performance Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-card rounded-lg border p-6">
            <div className="h-6 bg-muted rounded w-40 mb-4 animate-pulse"></div>
            <div className="space-y-3">
              {[1, 2, 3, 4].map((j) => (
                <div key={j} className="flex justify-between items-center">
                  <div className="h-4 bg-muted rounded w-24 animate-pulse"></div>
                  <div className="h-4 bg-muted rounded w-16 animate-pulse"></div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
