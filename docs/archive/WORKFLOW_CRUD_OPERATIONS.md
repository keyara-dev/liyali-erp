# Complete Workflow CRUD Operations Reference

This document outlines all available CRUD operations for the workflow system, now fully implemented with localStorage persistence.

## Quick Start

```typescript
import {
  getAllWorkflows,
  getWorkflowById,
  saveWorkflow,
  deleteWorkflow,
  duplicateWorkflow,
  updateWorkflowStatus,
  cloneWorkflowAsTemplate,
  exportWorkflows,
  importWorkflows,
  getWorkflowsByDocumentType,
  searchWorkflows,
  clearAllWorkflows,
  resetToMockData,
} from '@/lib/workflow-storage'
```

---

## CREATE Operations

### 1. **Save Workflow** (Create or Update)
```typescript
const workflow = saveWorkflow({
  id: 'wf-123',
  name: 'Requisition Approval',
  description: '4-stage approval process',
  documentType: 'REQUISITION',
  stages: 4,
  status: 'ACTIVE',
  updatedAt: new Date().toISOString(),
  fullData: {
    name: 'Requisition Approval',
    description: '4-stage approval process',
    documentType: 'REQUISITION',
    isDefault: true,
    stages: [
      {
        id: 'stage-1',
        order: 1,
        name: 'Department Manager Review',
        description: 'Initial review',
        approverRole: 'DEPARTMENT_MANAGER',
        requiredApprovals: 1,
        canReject: true,
        canReassign: true,
      },
      // ... more stages
    ],
  },
})

// Returns: StoredWorkflow | null
```

**When to use:** Creating new workflows or updating existing ones
**Stored in:** Browser localStorage under `liyali_workflows` key

---

## READ Operations

### 1. **Get All Workflows**
```typescript
const allWorkflows = getAllWorkflows()

// Returns: StoredWorkflow[]
// {
//   id: 'wf-1',
//   name: 'Standard Requisition Approval',
//   description: '4-stage approval process for purchase requisitions',
//   documentType: 'REQUISITION',
//   stages: 4,
//   status: 'ACTIVE',
//   createdAt: '2024-01-15T10:30:00Z',
//   updatedAt: '2024-11-20T14:22:00Z',
//   createdBy: 'admin@example.com',
//   fullData: { ... }
// }
```

### 2. **Get Workflow by ID**
```typescript
const workflow = getWorkflowById('wf-1')

// Returns: StoredWorkflow | null
```

### 3. **Get Workflows by Document Type**
```typescript
const poWorkflows = getWorkflowsByDocumentType('PURCHASE_ORDER')

// Returns: StoredWorkflow[] (only ACTIVE workflows)
```

### 4. **Search Workflows**
```typescript
const results = searchWorkflows('approval')

// Searches in: name, description, documentType
// Returns: StoredWorkflow[]
```

---

## UPDATE Operations

### 1. **Update Workflow** (Full Update)
```typescript
const updated = saveWorkflow({
  id: 'wf-1', // Keep same ID to update
  name: 'Updated Requisition Approval',
  description: 'Updated description',
  documentType: 'REQUISITION',
  stages: 5, // Changed from 4
  status: 'ACTIVE',
  updatedAt: new Date().toISOString(),
  fullData: {
    // ... updated data
  },
})

// Returns: StoredWorkflow | null
```

### 2. **Update Workflow Status** (ACTIVE/DEPRECATED)
```typescript
const deprecated = updateWorkflowStatus('wf-1', 'DEPRECATED')

// Returns: StoredWorkflow | null
```

---

## DELETE Operations

### 1. **Delete Workflow** (Hard Delete)
```typescript
const success = deleteWorkflow('wf-1')

// Returns: boolean (true if deleted, false if error)
// Removes workflow completely from localStorage
```

### 2. **Clear All Workflows**
```typescript
const success = clearAllWorkflows()

// Returns: boolean
// ⚠️ WARNING: This removes ALL workflows!
```

---

## SPECIAL Operations

### 1. **Duplicate Workflow**
```typescript
const copy = duplicateWorkflow('wf-1')

// Returns: StoredWorkflow | null
// Creates: New workflow with "(Copy)" suffix
// ID: Auto-generated new ID (wf-{timestamp})
// createdAt/updatedAt: Set to current time
```

