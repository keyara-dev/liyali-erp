# Workflow Builder - Code Reference Guide

## Quick Reference: File Structure

```
src/app/(private)/admin/workflows/
├── page.tsx                                    [25 lines]
│   └─ Server component
│   └─ Auth check + role validation
│   └─ Renders: WorkflowsClient
│
├── create/
│   ├── page.tsx                               [25 lines]
│   │   └─ Server component for create route
│   │   └─ Auth check + role validation
│   │   └─ Renders: CreateWorkflowClient
│   │
│   └── _components/
│       └── create-workflow-client.tsx         [79 lines]
│           ├─ 'use client' directive
│           ├─ Exports: WorkflowFormData, WorkflowStage interfaces
│           ├─ State: isSubmitting
│           ├─ Handlers: handleSubmit, handleBack
│           └─ Renders: PageHeader, WorkflowBuilder
│
├── [id]/edit/
│   ├── page.tsx                               [25 lines]
│   │   └─ Server component for edit route
│   │   └─ Gets [id] from params
│   │   └─ Renders: EditWorkflowClient
│   │
│   └── _components/
│       └── edit-workflow-client.tsx           [153 lines]
│           ├─ 'use client' directive
│           ├─ State: initialData, isLoading, isSubmitting
│           ├─ useEffect: fetch workflow data
│           ├─ Mock data map: mockWorkflows[id]
│           ├─ Handlers: handleSubmit, handleBack
│           └─ Renders: PageHeader, WorkflowBuilder with initialData
│
└── _components/
    ├── workflows-client.tsx                   [273 lines]
    │   ├─ 'use client' directive
    │   ├─ Mock data: mockWorkflows array
    │   ├─ State: workflows, deleteId, isDeleting
    │   ├─ Handlers: handleDelete, handleDuplicate
    │   └─ Renders: Table, AlertDialog
    │
    ├── workflow-builder.tsx                   [300 lines] ★ MAIN COMPONENT
    │   ├─ 'use client' directive
    │   ├─ State: formData, showStageDialog, editingStageId, errors
    │   ├─ dnd-kit setup: sensors, DndContext, SortableContext
    │   ├─ Handlers: add/edit/delete/reorder stages, form change, validation
    │   ├─ Validation: validateStage, validateForm
    │   └─ Renders: WorkflowDetailsForm, Card with DnD stages, Dialog with StageForm
    │
    ├── workflow-details-form.tsx              [104 lines]
    │   ├─ 'use client' directive
    │   ├─ Props: data, onChange, errors
    │   ├─ Constants: DOCUMENT_TYPES array
    │   └─ Renders: Input, Textarea, Select, Checkbox
    │
    ├── stage-form.tsx                         [187 lines]
    │   ├─ 'use client' directive
    │   ├─ Props: stage, onSave, onCancel, errors
    │   ├─ State: formData (WorkflowStage)
    │   ├─ Constants: APPROVER_ROLES array
    │   ├─ Handlers: handleChange, handleSubmit
    │   └─ Renders: Input, Textarea, Select, Checkbox
    │
    └── stage-item.tsx                         [123 lines]
        ├─ 'use client' directive
        ├─ Props: stage, onEdit, onDelete
        ├─ useSortable hook: for dnd-kit integration
        ├─ Constants: APPROVER_ROLE_LABELS map
        └─ Renders: Card with drag handle, stage info, action buttons
```

---

## Code Snippets: Core Implementation

### 1. WorkflowBuilder Main Component Structure

```typescript
export function WorkflowBuilder({
  onSubmit,
  isSubmitting,
  mode,
  initialData,
}: WorkflowBuilderProps) {
  // ==================== STATE ====================
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

  // ==================== DND-KIT SETUP ====================
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  // ==================== HANDLERS ====================
  const handleDragEnd = (event: DragEndEvent) => { ... }
  const handleAddStage = () => { ... }
  const handleEditStage = (stageId: string) => { ... }
  const handleDeleteStage = (stageId: string) => { ... }
  const handleSaveStage = (stage: WorkflowStage) => { ... }
  const validateStage = (stage: WorkflowStage) => { ... }
  const handleFormChange = (key: keyof WorkflowFormData, value: any) => { ... }
  const validateForm = (): boolean => { ... }
  const handleSubmit = async () => { ... }

  // ==================== COMPUTED ====================
  const editingStage = editingStageId
    ? formData.stages.find((s) => s.id === editingStageId)
    : null

  // ==================== RENDER ====================
  return (
    <div className="space-y-6">
      <WorkflowDetailsForm ... />
      <Card>...</Card>
      <div className="flex gap-3 justify-end">...</div>
      <Dialog>...</Dialog>
    </div>
  )
}
```

