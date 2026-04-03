# Vendor Selection from Quotations - Implementation Complete

## Overview

Added the ability to select a vendor from the uploaded quotations before submitting a Purchase Order for approval. This allows users to review all vendor quotes and choose the best option, automatically updating the PO's vendor and total amount.

## Features

### 1. Visual Vendor Selection in Quotations Table

**Location**: Purchase Order detail page → Documents tab → Quotations section

**Features**:

- "Select" button next to each quotation (only shown in DRAFT status)
- Currently selected vendor is highlighted with green background
- "Selected" badge shown next to the active vendor
- "✓ Active" indicator in the action column for selected vendor
- One-click vendor switching

### 2. Automatic PO Updates

When a vendor is selected from quotations:

- **Vendor ID** is updated to match the selected quotation
- **Vendor Name** is updated automatically
- **Total Amount** is updated to match the quotation amount
- **Audit log** captures the vendor change with old/new values
- **Activity log** shows who changed the vendor and when

### 3. Smart UI Behavior

- Vendor selection only available when PO status is DRAFT
- Selection disabled during the update process (loading state)
- Success toast notification when vendor is selected
- Error toast if selection fails
- Page automatically refreshes with updated data

## Implementation Details

### Frontend Changes

#### 1. Enhanced QuotationCollectionSection Component

**File**: `frontend/src/app/(private)/(main)/requisitions/_components/quotation-collection-section.tsx`

**New Props**:

```typescript
interface QuotationCollectionSectionProps {
  // ... existing props
  selectedVendorId?: string;
  onSelectVendor?: (
    vendorId: string,
    vendorName: string,
    amount: number,
  ) => Promise<void>;
  showVendorSelection?: boolean;
}
```

**New Features**:

- `handleSelectQuotationVendor` function to handle vendor selection
- `selectingVendor` state for loading indication
- Enhanced table with "Action" column
- Visual highlighting for selected vendor
- "Select" buttons for non-selected vendors

**Table Structure**:

```
| Vendor | Amount | Date | Quote | Action |
|--------|--------|------|-------|--------|
| Vendor A | ZMW 5,000 | Jan 15 | [PDF] | [Select] |
| Vendor B | ZMW 4,500 | Jan 16 | [PDF] | ✓ Active |
| Vendor C | ZMW 5,200 | Jan 17 | [PDF] | [Select] |
```

#### 2. Updated PO Detail Client

**File**: `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-detail-client.tsx`

**New Handler**:

```typescript
const handleSelectVendor = async (
  vendorId: string,
  vendorName: string,
  amount: number,
) => {
  await updatePurchaseOrder({
    purchaseOrderId,
    poId: purchaseOrderId,
    vendorId,
    vendorName,
    totalAmount: amount,
  });
  handleDocumentUpdated();
};
```

**Updated QuotationCollectionSection Usage**:

```typescript
<QuotationCollectionSection
  quotations={quotations}
  requisitionId={purchaseOrderId}
  currency={purchaseOrder.currency || "ZMW"}
  vendors={vendors}
  canEdit={canEditQuotations}
  onSave={handleSaveQuotations}
  selectedVendorId={purchaseOrder.vendorId}
  onSelectVendor={handleSelectVendor}
  showVendorSelection={isDraft}
/>
```

#### 3. Updated Type Definitions

**File**: `frontend/src/types/purchase-order.ts`

**Added to UpdatePurchaseOrderRequest**:

- `vendorName?: string` - Vendor name for display

**File**: `frontend/src/app/_actions/purchase-orders.ts`

**Updated API call** to include:

- `vendorName: data.vendorName`

### Backend Changes

#### 1. Updated Type Definition

**File**: `backend/types/documents.go`

**Added to UpdatePurchaseOrderRequest**:

```go
type UpdatePurchaseOrderRequest struct {
    VendorID   string `json:"vendorId"`
    VendorName string `json:"vendorName"` // NEW
    // ... other fields
}
```

#### 2. Existing Handler Support

**File**: `backend/handlers/purchase_order.go`

The `UpdatePurchaseOrder` handler already supports:

- Updating `VendorID` field
- Tracking vendor changes in audit log
- Creating snapshots after updates
- Validating PO status (only DRAFT/PENDING can be updated)

**Audit Trail Capture**:

```go
if req.VendorID != "" {
    fromVendorID := ""
    if order.VendorID != nil {
        fromVendorID = *order.VendorID
    }
    if fromVendorID != req.VendorID {
        changes["vendorId"] = map[string]string{"old": fromVendorID, "new": req.VendorID}
    }
    order.VendorID = &req.VendorID
}
```

## User Workflow

### Scenario: Selecting a Vendor from Quotations

1. **User creates a Purchase Order** (manually or from approved requisition)
   - PO is in DRAFT status
   - Vendor may or may not be selected yet

2. **User uploads quotations** (or they come from requisition)
   - Upload 3+ quotations from different vendors
   - Each quotation includes vendor name, amount, and quote document

3. **User reviews quotations**
   - Navigate to "Documents" tab
   - Scroll to "Quotations" section
   - Review all vendor quotes side-by-side

