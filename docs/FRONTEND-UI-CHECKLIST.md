# Frontend UI/UX Implementation Checklist - Phase 2

**Date:** December 24, 2025
**Framework:** Next.js 15 + React 19 + Tailwind CSS
**Status:** Implementation Ready

---

## 🎨 Design System Integration

### Color Scheme for Phase 2 Features

#### Category Management
- **Primary:** Blue (#3B82F6)
- **Background:** Blue-50 (#F0F9FF)
- **Border:** Blue-200 (#BFDBFE)
- **Text:** Blue-900 (#111827)

#### Estimate Badge
- **Background:** Amber-50 (#FFFBEB)
- **Border:** Amber-200 (#FDE68A)
- **Text:** Amber-900 (#78350F)
- **Icon:** ⚠️

#### Analytics
- **Approved:** Green-500 (#10B981)
- **Rejected:** Red-500 (#EF4444)
- **Pending:** Yellow-500 (#EAB308)
- **Draft:** Gray-500 (#6B7280)

#### Supplier/Vendor
- **Primary:** Indigo (#6366F1)
- **Background:** Indigo-50 (#F0F4FF)

---

## 📦 Component Specifications

### 1. CategorySelect Component

#### Specifications
```
Component Name: CategorySelect
Type: Form Control
Location: frontend/src/components/ui/category-select.tsx
```

#### Features
- [ ] Dropdown/Combobox with search functionality
- [ ] Show category ID and name
- [ ] Async loading from API
- [ ] Display loading state (skeleton or spinner)
- [ ] Display error state with retry button
- [ ] Show empty state if no categories exist
- [ ] Keyboard navigation support (arrow keys)
- [ ] Accessible (ARIA labels)
- [ ] Clear button to deselect

#### UI Layout
```
┌─────────────────────────────────┐
│ Category                        │
├─────────────────────────────────┤
│ [Search... ✕]                   │
├─────────────────────────────────┤
│ ☐ Office Supplies               │
│ ☐ Equipment                     │
│ ☐ Services                      │
│ ☐ Software Licenses             │
└─────────────────────────────────┘
```

#### Props Interface
```typescript
interface CategorySelectProps {
  value?: string
  onChange: (categoryId: string) => void
  disabled?: boolean
  required?: boolean
  error?: string
  placeholder?: string
  onError?: (error: Error) => void
}
```

#### States
- [ ] Idle (default)
- [ ] Loading (showing spinner)
- [ ] Open (dropdown visible)
- [ ] Focused (keyboard focus)
- [ ] Disabled (grayed out)
- [ ] Error (red border + error message)
- [ ] Empty (no options available)

---

### 2. Estimate Badge Component

#### Specifications
```
Component Name: EstimateBadge
Type: Display/Indicator
Location: frontend/src/components/ui/estimate-badge.tsx
```

#### Variants
- [ ] **Inline Badge** - Small, in lists/tables
- [ ] **Card Alert** - Large, in detail views
- [ ] **Text Badge** - Minimal, in headers

#### Visual Design

##### Inline Badge
```
⚠️ Estimate  [Small, inline display]
```

##### Card Alert
```
┌─────────────────────────────────────┐
│ ⚠️ Marked as Estimate              │
│ This is an estimate, not final      │
└─────────────────────────────────────┘
```

#### Props
```typescript
interface EstimateBadgeProps {
  variant?: 'inline' | 'card' | 'text'
  size?: 'sm' | 'md' | 'lg'
  className?: string
}
```

---

### 3. VendorSelect Component

#### Specifications
```
Component Name: VendorSelect
Type: Form Control
Location: frontend/src/components/ui/vendor-select.tsx
```

#### Features
- [ ] Similar to CategorySelect
- [ ] Search by vendor name, ID, or email
- [ ] Show vendor details on hover
- [ ] Optional "Create New Vendor" option
- [ ] Show vendor rating/status badge
- [ ] Async loading with debounce
- [ ] Keyboard navigation
- [ ] Accessible labels

#### Props
```typescript
interface VendorSelectProps {
  value?: string
  onChange: (vendorId: string) => void
  disabled?: boolean
  required?: boolean
  label?: string
  error?: string
  allowCreateNew?: boolean
}
```

---

### 4. Analytics Dashboard Components

#### MetricsCard Component
```
Component Name: MetricsCard
Type: Data Display
Location: frontend/src/components/workflows/metrics-card.tsx
```

##### Specifications
- [ ] Grid layout (1, 2, 3, or 5 columns based on screen size)
- [ ] Show metric title, value, and optional trend
- [ ] Color-coded background based on status
- [ ] Responsive padding
- [ ] Optional click handler for drill-down
- [ ] Loading skeleton state

##### Layout
```
┌─────────────────────────────┐
│ Total Requisitions          │
│ 250                         │
│ ↗ 12% from last month      │
└─────────────────────────────┘
```

##### Props
```typescript
interface MetricsCardProps {
  title: string
  value: number | string
  trend?: {
    direction: 'up' | 'down'
    percentage: number
    period: string
  }
  color?: 'blue' | 'green' | 'red' | 'amber'
  onClick?: () => void
  isLoading?: boolean
}
```

#### RejectionChart Component
```
Component Name: RejectionChart
Type: Visualization
Location: frontend/src/components/workflows/rejection-chart.tsx
```

##### Specifications
- [ ] Line chart (recharts or chart.js)
- [ ] X-axis: Dates (daily/weekly/monthly)
- [ ] Y-axis: Rejection count and rate
- [ ] Dual-axis visualization
- [ ] Interactive tooltips
- [ ] Legend showing both metrics
- [ ] Responsive sizing
- [ ] Loading state

##### Chart Data Example
```
Date        | Rejections | Total | Rate
------------|------------|-------|-----
2025-12-20 | 2          | 15    | 13.3%
2025-12-21 | 1          | 8     | 12.5%
2025-12-22 | 3          | 20    | 15.0%
```

#### RejectionReasonsChart Component
```
Component Name: RejectionReasonsChart
Type: Visualization
Location: frontend/src/components/workflows/rejection-reasons-chart.tsx
```

##### Specifications
- [ ] Horizontal bar chart (top reasons)
- [ ] Show reason, count, and percentage
- [ ] Sortable by count (descending)
- [ ] Color gradient (more common = darker)
- [ ] Interactive tooltips
- [ ] Responsive sizing

##### Chart Data
```
Reason                  | Count | %
------------------------|-------|-----
Budget exceeded         | 12    | 40%
Missing documentation   | 8     | 26.7%
Vendor not approved     | 6     | 20%
Incomplete details      | 4     | 13.3%
```

#### TopApproversTable Component
```
Component Name: TopApproversTable
Type: Data Table
Location: frontend/src/components/workflows/top-approvers-table.tsx
```

##### Specifications
- [ ] Sortable columns (click to sort)
- [ ] Pagination for large datasets
- [ ] Show: Name, Rejections, Approvals, Rate
- [ ] Highlight high rejection rate rows (>30%)
- [ ] Color-coded status bars
- [ ] Hover effects
- [ ] Responsive stacking on mobile

##### Table Layout
```
Approver Name  | Approvals | Rejections | Rate    | Status
---------------|-----------|------------|---------|--------
Jane Smith     | 45        | 5          | 10.0%   | Good ✓
John Doe       | 32        | 12         | 27.3%   | Warning ⚠
Sarah Johnson  | 28        | 2          | 6.7%    | Good ✓
```

---

## 🎯 Page-Level Checklists

### Categories Management Page

#### URL: `/admin/categories` (or similar)

#### Header Section
- [ ] Page title: "Category Management"
- [ ] Breadcrumb navigation
- [ ] Create Category button (primary action)
- [ ] Search/filter bar
- [ ] Pagination controls

#### Main Content
- [ ] Category table with columns:
  - [ ] Checkbox (bulk select)
  - [ ] Category Name
  - [ ] Budget Codes (count badge)
  - [ ] Active Status (toggle)
  - [ ] Created Date
  - [ ] Actions (Edit, Delete, Manage Codes)

#### Category Detail View
- [ ] Modal or side panel
- [ ] Show category details
- [ ] Edit category form
- [ ] Budget codes list with add/remove
- [ ] Delete category button (with confirmation)

#### States & Interactions
- [ ] Empty state (no categories)
- [ ] Loading state (skeleton)
- [ ] Error state (with retry)
- [ ] Confirmation dialogs for destructive actions
- [ ] Success/error toast messages
- [ ] Form validation errors

---

### Requisition Create/Edit Page

#### NEW SECTION: Category & Supplier

##### Category Selection
```
┌─────────────────────────────────────┐
│ Category *                          │
├─────────────────────────────────────┤
│ [Select category... ▼]              │
│ Help text: Choose the category      │
│ for this requisition                │
└─────────────────────────────────────┘
```

##### Preferred Supplier
```
┌─────────────────────────────────────┐
│ Preferred Supplier                  │
├─────────────────────────────────────┤
│ [Select supplier... ▼]              │
│ Help text: Optional - leave empty   │
│ if not predetermined                │
└─────────────────────────────────────┘
```

##### Estimate Checkbox
```
┌─────────────────────────────────────┐
│ ☐ Mark as Estimate                 │
│ Use this to indicate this is an     │
│ estimate, not a final cost          │
└─────────────────────────────────────┘
```

#### Features
- [ ] Category field is visible prominently
- [ ] Category field can be mandatory or optional
- [ ] Supplier field is optional
- [ ] Estimate checkbox is clearly labeled
- [ ] Help text explains each field
- [ ] Visual feedback when fields are filled
- [ ] Form validation for category (if required)
- [ ] Pre-fill from URL params if editing

#### Validation
- [ ] Show error if category required but not selected
- [ ] Show error if vendor ID invalid
- [ ] Estimate field accepts boolean only

---

### Requisition Detail View

#### Display New Fields
- [ ] Category name (as badge or text)
- [ ] Preferred supplier name (if set)
- [ ] Estimate indicator (prominent placement)

#### Estimate Indicator Placement
```
BEFORE TOTAL SECTION:
┌──────────────────────────────────────┐
│ ⚠️ MARKED AS ESTIMATE               │
│ This requisition is an estimate and  │
│ may not reflect final costs.         │
└──────────────────────────────────────┘

IN TOTALS SECTION:
Category: Office Supplies
Supplier: ACME Corp
Type: [Estimate Badge] or [Final]
```

---

### Analytics Dashboard Page

#### URL: `/analytics` (or `/dashboard`)

#### Layout
```
┌──────────────────────────────────────────┐
│ Analytics Dashboard                      │
│ [Date Range] [Period] [Department]       │
└──────────────────────────────────────────┘

[Metrics Cards Grid - 5 columns]
[Total] [Draft] [Approved] [Rejected] [Rate]

[Charts Grid - 2 columns]
[Rejections Over Time] [Rejection Reasons]

[Table - Full Width]
[Top Approvers Performance]
```

#### Filters
- [ ] Date Range Picker
  - [ ] From date input
  - [ ] To date input
  - [ ] Preset options (Last 7 days, Last 30 days, YTD)
  - [ ] Clear button

- [ ] Period Selector
  - [ ] Daily
  - [ ] Weekly
  - [ ] Monthly

- [ ] Department Filter
  - [ ] Multi-select dropdown
  - [ ] Search within dropdown
  - [ ] Clear selection button

#### Metrics Section
- [ ] Display 5 metric cards in responsive grid
- [ ] Show loading skeleton until data loads
- [ ] Show error state with retry button
- [ ] Format large numbers (1.2k, 45.3%)
- [ ] Clicking metric shows drill-down detail (optional)

#### Charts Section
- [ ] Left chart: Rejections over time (line chart)
- [ ] Right chart: Rejection reasons (horizontal bar)
- [ ] Both charts responsive
- [ ] Both charts show loading states
- [ ] Interactive legends
- [ ] Export data button (optional)

#### Table Section
- [ ] Sortable columns
- [ ] Pagination (10, 25, 50 rows per page)
- [ ] Hover effects on rows
- [ ] Color-coding for high rejection rates
- [ ] Export as CSV button
- [ ] No data state

---

## 📱 Responsive Design Checklist

### Mobile (< 640px)
- [ ] Category dropdown full width
- [ ] Metrics cards stack to 1 column
- [ ] Charts stack vertically
- [ ] Table becomes card view
- [ ] Touch-friendly button sizes (48px min)
- [ ] Filter collapsible/drawer
- [ ] Horizontal scroll for wide tables

### Tablet (640px - 1024px)
- [ ] Category dropdown full width or half width
- [ ] Metrics cards 2-3 columns
- [ ] Charts side by side with scroll
- [ ] Table remains tabular
- [ ] Larger touch targets
- [ ] Sidebar for filters

### Desktop (> 1024px)
- [ ] Full-width layouts
- [ ] Metrics cards 5 columns (responsive to fewer if needed)
- [ ] Charts grid 2 columns
- [ ] Full tables with horizontal scroll backup
- [ ] Inline filters
- [ ] Sidebar navigation visible

---

## 🎬 Animation & Interactions

### Loading States
- [ ] Show skeleton loaders for cards
- [ ] Show spinner in dropdown during search
- [ ] Show progress bar for long operations
- [ ] Pulse animation for "updating" state

### Transitions
- [ ] Smooth opacity fade for modals
- [ ] Slide-in for side panels
- [ ] Slide-down for dropdowns
- [ ] Fade-in for newly loaded content
- [ ] Bounce animation on success

### Hover Effects
- [ ] Card hover: slight shadow increase
- [ ] Button hover: slight background change
- [ ] Table row hover: subtle background color
- [ ] Link hover: underline or color change

### Feedback
- [ ] Toast notifications for:
  - [ ] Category created successfully
  - [ ] Category deleted successfully
  - [ ] Requisition saved with new fields
  - [ ] Error creating/updating category
  - [ ] No categories available

---

## ♿ Accessibility (A11y)

### Form Accessibility
- [ ] All form fields have `<label>` elements
- [ ] Category select has `aria-label` or `aria-labelledby`
- [ ] Error messages linked to fields via `aria-describedby`
- [ ] Required fields marked with `*` and `aria-required="true"`
- [ ] Focus management when opening modals

### Chart Accessibility
- [ ] Charts have alt text or ARIA label
- [ ] Data available in table format as alternative
- [ ] Color not sole indicator of status (use icons/text too)
- [ ] Sufficient color contrast (WCAG AA minimum)

### Keyboard Navigation
- [ ] Tab through all interactive elements
- [ ] Enter/Space to activate buttons
- [ ] Escape to close modals/dropdowns
- [ ] Arrow keys to navigate dropdowns
- [ ] Shift+Tab to go backwards

### Screen Reader Support
- [ ] All buttons have descriptive text
- [ ] Images have alt text
- [ ] Charts have ARIA labels
- [ ] Status indicators announced
- [ ] Loading states announced

---

## 🧪 Testing Checklist

### Unit Tests
- [ ] CategorySelect renders without crashing
- [ ] CategorySelect loads categories on mount
- [ ] CategorySelect updates value on selection
- [ ] EstimateBadge renders all variants
- [ ] MetricsCard formats numbers correctly
- [ ] RejectionChart renders with sample data

### Integration Tests
- [ ] Category create form → API call → success toast
- [ ] Requisition form with category → submit works
- [ ] Analytics filters → updates chart data
- [ ] Date range change → refetches metrics

### E2E Tests
- [ ] User can create category
- [ ] User can use category in requisition
- [ ] User can view analytics
- [ ] User can filter by date range
- [ ] Mobile navigation works

### Visual Regression Tests
- [ ] Compare component screenshots
- [ ] Test different data states
- [ ] Test responsive layouts

---

## 📊 Performance Checklist

- [ ] Category list lazy-loads (virtual scrolling for 100+ items)
- [ ] Analytics data cached (5 min TTL)
- [ ] Charts lazy-loaded (imported with dynamic())
- [ ] Images optimized (next/image)
- [ ] CSS tree-shaken (unused styles removed)
- [ ] Bundle size monitored
- [ ] PageSpeed Insights > 90

---

## 🎨 Design Files

### Figma/Design Deliverables
- [ ] Component library created
- [ ] Variants for all states
- [ ] Responsive breakpoints shown
- [ ] Color scheme defined
- [ ] Typography scale defined
- [ ] Spacing/grid system
- [ ] Icon set finalized

---

## 📝 Copy & Content

### Button Labels
- [ ] "Add Category"
- [ ] "Create Category"
- [ ] "Save Category"
- [ ] "Delete Category"
- [ ] "View Analytics"
- [ ] "Export Report"

### Placeholder Text
- [ ] "Select category..."
- [ ] "Search categories..."
- [ ] "No categories found"
- [ ] "Loading categories..."

### Help Text
- [ ] Category: "Choose a category to organize this requisition"
- [ ] Supplier: "Optional - your preferred vendor for this item"
- [ ] Estimate: "Mark this requisition as an estimate rather than final cost"

### Error Messages
- [ ] "Category not found"
- [ ] "Failed to load categories"
- [ ] "Category is required"
- [ ] "Failed to create category"
- [ ] "Please try again"

### Success Messages
- [ ] "Category created successfully"
- [ ] "Category updated successfully"
- [ ] "Category deleted successfully"
- [ ] "Requisition saved with category"

---

## 📋 Sign-Off Checklist

### Frontend Lead
- [ ] All components meet design spec
- [ ] Code follows project conventions
- [ ] No accessibility violations
- [ ] Performance acceptable
- [ ] Mobile responsive
- [ ] Error states handled

### QA Lead
- [ ] Manual testing completed
- [ ] Edge cases tested
- [ ] Cross-browser compatibility
- [ ] Mobile testing done
- [ ] No console errors
- [ ] API integration verified

### UX/Design Lead
- [ ] Visual consistency
- [ ] User flows intuitive
- [ ] Loading states clear
- [ ] Error states helpful
- [ ] Responsive design works
- [ ] Accessibility compliant

---

**Implementation Status:** Ready for Frontend Development
**Estimated Duration:** 5 days
**Team Size:** 2-3 frontend developers

---

*Last Updated: December 24, 2025*
*Questions? Refer to FRONTEND-INTEGRATION-GUIDE.md*
