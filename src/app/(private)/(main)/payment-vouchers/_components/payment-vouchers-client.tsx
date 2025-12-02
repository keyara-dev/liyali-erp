'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/base/page-header'
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
    <div className="space-y-6">
      {/* Page Header */}
      <PageHeader
        title="Payment Vouchers"
        subtitle="Manage payment vouchers for approved goods and services"
        showBackButton={false}
      />

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
