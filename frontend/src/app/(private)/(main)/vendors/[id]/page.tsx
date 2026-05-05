import { redirect } from "next/navigation";
import { verifySession } from "@/lib/auth";
import { VendorDetailClient } from "./_components/vendor-detail-client";

export const metadata = {
  title: "Vendor Detail",
  description: "View vendor profile, banking details, and recent purchase orders.",
};

interface VendorDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function VendorDetailPage({
  params,
}: VendorDetailPageProps) {
  const { session } = await verifySession();
  if (!session?.user) {
    redirect("/login");
  }

  const { id } = await params;
  return <VendorDetailClient vendorId={id} />;
}
