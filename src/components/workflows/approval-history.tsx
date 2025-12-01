"use client";

import { useMemo, useState } from "react";
import { useGetTaskHistory } from "@/hooks/use-workflows";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  CheckCircle2,
  XCircle,
  Repeat2,
  Clock,
  AlertCircle,
  ChevronDown,
  ChevronUp,
} from "lucide-react";

export interface ApprovalHistoryProps {
  entityId: string;
  entityType: string;
}

export function ApprovalHistory({
  entityId,
  entityType,
}: ApprovalHistoryProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const { data: historyData, isLoading } = useGetTaskHistory(entityId);

  const sortedHistory = useMemo(() => {
    if (!historyData?.history) return [];
    return [...historyData.history].sort(
      (a, b) =>
        new Date(b.actionDate || b.createdAt || 0).getTime() -
        new Date(a.actionDate || a.createdAt || 0).getTime()
    );
  }, [historyData?.history]);

  const getActionIcon = (action: string) => {
    switch (action) {
      case "approved":
        return <CheckCircle2 className="h-5 w-5 text-green-600" />;
      case "rejected":
        return <XCircle className="h-5 w-5 text-red-600" />;
      case "reassigned":
        return <Repeat2 className="h-5 w-5 text-blue-600" />;
      default:
        return <Clock className="h-5 w-5 text-gray-600" />;
    }
  };

  const getActionBadge = (action: string) => {
    switch (action) {
      case "approved":
        return (
          <Badge className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
            Approved
          </Badge>
        );
      case "rejected":
        return (
          <Badge className="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
            Rejected
          </Badge>
        );
      case "reassigned":
        return (
          <Badge className="bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
            Reassigned
          </Badge>
        );
      default:
        return <Badge variant="outline">{action}</Badge>;
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-40" />
          <Skeleton className="h-4 w-60 mt-2" />
        </CardHeader>
        <CardContent className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="space-y-3 pb-4 border-b last:border-b-0">
              <Skeleton className="h-4 w-48" />
              <Skeleton className="h-3 w-64" />
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  if (!sortedHistory || sortedHistory.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Clock className="h-12 w-12 text-muted-foreground mb-3 opacity-50" />
          <h3 className="font-semibold mb-1">No Actions Yet</h3>
          <p className="text-sm text-muted-foreground text-center">
            This {entityType.toLowerCase()} hasn't been approved or rejected yet
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval History</CardTitle>
        <CardDescription>
          Timeline of all approvals and actions for this {entityType.toLowerCase()}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {sortedHistory.map((entry, index) => {
            const isExpanded = expandedId === entry.id;
            const actionDate = new Date(entry.actionDate || entry.createdAt || 0);
            const dateStr = actionDate.toLocaleDateString();
            const timeStr = actionDate.toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            });

            return (
              <div
                key={entry.id}
                className="border rounded-lg hover:bg-muted/50 transition-colors"
              >
                <button
                  onClick={() =>
                    setExpandedId(isExpanded ? null : entry.id)
                  }
                  className="w-full p-4 text-left flex items-center justify-between gap-4"
                >
                  {/* Timeline Marker */}
                  <div className="flex-shrink-0">
                    {getActionIcon(entry.action || "pending")}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-semibold text-sm">
                        {entry.approverName || entry.approverId || "System"}
                      </h3>
                      {getActionBadge(entry.action || "pending")}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {dateStr} at {timeStr}
                    </p>
                  </div>

                  {/* Expand Icon */}
                  {(entry.remarks ||
                    entry.reason ||
                    entry.reassignedTo) && (
                    <div className="flex-shrink-0">
                      {isExpanded ? (
                        <ChevronUp className="h-4 w-4 text-muted-foreground" />
                      ) : (
                        <ChevronDown className="h-4 w-4 text-muted-foreground" />
                      )}
                    </div>
                  )}
                </button>

                {/* Expanded Content */}
                {isExpanded && (
                  <div className="px-4 pb-4 border-t space-y-3">
                    {/* Remarks */}
                    {entry.remarks && (
                      <div>
                        <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                          Remarks
                        </h4>
                        <p className="text-sm bg-muted rounded p-2">
                          {entry.remarks}
                        </p>
                      </div>
                    )}

                    {/* Rejection Reason */}
                    {entry.reason && (
                      <div>
                        <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                          Reason for Rejection
                        </h4>
                        <Alert variant="destructive" className="bg-red-50 border-red-200 dark:bg-red-900/20">
                          <AlertCircle className="h-4 w-4" />
                          <AlertDescription className="text-red-800 dark:text-red-200">
                            {entry.reason}
                          </AlertDescription>
                        </Alert>
                      </div>
                    )}

                    {/* Reassignment Info */}
                    {entry.reassignedTo && (
                      <div>
                        <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-2">
                          Reassigned To
                        </h4>
                        <div className="flex items-center gap-2 bg-blue-50 dark:bg-blue-900/20 p-2 rounded border border-blue-200 dark:border-blue-800">
                          <Avatar className="h-6 w-6">
                            <AvatarFallback>
                              {entry.reassignedTo.name?.charAt(0).toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium">
                              {entry.reassignedTo.name || entry.reassignedTo.email}
                            </p>
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Signature Note */}
                    {entry.signature && (
                      <div>
                        <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                          Digital Signature
                        </h4>
                        <p className="text-xs text-muted-foreground">
                          ✓ Digitally signed on {dateStr}
                        </p>
                      </div>
                    )}

                    {/* Reviewer Info */}
                    {entry.reviewedBy && (
                      <div>
                        <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                          Reviewed By
                        </h4>
                        <p className="text-sm">{entry.reviewedBy}</p>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* Summary */}
        <div className="mt-6 pt-6 border-t grid grid-cols-3 gap-4 text-center text-sm">
          <div>
            <h4 className="font-semibold text-muted-foreground mb-1">
              Total Actions
            </h4>
            <p className="text-lg font-bold">{sortedHistory.length}</p>
          </div>
          <div>
            <h4 className="font-semibold text-muted-foreground mb-1">
              Approvals
            </h4>
            <p className="text-lg font-bold text-green-600">
              {sortedHistory.filter((h) => h.action === "approved").length}
            </p>
          </div>
          <div>
            <h4 className="font-semibold text-muted-foreground mb-1">
              Rejections
            </h4>
            <p className="text-lg font-bold text-red-600">
              {sortedHistory.filter((h) => h.action === "rejected").length}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
