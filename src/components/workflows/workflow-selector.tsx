"use client";

import { useState, useMemo } from "react";
import { Workflow } from "@/types";
import { useWorkflows } from "@/hooks/use-workflows";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  AlertCircle,
  CheckCircle2,
  ChevronRight,
  Workflow as WorkflowIcon,
} from "lucide-react";

export interface WorkflowSelectorProps {
  entityType: string;
  onSelect: (workflow: Workflow) => void;
  disabled?: boolean;
  showRecent?: boolean;
}

export function WorkflowSelector({
  entityType,
  onSelect,
  disabled = false,
  showRecent = true,
}: WorkflowSelectorProps) {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);

  const { data: workflows, isLoading } = useWorkflows();

  // Filter workflows by entity type and get published ones
  const availableWorkflows = useMemo(() => {
    if (!workflows) return [];
    return workflows.filter(
      (w) =>
        w.entityType === entityType &&
        w.status === "published" &&
        !w.isDeleted
    );
  }, [workflows, entityType]);

  // Get recently used workflows from localStorage
  const recentWorkflows = useMemo(() => {
    if (!showRecent) return [];
    try {
      const recent = localStorage.getItem(
        `workflow-recent-${entityType}`
      );
      if (!recent) return [];
      const ids = JSON.parse(recent) as string[];
      return availableWorkflows.filter((w) => ids.includes(w.id)).slice(0, 3);
    } catch {
      return [];
    }
  }, [availableWorkflows, showRecent, entityType]);

  const handleSelect = (workflow: Workflow) => {
    setSelectedId(workflow.id);
    setIsExpanded(false);
    onSelect(workflow);

    // Store in recent
    if (showRecent) {
      try {
        const key = `workflow-recent-${entityType}`;
        const recent = localStorage.getItem(key);
        const ids = recent ? JSON.parse(recent) : [];
        const updated = [
          workflow.id,
          ...ids.filter((id: string) => id !== workflow.id),
        ].slice(0, 5);
        localStorage.setItem(key, JSON.stringify(updated));
      } catch {
        // Silent fail on localStorage
      }
    }
  };

  const selectedWorkflow = availableWorkflows.find(
    (w) => w.id === selectedId
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-40" />
          <Skeleton className="h-4 w-60 mt-2" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-10 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (availableWorkflows.length === 0) {
    return (
      <Card className="border-yellow-200 bg-yellow-50 dark:bg-yellow-900/20">
        <CardContent className="flex items-center gap-3 pt-6">
          <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-500 flex-shrink-0" />
          <div>
            <h3 className="font-semibold text-yellow-900 dark:text-yellow-200">
              No Workflows Available
            </h3>
            <p className="text-sm text-yellow-800 dark:text-yellow-300">
              No published workflows for {entityType}. Contact your administrator.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <WorkflowIcon className="h-5 w-5" />
          Select Approval Workflow
        </CardTitle>
        <CardDescription>
          Choose a workflow to route this {entityType.toLowerCase()} through
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Recent Workflows */}
        {recentWorkflows.length > 0 && (
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-muted-foreground">
              Recently Used
            </h3>
            <div className="grid gap-2">
              {recentWorkflows.map((workflow) => (
                <button
                  key={workflow.id}
                  onClick={() => handleSelect(workflow)}
                  className="flex items-center justify-between p-3 rounded-lg border border-primary/20 bg-primary/5 hover:bg-primary/10 transition-colors text-left"
                >
                  <div className="flex-1 min-w-0">
                    <h4 className="font-medium text-sm">{workflow.name}</h4>
                    <p className="text-xs text-muted-foreground truncate">
                      {workflow.description}
                    </p>
                  </div>
                  <ChevronRight className="h-4 w-4 text-muted-foreground flex-shrink-0 ml-2" />
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Workflow Selector */}
        <div className="space-y-2">
          <h3 className="text-sm font-medium text-muted-foreground">
            All Workflows
          </h3>
          <Select
            value={selectedId || ""}
            onValueChange={(id) => {
              const workflow = availableWorkflows.find((w) => w.id === id);
              if (workflow) {
                handleSelect(workflow);
              }
            }}
            disabled={disabled}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select a workflow..." />
            </SelectTrigger>
            <SelectContent>
              {availableWorkflows.map((workflow) => (
                <SelectItem key={workflow.id} value={workflow.id}>
                  <div className="flex items-center gap-2">
                    <span>{workflow.name}</span>
                    {workflow.id === selectedId && (
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                    )}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Selected Workflow Details */}
        {selectedWorkflow && (
          <div className="pt-4 border-t space-y-3">
            <div>
              <h3 className="font-semibold text-base">{selectedWorkflow.name}</h3>
              <p className="text-sm text-muted-foreground">
                {selectedWorkflow.description}
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                  Stages
                </h4>
                <p className="text-lg font-bold">
                  {selectedWorkflow.stages?.length || 0}
                </p>
              </div>
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                  Status
                </h4>
                <Badge className="mt-1">Published</Badge>
              </div>
            </div>

            {selectedWorkflow.stages && selectedWorkflow.stages.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold mb-2">Approval Chain</h4>
                <div className="space-y-1">
                  {selectedWorkflow.stages.map((stage, index) => (
                    <div
                      key={index}
                      className="flex items-center gap-2 text-xs"
                    >
                      <span className="inline-flex items-center justify-center w-5 h-5 rounded-full bg-primary/20 text-primary font-semibold">
                        {index + 1}
                      </span>
                      <span className="font-medium">{stage.name}</span>
                      {stage.approvers && stage.approvers.length > 0 && (
                        <span className="text-muted-foreground">
                          ({stage.approvers.length} approver
                          {stage.approvers.length !== 1 ? "s" : ""})
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <Button
              onClick={() => onSelect(selectedWorkflow)}
              className="w-full"
            >
              Use This Workflow
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
