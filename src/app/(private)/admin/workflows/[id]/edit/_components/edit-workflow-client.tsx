'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { PageHeader } from '@/components/base/page-header'
import { WorkflowBuilder } from '../../../_components/workflow-builder'
import { toast } from 'sonner'
import { WorkflowFormData } from '../../../create/_components/create-workflow-client'
import { getWorkflowById, saveWorkflow } from '@/lib/workflow-storage'

interface EditWorkflowClientProps {
  workflowId: string
  userId: string
  userRole: string
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
    // Load workflow data from localStorage
    const timer = setTimeout(() => {
      console.log('Requested workflowId:', workflowId)
      const workflow = getWorkflowById(workflowId)
      if (workflow) {
        setInitialData(workflow.fullData)
      } else {
        console.log('Workflow not found for ID:', workflowId)
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
        toast.success('Workflow updated successfully')
        router.push('/admin/workflows')
      } else {
        toast.error('Failed to save workflow')
      }
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
