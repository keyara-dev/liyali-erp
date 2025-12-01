'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Textarea } from '@/components/ui/textarea'
import { ArrowLeft, AlertTriangle, CheckCircle2, Package, Signature } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { GRNItemsMatchingTable } from '../_components/grn-items-matching-table'

interface GRNConfirmationClientProps {
  grnId: string
  userId: string
  userRole: string
}

interface ReceivedItem {
  id: string
  itemNumber: number
  description: string
  poQuantity: number
  receivedQuantity: number
  unit: string
  variance: number
  damage: number
  damageNotes?: string
  condition: 'GOOD' | 'DAMAGED' | 'PARTIAL'
}

interface GoodsReceivedNote {
  id: string
  grnNumber: string
  poNumber: string
  status: 'DRAFT' | 'SUBMITTED' | 'CONFIRMED' | 'REJECTED'
  warehouseLocation: string
  receivedDate: string
  receivedBy: string
  approvedBy?: string
  items: ReceivedItem[]
  qualityIssues: Array<{
    id: string
    itemId: string
    description: string
    severity: 'LOW' | 'MEDIUM' | 'HIGH'
  }>
  notes?: string
  currentStage: number
  stageName: string
  createdAt: string
  updatedAt: string
}

const STAGE_NAMES: Record<number, string> = {
  1: 'Warehouse Clerk Receipt',
  2: 'Department Manager Confirmation',
}

// Mock data generator
function generateMockGRN(grnId: string): GoodsReceivedNote {
  const currentStage = Math.floor(Math.random() * 2) + 1

  return {
    id: grnId,
    grnNumber: `GRN-2024-${String(Math.floor(Math.random() * 9000) + 1000).padStart(4, '0')}`,
    poNumber: `PO-2024-${String(Math.floor(Math.random() * 9000) + 1000).padStart(4, '0')}`,
    status: 'SUBMITTED',
    warehouseLocation: 'Warehouse A - Section 3',
    receivedDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    receivedBy: 'WAREHOUSE-USER-001',
    approvedBy: undefined,
    items: [
      {
        id: 'item-1',
        itemNumber: 1,
        description: 'Office Chairs - Ergonomic',
        poQuantity: 10,
        receivedQuantity: 10,
        unit: 'units',
        variance: 0,
        damage: 0,
        condition: 'GOOD',
      },
      {
        id: 'item-2',
        itemNumber: 2,
        description: 'Standing Desks - Electric',
        poQuantity: 5,
        receivedQuantity: 4,
        unit: 'units',
        variance: -1,
        damage: 1,
        damageNotes: 'One unit arrived with damaged motor',
        condition: 'DAMAGED',
      },
      {
        id: 'item-3',
        itemNumber: 3,
        description: 'Computer Monitors - 27 inch',
        poQuantity: 8,
        receivedQuantity: 8,
        unit: 'units',
        variance: 0,
        damage: 0,
        condition: 'GOOD',
      },
    ],
    qualityIssues: [
      {
        id: 'issue-1',
        itemId: 'item-2',
        description: 'Standing Desk motor malfunction',
        severity: 'HIGH',
      },
    ],
    notes: 'General inspection completed. One standing desk has motor issues.',
    currentStage,
    stageName: STAGE_NAMES[currentStage],
    createdAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
  }
}

