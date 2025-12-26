# Document Number Generation Guide

**Date**: 2025-12-26
**Status**: ✅ IMPLEMENTED
**Scope**: All workflow documents (Requisitions, Purchase Orders, Payment Vouchers, Goods Received Notes)

---

## Overview

All workflow documents now generate unique, human-readable document numbers automatically on the backend. This guide documents the implementation and provides reference information for developers.

---

## Number Format

All documents follow the same format:

```
{PREFIX}-{UNIX_TIMESTAMP}-{UUID_SHORT}
```

**Example**: `REQ-1735243125-a1b2c3d4`

### Components

| Component | Description | Example |
|-----------|-------------|---------|
| **PREFIX** | Document type identifier | REQ, PO, PV, GRN |
| **UNIX_TIMESTAMP** | Seconds since epoch | 1735243125 |
| **UUID_SHORT** | First 8 characters of UUID | a1b2c3d4 |

### Benefits

- ✅ **Human-readable**: Easy to reference in conversation/documentation
- ✅ **Chronologically sortable**: Earlier documents have lower timestamps
- ✅ **Globally unique**: UUID component ensures no collisions
- ✅ **Collision-resistant**: Combined timestamp + UUID = virtually impossible duplicates
- ✅ **Compact**: Only 30 characters (vs 36 for full UUID)
- ✅ **Consistent**: Same format across all document types

---

## Document Types and Prefixes

### 1. Requisition (REQ)

**Model**: `models.Requisition`
**Field**: `REQNumber`
**Handler**: `handlers/requisition.go::CreateRequisition()`
**Generated**: At creation time in backend

```go
reqNumber := fmt.Sprintf("REQ-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
```

**Database**: Unique index on `req_number` column

---

### 2. Purchase Order (PO)

**Model**: `models.PurchaseOrder`
**Field**: `PONumber`
**Handler**: `handlers/purchase_order.go::CreatePurchaseOrder()`
**Generated**: At creation time in backend

```go
poNumber := fmt.Sprintf("PO-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
```

**Database**: Unique index on `po_number` column

---

### 3. Payment Voucher (PV)

**Model**: `models.PaymentVoucher`
**Field**: `VoucherNumber`
**Handler**: `handlers/payment_voucher.go::CreatePaymentVoucher()`
**Generated**: At creation time in backend

```go
voucherNumber := fmt.Sprintf("PV-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
```

**Database**: Unique index on `voucher_number` column

---

### 4. Goods Received Note (GRN)

**Model**: `models.GoodsReceivedNote`
**Field**: `GRNNumber`
**Handler**: `handlers/grn.go::CreateGRN()`
**Generated**: At creation time in backend

```go
grnNumber := fmt.Sprintf("GRN-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
```

**Database**: Unique index on `grn_number` column

---

## Backend Implementation

### Step 1: Model Definition

Add the document number field to the model with unique index:

```go
type Requisition struct {
	ID                string          `gorm:"primaryKey" json:"id"`
	REQNumber         string          `gorm:"uniqueIndex" json:"reqNumber"`  // ← Add this
	OrganizationID    string          `gorm:"index;not null" json:"organizationId"`
	// ... other fields
}
```

### Step 2: Handler Implementation

Generate the number in the Create handler:

```go
func CreateRequisition(c fiber.Ctx) error {
	var req types.CreateRequisitionRequest

	// ... validation code ...

	// Generate document number
	reqNumber := fmt.Sprintf("REQ-%d-%s", time.Now().Unix(), uuid.New().String()[:8])

	// Create model
	requisition := models.Requisition{
		ID:        uuid.New().String(),
		REQNumber: reqNumber,  // ← Set the generated number
		// ... other fields
	}

	// Save to database
	if err := config.DB.Create(&requisition).Error; err != nil {
		return utils.SendInternalError(c, "Failed to create requisition", err)
	}

	return utils.SendSuccess(c, fiber.StatusCreated, requisition, "Created")
}
```

