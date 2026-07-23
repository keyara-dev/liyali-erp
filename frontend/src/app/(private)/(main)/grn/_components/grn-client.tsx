'use client'

import { useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { PageHeader } from '@/components/base/page-header'
import { GrnTable } from './grn-table'
import { CreateGRNDialog } from './create-grn-dialog'
import { ReadyForGrnSection } from './ready-for-grn-section'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Plus } from 'lucide-react'

interface GrnClientProps {
  userId: string
  userRole: string
}

export function GrnClient({ userId, userRole }: GrnClientProps) {
  const [refreshTrigger, setRefreshTrigger] = useState(0)
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  // ?po=PO-NUMBER deep-links from PO detail page to scope the list.
  const searchParams = useSearchParams()
  const poFilter = searchParams.get('po') ?? undefined

  const handleRefresh = () => {
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <PageHeader
          title="Goods Received Notes"
          subtitle={
            poFilter
              ? `GRNs linked to ${poFilter}`
              : 'View and manage goods received notes from purchase orders'
          }
          showBackButton={Boolean(poFilter)}
        />
        <Button onClick={() => setIsCreateDialogOpen(true)} className="shrink-0 mt-1">
          <Plus className="h-4 w-4" />
          Create GRN
        </Button>
      </div>

      {/* Ready for GRN — approved POs (goods-first) + approved/paid PVs
          (payment-first) awaiting goods receipt. Hidden when deep-linked
          from a PO so the filtered view stays focused. */}
      {!poFilter && (
        <>
          <div className="space-y-3">
            <div>
              <h2 className="text-lg font-semibold">Ready for GRN</h2>
              <p className="text-sm text-muted-foreground">
                Select an approved document to record goods received
              </p>
            </div>
            <ReadyForGrnSection
              userId={userId}
              userRole={userRole}
              onChanged={handleRefresh}
            />
          </div>

          <Separator className="my-8" />

          <div>
            <h2 className="text-lg font-semibold">All Goods Received Notes</h2>
            <p className="text-sm text-muted-foreground">
              View and manage all goods received notes
            </p>
          </div>
        </>
      )}

      <GrnTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
        onRefresh={handleRefresh}
        poFilter={poFilter}
      />

      <CreateGRNDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onSuccess={handleRefresh}
      />
    </div>
  )
}
