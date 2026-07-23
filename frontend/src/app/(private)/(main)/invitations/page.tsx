"use client";

import { formatDistanceToNow } from "date-fns";
import { Mail, Building2, CheckCircle2, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  useMyPendingInvitations,
  useAcceptInvitation,
  useDeclineInvitation,
} from "@/hooks/use-invitation-mutations";
import { type OrganizationInvitation } from "@/app/_actions/invitation-actions";

function InvitationCard({ inv }: { inv: OrganizationInvitation }) {
  const accept = useAcceptInvitation();
  const decline = useDeclineInvitation();

  // The backend returns the token embedded in the invitation.
  // Since token is hidden in list responses, we use the invitation ID as the lookup key
  // and pass the invitation id — the backend accept/decline routes accept the token.
  // For the invitee page we use a dedicated token field if available, fall back to id.
  const token = (inv as any).token ?? inv.id;

  const expiresIn = formatDistanceToNow(new Date(inv.expiresAt), { addSuffix: true });
  const orgName = inv.organization?.name ?? "an organization";
  const inviterName = inv.invitedByUser?.name ?? "An admin";

  return (
    <Card className="border">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/10">
              <Building2 className="h-5 w-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-base">{orgName}</CardTitle>
              <CardDescription className="text-xs mt-0.5">
                Invited by {inviterName}
              </CardDescription>
            </div>
          </div>
          <Badge variant="outline" className="capitalize shrink-0 text-xs">
            {inv.role}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="pb-3">
        <p className="text-sm text-muted-foreground">
          You have been invited to join <strong>{orgName}</strong> as{" "}
          <strong>{inv.role}</strong>.
        </p>
        <p className="mt-1 text-xs text-muted-foreground">
          Expires {expiresIn}
        </p>
      </CardContent>

      <CardFooter className="gap-3 pt-0">
        <Button
          size="sm"
          onClick={() => accept.mutate(token)}
          disabled={accept.isPending || decline.isPending}
          isLoading={accept.isPending}
          loadingText="Accepting…"
          className="gap-1.5"
        >
          <CheckCircle2 className="h-4 w-4" />
          Accept
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={() => decline.mutate(token)}
          disabled={accept.isPending || decline.isPending}
          isLoading={decline.isPending}
          loadingText="Declining…"
          className="gap-1.5 text-destructive hover:text-destructive"
        >
          <XCircle className="h-4 w-4" />
          Decline
        </Button>
      </CardFooter>
    </Card>
  );
}

export default function InvitationsPage() {
  const { data: invitations = [], isLoading } = useMyPendingInvitations();

  return (
    <div className="mx-auto max-w-2xl space-y-6 py-6">
      <div>
        <h1 className="text-2xl font-semibold">Pending Invitations</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Organizations that have invited you to join.
        </p>
      </div>

      {isLoading ? (
        <div className="flex h-40 items-center justify-center text-muted-foreground text-sm">
          Loading invitations…
        </div>
      ) : invitations.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <Mail className="mb-3 h-10 w-10 text-muted-foreground/40" />
          <p className="text-sm font-medium">No pending invitations</p>
          <p className="mt-1 text-xs text-muted-foreground">
            When an admin invites you to join their organization, it will appear here.
          </p>
        </div>
      ) : (
        <div className="grid gap-4">
          {invitations.map((inv) => (
            <InvitationCard key={inv.id} inv={inv} />
          ))}
        </div>
      )}
    </div>
  );
}