---

### 2. Drag-and-Drop Handler

```typescript
const handleDragEnd = (event: DragEndEvent) => {
  const { active, over } = event

  // Guard: Do nothing if dropped on same item or empty space
  if (over && active.id !== over.id) {
    // Find positions of active and over items
    const oldIndex = formData.stages.findIndex((s) => s.id === active.id)
    const newIndex = formData.stages.findIndex((s) => s.id === over.id)

    // Use arrayMove helper to reorder
    // arrayMove swaps elements: [A,B,C] with oldIndex=2, newIndex=0 → [C,A,B]
    const newStages = arrayMove(formData.stages, oldIndex, newIndex).map(
      (stage, idx) => ({
        ...stage,
        order: idx + 1,  // ← Auto-increment order
      })
    )

    // Update state with new order
    setFormData({ ...formData, stages: newStages })
  }
}

// Key learning:
// - arrayMove is from @dnd-kit/sortable
// - We manually increment the order property
// - order must match the array index for visual consistency
```

---

### 3. Add Stage Handler

```typescript
const handleAddStage = () => {
  // Enforce max 5 stages
  if (formData.stages.length >= 5) {
    toast.error('Maximum 5 stages allowed per workflow')
    return  // ← Exit early
  }

  // Clear edit state (we're creating, not editing)
  setEditingStageId(null)

  // Open dialog
  setShowStageDialog(true)
}

// Usage flow:
// Click "Add Stage" → handleAddStage() → Dialog opens
// User fills form → handleSaveStage(data) → Stage added
```

---

### 4. Save Stage Handler (Add/Update)

```typescript
const handleSaveStage = (stage: WorkflowStage) => {
  // ==================== VALIDATION ====================
  const errors = validateStage(stage)
  if (Object.keys(errors).length > 0) {
    setStageErrors(errors)
    return  // ← Exit early, dialog stays open
  }

  // ==================== ADD vs UPDATE ====================
  if (editingStageId) {
    // UPDATE: Replace stage with same id
    const updatedStages = formData.stages.map((s) =>
      s.id === editingStageId ? stage : s
    )
    setFormData({ ...formData, stages: updatedStages })
    toast.success('Stage updated')
  } else {
    // CREATE: Add new stage to end
    const newStage = {
      ...stage,
      id: `stage-${Date.now()}`,  // ⚠️ Timestamp-based ID
      order: formData.stages.length + 1,
    }
    setFormData({
      ...formData,
      stages: [...formData.stages, newStage],
    })
    toast.success('Stage added')
  }

  // ==================== CLEANUP ====================
  setShowStageDialog(false)
  setStageErrors({})
}

// Key learning:
// - Validation happens first
// - editingStageId determines add vs update
// - New stages get ID from Date.now() (not ideal, could collide)
// - New stages get order = array.length + 1 (sequential)
// - Dialog closes regardless of add/update
```

---

### 5. Validation Handlers

```typescript
// ==================== STAGE-LEVEL VALIDATION ====================
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

// Called when: User submits StageForm in modal
// Returns: Object of field → error message
// Used by: handleSaveStage() to decide if stage can be saved

// ==================== FORM-LEVEL VALIDATION ====================
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

// Called when: User clicks "Create/Update Workflow"
// Returns: true if valid, false if errors exist
// Effect: Errors set in state, displayed to user

// ==================== ERROR CLEARING ====================
const handleFormChange = (key: keyof WorkflowFormData, value: any) => {
  // Update state
  setFormData((prev) => ({
    ...prev,
    [key]: value,
  }))

  // Auto-clear error for this field when user starts editing
  if (formErrors[key]) {
    const newErrors = { ...formErrors }
    delete newErrors[key]
    setFormErrors(newErrors)
  }
}

// UX Improvement:
// When user corrects a field, error disappears immediately
// Doesn't wait for re-validation
```

---

### 6. Form Submission

```typescript
const handleSubmit = async () => {
  // ==================== VALIDATE ====================
  if (!validateForm()) {
    toast.error('Please fix the errors before submitting')
    return  // ← Exit early
  }

  // ==================== SUBMIT ====================
  // Call parent's callback with form data
  await onSubmit(formData)
}

// Parent receives formData: WorkflowFormData
// Parent is responsible for:
// - setIsSubmitting(true/false)
// - API call
// - Error handling
// - Navigation

// Data passed to parent:
// {
//   name: "Standard Requisition Approval",
//   description: "4-stage approval",
//   documentType: "REQUISITION",
//   stages: [
//     { id: "stage-1733328400000", order: 1, name: "...", ... },
//     { id: "stage-1733328401000", order: 2, name: "...", ... }
//   ],
//   isDefault: true
// }
```