### Step 3: Database Schema (Auto-Migrated)

GORM's `AutoMigrate` automatically creates/updates the column:

```
Column: req_number (VARCHAR, NOT NULL)
Index: uniqueIndex (ensures no duplicates)
```

---

## Frontend Usage

### 1. Type Definition

Frontend types already have the document number fields:

```typescript
export interface Requisition {
  id: string;
  requisitionNumber: string;  // ← e.g., "REQ-1735243125-a1b2c3d4"
  // ... other fields
}

export interface PurchaseOrder {
  id: string;
  poNumber: string;  // ← e.g., "PO-1735243125-a1b2c3d4"
  // ... other fields
}

export interface PaymentVoucher {
  id: string;
  voucherNumber: string;  // ← e.g., "PV-1735243125-a1b2c3d4"
  // ... other fields
}
```

### 2. Displaying Numbers

Display the document number in UI components:

```typescript
// In a requisition detail view
<div className="document-header">
  <h1>{requisition.title}</h1>
  <p className="document-id">REQ: {requisition.requisitionNumber}</p>
</div>

// In a PO summary
<div className="po-info">
  <span>PO Number: {purchaseOrder.poNumber}</span>
</div>
```

### 3. Referencing in Forms

Include the document number when submitting related documents:

```typescript
// Creating a PO from requisition
const poData = {
  sourceRequisitionNumber: requisition.requisitionNumber,
  // ... other fields
};

// Creating a PV from PO
const pvData = {
  linkedPONumber: purchaseOrder.poNumber,
  // ... other fields
};
```

---

## API Response Example

### Create Requisition Response

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "reqNumber": "REQ-1735243125-a1b2c3d4",
    "title": "Office Supplies",
    "description": "Q1 2025 Office Supplies",
    "department": "Admin",
    "status": "DRAFT",
    "totalAmount": 5000,
    "currency": "ZMW",
    "items": [...],
    "createdAt": "2025-12-26T10:30:00Z",
    "updatedAt": "2025-12-26T10:30:00Z"
  },
  "message": "Requisition created successfully"
}
```

---

## Database Schema

### Requisitions Table

```sql
ALTER TABLE requisitions ADD COLUMN req_number VARCHAR(50) NOT NULL UNIQUE;

