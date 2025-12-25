# Frontend Integration Guide - Phase 2 Features

**Date:** December 24, 2025
**Status:** Complete
**Framework:** Next.js 15 + React 19

---

## 📋 Overview

This guide maps Phase 2 backend features to frontend components and provides integration instructions for the frontend team.

### Phase 2 Features to Integrate
1. **Category Management** - Organize requisitions by category
2. **Requisition Enhancements** - Add category, supplier, estimate fields
3. **User Last Login Tracking** - Display user activity timestamps
4. **Analytics Engine** - Show metrics and insights

---

## 🔗 API Integration Points

### Base URL Configuration
```typescript
// frontend/src/app/_actions/api-config.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
```

### New Endpoints to Integrate

#### Category Management Endpoints
```
POST   /api/v1/categories                          - Create category
GET    /api/v1/categories                          - List categories (paginated)
GET    /api/v1/categories/{id}                     - Get category details
PUT    /api/v1/categories/{id}                     - Update category
DELETE /api/v1/categories/{id}                     - Delete category (soft delete)
GET    /api/v1/categories/{id}/budget-codes        - List budget codes
POST   /api/v1/categories/{id}/budget-codes        - Add budget code
DELETE /api/v1/categories/{id}/budget-codes/{code} - Remove budget code
```

#### Requisition Enhancement Endpoints
```
POST   /api/v1/requisitions       - Create (NEW FIELDS)
GET    /api/v1/requisitions       - List (UPDATED)
GET    /api/v1/requisitions/{id}  - Get details (UPDATED)
PUT    /api/v1/requisitions/{id}  - Update (NEW FIELDS)
```

#### Analytics Endpoints
```
GET /api/v1/analytics/requisitions/metrics  - Get requisition metrics
GET /api/v1/analytics/approvals/metrics     - Get approval metrics
GET /api/v1/analytics/dashboard             - Get dashboard overview
```

#### User Login Enhancement
```
POST /api/v1/auth/login - Returns lastLogin in response
```

---

## 🗂️ Frontend Component Mapping

### 1. Category Management Components

#### New Components to Create
| Component | Location | Purpose |
|-----------|----------|---------|
| `CategoryManager` | `frontend/src/app/(private)/(main)/categories` | Main category list page |
| `CategoryForm` | `frontend/src/app/(private)/(main)/categories/_components` | Create/Edit category modal |
| `CategorySelect` | `frontend/src/components/ui` | Reusable category dropdown |
| `BudgetCodeManager` | `frontend/src/app/(private)/(main)/categories/_components` | Manage budget codes |

#### CategoryManager Component
```typescript
// frontend/src/app/(private)/(main)/categories/page.tsx
'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { CategoryForm } from './_components/category-form'
import { CategoryTable } from './_components/category-table'
import { BudgetCodeManager } from './_components/budget-code-manager'

export default function CategoriesPage() {
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const { data: categories, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => fetch('/api/categories').then(r => r.json()),
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateCategoryRequest) =>
      fetch('/api/categories', {
        method: 'POST',
        body: JSON.stringify(data),
        headers: { 'Content-Type': 'application/json' },
      }).then(r => r.json()),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['categories'] }),
  })

  return (
    <div className="space-y-6">
      <CategoryForm onSubmit={createMutation.mutate} />
      <CategoryTable
        categories={categories?.data}
        isLoading={isLoading}
        onSelectCategory={setSelectedCategory}
      />
      {selectedCategory && (
        <BudgetCodeManager categoryId={selectedCategory} />
      )}
    </div>
  )
}
```

#### CategorySelect Component (Reusable)
```typescript
// frontend/src/components/ui/category-select.tsx
import { useQuery } from '@tanstack/react-query'
import { SelectField } from './select-field'

interface CategorySelectProps {
  value?: string
  onChange: (value: string) => void
  disabled?: boolean
}

export function CategorySelect({ value, onChange, disabled }: CategorySelectProps) {
  const { data: categories, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => fetch('/api/categories').then(r => r.json()),
  })

  const options = categories?.data?.map(cat => ({
    value: cat.id,
    label: cat.name,
  })) || []

  return (
    <SelectField
      label="Category"
      value={value}
      onChange={onChange}
      options={options}
      disabled={disabled || isLoading}
      placeholder="Select a category"
    />
  )
}
```

### 2. Requisition Enhancement Components

