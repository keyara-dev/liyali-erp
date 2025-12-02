import { getCurrentUser } from "@/auth";
import { redirect } from "next/navigation";
import { PropsWithChildren } from "react";

export default async function AdminLayout({ children }: PropsWithChildren) {
  const user = await getCurrentUser();

  if (!user) {
    redirect("/login");
  }

  // Check if user is admin
  if (user.role !== "ADMIN") {
    redirect("/home");
  }

  return <>{children}</>;
}
