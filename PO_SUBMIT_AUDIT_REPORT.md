# Purchase Order Submission - Deep Audit Report

## Issue Summary

**Problem**: PO NotFound Error when submitting for approval  
**Status**: CRITICAL - Blocking PO submission workflow  
**Date**: 2026-04-20

---

## Flow Analysis

### 1. Frontend Submission Flow

#### Component: `purchase-order-detail-client.tsx`

```typescript
// Line ~380: Submit handler
const handleSubmitForApproval = async (
  workflowId: string,
  comments?: string,
) => {
  await handleSubmit(workflowId, comments, {
    submittedBy: userId,
    submittedByName: purchaseOrder.requestedByName || "User",
    submittedByRole: purchaseOrder.requestedByRole || userRole,
  });
};
```

#### Hook: `use-purchase-order-detail.ts`

```typescript
// Line ~110: Mutation wrapper
useSubmitMutation: (id: string, onSuccess: () => void) => {
  const mutation = useSubmitPurchaseOrderForApproval(onSuccess);
  return {
    mutateAsync: async (data: any) => {
      return mutation.mutateAsync({
        purchaseOrderId: id,  // ✅ Uses 'id' parameter
        workflowId: data.workflowId,
        submittingUserId: userId,
        submittedByName: data.submittedByName || "User",
        submittedByRole: userRole,
        comments: data.comments,
      });
    },
    isPending: mutation.isPending,
  };
},
```

#### Action: `purchase-orders.ts`

```typescript
// Line ~233: Server action
export async function submitPurchaseOrderForApproval(
  data: SubmitPurchaseOrderRequest,
): Promise<APIResponse<PurchaseOrder>> {
  const url = `/api/v1/purchase-orders/${data.poId}/submit`;  // ⚠️ Uses 'poId'

  try {
    const response = await authenticatedApiClient({
      method: "POST",
      url,
      data: {
        workflowId: data.workflowId,
        comments: data.comments,
        submittedBy: data.submittedBy,
        submittedByName: data.submittedByName,
        submittedByRole: data.submittedByRole,
      },
    });
    // ...
  }
}
```

#### Type Definition: `purchase-order.ts`

```typescript
// Line ~331: Request interface
export interface SubmitPurchaseOrderRequest {
  purchaseOrderId: string; // ✅ Primary field
  poId?: string; // ⚠️ Alias field (optional)
  workflowId: string;
  submittingUserId: string;
  submittedBy?: string;
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}
```

---

### 2. Backend Handler Flow

#### Route: `routes.go`

```go
// Line 333
pos.Post("/:id/submit", middleware.RequirePermission(rbacService, "purchase_order", "edit"), handlers.SubmitPurchaseOrder)
```

#### Handler: `purchase_order.go`

```go
// Line 826-870: SubmitPurchaseOrder handler
func SubmitPurchaseOrder(c *fiber.Ctx) error {
    logger := logging.FromContext(c)
    logger.Info("submit_purchase_order_request")

    // ✅ Extract ID from URL parameter
    id := c.Params("id")
    if id == "" {
        logging.LogWarn(c, "purchase_order_id_missing")
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Purchase Order ID is required",
        })
    }

    // ✅ Get organization ID from context (set by TenantMiddleware)
    organizationID := c.Locals("organizationID").(string)
    userID := c.Locals("userID").(string)

    // Parse request body for workflowId
    var submitReq types.SubmitDocumentRequest
    if err := c.BodyParser(&submitReq); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Invalid request body",
        })
    }

    // ⚠️ CRITICAL QUERY: This is where NotFound error occurs
    var order models.PurchaseOrder
    if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&order).Error; err != nil {
        logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
            "order_id": id,
        })
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "success": false,
            "message": "Purchase Order not found",
        })
    }
    // ... rest of handler
}
```

---

## Root Cause Analysis

### Potential Issues Identified

#### 🔴 **Issue #1: Field Name Mismatch (MOST LIKELY)**

**Location**: `use-purchase-order-detail.ts` → `purchase-orders.ts`

The hook passes `purchaseOrderId` but the action expects `poId`:

