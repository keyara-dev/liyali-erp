# Currency Formatting - Quick Reference Guide

## What Changed

All currency formatting across the application now uses a centralized `formatCurrency` utility function instead of hardcoded formatting.

## How to Use

### Import the Function

```typescript
import { formatCurrency } from "@/lib/utils";
```

### Basic Usage

```typescript
// Format with currency code
formatCurrency(1234.56, "ZMW"); // Returns: "ZMW 1,234.56"
formatCurrency(1234.56, "USD"); // Returns: "USD 1,234.56"
formatCurrency(1234.56, "EUR"); // Returns: "EUR 1,234.56"

// Handles null/undefined gracefully
formatCurrency(null, "ZMW"); // Returns: "ZMW 0.00"
formatCurrency(undefined, "USD"); // Returns: "USD 0.00"

// Default currency (USD) if not specified
formatCurrency(1234.56); // Returns: "USD 1,234.56"
```

### In Components

```tsx
// Purchase Order
<p>{formatCurrency(purchaseOrder.totalAmount, purchaseOrder.currency)}</p>

// Payment Voucher
<p>{formatCurrency(pv.totalAmount || pv.amount, pv.currency)}</p>

// Requisition
<p>{formatCurrency(requisition.totalAmount, requisition.currency)}</p>

// Budget Item
<p>{formatCurrency(budgetItem.amount, budgetItem.currency)}</p>
```

## Function Signature

```typescript
export function formatCurrency(
  amount: number | null | undefined,
  currency: string = "USD",
  locale: string = "en-ZM",
): string;
```

### Parameters

- `amount`: The numeric amount to format (can be null/undefined)
- `currency`: The currency code (e.g., "ZMW", "USD", "EUR") - defaults to "USD"
- `locale`: The locale for number formatting - defaults to "en-ZM"

### Returns

A formatted string in the format: `"CURRENCY 1,234.56"`

## Currency Flow in Documents

```
REQ (ZMW) → PO (ZMW) → PV (ZMW)
                    → GRN (references PO)
```

Currency is always derived from the source requisition and propagated through all related documents.

## Default Currency

The system default currency is **ZMW (Zambian Kwacha)**.

## Examples from the Codebase

### Before (Hardcoded)

```tsx
{
  purchaseOrder.currency;
}
{
  purchaseOrder.totalAmount?.toLocaleString("en-ZM", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}
```

### After (Using Utility)

```tsx
{
  formatCurrency(purchaseOrder.totalAmount, purchaseOrder.currency);
}
```

## Benefits

1. **Consistency**: Same formatting everywhere
2. **Maintainability**: Change once, applies everywhere
3. **Type Safety**: Handles null/undefined gracefully
4. **Internationalization**: Easy to add new locales
5. **No Hardcoding**: Currency comes from document data

## Testing

To test currency formatting:

1. Create a REQ with ZMW currency
2. Create PO from REQ - should show "ZMW 1,234.56"
3. Create PV from PO - should show "ZMW 1,234.56"
4. Create GRN from PO - should show "ZMW 1,234.56"
5. Try with USD and EUR currencies
6. Verify all amounts display correctly

## Where It's Used

- Purchase Order detail page
- Payment Voucher detail page
- Payment Voucher approval page
- Payment Voucher submit dialog
- Payment Voucher items list
- Create PV from PO dialog
- Approved POs table
- GRN create dialog
- Requisitions table
- Budget items manager
- Budget item dialogs

## Notes

- Always pass the currency from the document object
- Don't hardcode currency codes in components
- The function handles edge cases (null, undefined, 0)
- Format is always: "CURRENCY 1,234.56" (with space and 2 decimals)
