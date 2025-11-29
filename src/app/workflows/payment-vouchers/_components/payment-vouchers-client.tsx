'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { PaymentVouchersTable } from './payment-vouchers-table'

interface PaymentVouchersClientProps {
  userId: string
  userRole: string
}

export function PaymentVouchersClient({
  userId,
  userRole,
}: PaymentVouchersClientProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [paymentVouchers, setPaymentVouchers] = useState([])
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  useEffect(() => {
    setIsLoading(true)
    // Load payment vouchers from mock data or API
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
            Payment Vouchers
          </h1>
          <p className="text-sm text-muted-foreground">
            Manage payment vouchers for approved goods and services
          </p>
        </div>
      </div>

      {/* Payment Vouchers Table */}
      <PaymentVouchersTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
        onRefresh={handleRefresh}
      />
    </div>
  )
}