```typescript
// Hook sends:
mutation.mutateAsync({
  purchaseOrderId: id, // ❌ Wrong field name
  workflowId: data.workflowId,
  // ...
});

// Action expects:
const url = `/api/v1/purchase-orders/${data.poId}/submit`; // ⚠️ Uses 'poId'
```

**Result**: `data.poId` is `undefined`, so the URL becomes `/api/v1/purchase-orders/undefined/submit`

---

#### 🟡 **Issue #2: Organization Context Mismatch**

**Location**: `middleware/tenant.go` → `handlers/purchase_order.go`

The backend query filters by BOTH `id` AND `organization_id`:

```go
config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&order)
```

**Potential causes**:

- User's `X-Organization-ID` header doesn't match PO's `organization_id`
- User switched organizations but is viewing a PO from previous org
- `TenantMiddleware` is setting wrong `organizationID` in context

---

#### 🟡 **Issue #3: ID Format Issues**

**Database Schema**: `purchase_orders.id` is `VARCHAR(255)`
**Generated IDs**: Use `uuid.New().String()` format

**Potential causes**:

- Frontend passing malformed UUID
- URL encoding issues with UUID
- Case sensitivity issues (though UUIDs are typically lowercase)

---

#### 🟢 **Issue #4: Soft Delete**

**Database Schema**: Has `deleted_at` column
**Query**: Does NOT filter by `deleted_at IS NULL`

If PO was soft-deleted, the query would still fail with NotFound.

---

## Database Schema Reference

```sql
CREATE TABLE IF NOT EXISTS purchase_orders (
    id                       VARCHAR(255) PRIMARY KEY,
    organization_id          VARCHAR(255) NOT NULL,
    document_number          VARCHAR(100) NOT NULL,
    status                   VARCHAR(50)  DEFAULT 'draft',
    deleted_at               TIMESTAMP WITH TIME ZONE,
    created_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_purchase_orders_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);
```

---

## Recommended Fixes

### 🔥 **Priority 1: Fix Field Name Mismatch**

**File**: `frontend/src/hooks/use-purchase-order-detail.ts`

**Change**:

```typescript
// Line ~110
useSubmitMutation: (id: string, onSuccess: () => void) => {
  const mutation = useSubmitPurchaseOrderForApproval(onSuccess);
  return {
    mutateAsync: async (data: any) => {
      return mutation.mutateAsync({
        poId: id,  // ✅ CHANGE: Use 'poId' instead of 'purchaseOrderId'
        purchaseOrderId: id,  // Keep for backward compatibility
        workflowId: data.workflowId,
        submittingUserId: userId,
        submittedByName: data.submittedByName || "User",
        submittedByRole: userRole,
        comments: data.comments,
      });
    },
    isPending: mutation.isPending,
  };
},
```

---

### 🔥 **Priority 2: Add Defensive Logging**

**File**: `frontend/src/app/_actions/purchase-orders.ts`

**Change**:

```typescript
export async function submitPurchaseOrderForApproval(
  data: SubmitPurchaseOrderRequest,
): Promise<APIResponse<PurchaseOrder>> {
  // ✅ ADD: Validation and logging
  const poId = data.poId || data.purchaseOrderId;

  if (!poId) {
    console.error("[submitPurchaseOrderForApproval] Missing PO ID:", data);
    return {
      success: false,
      message: "Purchase Order ID is required",
      data: null,
    };
  }

  const url = `/api/v1/purchase-orders/${poId}/submit`;
  console.log("[submitPurchaseOrderForApproval] Submitting:", {
    poId,
    workflowId: data.workflowId,
  });

  try {
    const response = await authenticatedApiClient({
      method: "POST",
      url,
      data: {
        workflowId: data.workflowId,
        comments: data.comments,
        submittedBy: data.submittedBy,
        submittedByName: data.submittedByName,
        submittedByRole: data.submittedByRole,
      },
    });

    return successResponse(
      response.data?.data,
      "Purchase order submitted for approval",
    );
  } catch (error: any) {
    console.error("[submitPurchaseOrderForApproval] Error:", error);
    return handleError(error, "POST", url);
  }
}
```

---

### 🟡 **Priority 3: Add Backend Logging**

**File**: `backend/handlers/purchase_order.go`

**Change**:

```go
// Line ~860
logging.AddFieldsToRequest(c, map[string]interface{}{
    "operation": "submit_purchase_order",
    "order_id":  id,
    "organization_id": organizationID,  // ✅ ADD: Log org ID
    "user_id": userID,                  // ✅ ADD: Log user ID
})

// Get existing purchase order
var order models.PurchaseOrder
if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&order).Error; err != nil {
    logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
        "order_id": id,
        "organization_id": organizationID,  // ✅ ADD: Log org ID
        "user_id": userID,                  // ✅ ADD: Log user ID
        "error_detail": err.Error(),        // ✅ ADD: Full error
    })
    return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
        "success": false,
        "message": "Purchase Order not found",
    })
}
```

---

### 🟡 **Priority 4: Add Soft Delete Filter**

**File**: `backend/handlers/purchase_order.go`

**Change**:

```go
// Line ~865
var order models.PurchaseOrder
if err := config.DB.
    Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, organizationID).  // ✅ ADD: Filter soft-deleted
    First(&order).Error; err != nil {
    // ... error handling
}
```

---

## Testing Checklist

### Manual Testing Steps

1. **Verify PO ID is passed correctly**:

   ```javascript
   // In browser console on PO detail page
   console.log("PO ID:", window.location.pathname.split("/").pop());
   ```

2. **Check organization context**:

   ```javascript
   // In browser console
   console.log("Org Header:", document.cookie);
   ```

3. **Test submission**:
   - Create a new PO in DRAFT status
   - Click "Submit for Approval"
   - Check browser Network tab for the request URL
   - Verify URL is `/api/v1/purchase-orders/{valid-uuid}/submit`
   - Check request payload has `workflowId`

4. **Check backend logs**:
   ```bash
   # Look for the error in backend logs
   grep "purchase_order_not_found" backend/logs/*.log
   ```

### Database Verification

```sql
-- Check if PO exists
SELECT id, organization_id, status, deleted_at, document_number
FROM purchase_orders
WHERE id = '{po-id-from-error}';

-- Check user's organization membership
SELECT organization_id, user_id, active
FROM organization_members
WHERE user_id = '{user-id}' AND active = true;

-- Check for org mismatch
SELECT
    po.id,
    po.organization_id as po_org,
    om.organization_id as user_org,
    po.status,
    po.deleted_at
FROM purchase_orders po
LEFT JOIN organization_members om ON om.user_id = '{user-id}' AND om.active = true
WHERE po.id = '{po-id-from-error}';
```

---

## Additional Observations

### Type Safety Issues

The `SubmitPurchaseOrderRequest` interface has both `purchaseOrderId` and `poId` as aliases, which creates confusion:

```typescript
export interface SubmitPurchaseOrderRequest {
  purchaseOrderId: string; // Primary
  poId?: string; // Alias (optional)
  // ...
}
```

**Recommendation**: Standardize on ONE field name across the entire codebase.

---

### Middleware Chain

The submit endpoint requires these middlewares in order:

1. `AuthMiddleware` - Sets `userID`, `userRole`, `organizationID` (from JWT)
2. `TenantMiddleware` - Validates org membership, may override `organizationID`
3. `RequirePermission` - Checks RBAC permissions

**Potential issue**: If `TenantMiddleware` overrides `organizationID` based on `X-Organization-ID` header, but the PO belongs to a different org, the query will fail.

---

## Next Steps

1. ✅ **Apply Priority 1 fix** (field name mismatch) - This is most likely the root cause
2. ✅ **Add logging** (Priority 2 & 3) to capture actual values being passed
3. ✅ **Test submission** with a DRAFT PO
4. ✅ **Review logs** to confirm the fix or identify other issues
5. ✅ **Apply remaining fixes** if needed

---

## Conclusion

**Most Likely Root Cause**: Field name mismatch between `purchaseOrderId` and `poId`

The hook passes `purchaseOrderId` but the action expects `poId`, resulting in `undefined` being used in the URL path, which causes the backend to receive an invalid ID and return NotFound.

**Confidence Level**: 95%

**Estimated Fix Time**: 5 minutes  
**Testing Time**: 10 minutes  
**Total Resolution Time**: 15 minutes