#### Update Requisition Form
```typescript
// frontend/src/app/(private)/(main)/requisitions/create/_components/create-form.tsx
// ADD THESE NEW FIELDS:

import { CategorySelect } from '@/components/ui/category-select'
import { VendorSelect } from '@/components/ui/vendor-select'

export function CreateRequisitionForm() {
  const [formData, setFormData] = useState({
    // ... existing fields ...

    // NEW PHASE 2 FIELDS:
    categoryId: '',          // Selected category ID
    preferredVendorId: '',   // Preferred supplier ID
    isEstimate: false,       // Mark as estimate
  })

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* ... existing fields ... */}

      {/* NEW PHASE 2 FIELDS */}
      <CategorySelect
        value={formData.categoryId}
        onChange={(categoryId) =>
          setFormData(prev => ({ ...prev, categoryId }))
        }
      />

      <VendorSelect
        value={formData.preferredVendorId}
        onChange={(vendorId) =>
          setFormData(prev => ({ ...prev, preferredVendorId: vendorId }))
        }
        label="Preferred Supplier (Optional)"
      />

      <div className="flex items-center space-x-2">
        <input
          type="checkbox"
          id="isEstimate"
          checked={formData.isEstimate}
          onChange={(e) =>
            setFormData(prev => ({ ...prev, isEstimate: e.target.checked }))
          }
        />
        <label htmlFor="isEstimate">Mark as Estimate</label>
      </div>

      <button type="submit">Create Requisition</button>
    </form>
  )
}
```

#### Update Requisition Response Display
```typescript
// Requisition Detail View - Show New Fields
interface RequisitionDetailProps {
  requisition: Requisition & {
    categoryName?: string
    preferredVendorName?: string
    isEstimate: boolean
  }
}

export function RequisitionDetail({ requisition }: RequisitionDetailProps) {
  return (
    <div className="space-y-4">
      {/* ... existing fields ... */}

      {/* NEW PHASE 2 FIELDS */}
      {requisition.categoryName && (
        <div>
          <label className="text-sm font-medium">Category</label>
          <p className="text-lg">{requisition.categoryName}</p>
        </div>
      )}

      {requisition.preferredVendorName && (
        <div>
          <label className="text-sm font-medium">Preferred Supplier</label>
          <p className="text-lg">{requisition.preferredVendorName}</p>
        </div>
      )}

      {requisition.isEstimate && (
        <div className="rounded-lg bg-amber-50 p-3 border border-amber-200">
          <p className="text-sm font-medium text-amber-900">⚠️ Marked as Estimate</p>
          <p className="text-sm text-amber-700">This is an estimate, not a final purchase</p>
        </div>
      )}
    </div>
  )
}
```

### 3. User Last Login Tracking

#### Update User Profile Display
```typescript
// frontend/src/app/(private)/settings/_components/account-settings.tsx

import { formatDistanceToNow } from 'date-fns'

export function AccountSettings({ user }) {
  return (
    <div className="space-y-6">
      {/* ... existing fields ... */}

      {/* NEW: Last Login Display */}
      {user.lastLogin && (
        <div>
          <label className="text-sm font-medium">Last Login</label>
          <p className="text-lg text-gray-600">
            {formatDistanceToNow(new Date(user.lastLogin), { addSuffix: true })}
          </p>
          <p className="text-xs text-gray-500">
            {new Date(user.lastLogin).toLocaleString()}
          </p>
        </div>
      )}
    </div>
  )
}
```

#### Update Login Response Handler
```typescript
// frontend/src/app/_actions/auth.ts

export async function handleLogin(email: string, password: string) {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })

  if (!response.ok) throw new Error('Login failed')

  const data = await response.json()

  // NEW: Store lastLogin timestamp
  if (data.user.lastLogin) {
    localStorage.setItem('lastLogin', data.user.lastLogin)
  }

  return data
}
```

### 4. Analytics Dashboard Components

#### New Analytics Components
| Component | Location | Purpose |
|-----------|----------|---------|
| `AnalyticsDashboard` | `frontend/src/app/(private)/(main)/analytics` | Main analytics page |
| `MetricsCards` | `frontend/src/components/workflows` | Status count cards |
| `RejectionChart` | `frontend/src/components/workflows` | Rejection rate visualization |
| `TopApproversTable` | `frontend/src/components/workflows` | Approver performance |

