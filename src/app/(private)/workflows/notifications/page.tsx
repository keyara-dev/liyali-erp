"use client";

import { useState } from "react";
import { useUserNotifications, useMarkAllNotificationsAsRead, useDeleteNotification } from "@/hooks/use-notifications";
import { getCurrentUser } from "@/lib/auth";
import { NotificationItem } from "@/components/notifications/notification-item";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckIcon, SearchIcon, MailIcon } from "lucide-react";
import type { NotificationType } from "@/types";

const NotificationSkeleton = () => (
  <div className="border rounded-lg p-4">
    <div className="flex gap-4">
      <Skeleton className="h-12 w-12 rounded" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-3/4" />
      </div>
    </div>
  </div>
);

interface NotificationsPageProps {
  userId: string;
}

function NotificationsPageContent({ userId }: NotificationsPageProps) {
  const [page, setPage] = useState(1);
  const [typeFilter, setTypeFilter] = useState<NotificationType | "all">("all");
  const [statusFilter, setStatusFilter] = useState<"all" | "read" | "unread">("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  const { data: result, isLoading } = useUserNotifications(
    userId,
    page,
    10,
    {
      type: typeFilter === "all" ? undefined : typeFilter,
      isRead: statusFilter === "all" ? undefined : statusFilter === "read",
    }
  );

  const markAllAsReadMutation = useMarkAllNotificationsAsRead();
  const deleteNotificationMutation = useDeleteNotification();

  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsReadMutation.mutateAsync({ userId });
    } catch (error) {
      console.error("Failed to mark all as read:", error);
    }
  };

  const handleDelete = async (notificationId: string) => {
    try {
      await deleteNotificationMutation.mutateAsync({ notificationId });
    } catch (error) {
      console.error("Failed to delete notification:", error);
    }
  };

  const filteredNotifications = result?.notifications.filter((n) => {
    if (searchQuery && !n.message.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }
    return true;
  }) || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Notifications</h1>
        <p className="text-muted-foreground">
          View all your notifications and approval updates
        </p>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Type</label>
              <Select
                value={typeFilter}
                onValueChange={(value) => {
                  setTypeFilter(value as NotificationType | "all");
                  setPage(1);
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All types</SelectItem>
                  {Object.entries(notificationTypeLabels).map(([key, label]) => (
                    <SelectItem key={key} value={key}>
                      {label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">Status</label>
              <Select
                value={statusFilter}
                onValueChange={(value) => {
                  setStatusFilter(value as "all" | "read" | "unread");
                  setPage(1);
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All status</SelectItem>
                  <SelectItem value="unread">Unread</SelectItem>
                  <SelectItem value="read">Read</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2 md:col-span-2">
              <label className="text-sm font-medium">Search</label>
              <div className="relative">
                <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search messages..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
          </div>

          <div className="flex gap-2 pt-4">
            <Button
              variant="outline"
              size="sm"
              onClick={handleMarkAllAsRead}
              disabled={markAllAsReadMutation.isPending}
            >
              <CheckIcon className="mr-2 h-4 w-4" />
              Mark all as read
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Notifications List */}
      <div className="space-y-3">
        {isLoading ? (
          <>
            <NotificationSkeleton />
            <NotificationSkeleton />
            <NotificationSkeleton />
          </>
        ) : filteredNotifications.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <MailIcon className="h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="font-semibold mb-2">No notifications</h3>
              <p className="text-muted-foreground text-sm">
                You're all caught up!
              </p>
            </CardContent>
          </Card>
        ) : (
          filteredNotifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              variant="full"
              onDelete={handleDelete}
              isDeleting={deleteNotificationMutation.isPending}
              showCheckbox={true}
              isSelected={selectedIds.includes(notification.id)}
              onSelectionChange={(checked) => {
                if (checked) {
                  setSelectedIds([...selectedIds, notification.id]);
                } else {
                  setSelectedIds(selectedIds.filter((id) => id !== notification.id));
                }
              }}
            />
          ))
        )}
      </div>

      {/* Pagination */}
      {result && result.hasMore && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            onClick={() => setPage(page + 1)}
          >
            Load more
          </Button>
        </div>
      )}
    </div>
  );
}

// Server component wrapper to get user
async function NotificationsPage() {
  const { getCurrentUser } = await import("@/lib/auth");
  const user = await getCurrentUser();

  if (!user) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Please log in to view notifications</p>
      </div>
    );
  }

  return <NotificationsPageContent userId={user.id} />;
}

export default NotificationsPage;
