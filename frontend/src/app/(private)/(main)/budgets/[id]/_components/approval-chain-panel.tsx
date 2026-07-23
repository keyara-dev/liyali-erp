"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ClipboardList, Plus } from "lucide-react";
import type { ApprovalRecord } from "@/types";
import Link from "next/link";
import {
  ApprovalChainStepper,
  type ApprovalStage,
  type StageStatus,
} from "@/components/workflows/approval-chain-stepper";

interface ApprovalChainPanelProps {
  approvalChain: ApprovalRecord[];
}

function recordsToStages(records: ApprovalRecord[]): ApprovalStage[] {
  return records.map((r, i) => {
    const upper = (r.status ?? "").toUpperCase();
    let status: StageStatus = "pending";
    if (upper === "APPROVED") status = "approved";
    else if (upper === "REJECTED") status = "rejected";
    else if (upper === "PENDING") status = "current";
    const at = r.approvedAt ?? r.actionTakenAt;
    return {
      id: r.approverId ?? `stage-${i}`,
      name: r.stageName ?? `Stage ${i + 1}`,
      status,
      actor: r.approverName,
      at: at instanceof Date ? at.toISOString() : at,
      comments: r.comments ?? r.remarks,
    };
  });
}

export function ApprovalChainPanel({ approvalChain }: ApprovalChainPanelProps) {
  if (!approvalChain || approvalChain.length === 0) {
    return (
      <Card className="border-2 border-dashed">
        <CardContent className="flex flex-col items-center justify-center px-8 py-8">
          <div className="relative mb-4">
            <div className="bg-primary/10 absolute inset-0 rounded-full blur-2xl" />
            <div className="bg-card border-primary/20 relative rounded-2xl border-2 p-6">
              <ClipboardList
                className="text-primary h-16 w-16"
                strokeWidth={1.5}
              />
            </div>
          </div>
          <h3 className="text-base font-semibold mb-1">No approval chain yet</h3>
          <p className="text-sm text-muted-foreground text-center max-w-sm mb-4">
            This budget has not entered an approval workflow. Submit it to
            start the chain.
          </p>
          <Button asChild variant="outline" size="sm">
            <Link href="/budgets">
              <Plus className="h-4 w-4 mr-2" />
              Back to budgets
            </Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  const stages = recordsToStages(approvalChain);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval Chain</CardTitle>
      </CardHeader>
      <CardContent>
        <ApprovalChainStepper stages={stages} />
      </CardContent>
    </Card>
  );
}
