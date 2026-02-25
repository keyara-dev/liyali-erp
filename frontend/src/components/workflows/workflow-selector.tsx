"use client";

import { useEffect, useState } from "react";
import { SelectField } from "@/components/ui/select-field";
import { useWorkflows, useDefaultWorkflow } from "@/hooks/use-workflow-queries";
import type { Workflow, WorkflowConditions } from "@/types/workflow-config";
import { Loader2, Info, AlertCircle, CheckCircle2, Zap, ShoppingCart } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { cn } from "@/lib/utils";

export interface WorkflowSelectorProps {
  entityType:
    | "requisition"
    | "purchase_order"
    | "budget"
    | "grn"
    | "payment_voucher";
  value: string;
  onChange: (workflowId: string) => void;
  onWorkflowSelect?: (workflow: Workflow | null) => void;
  disabled?: boolean;
  required?: boolean;
  error?: string;
  showDetails?: boolean;
  className?: string;
}

export function WorkflowSelector({
  entityType,
  value,
  onChange,
  onWorkflowSelect,
  disabled = false,
  required = true,
  error,
  showDetails = true,
  className,
}: WorkflowSelectorProps) {
  const [hasAutoSelected, setHasAutoSelected] = useState(false);

  // Fetch workflows for this entity type
  const {
    data: workflows,
    isLoading,
    error: fetchError,
  } = useWorkflows({
    filter: { entityType, isActive: true },
  });

  // Fetch default workflow
  const { data: defaultWorkflow } = useDefaultWorkflow(entityType);

  // Auto-select default workflow on mount
  useEffect(() => {
    if (hasAutoSelected) return;

    if (!value && defaultWorkflow) {
      onChange(defaultWorkflow.id);
      setHasAutoSelected(true);
    } else if (!value && workflows && workflows.length > 0) {
      // If no default, select the first workflow
      onChange(workflows[0].id);
      setHasAutoSelected(true);
    }
  }, [defaultWorkflow, workflows, value, onChange, hasAutoSelected]);

  // Find selected workflow for details
  const selectedWorkflow = workflows?.find((w) => w.id === value) ?? null;

  // Notify parent of selected workflow changes
  useEffect(() => {
    onWorkflowSelect?.(selectedWorkflow);
  }, [selectedWorkflow, onWorkflowSelect]);

  // Render loading state
  if (isLoading) {
    return (
      <div className={cn("space-y-2", className)}>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading workflows...
        </div>
      </div>
    );
  }

  // Render error state
  if (fetchError) {
    return (
      <div className={cn("space-y-2", className)}>
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load workflows. Please try again or contact support.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  // Render no workflows state
  if (!workflows || workflows.length === 0) {
    return (
      <div className={cn("space-y-2", className)}>
        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            No workflows available for this document type. Please contact your
            administrator to set up approval workflows.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  // Prepare options for SelectField with routing type indicator
  const options = workflows.map((workflow) => {
    const routingType = workflow.conditions?.routingType;
    const suffix = routingType === "accounting" ? " [Accounting]" : routingType === "procurement" ? " [Procurement]" : "";
    return {
      value: workflow.id,
      label: `${workflow.name}${suffix}`,
    };
  });

  // Handle select change - extract value from event or use directly
  const handleChange = (valueOrEvent: any) => {
    const newValue =
      typeof valueOrEvent === "string"
        ? valueOrEvent
        : valueOrEvent?.target?.value || valueOrEvent;
    onChange(newValue);
  };

  return (
    <div className={cn("space-y-3", className)}>
      <div className="space-y-2">
        <SelectField
          label="Approval Workflow"
          value={value}
          onChange={handleChange}
          options={options}
          placeholder="Select a workflow"
          disabled={disabled}
          required={required}
          error={error}
        />

        {selectedWorkflow && (
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <CheckCircle2 className="h-3 w-3 text-green-600" />
            <span>Selected workflow for {entityType.replace("_", " ")}</span>
          </div>
        )}
      </div>

      {showDetails && selectedWorkflow && (
        <WorkflowDetails workflow={selectedWorkflow} />
      )}
    </div>
  );
}

interface WorkflowDetailsProps {
  workflow: any; // Using any to handle different workflow type structures
}

function WorkflowDetails({ workflow }: WorkflowDetailsProps) {
  const stagesCount = workflow.stages?.length || 0;
  const conditions: WorkflowConditions | undefined = workflow.conditions;
  const routingType = conditions?.routingType || "procurement";
  const isAccounting = routingType === "accounting";

  return (
    <div className="rounded-lg border bg-muted/50 p-3 space-y-2">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 space-y-1">
          <div className="flex items-center gap-2">
            <h4 className="text-sm font-medium">{workflow.name}</h4>
            {isAccounting ? (
              <span className="inline-flex items-center gap-1 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-medium text-amber-800">
                <Zap className="h-3 w-3" />
                Accounting
              </span>
            ) : (
              <span className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-2 py-0.5 text-[10px] font-medium text-blue-800">
                <ShoppingCart className="h-3 w-3" />
                Procurement
              </span>
            )}
          </div>

          {workflow.description && (
            <p className="text-xs text-muted-foreground">
              {workflow.description}
            </p>
          )}
        </div>
      </div>

      <div className="flex items-center gap-4 text-xs text-muted-foreground">
        <div className="flex items-center gap-1">
          <Info className="h-3 w-3" />
          <span>
            {stagesCount} approval {stagesCount === 1 ? "stage" : "stages"}
          </span>
        </div>
        {isAccounting && conditions?.autoApprove && (
          <div className="flex items-center gap-1 text-amber-600">
            <Zap className="h-3 w-3" />
            <span>
              Auto-approval
              {conditions.autoApprovalMaxAmount
                ? ` up to ${conditions.autoApprovalMaxAmount.toLocaleString()}`
                : ""}
            </span>
          </div>
        )}
      </div>

      {workflow.stages && workflow.stages.length > 0 && (
        <div className="pt-2 border-t space-y-1">
          <p className="text-xs font-medium text-muted-foreground">
            Approval Stages:
          </p>
          <div className="space-y-1">
            {workflow.stages.slice(0, 3).map((stage: any, index: number) => (
              <div key={index} className="flex items-center gap-2 text-xs">
                <span className="flex items-center justify-center w-5 h-5 rounded-full bg-primary/10 text-primary font-medium">
                  {stage.stageNumber || index + 1}
                </span>
                <span className="text-muted-foreground">
                  {stage.stageName || stage.name}
                  {stage.requiredRole && (
                    <span className="text-xs ml-1">({stage.requiredRole})</span>
                  )}
                </span>
              </div>
            ))}
            {workflow.stages.length > 3 && (
              <p className="text-xs text-muted-foreground pl-7">
                +{workflow.stages.length - 3} more{" "}
                {workflow.stages.length - 3 === 1 ? "stage" : "stages"}
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
