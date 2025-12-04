# Workflow Builder Deep Dive Analysis

## Executive Summary

The Workflow Builder is a **React-based visual workflow designer** that allows administrators to create multi-stage approval workflows. It combines:
- **Drag-and-drop UI** for stage reordering (dnd-kit)
- **Modal dialogs** for stage configuration
- **Real-time form validation** with error feedback
- **State management** using React hooks
- **Two-tier validation**: Stage-level and form-level

The builder is currently a **complete client-side solution** that manages an in-memory data model, ready for backend integration.

---

## Component Architecture

### Component Hierarchy

```
WorkflowBuilder (Main orchestrator)
├── WorkflowDetailsForm
│   ├── Input (name)
│   ├── Textarea (description)
│   ├── Select (documentType)
│   └── Checkbox (isDefault)
│
├── Card (Stages Section)
│   └── DndContext (Drag-drop context)
│       └── SortableContext
│           └── StageItem[] (Draggable stage cards)
│               ├── GripVertical (Drag handle)
│               ├── Edit button
│               ├── Delete button
│               └── ArrowRight (Visual connector)
│
└── Dialog (Stage editor modal)
    └── StageForm
        ├── Input (stage name)
        ├── Textarea (description)
        ├── Select (approverRole)
        ├── Select (requiredApprovals)
        ├── Checkbox (canReject)
        └── Checkbox (canReassign)
```

---

## State Management Deep Dive

### Main State Variables

```typescript
// WorkflowBuilder state
const [formData, setFormData] = useState<WorkflowFormData>({
  name: '',
  description: '',
  documentType: 'REQUISITION',
  stages: [],
  isDefault: false,
})

const [showStageDialog, setShowStageDialog] = useState(false)
const [editingStageId, setEditingStageId] = useState<string | null>(null)
const [stageErrors, setStageErrors] = useState<Record<string, string>>({})
const [formErrors, setFormErrors] = useState<Record<string, string>>({})
```

### State Update Flow

```
Initial State
    ↓
User interacts (add/edit/delete/reorder)
    ↓
Handler function triggered
    ↓
Validation (if applicable)
    ↓
setFormData() updates root state
    ↓
Component re-renders
    ↓
UI reflects new state
```

---

## Data Model Structure

### WorkflowFormData (Top-level model)

```typescript
interface WorkflowFormData {
  name: string                    // e.g., "Standard Requisition Approval"
  description: string             // e.g., "4-stage approval process"
  documentType: string            // REQUISITION | PURCHASE_ORDER | etc.
  stages: WorkflowStage[]         // Array of approval stages (max 5)
  isDefault: boolean              // Mark as default for this docType
}
```

### WorkflowStage (Individual stage model)

```typescript
interface WorkflowStage {
  id: string                      // Unique ID: stage-${Date.now()}
  order: number                   // Position in workflow (1, 2, 3...)
  name: string                    // e.g., "Department Manager Review"
  description: string             // Purpose of this stage
  approverRole: string            // Required role (DEPARTMENT_MANAGER, CFO, etc.)
  requiredApprovals: number       // 1, 2, 3, or 5 (all approvals needed)
  canReject: boolean              // Stage permission
  canReassign: boolean            // Stage permission
}
```

---

## Lifecycle: Adding a New Stage

### Step-by-Step Flow

```
User clicks "Add Stage" button
        ↓
handleAddStage() called
        ↓
Check: formData.stages.length >= 5?
   ├─ YES → Toast error, return
   └─ NO → Continue
        ↓
setEditingStageId(null)         // Clear edit state
setShowStageDialog(true)        // Open dialog
        ↓
Dialog renders StageForm with empty initial data
        ↓
User fills form:
  - Stage name
  - Description
  - Approver role
  - Required approvals
  - Permissions (canReject, canReassign)
        ↓
User clicks "Add Stage" button in form
        ↓
handleSaveStage(stage) called
        ↓
validateStage(stage) runs
   ├─ ERRORS FOUND:
   │   ├─ setStageErrors(errors)
   │   └─ return (form stays open)
   └─ NO ERRORS: Continue
        ↓
editingStageId === null?
   ├─ YES (Creating new stage):
   │   ├─ Generate ID: stage-${Date.now()}
   │   ├─ Set order: formData.stages.length + 1
   │   ├─ Add to stages array
   │   ├─ setFormData with updated stages
   │   └─ Toast: "Stage added"
   └─ NO (Editing existing):
       └─ [See Editing flow below]
        ↓
setShowStageDialog(false)       // Close dialog
setStageErrors({})              // Clear errors
```