---

### 7. DndContext Setup

```typescript
const sensors = useSensors(
  // Mouse/Touch drag detection
  useSensor(PointerSensor),

  // Keyboard arrow key support
  useSensor(KeyboardSensor, {
    coordinateGetter: sortableKeyboardCoordinates,
  })
)

// In JSX:
<DndContext
  sensors={sensors}                    // How drag is detected
  collisionDetection={closestCenter}   // Snap to closest stage
  onDragEnd={handleDragEnd}            // Called when drag ends
>
  <SortableContext
    items={formData.stages.map((s) => s.id)}      // IDs of sortable items
    strategy={verticalListSortingStrategy}        // Vertical list layout
  >
    <div className="space-y-3">
      {formData.stages.map((stage, index) => (
        <div key={stage.id} className="flex items-start gap-3">
          <div className="flex flex-col items-center gap-2 pt-3">
            <StageItem
              stage={stage}
              onEdit={() => handleEditStage(stage.id)}
              onDelete={() => handleDeleteStage(stage.id)}
            />
            {index < formData.stages.length - 1 && (
              <ArrowRight className="h-4 w-4 text-muted-foreground rotate-90 mt-2" />
            )}
          </div>
        </div>
      ))}
    </div>
  </SortableContext>
</DndContext>

// Key learning:
// - DndContext provides drag context
// - SortableContext provides sortable functionality
// - sensors define how drag is triggered
// - collisionDetection determines drop target
// - onDragEnd is the callback for drag completion
```

---

### 8. StageItem (Draggable Card)

```typescript
export function StageItem({ stage, onEdit, onDelete }: StageItemProps) {
  // ==================== DND-KIT HOOKS ====================
  const {
    attributes,           // HTML attributes for dnd-kit
    listeners,           // Event listeners (onPointerDown, etc.)
    setNodeRef,          // Register this DOM node
    transform,           // Current drag transform
    transition,          // CSS transition
    isDragging,          // Is this item currently being dragged?
  } = useSortable({ id: stage.id })

  // ==================== COMPUTE STYLE ====================
  const style = {
    transform: CSS.Transform.toString(transform),  // dnd-kit → CSS
    transition,
    opacity: isDragging ? 0.5 : 1,  // Fade when dragging
  }

  // ==================== RENDER ====================
  return (
    <div ref={setNodeRef} style={style} className="w-full">
      <Card className="border-l-4 border-l-blue-500">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-start gap-3 flex-1">
              {/* Drag Handle */}
              <button
                {...attributes}              // Spread dnd-kit attributes
                {...listeners}               // Spread dnd-kit listeners
                className="text-muted-foreground hover:text-foreground cursor-grab active:cursor-grabbing mt-1"
              >
                <GripVertical className="h-4 w-4" />
              </button>

              {/* Stage Info */}
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

            {/* Action Buttons */}
            <div className="flex gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={onEdit}
              >
                <Edit2 className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={onDelete}
                className="text-destructive hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>

        {/* Stage Details */}
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

// Key learning:
// - useSortable hook integrates with dnd-kit
// - attributes and listeners MUST be spread on drag handle
// - setNodeRef registers this component with dnd-kit
// - transform is updated by dnd-kit as user drags
// - isDragging provides visual feedback (opacity)
// - CSS.Transform.toString() converts transform to CSS
```

---

### 9. Delete Stage Handler

```typescript
const handleDeleteStage = (stageId: string) => {
  // ==================== FILTER ====================
  const newStages = formData.stages
    .filter((s) => s.id !== stageId)  // Remove matching stage
    .map((s, idx) => ({               // Renumber remaining stages
      ...s,
      order: idx + 1,                 // 1-indexed order
    }))

  // ==================== UPDATE STATE ====================
  setFormData({ ...formData, stages: newStages })

  // ==================== FEEDBACK ====================
  toast.success('Stage removed')
}

// Example:
// Before: stages = [stage-1(order:1), stage-2(order:2), stage-3(order:3)]
// Delete stage-2
// Filter: [stage-1, stage-3]
// Map with new order:
//   stage-1: order = 0+1 = 1 (unchanged)
//   stage-3: order = 1+1 = 2 (was 3)
// After: stages = [stage-1(order:1), stage-3(order:2)]

// Key learning:
// - Filter removes the stage
// - Map renumbers the order property
// - Order must be sequential starting from 1
// - This maintains display consistency
```

