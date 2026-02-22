# Requisition Configuration Checklist Implementation

## Overview

This implementation adds intelligent configuration requirement banners to the requisition creation and submission flows. The system checks for required configurations and displays helpful checklists to guide users through setup.

## Problem Statement

Users attempting to create requisitions need:

1. **Departments** - To categorize the requisition
2. **Categories** - To classify the type of items
3. **Budget Codes** - To allocate costs

Users attempting to submit requisitions for approval additionally need: 4. **Workflows** - To define the approval process

Without these configurations, users would encounter errors or be unable to complete their tasks.

## Solution Architecture

### 1. Configuration Status Hook (`use-configuration-status.ts`)

A reusable hook that checks the status of required configurations:

```typescript
const configStatus = useConfigurationStatus({
  includeWorkflow: false, // Set to true when checking submission requirements
  workflowEntityType: "requisition",
});
```

**Returns:**

- `requirements`: Array of configuration requirements with status
- `allConfigured`: Boolean indicating if all requirements are met
- `missingCount`: Number of missing configurations
- `isLoading`: Loading state

**Features:**

- Checks departments, categories, budgets, and optionally workflows
- Provides counts for each configuration type
- Includes navigation paths to configuration pages
- Handles loading states gracefully

### 2. Configuration Checklist Banner (`configuration-checklist-banner.tsx`)

A reusable banner component that displays configuration requirements:

**Props:**

- `requirements`: Array of configuration requirements
- `title`: Custom title (optional)
- `description`: Custom description (optional)
- `variant`: "creation" or "submission"
- `className`: Additional CSS classes

**Features:**

- Visual checklist with status indicators (✓ or ✗)
- Shows count of configured items
- "Configure" buttons that navigate to admin pages
- Automatically hides when all requirements are met
- Responsive design with dark mode support
- Color-coded: Amber for missing, Green for complete

**Visual Design:**

```
┌─────────────────────────────────────────────────────────┐
│ ⚠️  Configuration Required                              │
│                                                          │
│ Complete the following configurations to create...      │
│                                                          │
│ ┌─────────────────────────────────────────────────┐    │
│ │ ✗ Departments                    [Configure →]  │    │
│ │   At least one active department must be...     │    │
│ └─────────────────────────────────────────────────┘    │
│                                                          │
│ ┌─────────────────────────────────────────────────┐    │
│ │ ✓ Categories          [3 configured]            │    │
│ │   At least one active category must be...       │    │
│ └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

### 3. Workflow Requirement Banner (`workflow-requirement-banner.tsx`)

A specialized banner for workflow configuration requirements:

**Props:**

- `entityType`: Document type (requisition, budget, etc.)
- `className`: Additional CSS classes
- `onDismiss`: Optional dismiss handler

**Features:**

- Specifically designed for workflow requirements
- Explains what users can still do without workflows
- Provides context about draft status
- Only shows when no workflows are configured
- Blue color scheme to differentiate from creation requirements

**Use Case:**

- Shown in submission dialogs when no workflow exists
- Informs users they can create but not submit documents
- Guides users to workflow configuration

### 4. Integration Points

#### A. Create Requisition Dialog

**Location:** `frontend/src/app/(private)/(main)/requisitions/_components/create-requisition-dialog.tsx`

**Implementation:**

```typescript
// Check configuration status (without workflow)
const configStatus = useConfigurationStatus({
  includeWorkflow: false,
});

// Show banner at top of form
{!configStatus.allConfigured && !configStatus.isLoading && (
  <ConfigurationChecklistBanner
    requirements={configStatus.requirements}
    variant="creation"
  />
)}

// Disable submit button if configs missing
<Button
  disabled={!configStatus.allConfigured}
  onClick={handleSubmit}
>
  Create Requisition
</Button>
```

**Behavior:**

- Banner appears at the top of the dialog
- Shows missing configurations with navigation links
- Submit button is disabled until all requirements are met
- Banner automatically hides when all configs are complete

#### B. Submit Requisition Dialog

**Location:** `frontend/src/app/(private)/(main)/requisitions/_components/requisition-submit-dialog.tsx`

**Implementation:**

```typescript
// Show workflow requirement banner
<WorkflowRequirementBanner entityType="requisition" />

// Workflow selector (existing)
<WorkflowSelector
  entityType="requisition"
  value={workflowId}
  onChange={setWorkflowId}
  required
/>
```

**Behavior:**

- Banner appears if no workflows are configured
- Explains that workflows are needed for submission
- Provides context about what users can still do
- Banner automatically hides when workflows exist

## User Experience Flow

### Scenario 1: New User - No Configurations

1. User clicks "Create Requisition"
2. Dialog opens with prominent amber banner at top
3. Banner shows checklist:
   - ✗ Departments (0 configured) [Configure →]
   - ✗ Categories (0 configured) [Configure →]
   - ✗ Budget Codes (0 configured) [Configure →]
4. User clicks "Configure" buttons to set up each requirement
5. As each is configured, it turns green with checkmark
6. When all are complete, banner disappears
7. Submit button becomes enabled

### Scenario 2: Partial Configuration

1. User has departments and categories but no budgets
2. Dialog shows banner with:
   - ✓ Departments (5 configured)
   - ✓ Categories (12 configured)
   - ✗ Budget Codes (0 configured) [Configure →]
3. User configures budgets
4. Banner disappears, form becomes fully functional

### Scenario 3: Attempting Submission Without Workflow

1. User creates requisition successfully
2. User clicks "Submit for Approval"
3. Submit dialog shows blue workflow banner
4. Banner explains:
   - No workflow configured for requisitions
   - Can still save as draft
   - Need to configure workflow to submit
5. User clicks "Configure Workflow"
6. After workflow setup, banner disappears
7. User can now submit for approval

## Reusability

### For Other Document Types

The components are designed to be reusable for all document types:

**Budget Creation:**

```typescript
const configStatus = useConfigurationStatus({
  includeWorkflow: false,
});

