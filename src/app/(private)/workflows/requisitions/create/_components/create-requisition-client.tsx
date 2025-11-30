'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { CreateRequisitionForm } from './create-form'
import { FormPreview } from './form-preview'
import { RequisitionItem } from '@/types/workflow'

interface CreateRequisitionClientProps {
  userId: string
  userRole: string
  userName: string
}

export interface RequisitionFormData {
  department: string
  requestedFor: string
  justification: string
  budgetCode: string
  items: RequisitionItem[]
}

export function CreateRequisitionClient({
  userId,
  userRole,
  userName,
}: CreateRequisitionClientProps) {
  const router = useRouter()
  const [currentStep, setCurrentStep] = useState<'form' | 'preview'>('form')
  const [formData, setFormData] = useState<RequisitionFormData>({
    department: '',
    requestedFor: userName,
    justification: '',
    budgetCode: '',
    items: [],
  })

  const handleFormSubmit = (data: RequisitionFormData) => {
    setFormData(data)
    setCurrentStep('preview')
  }

  const handleBack = () => {
    setCurrentStep('form')
  }

  const handleSubmit = async (data: RequisitionFormData) => {
    // Submit to server action
    try {
      // TODO: Call createWorkflowDocument server action
      router.push('/workflows/requisitions')
    } catch (error) {
      console.error('Failed to create requisition:', error)
    }
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">
          {currentStep === 'form' ? 'Create Requisition' : 'Preview & Submit'}
        </h1>
        <p className="text-sm text-muted-foreground">
          {currentStep === 'form'
            ? 'Fill in the requisition details and items'
            : 'Review your requisition before submitting'}
        </p>
      </div>

      {/* Progress Indicator */}
      <div className="flex items-center gap-4">
        <div
          className={`flex h-8 w-8 items-center justify-center rounded-full border-2 ${
            currentStep === 'form'
              ? 'border-primary bg-primary text-primary-foreground'
              : 'border-secondary bg-secondary text-secondary-foreground'
          }`}
        >
          1
        </div>
        <div className="flex-1 h-1 bg-border" />
        <div
          className={`flex h-8 w-8 items-center justify-center rounded-full border-2 ${
            currentStep === 'preview'
              ? 'border-primary bg-primary text-primary-foreground'
              : 'border-muted bg-muted text-muted-foreground'
          }`}
        >
          2
        </div>
      </div>

      {/* Content */}
      {currentStep === 'form' ? (
        <CreateRequisitionForm onSubmit={handleFormSubmit} initialData={formData} />
      ) : (
        <FormPreview
          formData={formData}
          onBack={handleBack}
          onSubmit={handleSubmit}
        />
      )}
    </div>
  )
}
