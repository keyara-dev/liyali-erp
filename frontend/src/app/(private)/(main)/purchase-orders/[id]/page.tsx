import { verifySession } from "@/lib/auth";
import { redirect } from "next/navigation";
import { PODetailClient } from "./_components/po-detail-client";

export const metadata = {
  title: "Purchase Order Details",
  description: "View and manage purchase order details",
};

interface PODetailPageProps {
  params: Promise<{
    id: string;
  }>;
}

export default async function PODetailPage({ params }: PODetailPageProps) {
  const { session } = await verifySession();

  if (!session?.user) {
    redirect("/login");
  }

  const POId = (await params).id;

  return (
    <PODetailClient
      poId={POId}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  );
}