---

### 10. Edit Stage Handler

```typescript
const handleEditStage = (stageId: string) => {
  // Set which stage is being edited
  setEditingStageId(stageId)

  // Open dialog (same dialog as Add)
  setShowStageDialog(true)
}

// Then in render, compute editingStage:
const editingStage = editingStageId
  ? formData.stages.find((s) => s.id === editingStageId)
  : null

// Dialog uses editingStage for pre-population:
<StageForm
  stage={editingStage}                // null in create mode, object in edit
  onSave={handleSaveStage}            // Same handler for both modes
  onCancel={() => setShowStageDialog(false)}
  errors={stageErrors}
/>

// StageForm detects mode by checking if stage prop exists:
// - If stage is null → empty form, button says "Add Stage"
// - If stage is object → pre-filled form, button says "Update Stage"

// Key learning:
// - Same dialog used for add and edit
// - editingStageId determines the mode
// - StageForm updates or creates based on receiving stage data
// - handleSaveStage handles both paths
```

---

### 11. Dialog Structure

```typescript
const editingStage = editingStageId
  ? formData.stages.find((s) => s.id === editingStageId)
  : null

return (
  <div className="space-y-6">
    {/* ... other components ... */}

    {/* Stage Dialog - Controlled by showStageDialog state */}
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

        {/* StageForm component inside dialog */}
        <StageForm
          stage={editingStage}              // null or object
          onSave={handleSaveStage}          // Add/Update handler
          onCancel={() => setShowStageDialog(false)}  // Close
          errors={stageErrors}              // Display errors
        />
      </DialogContent>
    </Dialog>
  </div>
)

// Dialog lifecycle:
// 1. User clicks "Add Stage" → handleAddStage()
// 2. setShowStageDialog(true) → Dialog opens
// 3. StageForm renders with stage={null}
// 4. User fills form and clicks save
// 5. handleSaveStage() called
// 6. If valid: setShowStageDialog(false) → Dialog closes
// 7. If invalid: error shown, dialog stays open

// Key learning:
// - Dialog open state controlled by showStageDialog
// - Dialog content changes based on editingStageId
// - Same dialog reused for add and edit
// - Dialog doesn't automatically close on error
```

---

### 12. Parent Component Integration

```typescript
// CreateWorkflowClient.tsx
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
      // TODO: Replace with actual server action
      console.log('Creating workflow:', formData)

      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000))

      toast.success('Workflow created successfully')
      router.push('/admin/workflows')
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

      {/* Pass WorkflowBuilder props */}
      <WorkflowBuilder
        onSubmit={handleSubmit}          // Callback when workflow submitted
        isSubmitting={isSubmitting}      // Disables buttons while loading
        mode="create"                    // Controls button text
        initialData={undefined}          // No initial data for create
      />
    </div>
  )
}

// Key learning:
// - Parent manages isSubmitting state
// - Parent provides onSubmit callback
// - Parent handles API call and navigation
// - WorkflowBuilder is "dumb" container
// - Communication is one-way (data up to parent)
```

---

## Type Definitions Reference

```typescript
// Main workflow data structure
interface WorkflowFormData {
  name: string                    // Workflow name
  description: string             // Purpose
  documentType: string            // REQUISITION, PURCHASE_ORDER, etc.
  stages: WorkflowStage[]         // Array of approval stages
  isDefault: boolean              // Mark as default for doc type
}

// Individual approval stage
interface WorkflowStage {
  id: string                      // Unique ID (stage-${Date.now()})
  order: number                   // Display order (1, 2, 3...)
  name: string                    // Stage name
  description: string             // Purpose of this stage
  approverRole: string            // Required role (DEPARTMENT_MANAGER, etc.)
  requiredApprovals: number       // 1, 2, 3, or 5 (all)
  canReject: boolean              // Permission
  canReassign: boolean            // Permission
}

// Component props
interface WorkflowBuilderProps {
  onSubmit: (data: WorkflowFormData) => Promise<void>
  isSubmitting: boolean
  mode: 'create' | 'edit'
  initialData?: WorkflowFormData
}

interface WorkflowDetailsFormProps {
  data: WorkflowFormData
  onChange: (key: keyof WorkflowFormData, value: any) => void
  errors: Record<string, string>
}

interface StageFormProps {
  stage?: WorkflowStage | null
  onSave: (stage: WorkflowStage) => void
  onCancel: () => void
  errors: Record<string, string>
}

interface StageItemProps {
  stage: WorkflowStage
  onEdit: () => void
  onDelete: () => void
}
```

