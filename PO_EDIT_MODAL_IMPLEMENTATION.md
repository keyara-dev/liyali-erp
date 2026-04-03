# Purchase Order Edit Modal Implementation - Complete

## Summary

Successfully implemented the edit modal for Purchase Orders with comprehensive audit logging that captures snapshots and field-level changes for complete transparency.

## What Was Completed

### 1. Backend - Enhanced Audit Logging in UpdatePurchaseOrder Handler

**File**: `backend/handlers/purchase_order.go`

**Changes**:

- Enhanced the `UpdatePurchaseOrder` handler to create a complete snapshot after changes
- Updated audit logging to use both `Changes` and `Snapshot` fields
- Properly structured the audit event with:
  - `Changes`: Field-level changes with old/new values
  - `Snapshot`: Complete document state after changes
  - `Details`: Additional context (documentNumber, updateType)

**Code Added** (lines ~610-625):

```go
if len(changes) > 0 {
    // Create snapshot of current state after changes
    snapshot := services.CreateDocumentSnapshot(order)

    // Log the audit event with changes and snapshot for full transparency
    go services.LogDocumentEvent(config.DB, services.DocumentEvent{
        OrganizationID: orgID,
        DocumentID:     order.ID,
        DocumentType:   "purchase_order",
        UserID:         actorID,
        ActorName:      updateUser.Name,
        ActorRole:      actorRole,
        Action:         "updated",
        Changes:        changes,
        Snapshot:       snapshot,
        Details: map[string]interface{}{
            "documentNumber": order.DocumentNumber,
            "updateType":     "manual_edit",
        },
    })
}
```

### 2. Frontend - Edit Purchase Order Dialog

**File**: `frontend/src/app/(private)/(main)/purchase-orders/_components/edit-purchase-order-dialog.tsx`

**Features**:

- Modal dialog for editing purchase orders
- Only allows editing when PO status is DRAFT or REJECTED
- Shows warning message when PO cannot be edited
- Editable fields:
  - Title (required)
  - Description
  - Department (required, dropdown)
  - Priority (dropdown: LOW, MEDIUM, HIGH, URGENT)
  - Budget Code (dropdown from budgets)
  - Cost Center
  - Project Code
  - Delivery Date (date picker)
- Read-only fields (displayed but disabled):
  - Document Number
  - Vendor Name
  - Total Amount
  - Currency
  - Status
  - Items
- Form validation:
  - Title is required
  - Department is required
- Loading states during save
- Success/error toast notifications

### 3. Frontend - Integration with Detail Page

**File**: `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-detail-client.tsx`

**Integration**:

- Edit button in header (only shown when user has edit permission)
- Dialog state management via `usePurchaseOrderDetail` hook
- Automatic data refresh after successful edit
- Proper permission checks before showing edit button

### 4. Frontend - Update Mutation Hook

**File**: `frontend/src/hooks/use-purchase-order-mutations.ts`

**Added**:

- `useUpdatePurchaseOrder` hook for updating purchase orders
- Automatic query invalidation after successful update
- Toast notifications for success/error
- Invalidates:
  - Purchase order list and stats
  - Specific purchase order detail
  - Dashboard metrics and activities
  - Audit events for the document

### 5. Frontend - Type Definitions

**File**: `frontend/src/types/purchase-order.ts`

**Added**:

- `department` field to `UpdatePurchaseOrderRequest`
- `departmentId` field to `UpdatePurchaseOrderRequest`

### 6. Frontend - Update Action

**File**: `frontend/src/app/_actions/purchase-orders.ts`

**Updated**:

- Added `department` and `departmentId` to the API request payload
- Added `deliveryDate` to the API request payload

### 7. Bug Fix - Unused Variable

**File**: `backend/utils/audit_helper.go`

**Fix**: Removed unused `changesJSON` variable that was causing compilation error

### 8. Bug Fix - QueryClient SSR Error

**File**: `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-detail-client.tsx`

**Fix**:

- Removed `useQueryClient` hook call that was causing server-side rendering error
- Replaced direct `queryClient.invalidateQueries` calls with `handleDocumentUpdated` callback from the hook
- Removed unused imports: `useQueryClient` and `QUERY_KEYS`
- This fixes the "No QueryClient set" error during server rendering

## Audit Trail Features

The implementation provides complete transparency by capturing:

