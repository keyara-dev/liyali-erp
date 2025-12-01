# Approval Modal Patterns

This document provides guidance on using the standardized ApprovalConfirmationModal component for all approval and rejection workflows in Liyali Gateway.

## Overview

The `ApprovalConfirmationModal` ensures consistent user experience across all document approval workflows with:
- Required digital signatures
- Required remarks for rejections
- Optional comments for both approval and rejection
- Clear confirmation with audit trail
- Accessible form validation

---

## Component Location

**File**: `src/components/approval-confirmation-modal.tsx`

**Type Definitions**:
- `ApprovalAction`: Union type for 'approve' | 'reject'
- `ApprovalData`: Interface containing signature, remarks, and comments

---

## Basic Usage

### Setup in Component

```typescript
'use client';

import { useState } from 'react';
import { ApprovalConfirmationModal, ApprovalData } from '@/components/approval-confirmation-modal';
import { useApproveBudget, useRejectBudget } from '@/hooks/use-budget-queries';

interface BudgetApprovalProps {
  budgetId: string;
  budgetTitle: string;
}

export function BudgetApprovalComponent({ budgetId, budgetTitle }: BudgetApprovalProps) {
  const [modalAction, setModalAction] = useState<'approve' | 'reject' | null>(null);
  const approveMutation = useApproveBudget(budgetId);
  const rejectMutation = useRejectBudget(budgetId);

  const handleApprovalConfirm = async (data: ApprovalData) => {
    if (modalAction === 'approve') {
      await approveMutation.mutateAsync({
        approvingUserId: userId,
        approvingUserRole: userRole,
        approvingUserName: userName,
        signature: data.signature,
        comments: data.comments,
      });
    } else if (modalAction === 'reject') {
      await rejectMutation.mutateAsync({
        rejectingUserId: userId,
        rejectingUserRole: userRole,
        rejectingUserName: userName,
        remarks: data.remarks!,
        signature: data.signature,
        comments: data.comments,
      });
    }
  };

  return (
    <>
      {/* Approval Buttons */}
      <div className="flex gap-2">
        <button
          onClick={() => setModalAction('approve')}
          className="bg-green-600 text-white px-4 py-2 rounded"
        >
          Approve
        </button>
        <button
          onClick={() => setModalAction('reject')}
          className="bg-red-600 text-white px-4 py-2 rounded"
        >
          Reject
        </button>
      </div>

      {/* Approval Modal */}
      <ApprovalConfirmationModal
        isOpen={modalAction !== null}
        onClose={() => setModalAction(null)}
        onConfirm={handleApprovalConfirm}
        action={modalAction as 'approve' | 'reject'}
        documentTitle={budgetTitle}
        documentNumber={budgetId}
        isLoading={approveMutation.isPending || rejectMutation.isPending}
      />
    </>
  );
}
```

---

## Props Reference

```typescript
interface ApprovalConfirmationModalProps {
  // State management
  isOpen: boolean;                    // Whether modal is visible
  onClose: () => void;                // Callback to close modal

  // Submission
  onConfirm: (data: ApprovalData) => Promise<void>;  // Called when approved/rejected

  // Configuration
  action: 'approve' | 'reject';       // Type of action
  documentTitle: string;              // Display name of document
  documentNumber?: string;            // Document identifier (optional)

  // Loading state
  isLoading?: boolean;                // Disable modal during submission
}
```

---

## ApprovalData Response

```typescript
interface ApprovalData {
  signature: string;                  // Digital signature (base64 PNG) - REQUIRED
  remarks?: string;                   // Rejection reason - REQUIRED for rejection
  comments?: string;                  // Optional comments - both action types
}
```

---

## Feature Breakdown

### 1. Digital Signature (Required)

- Canvas-based drawing interface
- Captures signature as base64 PNG
- "Clear Signature" button to redraw
- Visual confirmation when captured
- Required for both approval and rejection

**Used for:**
- Legal authorization
- Audit trail
- Non-repudiation

### 2. Remarks (Required for Rejection Only)

- Multi-line text area
- Minimum length validation
- Hidden for approval action
- Visible and required for rejection action
- Helps document rejection reasons

**Best practices:**
- Be specific and constructive
- Explain what needs to be fixed
- Provide actionable feedback
- Keep professional tone

### 3. Optional Comments

- Available for both approval and rejection
- Additional context or instructions
- Not required but recommended
- Appears in approval history

---

## Usage Examples

### Example 1: Budget Approval

```typescript
import { ApprovalConfirmationModal, ApprovalData } from '@/components/approval-confirmation-modal';
import { useApproveBudget } from '@/hooks/use-budget-queries';

export function BudgetApprovalPanel({ budgetId, budget }: BudgetApprovalPanelProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const approveMutation = useApproveBudget(budgetId, () => {
    // Callback after successful approval
    router.refresh();
  });

  const handleApprove = async (data: ApprovalData) => {
    await approveMutation.mutateAsync({
      approvingUserId: session.user.id,
      approvingUserName: session.user.name,
      approvingUserRole: session.user.role,
      signature: data.signature,
      comments: data.comments,
    });
  };

  return (
    <>
      <button
        onClick={() => setIsModalOpen(true)}
        className="bg-green-600 text-white px-4 py-2 rounded"
      >
        Approve Budget
      </button>

      <ApprovalConfirmationModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onConfirm={handleApprove}
        action="approve"
        documentTitle={budget.name}
        documentNumber={budget.budgetNumber}
        isLoading={approveMutation.isPending}
      />
    </>
  );
}
```