### Code Implementation

```typescript
// Handler that opens the dialog
const handleAddStage = () => {
  if (formData.stages.length >= 5) {
    toast.error('Maximum 5 stages allowed per workflow')
    return
  }
  setEditingStageId(null)              // Create mode
  setShowStageDialog(true)              // Open dialog
}

// Handler that saves the stage
const handleSaveStage = (stage: WorkflowStage) => {
  const errors = validateStage(stage)
  if (Object.keys(errors).length > 0) {
    setStageErrors(errors)              // Show errors, stay open
    return
  }

  if (editingStageId) {
    // Edit mode: update existing stage
    const updatedStages = formData.stages.map((s) =>
      s.id === editingStageId ? stage : s
    )
    setFormData({ ...formData, stages: updatedStages })
    toast.success('Stage updated')
  } else {
    // Create mode: add new stage
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
```

---

## Lifecycle: Editing a Stage

### Step-by-Step Flow

```
User clicks Edit icon on a stage card
        ↓
handleEditStage(stageId) called
        ↓
setEditingStageId(stageId)              // Set edit state
setShowStageDialog(true)                // Open dialog
        ↓
Dialog renders StageForm with:
  editingStage = formData.stages.find(s => s.id === stageId)
  (Pre-populated with current values)
        ↓
User modifies fields
        ↓
User clicks "Update Stage" button
        ↓
handleSaveStage(updatedStage) called
        ↓
validateStage(updatedStage) runs
   ├─ ERRORS: setStageErrors(errors), return
   └─ NO ERRORS: Continue
        ↓
editingStageId !== null? YES
        ↓
Map through stages array:
  - If stage.id === editingStageId: replace with updatedStage
  - Otherwise: keep as is
        ↓
setFormData with updated stages array
        ↓
Toast: "Stage updated"
setShowStageDialog(false)
setStageErrors({})
```

### Code Details

```typescript
const handleEditStage = (stageId: string) => {
  setEditingStageId(stageId)
  setShowStageDialog(true)
}

// Find stage to pre-populate form
const editingStage = editingStageId
  ? formData.stages.find((s) => s.id === editingStageId)
  : null

// Pass to StageForm
<StageForm
  stage={editingStage}           // Pre-populated data
  onSave={handleSaveStage}
  onCancel={() => setShowStageDialog(false)}
  errors={stageErrors}
/>
```

---

## Lifecycle: Deleting a Stage

### Step-by-Step Flow

```
User clicks Delete (trash icon) on stage card
        ↓
handleDeleteStage(stageId) called
        ↓
Filter out stage with matching ID:
  newStages = formData.stages.filter(s => s.id !== stageId)
        ↓
Re-calculate order for remaining stages:
  newStages = newStages.map((s, idx) => ({
    ...s,
    order: idx + 1    // Renumber: 1, 2, 3...
  }))
        ↓
setFormData with updated stages
        ↓
Toast: "Stage removed"
```

### Code Implementation

```typescript
const handleDeleteStage = (stageId: string) => {
  const newStages = formData.stages
    .filter((s) => s.id !== stageId)
    .map((s, idx) => ({
      ...s,
      order: idx + 1,               // Auto-renumber
    }))
  setFormData({ ...formData, stages: newStages })
  toast.success('Stage removed')
}
```

**Important**: After deletion, the `order` property is automatically recalculated to maintain sequence integrity.

---

## Lifecycle: Reordering Stages (Drag-and-Drop)

### dnd-kit Configuration

