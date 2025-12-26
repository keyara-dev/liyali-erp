'use client'

import { useRouter } from 'next/navigation'
import { PageHeader } from '@/components/base/page-header'
import { WorkflowBuilder } from '../../_components/workflow-builder'
import { useCreateWorkflow, type WorkflowFormData } from '@/hooks/use-workflow-queries'

interface CreateWorkflowClientProps {
  userId: string
  userRole: string
}

export function CreateWorkflowClient({
  userId,
  userRole,
}: CreateWorkflowClientProps) {
  const router = useRouter()

  // Create workflow mutation
  const createMutation = useCreateWorkflow(() => {
    router.push('/admin/workflows')
  })

  const handleBack = () => {
    router.back()
  }

  const handleSubmit = async (formData: WorkflowFormData) => {
    createMutation.mutate(formData)
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
        isSubmitting={createMutation.isPending}
        mode="create"
      />
    </div>
  )
}
