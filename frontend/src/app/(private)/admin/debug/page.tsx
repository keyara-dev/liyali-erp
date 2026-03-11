import { PermissionsDebug } from "@/components/debug/permissions-debug";
import { PageHeader } from "@/components/base/page-header";
import { requireAdminRole } from "@/lib/admin-guard";

export const metadata = {
  title: "Permissions Debug",
  description: "Debug user permissions and role system",
};

export default async function PermissionsDebugPage() {
  await requireAdminRole();

  return (
    <div>
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
