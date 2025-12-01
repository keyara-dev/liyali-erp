import { Suspense } from "react";
import { NotificationBell } from "@/components/notifications/notification-bell";
import { Button } from "@/components/ui/button";
import { BellIcon } from "lucide-react";
import { getCurrentUser } from "@/lib/auth";

// Fallback while loading user
function NotificationFallback() {
  return (
    <Button size="icon" variant="ghost" disabled>
      <BellIcon className="h-5 w-5" />
    </Button>
  );
}

// Server component that fetches user data
async function NotificationsContent() {
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

// Server component - no "use client" directive
async function Notifications() {
  return (
    <Suspense fallback={<NotificationFallback />}>
      <NotificationsContent />
    </Suspense>
  );
}

export default Notifications;
