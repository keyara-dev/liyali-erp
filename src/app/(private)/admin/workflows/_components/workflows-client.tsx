'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PageHeader } from '@/components/base/page-header'
import { Plus, Edit2, Trash2, Copy } from 'lucide-react'
import Link from 'next/link'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { toast } from 'sonner'
import { StatusBadge } from '@/components/status-badge'

interface WorkflowsClientProps {
  userId: string
  userRole: string
}

interface Workflow {
  id: string
  name: string
  description: string
  documentType: string
  stages: number
  status: 'ACTIVE' | 'DEPRECATED'
  createdAt: string
  updatedAt: string
  createdBy: string
}

// Mock workflows data
const mockWorkflows: Workflow[] = [
  {
    id: 'wf-1',
    name: 'Standard Requisition Approval',
    description: '4-stage approval process for purchase requisitions',
    documentType: 'REQUISITION',
    stages: 4,
    status: 'ACTIVE',
    createdAt: '2024-01-15T10:30:00Z',
    updatedAt: '2024-11-20T14:22:00Z',
    createdBy: 'admin@example.com',
  },
  {
    id: 'wf-2',
    name: 'Purchase Order Approval',
    description: '4-stage approval with CFO override capability',
    documentType: 'PURCHASE_ORDER',
    stages: 4,
    status: 'ACTIVE',
    createdAt: '2024-02-10T08:15:00Z',
    updatedAt: '2024-11-18T11:45:00Z',
    createdBy: 'admin@example.com',
  },
  {
    id: 'wf-3',
    name: 'Payment Voucher Review',
    description: 'Finance review workflow for payment processing',
    documentType: 'PAYMENT_VOUCHER',
    stages: 3,
    status: 'ACTIVE',
    createdAt: '2024-03-05T09:20:00Z',
    updatedAt: '2024-11-19T13:10:00Z',
    createdBy: 'finance-admin@example.com',
  },
  {
    id: 'wf-4',
    name: 'GRN Confirmation Flow',
    description: 'Simple goods receipt confirmation workflow',
    documentType: 'GOODS_RECEIVED_NOTE',
    stages: 2,
    status: 'ACTIVE',
    createdAt: '2024-04-12T11:00:00Z',
    updatedAt: '2024-11-17T10:30:00Z',
    createdBy: 'warehouse@example.com',
  },
]

export function WorkflowsClient({ userId, userRole }: WorkflowsClientProps) {
  const router = useRouter()
  const [workflows, setWorkflows] = useState<Workflow[]>(mockWorkflows)
  const [deleteId, setDeleteId] = useState<string | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleDelete = async () => {
    if (!deleteId) return

    setIsDeleting(true)
    try {
      // Simulate delete
      await new Promise((resolve) => setTimeout(resolve, 500))
      setWorkflows(workflows.filter((w) => w.id !== deleteId))
      toast.success('Workflow deleted successfully')
      setDeleteId(null)
    } catch (error) {
      toast.error('Failed to delete workflow')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleDuplicate = (workflow: Workflow) => {
    const newWorkflow = {
      ...workflow,
      id: `wf-${Date.now()}`,
      name: `${workflow.name} (Copy)`,
      status: 'ACTIVE' as const,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      createdBy: userId,
    }
    setWorkflows([...workflows, newWorkflow])
    toast.success(`${workflow.name} duplicated successfully`)
  }

  const getDocumentTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      'REQUISITION': 'Requisition',
      'PURCHASE_ORDER': 'Purchase Order',
      'PAYMENT_VOUCHER': 'Payment Voucher',
      'GOODS_RECEIVED_NOTE': 'GRN',
      'BUDGET': 'Budget',
    }
    return labels[type] || type
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <PageHeader
          title="Workflow Management"
          subtitle="Create and manage custom approval workflows"
          showBackButton={false}
        />
        <Link href="/admin/workflows/create">
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            Create Workflow
          </Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Approval Workflows</CardTitle>
        </CardHeader>
        <CardContent>
          {workflows.length === 0 ? (
            <div className="py-8 text-center">
              <p className="text-muted-foreground mb-4">
                No workflows created yet
              </p>
              <Link href="/admin/workflows/create">
                <Button variant="outline">Create Your First Workflow</Button>
              </Link>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Document Type</TableHead>
                    <TableHead className="text-center">Stages</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Last Updated</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {workflows.map((workflow) => (
                    <TableRow key={workflow.id}>
                      <TableCell>
                        <div>
                          <p className="font-medium">{workflow.name}</p>
                          <p className="text-sm text-muted-foreground">
                            {workflow.description}
                          </p>
                        </div>
                      </TableCell>
                      <TableCell>
                        {getDocumentTypeLabel(workflow.documentType)}
                      </TableCell>
                      <TableCell className="text-center">
                        {workflow.stages}
                      </TableCell>
                      <TableCell>
                        <StatusBadge
                          status={workflow.status}
                          type="document"
                        />
                      </TableCell>
                      <TableCell>
                        {new Date(workflow.updatedAt).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Link href={`/admin/workflows/${workflow.id}/edit`}>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="gap-2"
                            >
                              <Edit2 className="h-4 w-4" />
                            </Button>
                          </Link>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDuplicate(workflow)}
                            className="gap-2"
                          >
                            <Copy className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="gap-2 text-destructive hover:text-destructive"
                            onClick={() => setDeleteId(workflow.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <AlertDialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Workflow?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. The workflow will be permanently deleted
              and any active assignments will be affected.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="flex gap-2 justify-end">
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </div>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