export function GRNConfirmationClient({
  grnId,
  userId,
  userRole,
}: GRNConfirmationClientProps) {
  const router = useRouter()
  const [grn, setGRN] = useState<GoodsReceivedNote | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [confirmCheckbox, setConfirmCheckbox] = useState(false)
  const [confirmationNotes, setConfirmationNotes] = useState('')
  const [signature, setSignature] = useState('')

  useEffect(() => {
    // Simulate data loading
    const timer = setTimeout(() => {
      setGRN(generateMockGRN(grnId))
      setIsLoading(false)
    }, 500)

    return () => clearTimeout(timer)
  }, [grnId])

  const handleConfirm = async () => {
    if (!confirmCheckbox) {
      toast.error('Please confirm all items have been checked')
      return
    }

    if (!signature) {
      toast.error('Please provide your signature')
      return
    }

    setIsSubmitting(true)
    try {
      // Simulate confirmation process
      await new Promise((resolve) => setTimeout(resolve, 1000))
      toast.success('GRN confirmed successfully')
      router.push('/workflows/grn')
    } catch (error) {
      toast.error('Failed to confirm GRN')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleReject = async () => {
    if (!confirmationNotes) {
      toast.error('Please provide rejection reason')
      return
    }

    if (!signature) {
      toast.error('Please provide your signature')
      return
    }

    setIsSubmitting(true)
    try {
      // Simulate rejection process
      await new Promise((resolve) => setTimeout(resolve, 1000))
      toast.success('GRN rejected successfully')
      router.push('/workflows/grn')
    } catch (error) {
      toast.error('Failed to reject GRN')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleBack = () => {
    router.back()
  }

  if (isLoading || !grn) {
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

  const hasQualityIssues = grn.qualityIssues.length > 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">{grn.grnNumber}</h1>
          <p className="text-muted-foreground">Goods Received Confirmation</p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        {/* Main Content */}
        <div className="md:col-span-2 space-y-6">
          {/* Quality Alert */}
          {hasQualityIssues && (
            <Card className="border-yellow-200 bg-yellow-50">
              <CardContent className="pt-4 flex gap-3">
                <AlertTriangle className="h-5 w-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-semibold text-yellow-900">Quality Issues Detected</p>
                  <p className="text-sm text-yellow-800 mt-1">
                    {grn.qualityIssues.length} quality issue(s) have been reported. Please review
                    before confirming.
                  </p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* GRN Summary */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="h-5 w-5" />
                Goods Received Summary
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 md:grid-cols-3">
              <div>
                <p className="text-sm text-muted-foreground">PO Number</p>
                <p className="font-semibold">{grn.poNumber}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Warehouse Location</p>
                <p className="font-semibold">{grn.warehouseLocation}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Received Date</p>
                <p className="font-semibold">
                  {new Date(grn.receivedDate).toLocaleDateString()}
                </p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Items</p>
                <p className="font-semibold">{grn.items.length}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Good Condition</p>
                <p className="font-semibold text-green-600">
                  {grn.items.filter((i) => i.condition === 'GOOD').length}
                </p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Damaged</p>
                <p className="font-semibold text-red-600">
                  {grn.items.filter((i) => i.condition === 'DAMAGED').length}
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Items Matching */}
          <Card>
            <CardHeader>
              <CardTitle>Items Received vs. Ordered</CardTitle>
            </CardHeader>
            <CardContent>
              <GRNItemsMatchingTable items={grn.items} />
            </CardContent>
          </Card>

          {/* Original Notes */}
          {grn.notes && (
            <Card>
              <CardHeader>
                <CardTitle>Warehouse Inspection Notes</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm whitespace-pre-wrap bg-muted/30 p-3 rounded">
                  {grn.notes}
                </p>
              </CardContent>
            </Card>
          )}

          {/* Confirmation Form */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Confirmation Required</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Confirmation Checkbox */}
              <div className="flex gap-3 items-start p-4 border rounded-lg bg-muted/30">
                <Checkbox
                  id="confirm"
                  checked={confirmCheckbox}
                  onCheckedChange={(checked) => setConfirmCheckbox(checked as boolean)}
                  className="mt-1"
                />
                <label htmlFor="confirm" className="cursor-pointer flex-1">
                  <p className="font-medium">I have checked all items and confirm</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    I certify that all items listed above have been physically inspected and match
                    the quantities and conditions recorded
                  </p>
                </label>
              </div>

              {/* Confirmation Notes */}
              <div>
                <label className="text-sm font-medium">
                  Your Confirmation Notes (or Rejection Reason)
                </label>
                <Textarea
                  placeholder="Enter any additional notes, observations, or reasons for rejection..."
                  value={confirmationNotes}
                  onChange={(e) => setConfirmationNotes(e.target.value)}
                  className="mt-2"
                  rows={4}
                />
              </div>

              {/* Signature */}
              <div>
                <label className="text-sm font-medium flex items-center gap-2">
                  <Signature className="h-4 w-4" />
                  Your Signature (Name)
                </label>
                <input
                  type="text"
                  placeholder="Enter your full name"
                  value={signature}
                  onChange={(e) => setSignature(e.target.value)}
                  className="w-full mt-2 px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Your signature confirms this GRN confirmation
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Action Panel */}
        <div>
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Confirmation Status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Requirements Checklist */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  {confirmCheckbox ? (
                    <CheckCircle2 className="h-5 w-5 text-green-600" />
                  ) : (
                    <div className="h-5 w-5 border-2 border-gray-300 rounded-full" />
                  )}
                  <span
                    className={
                      confirmCheckbox
                        ? 'text-sm font-medium text-green-600'
                        : 'text-sm text-muted-foreground'
                    }
                  >
                    Items Checked
                  </span>
                </div>

                <div className="flex items-center gap-2">
                  {signature ? (
                    <CheckCircle2 className="h-5 w-5 text-green-600" />
                  ) : (
                    <div className="h-5 w-5 border-2 border-gray-300 rounded-full" />
                  )}
                  <span
                    className={
                      signature
                        ? 'text-sm font-medium text-green-600'
                        : 'text-sm text-muted-foreground'
                    }
                  >
                    Signature Provided
                  </span>
                </div>
              </div>

              <div className="border-t pt-4 space-y-2">
                <Button
                  onClick={handleConfirm}
                  disabled={isSubmitting}
                  className="w-full bg-green-600 hover:bg-green-700"
                >
                  {isSubmitting ? 'Confirming...' : 'Confirm Receipt'}
                </Button>

                <Button
                  onClick={handleReject}
                  disabled={isSubmitting}
                  variant="destructive"
                  className="w-full"
                >
                  {isSubmitting ? 'Rejecting...' : 'Reject GRN'}
                </Button>

                <Button onClick={handleBack} variant="outline" className="w-full">
                  Cancel
                </Button>
              </div>

              {hasQualityIssues && (
                <div className="pt-4 border-t text-xs text-muted-foreground">
                  <p className="font-medium mb-2">⚠️ Quality Issues Present</p>
                  <p>
                    This GRN has {grn.qualityIssues.length} reported quality issue(s). Please address
                    before confirming if they affect acceptance.
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
