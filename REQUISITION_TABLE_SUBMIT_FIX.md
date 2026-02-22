# Requisition Table Submit Error Fix

## Problem

The requisitions table had a "Submit for Approval" action in the dropdown menu that was trying to submit without the required `workflowId` parameter, causing a TypeScript error:

```
Property 'workflowId' is missing in type '{ submittedBy: string; submittedByName: string; submittedByRole: string; comments: string; }' but required in type 'Omit<SubmitRequisitionRequest, "requisitionId">'
```

## Root Cause

The backend API requires `workflowId` in all document submit requests. The quick submit action from the table dropdown was attempting to submit without workflow selection, which is not possible.

## Solution

Removed the quick submit functionality from the table dropdown menu and changed the "Submit for Approval" action to navigate to the detail page instead. This ensures:

1. Users always go through the proper workflow selection dialog
2. All submissions include the required `workflowId` parameter
3. Consistent user experience across all document types

## Changes Made

### File: `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx`

**Removed:**

- Import of `useSubmitRequisitionForApproval` hook
- `showSubmitModal` state
- `handleSubmitForApproval` function
- Submit confirmation modal component
- Quick submit logic from dropdown menu

**Modified:**

- "Submit for Approval" dropdown action now navigates to detail page:
  ```typescript
  {canSubmit && (
    <DropdownMenuItem
      onClick={() => router.push(`/requisitions/${req.id}`)}
    >
      <Send className="mr-2 h-4 w-4 text-blue-600" />
      Submit for Approval
    </DropdownMenuItem>
  )}
  ```

## Submission Flow

### Current (Correct) Flow:

1. User clicks "Submit for Approval" in table dropdown
2. Navigates to requisition detail page
3. Clicks "Submit for Approval" button on detail page
4. `RequisitionSubmitDialog` opens with `WorkflowSelector`
5. User selects workflow and adds optional comments
6. Submission includes `workflowId` parameter
7. Backend processes submission successfully

### Components Involved:

- **RequisitionsTable**: Provides navigation to detail page
- **RequisitionDetailClient**: Shows submit button and manages dialog
- **RequisitionSubmitDialog**: Handles workflow selection and submission
- **WorkflowSelector**: Allows user to choose approval workflow
- **useSubmitRequisitionForApproval**: Hook that sends request with `workflowId`

## Verification

### TypeScript Compilation

✅ No diagnostics found in `requisitions-table.tsx`

### User Experience

✅ Consistent with other document types (Budget, PO, PV, GRN)
✅ Forces workflow selection before submission
✅ Provides clear submission summary and validation

## Related Files

- `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx` - Fixed
- `frontend/src/app/(private)/(main)/requisitions/_components/requisition-detail-client.tsx` - Verified
- `frontend/src/app/(private)/(main)/requisitions/_components/requisition-submit-dialog.tsx` - Verified
- `frontend/src/hooks/use-requisition-queries.ts` - Verified

## Status

✅ **COMPLETE** - Fix implemented and verified