### Example 2: Requisition Rejection

```typescript
import { ApprovalConfirmationModal, ApprovalData } from '@/components/approval-confirmation-modal';
import { useRejectRequisition } from '@/hooks/use-requisition-queries';

export function RequisitionApprovalPanel({ requisitionId, requisition }: RequisitionApprovalPanelProps) {
  const [isRejectModalOpen, setIsRejectModalOpen] = useState(false);
  const rejectMutation = useRejectRequisition(requisitionId, () => {
    // Callback after successful rejection
    toast.success('Requisition rejected successfully');
  });

  const handleReject = async (data: ApprovalData) => {
    await rejectMutation.mutateAsync({
      rejectingUserId: session.user.id,
      rejectingUserName: session.user.name,
      rejectingUserRole: session.user.role,
      remarks: data.remarks!, // Required for rejection
      signature: data.signature,
      comments: data.comments, // Optional
    });
  };

  return (
    <>
      <button
        onClick={() => setIsRejectModalOpen(true)}
        className="bg-red-600 text-white px-4 py-2 rounded"
      >
        Reject Requisition
      </button>

      <ApprovalConfirmationModal
        isOpen={isRejectModalOpen}
        onClose={() => setIsRejectModalOpen(false)}
        onConfirm={handleReject}
        action="reject"
        documentTitle={requisition.title}
        documentNumber={requisition.requisitionNumber}
        isLoading={rejectMutation.isPending}
      />
    </>
  );
}
```

### Example 3: Dual Action Button

```typescript
import { ApprovalConfirmationModal, ApprovalData, ApprovalAction } from '@/components/approval-confirmation-modal';
import { useApproveBudget, useRejectBudget } from '@/hooks/use-budget-queries';

export function BudgetActionButtons({ budgetId, budget }: BudgetActionProps) {
  const [modalAction, setModalAction] = useState<ApprovalAction | null>(null);

  const approveMutation = useApproveBudget(budgetId);
  const rejectMutation = useRejectBudget(budgetId);

  const handleConfirm = async (data: ApprovalData) => {
    if (modalAction === 'approve') {
      await approveMutation.mutateAsync({
        approvingUserId: userId,
        approvingUserName: userName,
        approvingUserRole: userRole,
        signature: data.signature,
        comments: data.comments,
      });
    } else if (modalAction === 'reject') {
      await rejectMutation.mutateAsync({
        rejectingUserId: userId,
        rejectingUserName: userName,
        rejectingUserRole: userRole,
        remarks: data.remarks!,
        signature: data.signature,
        comments: data.comments,
      });
    }
  };

  return (
    <>
      <div className="flex gap-2">
        <button
          onClick={() => setModalAction('approve')}
          className="bg-green-600 text-white px-4 py-2 rounded"
          disabled={approveMutation.isPending || rejectMutation.isPending}
        >
          Approve
        </button>
        <button
          onClick={() => setModalAction('reject')}
          className="bg-red-600 text-white px-4 py-2 rounded"
          disabled={approveMutation.isPending || rejectMutation.isPending}
        >
          Reject
        </button>
      </div>

      <ApprovalConfirmationModal
        isOpen={modalAction !== null}
        onClose={() => setModalAction(null)}
        onConfirm={handleConfirm}
        action={modalAction!}
        documentTitle={budget.name}
        documentNumber={budget.budgetNumber}
        isLoading={approveMutation.isPending || rejectMutation.isPending}
      />
    </>
  );
}
```

---

## Validation Rules

### Signature
- **Required**: Yes (both actions)
- **Format**: Base64 PNG data URL
- **Must**: Have visible strokes drawn on canvas

### Remarks
- **Required**: Yes (rejection only)
- **Minimum length**: 1 character
- **Recommended**: 10+ characters for clarity
- **Not shown**: During approval action

### Comments
- **Required**: No (optional)
- **Minimum length**: None
- **Shown**: Both approval and rejection
- **Use for**: Additional context or instructions

---

## Error Handling

The modal validates inputs and provides clear error messages:

```typescript
// Error: Missing signature
"Signature is required"

// Error: Missing remarks on rejection
"Remarks are required for rejection"
```

Errors appear in red box with AlertCircle icon. Clear after user starts fixing the issue.

---

## Integration with Hooks

### Pattern: Mutation with onSuccess Callback

