'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/base/page-header'
import { PurchaseOrdersTable } from './purchase-orders-table'

interface PurchaseOrdersClientProps {
  userId: string
  userRole: string
}

export function PurchaseOrdersClient({
  userId,
  userRole,
}: PurchaseOrdersClientProps) {
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleRefresh = () => {
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <PageHeader
        title="Purchase Orders"
        subtitle="View and manage purchase orders through the approval workflow"
        showBackButton={false}
      />

      {/* Purchase Orders Table */}
      <PurchaseOrdersTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
        onRefresh={handleRefresh}
      />
    </div>
  )
}
