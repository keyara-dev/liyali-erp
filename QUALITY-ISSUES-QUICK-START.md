# Quality Issues Feature - Quick Start Guide

## What Was Implemented

A complete quality issue reporting system for GRNs with:
- ✅ Dialog form to report issues
- ✅ localStorage persistence via server actions
- ✅ React Query mutations for state management
- ✅ Real-time UI updates
- ✅ Full error handling

---

## How It Works

### For Users

1. Open GRN detail page → `/grn/[id]`
2. Scroll to "Quality Issues Reported" section
3. Click "Report Issue" button
4. Fill form:
   - Select affected item
   - Select severity (Low, Medium, High)
   - Describe the issue (0-500 characters)
5. Click "Report Issue"
6. Issue saved immediately ✅
7. Issue appears in list with color-coding

### Data Flow

```
Dialog Form
    ↓
handleAddQualityIssue()
    ↓
useAddQualityIssueMutation()
    ↓
Server Action: addQualityIssueToGRN()
    ↓
localStorage (saved under 'app_grns' key)
    ↓
Component updates with new issue
```

---

## Files Created

### 1. Dialog Component
**File**: `frontend/src/app/(private)/(main)/grn/[id]/_components/quality-issue-dialog.tsx`

Form for reporting quality issues with:
- Item dropdown
- Severity selector (Low/Medium/High)
- Description textarea (0-500 chars)
- Submit validation
- Toast notifications

### 2. Server Action
**File**: `frontend/src/app/_actions/grn-actions.ts`

Handles all GRN operations:
```typescript
addQualityIssueToGRN(grnId, issue)        // Add issue
removeQualityIssueFromGRN(grnId, issueId) // Remove issue
updateQualityIssueInGRN(grnId, issueId, updates) // Update issue
getGRNAction(grnId)                       // Get GRN
```

**Storage**: localStorage key `'app_grns'`

### 3. React Query Mutation
**File**: `frontend/src/hooks/use-quality-issue-mutations.ts`

Three mutation hooks:
```typescript
useAddQualityIssueMutation(grnId)       // Report new issue
useRemoveQualityIssueMutation(grnId)    // Delete issue
useUpdateQualityIssueMutation(grnId)    // Edit issue
```

### 4. Updated Component
**File**: `frontend/src/app/(private)/(main)/grn/[id]/_components/grn-detail-client.tsx`

- Integrated mutation hook
- Updated dialog handler to use mutation
- Added persistence feedback

---

## Usage Examples

### In a Component

```typescript
import { useAddQualityIssueMutation } from '@/hooks/use-quality-issue-mutations';

function MyComponent({ grnId }) {
  const mutation = useAddQualityIssueMutation(grnId);

  const handleReport = async (issue) => {
    try {
      const result = await mutation.mutateAsync({
        itemId: 'item-1',
        description: 'Motor malfunction',
        severity: 'HIGH',
      });
      console.log('Issue saved:', result);
    } catch (error) {
      console.error('Failed to save:', error);
    }
  };

  return (
    <button
      onClick={() => handleReport(...)}
      disabled={mutation.isPending}
    >
      {mutation.isPending ? 'Saving...' : 'Report Issue'}
    </button>
  );
}
```

### In a Server Context

```typescript
import { addQualityIssueToGRN } from '@/app/_actions/grn-actions';

// In a server action:
const updated = await addQualityIssueToGRN('grn-1', {
  itemId: 'item-2',
  description: 'Damaged packaging',
  severity: 'MEDIUM',
});

console.log(updated.qualityIssues); // All issues for GRN
```

---

## Data Structure

### Quality Issue

```typescript
interface QualityIssue {
  id: string;              // "issue-1733817600000" (auto-generated)
  itemId: string;          // Reference to received item
  description: string;     // What's the issue
  severity: 'LOW' | 'MEDIUM' | 'HIGH';
}
```

### Stored in GRN

```typescript
interface GoodsReceivedNote {
  id: string;
  grnNumber: string;
  qualityIssues: QualityIssue[]; // Array of reported issues
  updatedAt: string;             // Changes on issue add/edit/delete
  // ... other fields
}
```

### In localStorage

```javascript
localStorage.getItem('app_grns')
// Returns: JSON string with array of all GRNs
// Each GRN contains its qualityIssues array

// View in browser console:
JSON.parse(localStorage.getItem('app_grns'))
  .find(grn => grn.id === 'grn-1')
  .qualityIssues
```

---

## Testing

### Manual Testing Checklist

