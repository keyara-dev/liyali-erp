'use client'

import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Edit2, Trash2, GripVertical } from 'lucide-react'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { WorkflowStage } from '@/types/workflow-config'

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
  } = useSortable({ id: stage.id || `stage-${stage.stageNumber || stage.order || 0}` })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const approverRoleLabel = APPROVER_ROLE_LABELS[stage.approverRole as keyof typeof APPROVER_ROLE_LABELS] || stage.approverRole || stage.requiredRole || 'Not Set'
  const requiredApprovalsLabel = stage.requiredApprovals === 5 ? 'All' : stage.requiredApprovals
  const hasPermissions = stage.canReject || stage.canReassign

  return (
    <div ref={setNodeRef} style={style} className="w-full">
      <Card className="border-l-4 border-l-blue-500">
        <CardContent className="p-4">
          <div className="flex items-center justify-between gap-6">
            {/* Left Section: Stage Info */}
            <div className="flex items-start gap-4 flex-1 min-w-0">
              {/* Drag Handle */}
              <button
                {...attributes}
                {...listeners}
                className="text-muted-foreground hover:text-foreground cursor-grab active:cursor-grabbing shrink-0 mt-1"
              >
                <GripVertical className="h-4 w-4" />
              </button>

              {/* Stage Number & Title */}
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-500 text-white text-sm font-semibold shrink-0">
                  {stage.order}
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-base leading-tight">
                    {stage.name}
                  </h3>
                  {stage.description && (
                    <p className="text-sm text-muted-foreground line-clamp-2 mt-0.5">
                      {stage.description}
                    </p>
                  )}
                </div>
              </div>
            </div>

            {/* Middle Section: Stage Details */}
            <div className="flex items-center gap-8 shrink-0">
              {/* Approver Role */}
              <div className="text-center">
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  Approver Role
                </p>
                <p className="text-sm font-semibold mt-1">
                  {approverRoleLabel}
                </p>
              </div>

              {/* Required Approvals */}
              <div className="text-center">
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  Required Approvals
                </p>
                <p className="text-sm font-semibold mt-1">
                  {requiredApprovalsLabel}
                </p>
              </div>

              {/* Permissions */}
              {hasPermissions && (
                <div className="text-center">
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">
                    Permissions
                  </p>
                  <div className="flex gap-2">
                    {stage.canReject && (
                      <span className="inline-flex items-center bg-green-100 text-green-800 px-2 py-1 rounded text-xs font-medium">
                        Can Reject
                      </span>
                    )}
                    {stage.canReassign && (
                      <span className="inline-flex items-center bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs font-medium">
                        Can Reassign
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Right Section: Actions */}
            <div className="flex gap-1 shrink-0">
              <Button
                variant="ghost"
                size="sm"
                onClick={onEdit}
                className="h-8 w-8 p-0"
                title="Edit stage"
              >
                <Edit2 className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={onDelete}
                className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                title="Delete stage"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
