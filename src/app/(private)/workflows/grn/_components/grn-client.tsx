'use client'

import { useState } from 'react'
import { PageHeader } from '@/components/base/page-header'
import { GrnTable } from './grn-table'

interface GrnClientProps {
  userId: string
  userRole: string
}

export function GrnClient({ userId, userRole }: GrnClientProps) {
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleRefresh = () => {
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Goods Received Notes"
        subtitle="View and manage goods received notes from purchase orders"
        showBackButton={false}
      />

      <GrnTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
        onRefresh={handleRefresh}
      />
    </div>
  )
}