#### Analytics Dashboard Component
```typescript
// frontend/src/app/(private)/(main)/analytics/page.tsx
'use client'

import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { DateRangePicker } from '@/components/ui/date-range-picker'
import { Select } from '@/components/ui/select'
import { MetricsCards } from './_components/metrics-cards'
import { RejectionChart } from './_components/rejection-chart'
import { TopApproversTable } from './_components/top-approvers-table'
import { RejectionReasonsChart } from './_components/rejection-reasons-chart'

export default function AnalyticsPage() {
  const [dateRange, setDateRange] = useState({ start: null, end: null })
  const [department, setDepartment] = useState('')
  const [period, setPeriod] = useState('daily')

  // Build query params
  const params = new URLSearchParams()
  if (dateRange.start) params.append('start_date', dateRange.start.toISOString())
  if (dateRange.end) params.append('end_date', dateRange.end.toISOString())
  if (department) params.append('department', department)
  if (period) params.append('period', period)

  const { data: metrics, isLoading } = useQuery({
    queryKey: ['analytics:metrics', params.toString()],
    queryFn: () =>
      fetch(`/api/analytics/requisitions/metrics?${params}`).then(r => r.json()),
  })

  if (isLoading) return <div>Loading analytics...</div>

  return (
    <div className="space-y-8 p-6">
      {/* Filters */}
      <div className="flex gap-4 items-end">
        <DateRangePicker
          value={dateRange}
          onChange={setDateRange}
          label="Date Range"
        />
        <Select
          value={period}
          onChange={setPeriod}
          options={[
            { value: 'daily', label: 'Daily' },
            { value: 'weekly', label: 'Weekly' },
            { value: 'monthly', label: 'Monthly' },
          ]}
          label="Period"
        />
      </div>

      {/* Key Metrics */}
      <MetricsCards
        totalRequisitions={metrics?.data?.totalRequisitions}
        statusCounts={metrics?.data?.statusCounts}
        rejectionRate={metrics?.data?.rejectionRate}
      />

      {/* Charts */}
      <div className="grid grid-cols-2 gap-6">
        <RejectionChart
          data={metrics?.data?.rejectionsOverTime}
          title="Rejections Over Time"
        />
        <RejectionReasonsChart
          data={metrics?.data?.rejectionReasons}
          title="Top Rejection Reasons"
        />
      </div>

      {/* Top Approvers */}
      <TopApproversTable
        data={metrics?.data?.topRejectingApprovers}
        title="Approver Performance"
      />
    </div>
  )
}
```

#### Metrics Cards Component
```typescript
// frontend/src/components/workflows/metrics-cards.tsx

export function MetricsCards({ totalRequisitions, statusCounts, rejectionRate }) {
  const cards = [
    {
      title: 'Total Requisitions',
      value: totalRequisitions,
      color: 'bg-blue-50 border-blue-200',
    },
    {
      title: 'Draft',
      value: statusCounts?.draft || 0,
      color: 'bg-gray-50 border-gray-200',
    },
    {
      title: 'Approved',
      value: statusCounts?.approved || 0,
      color: 'bg-green-50 border-green-200',
    },
    {
      title: 'Rejected',
      value: statusCounts?.rejected || 0,
      color: 'bg-red-50 border-red-200',
    },
    {
      title: 'Rejection Rate',
      value: `${rejectionRate?.toFixed(2)}%`,
      color: 'bg-orange-50 border-orange-200',
    },
  ]

  return (
    <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
      {cards.map(card => (
        <div
          key={card.title}
          className={`p-6 rounded-lg border ${card.color}`}
        >
          <h3 className="text-sm font-medium text-gray-600">{card.title}</h3>
          <p className="text-2xl font-bold mt-2">{card.value}</p>
        </div>
      ))}
    </div>
  )
}
```

---

## 📝 TypeScript Type Updates

### Updated Requisition Types
```typescript
// frontend/src/types/requisition.ts - ADD THESE FIELDS:

export interface Requisition {
  // ... existing fields ...

  // NEW PHASE 2 FIELDS:
  categoryId?: string          // Category ID
  categoryName?: string        // Category name (from backend)
  preferredVendorId?: string   // Preferred supplier ID
  preferredVendorName?: string // Supplier name (from backend)
  isEstimate: boolean          // Mark as estimate
}

export interface CreateRequisitionRequest {
  // ... existing fields ...

  // NEW PHASE 2 FIELDS:
  categoryId?: string
  preferredVendorId?: string
  isEstimate?: boolean
}
```

### New Category Types
```typescript
// frontend/src/types/category.ts - CREATE NEW FILE

export interface Category {
  id: string
  name: string
  description?: string
  budgetCodes: string[]
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateCategoryRequest {
  name: string
  description?: string
  budgetCodes?: string[]
}

export interface UpdateCategoryRequest {
  name?: string
  description?: string
  budgetCodes?: string[]
}
```

