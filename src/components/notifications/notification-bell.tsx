"use client";

import { BellIcon, ClockIcon, CheckIcon } from "lucide-react";
import Link from "next/link";
import { useIsMobile } from "@/hooks/use-mobile";
import {
  useNotificationBell,
  useMarkNotificationAsRead,
  useNotificationPolling,
} from "@/hooks/use-notifications";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { formatDistanceToNow } from "date-fns";

const getNotificationIcon = (type: string) => {
  switch (type) {
    case "TASK_ASSIGNED":
      return "📋";
    case "TASK_REASSIGNED":
      return "🔄";
    case "TASK_APPROVED":
      return "✅";
    case "TASK_REJECTED":
      return "❌";
    case "WORKFLOW_COMPLETE":
      return "🎉";
    default:
      return "🔔";
  }
};

const NotificationSkeleton = () => (
  <div className="flex items-start gap-2 p-4 border-b">
    <Skeleton className="h-8 w-8 rounded-full" />
    <div className="flex-1">
      <Skeleton className="h-4 w-32 mb-2" />
      <Skeleton className="h-3 w-full" />
    </div>
  </div>
);

interface NotificationBellProps {
  userId: string;
}

export function NotificationBell({ userId }: NotificationBellProps) {
  const isMobile = useIsMobile();

  const { unreadCount, recentNotifications, isLoading } =
    useNotificationBell(userId);
  const markAsReadMutation = useMarkNotificationAsRead();

  // Set up automatic polling
  useNotificationPolling(userId, 30 * 1000); // Poll every 30 seconds

  const handleMarkAsRead = async (notificationId: string) => {
    try {
      await markAsReadMutation.mutateAsync({ notificationId });
    } catch (error) {
      console.error("Failed to mark as read:", error);
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="icon" variant="ghost" className="relative">
          <>
            <BellIcon className="h-5 w-5" />
            {unreadCount > 0 && (
              <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-destructive text-xs font-semibold text-background">
                {unreadCount > 9 ? "9+" : unreadCount}
              </span>
            )}
          </>
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align={isMobile ? "center" : "end"}
        className="ms-4 w-96 p-0"
      >
        <DropdownMenuLabel className="sticky top-0 z-10 border-b bg-background p-0 dark:bg-muted">
          <div className="flex items-center justify-between px-6 py-4">
            <div className="font-semibold">
              Notifications {unreadCount > 0 && `(${unreadCount})`}
            </div>
            <Button
              variant="link"
              className="h-auto p-0 text-xs"
              size="sm"
              asChild
            >
              <Link href="/workflows/notifications">View all</Link>
            </Button>
          </div>
        </DropdownMenuLabel>

        <ScrollArea className="h-[400px]">
          {isLoading ? (
            <>
              <NotificationSkeleton />
              <NotificationSkeleton />
              <NotificationSkeleton />
            </>
          ) : recentNotifications.length === 0 ? (
            <div className="flex h-32 items-center justify-center text-muted-foreground">
              <p className="text-sm">No notifications yet</p>
            </div>
          ) : (
            <div>
              {recentNotifications.map((notification) => (
                <div
                  key={notification.id}
                  className="group border-b px-4 py-3 hover:bg-muted/50 cursor-pointer transition-colors"
                  onClick={() => handleMarkAsRead(notification.id)}
                >
                  <div className="flex items-start gap-3">
                    <div className="flex-none pt-1 text-xl">
                      {getNotificationIcon(notification.type)}
                    </div>
                    <div className="flex flex-1 flex-col gap-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <div className="truncate text-sm font-semibold dark:group-hover:text-default-800">
                          {notification.title}
                        </div>
                        {!notification.isRead && (
                          <span className="mt-1 flex-none block size-2 rounded-full bg-destructive/80" />
                        )}
                      </div>
                      <div className="line-clamp-2 text-xs text-muted-foreground dark:group-hover:text-default-700">
                        {notification.message}
                      </div>
                      {notification.entityNumber && (
                        <div className="text-xs font-mono text-muted-foreground">
                          {notification.entityType} #{notification.entityNumber}
                        </div>
                      )}
                      <div className="flex items-center gap-1 text-xs text-muted-foreground dark:group-hover:text-default-500">
                        <ClockIcon className="h-3 w-3" />
                        {formatDistanceToNow(new Date(notification.createdAt), {
                          addSuffix: true,
                        })}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
