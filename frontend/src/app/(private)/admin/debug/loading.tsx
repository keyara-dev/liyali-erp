export default function Loading() {
  return (
    <div className="space-y-6">
      {/* Page Header Skeleton */}
      <div>
        <div className="h-8 bg-muted rounded-lg w-48 mb-2 animate-pulse"></div>
        <div className="h-5 bg-muted rounded-lg w-64 animate-pulse"></div>
      </div>

      {/* Debug Tools Tabs */}
      <div className="bg-card rounded-lg border">
        {/* Tab Headers */}
        <div className="flex border-b">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="px-6 py-3">
              <div className="h-5 bg-muted rounded w-24 animate-pulse"></div>
            </div>
          ))}
        </div>

        {/* Tab Content */}
        <div className="p-6">
          {/* System Information */}
          <div className="space-y-6">
            <div>
              <div className="h-6 bg-muted rounded w-40 mb-4 animate-pulse"></div>
              <div className="grid grid-cols-2 gap-4">
                {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
                  <div
                    key={i}
                    className="flex justify-between p-3 bg-muted/20 rounded"
                  >
                    <div className="h-4 bg-muted rounded w-24 animate-pulse"></div>
                    <div className="h-4 bg-muted rounded w-32 animate-pulse"></div>
                  </div>
                ))}
              </div>
            </div>

            {/* Database Status */}
            <div>
              <div className="h-6 bg-muted rounded w-36 mb-4 animate-pulse"></div>
              <div className="space-y-3">
                {[1, 2, 3, 4].map((i) => (
                  <div
                    key={i}
                    className="flex items-center justify-between p-3 bg-muted/20 rounded"
                  >
                    <div className="flex items-center space-x-3">
                      <div className="w-3 h-3 bg-muted rounded-full animate-pulse"></div>
                      <div className="h-4 bg-muted rounded w-32 animate-pulse"></div>
                    </div>
                    <div className="h-4 bg-muted rounded w-20 animate-pulse"></div>
                  </div>
                ))}
              </div>
            </div>

            {/* API Endpoints */}
            <div>
              <div className="h-6 bg-muted rounded w-32 mb-4 animate-pulse"></div>
              <div className="space-y-2">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                  <div
                    key={i}
                    className="flex items-center justify-between p-3 bg-muted/20 rounded"
                  >
                    <div className="flex items-center space-x-3">
                      <div className="h-6 bg-muted rounded w-12 animate-pulse"></div>
                      <div className="h-4 bg-muted rounded w-48 animate-pulse"></div>
                    </div>
                    <div className="flex space-x-2">
                      <div className="h-6 bg-muted rounded-full w-16 animate-pulse"></div>
                      <div className="h-8 bg-muted rounded w-16 animate-pulse"></div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Recent Errors */}
            <div>
              <div className="h-6 bg-muted rounded w-32 mb-4 animate-pulse"></div>
              <div className="space-y-3">
                {[1, 2, 3].map((i) => (
                  <div
                    key={i}
                    className="p-4 bg-red-50 border border-red-200 rounded"
                  >
                    <div className="flex items-start space-x-3">
                      <div className="w-5 h-5 bg-muted rounded animate-pulse"></div>
                      <div className="flex-1 space-y-2">
                        <div className="h-4 bg-muted rounded w-full animate-pulse"></div>
                        <div className="h-3 bg-muted rounded w-32 animate-pulse"></div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
