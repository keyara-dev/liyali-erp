"use client";
import { ArrowLeft } from "lucide-react";
import { Button } from "../ui/button";
import { StatusBadge } from "../status-badge";
import { useRouter } from "next/navigation";

interface TStatusBadge {
  status: string;
  type:
    | "document"
    | "action"
    | "execution"
    | "approval"
    | "compliance"
    | "role"
    | "health";
}

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  description?: string; // Add description field for compatibility
  badges?: TStatusBadge[];
  onBackClick?: () => void;
  showBackButton?: boolean;
}

export function PageHeader({
  title,
  subtitle,
  badges,
  onBackClick,
  showBackButton = false,
}: PageHeaderProps) {
  const router = useRouter();
  return (
    <div className="mb-4 transition-colors duration-300 w-full">
      <div className="flex gap-4 items-center w-full">
        {showBackButton && (
          <Button
            onClick={() => {
              if (onBackClick !== undefined) {
                onBackClick?.();
              } else {
                router.back();
              }
            }}
            variant={"outline"}
            className="flex items-center h-10 w-10 aspect-square gap-2 text-foreground/70 transition-colors group"
          >
            <ArrowLeft className="w-6 h-6 group-hover:-translate-x-1 transition-transform" />
          </Button>
        )}

        <div className="space-y-1">
          <div className="flex flex-wrap items-end gap-2 md:gap-3">
            <h1
              className="text-2xl md:text-3xl font-bold text-foreground
              dark:text-foreground/90 tracking-tight"
            >
              {title}
            </h1>

            {badges && badges.length > 0 && (
              <div className="flex flex-wrap gap-2 pb-0.5">
                {badges.map((badge, index) => (
                  <StatusBadge
                    key={index}
                    status={badge.status}
                    type={badge.type}
                  />
                ))}
              </div>
            )}
          </div>

          {subtitle && (
            <p className="text-slate-600 dark:text-slate-400 font-medium text-sm leading-relaxed">
              {subtitle}
            </p>
          )}
        </div>
      </div>
      <div className="mt-6 h-0.5 bg-linear-to-r from-slate-200 via-slate-300 to-transparent  dark:to-transparent rounded-full"></div>
    </div>
  );
}
