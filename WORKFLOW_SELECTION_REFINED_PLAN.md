# Workflow Selection - Refined Implementation Plan

## Current State

### ✅ What Already Exists

1. **Workflow Hooks** (`frontend/src/hooks/use-workflow-queries.ts`)
   - `useWorkflows(filter)` - Fetch all workflows with filtering
   - `useDefaultWorkflow(entityType)` - Get default workflow for document type
   - `useWorkflowById(workflowId)` - Get specific workflow details
   - All CRUD operations (create, update, delete, duplicate, activate, deactivate)

2. **Workflow Actions** (`frontend/src/app/_actions/workflows.ts`)
   - Already implemented and working
   - Connected to backend API endpoints

3. **Workflow Types** (`frontend/src/types/workflow-config.ts`)
   - Complete type definitions exist
   - Includes `Workflow`, `WorkflowFormData`, `WorkflowListFilter`

4. **Submit Dialogs**
   - Budget submit dialog exists
   - Other document types may have similar dialogs

### ❌ What's Missing

1. **Workflow Selector Component** - UI to select workflow
2. **workflowId in Submit Requests** - Not being sent to backend
3. **Integration in Submit Dialogs** - Workflow selection not included
4. **Updated Submit Types** - Types don't include workflowId

## Implementation Strategy

### Phase 1: Update Types & Actions (1-2 hours)

#### 1.1 Update Submit Request Types

Add `workflowId` to all document submit request interfaces:

**Files to Update:**

- `frontend/src/types/requisition.ts`
- `frontend/src/types/budget.ts`
- `frontend/src/types/purchase-order.ts`
- `frontend/src/types/payment-voucher.ts`
- `frontend/src/types/goods-received-note.ts`

**Example Change:**

```typescript
// BEFORE
export interface SubmitBudgetRequest {
  budgetId: string;
  submittedBy: string;
  comments?: string;
}

// AFTER
export interface SubmitBudgetRequest {
  budgetId: string;
  workflowId: string; // NEW - REQUIRED
  submittedBy: string;
  comments?: string;
}
```

#### 1.2 Update Submit Actions

Update all submit actions to send `workflowId` to backend:

**Files to Update:**

- `frontend/src/app/_actions/requisitions.ts`
- `frontend/src/app/_actions/budgets.ts`
- `frontend/src/app/_actions/purchase-orders.ts`
- `frontend/src/app/_actions/payment-vouchers.ts`
- `frontend/src/app/_actions/grns.ts`

**Example Change:**

```typescript
// In frontend/src/app/_actions/budgets.ts

export async function submitBudgetForApproval(
  request: SubmitBudgetRequest,
): Promise<APIResponse<Budget | null>> {
  try {
    const response = await authenticatedApiClient({
      method: "POST",
      url: `/api/v1/budgets/${request.budgetId}/submit`,
      data: {
        workflowId: request.workflowId, // NEW - REQUIRED
        comments: request.comments,
        submittedBy: request.submittedBy,
      },
    });

    return successResponse(
      response.data?.data,
      "Budget submitted for approval",
    );
  } catch (error: any) {
    return handleError(
      error,
      "POST",
      `/api/v1/budgets/${request.budgetId}/submit`,
    );
  }
}
```

### Phase 2: Create Workflow Selector Component (2-3 hours)

#### 2.1 Create WorkflowSelector Component

**File:** `frontend/src/components/workflows/workflow-selector.tsx` (NEW)

**Features:**

- Fetches workflows for specific document type using `useWorkflows`
- Auto-selects default workflow if available
- Shows workflow details (name, description, stages count)
- Handles loading and error states
- Validates selection

**Props:**

```typescript
interface WorkflowSelectorProps {
  entityType:
    | "requisition"
    | "purchase_order"
    | "budget"
    | "grn"
    | "payment_voucher";
  value: string;
  onChange: (workflowId: string) => void;
  disabled?: boolean;
  required?: boolean;
  error?: string;
  showDetails?: boolean; // Show workflow description and stages
}
```

**Implementation Approach:**

