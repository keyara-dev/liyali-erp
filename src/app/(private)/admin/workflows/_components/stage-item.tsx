'use client'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Edit2, Trash2, GripVertical } from 'lucide-react'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { WorkflowStage } from '../create/_components/create-workflow-client'

interface StageItemProps {
  stage: WorkflowStage
  onEdit: () => void
  onDelete: () => void
}

const APPROVER_ROLE_LABELS: Record<string, string> = {
  DEPARTMENT_MANAGER: 'Department Manager',
  FINANCE_OFFICER: 'Finance Officer',
  CFO: 'CFO',
  WAREHOUSE_MANAGER: 'Warehouse Manager',
  PROCUREMENT_OFFICER: 'Procurement Officer',
  ADMIN: 'Admin',
}

export function StageItem({ stage, onEdit, onDelete }: StageItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: stage.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  return (
    <div ref={setNodeRef} style={style} className="w-full">
      <Card className="border-l-4 border-l-blue-500">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-start gap-3 flex-1">
              <button
                {...attributes}
                {...listeners}
                className="text-muted-foreground hover:text-foreground cursor-grab active:cursor-grabbing mt-1"
              >
                <GripVertical className="h-4 w-4" />
              </button>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <div className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-500 text-white text-xs font-medium">
                    {stage.order}
                  </div>
                  <CardTitle className="text-base">{stage.name}</CardTitle>
                </div>
                {stage.description && (
                  <p className="text-sm text-muted-foreground mt-1">
                    {stage.description}
                  </p>
                )}
              </div>
            </div>

            <div className="flex gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={onEdit}
                className="gap-2"
              >
                <Edit2 className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={onDelete}
                className="gap-2 text-destructive hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Approver Role</p>
              <p className="font-medium">
                {APPROVER_ROLE_LABELS[stage.approverRole] || stage.approverRole}
              </p>
            </div>
            <div>
              <p className="text-muted-foreground">Required Approvals</p>
              <p className="font-medium">
                {stage.requiredApprovals === 5 ? 'All' : stage.requiredApprovals}
              </p>
            </div>
            <div className="col-span-2">
              <p className="text-muted-foreground mb-1">Permissions</p>
              <div className="flex gap-4 text-xs">
                {stage.canReject && (
                  <span className="bg-green-100 text-green-800 px-2 py-1 rounded">
                    Can Reject
                  </span>
                )}
                {stage.canReassign && (
                  <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded">
                    Can Reassign
                  </span>
                )}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
