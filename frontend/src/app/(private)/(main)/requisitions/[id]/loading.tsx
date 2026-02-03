export default function Loading() {
  return (
    <div className="space-y-6">
      {/* Header Skeleton */}
      <div className="flex items-center justify-between">
        <div>
          <div className="h-8 bg-muted rounded-lg w-64 mb-2 animate-pulse"></div>
          <div className="h-5 bg-muted rounded-lg w-48 animate-pulse"></div>
        </div>
        <div className="flex space-x-2">
          <div className="h-10 bg-muted rounded-md w-24 animate-pulse"></div>
          <div className="h-10 bg-muted rounded-md w-24 animate-pulse"></div>
        </div>
      </div>

      {/* Status Badge Skeleton */}
      <div className="h-6 bg-muted rounded-full w-20 animate-pulse"></div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Details */}
        <div className="lg:col-span-2 space-y-6">
          {/* Basic Information Card */}
          <div className="bg-card rounded-lg border p-6">
            <div className="h-6 bg-muted rounded w-40 mb-4 animate-pulse"></div>
            <div className="grid grid-cols-2 gap-4">
              {[1, 2, 3, 4, 5, 6].map((i) => (
                <div key={i} className="space-y-2">
                  <div className="h-4 bg-muted rounded w-24 animate-pulse"></div>
                  <div className="h-5 bg-muted rounded w-32 animate-pulse"></div>
                </div>
              ))}
            </div>
          </div>

          {/* Items Table */}
          <div className="bg-card rounded-lg border p-6">
            <div className="h-6 bg-muted rounded w-32 mb-4 animate-pulse"></div>
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="grid grid-cols-4 gap-4 p-3 bg-muted/20 rounded"
                >
                  <div className="h-4 bg-muted rounded animate-pulse"></div>
                  <div className="h-4 bg-muted rounded animate-pulse"></div>
                  <div className="h-4 bg-muted rounded animate-pulse"></div>
                  <div className="h-4 bg-muted rounded animate-pulse"></div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Right Column - Actions & History */}
        <div className="space-y-6">
          {/* Actions Card */}
          <div className="bg-card rounded-lg border p-6">
            <div className="h-6 bg-muted rounded w-24 mb-4 animate-pulse"></div>
            <div className="space-y-3">
              <div className="h-10 bg-muted rounded-md animate-pulse"></div>
              <div className="h-10 bg-muted rounded-md animate-pulse"></div>
            </div>
          </div>

          {/* History Card */}
          <div className="bg-card rounded-lg border p-6">
            <div className="h-6 bg-muted rounded w-32 mb-4 animate-pulse"></div>
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex space-x-3">
                  <div className="w-8 h-8 bg-muted rounded-full animate-pulse"></div>
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-muted rounded w-full animate-pulse"></div>
                    <div className="h-3 bg-muted rounded w-24 animate-pulse"></div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
