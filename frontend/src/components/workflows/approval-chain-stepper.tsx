import * as React from "react";
import { Check, Clock, X } from "lucide-react";
import { cn } from "@/lib/utils";

export type StageStatus = "approved" | "rejected" | "current" | "pending" | "skipped";

export interface ApprovalStage {
  id: string;
  name: string;
  status: StageStatus;
  /** Approver / rejecter display name, when available. */
  actor?: string;
  /** ISO timestamp of completion, when available. */
  at?: string;
  /** Optional comments/remarks attached to the action. */
  comments?: string;
}

export interface ApprovalChainStepperProps {
  stages: ApprovalStage[];
  className?: string;
}

const MARKER: Record<StageStatus, string> = {
  approved: "bg-emerald-600 text-white border-emerald-600",
  rejected: "bg-rose-600 text-white border-rose-600",
  current: "bg-background text-primary border-primary ring-2 ring-primary/30",
  pending: "bg-muted text-muted-foreground border-border",
  skipped: "bg-muted/40 text-muted-foreground border-dashed border-border",
};

const CONNECTOR: Record<StageStatus, string> = {
  approved: "bg-emerald-600",
  rejected: "bg-rose-600",
  current: "bg-border",
  pending: "bg-border",
  skipped: "bg-border/40",
};

function StageIcon({ status }: { status: StageStatus }) {
  if (status === "approved") return <Check className="h-3 w-3" aria-hidden="true" />;
  if (status === "rejected") return <X className="h-3 w-3" aria-hidden="true" />;
  if (status === "current") return <Clock className="h-3 w-3" aria-hidden="true" />;
  return null;
}

function fmtDate(iso?: string) {
  if (!iso) return null;
  try {
    return new Date(iso).toLocaleDateString([], {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return null;
  }
}

export function ApprovalChainStepper({ stages, className }: ApprovalChainStepperProps) {
  return (
    <ol className={cn("space-y-3", className)} aria-label="Approval chain">
      {stages.map((stage, idx) => {
        const isLast = idx === stages.length - 1;
        const dateStr = fmtDate(stage.at);
        return (
          <li
            key={stage.id}
            aria-current={stage.status === "current" ? "step" : undefined}
            className="relative flex gap-3"
          >
            <div className="flex flex-col items-center">
              <span
                data-testid={`stage-marker-${stage.id}`}
                data-status={stage.status}
                className={cn(
                  "flex h-6 w-6 items-center justify-center rounded-full border text-xs font-semibold shrink-0",
                  MARKER[stage.status]
                )}
              >
                <StageIcon status={stage.status} />
                {stage.status === "pending" && idx + 1}
                {stage.status === "skipped" && "·"}
              </span>
              {!isLast && (
                <span
                  aria-hidden="true"
                  className={cn("mt-1 w-0.5 flex-1 min-h-4", CONNECTOR[stage.status])}
                />
              )}
            </div>
            <div className="flex-1 pb-3">
              <div className="text-sm font-medium leading-tight">{stage.name}</div>
              {(stage.actor || dateStr) && (
                <div className="text-xs text-muted-foreground mt-0.5">
                  {stage.actor}
                  {stage.actor && dateStr && " · "}
                  {dateStr}
                </div>
              )}
              {stage.comments && (
                <p className="text-xs text-foreground/80 mt-1 italic">"{stage.comments}"</p>
              )}
            </div>
          </li>
        );
      })}
    </ol>
  );
}
