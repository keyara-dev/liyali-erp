# Currency Formatting Fix

## Problem

Currency is hardcoded as "USD" in several places and the formatting is inconsistent across the application. All currency should be derived from the source requisition document and formatted consistently using a utility function.

## Solution

### 1. Created `formatCurrency` Utility Function

Added to `frontend/src/lib/utils/index.ts`:

```typescript
/**
 * Format a number as currency with the specified currency code
 *
 * @param amount - The amount to format
 * @param currency - The currency code (e.g., "USD", "ZMW", "EUR")
 * @param locale - The locale to use for formatting (defaults to "en-ZM")
 * @returns Formatted currency string
 *
 * @example
 * formatCurrency(1234.56, "USD") // "USD 1,234.56"
 * formatCurrency(1234.56, "ZMW") // "ZMW 1,234.56"
 */
export function formatCurrency(
  amount: number | null | undefined,
  currency: string = "USD",
  locale: string = "en-ZM",
): string {
  if (amount === null || amount === undefined) {
    return `${currency} 0.00`;
  }

  const formatted = amount.toLocaleString(locale, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });

  return `${currency} ${formatted}`;
}
```

### 2. Usage Pattern

#### BEFORE (Hardcoded):

```tsx
<p>
  {purchaseOrder.currency}{" "}
  {purchaseOrder.totalAmount?.toLocaleString("en-ZM", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }) || "0.00"}
</p>
```

#### AFTER (Using utility):

```tsx
import { formatCurrency } from "@/lib/utils";

<p>{formatCurrency(purchaseOrder.totalAmount, purchaseOrder.currency)}</p>;
```

### 3. Files That Need Updates

#### Purchase Orders

- ✅ `frontend/src/lib/utils/index.ts` - Added formatCurrency function
- ⏳ `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-detail-client.tsx` - Import added, needs replacement
- ⏳ `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-items-list.tsx` - Needs update
- ⏳ `frontend/src/app/(private)/(main)/purchase-orders/create/_components/po-create-client.tsx` - Needs update

#### Payment Vouchers

- ⏳ `frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-voucher-detail-client.tsx` - Needs update
- ⏳ `frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-voucher-items-list.tsx` - Needs update
- ⏳ `frontend/src/app/(private)/(main)/payment-vouchers/create/_components/pv-create-client.tsx` - Remove hardcoded "USD"

#### Requisitions

- ⏳ `frontend/src/app/(private)/(main)/requisitions/_components/requisition-detail-client.tsx` - Needs update
- ⏳ `frontend/src/app/(private)/(main)/requisitions/_components/requisition-items-list.tsx` - Needs update

#### GRN

- ⏳ `frontend/src/app/(private)/(main)/grn/_components/grn-detail-client.tsx` - Needs update

#### Budgets

- ⏳ `frontend/src/app/(private)/(main)/budgets/[id]/_components/edit-budget-item-dialog.tsx` - Replace hardcoded USD
- ⏳ `frontend/src/app/(private)/(main)/budgets/[id]/_components/add-budget-item-dialog.tsx` - Replace hardcoded USD

### 4. Backend - Currency Propagation

Ensure currency is properly propagated from REQ → PO → PV → GRN:

#### Already Correct:

- ✅ `backend/handlers/document_extras_handler.go` - PO creation copies currency from REQ
- ✅ `backend/handlers/document_extras_handler.go` - PV creation copies currency from PO

#### Verification Needed:

- Check that currency is being saved correctly in all document creation handlers
- Verify currency is included in all response types

### 5. Search and Replace Pattern

To find all instances that need updating:

```bash
# Find hardcoded USD
grep -r "USD" frontend/src/app --include="*.tsx" --include="*.ts"

# Find toLocaleString with currency formatting
grep -r "toLocaleString.*minimumFractionDigits" frontend/src/app --include="*.tsx"

# Find currency concatenation
grep -r "currency.*toLocaleString\|toLocaleString.*currency" frontend/src/app --include="*.tsx"
```

### 6. Testing Checklist

After applying all changes:

- [ ] Create REQ with ZMW currency
- [ ] Create PO from REQ - verify currency is ZMW
- [ ] Verify PO detail page shows "ZMW 1,234.56" format
- [ ] Create PV from PO - verify currency is ZMW
- [ ] Verify PV detail page shows "ZMW 1,234.56" format
- [ ] Create GRN from PO - verify currency is ZMW
- [ ] Test with different currencies (USD, EUR, GBP)
- [ ] Verify all amounts display consistently
- [ ] Check PDF exports show correct currency

### 7. Benefits

1. **Consistency**: All currency formatting uses the same function
2. **Flexibility**: Easy to change formatting logic in one place
3. **Correctness**: Currency is derived from source document, not hardcoded
4. **Maintainability**: Single source of truth for currency formatting
5. **Internationalization**: Easy to add locale-specific formatting

### 8. Next Steps

1. Update all component files to import and use `formatCurrency`
2. Remove all hardcoded "USD" references
3. Test with multiple currencies
4. Update PDF generation to use formatCurrency
5. Add currency validation in forms

## Status

✅ Utility function created
✅ Import added to PO detail page
⏳ Need to replace all instances in components
⏳ Need to test with different currencies
