"use client";

import { Suspense } from "react";
import { NotificationBell } from "@/components/notifications/notification-bell";
import { Button } from "@/components/ui/button";
import { BellIcon } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

// Fallback while loading user
function NotificationFallback() {
  return (
    <Button size="icon" variant="ghost" disabled>
      <BellIcon className="h-5 w-5" />
    </Button>
  );
}

// Client component that will be wrapped by server component
async function NotificationsContent() {
  // Import here to avoid circular dependencies
  const { getCurrentUser } = await import("@/lib/auth");
  const user = await getCurrentUser();

  if (!user) {
    return (
      <Button size="icon" variant="ghost" disabled>
        <BellIcon className="h-5 w-5" />
      </Button>
    );
  }

  return <NotificationBell userId={user.id} />;
}

const Notifications = () => {
  
  return (
    <Suspense fallback={<NotificationFallback />}>
      <NotificationsContent />
    </Suspense>
  );
};

export default Notifications;