---

## Common Patterns Used

### Pattern 1: Lifting State Up

```typescript
// Child component (StageForm)
const handleSubmit = () => {
  onSave(formData)  // Call parent callback
}

// Parent component (WorkflowBuilder)
const handleSaveStage = (stage: WorkflowStage) => {
  // Receives data from child
  // Updates parent state
  setFormData({ ...formData, stages: [...formData.stages, stage] })
}

// Key: Parent manages state, child calls callbacks
```

### Pattern 2: Controlled Components

```typescript
// All inputs are controlled (value from state)
<Input
  value={formData.name}
  onChange={(e) => handleFormChange('name', e.target.value)}
/>

// Changes flow: User types → onChange fires → handleFormChange
// → setFormData → re-render → input shows new value
```

### Pattern 3: Immutable State Updates

```typescript
// ✓ Correct: Create new object, don't mutate
setFormData({ ...formData, stages: newStages })

// ✗ Wrong: Mutating existing object
formData.stages.push(newStage)
setFormData(formData)
```

### Pattern 4: Compound Components

```typescript
// Parent (WorkflowBuilder) orchestrates
<DndContext>
  <SortableContext>
    <StageItem />      // Child doesn't know about dnd-kit context
  </SortableContext>
</DndContext>

// StageItem uses useSortable hook to participate in context
const StageItem = () => {
  const { attributes, listeners } = useSortable({ id: stage.id })
  // Now it's sortable without knowing DndContext exists
}
```

### Pattern 5: Error Handling

```typescript
// Validate before action
const errors = validateStage(stage)
if (Object.keys(errors).length > 0) {
  setErrors(errors)
  return  // Early exit
}

// Proceed if valid
doAction()
```

---

## Performance Notes

### Current Bottlenecks

1. **No Memoization**: Every render causes all children to re-render
2. **Stage ID Generation**: `Date.now()` not guaranteed unique
3. **Array Lookup**: Finding stage by ID uses `.findIndex()` O(n)

### Optimization Ideas

```typescript
// Memoize the stages map
const stageIds = useMemo(
  () => formData.stages.map((s) => s.id),
  [formData.stages]
)

// Memoize stage item component
const MemoizedStageItem = React.memo(StageItem)

// Better ID generation
import { v4 as uuidv4 } from 'uuid'
id: uuidv4()

// Memoize handlers with useCallback
const handleEditStage = useCallback((id: string) => {
  setEditingStageId(id)
  setShowStageDialog(true)
}, [])
```

---

## Testing Checklist

```typescript
describe('WorkflowBuilder', () => {
  it('should add a stage when Add Stage is clicked', () => {
    // Click Add Stage button
    // Fill form
    // Click save
    // Expect stage in list
  })

  it('should validate stage name is required', () => {
    // Click Add Stage
    // Leave name empty
    // Click save
    // Expect error message
  })

  it('should enforce max 5 stages', () => {
    // Add 5 stages
    // Click Add Stage again
    // Expect error toast
  })

  it('should reorder stages on drag drop', () => {
    // Drag stage 3 to position 1
    // Expect order property updated
    // Expect UI reflects new order
  })

  it('should validate form before submit', () => {
    // Leave workflow name empty
    // Click Create
    // Expect error, stay on page
  })
})
```

---

## Migration Checklist: Mock → Real API

```typescript
// Current (Mock)
const handleSubmit = async (formData: WorkflowFormData) => {
  await new Promise((resolve) => setTimeout(resolve, 1000))
  toast.success('Workflow created')
  router.push('/admin/workflows')
}

// Target (Real API)
const handleSubmit = async (formData: WorkflowFormData) => {
  try {
    const response = await fetch('/api/workflows', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(formData)
    })

    if (!response.ok) {
      throw new Error('Failed to create workflow')
    }

    const result = await response.json()
    toast.success('Workflow created')
    router.push(`/admin/workflows/${result.id}`)
  } catch (error) {
    toast.error(error.message)
  } finally {
    setIsSubmitting(false)
  }
}

// Or with Server Actions (Recommended)
const handleSubmit = async (formData: WorkflowFormData) => {
  try {
    const result = await createWorkflow(formData)
    toast.success('Workflow created')
    router.push(`/admin/workflows/${result.id}`)
  } catch (error) {
    toast.error('Failed to create workflow')
  }
}
```