```typescript
export function WorkflowSelector({
  entityType,
  value,
  onChange,
  disabled = false,
  required = true,
  error,
  showDetails = true,
}: WorkflowSelectorProps) {
  // Fetch workflows for this entity type
  const { data: workflows, isLoading, error: fetchError } = useWorkflows({
    entityType,
    activeOnly: true,
  });

  // Fetch default workflow
  const { data: defaultWorkflow } = useDefaultWorkflow(entityType);

  // Auto-select default workflow on mount
  useEffect(() => {
    if (!value && defaultWorkflow) {
      onChange(defaultWorkflow.id);
    } else if (!value && workflows && workflows.length > 0) {
      // If no default, select first workflow
      onChange(workflows[0].id);
    }
  }, [defaultWorkflow, workflows, value, onChange]);

  // Find selected workflow for details
  const selectedWorkflow = workflows?.find(w => w.id === value);

  // Render loading state
  if (isLoading) {
    return <LoadingState />;
  }

  // Render error state
  if (fetchError) {
    return <ErrorState error={fetchError} />;
  }

  // Render no workflows state
  if (!workflows || workflows.length === 0) {
    return <NoWorkflowsState />;
  }

  // Render selector
  return (
    <div className="space-y-2">
      <SelectField
        label="Approval Workflow"
        value={value}
        onChange={onChange}
        options={workflows.map(w => ({
          value: w.id,
          label: w.name,
          badge: w.isDefault ? 'Default' : undefined,
        }))}
        placeholder="Select a workflow"
        disabled={disabled}
        required={required}
        error={error}
      />

      {showDetails && selectedWorkflow && (
        <WorkflowDetails workflow={selectedWorkflow} />
      )}
    </div>
  );
}
```

#### 2.2 Create WorkflowDetails Component (Optional)

**File:** `frontend/src/components/workflows/workflow-details.tsx` (NEW)

Shows workflow information in a compact card:

- Description
- Number of stages
- Estimated approval time (if available)

### Phase 3: Update Submit Dialogs (2-3 hours)

Update all existing submit dialogs to include workflow selection.

#### 3.1 Budget Submit Dialog

**File:** `frontend/src/app/(private)/(main)/budgets/[id]/_components/budget-submit-dialog.tsx`

**Changes:**

```typescript
export function BudgetSubmitDialog({
  open,
  onOpenChange,
  budget,
  onSubmit,
  isSubmitting,
}: BudgetSubmitDialogProps) {
  const [comments, setComments] = useState("");
  const [workflowId, setWorkflowId] = useState(""); // NEW
  const [workflowError, setWorkflowError] = useState<string | null>(null); // NEW

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate workflow selection
    if (!workflowId) {
      setWorkflowError("Please select a workflow");
      return;
    }

    if (!canSubmit) return;

    setWorkflowError(null);
    await onSubmit(workflowId, comments); // Pass workflowId
    setComments("");
    setWorkflowId("");
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-lg">
        {/* ... existing header ... */}

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* NEW: Workflow Selector */}
          <WorkflowSelector
            entityType="budget"
            value={workflowId}
            onChange={setWorkflowId}
            disabled={isSubmitting}
            required
            error={workflowError || undefined}
          />

          {/* ... existing budget summary ... */}
          {/* ... existing validation alerts ... */}
          {/* ... existing comments field ... */}
          {/* ... existing actions ... */}
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

**Update Props:**

```typescript
interface BudgetSubmitDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  budget: Budget;
  onSubmit: (workflowId: string, comments?: string) => Promise<void>; // Updated
  isSubmitting: boolean;
}
```

#### 3.2 Other Submit Dialogs

Apply similar changes to:

- `frontend/src/app/(private)/(main)/requisitions/_components/requisition-submit-dialog.tsx`
- `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-submit-dialog.tsx`
- `frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-voucher-submit-dialog.tsx`
- `frontend/src/app/(private)/(main)/grns/_components/grn-submit-dialog.tsx`

**Note:** Some of these dialogs may not exist yet. If they don't, create them following the budget dialog pattern.

### Phase 4: Update Detail Pages (1-2 hours)

Update all document detail pages to pass `workflowId` to submit mutations.

#### 4.1 Budget Detail Page

**File:** `frontend/src/app/(private)/(main)/budgets/[id]/_components/budget-detail-client.tsx`

**Changes:**

```typescript
// Update submit handler
const handleSubmit = async (workflowId: string, comments?: string) => {
  if (!user) return;

  await submitMutation.mutateAsync({
    budgetId: budget.id,
    workflowId, // NEW - REQUIRED
    submittedBy: user.id,
    comments,
  });

  setShowSubmitDialog(false);
};