```typescript
// Setup sensors for drag interaction
const sensors = useSensors(
  useSensor(PointerSensor),                    // Mouse/touch drag
  useSensor(KeyboardSensor, {
    coordinateGetter: sortableKeyboardCoordinates,  // Keyboard arrow keys
  })
)

// DndContext wraps the sortable stages
<DndContext
  sensors={sensors}
  collisionDetection={closestCenter}          // Snap to closest stage
  onDragEnd={handleDragEnd}                   // Callback when drag ends
>
  <SortableContext
    items={formData.stages.map((s) => s.id)} // Array of sortable IDs
    strategy={verticalListSortingStrategy}   // Vertical list layout
  >
    {/* StageItem components rendered here */}
  </SortableContext>
</DndContext>
```

### Drag-and-Drop Flow

```
User starts dragging a stage card
        ↓
dnd-kit detects active.id (dragged item)
        ↓
Visual feedback:
  - Stage opacity: 0.5 (isDragging)
  - Drag cursor appears
        ↓
User moves cursor over another stage
        ↓
collisionDetection algorithm:
  - Finds closestCenter item
  - Updates visual preview
        ↓
User releases mouse
        ↓
handleDragEnd({ active, over }) called
        ↓
Check: over && active.id !== over.id?
   ├─ YES: Reorder needed
   └─ NO: No change, return
        ↓
Find indices in stages array:
  oldIndex = formData.stages.findIndex(s => s.id === active.id)
  newIndex = formData.stages.findIndex(s => s.id === over.id)
        ↓
Use arrayMove helper from @dnd-kit/sortable:
  newStages = arrayMove(formData.stages, oldIndex, newIndex)
        ↓
Re-calculate order property:
  newStages = newStages.map((stage, idx) => ({
    ...stage,
    order: idx + 1    // Renumber based on new position
  }))
        ↓
setFormData({ ...formData, stages: newStages })
        ↓
Component re-renders with new order
```

### Code Implementation

```typescript
const handleDragEnd = (event: DragEndEvent) => {
  const { active, over } = event

  if (over && active.id !== over.id) {
    // Find current positions
    const oldIndex = formData.stages.findIndex((s) => s.id === active.id)
    const newIndex = formData.stages.findIndex((s) => s.id === over.id)

    // arrayMove: swaps elements in array
    // Example: [A, B, C] moved 0→2 becomes [B, C, A]
    const newStages = arrayMove(formData.stages, oldIndex, newIndex).map(
      (stage, idx) => ({
        ...stage,
        order: idx + 1,    // Auto-increment order
      })
    )

    setFormData({ ...formData, stages: newStages })
  }
}
```

### StageItem Integration with dnd-kit

```typescript
// StageItem uses useSortable hook for dnd-kit
export function StageItem({ stage, onEdit, onDelete }: StageItemProps) {
  const {
    attributes,           // HTML attributes (data-* for dnd-kit)
    listeners,           // Event listeners (onPointerDown, etc.)
    setNodeRef,          // Register DOM node with dnd-kit
    transform,           // Current transform from dnd-kit
    transition,          // CSS transition
    isDragging,          // Boolean: is this item being dragged?
  } = useSortable({ id: stage.id })

  const style = {
    transform: CSS.Transform.toString(transform),  // Convert to CSS
    transition,
    opacity: isDragging ? 0.5 : 1,                // Visual feedback
  }

  return (
    <div ref={setNodeRef} style={style} className="w-full">
      <Card>
        <button
          {...attributes}
          {...listeners}              // Attach drag listeners here
          className="cursor-grab active:cursor-grabbing"
        >
          <GripVertical className="h-4 w-4" />
        </button>
        {/* Rest of card content */}
      </Card>
    </div>
  )
}
```

---

## Validation System

### Two-Level Validation

#### 1. Stage-Level Validation

```typescript
const validateStage = (stage: WorkflowStage): Record<string, string> => {
  const errors: Record<string, string> = {}

  // Rule 1: Stage name required
  if (!stage.name.trim()) {
    errors.name = 'Stage name is required'
  }

  // Rule 2: Approver role required
  if (!stage.approverRole.trim()) {
    errors.approverRole = 'Approver role is required'
  }

  // Rule 3: At least 1 approval required
  if (stage.requiredApprovals < 1) {
    errors.requiredApprovals = 'At least 1 approval is required'
  }

  return errors
}
```

**When called**: When user submits StageForm (modal)
**Result**: If errors exist, dialog stays open; if valid, stage is added/updated

