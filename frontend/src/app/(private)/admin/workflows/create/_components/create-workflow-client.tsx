'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { PageHeader } from '@/components/base/page-header'
import { WorkflowBuilder } from '../../_components/workflow-builder'
import { toast } from 'sonner'
import { saveWorkflow } from '@/lib/workflow-storage'

interface CreateWorkflowClientProps {
  userId: string
  userRole: string
}

export interface WorkflowStage {
  id: string
  order: number
  name: string
  description: string
  approverRole: string
  requiredApprovals: number
  canReject: boolean
  canReassign: boolean
}

export interface WorkflowFormData {
  name: string
  description: string
  documentType: string
  stages: WorkflowStage[]
  isDefault: boolean
}

export function CreateWorkflowClient({
  userId,
  userRole,
}: CreateWorkflowClientProps) {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleBack = () => {
    router.back()
  }

  const handleSubmit = async (formData: WorkflowFormData) => {
    setIsSubmitting(true)
    try {
      // Generate workflow ID
      const workflowId = `wf-${Date.now()}`

      // Prepare workflow for storage
      const workflow = {
        id: workflowId,
        name: formData.name,
        description: formData.description,
        documentType: formData.documentType,
        stages: formData.stages.length,
        status: 'ACTIVE' as const,
        updatedAt: new Date().toISOString(),
        fullData: formData,
      }

      // Save to localStorage
      const result = saveWorkflow(workflow)

      if (result) {
        toast.success('Workflow created successfully')
        router.push('/admin/workflows')
      } else {
        toast.error('Failed to save workflow')
      }
    } catch (error) {
      console.error('Failed to create workflow:', error)
      toast.error('Failed to create workflow')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Workflow"
        subtitle="Design a new custom approval workflow"
        onBackClick={handleBack}
        showBackButton={true}
      />

      <WorkflowBuilder
        onSubmit={handleSubmit}
        isSubmitting={isSubmitting}
        mode="create"
      />
    </div>
  )
}
