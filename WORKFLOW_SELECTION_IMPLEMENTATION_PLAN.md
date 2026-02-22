# Workflow Selection Implementation Plan

## Overview

Implement a feature that allows users to select a workflow before submitting documents for approval. This applies to all document types: Requisitions, Purchase Orders, Budgets, Payment Vouchers, and GRNs.

## Current State Analysis

### Backend Implementation ✅

The backend is **FULLY READY** to support workflow selection:

#### 1. Workflow Model

**File:** `backend/models/enhanced_auth.go`

```go
type Workflow struct {
    ID             uuid.UUID      `json:"id"`
    OrganizationID string         `json:"organizationId"`
    Name           string         `json:"name"`
    Description    string         `json:"description"`
    DocumentType   string         `json:"documentType"` // For compatibility
    EntityType     string         `json:"entityType"`   // "requisition", "purchase_order", "budget", "grn", "payment_voucher"
    Version        int            `json:"version"`
    IsActive       bool           `json:"isActive"`
    IsDefault      bool           `json:"isDefault"`
    Conditions     datatypes.JSON `json:"conditions,omitempty"`
    Stages         datatypes.JSON `json:"stages"`
    // ... other fields
}
```

#### 2. Available API Endpoints

**Get Workflows (with filtering):**

```
GET /api/v1/workflows?entityType={type}&activeOnly=true
```

**Get Default Workflow:**

```
GET /api/v1/workflows/default/:documentType
```

**Get Workflow by ID:**

```
GET /api/v1/workflows/:id
```

#### 3. Submit Request Structure

**File:** `backend/types/documents.go`

```go
type SubmitDocumentRequest struct {
    WorkflowID string `json:"workflowId" validate:"required,uuid"`
}
```

#### 4. Submit Endpoints (All Document Types)

All submit handlers **REQUIRE** `workflowId` in the request body:

- `POST /api/v1/requisitions/:id/submit`
- `POST /api/v1/purchase-orders/:id/submit`
- `POST /api/v1/budgets/:id/submit`
- `POST /api/v1/payment-vouchers/:id/submit`
- `POST /api/v1/grns/:id/submit`

**Example from `backend/handlers/requisition.go`:**

```go
func SubmitRequisition(c *fiber.Ctx) error {
    var submitReq types.SubmitDocumentRequest
    if err := c.BodyParser(&submitReq); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Invalid request body",
        })
    }
    if submitReq.WorkflowID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "workflowId is required",
        })
    }
    // ... workflow assignment logic
}
```

### Frontend Current State ❌

The frontend is **NOT SENDING** `workflowId` in submit requests:

#### Current Submit Request Types

**Requisition:**

```typescript
export interface SubmitRequisitionRequest {
  requisitionId: string;
  submittedBy: string;
  submittedByName?: string;
  submittedByRole?: string;
  comments?: string;
}
```

**Budget:**

```typescript
export interface SubmitBudgetRequest {
  budgetId: string;
  submittedBy: string;
  comments?: string;
}
```

**Purchase Order:**

```typescript
export interface SubmitPurchaseOrderRequest {
  purchaseOrderId: string;
  poId?: string;
  submittedBy: string;
  comments?: string;
}
```

#### Current Submit Action (Example)

**File:** `frontend/src/app/_actions/requisitions.ts`