// In JSX
<BudgetSubmitDialog
  open={showSubmitDialog}
  onOpenChange={setShowSubmitDialog}
  budget={budget}
  onSubmit={handleSubmit} // Now expects workflowId
  isSubmitting={submitMutation.isPending}
/>
```

#### 4.2 Other Detail Pages

Apply similar changes to:

- `frontend/src/app/(private)/(main)/requisitions/[id]/_components/requisition-detail-client.tsx`
- `frontend/src/app/(private)/(main)/purchase-orders/[id]/_components/purchase-order-detail-client.tsx`
- `frontend/src/app/(private)/(main)/payment-vouchers/[id]/_components/payment-voucher-detail-client.tsx`
- `frontend/src/app/(private)/(main)/grns/[id]/_components/grn-detail-client.tsx`

### Phase 5: Update Hooks (1 hour)

Update submit mutation hooks to accept and pass `workflowId`.

#### 5.1 Budget Hooks

**File:** `frontend/src/hooks/use-budget-queries.ts`

**Changes:**

```typescript
export const useSubmitBudgetForApproval = (
  budgetId: string,
  onSuccess?: () => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      submittingUserId: string;
      workflowId: string; // NEW - REQUIRED
      comments?: string;
    }) => {
      const response = await submitBudgetForApproval({
        budgetId,
        workflowId: data.workflowId, // NEW
        submittedBy: data.submittingUserId,
        comments: data.comments,
      });

      if (!response.success) {
        throw new Error(response.message);
      }

      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.ALL] });
      queryClient.setQueryData([QUERY_KEYS.BUDGETS.DETAIL, budgetId], data);
      toast.success("Budget submitted for approval");
      onSuccess?.();
    },
    onError: (error: any) => {
      toast.error(error.message || "Failed to submit budget");
    },
  });
};
```

#### 5.2 Other Hooks

Apply similar changes to:

- `frontend/src/hooks/use-requisition-queries.ts`
- `frontend/src/hooks/use-purchase-order-queries.ts`
- `frontend/src/hooks/use-payment-voucher-queries.ts`
- `frontend/src/hooks/use-grn-queries.ts`

## Implementation Checklist

### Phase 1: Types & Actions ✅

- [ ] Update `frontend/src/types/requisition.ts` - Add workflowId
- [ ] Update `frontend/src/types/budget.ts` - Add workflowId
- [ ] Update `frontend/src/types/purchase-order.ts` - Add workflowId
- [ ] Update `frontend/src/types/payment-voucher.ts` - Add workflowId
- [ ] Update `frontend/src/types/goods-received-note.ts` - Add workflowId
- [ ] Update `frontend/src/app/_actions/requisitions.ts` - Send workflowId
- [ ] Update `frontend/src/app/_actions/budgets.ts` - Send workflowId
- [ ] Update `frontend/src/app/_actions/purchase-orders.ts` - Send workflowId
- [ ] Update `frontend/src/app/_actions/payment-vouchers.ts` - Send workflowId
- [ ] Update `frontend/src/app/_actions/grns.ts` - Send workflowId

### Phase 2: Components ✅

- [ ] Create `frontend/src/components/workflows/workflow-selector.tsx`
- [ ] Create `frontend/src/components/workflows/workflow-details.tsx` (optional)
- [ ] Test WorkflowSelector with different entity types
- [ ] Test auto-selection of default workflow
- [ ] Test error handling (no workflows, fetch error)

### Phase 3: Submit Dialogs ✅

- [ ] Update `budget-submit-dialog.tsx` - Add workflow selector
- [ ] Update/Create `requisition-submit-dialog.tsx` - Add workflow selector
- [ ] Update/Create `purchase-order-submit-dialog.tsx` - Add workflow selector
- [ ] Update/Create `payment-voucher-submit-dialog.tsx` - Add workflow selector
- [ ] Update/Create `grn-submit-dialog.tsx` - Add workflow selector

### Phase 4: Detail Pages ✅

- [ ] Update `budget-detail-client.tsx` - Pass workflowId
- [ ] Update `requisition-detail-client.tsx` - Pass workflowId
- [ ] Update `purchase-order-detail-client.tsx` - Pass workflowId
- [ ] Update `payment-voucher-detail-client.tsx` - Pass workflowId
- [ ] Update `grn-detail-client.tsx` - Pass workflowId

### Phase 5: Hooks ✅

- [ ] Update `use-budget-queries.ts` - Accept workflowId
- [ ] Update `use-requisition-queries.ts` - Accept workflowId
- [ ] Update `use-purchase-order-queries.ts` - Accept workflowId
- [ ] Update `use-payment-voucher-queries.ts` - Accept workflowId
- [ ] Update `use-grn-queries.ts` - Accept workflowId

### Testing ✅

- [ ] Test budget submission with workflow selection
- [ ] Test requisition submission with workflow selection
- [ ] Test purchase order submission with workflow selection
- [ ] Test payment voucher submission with workflow selection
- [ ] Test GRN submission with workflow selection
- [ ] Test with no workflows available
- [ ] Test with single workflow (auto-select)
- [ ] Test with multiple workflows
- [ ] Test default workflow selection
- [ ] Test validation (workflowId required)
- [ ] Test error handling

## Document Types & Entity Types Mapping

| Document Type   | Entity Type       | Submit Endpoint                            |
| --------------- | ----------------- | ------------------------------------------ |
| Requisition     | `requisition`     | `POST /api/v1/requisitions/:id/submit`     |
| Budget          | `budget`          | `POST /api/v1/budgets/:id/submit`          |
| Purchase Order  | `purchase_order`  | `POST /api/v1/purchase-orders/:id/submit`  |
| Payment Voucher | `payment_voucher` | `POST /api/v1/payment-vouchers/:id/submit` |
| GRN             | `grn`             | `POST /api/v1/grns/:id/submit`             |

## Estimated Timeline

| Phase     | Tasks                  | Estimated Time |
| --------- | ---------------------- | -------------- |
| Phase 1   | Update types & actions | 1-2 hours      |
| Phase 2   | Create components      | 2-3 hours      |
| Phase 3   | Update submit dialogs  | 2-3 hours      |
| Phase 4   | Update detail pages    | 1-2 hours      |
| Phase 5   | Update hooks           | 1 hour         |
| Testing   | Comprehensive testing  | 2 hours        |
| **Total** |                        | **9-13 hours** |

## Priority Order

1. **Budget** (Has existing submit dialog) - Start here
2. **Requisition** (Most common document type)
3. **Purchase Order** (Follows requisition)
4. **Payment Voucher** (Follows PO)
5. **GRN** (Goods receipt)

## Notes

1. The workflow hooks (`use-workflow-queries.ts`) are already complete and ready to use
2. The backend is fully ready and requires `workflowId` in all submit requests
3. Start with Budget since it already has a submit dialog
4. Use the Budget implementation as a template for other document types
5. Consider creating a shared `DocumentSubmitDialog` component to reduce duplication
6. The `SelectField` component already exists and can be used in the workflow selector
7. Auto-selecting the default workflow improves UX
8. Show workflow details (stages count, description) to help users make informed choices

## Success Criteria

- [ ] Users can select a workflow before submitting any document type
- [ ] Default workflow is auto-selected when available
- [ ] Workflow details are displayed to help users understand the approval process
- [ ] Submit requests include `workflowId` and succeed on backend
- [ ] Error handling works for all edge cases
- [ ] UI is consistent across all document types
- [ ] No breaking changes to existing functionality
