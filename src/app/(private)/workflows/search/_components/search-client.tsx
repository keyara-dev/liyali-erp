'use client'

import { useState } from 'react'
import { SearchForm } from './search-form'
import { TransactionResults } from './transaction-results'
import { SearchFilters } from '@/types/workflow'

interface SearchClientProps {
  userId: string
  userRole: string
}

export function SearchClient({ userId, userRole }: SearchClientProps) {
  const [filters, setFilters] = useState<SearchFilters>({
    documentNumber: '',
    documentType: 'ALL',
    status: 'ALL',
    startDate: '',
    endDate: '',
  })
  const [refreshTrigger, setRefreshTrigger] = useState(0)
  const [isSearching, setIsSearching] = useState(false)

  const handleSearch = (newFilters: SearchFilters) => {
    setFilters(newFilters)
    setIsSearching(true)
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">
          Search Transactions
        </h1>
        <p className="text-sm text-muted-foreground">
          Find requisitions, purchase orders, and GRNs by searching filters
        </p>
      </div>

      {/* Search Form */}
      <SearchForm onSearch={handleSearch} isSearching={isSearching} />

      {/* Results */}
      <TransactionResults
        filters={filters}
        refreshTrigger={refreshTrigger}
        userRole={userRole}
      />
    </div>
  )
}
