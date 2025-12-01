'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ArrowLeft, AlertCircle, CheckCircle2, XCircle } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { ApprovalActionPanel } from '@/components/workflows/approval-action-panel'
import { POItemsTable } from '../_components/po-items-table'

interface POApprovalClientProps {
  poId: string
  userId: string
  userRole: string
}

interface POItem {
  id: string
  itemNumber: number
  description: string
  quantity: number
  unitPrice: number
  totalPrice: number
  unit: string
  expectedDelivery?: string
}

interface PurchaseOrder {
  id: string
  poNumber: string
  status: 'DRAFT' | 'SUBMITTED' | 'IN_APPROVAL' | 'APPROVED' | 'REJECTED'
  vendor: {
    name: string
    contactPerson: string
    email: string
    phone: string
    address: string
  }
  requestedBy: string
  requestDate: string
  deliveryDate: string
  paymentTerms: string
  items: POItem[]
  subtotal: number
  tax: number
  total: number
  currentStage: number
  stageName: string
  createdAt: string
  updatedAt: string
}

const STAGE_NAMES: Record<number, string> = {
  1: 'Department Manager Review',
  2: 'Finance Officer Review',
  3: 'CFO Approval',
}

// Mock data generator
function generateMockPO(poId: string): PurchaseOrder {
  const currentStage = Math.floor(Math.random() * 3) + 1

  return {
    id: poId,
    poNumber: `PO-2024-${String(Math.floor(Math.random() * 9000) + 1000).padStart(4, '0')}`,
    status: 'IN_APPROVAL',
    vendor: {
      name: 'Global Supplies Inc.',
      contactPerson: 'John Smith',
      email: 'john.smith@globalsupplies.com',
      phone: '+1 (555) 123-4567',
      address: '123 Business Park, NY 10001',
    },
    requestedBy: 'REQ-USER-001',
    requestDate: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
    deliveryDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
    paymentTerms: 'Net 30',
    items: [
      {
        id: 'item-1',
        itemNumber: 1,
        description: 'Office Chairs - Ergonomic',
        quantity: 10,
        unitPrice: 250,
        totalPrice: 2500,
        unit: 'units',
        expectedDelivery: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000).toISOString(),
      },
      {
        id: 'item-2',
        itemNumber: 2,
        description: 'Standing Desks - Electric',
        quantity: 5,
        unitPrice: 800,
        totalPrice: 4000,
        unit: 'units',
        expectedDelivery: new Date(Date.now() + 20 * 24 * 60 * 60 * 1000).toISOString(),
      },
      {
        id: 'item-3',
        itemNumber: 3,
        description: 'Computer Monitors - 27 inch',
        quantity: 8,
        unitPrice: 350,
        totalPrice: 2800,
        unit: 'units',
        expectedDelivery: new Date(Date.now() + 25 * 24 * 60 * 60 * 1000).toISOString(),
      },
    ],
    subtotal: 9300,
    tax: 930,
    total: 10230,
    currentStage,
    stageName: STAGE_NAMES[currentStage],
    createdAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
  }
}

// Convert PO to ApprovalTask format
function convertPOToApprovalTask(po: PurchaseOrder, userId: string) {
  return {
    id: po.id,
    entityId: po.id,
    entityType: 'PURCHASE_ORDER',
    entityNumber: po.poNumber,
    status: 'pending',
    stageName: po.stageName,
    stageIndex: po.currentStage,
    importance: 'MEDIUM',
    approverName: 'Current Approver',
    approverUserId: userId,
    createdAt: po.createdAt,
    actionDate: new Date().toISOString(),
    dueDate: po.deliveryDate,
    workflowId: 'po-workflow-v1',
    workflowName: '3-Stage PO Approval',
  }
}

export function POApprovalClient({
  poId,
  userId,
  userRole,
}: POApprovalClientProps) {
  const router = useRouter()
  const [po, setPO] = useState<PurchaseOrder | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Simulate data loading
    const timer = setTimeout(() => {
      setPO(generateMockPO(poId))
      setIsLoading(false)
    }, 500)

    return () => clearTimeout(timer)
  }, [poId])

  const handleBack = () => {
    router.back()
  }

  if (isLoading || !po) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" size="sm" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <div className="space-y-4">
          <Skeleton className="h-12 w-48" />
          <Skeleton className="h-96 w-full" />
          <Skeleton className="h-96 w-full" />
        </div>
      </div>
    )
  }

  const approvalTask = convertPOToApprovalTask(po, userId)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">{po.poNumber}</h1>
          <p className="text-muted-foreground">Purchase Order Approval</p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        {/* Main Content */}
        <div className="md:col-span-2 space-y-6">
          {/* Vendor Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Vendor Information</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 md:grid-cols-2">
              <div>
                <p className="text-sm text-muted-foreground">Vendor Name</p>
                <p className="font-semibold">{po.vendor.name}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Contact Person</p>
                <p className="font-semibold">{po.vendor.contactPerson}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Email</p>
                <p className="font-semibold text-blue-600">{po.vendor.email}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Phone</p>
                <p className="font-semibold">{po.vendor.phone}</p>
              </div>
            </CardContent>
          </Card>

          {/* Line Items */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Line Items</CardTitle>
            </CardHeader>
            <CardContent>
              <POItemsTable items={po.items} />
            </CardContent>
          </Card>

          {/* Cost Summary */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Cost Summary</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 max-w-xs ml-auto">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Subtotal:</span>
                  <span className="font-semibold">K{po.subtotal.toLocaleString('en-ZM')}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Tax (10%):</span>
                  <span className="font-semibold">K{po.tax.toLocaleString('en-ZM')}</span>
                </div>
                <div className="border-t pt-2 flex justify-between">
                  <span className="font-semibold">Total:</span>
                  <span className="text-lg font-bold text-green-600">
                    K{po.total.toLocaleString('en-ZM')}
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Approval Panel */}
        <div>
          <ApprovalActionPanel
            task={approvalTask}
            onSuccess={() => {
              toast.success('Purchase order approved successfully')
              router.push('/workflows/purchase-orders')
            }}
            onError={(error) => {
              toast.error(error || 'Failed to process approval')
            }}
          />
        </div>
      </div>
    </div>
  )
}
