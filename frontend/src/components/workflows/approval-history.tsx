"use client";

import { useMemo } from "react";
import { useApprovalHistory } from "@/hooks/use-approval-workflow";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, History } from "lucide-react";
import {
  ApprovalChainStepper,
  type ApprovalStage,
  type StageStatus,
} from "@/components/workflows/approval-chain-stepper";

export interface ApprovalHistoryProps {
  documentId?: string;
  entityId?: string; // Legacy compatibility
  entityType?: string; // Legacy compatibility — currently unused
}

export function ApprovalHistory({
  documentId,
  entityId,
  entityType: _entityType,
}: ApprovalHistoryProps) {
  const actualDocumentId = documentId || entityId || "";
  const { data: historyData, isLoading } = useApprovalHistory(actualDocumentId);

  const stages: ApprovalStage[] = useMemo(() => {
    if (!historyData) return [];
    const sorted = [...historyData].sort(
      (a: any, b: any) =>
        new Date(b.approvedAt ?? 0).getTime() -
        new Date(a.approvedAt ?? 0).getTime()
    );
    return sorted.map((r: any, i: number) => {
      const upper = (r.status ?? "").toUpperCase();
      let status: StageStatus = "pending";
      if (upper === "APPROVED") status = "approved";
      else if (upper === "REJECTED") status = "rejected";
      else if (upper === "RETURNED" || upper === "REASSIGNED") status = "skipped";
      const at =
        r.approvedAt ?? r.actionTakenAt;
      return {
        id: r.id ?? r.approverId ?? `entry-${i}`,
        name: r.stageName ?? r.action ?? `Entry ${i + 1}`,
        status,
        actor: r.approverName,
        at: at instanceof Date ? at.toISOString() : at,
        comments: r.comments ?? r.remarks,
      };
    });
  }, [historyData]);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Approval History</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full rounded-md" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!actualDocumentId) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Document id is required to load approval history.
        </AlertDescription>
      </Alert>
    );
  }

  if (stages.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <History className="h-4 w-4" />
            Approval History
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No approval actions recorded yet.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <History className="h-4 w-4" />
          Approval History
        </CardTitle>
        <CardDescription>
          {stages.length} action{stages.length === 1 ? "" : "s"} recorded
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ApprovalChainStepper stages={stages} />
      </CardContent>
    </Card>
  );
}