**Example Result:**
```
Original: "Standard Requisition Approval" (wf-1)
Duplicate: "Standard Requisition Approval (Copy)" (wf-1733000000000)
```

### 2. **Clone Workflow as Template**
```typescript
const template = cloneWorkflowAsTemplate('wf-1', 'My Custom Template')

// Returns: StoredWorkflow | null
// Creates: New workflow with custom name
// Additionally: Regenerates all stage IDs for independence
// Sets: isDefault = false
```

### 3. **Export Workflows** (JSON)
```typescript
const jsonString = exportWorkflows()

// Returns: JSON string of all workflows
// Usage: Download or backup workflows
// Example:
// const element = document.createElement('a')
// element.href = 'data:text/json;charset=utf-8,' + encodeURIComponent(jsonString)
// element.download = 'workflows.json'
// element.click()
```

### 4. **Import Workflows** (JSON)
```typescript
const result = importWorkflows(jsonString)

// Returns: { success: boolean; count: number }
// - If workflow exists: Updates it
// - If workflow is new: Adds it
// - Preserves IDs from import data
```

**Example:**
```typescript
const fileInput = document.getElementById('file-input') as HTMLInputElement
const file = fileInput.files?.[0]

if (file) {
  const text = await file.text()
  const result = importWorkflows(text)
  console.log(`Imported ${result.count} workflows`)
}
```

### 5. **Reset to Mock Data**
```typescript
const success = resetToMockData()

// Returns: boolean
// Resets localStorage to 4 default workflows:
// 1. Standard Requisition Approval
// 2. Purchase Order Approval
// 3. Payment Voucher Review
// 4. GRN Confirmation Flow
// ⚠️ WARNING: Overwrites all current workflows!
```

---

## Data Structure

### StoredWorkflow Interface
```typescript
interface StoredWorkflow {
  id: string // 'wf-123'
  name: string // 'Standard Requisition Approval'
  description: string // '4-stage approval process...'
  documentType: string // 'REQUISITION' | 'PURCHASE_ORDER' | etc.
  stages: number // Count of approval stages
  status: 'ACTIVE' | 'DEPRECATED'
  createdAt: string // ISO timestamp
  updatedAt: string // ISO timestamp
  createdBy: string // User email
  fullData: {
    name: string
    description: string
    documentType: string
    isDefault: boolean
    stages: WorkflowStage[]
  }
}

interface WorkflowStage {
  id: string
  order: number
  name: string
  description: string
  approverRole: string // 'DEPARTMENT_MANAGER' | 'FINANCE_OFFICER' | etc.
  requiredApprovals: number
  canReject: boolean
  canReassign: boolean
}
```

---

## Usage Examples

### Example 1: Create and Save New Workflow
```typescript
import { saveWorkflow } from '@/lib/workflow-storage'

const newWorkflow = saveWorkflow({
  id: `wf-${Date.now()}`,
  name: 'Budget Approval Flow',
  description: 'Multi-level budget review process',
  documentType: 'BUDGET',
  stages: 3,
  status: 'ACTIVE',
  updatedAt: new Date().toISOString(),
  fullData: {
    name: 'Budget Approval Flow',
    description: 'Multi-level budget review process',
    documentType: 'BUDGET',
    isDefault: false,
    stages: [
      {
        id: 'stage-1',
        order: 1,
        name: 'Department Head Review',
        description: 'Department head review',
        approverRole: 'DEPARTMENT_MANAGER',
        requiredApprovals: 1,
        canReject: true,
        canReassign: true,
      },
      {
        id: 'stage-2',
        order: 2,
        name: 'Finance Review',
        description: 'Finance officer review',
        approverRole: 'FINANCE_OFFICER',
        requiredApprovals: 1,
        canReject: true,
        canReassign: true,
      },
      {
        id: 'stage-3',
        order: 3,
        name: 'CFO Approval',
        description: 'Final CFO approval',
        approverRole: 'CFO',
        requiredApprovals: 1,
        canReject: false,
        canReassign: false,
      },
    ],
  },
})

if (newWorkflow) {
  console.log('Workflow created:', newWorkflow.name)
}
```