#### 2. Form-Level Validation

```typescript
const validateForm = (): boolean => {
  const errors: Record<string, string> = {}

  // Rule 1: Workflow name required
  if (!formData.name.trim()) {
    errors.name = 'Workflow name is required'
  }

  // Rule 2: Document type required
  if (!formData.documentType) {
    errors.documentType = 'Document type is required'
  }

  // Rule 3: At least 1 stage required
  if (formData.stages.length === 0) {
    errors.stages = 'At least one stage is required'
  }

  setFormErrors(errors)
  return Object.keys(errors).length === 0  // Returns true if valid
}
```

**When called**: Just before submitting workflow (final validation)
**Result**: If errors exist, toast appears and form doesn't submit; if valid, onSubmit is called

### Error Clearing on Field Change

```typescript
const handleFormChange = (key: keyof WorkflowFormData, value: any) => {
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
```

**Benefit**: When user corrects an error, the error message disappears immediately (UX improvement).

---

## Full Workflow Submission Flow

### Create Mode

```
User clicks "Create Workflow" button
        ↓
handleSubmit() called
        ↓
validateForm() runs:
  - Check name not empty
  - Check documentType selected
  - Check stages.length > 0
        ↓
Validation failed?
   ├─ YES:
   │   ├─ setFormErrors(errors)
   │   ├─ Toast: "Please fix the errors..."
   │   └─ return (stop here)
   └─ NO: Continue
        ↓
onSubmit(formData) called (parent callback)
        ↓
Parent (CreateWorkflowClient) receives formData
        ↓
Parent validation logic:
  - API call to create workflow
  - Error handling
  - Navigation on success
```

### Code Flow

```typescript
// WorkflowBuilder
const handleSubmit = async () => {
  if (!validateForm()) {
    toast.error('Please fix the errors before submitting')
    return
  }

  await onSubmit(formData)  // Call parent's handler
}

// Parent (CreateWorkflowClient)
const handleSubmit = async (formData: WorkflowFormData) => {
  setIsSubmitting(true)
  try {
    // TODO: Call createWorkflow server action
    console.log('Creating workflow:', formData)

    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000))

    toast.success('Workflow created successfully')
    router.push('/admin/workflows')  // Redirect to list
  } catch (error) {
    console.error('Failed to create workflow:', error)
    toast.error('Failed to create workflow')
  } finally {
    setIsSubmitting(false)
  }
}
```

---

## UI Rendering System

### Stages Display with Visual Connectors

```
┌─────────────────────────────────────┐
│ Department Manager Review           │
│ ├─ Approve: next stage              │
│ └─ Reject: return to draft          │
└─────────────────────────────────────┘
            ↓ (ArrowRight rotated 90°)
┌─────────────────────────────────────┐
│ Finance Officer Review              │
│ ├─ Approve: next stage              │
│ └─ Reject: return to draft          │
└─────────────────────────────────────┘
            ↓ (ArrowRight rotated 90°)
┌─────────────────────────────────────┐
│ CFO Approval                        │
│ ├─ Approve: complete                │
│ └─ Reject: return to draft          │
└─────────────────────────────────────┘
```

### Code for Visual Connectors

```typescript
{formData.stages.map((stage, index) => (
  <div key={stage.id} className="flex items-start gap-3">
    <div className="flex flex-col items-center gap-2 pt-3">
      {/* Draggable stage card */}
      <StageItem
        stage={stage}
        onEdit={() => handleEditStage(stage.id)}
        onDelete={() => handleDeleteStage(stage.id)}
      />

      {/* Arrow connector (if not last stage) */}
      {index < formData.stages.length - 1 && (
        <ArrowRight className="h-4 w-4 text-muted-foreground rotate-90 mt-2" />
      )}
    </div>
  </div>
))}
```

**Logic**:
- Only show arrow if current stage is NOT the last one
- `rotate-90`: Rotates arrow from right-pointing (→) to down-pointing (↓)

---

## State Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    WorkflowBuilder State                     │
└─────────────────────────────────────────────────────────────┘

formData: WorkflowFormData
├── name: string
├── description: string
├── documentType: string
├── stages: WorkflowStage[]              ← Main data model
└── isDefault: boolean

