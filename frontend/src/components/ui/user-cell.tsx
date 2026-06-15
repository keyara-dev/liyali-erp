"use client";

import { UserAvatar } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";
import { formatRoleForDisplay } from "@/lib/workflow-utils";
import type { UserRef } from "@/types/core";

interface UserCellProps {
  /** Resolved user object from the API ({id,name,email,role}). */
  user?: UserRef | null;
  /** Display name to use when the resolved user object isn't present. */
  fallbackName?: string | null;
  /** Render the avatar (default true). */
  showAvatar?: boolean;
  className?: string;
}

/**
 * Renders a user reference as name + role (with an avatar), never a raw user ID.
 * Prefers the resolved `user` object; falls back to `fallbackName`. Shows "—"
 * when no name is available — so a bare UUID is never displayed in a table.
 */
export function UserCell({
  user,
  fallbackName,
  showAvatar = true,
  className,
}: UserCellProps) {
  const name = user?.name?.trim() || fallbackName?.trim() || "";

  if (!name) {
    return <span className="text-muted-foreground">—</span>;
  }

  const role = user?.role ? formatRoleForDisplay(user.role) : "";

  return (
    <div className={cn("flex min-w-0 items-center gap-2", className)}>
      {showAvatar && <UserAvatar name={name} size="xs" />}
      <div className="min-w-0 leading-tight">
        <div className="truncate text-sm font-medium text-foreground">
          {name}
        </div>
        {role && (
          <div className="truncate text-xs text-muted-foreground">{role}</div>
        )}
      </div>
    </div>
  );
}