<ConfigurationChecklistBanner
  requirements={configStatus.requirements}
  variant="creation"
  title="Budget Configuration Required"
/>
```

**Purchase Order Submission:**

```typescript
<WorkflowRequirementBanner entityType="purchase_order" />
```

**Payment Voucher:**

```typescript
<WorkflowRequirementBanner entityType="payment_voucher" />
```

### Customization Options

1. **Custom Requirements:**

```typescript
const customRequirements: ConfigurationRequirement[] = [
  {
    id: "vendors",
    label: "Vendors",
    description: "At least one vendor must be configured",
    isConfigured: vendors.length > 0,
    count: vendors.length,
    navigateTo: "/admin/vendors",
  },
];

<ConfigurationChecklistBanner requirements={customRequirements} />
```

2. **Custom Styling:**

```typescript
<ConfigurationChecklistBanner
  requirements={requirements}
  className="mb-6 shadow-lg"
/>
```

3. **Custom Messages:**

```typescript
<ConfigurationChecklistBanner
  requirements={requirements}
  title="Setup Required"
  description="Complete these steps to get started:"
/>
```

## Technical Details

### Data Fetching

The hook uses existing query hooks:

- `useActiveDepartments()` - Fetches active departments
- `useCategories()` - Fetches active categories
- `useAllBudgets()` - Fetches all budgets
- `useWorkflows()` - Fetches workflows by entity type

### Performance

- Queries are cached by React Query
- Banner only re-renders when configuration data changes
- Automatically hides when not needed (no DOM overhead)
- Loading states prevent layout shift

### Accessibility

- Semantic HTML with proper ARIA labels
- Color is not the only indicator (icons + text)
- Keyboard navigable buttons
- Screen reader friendly descriptions
- High contrast in both light and dark modes

## Configuration Navigation Paths

| Requirement | Admin Path           |
| ----------- | -------------------- |
| Departments | `/admin/departments` |
| Categories  | `/admin/categories`  |
| Budgets     | `/admin/budgets`     |
| Workflows   | `/admin/workflows`   |

## Future Enhancements

### Potential Additions:

1. **Vendors Requirement:**
   - Add vendor configuration check for purchase orders
   - Show vendor count and setup link

2. **Currency Configuration:**
   - Check if multiple currencies are needed
   - Guide setup of exchange rates

3. **Approval Limits:**
   - Verify approval limits are configured
   - Show which roles need limits set

4. **Email Templates:**
   - Check if notification templates exist
   - Guide template customization

5. **Progress Indicator:**
   - Show overall setup progress (e.g., "3 of 5 complete")
   - Gamify the configuration process

6. **Guided Setup Wizard:**
   - Multi-step wizard for first-time setup
   - Walk through all configurations in order

## Testing Checklist

- [ ] Banner shows when departments are missing
- [ ] Banner shows when categories are missing
- [ ] Banner shows when budgets are missing
- [ ] Banner hides when all requirements are met
- [ ] Configure buttons navigate to correct admin pages
- [ ] Submit button is disabled when configs missing
- [ ] Loading states display correctly
- [ ] Workflow banner shows in submit dialog
- [ ] Workflow banner hides when workflows exist
- [ ] Dark mode displays correctly
- [ ] Mobile responsive layout works
- [ ] Keyboard navigation functions properly

## Files Created/Modified

### New Files:

1. `frontend/src/hooks/use-configuration-status.ts` - Configuration status hook
2. `frontend/src/components/ui/configuration-checklist-banner.tsx` - Main banner component
3. `frontend/src/components/ui/workflow-requirement-banner.tsx` - Workflow-specific banner

### Modified Files:

1. `frontend/src/app/(private)/(main)/requisitions/_components/create-requisition-dialog.tsx`
   - Added configuration status check
   - Added banner display
   - Added button disable logic

2. `frontend/src/app/(private)/(main)/requisitions/_components/requisition-submit-dialog.tsx`
   - Added workflow requirement banner

## Summary

This implementation provides a user-friendly way to guide users through required configurations before they can create or submit requisitions. The components are:

- **Reusable** - Can be used for any document type
- **Intelligent** - Automatically detects missing configurations
- **Helpful** - Provides clear guidance and navigation
- **Performant** - Minimal overhead, hides when not needed
- **Accessible** - Works for all users regardless of ability

The system improves the user experience by preventing errors and providing clear paths to resolution when configurations are missing.