showStageDialog: boolean                 ← Dialog visibility
editingStageId: string | null            ← Edit mode state
stageErrors: Record<string, string>      ← Stage form errors
formErrors: Record<string, string>       ← Workflow form errors

┌─────────────────────────────────────────────────────────────┐
│                    Event Flow Diagram                        │
└─────────────────────────────────────────────────────────────┘

                    User Action
                         ↓
    ┌────────────────────┬────────────────────┐
    │                    │                    │
  Add              Edit/Delete            Reorder
  Stage            Stage                  Stages
    │                    │                    │
    ├→ Open Dialog  ├→ Set editingId    ├→ handleDragEnd
    │   dialogId=null    └→ Open Dialog       │
    │                       dialogId=id       │
    │                                        │
    └→ User fills form                       │
       User submits                          │
                                             │
    ├→ validateStage()                       │
    │   ├─ Pass: add to stages              │
    │   └─ Fail: show errors                │
    │                                        │
    ├→ setFormData()                        ├→ arrayMove()
    │   Close Dialog                        └→ setFormData()
    │   Show toast                             Close Dialog
    │                                          Show toast
    └→ Re-render UI
```

---

## Comparison: Create vs Edit Mode

| Aspect | Create Mode | Edit Mode |
|--------|------------|-----------|
| **Initial Data** | Empty form | Pre-populated from initialData |
| **First Load** | Immediate | May load with suspense/skeleton |
| **Button Text** | "Create Workflow" | "Update Workflow" |
| **Data Origin** | All new | Merged with existing data |
| **Stage IDs** | Generated as new | Preserved from existing |
| **API Call** | POST /workflows | PATCH /workflows/{id} |

### Code Handling

```typescript
// Create mode
<WorkflowBuilder
  onSubmit={handleSubmit}
  isSubmitting={isSubmitting}
  mode="create"
  initialData={undefined}    // No initial data
/>

// Edit mode
<WorkflowBuilder
  onSubmit={handleSubmit}
  isSubmitting={isSubmitting}
  mode="edit"
  initialData={existingWorkflow}  // Pre-loaded data
/>

// Inside WorkflowBuilder
const [formData, setFormData] = useState<WorkflowFormData>(
  initialData || {              // Use initialData if provided, else default
    name: '',
    description: '',
    documentType: 'REQUISITION',
    stages: [],
    isDefault: false,
  }
)
```

---

## Dialog System (StageForm)

### Dialog States

```
┌─ Dialog Closed (showStageDialog = false)
│  └─ No StageForm rendered
│
└─ Dialog Open (showStageDialog = true)
   ├─ Add Mode (editingStageId = null)
   │  └─ StageForm with empty form
   │     └─ "Add Stage" button
   │
   └─ Edit Mode (editingStageId = "stage-123")
      └─ StageForm with pre-filled values
         └─ "Update Stage" button
```

### Dialog Integration

```typescript
// Trigger dialog opening
<Button onClick={handleAddStage}>Add Stage</Button>

// Dialog component
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
      stage={editingStage}           // null in create mode
      onSave={handleSaveStage}       // Both modes
      onCancel={() => setShowStageDialog(false)}
      errors={stageErrors}
    />
  </DialogContent>
</Dialog>

// Find stage for editing
const editingStage = editingStageId
  ? formData.stages.find((s) => s.id === editingStageId)
  : null
```

---

## Performance Considerations

### Re-render Optimization

```typescript
// ✅ Good: State updates are specific
setFormData({ ...formData, stages: newStages })  // Only updates stages

