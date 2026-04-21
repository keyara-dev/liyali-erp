# PO Submission Fix - Verification Steps

## What Was Fixed

### 1. **Frontend Hook** (`use-purchase-order-detail.ts`)

- **Issue**: Was passing `purchaseOrderId` but action expected `poId`
- **Fix**: Now passes both `poId` (primary) and `purchaseOrderId` (backward compatibility)

### 2. **Frontend Action** (`purchase-orders.ts`)

- **Issue**: Would use `undefined` if `poId` was missing
- **Fix**: Added fallback logic and validation to support both field names

### 3. **Backend Handler** (`purchase_order.go`)

- **Issue**: Limited logging and no soft-delete filter
- **Fix**: Added comprehensive logging and soft-delete filter

---

## Testing Steps

### Step 1: Rebuild Frontend

```bash
cd frontend
npm run build
# or for development
npm run dev
```

### Step 2: Restart Backend

```bash
cd backend
# If using air for hot reload, it should auto-restart
# Otherwise:
go run main.go
```

### Step 3: Test PO Submission

1. **Navigate to a DRAFT Purchase Order**
   - Go to Purchase Orders list
   - Open any PO with status "DRAFT"
   - Or create a new PO and save as draft

2. **Open Browser DevTools**
   - Press F12
   - Go to "Network" tab
   - Filter by "Fetch/XHR"

3. **Click "Submit for Approval"**
   - Select a workflow
   - Add optional comments
   - Click Submit

4. **Verify the Request**
   - Check Network tab for the POST request
   - URL should be: `/api/v1/purchase-orders/{valid-uuid}/submit`
   - NOT: `/api/v1/purchase-orders/undefined/submit`

5. **Check Response**
   - Should return 200 OK with success message
   - PO status should change to "PENDING"

---

## Database Verification

If you still encounter issues, run these SQL queries:

### Check if PO exists

```sql
SELECT
    id,
    organization_id,
    status,
    deleted_at,
    document_number,
    created_by
FROM purchase_orders
WHERE document_number = 'PO-XXXX-XXX'  -- Replace with your PO number
ORDER BY created_at DESC
LIMIT 1;
```

### Check user's organization

```sql
SELECT
    u.id as user_id,
    u.email,
    u.current_organization_id,
    om.organization_id as member_org_id,
    om.active
FROM users u
LEFT JOIN organization_members om ON om.user_id = u.id
WHERE u.email = 'your-email@example.com';  -- Replace with your email
```

### Check for org mismatch

```sql
SELECT
    po.id as po_id,
    po.document_number,
    po.organization_id as po_org,
    po.status,
    po.deleted_at,
    om.organization_id as user_org,
    om.active as user_active
FROM purchase_orders po
CROSS JOIN organization_members om
WHERE po.document_number = 'PO-XXXX-XXX'  -- Replace with your PO number
  AND om.user_id = 'your-user-id'  -- Replace with your user ID
  AND om.active = true;
```

---

## Backend Log Verification

### Check logs for the submission attempt

```bash
# If using structured logging
cd backend
tail -f logs/app.log | grep "submit_purchase_order"

# Look for these log entries:
# - submit_purchase_order_request (entry point)
# - order_id, organization_id, user_id (context)
# - purchase_order_not_found (if error occurs)
```

### Expected log output (SUCCESS):

```json
{
  "level": "info",
  "msg": "submit_purchase_order_request",
  "operation": "submit_purchase_order",
  "order_id": "abc-123-def",
  "organization_id": "org-456",
  "user_id": "user-789"
}
```

### Expected log output (ERROR - if still failing):

```json
{
  "level": "error",
  "msg": "purchase_order_not_found",
  "order_id": "abc-123-def",
  "organization_id": "org-456",
  "user_id": "user-789",
  "error_detail": "record not found"
}
```

---

## Common Issues & Solutions

### Issue 1: Still getting "undefined" in URL

**Cause**: Frontend not rebuilt  
**Solution**:

```bash
cd frontend
rm -rf .next
npm run build
```

### Issue 2: PO not found but exists in database

**Cause**: Organization mismatch  
**Solution**:

- Check user's `X-Organization-ID` header in browser DevTools
- Verify it matches the PO's `organization_id` in database
- User may need to switch organizations

### Issue 3: PO was soft-deleted

**Cause**: `deleted_at` is not NULL  
**Solution**:

```sql
-- Check if soft-deleted
SELECT id, deleted_at FROM purchase_orders WHERE id = 'your-po-id';

-- Restore if needed (admin only)
UPDATE purchase_orders SET deleted_at = NULL WHERE id = 'your-po-id';
```

### Issue 4: Permission denied

**Cause**: User doesn't have "purchase_order:edit" permission  
**Solution**:

- Check user's role permissions
- Verify RBAC configuration
- User must be creator or have appropriate role

---

## Success Indicators

✅ **Frontend**:

- Network request shows valid UUID in URL
- Request payload includes `workflowId`
- Response is 200 OK with success message

✅ **Backend**:

- Logs show all context fields (order_id, organization_id, user_id)
- No "purchase_order_not_found" error
- PO status updated to "PENDING"

✅ **Database**:

- PO status changed from "DRAFT" to "PENDING"
- `action_history` includes "SUBMIT" entry
- Workflow assignment created

✅ **UI**:

- Success toast notification appears
- PO detail page refreshes
- Status badge shows "PENDING"
- Submit button disappears
- Withdraw button appears

---

## Rollback Plan

If the fix causes issues, revert these files:

```bash
git checkout HEAD -- frontend/src/hooks/use-purchase-order-detail.ts
git checkout HEAD -- frontend/src/app/_actions/purchase-orders.ts
git checkout HEAD -- backend/handlers/purchase_order.go
```

---

## Additional Debugging

### Enable verbose logging in frontend

Add to `frontend/src/app/_actions/purchase-orders.ts`:

```typescript
console.log("[DEBUG] Submit PO Request:", {
  poId: data.poId,
  purchaseOrderId: data.purchaseOrderId,
  workflowId: data.workflowId,
  url: `/api/v1/purchase-orders/${poId}/submit`,
});
```

### Enable verbose logging in backend

Add to `backend/handlers/purchase_order.go`:

```go
log.Printf("[DEBUG] Submit PO: id=%s, orgID=%s, userID=%s", id, organizationID, userID)
```

---

## Contact & Support

If issues persist after applying these fixes:

1. Capture the following information:
   - Browser console logs (with Network tab)
   - Backend logs (last 50 lines)
   - Database query results from above
   - Screenshot of the error

2. Check the audit report: `PO_SUBMIT_AUDIT_REPORT.md`

3. Review the complete flow diagram in the audit report
