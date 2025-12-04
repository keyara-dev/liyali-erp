'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/base/page-header'
import { Plus } from 'lucide-react'
import { PaymentVouchersTable } from './payment-vouchers-table'

interface PaymentVouchersClientProps {
  userId: string
  userRole: string
}

export function PaymentVouchersClient({
  userId,
  userRole,
}: PaymentVouchersClientProps) {
  const router = useRouter()
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
      {/* Page Header with Create Button */}
      <div className="flex items-center justify-between">
        <PageHeader
          title="Payment Vouchers"
          subtitle="Manage payment vouchers for approved goods and services"
          showBackButton={false}
        />
        <Button
          onClick={() => router.push('/payment-vouchers/create')}
          className="bg-blue-600 hover:bg-blue-700 gap-2"
        >
          <Plus className="h-4 w-4" />
          Create Payment Voucher
        </Button>
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
