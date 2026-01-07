import { PermissionsDebug } from "@/components/debug/permissions-debug";
import { PageHeader } from "@/components/base/page-header";
import { verifySession } from "@/lib/auth";
import { redirect } from "next/navigation";

export const metadata = {
  title: "Permissions Debug",
  description: "Debug user permissions and role system",
};

export default async function PermissionsDebugPage() {
  // Get authenticated user context
  const { session, isAuthenticated } = await verifySession();

  if (!isAuthenticated || !session?.user) {
    redirect("/login");
  }

  // Verify admin role (only admins can access debug pages)
  if (session.user.role !== "admin") {
    redirect("/access-denied");
  }

  return (
    <div>
      {/* Header */}
      <div className="bg-card border-b">
        <div className="container mx-auto px-4 py-6">
          <PageHeader
            title="Permissions Debug"
            subtitle="Debug and test the dynamic permission system"
          />
        </div>
      </div>

      <div className="container mx-auto p-4">
        <div className="flex justify-center">
          <PermissionsDebug />
        </div>
      </div>
    </div>
  );
}