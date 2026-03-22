"use client";

import { useState } from "react";
import { formatDistanceToNow, isPast } from "date-fns";
import {
  Mail,
  MoreHorizontal,
  RefreshCw,
  XCircle,
  Clock,
  CheckCircle2,
  MinusCircle,
  Ban,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  useOrgInvitations,
  useCancelInvitation,
  useResendInvitation,
} from "@/hooks/use-invitation-mutations";
import { type OrganizationInvitation } from "@/app/_actions/invitation-actions";

const STATUS_CONFIG = {
  PENDING: {
    label: "Pending",
    icon: Clock,
    variant: "secondary" as const,
    className: "bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300",
  },
  ACCEPTED: {
    label: "Accepted",
    icon: CheckCircle2,
    variant: "secondary" as const,
    className: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300",
  },
  DECLINED: {
    label: "Declined",
    icon: MinusCircle,
    variant: "secondary" as const,
    className: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300",
  },
  EXPIRED: {
    label: "Expired",
    icon: Clock,
    variant: "secondary" as const,
    className: "bg-muted text-muted-foreground",
  },
  CANCELLED: {
    label: "Cancelled",
    icon: Ban,
    variant: "secondary" as const,
    className: "bg-muted text-muted-foreground",
  },
} as const;

function StatusBadge({ status }: { status: OrganizationInvitation["status"] }) {
  const config = STATUS_CONFIG[status.toUpperCase() as keyof typeof STATUS_CONFIG] ?? STATUS_CONFIG.PENDING;
  const Icon = config.icon;
  return (
    <Badge variant="secondary" className={`gap-1 ${config.className}`}>
      <Icon className="h-3 w-3" />
      {config.label}
    </Badge>
  );
}

function InvitationRow({ inv }: { inv: OrganizationInvitation }) {
  const cancelMutation = useCancelInvitation();
  const resendMutation = useResendInvitation();
  const s = inv.status?.toUpperCase();
  const isExpired = s === "PENDING" && isPast(new Date(inv.expiresAt));
  const effectiveStatus = isExpired ? "EXPIRED" : inv.status;
  const canCancel = s === "PENDING" && !isExpired;
  const canResend = s === "DECLINED" || s === "EXPIRED" || isExpired;

  return (
    <TableRow>
      <TableCell>
        <div className="flex flex-col">
          <span className="font-medium text-sm">{inv.invitedEmail}</span>
          {inv.invitedUser && (
            <span className="text-xs text-muted-foreground">{inv.invitedUser.name}</span>
          )}
        </div>
      </TableCell>
      <TableCell>
        <Badge variant="outline" className="capitalize text-xs">
          {inv.role}
        </Badge>
      </TableCell>
      <TableCell>
        <StatusBadge status={effectiveStatus as OrganizationInvitation["status"]} />
      </TableCell>
      <TableCell className="text-sm text-muted-foreground">
        {inv.invitedByUser?.name ?? "—"}
      </TableCell>
      <TableCell className="text-sm text-muted-foreground">
        {isExpired ? (
          <span className="text-destructive text-xs">Expired</span>
        ) : s === "PENDING" ? (
          <span className="text-xs">
            Expires {formatDistanceToNow(new Date(inv.expiresAt), { addSuffix: true })}
          </span>
        ) : (
          <span className="text-xs">
            {new Date(inv.createdAt).toLocaleDateString()}
          </span>
        )}
      </TableCell>
      <TableCell className="text-right">
        {(canCancel || canResend) && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-7 w-7">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {canResend && (
                <DropdownMenuItem
                  onClick={() => resendMutation.mutate(inv.id)}
                  disabled={resendMutation.isPending}
                >
                  <RefreshCw className="mr-2 h-4 w-4" />
                  Resend Invitation
                </DropdownMenuItem>
              )}
              {canCancel && (
                <DropdownMenuItem
                  onClick={() => cancelMutation.mutate(inv.id)}
                  disabled={cancelMutation.isPending}
                  className="text-destructive focus:text-destructive"
                >
                  <XCircle className="mr-2 h-4 w-4" />
                  Cancel Invitation
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </TableCell>
    </TableRow>
  );
}

export function InvitationsClient() {
  const { data: invitations = [], isLoading, error } = useOrgInvitations();
  const [filter, setFilter] = useState<"all" | OrganizationInvitation["status"]>("all");

  const filtered =
    filter === "all" ? invitations : invitations.filter((i) => i.status === filter);

  if (isLoading) {
    return (
      <div className="flex h-40 items-center justify-center text-muted-foreground text-sm">
        Loading invitations…
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-40 items-center justify-center text-destructive text-sm">
        Failed to load invitations.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Filter tabs */}
      <div className="flex flex-wrap gap-2">
        {(["all", "PENDING", "ACCEPTED", "DECLINED", "EXPIRED", "CANCELLED"] as const).map(
          (s) => (
            <Button
              key={s}
              size="sm"
              variant={filter === s ? "default" : "outline"}
              onClick={() => setFilter(s)}
              className="h-7 capitalize text-xs"
            >
              {s === "all" ? `All (${invitations.length})` : STATUS_CONFIG[s].label}
            </Button>
          )
        )}
      </div>

      {filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
          <Mail className="mb-3 h-8 w-8 text-muted-foreground/50" />
          <p className="text-sm text-muted-foreground">No invitations found.</p>
        </div>
      ) : (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Invitee</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Sent By</TableHead>
                <TableHead>Expiry</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((inv) => (
                <InvitationRow key={inv.id} inv={inv} />
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
