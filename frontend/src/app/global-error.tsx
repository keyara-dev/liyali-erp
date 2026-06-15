"use client";

import { AlertCircle } from "lucide-react";

export default function GlobalError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body>
        <div className="grid min-h-screen place-content-center place-items-center bg-linear-to-br from-gray-50 via-white to-purple-50 p-8">
          <div className="max-w-lg text-center">
            <div className="mb-4 flex justify-center">
              <AlertCircle className="h-20 w-20 text-red-500" />
            </div>
            <h1 className="mb-4 text-3xl font-bold text-gray-900">
              Something went wrong
            </h1>
            <p className="mb-8 text-lg text-gray-600">
              An unexpected error occurred. Please try again, and contact
              support if the problem persists.
            </p>
            <button
              onClick={() => reset()}
              className="inline-flex items-center gap-2 rounded-full px-6 py-3 font-medium text-white shadow-xl transition-all duration-200 hover:scale-105 hover:opacity-90"
              style={{ backgroundColor: "#0c54e7" }}
            >
              Try again
            </button>
          </div>
        </div>
      </body>
    </html>
  );
}
