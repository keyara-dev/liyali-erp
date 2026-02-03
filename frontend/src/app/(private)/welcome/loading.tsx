export default function Loading() {
  return (
    <div className="space-y-6">
      {/* Welcome Header Skeleton */}
      <div className="text-center space-y-4">
        <div className="h-10 bg-muted rounded-lg w-64 mx-auto animate-pulse"></div>
        <div className="h-5 bg-muted rounded-lg w-80 mx-auto animate-pulse"></div>
      </div>

      {/* Organization Selection Card */}
      <div className="bg-card rounded-lg border p-8 max-w-2xl mx-auto">
        <div className="space-y-6">
          <div className="text-center">
            <div className="h-6 bg-muted rounded w-48 mx-auto mb-2 animate-pulse"></div>
            <div className="h-4 bg-muted rounded w-64 mx-auto animate-pulse"></div>
          </div>

          {/* Organization Options */}
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="p-4 border rounded-lg">
                <div className="flex items-center space-x-4">
                  <div className="w-12 h-12 bg-muted rounded-full animate-pulse"></div>
                  <div className="flex-1 space-y-2">
                    <div className="h-5 bg-muted rounded w-48 animate-pulse"></div>
                    <div className="h-4 bg-muted rounded w-32 animate-pulse"></div>
                  </div>
                  <div className="w-4 h-4 bg-muted rounded-full animate-pulse"></div>
                </div>
              </div>
            ))}
          </div>

          {/* Create New Organization Option */}
          <div className="border-t pt-6">
            <div className="p-4 border-2 border-dashed rounded-lg text-center">
              <div className="w-8 h-8 bg-muted rounded mx-auto mb-2 animate-pulse"></div>
              <div className="h-5 bg-muted rounded w-40 mx-auto mb-1 animate-pulse"></div>
              <div className="h-4 bg-muted rounded w-56 mx-auto animate-pulse"></div>
            </div>
          </div>

          {/* Continue Button */}
          <div className="text-center">
            <div className="h-12 bg-muted rounded-md w-32 mx-auto animate-pulse"></div>
          </div>
        </div>
      </div>

      {/* Quick Start Guide */}
      <div className="bg-card rounded-lg border p-6 max-w-2xl mx-auto">
        <div className="h-6 bg-muted rounded w-32 mb-4 animate-pulse"></div>
        <div className="space-y-3">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="flex items-center space-x-3">
              <div className="w-6 h-6 bg-muted rounded-full animate-pulse"></div>
              <div className="h-4 bg-muted rounded w-64 animate-pulse"></div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