### Example 2: Search and Update
```typescript
import { searchWorkflows, saveWorkflow, updateWorkflowStatus } from '@/lib/workflow-storage'

// Search for requisition workflows
const requisitionWorkflows = searchWorkflows('requisition')

// Deprecate old ones
for (const workflow of requisitionWorkflows) {
  if (workflow.name.includes('Old')) {
    updateWorkflowStatus(workflow.id, 'DEPRECATED')
  }
}
```

### Example 3: Export and Backup
```typescript
import { exportWorkflows } from '@/lib/workflow-storage'

const backupData = exportWorkflows()
const backupFile = new Blob([backupData], { type: 'application/json' })
const url = URL.createObjectURL(backupFile)
const link = document.createElement('a')
link.href = url
link.download = `workflows-backup-${new Date().toISOString()}.json`
link.click()
```

### Example 4: Filter by Document Type
```typescript
import { getWorkflowsByDocumentType } from '@/lib/workflow-storage'

// Get all active payment voucher workflows
const pvWorkflows = getWorkflowsByDocumentType('PAYMENT_VOUCHER')

console.log(`Found ${pvWorkflows.length} payment voucher workflows`)
```

---

## Operations Matrix

| Operation | Function | Input | Output | Side Effect |
|-----------|----------|-------|--------|-------------|
| Create | `saveWorkflow()` | WorkflowData | StoredWorkflow \| null | Saves to localStorage |
| Read All | `getAllWorkflows()` | - | StoredWorkflow[] | Initializes with mock data if empty |
| Read One | `getWorkflowById()` | ID | StoredWorkflow \| null | - |
| Read Filtered | `getWorkflowsByDocumentType()` | Type | StoredWorkflow[] | - |
| Search | `searchWorkflows()` | Query | StoredWorkflow[] | - |
| Update Full | `saveWorkflow()` (with existing ID) | WorkflowData | StoredWorkflow \| null | Updates in localStorage |
| Update Status | `updateWorkflowStatus()` | ID, Status | StoredWorkflow \| null | Updates in localStorage |
| Delete | `deleteWorkflow()` | ID | boolean | Removes from localStorage |
| Duplicate | `duplicateWorkflow()` | ID | StoredWorkflow \| null | Creates new workflow |
| Clone Template | `cloneWorkflowAsTemplate()` | ID, Name | StoredWorkflow \| null | Creates new with fresh IDs |
| Export | `exportWorkflows()` | - | JSON string | - |
| Import | `importWorkflows()` | JSON string | { success, count } | Merges into localStorage |
| Clear All | `clearAllWorkflows()` | - | boolean | ⚠️ Deletes all workflows |
| Reset | `resetToMockData()` | - | boolean | ⚠️ Overwrites with defaults |

---

## Error Handling

All functions include try-catch blocks and return `null` or `false` on error. Check browser console for detailed error messages.

```typescript
const workflow = getWorkflowById('invalid-id')

if (!workflow) {
  console.log('Workflow not found')
}
```

---

## localStorage Storage Key

All workflows are stored under the key: `liyali_workflows`

To inspect workflows in browser DevTools:
```javascript
localStorage.getItem('liyali_workflows')
```

---

## Limitations

- **Browser Storage**: Only persistent in same browser/device
- **Size Limit**: Depends on browser (typically 5-10MB)
- **No Transactions**: Multiple operations could lead to inconsistencies
- **No Concurrent Editing**: Real-time sync not supported

---

## Roadmap

Future enhancements planned:
- [ ] Database backend (PostgreSQL)
- [ ] Real-time sync with WebSockets
- [ ] Workflow versioning
- [ ] SLA enforcement
- [ ] Concurrent approvals
- [ ] Advanced validations
- [ ] Audit trail

---

## See Also

- [Workflow Builder Component](src/app/(private)/admin/workflows/_components/workflow-builder.tsx)
- [Workflows Client Component](src/app/(private)/admin/workflows/_components/workflows-client.tsx)
- [Create Workflow Page](src/app/(private)/admin/workflows/create/)
- [Edit Workflow Page](src/app/(private)/admin/workflows/[id]/edit/)
