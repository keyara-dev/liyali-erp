import { UserManagementClient } from "./_components/user-management-client";

export const metadata = {
  title: "User Management",
  description: "Manage user roles and access permissions",
};

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default function UserManagementPage() {
  // Use default values for client rendering
  return <UserManagementClient userId="system" userRole="ADMIN" />;
}