4. **User selects preferred vendor**
   - Click "Select" button next to chosen quotation
   - System updates PO with:
     - Selected vendor ID
     - Selected vendor name
     - Quotation amount as total amount

5. **System confirms selection**
   - Success toast: "Selected [Vendor Name] as vendor"
   - Vendor row highlighted in green
   - "Selected" badge appears
   - PO details section updates with new vendor and amount

6. **Audit trail is captured**
   - Activity log shows: "Updated purchase order"
   - Changes show: vendor changed from X to Y
   - Snapshot captures complete PO state

7. **User submits for approval**
   - With selected vendor and amount
   - Quotation gate validation passes (3+ quotations)

## Benefits

### 1. Improved Decision Making

- Side-by-side comparison of all vendor quotes
- Easy switching between vendors
- Clear visual indication of selected vendor

### 2. Streamlined Workflow

- No need to manually enter vendor details
- Automatic amount updates
- One-click vendor selection

### 3. Complete Audit Trail

- Every vendor change is logged
- Old and new values captured
- Who made the change and when

### 4. Data Consistency

- Vendor information matches quotation exactly
- No manual entry errors
- Amount automatically synced

### 5. Flexibility

- Can change vendor selection before submission
- Can switch between vendors multiple times
- All changes are tracked

## Testing Checklist

- [x] Backend compiles without errors
- [ ] Quotations table shows "Action" column in DRAFT status
- [ ] "Select" button appears for non-selected vendors
- [ ] Currently selected vendor shows "✓ Active"
- [ ] Selected vendor row has green background
- [ ] Clicking "Select" updates the PO vendor
- [ ] PO total amount updates to match quotation
- [ ] Success toast appears after selection
- [ ] PO details section updates with new vendor
- [ ] Activity log shows vendor change
- [ ] Audit log captures old and new vendor IDs
- [ ] Snapshot is created after vendor change
- [ ] Can switch between vendors multiple times
- [ ] Selection only available in DRAFT status
- [ ] Selection disabled during update (loading state)

## How to Test

### 1. Create a Purchase Order

```
1. Navigate to Purchase Orders
2. Click "Create Purchase Order"
3. Fill in required fields
4. Leave vendor empty or select any vendor
5. Save as DRAFT
```

### 2. Add Quotations

```
1. Open the PO detail page
2. Go to "Documents" tab
3. Scroll to "Quotations" section
4. Click "Add Quotation"
5. Add 3 quotations from different vendors:
   - Vendor A: ZMW 5,000
   - Vendor B: ZMW 4,500
   - Vendor C: ZMW 5,200
6. Upload quote documents (optional)
```

### 3. Select a Vendor

```
1. Review the quotations table
2. Click "Select" next to Vendor B (lowest quote)
3. Verify:
   - Success toast appears
   - Vendor B row turns green
   - "Selected" badge appears
   - "✓ Active" shows in action column
   - PO details section shows Vendor B
   - Total amount shows ZMW 4,500
```

### 4. Switch Vendor

```
1. Click "Select" next to Vendor A
2. Verify:
   - Vendor A is now selected
   - Vendor B is no longer highlighted
   - Total amount updates to ZMW 5,000
```

### 5. Check Audit Trail

```
1. Go to "Activity Log" tab
2. Verify entries show:
   - "Updated purchase order"
   - Changes: vendorId from [old] to [new]
   - Changes: totalAmount from 4,500 to 5,000
   - Actor name and timestamp
```

### 6. Submit for Approval

```
1. Click "Submit for Approval"
2. Select workflow
3. Verify submission succeeds with selected vendor
```

## Files Modified

1. `frontend/src/app/(private)/(main)/requisitions/_components/quotation-collection-section.tsx`
   - Added vendor selection props and UI
   - Added "Select" buttons and visual indicators
   - Added `handleSelectQuotationVendor` handler

2. `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-detail-client.tsx`
   - Added `handleSelectVendor` function
   - Updated QuotationCollectionSection props

3. `frontend/src/types/purchase-order.ts`
   - Added `vendorName` to UpdatePurchaseOrderRequest

4. `frontend/src/app/_actions/purchase-orders.ts`
   - Added `vendorName` to API request payload

5. `backend/types/documents.go`
   - Added `VendorName` field to UpdatePurchaseOrderRequest

## Related Features

- **Edit Purchase Order Modal**: Allows editing other PO fields
- **Quotation Upload**: Upload vendor quotations with documents
- **Audit Logging**: Complete transparency of all changes
- **Activity Log**: Visual display of document history

## Future Enhancements

1. **Comparison View**: Side-by-side detailed comparison of quotations
2. **Vendor Scoring**: Automatic scoring based on price, delivery time, etc.
3. **Recommendation**: AI-powered vendor recommendation
4. **Bulk Selection**: Select vendor for multiple POs at once
5. **Vendor History**: Show past performance of each vendor
6. **Price Negotiation**: Track negotiation history with vendors

## Notes

- Vendor selection is only available in DRAFT status
- Once submitted for approval, vendor cannot be changed via quotations
- Use the Edit modal to change vendor after submission (if allowed by status)
- All vendor changes are captured in audit trail
- VendorName is computed from Vendor relationship in backend
- Frontend sends vendorName for display purposes only
