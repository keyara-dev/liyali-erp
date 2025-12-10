import { redirect } from "next/navigation";
import { PVApprovalClient } from "./_components/pv-approval-client";
import { verifySession } from "@/lib/auth";

export const metadata = {
  title: "Payment Voucher Approval",
  description: "Review and approve payment voucher",
};

interface PVApprovalPageProps {
  params: {
    id: string;
  };
}

export default async function PVApprovalPage({ params }: PVApprovalPageProps) {
  const { session } = await verifySession();

  if (!session?.user) {
    redirect("/login");
  }

  return (
    <PVApprovalClient
      pvId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  );
}
