'use client'

import { useState, useEffect } from 'react'
import { PlusCircledIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { PurchaseOrdersTable } from './purchase-orders-table'

interface PurchaseOrdersClientProps {
  userId: string
  userRole: string
}

export function PurchaseOrdersClient({
  userId,
  userRole,
}: PurchaseOrdersClientProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [purchaseOrders, setPurchaseOrders] = useState([])
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  useEffect(() => {
    setIsLoading(true)
    // Load purchase orders from mock data or API
    // This will be implemented when we create the workflow actions
    setIsLoading(false)
  }, [refreshTrigger])

  const handleRefresh = () => {
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-4">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold tracking-tight lg:text-2xl">
            Purchase Orders
          </h1>
          <p className="text-sm text-muted-foreground">
            View and manage purchase orders through the approval workflow
          </p>
        </div>
      </div>

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