```typescript
// Server action defines behavior
export const useApproveBudget = (budgetId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<ApproveBudgetRequest, 'budgetId'>) => {
      const response = await approveBudget({ budgetId, ...data });
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: () => {
      toast.success('Budget approved successfully');
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId] });
      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to approve budget');
    },
  });
};

// Component uses mutation with modal
const approveMutation = useApproveBudget(budgetId, () => {
  setIsModalOpen(false); // Close modal on success
});

const handleApproveClick = async (data: ApprovalData) => {
  await approveMutation.mutateAsync({
    approvingUserId: userId,
    approvingUserName: userName,
    approvingUserRole: userRole,
    signature: data.signature,
    comments: data.comments,
  });
};
```

---

## Best Practices

### Do's ✅

- ✅ Always use the modal for approvals and rejections
- ✅ Require signature for legal compliance
- ✅ Require remarks for rejections (user feedback)
- ✅ Disable buttons while loading
- ✅ Show clear confirmation messages
- ✅ Provide actionable error messages
- ✅ Include document details in modal header
- ✅ Reset form after successful submission
- ✅ Prevent double-submission with loading state

### Don'ts ❌

- ❌ Don't skip signature requirement
- ❌ Don't allow blank remarks on rejection
- ❌ Don't leave modal open after submission
- ❌ Don't forget to invalidate queries on success
- ❌ Don't hide error messages
- ❌ Don't allow form submission while loading
- ❌ Don't skip document identifier in modal
- ❌ Don't mix approval/rejection logic

---

## Accessibility

The modal includes:

- Semantic HTML with proper heading hierarchy
- ARIA labels on form fields
- Clear icon indicators for action type
- Error messages announced to screen readers
- Keyboard navigation support
- Focus management
- Proper color contrast

---

## Styling and Customization

The modal uses:
- Tailwind CSS classes
- Shadcn/ui components (Dialog, Button, Textarea)
- Conditional styling based on action type:
  - Approval: Green accent (CheckCircle2 icon)
  - Rejection: Red accent (AlertCircle icon)

### Button Colors

```typescript
// Approval
<Button variant="default" onClick={handleSubmit}>
  Confirm Approval
</Button>

// Rejection
<Button variant="destructive" onClick={handleSubmit}>
  Confirm Rejection
</Button>
```

---

## Testing

### Unit Test Example

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { ApprovalConfirmationModal, ApprovalData } from '@/components/approval-confirmation-modal';

describe('ApprovalConfirmationModal', () => {
  it('requires signature for approval', async () => {
    const mockOnConfirm = jest.fn();

    render(
      <ApprovalConfirmationModal
        isOpen={true}
        onClose={jest.fn()}
        onConfirm={mockOnConfirm}
        action="approve"
        documentTitle="Test Budget"
      />
    );

    const submitButton = screen.getByRole('button', { name: /confirm approval/i });
    fireEvent.click(submitButton);

    expect(screen.getByText('Signature is required')).toBeInTheDocument();
    expect(mockOnConfirm).not.toHaveBeenCalled();
  });

  it('requires remarks for rejection', async () => {
    const mockOnConfirm = jest.fn();

    render(
      <ApprovalConfirmationModal
        isOpen={true}
        onClose={jest.fn()}
        onConfirm={mockOnConfirm}
        action="reject"
        documentTitle="Test Budget"
      />
    );

    // Draw signature
    const canvas = screen.getByRole('button', { name: /clear signature/i });
    // ... simulate signature drawing

    const submitButton = screen.getByRole('button', { name: /confirm rejection/i });
    fireEvent.click(submitButton);

    expect(screen.getByText('Remarks are required for rejection')).toBeInTheDocument();
    expect(mockOnConfirm).not.toHaveBeenCalled();
  });
});
```

---

## FAQs

**Q: Can I use this modal for other workflows besides approvals?**
A: This modal is designed specifically for approvals and rejections. For other actions, create specialized modals.

**Q: How do I change the signature canvas size?**
A: The SignatureCanvas component controls this. Modify the component props if needed.

**Q: Can users save incomplete approvals?**
A: No, all required fields (signature, remarks for rejection) must be filled before submission is allowed.

**Q: How are signatures stored?**
A: Signatures are stored as base64 PNG data URLs in the approval record for audit trail and compliance.

**Q: Can I customize the confirmation message?**
A: Yes, the modal accepts `documentTitle` and `documentNumber` props for customization.

---

## Integration Checklist

When adding approval to a new document type:

- [ ] Import ApprovalConfirmationModal
- [ ] Import/create mutation hook (useApprove*, useReject*)
- [ ] Add state for modal open/action
- [ ] Create handleConfirm function
- [ ] Add approval button(s)
- [ ] Add rejection button(s) (if applicable)
- [ ] Connect modal to buttons
- [ ] Pass required props (isOpen, action, title, number)
- [ ] Test signature requirement
- [ ] Test remarks requirement for rejection
- [ ] Test successful submission
- [ ] Test error handling

---

## Related Documentation

- [Query Hooks Patterns](./QUERY_HOOKS_PATTERNS.md) - Data fetching with mutations
- [Approval Workflow System](./FEATURES.md#4-approval-workflow-system) - Feature overview
- [Signature Canvas Component](../src/components/ui/signature-canvas.tsx) - Canvas implementation

---

**Last Updated**: 2025-11-30
**Version**: 1.0.0
