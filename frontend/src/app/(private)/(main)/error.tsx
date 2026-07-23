"use client";

import ErrorDisplay from "@/components/base/error-display";
import { Button } from "@/components/ui/button";
import { RefreshCw } from "lucide-react";

export default function Error({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <ErrorDisplay
      status={500}
      title="Something went wrong"
      message="An unexpected error occurred while loading this page. Please try again, and contact support if the problem persists."
    >
      <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
        <Button
          onClick={() => reset()}
          className="inline-flex items-center gap-3 px-4 py-2 bg-primary text-white rounded-full hover:bg-primary/80 transition-all duration-200 transform hover:scale-105 shadow-xl hover:shadow-2xl font-medium text-lg group"
        >
          <RefreshCw className="w-5 h-5" />
          Try again
        </Button>
      </div>
    </ErrorDisplay>
  );
}
