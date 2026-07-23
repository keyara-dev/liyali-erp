import ErrorDisplay from "@/components/base/error-display";
import Link from "next/link";
import { ArrowLeftIcon } from "lucide-react";

export default function NotFound() {
  return (
    <ErrorDisplay
      status={404}
      title="Page Not Found"
      message="Oops! The page you're looking for doesn't exist or has been moved."
    >
      <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
        <Link
          href="/home"
          className="inline-flex items-center gap-3 px-4 py-2 bg-primary text-white rounded-full hover:bg-primary/80 transition-all duration-200 transform hover:scale-105 shadow-xl hover:shadow-2xl font-medium text-lg group"
        >
          <ArrowLeftIcon className="w-5 h-5 transition-transform group-hover:-translate-x-1" />
          Back to Home
        </Link>
      </div>
    </ErrorDisplay>
  );
}