```typescript
export async function submitRequisitionForApproval(
  data: SubmitRequisitionRequest,
): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions/${data.requisitionId}/submit`;

  const response = await authenticatedApiClient({
    method: "POST",
    url,
    data: {
      comments: data.comments,
      submittedBy: data.submittedBy,
      submittedByName: data.submittedByName,
      submittedByRole: data.submittedByRole,
      // ❌ Missing: workflowId
    },
  });

  return successResponse(
    response.data?.data,
    "Requisition submitted for approval",
  );
}
```

## Gap Analysis

### What's Missing in Frontend

1. **No workflow selection UI** - Users cannot choose a workflow before submitting
2. **No workflowId in submit requests** - Backend requires it but frontend doesn't send it
3. **No workflow fetching** - Frontend doesn't fetch available workflows
4. **No workflow types** - TypeScript types don't include workflow data
5. **No workflow actions** - No API client functions to fetch workflows

## Implementation Plan

### Phase 1: Backend Integration (API Actions & Types)

#### 1.1 Create Workflow Types

**File:** `frontend/src/types/workflow.ts` (NEW)

```typescript
export interface Workflow {
  id: string;
  organizationId: string;
  name: string;
  description: string;
  documentType: string;
  entityType:
    | "requisition"
    | "purchase_order"
    | "budget"
    | "grn"
    | "payment_voucher";
  version: number;
  isActive: boolean;
  isDefault: boolean;
  conditions?: WorkflowConditions;
  stages: WorkflowStage[];
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  totalStages?: number;
  usageCount?: number;
}

export interface WorkflowStage {
  stageNumber: number;
  stageName: string;
  description?: string;
  requiredRole: string;
  requiredApprovals: number;
  timeoutHours?: number;
  canReject: boolean;
  canReassign: boolean;
  requiredApprovalCount: number;
  approvalType: "any" | "all" | "majority" | "quorum";
  quorumCount?: number;
  allowSelfApproval: boolean;
  requireUnanimous: boolean;
  escalationUserId?: string;
  assignmentStrategy: "role" | "round_robin" | "specific_user" | "user_group";
  assignedUserIds?: string[];
  assignedGroupId?: string;
}

export interface WorkflowConditions {
  amountRange?: {
    min?: number;
    max?: number;
  };
  departments?: string[];
  priority?: string[];
  categories?: string[];
  customFields?: Record<string, any>;
}

export interface WorkflowFilters {
  entityType?: string;
  activeOnly?: boolean;
  isDefault?: boolean;
  limit?: number;
  offset?: number;
}
```

#### 1.2 Create Workflow Actions

**File:** `frontend/src/app/_actions/workflows.ts` (NEW)

```typescript
import { authenticatedApiClient } from "@/lib/api-client";
import { APIResponse } from "@/types";
import { Workflow, WorkflowFilters } from "@/types/workflow";
import { successResponse, handleError } from "@/lib/response-helpers";

/**
 * Get workflows with optional filtering
 * Calls: GET /api/v1/workflows
 */
