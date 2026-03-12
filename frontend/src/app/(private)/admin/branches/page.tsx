import { requireAdminRole } from "@/lib/admin-guard";
import BranchesClient from "../_components/branches-client";

export const dynamic = "force-dynamic";

export default async function BranchesPage() {
  await requireAdminRole();

  return (
    <div className="container mx-auto p-6 px-4">
      <BranchesClient />
    </div>
  );
}
