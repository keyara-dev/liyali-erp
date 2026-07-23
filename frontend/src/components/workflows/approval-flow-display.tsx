"use client";

import {
  ApprovalChainStepper,
  type ApprovalStage,
  type StageStatus,
} from "@/components/workflows/approval-chain-stepper";
import type { ApprovalRecord } from "@/types";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { AlertCircle } from "lucide-react";

export interface ApprovalFlowDisplayProps {
  approvalHistory: ApprovalRecord[];
  currentStage: number;
  totalStages?: number;
  isCompleted?: boolean;
}

function recordsToStages(
  records: ApprovalRecord[],
  currentStage: number,
  totalStages: number,
  isCompleted: boolean
): ApprovalStage[] {
  const stages: ApprovalStage[] = records.map((r, i) => {
    const upper = (r.status ?? "").toUpperCase();
    let status: StageStatus = "pending";
    if (upper === "APPROVED") status = "approved";
    else if (upper === "REJECTED") status = "rejected";
    else if (i < currentStage) status = "approved";
    else if (i === currentStage && !isCompleted) status = "current";
    return {
      id: r.approverId ?? `stage-${i}`,
      name: r.stageName ?? `Stage ${i + 1}`,
      status,
      actor: r.approverName,
      at: (r.approvedAt ?? r.actionTakenAt)?.toISOString(),
      comments: r.comments ?? r.remarks,
    };
  });

  // Append pending placeholders for stages not yet in approvalHistory
  if (totalStages > stages.length) {
    const remaining = totalStages - stages.length;
    for (let i = 0; i < remaining; i++) {
      const idx = stages.length;
      stages.push({
        id: `pending-${idx}`,
        name: `Stage ${idx + 1}`,
        status: idx === currentStage && !isCompleted ? "current" : "pending",
      });
    }
  }
  return stages;
}

export function ApprovalFlowDisplay({
  approvalHistory,
  currentStage,
  totalStages = 0,
  isCompleted = false,
}: ApprovalFlowDisplayProps) {
  const effectiveTotal = totalStages || approvalHistory.length;

  if (effectiveTotal === 0 && approvalHistory.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <AlertCircle className="h-8 w-8 text-amber-600 mr-3" />
          <div>
            <h3 className="font-semibold">No Approval History</h3>
            <p className="text-sm text-muted-foreground">
              This document has not been submitted for approval yet.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const stages = recordsToStages(
    approvalHistory,
    currentStage,
    effectiveTotal,
    isCompleted
  );
  const completed = isCompleted ? effectiveTotal : currentStage;
  const remaining = Math.max(0, effectiveTotal - completed);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval Workflow Progress</CardTitle>
        <CardDescription>
          {isCompleted
            ? "Workflow completed successfully"
            : `Currently at stage ${Math.min(currentStage + 1, effectiveTotal)} of ${effectiveTotal}`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          <ApprovalChainStepper stages={stages} />

          <div className="pt-4 border-t">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Total Stages
                </h4>
                <p className="text-lg font-bold tabular-nums">{effectiveTotal}</p>
              </div>
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Completed
                </h4>
                <p className="text-lg font-bold tabular-nums">{completed}</p>
              </div>
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Remaining
                </h4>
                <p className="text-lg font-bold tabular-nums">{remaining}</p>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