export async function getWorkflows(
  filters?: WorkflowFilters,
): Promise<APIResponse<Workflow[]>> {
  const url = "/api/v1/workflows";

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
      params: filters,
    });

    return successResponse(
      response.data?.data || [],
      "Workflows retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get workflows for a specific document type
 * Calls: GET /api/v1/workflows?entityType={type}&activeOnly=true
 */
export async function getWorkflowsForDocumentType(
  entityType: string,
): Promise<APIResponse<Workflow[]>> {
  return getWorkflows({
    entityType,
    activeOnly: true,
  });
}

/**
 * Get default workflow for a document type
 * Calls: GET /api/v1/workflows/default/:documentType
 */
export async function getDefaultWorkflow(
  documentType: string,
): Promise<APIResponse<Workflow>> {
  const url = `/api/v1/workflows/default/${documentType}`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return successResponse(
      response.data?.data,
      "Default workflow retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get workflow by ID
 * Calls: GET /api/v1/workflows/:id
 */
export async function getWorkflowById(
  workflowId: string,
): Promise<APIResponse<Workflow>> {
  const url = `/api/v1/workflows/${workflowId}`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return successResponse(
      response.data?.data,
      "Workflow retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}
```

#### 1.3 Update Submit Request Types

**Update all document submit request types to include workflowId:**

**File:** `frontend/src/types/requisition.ts`

```typescript
export interface SubmitRequisitionRequest {
  requisitionId: string;
  workflowId: string; // NEW - REQUIRED
  submittedBy: string;
  submittedByName?: string;
  submittedByRole?: string;
  comments?: string;
}
```

**File:** `frontend/src/types/budget.ts`

```typescript
export interface SubmitBudgetRequest {
  budgetId: string;
  workflowId: string; // NEW - REQUIRED
  submittedBy: string;
  comments?: string;
}
```

**File:** `frontend/src/types/purchase-order.ts`

```typescript
export interface SubmitPurchaseOrderRequest {
  purchaseOrderId: string;
  poId?: string;
  workflowId: string; // NEW - REQUIRED
  submittedBy: string;
  comments?: string;
}
```

**Similar updates for:**

- `frontend/src/types/payment-voucher.ts`
- `frontend/src/types/goods-received-note.ts`

#### 1.4 Update Submit Actions

**Update all submit actions to include workflowId in the request:**

**File:** `frontend/src/app/_actions/requisitions.ts`

```typescript
export async function submitRequisitionForApproval(
  data: SubmitRequisitionRequest,
): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions/${data.requisitionId}/submit`;

  try {
    const response = await authenticatedApiClient({
      method: "POST",
      url,
      data: {
        workflowId: data.workflowId, // NEW - REQUIRED
        comments: data.comments,
        submittedBy: data.submittedBy,
        submittedByName: data.submittedByName,
        submittedByRole: data.submittedByRole,
      },
    });

    return successResponse(
      response.data?.data,
      "Requisition submitted for approval",
    );
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}
```

**Similar updates for:**

- `frontend/src/app/_actions/budgets.ts`
- `frontend/src/app/_actions/purchase-orders.ts`
- `frontend/src/app/_actions/payment-vouchers.ts`
- `frontend/src/app/_actions/grns.ts`

### Phase 2: UI Components

#### 2.1 Create Workflow Selector Component

**File:** `frontend/src/components/workflows/workflow-selector.tsx` (NEW)

```typescript
"use client";

import { useEffect, useState } from "react";
import { SelectField } from "@/components/ui/select-field";
import { Workflow } from "@/types/workflow";
import { getWorkflowsForDocumentType, getDefaultWorkflow } from "@/app/_actions/workflows";
import { Loader2, Info } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface WorkflowSelectorProps {
  documentType: 'requisition' | 'purchase_order' | 'budget' | 'grn' | 'payment_voucher';
  value: string;
  onChange: (workflowId: string) => void;
  disabled?: boolean;
  required?: boolean;
  error?: string;
}

export function WorkflowSelector({
  documentType,
  value,
  onChange,
  disabled = false,
  required = true,
  error,
}: WorkflowSelectorProps) {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null);

  useEffect(() => {
    loadWorkflows();
  }, [documentType]);

  useEffect(() => {
    if (value && workflows.length > 0) {
      const workflow = workflows.find(w => w.id === value);
      setSelectedWorkflow(workflow || null);
    }
  }, [value, workflows]);

  const loadWorkflows = async () => {
    setLoading(true);
    setFetchError(null);

    try {
      // Fetch workflows for this document type
      const response = await getWorkflowsForDocumentType(documentType);

      if (response.success && response.data) {
        setWorkflows(response.data);

        // Auto-select default workflow if no value is set
        if (!value && response.data.length > 0) {
          const defaultWorkflow = response.data.find(w => w.isDefault);
          if (defaultWorkflow) {
            onChange(defaultWorkflow.id);
          } else {
            // If no default, select the first active workflow
            onChange(response.data[0].id);
          }
        }
      } else {
        setFetchError(response.message || 'Failed to load workflows');
      }
    } catch (err: any) {
      setFetchError(err.message || 'An error occurred while loading workflows');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading workflows...
      </div>
    );
  }

  if (fetchError) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{fetchError}</AlertDescription>
      </Alert>
    );
  }

  if (workflows.length === 0) {
    return (
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>
          No workflows available for this document type. Please contact your administrator.
        </AlertDescription>
      </Alert>
    );
  }

  const options = workflows.map(workflow => ({
    value: workflow.id,
    label: workflow.name,
    description: workflow.description,
    badge: workflow.isDefault ? 'Default' : undefined,
  }));

  return (
    <div className="space-y-2">
      <SelectField
        label="Approval Workflow"
        value={value}
        onChange={onChange}
        options={options}
        placeholder="Select a workflow"
        disabled={disabled}
        required={required}
        error={error}
      />

      {selectedWorkflow && (
        <div className="text-sm text-muted-foreground space-y-1">
          {selectedWorkflow.description && (
            <p>{selectedWorkflow.description}</p>
          )}
          <p className="flex items-center gap-1">
            <Info className="h-3 w-3" />
            {selectedWorkflow.totalStages || selectedWorkflow.stages?.length || 0} approval stages
          </p>
        </div>
      )}
    </div>
  );
}
```

#### 2.2 Update Submit Dialogs

Update all submit dialogs to include the workflow selector:

**Example: Requisition Submit Dialog**
**File:** `frontend/src/app/(private)/(main)/requisitions/_components/requisition-submit-dialog.tsx`

```typescript
"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { WorkflowSelector } from "@/components/workflows/workflow-selector";
import { Loader2 } from "lucide-react";

interface RequisitionSubmitDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (workflowId: string, comments?: string) => Promise<void>;
  isSubmitting: boolean;
}

export function RequisitionSubmitDialog({
  open,
  onOpenChange,
  onSubmit,
  isSubmitting,
}: RequisitionSubmitDialogProps) {
  const [workflowId, setWorkflowId] = useState("");
  const [comments, setComments] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!workflowId) {
      setError("Please select a workflow");
      return;
    }

    setError(null);
    await onSubmit(workflowId, comments);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Submit Requisition for Approval</DialogTitle>
          <DialogDescription>
            Select an approval workflow and optionally add comments before submitting.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <WorkflowSelector
            documentType="requisition"
            value={workflowId}
            onChange={setWorkflowId}
            disabled={isSubmitting}
            required
            error={error || undefined}
          />

          <div className="space-y-2">
            <Label htmlFor="comments">Comments (Optional)</Label>
            <Textarea
              id="comments"
              placeholder="Add any comments or notes..."
              value={comments}
              onChange={(e) => setComments(e.target.value)}
              disabled={isSubmitting}
              rows={3}
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isSubmitting || !workflowId}
          >
            {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {isSubmitting ? "Submitting..." : "Submit for Approval"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

### Phase 3: Integration with Existing Components

#### 3.1 Update Document Detail Pages

For each document type, update the submit button handler to:

1. Open a submit dialog with workflow selector
2. Pass the selected workflowId to the submit action

**Example: Requisition Detail Page**

```typescript
const [showSubmitDialog, setShowSubmitDialog] = useState(false);

const handleSubmit = async (workflowId: string, comments?: string) => {
  await submitMutation.mutateAsync({
    requisitionId: requisition.id,
    workflowId, // NEW
    submittedBy: user.id,
    submittedByName: user.name,
    submittedByRole: user.role,
    comments,
  });
  setShowSubmitDialog(false);
};

// In JSX:
<Button onClick={() => setShowSubmitDialog(true)}>
  Submit for Approval
</Button>

<RequisitionSubmitDialog
  open={showSubmitDialog}
  onOpenChange={setShowSubmitDialog}
  onSubmit={handleSubmit}
  isSubmitting={submitMutation.isPending}
/>
```

## Implementation Checklist

### Backend (Already Complete) ✅

- [x] Workflow model with entityType field
- [x] GET /api/v1/workflows endpoint with filtering
- [x] GET /api/v1/workflows/default/:documentType endpoint
- [x] Submit endpoints require workflowId
- [x] Workflow assignment logic

### Frontend - Phase 1: API Integration

- [ ] Create `frontend/src/types/workflow.ts`
- [ ] Create `frontend/src/app/_actions/workflows.ts`
- [ ] Update `frontend/src/types/requisition.ts` (add workflowId)
- [ ] Update `frontend/src/types/budget.ts` (add workflowId)
- [ ] Update `frontend/src/types/purchase-order.ts` (add workflowId)
- [ ] Update `frontend/src/types/payment-voucher.ts` (add workflowId)
- [ ] Update `frontend/src/types/goods-received-note.ts` (add workflowId)
- [ ] Update `frontend/src/app/_actions/requisitions.ts` (send workflowId)
- [ ] Update `frontend/src/app/_actions/budgets.ts` (send workflowId)
- [ ] Update `frontend/src/app/_actions/purchase-orders.ts` (send workflowId)
- [ ] Update `frontend/src/app/_actions/payment-vouchers.ts` (send workflowId)
- [ ] Update `frontend/src/app/_actions/grns.ts` (send workflowId)

### Frontend - Phase 2: UI Components

- [ ] Create `frontend/src/components/workflows/workflow-selector.tsx`
- [ ] Create `frontend/src/components/workflows/workflow-info-card.tsx` (optional)
- [ ] Create submit dialogs for each document type:
  - [ ] `requisition-submit-dialog.tsx`
  - [ ] `budget-submit-dialog.tsx`
  - [ ] `purchase-order-submit-dialog.tsx`
  - [ ] `payment-voucher-submit-dialog.tsx`
  - [ ] `grn-submit-dialog.tsx`

### Frontend - Phase 3: Integration

- [ ] Update requisition detail page
- [ ] Update budget detail page
- [ ] Update purchase order detail page
- [ ] Update payment voucher detail page
- [ ] Update GRN detail page
- [ ] Update hooks to pass workflowId
- [ ] Test all document types

### Testing

- [ ] Test workflow fetching for each document type
- [ ] Test default workflow selection
- [ ] Test manual workflow selection
- [ ] Test submit with selected workflow
- [ ] Test error handling (no workflows available)
- [ ] Test validation (workflowId required)
- [ ] Test with multiple workflows
- [ ] Test with single workflow (auto-select)

## Estimated Effort

- **Phase 1 (API Integration):** 2-3 hours
- **Phase 2 (UI Components):** 3-4 hours
- **Phase 3 (Integration):** 2-3 hours
- **Testing:** 2 hours
- **Total:** 9-12 hours

## Priority

**HIGH** - This is a critical feature as the backend requires workflowId for document submission. Without this, users cannot submit documents for approval.

## Notes

1. The backend is fully ready and waiting for the frontend to send workflowId
2. All submit endpoints will fail without workflowId in the request
3. The workflow selector should auto-select the default workflow if available
4. Consider caching workflows to avoid repeated API calls
5. The SelectField component already exists and can be used for the workflow selector
6. Consider adding workflow preview/details modal for users to see stages before selecting

## Related Files

### Backend

- `backend/models/enhanced_auth.go` - Workflow model
- `backend/handlers/workflow_handler.go` - Workflow API endpoints
- `backend/handlers/requisition.go` - Requisition submit handler
- `backend/handlers/budget.go` - Budget submit handler
- `backend/handlers/purchase_order.go` - PO submit handler
- `backend/handlers/payment_voucher.go` - PV submit handler
- `backend/handlers/grn.go` - GRN submit handler
- `backend/types/documents.go` - SubmitDocumentRequest type

### Frontend (To be created/updated)

- `frontend/src/types/workflow.ts` (NEW)
- `frontend/src/app/_actions/workflows.ts` (NEW)
- `frontend/src/components/workflows/workflow-selector.tsx` (NEW)
- All submit request types and actions (UPDATE)
- All document detail pages (UPDATE)
