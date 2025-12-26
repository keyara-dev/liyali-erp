"use client";

import { useParams } from "next/navigation";
import { useApprovalTaskDetail } from "@/hooks/use-approval-workflow";
import {
  ApprovalFlowDisplay,
  ApprovalActionPanel,
  ApprovalHistory,
} from "@/components/";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  AlertCircle,
  CheckCircle2,
  ClockIcon,
  FileText,
  User,
  Calendar,
} from "lucide-react";

export default function RequisitionApprovalPage() {
  const params = useParams();
  const taskId = params.id as string;

  const { data: task, isLoading } = useApprovalTaskDetail(taskId);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-12 w-48" />
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (!task) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load approval task details. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <div className="flex items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              Requisition Approval
            </h1>
            <p className="text-muted-foreground">
              {requisition?.name || `Requisition #${task.entityNumber}`}
            </p>
          </div>
          <Badge
            variant={
              task.status === "pending"
                ? "default"
                : task.status === "approved"
                  ? "secondary"
                  : "destructive"
            }
          >
            {task.status === "pending"
              ? "Pending Approval"
              : task.status === "approved"
                ? "Approved"
                : "Rejected"}
          </Badge>
        </div>
      </div>

      {/* Info Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Entity
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <FileText className="h-4 w-4 text-muted-foreground" />
              <span className="font-mono font-semibold">
                {task.entityType} #{task.entityNumber}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Workflow
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-semibold text-sm">
              {workflow?.name || "Unknown"}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Current Stage
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-semibold text-sm">{task.stageName}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Assigned To
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-muted-foreground" />
              <span className="font-semibold text-sm">
                {task.approverName || "Unassigned"}
              </span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left Column - Forms and Actions */}
        <div className="lg:col-span-2 space-y-6">
          {/* Requisition Details */}
          {requisition && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="h-5 w-5" />
                  Requisition Details
                </CardTitle>
                <CardDescription>
                  Review the requisition information below
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                      Requisition ID
                    </h4>
                    <p className="font-mono">{requisition.id}</p>
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                      Amount
                    </h4>
                    <p className="font-semibold">
                      {requisition.amount
                        ? `K${requisition.amount.toLocaleString()}`
                        : "N/A"}
                    </p>
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                      Department
                    </h4>
                    <p>{requisition.department || "N/A"}</p>
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                      Created Date
                    </h4>
                    <p>
                      {requisition.createdAt
                        ? new Date(requisition.createdAt).toLocaleDateString()
                        : "N/A"}
                    </p>
                  </div>
                </div>

                {requisition.description && (
                  <div>
                    <h4 className="text-sm font-semibold text-muted-foreground mb-2">
                      Description
                    </h4>
                    <p className="text-sm bg-muted p-3 rounded">
                      {requisition.description}
                    </p>
                  </div>
                )}

                {requisition.items && requisition.items.length > 0 && (
                  <div>
                    <h4 className="text-sm font-semibold text-muted-foreground mb-2">
                      Items ({requisition.items.length})
                    </h4>
                    <div className="space-y-2">
                      {requisition.items.map((item: any, index: number) => (
                        <div
                          key={index}
                          className="flex justify-between items-center p-2 bg-muted rounded text-sm"
                        >
                          <span>{item.description || item.name}</span>
                          <span className="font-mono">
                            Qty: {item.quantity} @ K{item.unitPrice}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Approval Actions */}
          {task.status === "pending" && (
            <ApprovalActionPanel
              task={task}
              onApprovalComplete={() => {
                // Refresh page or update state
                window.location.reload();
              }}
            />
          )}

          {/* Completed Approval Alert */}
          {task.status !== "pending" && (
            <Alert
              className={
                task.status === "approved"
                  ? "bg-green-50 border-green-200 dark:bg-green-900/20"
                  : "bg-red-50 border-red-200 dark:bg-red-900/20"
              }
            >
              <CheckCircle2
                className={`h-4 w-4 ${task.status === "approved" ? "text-green-600" : "text-red-600"}`}
              />
              <AlertDescription
                className={
                  task.status === "approved"
                    ? "text-green-700 dark:text-green-200"
                    : "text-red-700 dark:text-red-200"
                }
              >
                {task.status === "approved"
                  ? "This requisition has been approved and is proceeding to the next stage."
                  : "This requisition has been rejected. Contact the requester for more information."}
              </AlertDescription>
            </Alert>
          )}
        </div>

        {/* Right Column - Workflow Progress */}
        <div className="space-y-6">
          {/* Workflow Timeline */}
          {workflow && (
            <ApprovalFlowDisplay
              workflow={workflow}
              currentStageIndex={task.stageIndex || 0}
              approvals={taskData.relatedApprovals || []}
              isCompleted={task.status !== "pending"}
            />
          )}

          {/* Quick Info */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Timeline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div>
                <h4 className="font-semibold text-muted-foreground mb-1">
                  Created
                </h4>
                <p className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  {new Date(task.createdAt || new Date()).toLocaleString()}
                </p>
              </div>

              {task.actionDate && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Action Date
                  </h4>
                  <p className="flex items-center gap-2">
                    <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
                    {new Date(task.actionDate).toLocaleString()}
                  </p>
                </div>
              )}

              {task.dueDate && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Due Date
                  </h4>
                  <p className="flex items-center gap-2">
                    <ClockIcon className="h-4 w-4 text-muted-foreground" />
                    {new Date(task.dueDate).toLocaleString()}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Approval History */}
      <ApprovalHistory
        entityId={task.entityId || taskId}
        entityType={task.entityType || "Requisition"}
      />
    </div>
  );
}
