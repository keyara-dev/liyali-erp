import { PageHeader } from "@/components/base/page-header";
import { InvitationsClient } from "./_components/invitations-client";

export const metadata = {
  title: "Invitations",
  description: "Manage pending and past organization invitations",
};

export const dynamic = "force-dynamic";

export default function InvitationsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Invitations"
        description="Manage invitations sent to existing platform users to join your organization."
      />
      <InvitationsClient />
    </div>
  );
}