1. **What Changed**: Field-level changes with before/after values

   ```json
   {
     "title": {
       "old": "Office Supplies Request",
       "new": "Updated Office Supplies Request"
     },
     "priority": {
       "old": "MEDIUM",
       "new": "HIGH"
     }
   }
   ```

2. **Who Changed It**: User information

   ```json
   {
     "userId": "user-123",
     "actorName": "John Doe",
     "actorRole": "PROCUREMENT_OFFICER"
   }
   ```

3. **When It Changed**: Timestamp in audit log

4. **Complete Snapshot**: Full document state after change
   ```json
   {
     "snapshot": {
       "id": "po-456",
       "documentNumber": "PO-2025-001",
       "title": "Updated Office Supplies Request",
       "status": "DRAFT",
       "totalAmount": 5000.00,
       "currency": "ZMW",
       "vendorName": "Office Supplies Inc.",
       "items": [...],
       "snapshotTimestamp": "2025-04-03T10:30:00Z"
     }
   }
   ```

## Testing Checklist

- [x] Backend compiles without errors
- [ ] Edit button appears on PO detail page for authorized users
- [ ] Edit dialog opens when clicking Edit button
- [ ] Form is pre-populated with current PO data
- [ ] Only editable fields are enabled
- [ ] Read-only fields are properly disabled
- [ ] Warning message shows for non-DRAFT/REJECTED POs
- [ ] Form validation works (title and department required)
- [ ] Save button is disabled when form is invalid
- [ ] Changes are saved to database
- [ ] Audit log entry is created with changes and snapshot
- [ ] Activity log tab shows the update with field changes
- [ ] Page refreshes with updated data after save
- [ ] Toast notifications appear for success/error

## How to Test

1. **Open a Purchase Order in DRAFT status**:
   - Navigate to Purchase Orders list
   - Click on a PO with DRAFT status
   - Verify Edit button appears in header

2. **Open Edit Dialog**:
   - Click Edit button
   - Verify dialog opens with current PO data
   - Verify editable fields are enabled
   - Verify read-only fields are disabled

3. **Make Changes**:
   - Change title to "Updated PO Title"
   - Change priority from MEDIUM to HIGH
   - Change description
   - Add/update budget code, cost center, project code
   - Change delivery date

4. **Save Changes**:
   - Click "Save Changes" button
   - Verify loading state appears
   - Verify success toast notification
   - Verify dialog closes
   - Verify page refreshes with updated data

5. **Verify Audit Log**:
   - Click on "Activity Log" tab
   - Verify new entry shows "updated" action
   - Verify entry shows field changes with old/new values
   - Verify entry shows who made the change and when
   - Verify snapshot is captured (check database or API response)

6. **Test with Non-Editable Status**:
   - Open a PO with APPROVED or PENDING status
   - Click Edit button
   - Verify warning message appears
   - Verify all fields are disabled
   - Verify Save button is disabled

## Files Modified

1. `backend/handlers/purchase_order.go` - Enhanced audit logging with snapshot
2. `backend/utils/audit_helper.go` - Fixed unused variable
3. `frontend/src/app/(private)/(main)/purchase-orders/_components/edit-purchase-order-dialog.tsx` - Created edit modal
4. `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-detail-client.tsx` - Already integrated (from previous work)
5. `frontend/src/hooks/use-purchase-order-mutations.ts` - Added useUpdatePurchaseOrder hook
6. `frontend/src/types/purchase-order.ts` - Added department and departmentId fields to UpdatePurchaseOrderRequest
7. `frontend/src/app/_actions/purchase-orders.ts` - Updated updatePurchaseOrder action to include department fields

## Next Steps

1. Test the edit modal functionality end-to-end
2. Verify audit logs are created correctly with snapshots
3. Test with different user roles and permissions
4. Test with different PO statuses (DRAFT, PENDING, APPROVED, REJECTED)
5. Verify activity log displays changes correctly
6. Consider adding similar edit modals for other document types (REQ, PV, GRN)

## Related Documentation

- `backend/AUDIT_SNAPSHOT_IMPLEMENTATION.md` - Comprehensive guide on audit logging with snapshots
- `backend/AUDIT_LOGGING_IMPLEMENTATION.md` - Audit logging implementation details
- `AUDIT_TRAIL_TRANSPARENCY_SUMMARY.md` - Overview of audit trail features
- `AUDIT_TRAIL_VISUAL_GUIDE.md` - Visual guide for audit trail display