### New Analytics Types
```typescript
// frontend/src/types/analytics.ts - CREATE NEW FILE

export interface StatusCounts {
  draft: number
  pending: number
  approved: number
  rejected: number
}

export interface RejectionsOverTime {
  date: string
  rejections: number
  total: number
  rate: number
}

export interface RejectionReason {
  reason: string
  count: number
  percentage: number
}

export interface ApproverStats {
  approverId: string
  approverName: string
  rejections: number
  approvals: number
  rejectionRate: number
}

export interface RequisitionMetrics {
  statusCounts: StatusCounts
  rejectionRate: number
  rejectionsOverTime: RejectionsOverTime[]
  rejectionReasons: RejectionReason[]
  topRejectingApprovers: ApproverStats[]
  totalRequisitions: number
}
```

---

## 🔧 State Management Updates

### React Query Configuration
```typescript
// frontend/src/lib/query-client.ts

import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000,   // 10 minutes
    },
  },
})

// ADD NEW QUERY KEYS:
export const queryKeys = {
  // Existing...
  categories: {
    all: ['categories'],
    detail: (id: string) => ['categories', id],
    budgetCodes: (id: string) => ['categories', id, 'budget-codes'],
  },
  analytics: {
    metrics: ['analytics', 'metrics'],
    approvals: ['analytics', 'approvals'],
    dashboard: ['analytics', 'dashboard'],
  },
}
```

### Custom Hooks for Phase 2

#### useCategories Hook
```typescript
// frontend/src/hooks/use-categories.ts

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await fetch('/api/categories')
      if (!res.ok) throw new Error('Failed to fetch categories')
      return res.json()
    },
  })
}

export function useCreateCategory() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateCategoryRequest) => {
      const res = await fetch('/api/categories', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      })
      if (!res.ok) throw new Error('Failed to create category')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] })
    },
  })
}
```

#### useAnalytics Hook
```typescript
// frontend/src/hooks/use-analytics.ts

import { useQuery } from '@tanstack/react-query'

interface AnalyticsParams {
  startDate?: string
  endDate?: string
  period?: 'daily' | 'weekly' | 'monthly'
  department?: string
}

export function useAnalytics(params?: AnalyticsParams) {
  const queryParams = new URLSearchParams()
  if (params?.startDate) queryParams.append('start_date', params.startDate)
  if (params?.endDate) queryParams.append('end_date', params.endDate)
  if (params?.period) queryParams.append('period', params.period)
  if (params?.department) queryParams.append('department', params.department)

  return useQuery({
    queryKey: ['analytics:metrics', queryParams.toString()],
    queryFn: async () => {
      const res = await fetch(
        `/api/analytics/requisitions/metrics?${queryParams}`
      )
      if (!res.ok) throw new Error('Failed to fetch analytics')
      return res.json()
    },
  })
}
```

---

## 🎯 Implementation Checklist

### Phase 1: Type Definitions & API Setup (Day 1)
- [ ] Create `frontend/src/types/category.ts`
- [ ] Create `frontend/src/types/analytics.ts`
- [ ] Update `frontend/src/types/requisition.ts` with new fields
- [ ] Update API config with new endpoints
- [ ] Create `frontend/src/hooks/use-categories.ts`
- [ ] Create `frontend/src/hooks/use-analytics.ts`

### Phase 2: Category Management (Day 2)
- [ ] Create `CategorySelect` component
- [ ] Create category management page
- [ ] Create `CategoryForm` component
- [ ] Create `BudgetCodeManager` component
- [ ] Create category table component
- [ ] Test category CRUD operations

### Phase 3: Requisition Enhancements (Day 3)
- [ ] Update requisition create form with new fields
- [ ] Update requisition detail view
- [ ] Update requisition list view to show new fields
- [ ] Create vendor select component
- [ ] Test requisition creation with new fields
- [ ] Test requisition updates

### Phase 4: Analytics Dashboard (Day 4)
- [ ] Create analytics page
- [ ] Create metrics cards component
- [ ] Create rejection chart component
- [ ] Create rejection reasons chart component
- [ ] Create top approvers table component
- [ ] Add date range and department filters
- [ ] Test all analytics endpoints

### Phase 5: User Activity & Polish (Day 5)
- [ ] Update user profile to show last login
- [ ] Update login response handling
- [ ] Add loading states to all components
- [ ] Add error handling to all API calls
- [ ] Test end-to-end workflows
- [ ] Performance optimization