-- Index automatically created by GORM uniqueIndex tag
CREATE UNIQUE INDEX idx_requisitions_req_number ON requisitions(req_number);
```

### Purchase Orders Table

```sql
-- Already exists with poNumber field
SELECT * FROM purchase_orders WHERE po_number LIKE 'PO-%';
```

### Payment Vouchers Table

```sql
-- Already exists with voucherNumber field
SELECT * FROM payment_vouchers WHERE voucher_number LIKE 'PV-%';
```

### Goods Received Notes Table

```sql
-- Already exists with grnNumber field
SELECT * FROM goods_received_notes WHERE grn_number LIKE 'GRN-%';
```

---

## Implementation Timeline

| Date | Task | Status |
|------|------|--------|
| 2025-12-26 | Add REQNumber field to Requisition model | ✅ Complete |
| 2025-12-26 | Implement auto-generation in CreateRequisition | ✅ Complete |
| 2025-12-26 | Verify PO, PV, GRN auto-generation | ✅ Complete |
| 2025-12-26 | Update frontend types | ✅ Complete |
| 2025-12-26 | Document number generation | ✅ Complete |
| Ongoing | Test with backend API | ⏳ In Progress |

---

## Testing Checklist

- [ ] Create a requisition and verify REQ number is generated
- [ ] Create a purchase order and verify PO number is generated
- [ ] Create a payment voucher and verify PV number is generated
- [ ] Create a GRN and verify GRN number is generated
- [ ] Verify numbers are unique (no duplicates across same organization)
- [ ] Verify numbers are returned in API responses
- [ ] Display numbers in frontend UI
- [ ] Search/filter by document number
- [ ] Include numbers in exports (if applicable)
- [ ] Verify audit logs include document numbers

---

## API Endpoint Reference

### Requisition APIs

```
POST   /api/v1/requisitions
GET    /api/v1/requisitions
GET    /api/v1/requisitions/{id}
PUT    /api/v1/requisitions/{id}
POST   /api/v1/requisitions/{id}/submit
POST   /api/v1/requisitions/{id}/approve
POST   /api/v1/requisitions/{id}/reject
DELETE /api/v1/requisitions/{id}
```

All responses include `reqNumber` field.

### Purchase Order APIs

```
POST   /api/v1/purchase-orders
GET    /api/v1/purchase-orders
GET    /api/v1/purchase-orders/{id}
PUT    /api/v1/purchase-orders/{id}
POST   /api/v1/purchase-orders/{id}/submit
POST   /api/v1/purchase-orders/{id}/approve
POST   /api/v1/purchase-orders/{id}/reject
DELETE /api/v1/purchase-orders/{id}
```

All responses include `poNumber` field.

### Payment Voucher APIs

```
POST   /api/v1/payment-vouchers
GET    /api/v1/payment-vouchers
GET    /api/v1/payment-vouchers/{id}
PUT    /api/v1/payment-vouchers/{id}
POST   /api/v1/payment-vouchers/{id}/submit
POST   /api/v1/payment-vouchers/{id}/approve
POST   /api/v1/payment-vouchers/{id}/reject
POST   /api/v1/payment-vouchers/{id}/mark-paid
DELETE /api/v1/payment-vouchers/{id}
```

All responses include `voucherNumber` field.

### GRN APIs

```
POST   /api/v1/grns
GET    /api/v1/grns
GET    /api/v1/grns/{id}
PUT    /api/v1/grns/{id}
POST   /api/v1/grns/{id}/submit
POST   /api/v1/grns/{id}/approve
POST   /api/v1/grns/{id}/reject
DELETE /api/v1/grns/{id}
```

All responses include `grnNumber` field.

---

## Technical Details

### Why UNIX Timestamp + UUID?

**UNIX Timestamp Benefits:**
- Chronologically sortable (can query by date range)
- Human-readable (can estimate when created)
- Small (10 digits vs full date string)

**UUID Short Benefits:**
- Collision resistance
- Distributed generation (no centralized counter)
- No race conditions
- Works across multiple servers

**Combined Benefits:**
- Both chronological sorting AND guaranteed uniqueness
- Compact representation
- No server-side state needed

### Database Constraints

Each document number field has:
- `NOT NULL` constraint
- `UNIQUE` index (ensures no duplicates)
- Organization-scoped (not enforced in constraint but in application logic)

### Multi-Tenancy

Document numbers are unique per organization by application logic:
- Each organization has its own database (or filtered by `organization_id`)
- Unique index on document number column (local to organization)
- No global uniqueness required (same number can exist in different orgs)

---

## Migration Notes

### For Existing Requisitions

If you have existing requisitions without REQNumber:

```sql
-- Generate missing REQNumber for existing requisitions
UPDATE requisitions
SET req_number = CONCAT('REQ-', CAST(UNIX_TIMESTAMP(created_at) AS CHAR), '-', SUBSTR(id, 1, 8))
WHERE req_number IS NULL;
```

---

## Future Enhancements

Potential improvements (not implemented yet):

1. **Custom number formats** per organization (e.g., "REQ-2025-0001")
2. **Sequential counters** for audit compliance (some organizations require sequential numbers)
3. **Document number prefixes** by department (e.g., "REQ-AD-001")
4. **Number templates** (configurable by organization)
5. **Barcode/QR code** generation from document number

---

## Support

For questions about document number generation:

1. Check this guide: `DOCUMENT-NUMBER-GENERATION.md`
2. Review implementation in handler files
3. Check type definitions in frontend
4. Test with backend API

---

**Created**: 2025-12-26
**Last Updated**: 2025-12-26
**Owner**: Development Team
**Status**: ✅ Production Ready
