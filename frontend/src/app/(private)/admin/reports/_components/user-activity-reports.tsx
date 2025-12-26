'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { getDashboardMetrics } from '@/app/_actions/dashboard'
import { DashboardMetrics } from '@/types'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { User, Users, CheckCircle2 } from 'lucide-react'
import { QUERY_KEYS } from '@/lib/constants'

interface UserStat {
  id: string
  name: string
  role: string
  approvedCount: number
  activeDocuments: number
  lastActivity: string
}

export function UserActivityReports() {
  // Fetch dashboard metrics
  const { data: metrics, isLoading } = useQuery<DashboardMetrics>({
    queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
    queryFn: async () => {
      const result = await getDashboardMetrics()
      if (result.success && result.data) {
        return result.data
      }
      throw new Error('Failed to fetch metrics')
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  // Fetch user activity statistics from backend
  const { data: userStats = [] } = useQuery<UserStat[]>({
    queryKey: ['user-activity-stats'],
    queryFn: async () => {
      const response = await fetch('/api/users/activity-stats')
      if (!response.ok) throw new Error('Failed to fetch user stats')
      return response.json()
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  if (isLoading || !metrics) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Loading user activity reports...
      </div>
    )
  }

  const topContributors = userStats.slice(0, 3)
  const totalUsers = userStats.length

  return (
    <div className="space-y-6">
      {/* Activity Overview Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Users
            </CardTitle>
            <Users className="h-5 w-5 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{totalUsers}</div>
            <p className="text-xs text-muted-foreground mt-1">
              {Math.round((totalUsers / (totalUsers || 1)) * 100)}% active
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Docs in Progress
            </CardTitle>
            <User className="h-5 w-5 text-secondary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {userStats.reduce((sum, user) => sum + user.activeDocuments, 0)}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Across all users
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Actions
            </CardTitle>
            <CheckCircle2 className="h-5 w-5 text-accent" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {userStats.reduce((sum, user) => sum + user.approvedCount, 0)}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Approvals and rejections
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Top Contributors */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Top Contributors</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {topContributors.map((user, index) => (
              <div key={index} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-3 flex-1">
                  <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                    <span className="text-sm font-semibold text-primary">
                      {user.name.charAt(0)}
                    </span>
                  </div>
                  <div className="flex-1">
                    <p className="font-medium">{user.name}</p>
                    <p className="text-xs text-muted-foreground">{user.role.replace(/_/g, ' ')}</p>
                  </div>
                </div>
                <div className="text-right">
                  <Badge variant="secondary">{user.approvedCount} approvals</Badge>
                  <p className="text-xs text-muted-foreground mt-1">{user.activeDocuments} active</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* All Users Activity Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">User Activity Log</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead className="text-right">Approvals</TableHead>
                  <TableHead className="text-right">Active Docs</TableHead>
                  <TableHead>Last Activity</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {userStats.map((user, index) => (
                  <TableRow key={index}>
                    <TableCell className="font-medium">{user.name}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{user.role.replace(/_/g, ' ')}</Badge>
                    </TableCell>
                    <TableCell className="text-right">{user.approvedCount}</TableCell>
                    <TableCell className="text-right">{user.activeDocuments}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {user.lastActivity}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
