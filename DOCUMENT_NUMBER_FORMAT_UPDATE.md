# Document Number Format Update

## ✅ **Updated Document Number Format**

### **New Format**: `{PREFIX}-{6CHARS_TIMESTAMP}-{4CHARS_UUID}`

**Examples:**
- **Requisitions**: `REQ-260108-2D7C`
- **Purchase Orders**: `PO-260108-F5D0`
- **Payment Vouchers**: `PV-260108-507C`
- **Goods Received Notes**: `GRN-260108-DD75`
- **Vendor Codes**: `VND-260108-F124`

### **Format Breakdown:**
1. **PREFIX**: Document type in uppercase (REQ, PO, PV, GRN, VND)
2. **6 CHARS TIMESTAMP**: YYMMDD format (Year Month Day)
   - `26` = Year 2026
   - `01` = January
   - `08` = 8th day
3. **4 CHARS UUID**: First 4 characters of UUID in uppercase
   - Ensures uniqueness even for documents created at the same time
   - Always uppercase for consistency

### **Previous Format**: `{PREFIX}-{UNIX_TIMESTAMP}-{8CHARS_UUID}`
- Example: `REQ-1736347852-a1b2c3d4`
- Used Unix timestamp (seconds since epoch)
- Used 8 characters from UUID

### **Benefits of New Format:**
1. **Shorter**: More compact and readable
2. **Human-Readable**: Date is immediately recognizable (YYMMDD)
3. **Consistent**: All document types follow same pattern
4. **Unique**: 4-char UUID ensures uniqueness
5. **Uppercase**: All characters are uppercase for consistency

## 🔧 **Implementation Details**

### **Files Updated:**
1. **Created**: `backend/utils/document_numbers.go` - Centralized document number generation
2. **Updated**: `backend/handlers/requisition.go` - Uses new requisition number generator
3. **Updated**: `backend/handlers/purchase_order.go` - Uses new PO number generator
4. **Updated**: `backend/handlers/payment_voucher.go` - Uses new PV number generator
5. **Updated**: `backend/handlers/grn.go` - Uses new GRN number generator
6. **Updated**: `backend/handlers/vendor.go` - Uses new vendor code generator

### **Utility Functions:**
```go
utils.GenerateRequisitionNumber()      // REQ-260108-2D7C
utils.GeneratePurchaseOrderNumber()    // PO-260108-F5D0
utils.GeneratePaymentVoucherNumber()   // PV-260108-507C
utils.GenerateGRNNumber()              // GRN-260108-DD75
utils.GenerateVendorCode()             // VND-260108-F124
```

### **Backward Compatibility:**
- Legacy function available: `utils.GenerateDocumentNumberLegacy(prefix)`
- Existing documents in database retain their original format
- New documents will use the updated format

## 🚀 **Status**

- ✅ **Backend Updated**: All handlers use new document number generation
- ✅ **Compilation Successful**: No build errors
- ✅ **Server Running**: Backend server restarted with new changes
- ✅ **Format Tested**: Confirmed new format works correctly

### **Next Document Numbers Will Be:**
- **Requisitions**: `REQ-260108-XXXX` (where XXXX is 4-char UUID)
- **Purchase Orders**: `PO-260108-XXXX`
- **Payment Vouchers**: `PV-260108-XXXX`
- **GRNs**: `GRN-260108-XXXX`
- **Vendor Codes**: `VND-260108-XXXX`

The document number format has been successfully updated as requested! 🎉