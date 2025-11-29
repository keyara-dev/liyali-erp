'use client'

import { useState } from 'react'
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
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Goods Received Notes</h1>
        <p className="text-gray-600 mt-2">
          View and manage goods received notes from purchase orders
        </p>
      </div>

      <GrnTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
        onRefresh={handleRefresh}
      />
    </div>
  )
}