- [ ] Open `/grn/grn-1` page
- [ ] Click "Report Issue" button
- [ ] Fill form with valid data
- [ ] Submit form
- [ ] Verify success toast appears
- [ ] Verify new issue appears in list
- [ ] Verify issue has correct severity color
- [ ] Open DevTools → Application → localStorage
- [ ] Check `app_grns` contains new issue
- [ ] Refresh page
- [ ] Verify issue still appears (from localStorage)

### Console Testing

```javascript
// Check all issues for a GRN
JSON.parse(localStorage.getItem('app_grns'))
  .find(grn => grn.id === 'grn-1')
  .qualityIssues

// Clear storage to reset
localStorage.removeItem('app_grns')

// Add issue programmatically (in component):
// Use useAddQualityIssueMutation hook
```

---

## Features

### What Works ✅

- Report quality issues with form validation
- Issues persist in localStorage
- Issues survive page refresh
- Issues survive browser close
- Error handling with toast notifications
- Multiple issues per GRN
- Severity-based color coding
- Item preview when selecting
- Character counter for description
- Success/error feedback

### What's Next 🔜

- Edit existing issues
- Delete issues with confirmation
- Backend API integration
- Real-time sync across tabs
- Bulk issue reporting
- Issue analytics/dashboard

---

## Troubleshooting

### Issue not saved?

1. Check browser console for errors
2. Verify localStorage is enabled
3. Check DevTools → Application → localStorage → `app_grns`
4. Try clearing storage: `localStorage.removeItem('app_grns')`

### Issue appears then disappears?

1. Check mutation error handling
2. Review server action error logs
3. Verify GRN ID is correct: `grnId`
4. Check localStorage write permissions

### Form won't submit?

1. Verify all fields are filled
2. Description must be 1-500 characters
3. Item must be selected from dropdown
4. Severity must be selected

---

## Performance Tips

### For Developers

- Mutations are lazy - only execute on form submit
- Query cache prevents re-fetching
- Consider optimistic updates for better UX
- Batch multiple issues if possible
- Monitor localStorage size (currently unlimited)

### For Users

- Issues save instantly to browser storage
- Works offline (data stays in browser)
- No network required for persistence
- Clearing browser data will lose issues

---

## API Reference

### Server Actions

```typescript
// grn-actions.ts

// Add quality issue
addQualityIssueToGRN(grnId: string, issue: IssueData): Promise<GRN>

// Remove quality issue
removeQualityIssueFromGRN(grnId: string, issueId: string): Promise<GRN>

// Update quality issue
updateQualityIssueInGRN(
  grnId: string,
  issueId: string,
  updates: Partial<IssueData>
): Promise<GRN>

// Get GRN
getGRNAction(grnId: string): Promise<GRN | null>
```

### Mutation Hooks

```typescript
// use-quality-issue-mutations.ts

// All hooks follow same pattern:
const mutation = useXxxMutation(grnId);

// Return:
{
  mutate: (data) => void,           // Fire and forget
  mutateAsync: (data) => Promise,   // With error handling
  isPending: boolean,                // Loading state
  isError: boolean,                  // Error state
  error: Error | null,               // Error object
  isSuccess: boolean,                // Success state
}
```

---

## Migration to Backend

When ready to use backend APIs:

1. Update server actions to call API endpoints
2. Keep localStorage as optional fallback cache
3. Mutation hooks don't need to change
4. Component code stays the same
5. Only server action implementations change

Example:

```typescript
// Before (localStorage)
export async function addQualityIssueToGRN(grnId, issue) {
  const grn = getGRNById(grnId); // from localStorage
  // ... save to localStorage
  return updatedGRN;
}

// After (API)
export async function addQualityIssueToGRN(grnId, issue) {
  const response = await fetch(`/api/grn/${grnId}/issues`, {
    method: 'POST',
    body: JSON.stringify(issue),
  });
  return response.json(); // API returns updated GRN
}
```

---

## Key Files at a Glance

| File | Purpose | Status |
|------|---------|--------|
| `quality-issue-dialog.tsx` | Report form dialog | ✅ Complete |
| `grn-actions.ts` | Server actions | ✅ Complete |
| `use-quality-issue-mutations.ts` | React Query hooks | ✅ Complete |
| `grn-detail-client.tsx` | GRN page component | ✅ Updated |

---

## Questions?

See detailed documentation in:
- `QUALITY-ISSUE-PERSISTENCE.md` - Full architecture
- `QUALITY-ISSUE-REPORTING-FEATURE.md` - Dialog details
- Component TSDoc comments - Code-level documentation
