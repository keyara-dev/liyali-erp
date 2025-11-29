# Workflow System - Component Integration Examples

This document provides examples of how to integrate the mocked server actions into React components.

---

## Table of Contents

1. [Form Components](#form-components)
2. [Approval Components](#approval-components)
3. [Dashboard Components](#dashboard-components)
4. [Admin Components](#admin-components)

---

## Form Components

### Purchase Order Form Component

```typescript
'use client'

import { useState } from 'react'
import { createWorkflowDocument, submitDocument, updateDocumentDraft } from '@/app/_actions/workflow'
import { PurchaseOrder } from '@/types/workflow'

interface PurchaseOrderFormProps {
  documentId?: string
  initialData?: PurchaseOrder
}

export function PurchaseOrderForm({ documentId, initialData }: PurchaseOrderFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    vendorName: initialData?.metadata.vendorName || '',
    vendorId: initialData?.metadata.vendorId || '',
    items: initialData?.metadata.items || [],
    deliveryDate: initialData?.metadata.deliveryDate || '',
    specialInstructions: initialData?.metadata.specialInstructions || '',
  })

  const totalAmount = formData.items.reduce((sum, item) => sum + item.totalCost, 0)

  const handleAddItem = () => {
    setFormData({
      ...formData,
      items: [
        ...formData.items,
        {
          id: Date.now().toString(),
          description: '',
          quantity: 1,
          unitCost: 0,
          totalCost: 0,
        },
      ],
    })
  }

  const handleRemoveItem = (itemId: string) => {
    setFormData({
      ...formData,
      items: formData.items.filter((item) => item.id !== itemId),
    })
  }

  const handleUpdateItem = (itemId: string, field: string, value: any) => {
    setFormData({
      ...formData,
      items: formData.items.map((item) => {
        if (item.id === itemId) {
          const updated = { ...item, [field]: value }
          if (field === 'quantity' || field === 'unitCost') {
            updated.totalCost = updated.quantity * updated.unitCost
          }
          return updated
        }
        return item
      }),
    })
  }

  const handleSaveDraft = async () => {
    setIsLoading(true)
    try {
      if (documentId) {
        await updateDocumentDraft(documentId, {
          ...formData,
          totalAmount,
        })
      } else {
        const result = await createWorkflowDocument('PURCHASE_ORDER', {
          ...formData,
          totalAmount,
          currency: 'ZMW',
        })
        // Redirect to document detail page
        // window.location.href = `/workflows/purchase-orders/${result.data.id}`
      }
    } finally {
      setIsLoading(false)
    }
  }

  const handleSubmit = async () => {
    if (!documentId) return

    setIsLoading(true)
    try {
      await submitDocument(documentId)
      // Show success notification
      // Redirect to document detail page
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto p-6 bg-white rounded-lg shadow">
      <h1 className="text-3xl font-bold mb-6">Purchase Order</h1>

      <div className="grid grid-cols-2 gap-4 mb-6">
        <div>
          <label className="block font-medium mb-2">Vendor Name</label>
          <input
            type="text"
            value={formData.vendorName}
            onChange={(e) => setFormData({ ...formData, vendorName: e.target.value })}
            className="w-full border rounded px-3 py-2"
          />
        </div>
        <div>
          <label className="block font-medium mb-2">Vendor ID</label>
          <input
            type="text"
            value={formData.vendorId}
            onChange={(e) => setFormData({ ...formData, vendorId: e.target.value })}
            className="w-full border rounded px-3 py-2"
          />
        </div>
      </div>

      <div className="mb-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Items</h2>
          <button
            onClick={handleAddItem}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            + Add Item
          </button>
        </div>

        {formData.items.map((item) => (
          <div key={item.id} className="border p-4 mb-3 rounded">
            <div className="grid grid-cols-4 gap-2">
              <input
                type="text"
                placeholder="Description"
                value={item.description}
                onChange={(e) => handleUpdateItem(item.id, 'description', e.target.value)}
                className="border rounded px-2 py-1"
              />
              <input
                type="number"
                placeholder="Qty"
                value={item.quantity}
                onChange={(e) => handleUpdateItem(item.id, 'quantity', parseFloat(e.target.value))}
                className="border rounded px-2 py-1"
              />
              <input
                type="number"
                placeholder="Unit Cost"
                value={item.unitCost}
                onChange={(e) => handleUpdateItem(item.id, 'unitCost', parseFloat(e.target.value))}
                className="border rounded px-2 py-1"
              />
              <input
                type="text"
                disabled
                value={`ZMW ${item.totalCost.toFixed(2)}`}
                className="border rounded px-2 py-1 bg-gray-100"
              />
              <button
                onClick={() => handleRemoveItem(item.id)}
                className="bg-red-500 text-white px-2 py-1 rounded hover:bg-red-600 col-span-1"
              >
                Remove
              </button>
            </div>
          </div>
        ))}
      </div>

      <div className="mb-6 p-3 bg-gray-100 rounded">
        <p className="text-lg font-semibold">
          Total Amount: ZMW {totalAmount.toFixed(2)}
        </p>
      </div>

      <div className="flex gap-3">
        <button
          onClick={handleSaveDraft}
          disabled={isLoading}
          className="bg-gray-500 text-white px-6 py-2 rounded hover:bg-gray-600 disabled:opacity-50"
        >
          {documentId ? 'Update Draft' : 'Save Draft'}
        </button>
        <button
          onClick={handleSubmit}
          disabled={isLoading || !documentId}
          className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
        >
          Submit for Approval
        </button>
      </div>
    </div>
  )
}
```

---

## Approval Components

### Approval Actions Panel

```typescript
'use client'

import { useState } from 'react'
import { approveDocument, rejectDocument } from '@/app/_actions/workflow'
import { WorkflowDocument } from '@/types/workflow'

interface ApprovalActionsPanelProps {
  document: WorkflowDocument
  onApprovalComplete?: () => void
}

export function ApprovalActionsPanel({
  document,
  onApprovalComplete,
}: ApprovalActionsPanelProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [showApprovalForm, setShowApprovalForm] = useState(false)
  const [comments, setComments] = useState('')

  if (document.status !== 'IN_APPROVAL') {
    return (
      <div className="p-4 bg-gray-100 rounded text-center">
        <p className="text-gray-600">
          This document is not pending approval (Status: {document.status})
        </p>
      </div>
    )
  }

  const handleApprove = async () => {
    setIsLoading(true)
    try {
      const result = await approveDocument(document.id, comments)
      if (result.success) {
        // Show success notification
        setComments('')
        setShowApprovalForm(false)
        onApprovalComplete?.()
      }
    } finally {
      setIsLoading(false)
    }
  }

  const handleReject = async () => {
    if (!comments.trim()) {
      alert('Please provide a reason for rejection')
      return
    }

    setIsLoading(true)
    try {
      const result = await rejectDocument(document.id, comments)
      if (result.success) {
        // Show success notification
        setComments('')
        setShowApprovalForm(false)
        onApprovalComplete?.()
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto p-6 bg-blue-50 border border-blue-200 rounded-lg">
      <h3 className="text-lg font-bold mb-4">Approval Required</h3>

      {showApprovalForm ? (
        <div className="space-y-4">
          <textarea
            placeholder="Add approval comments (optional for approval, required for rejection)"
            value={comments}
            onChange={(e) => setComments(e.target.value)}
            rows={4}
            className="w-full border rounded px-3 py-2"
          />

          <div className="flex gap-2">
            <button
              onClick={handleApprove}
              disabled={isLoading}
              className="flex-1 bg-green-500 text-white py-2 rounded hover:bg-green-600 disabled:opacity-50"
            >
              Approve
            </button>
            <button
              onClick={handleReject}
              disabled={isLoading}
              className="flex-1 bg-red-500 text-white py-2 rounded hover:bg-red-600 disabled:opacity-50"
            >
              Reject
            </button>
            <button
              onClick={() => {
                setShowApprovalForm(false)
                setComments('')
              }}
              disabled={isLoading}
              className="flex-1 bg-gray-500 text-white py-2 rounded hover:bg-gray-600"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <div className="flex gap-2">
          <button
            onClick={() => setShowApprovalForm(true)}
            className="flex-1 bg-blue-500 text-white py-2 rounded hover:bg-blue-600"
          >
            Review for Approval
          </button>
        </div>
      )}
    </div>
  )
}
```

### Approval History Component

```typescript
'use client'

import { useEffect, useState } from 'react'
import { getAuditLog } from '@/app/_actions/workflow'
import { ApprovalLogEntry } from '@/types/workflow'

interface ApprovalHistoryProps {
  documentId: string
}

export function ApprovalHistory({ documentId }: ApprovalHistoryProps) {
  const [logs, setLogs] = useState<ApprovalLogEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    async function fetchLogs() {
      const result = await getAuditLog(documentId)
      if (result.success) {
        setLogs(result.data || [])
      }
      setIsLoading(false)
    }
    fetchLogs()
  }, [documentId])

  if (isLoading) {
    return <div className="text-center py-8">Loading approval history...</div>
  }

  if (logs.length === 0) {
    return <div className="text-center py-8 text-gray-500">No approval history yet</div>
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h2 className="text-2xl font-bold mb-6">Approval History</h2>

      <div className="space-y-4">
        {logs.map((log, index) => (
          <div key={log.id} className="border-l-4 border-blue-500 pl-4 py-2">
            <div className="flex items-start justify-between">
              <div>
                <p className="font-semibold">
                  {log.action} by {log.approver.name}
                </p>
                <p className="text-sm text-gray-600">
                  {new Date(log.timestamp).toLocaleString()}
                </p>
                {log.comments && (
                  <p className="mt-2 text-gray-700 italic">"{log.comments}"</p>
                )}
              </div>
              <span
                className={`px-3 py-1 rounded text-sm font-medium ${
                  log.action === 'APPROVED'
                    ? 'bg-green-100 text-green-800'
                    : log.action === 'REJECTED'
                    ? 'bg-red-100 text-red-800'
                    : 'bg-blue-100 text-blue-800'
                }`}
              >
                {log.action}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
```

---

## Dashboard Components

### Pending Approvals Dashboard

```typescript
'use client'

import { useEffect, useState } from 'react'
import { getPendingApprovals } from '@/app/_actions/workflow'
import { WorkflowDocument } from '@/types/workflow'

interface PendingApprovalsProps {
  userRole: string
}

export function PendingApprovalsDashboard({ userRole }: PendingApprovalsProps) {
  const [documents, setDocuments] = useState<WorkflowDocument[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    async function fetchPendingApprovals() {
      const result = await getPendingApprovals(userRole)
      if (result.success) {
        setDocuments(result.data || [])
      }
      setIsLoading(false)
    }
    fetchPendingApprovals()
  }, [userRole])

  if (isLoading) {
    return <div className="text-center py-8">Loading pending approvals...</div>
  }

  if (documents.length === 0) {
    return (
      <div className="text-center py-8 bg-gray-50 rounded">
        <p className="text-gray-600">No pending approvals for your role</p>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h2 className="text-2xl font-bold mb-6">
        Pending Approvals ({documents.length})
      </h2>

      <div className="space-y-4">
        {documents.map((doc) => (
          <div
            key={doc.id}
            className="border rounded-lg p-4 hover:shadow-lg transition"
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <h3 className="font-bold text-lg">{doc.documentNumber}</h3>
                <p className="text-gray-600">
                  Type: {doc.type.replace('_', ' ')}
                </p>
                <p className="text-sm text-gray-500">
                  Created: {new Date(doc.createdAt).toLocaleDateString()}
                </p>
              </div>

              <div className="flex items-center gap-4">
                <div className="text-right">
                  <p className="text-sm text-gray-600">Stage</p>
                  <p className="text-2xl font-bold">{doc.currentStage}</p>
                </div>

                <a
                  href={`/workflows/${doc.type.toLowerCase()}s/${doc.id}`}
                  className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600"
                >
                  Review
                </a>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
```

### Dashboard Stats

```typescript
'use client'

import { useEffect, useState } from 'react'
import { getDashboardStats } from '@/app/_actions/workflow'

interface DashboardStatsProps {
  userId: string
}

export function DashboardStats({ userId }: DashboardStatsProps) {
  const [stats, setStats] = useState<any>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    async function fetchStats() {
      const result = await getDashboardStats(userId)
      if (result.success) {
        setStats(result.data)
      }
      setIsLoading(false)
    }
    fetchStats()
  }, [userId])

  if (isLoading) {
    return <div className="text-center py-8">Loading stats...</div>
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div className="bg-white p-6 rounded-lg shadow">
        <p className="text-gray-600 text-sm font-medium">Created Documents</p>
        <p className="text-3xl font-bold">{stats?.createdDocuments ?? 0}</p>
      </div>

      <div className="bg-white p-6 rounded-lg shadow">
        <p className="text-gray-600 text-sm font-medium">Pending Approvals</p>
        <p className="text-3xl font-bold text-orange-500">
          {stats?.pendingApprovals ?? 0}
        </p>
      </div>

      <div className="bg-white p-6 rounded-lg shadow">
        <p className="text-gray-600 text-sm font-medium">Approved</p>
        <p className="text-3xl font-bold text-green-500">
          {stats?.approvedDocuments ?? 0}
        </p>
      </div>

      <div className="bg-white p-6 rounded-lg shadow">
        <p className="text-gray-600 text-sm font-medium">Rejected</p>
        <p className="text-3xl font-bold text-red-500">
          {stats?.rejectedDocuments ?? 0}
        </p>
      </div>
    </div>
  )
}
```

---

## Admin Components

### Role Management Component

```typescript
'use client'

import { useEffect, useState } from 'react'
import {
  getAllRoles,
  createRole,
  getAllPermissions,
} from '@/app/_actions/rbac'
import { CustomRole, Permission } from '@/lib/rbac'

export function RoleManagement() {
  const [roles, setRoles] = useState<CustomRole[]>([])
  const [permissions, setPermissions] = useState<
    Array<{ name: Permission; description: string }>
  >([])
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newRoleName, setNewRoleName] = useState('')
  const [newRoleDescription, setNewRoleDescription] = useState('')
  const [selectedPermissions, setSelectedPermissions] = useState<Permission[]>([])

  useEffect(() => {
    async function fetchData() {
      const rolesResult = await getAllRoles()
      const permissionsResult = await getAllPermissions()

      if (rolesResult.success) {
        setRoles(rolesResult.data || [])
      }
      if (permissionsResult.success) {
        setPermissions(permissionsResult.data || [])
      }
      setIsLoading(false)
    }
    fetchData()
  }, [])

  const handleCreateRole = async () => {
    if (!newRoleName.trim() || selectedPermissions.length === 0) {
      alert('Please enter role name and select at least one permission')
      return
    }

    const result = await createRole(
      newRoleName,
      newRoleDescription,
      selectedPermissions
    )

    if (result.success) {
      setRoles([...roles, result.data])
      setNewRoleName('')
      setNewRoleDescription('')
      setSelectedPermissions([])
      setShowCreateForm(false)
    }
  }

  const togglePermission = (permission: Permission) => {
    setSelectedPermissions((prev) =>
      prev.includes(permission)
        ? prev.filter((p) => p !== permission)
        : [...prev, permission]
    )
  }

  if (isLoading) {
    return <div className="text-center py-8">Loading...</div>
  }

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Role Management</h2>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Create New Role
        </button>
      </div>

      {showCreateForm && (
        <div className="mb-6 p-4 bg-gray-50 rounded-lg border">
          <h3 className="font-bold mb-4">Create New Role</h3>

          <div className="mb-4">
            <label className="block font-medium mb-1">Role Name</label>
            <input
              type="text"
              value={newRoleName}
              onChange={(e) => setNewRoleName(e.target.value)}
              placeholder="e.g., Senior Approver"
              className="w-full border rounded px-3 py-2"
            />
          </div>

          <div className="mb-4">
            <label className="block font-medium mb-1">Description</label>
            <textarea
              value={newRoleDescription}
              onChange={(e) => setNewRoleDescription(e.target.value)}
              placeholder="Role description"
              rows={2}
              className="w-full border rounded px-3 py-2"
            />
          </div>

          <div className="mb-4">
            <label className="block font-medium mb-2">Permissions</label>
            <div className="grid grid-cols-2 gap-2">
              {permissions.map((perm) => (
                <label key={perm.name} className="flex items-center">
                  <input
                    type="checkbox"
                    checked={selectedPermissions.includes(perm.name)}
                    onChange={() => togglePermission(perm.name)}
                    className="mr-2"
                  />
                  <span className="text-sm">{perm.description}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={handleCreateRole}
              className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
            >
              Create Role
            </button>
            <button
              onClick={() => setShowCreateForm(false)}
              className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <div className="space-y-4">
        {roles.map((role) => (
          <div key={role.id} className="border rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-bold text-lg">{role.name}</h3>
                <p className="text-gray-600 text-sm mb-3">{role.description}</p>
                <div className="flex flex-wrap gap-1">
                  {role.permissions.map((perm) => (
                    <span
                      key={perm}
                      className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded"
                    >
                      {perm}
                    </span>
                  ))}
                </div>
              </div>
              {!role.isBuiltIn && (
                <div className="flex gap-2">
                  <button className="text-blue-500 hover:text-blue-700">
                    Edit
                  </button>
                  <button className="text-red-500 hover:text-red-700">
                    Delete
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
```

---

## Usage Pattern Summary

### In a `'use client'` Component:

```typescript
'use client'

import { useState } from 'react'
import { someServerAction } from '@/app/_actions/workflow'

export function MyComponent() {
  const [isLoading, setIsLoading] = useState(false)

  const handleAction = async () => {
    setIsLoading(true)
    try {
      const result = await someServerAction(params)

      if (result.success) {
        // Handle success
        console.log(result.data)
      } else {
        // Handle error
        console.error(result.message)
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    // JSX here
  )
}
```

---

## Key Points

✅ All server actions work with NextAuth session
✅ Responses follow consistent format
✅ Always check `result.success` before using data
✅ Use `'use client'` directive for interactive components
✅ Handle loading and error states
✅ Data persists during session, resets on server restart

---

## Files Ready for UI Development

These components can now be built using the mocked server actions:

- Dashboard (stats, pending approvals, submitted documents)
- Form pages (Purchase Order, Payment Voucher, Requisition)
- Document detail page (view, approve, reject, attachments)
- Approval workflow page (approve/reject actions)
- Role management page (admin only)
- User management page (admin only)
- Audit report page (compliance view)

All mocked data is ready to use!
