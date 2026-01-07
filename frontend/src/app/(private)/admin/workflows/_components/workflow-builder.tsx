'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Plus, Trash2, GripVertical, ArrowRight } from 'lucide-react'
import type { WorkflowFormData, WorkflowStage } from '@/app/_actions/workflows'
import { WorkflowDetailsForm } from './workflow-details-form'
import { StageForm } from './stage-form'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { StageItem } from './stage-item'
import { toast } from 'sonner'

interface WorkflowBuilderProps {
  onSubmit: (data: WorkflowFormData) => Promise<void>
  isSubmitting: boolean
  mode: 'create' | 'edit'
  initialData?: WorkflowFormData
}

export function WorkflowBuilder({
  onSubmit,
  isSubmitting,
  mode,
  initialData,
}: WorkflowBuilderProps) {
  const [formData, setFormData] = useState<WorkflowFormData>(
    initialData || {
      name: '',
      description: '',
      documentType: 'REQUISITION',
      stages: [],
      isDefault: false,
    }
  )
  const [showStageDialog, setShowStageDialog] = useState(false)
  const [editingStageId, setEditingStageId] = useState<string | null>(null)
  const [stageErrors, setStageErrors] = useState<Record<string, string>>({})
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event

    if (over && active.id !== over.id) {
      const oldIndex = formData.stages.findIndex((s) => s.id === active.id)
      const newIndex = formData.stages.findIndex((s) => s.id === over.id)

      const newStages = arrayMove(formData.stages, oldIndex, newIndex).map(
        (stage, idx) => ({
          ...stage,
          order: idx + 1,
        })
      )

      setFormData({ ...formData, stages: newStages })
    }
  }

  const handleAddStage = () => {
    if (formData.stages.length >= 5) {
      toast.error('Maximum 5 stages allowed per workflow')
      return
    }
    setEditingStageId(null)
    setShowStageDialog(true)
  }

  const handleEditStage = (stageId: string) => {
    setEditingStageId(stageId)
    setShowStageDialog(true)
  }

  const handleDeleteStage = (stageId: string) => {
    const newStages = formData.stages
      .filter((s) => s.id !== stageId)
      .map((s, idx) => ({
        ...s,
        order: idx + 1,
      }))
    setFormData({ ...formData, stages: newStages })
    toast.success('Stage removed')
  }

  const handleSaveStage = (stage: WorkflowStage) => {
    const errors = validateStage(stage)
    if (Object.keys(errors).length > 0) {
      setStageErrors(errors)
      return
    }

    if (editingStageId) {
      const updatedStages = formData.stages.map((s) =>
        s.id === editingStageId ? stage : s
      )
      setFormData({ ...formData, stages: updatedStages })
      toast.success('Stage updated')
    } else {
      const newStage = {
        ...stage,
        id: `stage-${Date.now()}`,
        order: formData.stages.length + 1,
      }
      setFormData({
        ...formData,
        stages: [...formData.stages, newStage],
      })
      toast.success('Stage added')
    }

    setShowStageDialog(false)
    setStageErrors({})
  }

  const validateStage = (stage: WorkflowStage): Record<string, string> => {
    const errors: Record<string, string> = {}

    if (!stage.name.trim()) {
      errors.name = 'Stage name is required'
    }
    if (!stage.approverRole.trim()) {
      errors.approverRole = 'Approver role is required'
    }
    if (stage.requiredApprovals < 1) {
      errors.requiredApprovals = 'At least 1 approval is required'
    }

    return errors
  }

  const handleFormChange = (key: keyof WorkflowFormData, value: any) => {
    setFormData((prev) => ({
      ...prev,
      [key]: value,
    }))
    if (formErrors[key]) {
      const newErrors = { ...formErrors }
      delete newErrors[key]
      setFormErrors(newErrors)
    }
  }

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {}

    if (!formData.name.trim()) {
      errors.name = 'Workflow name is required'
    }
    if (!formData.documentType) {
      errors.documentType = 'Document type is required'
    }
    if (formData.stages.length === 0) {
      errors.stages = 'At least one stage is required'
    }

    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async () => {
    if (!validateForm()) {
      toast.error('Please fix the errors before submitting')
      return
    }

    await onSubmit(formData)
  }

  const editingStage = editingStageId
    ? formData.stages.find((s) => s.id === editingStageId)
    : null

  return (
    <div className="space-y-6">
      {/* Workflow Details */}
      <WorkflowDetailsForm
        data={formData}
        onChange={handleFormChange}
        errors={formErrors}
      />

      {/* Stages Section */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle>Approval Stages</CardTitle>
          <Button onClick={handleAddStage} size="sm" className="gap-2">
            <Plus className="h-4 w-4" />
            Add Stage
          </Button>
        </CardHeader>
        <CardContent>
          {formData.stages.length === 0 ? (
            <div className="py-8 text-center">
              <p className="text-muted-foreground mb-4">
                No stages added yet. Create your first approval stage.
              </p>
              <Button onClick={handleAddStage} variant="outline">
                Add First Stage
              </Button>
            </div>
          ) : (
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={handleDragEnd}
            >
              <SortableContext
                items={formData.stages.map((s) => s.id)}
                strategy={verticalListSortingStrategy}
              >
                <div className="space-y-3">
                  {formData.stages.map((stage, index) => (
                    <div key={stage.id} className="flex flex-col items-center w-full">
                      <StageItem
                        stage={stage}
                        onEdit={() => handleEditStage(stage.id)}
                        onDelete={() => handleDeleteStage(stage.id)}
                      />
                      {index < formData.stages.length - 1 && (
                        <ArrowRight className="h-4 w-4 text-muted-foreground rotate-90 mt-2" />
                      )}
                    </div>
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          )}
          {formErrors.stages && (
            <p className="text-sm text-destructive mt-4">{formErrors.stages}</p>
          )}
        </CardContent>
      </Card>

      {/* Action Buttons */}
      <div className="flex gap-3 justify-end">
        <Button variant="outline" disabled={isSubmitting}>
          Cancel
        </Button>
        <Button onClick={handleSubmit} disabled={isSubmitting}>
          {isSubmitting
            ? 'Creating...'
            : mode === 'create'
              ? 'Create Workflow'
              : 'Update Workflow'}
        </Button>
      </div>

      {/* Stage Dialog */}
      <Dialog open={showStageDialog} onOpenChange={setShowStageDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingStageId ? 'Edit Stage' : 'Add Stage'}
            </DialogTitle>
            <DialogDescription>
              {editingStageId
                ? 'Update the stage details'
                : 'Create a new approval stage for your workflow'}
            </DialogDescription>
          </DialogHeader>
          <StageForm
            stage={editingStage}
            onSave={handleSaveStage}
            onCancel={() => setShowStageDialog(false)}
            errors={stageErrors}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}
