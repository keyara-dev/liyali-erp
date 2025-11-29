'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
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
import { MOCK_USERS } from '@/lib/mock-data'
import { UserRole } from '@/types/workflow'
import { Search, Edit2, Shield } from 'lucide-react'

interface UserManagementClientProps {
  userId: string
  userRole: string
}

interface UserWithStatus {
  id: string
  name: string
  email: string
  role: UserRole
  department?: string
  status: 'active' | 'inactive'
  lastLogin: string
  approvalCount: number
}

const ROLE_COLORS: Record<UserRole, string> = {
  REQUESTER: 'outline',
  DEPARTMENT_MANAGER: 'secondary',
  FINANCE_OFFICER: 'default',
  DIRECTOR: 'default',
  CFO: 'default',
  COMPLIANCE_OFFICER: 'secondary',
  ADMIN: 'destructive',
}

export function UserManagementClient({
  userId,
  userRole,
}: UserManagementClientProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedRole, setSelectedRole] = useState<UserRole | 'ALL'>('ALL')
  const [selectedUser, setSelectedUser] = useState<UserWithStatus | null>(null)
  const [showRoleModal, setShowRoleModal] = useState(false)

  // Convert mock users to extended format with status
  const allUsers: UserWithStatus[] = []
  Object.entries(MOCK_USERS).forEach(([, users]) => {
    users.forEach((user, index) => {
      allUsers.push({
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        department: user.department,
        status: index === 0 ? 'active' : 'active',
        lastLogin: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toLocaleDateString(),
        approvalCount: Math.floor(Math.random() * 30) + 1,
      })
    })
  })

  // Filter users
  let filteredUsers = allUsers
  if (searchTerm) {
    filteredUsers = filteredUsers.filter(
      (user) =>
        user.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        user.email.toLowerCase().includes(searchTerm.toLowerCase())
    )
  }
  if (selectedRole !== 'ALL') {
    filteredUsers = filteredUsers.filter((user) => user.role === selectedRole)
  }

  const handleChangeRole = (user: UserWithStatus, newRole: UserRole) => {
    // In a real app, would call server action here
    console.log(`Changing ${user.name}'s role to ${newRole}`)
    setShowRoleModal(false)
    setSelectedUser(null)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">User Management</h1>
        <p className="text-sm text-muted-foreground">
          Manage user roles and access permissions across the system
        </p>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid gap-4 md:grid-cols-2">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search by name or email..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Role Filter */}
            <Select value={selectedRole} onValueChange={(value) => setSelectedRole(value as UserRole | 'ALL')}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">All Roles</SelectItem>
                <SelectItem value="REQUESTER">Requester</SelectItem>
                <SelectItem value="DEPARTMENT_MANAGER">Department Manager</SelectItem>
                <SelectItem value="FINANCE_OFFICER">Finance Officer</SelectItem>
                <SelectItem value="DIRECTOR">Director</SelectItem>
                <SelectItem value="CFO">CFO</SelectItem>
                <SelectItem value="COMPLIANCE_OFFICER">Compliance Officer</SelectItem>
                <SelectItem value="ADMIN">Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Users Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">
            Users ({filteredUsers.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Department</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead className="text-right">Approvals</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Login</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredUsers.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-4 text-muted-foreground">
                      No users found
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredUsers.map((user) => (
                    <TableRow key={user.id}>
                      <TableCell className="font-medium">{user.name}</TableCell>
                      <TableCell className="text-sm">{user.email}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {user.department || '—'}
                      </TableCell>
                      <TableCell>
                        <Badge variant={ROLE_COLORS[user.role] as any}>
                          {user.role.replace(/_/g, ' ')}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {user.approvalCount}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="bg-secondary/10 text-secondary border-secondary/30">
                          {user.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {user.lastLogin}
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setSelectedUser(user)
                            setShowRoleModal(true)
                          }}
                          className="gap-1"
                        >
                          <Shield className="h-4 w-4" />
                          Edit Role
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Role Modal */}
      {showRoleModal && selectedUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>Change User Role</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">User</p>
                <p className="text-base font-semibold">{selectedUser.name}</p>
              </div>

              <div>
                <p className="text-sm font-medium mb-2">New Role</p>
                <Select defaultValue={selectedUser.role}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="REQUESTER">Requester</SelectItem>
                    <SelectItem value="DEPARTMENT_MANAGER">Department Manager</SelectItem>
                    <SelectItem value="FINANCE_OFFICER">Finance Officer</SelectItem>
                    <SelectItem value="DIRECTOR">Director</SelectItem>
                    <SelectItem value="CFO">CFO</SelectItem>
                    <SelectItem value="COMPLIANCE_OFFICER">Compliance Officer</SelectItem>
                    <SelectItem value="ADMIN">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex gap-2 justify-end pt-4">
                <Button
                  variant="outline"
                  onClick={() => {
                    setShowRoleModal(false)
                    setSelectedUser(null)
                  }}
                >
                  Cancel
                </Button>
                <Button
                  onClick={() => handleChangeRole(selectedUser, selectedUser.role)}
                >
                  Save Changes
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