// ⚠️ Watch out: Unnecessary re-renders
// If parent re-renders, all StageItem components re-render
// Currently no memoization in place
```

### Potential Issues

1. **No useMemo for stages list**: Every render creates new array reference
   ```typescript
   // Current (inefficient for large lists)
   items={formData.stages.map((s) => s.id)}

   // Could be optimized with useMemo
   const stageIds = useMemo(
     () => formData.stages.map((s) => s.id),
     [formData.stages]
   )
   ```

2. **No React.memo on StageItem**: Each stage re-renders on any parent update
   ```typescript
   // Could wrap StageItem with memo
   export const StageItem = React.memo(function StageItem(...) {...})
   ```

3. **Stage ID generation using Date.now()**: Risky in rapid clicks
   ```typescript
   // Current
   id: `stage-${Date.now()}`  // Could collide in < 1ms

   // Better
   id: `stage-${crypto.randomUUID()}`  // Truly unique
   ```

---

## Integration Points

### Parent Component (CreateWorkflowClient)

```typescript
<WorkflowBuilder
  onSubmit={handleSubmit}         // Parent handler
  isSubmitting={isSubmitting}     // Disables buttons while loading
  mode="create"                   // Controls button text
  initialData={undefined}         // For edit mode
/>
```

### Child Components

```typescript
// WorkflowDetailsForm
<WorkflowDetailsForm
  data={formData}                 // Current form data
  onChange={handleFormChange}     // Update parent state
  errors={formErrors}             // Display errors
/>

// StageForm (in modal)
<StageForm
  stage={editingStage}            // Pre-fill values
  onSave={handleSaveStage}        // Add/update stage
  onCancel={() => setShowStageDialog(false)}
  errors={stageErrors}
/>

// StageItem (list)
<StageItem
  stage={stage}
  onEdit={() => handleEditStage(stage.id)}
  onDelete={() => handleDeleteStage(stage.id)}
/>
```

---

## Key Insights

### Design Patterns Used

1. **Container/Presentational Pattern**
   - WorkflowBuilder: Container (handles logic)
   - WorkflowDetailsForm, StageForm, StageItem: Presentational (just render)

2. **Controlled Components**
   - All form inputs are controlled (value from state)
   - Changes flow through handlers

3. **Modal Dialog Pattern**
   - StageForm lives in a Dialog
   - Modal state controlled by parent

4. **Optimistic Updates**
   - State updates immediately on user action
   - No waiting for backend

### What Makes It Work

1. **Immutable state updates**: `...formData` spreads ensure no mutations
2. **Auto-renumbering**: `order` property recalculated after any stage change
3. **Error isolation**: Stage errors don't block form submission
4. **Visual feedback**: Toasts and error messages keep user informed

### Limitations

1. **No undo/redo**: Changes are permanent until page refresh
2. **No concurrent editing**: Two users can't edit same workflow
3. **No draft saving**: Progress lost on page close
4. **Limited stage config**: Many advanced fields not exposed in UI
5. **No workflow versioning UI**: Type support exists, UI doesn't use it

---

## Complete Event Handler Map

| Handler | Triggered By | Actions |
|---------|--------------|---------|
| `handleAddStage()` | "Add Stage" button | Check max 5, open dialog |
| `handleEditStage(id)` | Edit icon | Set editingStageId, open dialog |
| `handleDeleteStage(id)` | Delete icon | Remove from array, renumber |
| `handleDragEnd(event)` | Drag release | Reorder stages, renumber |
| `handleSaveStage(stage)` | Form submit | Validate, add/update, close dialog |
| `validateStage(stage)` | handleSaveStage() | Check name, role, approvals |
| `handleFormChange(key, value)` | Input change | Update state, clear error |
| `validateForm()` | Before submit | Check workflow-level rules |
| `handleSubmit()` | Submit button | Validate, call parent onSubmit |

---

## Data Flow Summary

```
        User Input
            ↓
      Handler Function
            ↓
      Validation (optional)
            ↓
      setState(new data)
            ↓
    Component Re-renders
            ↓
         UI Updates
            ↓
    User Sees Changes
```

Example for adding a stage:
```
Click "Add Stage" → handleAddStage() → setShowStageDialog(true)
→ Dialog renders → User fills form → Click save
→ handleSaveStage(data) → validateStage() → setFormData()
→ Dialog closes → UI re-renders with new stage visible
```

---

## Next Steps for Production

1. **Replace mock onSubmit** with actual API call (server action)
2. **Add optimistic UI updates** while server processes
3. **Implement error recovery** (rollback on API failure)
4. **Add loading skeleton** for edit mode data fetching
5. **Persist unsaved changes** to browser storage
6. **Implement undo/redo** for better UX
7. **Add workflow versioning** UI
8. **Optimize re-renders** with memo and useMemo
