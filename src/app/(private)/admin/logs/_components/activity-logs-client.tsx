'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Search, Download } from 'lucide-react'

interface ActivityLog {
  id: string
  timestamp: string
  user: string
  action: string
  resource: string
  resourceId: string
  status: 'success' | 'failed' | 'pending'
  details: string
  ipAddress: string
}

interface ActivityLogsClientProps {
  userId: string
  userRole: string
}

const ACTION_COLORS: Record<string, string> = {
  created: 'default',
  approved: 'secondary',
  rejected: 'destructive',
  submitted: 'default',
  edited: 'outline',
  viewed: 'outline',
  deleted: 'destructive',
}

const STATUS_COLORS: Record<string, string> = {
  success: 'default',
  failed: 'destructive',
  pending: 'outline',
}

export function ActivityLogsClient({ userId, userRole }: ActivityLogsClientProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedAction, setSelectedAction] = useState<string>('ALL')
  const [selectedStatus, setSelectedStatus] = useState<string>('ALL')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  // Generate mock activity logs
  const mockLogs: ActivityLog[] = [
    {
      id: '1',
      timestamp: new Date(Date.now() - 1 * 60 * 1000).toLocaleString(),
      user: 'John Mwale',
      action: 'submitted',
      resource: 'REQUISITION',
      resourceId: 'REQ-2024-001',
      status: 'success',
      details: 'Submitted purchase requisition for office equipment',
      ipAddress: '192.168.1.100',
    },
    {
      id: '2',
      timestamp: new Date(Date.now() - 5 * 60 * 1000).toLocaleString(),
      user: 'James Chileshe',
      action: 'approved',
      resource: 'PURCHASE_ORDER',
      resourceId: 'PO-2024-042',
      status: 'success',
      details: 'Approved purchase order for supplies',
      ipAddress: '192.168.1.105',
    },
    {
      id: '3',
      timestamp: new Date(Date.now() - 15 * 60 * 1000).toLocaleString(),
      user: 'Sarah Banda',
      action: 'edited',
      resource: 'REQUISITION',
      resourceId: 'REQ-2024-003',
      status: 'success',
      details: 'Edited draft requisition',
      ipAddress: '192.168.1.110',
    },
    {
      id: '4',
      timestamp: new Date(Date.now() - 30 * 60 * 1000).toLocaleString(),
      user: 'Paul Nkosi',
      action: 'rejected',
      resource: 'PAYMENT_VOUCHER',
      resourceId: 'PV-2024-015',
      status: 'success',
      details: 'Rejected payment voucher - budget exceeded',
      ipAddress: '192.168.1.112',
    },
    {
      id: '5',
      timestamp: new Date(Date.now() - 1 * 60 * 60 * 1000).toLocaleString(),
      user: 'Maria Chiyanda',
      action: 'viewed',
      resource: 'REQUISITION',
      resourceId: 'REQ-2024-002',
      status: 'success',
      details: 'Viewed requisition details',
      ipAddress: '192.168.1.115',
    },
    {
      id: '6',
      timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000).toLocaleString(),
      user: 'Grace Mvula',
      action: 'created',
      resource: 'PAYMENT_VOUCHER',
      resourceId: 'PV-2024-016',
      status: 'failed',
      details: 'Failed to create payment voucher - missing required fields',
      ipAddress: '192.168.1.118',
    },
    {
      id: '7',
      timestamp: new Date(Date.now() - 3 * 60 * 60 * 1000).toLocaleString(),
      user: 'David Moyo',
      action: 'approved',
      resource: 'REQUISITION',
      resourceId: 'REQ-2024-001',
      status: 'success',
      details: 'Final approval for requisition',
      ipAddress: '192.168.1.120',
    },
    {
      id: '8',
      timestamp: new Date(Date.now() - 4 * 60 * 60 * 1000).toLocaleString(),
      user: 'Catherine Phiri',
      action: 'viewed',
      resource: 'PURCHASE_ORDER',
      resourceId: 'PO-2024-041',
      status: 'success',
      details: 'Viewed purchase order for audit',
      ipAddress: '192.168.1.122',
    },
  ]

  // Filter logs
  let filteredLogs = mockLogs
  if (searchTerm) {
    filteredLogs = filteredLogs.filter(
      (log) =>
        log.user.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.resourceId.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.details.toLowerCase().includes(searchTerm.toLowerCase())
    )
  }
  if (selectedAction !== 'ALL') {
    filteredLogs = filteredLogs.filter((log) => log.action === selectedAction)
  }
  if (selectedStatus !== 'ALL') {
    filteredLogs = filteredLogs.filter((log) => log.status === selectedStatus)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">Activity Logs</h1>
        <p className="text-sm text-muted-foreground">
          Audit trail of all system activities and user actions
        </p>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {/* Search */}
            <div className="relative lg:col-span-2">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search by user, resource, or action..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Action Filter */}
            <Select value={selectedAction} onValueChange={setSelectedAction}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">All Actions</SelectItem>
                <SelectItem value="created">Created</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
                <SelectItem value="rejected">Rejected</SelectItem>
                <SelectItem value="submitted">Submitted</SelectItem>
                <SelectItem value="edited">Edited</SelectItem>
                <SelectItem value="viewed">Viewed</SelectItem>
                <SelectItem value="deleted">Deleted</SelectItem>
              </SelectContent>
            </Select>

            {/* Status Filter */}
            <Select value={selectedStatus} onValueChange={setSelectedStatus}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">All Statuses</SelectItem>
                <SelectItem value="success">Success</SelectItem>
                <SelectItem value="failed">Failed</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Date Range */}
          <div className="grid gap-4 md:grid-cols-2 mt-4">
            <div>
              <label className="text-sm font-medium text-muted-foreground">Start Date</label>
              <Input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>
            <div>
              <label className="text-sm font-medium text-muted-foreground">End Date</label>
              <Input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
              />
            </div>
          </div>

          {/* Export Button */}
          <div className="mt-4">
            <Button variant="outline" className="gap-2">
              <Download className="h-4 w-4" />
              Export Logs
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Activity Logs Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Activity Log ({filteredLogs.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>User</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Resource</TableHead>
                  <TableHead>Details</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>IP Address</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredLogs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-4 text-muted-foreground">
                      No activity logs found
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredLogs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="text-sm whitespace-nowrap">
                        {log.timestamp}
                      </TableCell>
                      <TableCell className="font-medium">{log.user}</TableCell>
                      <TableCell>
                        <Badge variant={ACTION_COLORS[log.action] as any}>
                          {log.action}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono text-sm">{log.resourceId}</TableCell>
                      <TableCell className="text-sm text-muted-foreground max-w-xs">
                        {log.details}
                      </TableCell>
                      <TableCell>
                        <Badge variant={STATUS_COLORS[log.status] as any}>
                          {log.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {log.ipAddress}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
