"use client";

import { Workflow, WorkflowStage, ApprovalTask } from "@/types";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  CheckCircle2,
  Clock,
  AlertCircle,
  ChevronRight,
  User,
  Users,
} from "lucide-react";

export interface ApprovalFlowDisplayProps {
  workflow: Workflow;
  currentStageIndex: number;
  approvals: ApprovalTask[];
  isCompleted?: boolean;
}

export function ApprovalFlowDisplay({
  workflow,
  currentStageIndex,
  approvals,
  isCompleted = false,
}: ApprovalFlowDisplayProps) {
  const stages = workflow.stages || [];

  const getStageStatus = (stageIndex: number) => {
    if (isCompleted) return "completed";
    if (stageIndex < currentStageIndex) return "completed";
    if (stageIndex === currentStageIndex) return "current";
    return "pending";
  };

  const getStageApprovals = (stageIndex: number) => {
    return approvals.filter((a) => a.stageIndex === stageIndex);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "completed":
        return <CheckCircle2 className="h-6 w-6 text-green-600" />;
      case "current":
        return <Clock className="h-6 w-6 text-blue-600 animate-pulse" />;
      default:
        return <div className="h-6 w-6 rounded-full border-2 border-gray-300" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "completed":
        return "bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800";
      case "current":
        return "bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800";
      default:
        return "bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-800";
    }
  };

  if (stages.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <AlertCircle className="h-8 w-8 text-yellow-600 mr-3" />
          <div>
            <h3 className="font-semibold">No Stages Configured</h3>
            <p className="text-sm text-muted-foreground">
              This workflow has no approval stages
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval Workflow Progress</CardTitle>
        <CardDescription>
          {isCompleted
            ? "Workflow completed successfully"
            : `Currently at stage ${currentStageIndex + 1} of ${stages.length}`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* Timeline View */}
          <div className="relative">
            {stages.map((stage, index) => {
              const status = getStageStatus(index);
              const stageApprovals = getStageApprovals(index);

              return (
                <div key={index}>
                  {/* Stage Card */}
                  <div className={`border rounded-lg p-4 ${getStatusColor(status)}`}>
                    <div className="flex items-start gap-4">
                      {/* Status Icon */}
                      <div className="flex-shrink-0">
                        {getStatusIcon(status)}
                      </div>

                      {/* Stage Info */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-2">
                          <h3 className="font-semibold">
                            Stage {index + 1}: {stage.name}
                          </h3>
                          <Badge
                            variant={
                              status === "completed"
                                ? "secondary"
                                : status === "current"
                                ? "default"
                                : "outline"
                            }
                          >
                            {status === "completed"
                              ? "Completed"
                              : status === "current"
                              ? "Current"
                              : "Pending"}
                          </Badge>
                        </div>

                        {stage.description && (
                          <p className="text-sm text-muted-foreground mb-3">
                            {stage.description}
                          </p>
                        )}

                        {/* Approvers */}
                        {stage.approvers && stage.approvers.length > 0 && (
                          <div className="mb-3">
                            <h4 className="text-xs font-semibold text-muted-foreground mb-2 uppercase">
                              Approvers
                            </h4>
                            <div className="flex flex-wrap gap-2">
                              {stage.approvers.map((approver) => (
                                <div
                                  key={approver.id}
                                  className="flex items-center gap-1 bg-background px-2 py-1 rounded border text-xs"
                                >
                                  <Avatar className="h-4 w-4">
                                    <AvatarImage src={approver.avatar} />
                                    <AvatarFallback>
                                      {approver.name?.charAt(0).toUpperCase()}
                                    </AvatarFallback>
                                  </Avatar>
                                  <span>{approver.name || approver.email}</span>
                                </div>
                              ))}
                            </div>
                          </div>
                        )}

                        {/* Stage Approvals History */}
                        {stageApprovals.length > 0 && (
                          <div className="space-y-2 pt-3 border-t">
                            <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                              Approvals
                            </h4>
                            {stageApprovals.map((approval) => (
                              <div
                                key={approval.id}
                                className="flex items-center justify-between bg-background px-2 py-2 rounded text-xs"
                              >
                                <div className="flex items-center gap-2">
                                  <User className="h-3 w-3 text-muted-foreground" />
                                  <span className="font-medium">
                                    {approval.approverName || approval.approverId}
                                  </span>
                                </div>
                                <div className="flex items-center gap-2">
                                  {approval.status === "approved" && (
                                    <CheckCircle2 className="h-3 w-3 text-green-600" />
                                  )}
                                  {approval.status === "rejected" && (
                                    <AlertCircle className="h-3 w-3 text-red-600" />
                                  )}
                                  <span className="text-muted-foreground">
                                    {new Date(approval.actionDate || new Date()).toLocaleDateString()}
                                  </span>
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Connector Line */}
                  {index < stages.length - 1 && (
                    <div className="flex justify-center py-2">
                      <ChevronRight className="h-6 w-6 text-muted-foreground transform rotate-90" />
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          {/* Summary */}
          <div className="pt-4 border-t">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Total Stages
                </h4>
                <p className="text-lg font-bold">{stages.length}</p>
              </div>
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Completed
                </h4>
                <p className="text-lg font-bold">{currentStageIndex}</p>
              </div>
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Remaining
                </h4>
                <p className="text-lg font-bold">
                  {stages.length - currentStageIndex - (isCompleted ? 1 : 0)}
                </p>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
