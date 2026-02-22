"use client";
import React, { PropsWithChildren } from "react";

import Link from "next/link";
import { ArrowLeftIcon, AlertCircle } from "lucide-react";
import { Button } from "../ui/button";

function ErrorDisplay({
  status = 404,
  title = "Store Not Found",
  message = "Oops! This store doesn't exist or the URL is incorrect.",
  showBackButton = false,
  children,
}: PropsWithChildren & {
  status?: number;
  title?: string;
  message?: string;
  showBackButton?: boolean;
}) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-purple-50 dark:from-gray-950 dark:via-gray-900 dark:to-purple-950 flex items-center justify-center relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-gradient-to-r from-purple-200 to-pink-200 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000"></div>
      <div className="absolute top-1/2 right-1/4 w-96 h-96 bg-gradient-to-r from-blue-200 to-cyan-200 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-4000"></div>
      <div className="absolute bottom-1/4 left-1/2 w-96 h-96 bg-gradient-to-r from-red-200 to-orange-200 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob"></div>

      <div className="relative z-10 text-center max-w-2xl mx-auto p-8 animate-fade-in">
        {/* Error Icon */}
        <div className="mb-8 flex justify-center">
          <AlertCircle className="w-24 h-24 text-muted-foreground" />
        </div>

        {/* Dramatic Typography */}
        <h1 className="text-7xl md:text-8xl font-light text-foreground mb-6 leading-tight">
          <span className="bg-gradient-to-r from-gray-600 to-black dark:from-gray-400 dark:to-white bg-clip-text text-transparent">
            {status}
          </span>
        </h1>

        <h2 className="text-3xl md:text-4xl font-light text-foreground mb-6 tracking-wide">
          {title}
        </h2>

        <p className="text-xl text-muted-foreground mb-4 leading-relaxed max-w-lg mx-auto">
          {message}
        </p>

        {/* Action Buttons */}
        {children ? (
          children
        ) : (
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            {!showBackButton ? (
              <Link
                href="/"
                className="inline-flex items-center gap-3 px-8 py-4 bg-foreground text-background rounded-full hover:bg-foreground/80 transition-all duration-300 transform hover:scale-105 shadow-xl hover:shadow-2xl font-medium text-lg group"
              >
                <ArrowLeftIcon className="w-5 h-5 transition-transform group-hover:-translate-x-1" />
                Go Home
              </Link>
            ) : (
              <Button
                className="inline-flex items-center gap-3 px-8 py-4 bg-foreground text-background rounded-full hover:bg-foreground/80 transition-all duration-300 transform hover:scale-105 shadow-xl hover:shadow-2xl font-medium text-lg group"
                onClick={() => window.history.back()}
              >
                Go Back
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default ErrorDisplay;