---

## 🔌 API Request Examples

### Create Category
```typescript
const response = await fetch('/api/categories', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    name: 'Office Supplies',
    description: 'General office supplies',
    budgetCodes: ['BDG-001', 'BDG-002'],
  }),
})

// Response:
{
  "success": true,
  "data": {
    "id": "uuid...",
    "name": "Office Supplies",
    "description": "General office supplies",
    "budgetCodes": ["BDG-001", "BDG-002"],
    "active": true,
    "createdAt": "2025-12-24T10:30:00Z",
    "updatedAt": "2025-12-24T10:30:00Z"
  }
}
```

### Create Requisition with New Fields
```typescript
const response = await fetch('/api/requisitions', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    // ... existing fields ...
    categoryId: 'uuid-of-category',
    preferredVendorId: 'uuid-of-vendor',
    isEstimate: true,
  }),
})

// Response includes:
{
  "success": true,
  "data": {
    // ... all requisition fields ...
    "categoryId": "uuid-of-category",
    "categoryName": "Office Supplies",
    "preferredVendorId": "uuid-of-vendor",
    "preferredVendorName": "Vendor Name",
    "isEstimate": true,
  }
}
```

### Get Analytics
```typescript
const params = new URLSearchParams({
  start_date: '2025-12-01',
  end_date: '2025-12-31',
  period: 'daily',
  department: 'Finance',
})

const response = await fetch(
  `/api/analytics/requisitions/metrics?${params}`,
  {
    headers: { 'Authorization': `Bearer ${token}` },
  }
)

// Response:
{
  "success": true,
  "data": {
    "statusCounts": {
      "draft": 5,
      "pending": 3,
      "approved": 10,
      "rejected": 2
    },
    "rejectionRate": 11.76,
    "rejectionsOverTime": [...],
    "rejectionReasons": [...],
    "topRejectingApprovers": [...],
    "totalRequisitions": 20
  }
}
```

---

## 📱 UI Component Requirements

### Required UI Updates

#### 1. Category Selector
- Dropdown/combobox with search
- Show category name
- Show budget codes as badge
- Load from API on mount
- Handle loading/error states

#### 2. Estimate Badge
- Visual indicator on requisition cards
- Warning message in detail view
- Different styling from regular requisitions

#### 3. Supplier/Vendor Select
- Similar to category select
- Multi-select optional enhancement
- Fallback to text input if vendor not found

#### 4. Analytics Cards
- Color-coded status cards
- Large metric display
- Percentage for rejection rate
- Trend indicators (optional)

#### 5. Charts
- Line chart for rejections over time
- Bar chart for rejection reasons
- Table for approver statistics
- Date range filtering

---

## 🚀 Deployment Checklist

### Before Going Live
- [ ] All new components tested in development
- [ ] API endpoints tested with Postman collection
- [ ] TypeScript compilation passes without errors
- [ ] No console errors in browser dev tools
- [ ] Responsive design tested on mobile
- [ ] Loading states working correctly
- [ ] Error messages user-friendly
- [ ] Analytics data accurate
- [ ] Database migration has run
- [ ] Environment variables set

### Performance Considerations
- [ ] Lazy load analytics components
- [ ] Cache category list (5 min TTL)
- [ ] Pagination on large category lists
- [ ] Debounce search inputs
- [ ] Consider virtualization for large tables

---

## 🔐 Security Considerations

- All API calls require authentication token
- Validate user permissions before displaying category/analytics data
- Sanitize user input in category names
- Prevent SQL injection via parameterized queries (backend handles)
- CORS headers properly configured

---

## 📞 Support & Troubleshooting

### Common Issues

**Q: Category dropdown shows no options**
- Verify categories API endpoint returns data
- Check Authorization header is included
- Verify user has permission to view categories

**Q: Analytics shows no data**
- Ensure requisitions exist with various statuses
- Check date range includes requisition dates
- Verify analytics service is running

**Q: Estimate flag not saving**
- Confirm `isEstimate` field is in request body
- Check API response includes the field
- Verify frontend is sending boolean not string

---

## 📚 Related Documentation

- [QUICK-START.md](QUICK-START.md) - Backend setup
- [TESTING-GUIDE.md](TESTING-GUIDE.md) - API testing
- [postman-collection.json](postman-collection.json) - Pre-configured requests

---

**Status:** Ready for Frontend Implementation
**Last Updated:** December 24, 2025
**Questions?** Check the TESTING-GUIDE.md for API examples
