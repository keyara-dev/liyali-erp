'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { PageHeader } from '@/components/base/page-header'
import { WorkflowBuilder } from '../../../_components/workflow-builder'
import { toast } from 'sonner'
import { WorkflowFormData } from '../../../create/_components/create-workflow-client'

interface EditWorkflowClientProps {
  workflowId: string
  userId: string
  userRole: string
}

// Mock workflow data - in real app, fetch from server
const mockWorkflows: Record<string, WorkflowFormData> = {
  'wf-1': {
    name: 'Standard Requisition Approval',
    description: '4-stage approval process for purchase requisitions',
    documentType: 'REQUISITION',
    isDefault: true,
    stages: [
      {
        id: 'stage-1',
        order: 1,
        name: 'Department Manager Review',
        description: 'Initial review by department manager',
        approverRole: 'DEPARTMENT_MANAGER',
        requiredApprovals: 1,
        canReject: true,
        canReassign: true,
      },
      {
        id: 'stage-2',
        order: 2,
        name: 'Finance Officer Review',
        description: 'Budget and finance validation',
        approverRole: 'FINANCE_OFFICER',
        requiredApprovals: 1,
        canReject: true,
        canReassign: true,
      },
      {
        id: 'stage-3',
        order: 3,
        name: 'CFO Approval',
        description: 'Final approval by CFO',
        approverRole: 'CFO',
        requiredApprovals: 1,
        canReject: true,
        canReassign: false,
      },
    ],
  },
}

export function EditWorkflowClient({
  workflowId,
  userId,
  userRole,
}: EditWorkflowClientProps) {
  const router = useRouter()
  const [initialData, setInitialData] = useState<WorkflowFormData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    // Simulate loading workflow data
    const timer = setTimeout(() => {
      const data = mockWorkflows[workflowId]
      if (data) {
        setInitialData(data)
      } else {
        toast.error('Workflow not found')
        router.push('/admin/workflows')
      }
      setIsLoading(false)
    }, 500)

    return () => clearTimeout(timer)
  }, [workflowId, router])

  const handleBack = () => {
    router.back()
  }

  const handleSubmit = async (formData: WorkflowFormData) => {
    setIsSubmitting(true)
    try {
      // TODO: Call updateWorkflow server action
      console.log('Updating workflow:', workflowId, formData)

      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000))

      toast.success('Workflow updated successfully')
      router.push('/admin/workflows')
    } catch (error) {
      console.error('Failed to update workflow:', error)
      toast.error('Failed to update workflow')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader
          title="Edit Workflow"
          subtitle="Loading workflow details..."
          onBackClick={handleBack}
          showBackButton={true}
        />
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    )
  }

  if (!initialData) {
    return (
      <div className="space-y-6">
        <PageHeader
          title="Edit Workflow"
          subtitle="Workflow not found"
          onBackClick={handleBack}
          showBackButton={true}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Edit: ${initialData.name}`}
        subtitle="Update the workflow configuration"
        onBackClick={handleBack}
        showBackButton={true}
      />

      <WorkflowBuilder
        onSubmit={handleSubmit}
        isSubmitting={isSubmitting}
        mode="edit"
        initialData={initialData}
      />
    </div>
  )
}
