import { redirect } from "next/navigation";
import { verifyAdminSession } from "@/lib/auth";

// Force dynamic rendering for authentication
export const dynamic = "force-dynamic";

export default async function HomePage() {
  const { isAuthenticated } = await verifyAdminSession();

  if (isAuthenticated) {
    redirect("/admin/dashboard");
  } else {
    redirect("/login");
  }
}
