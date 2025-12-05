# Workflow Builder - Architecture Diagram

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     WORKFLOW BUILDER SYSTEM                              │
└─────────────────────────────────────────────────────────────────────────┘

LAYER 1: PAGES (Server Components)
┌──────────────────────────┐  ┌──────────────────────────┐  ┌──────────────┐
│  /admin/workflows        │  │  /admin/workflows/create │  │ /admin/[id]  │
│  page.tsx                │  │  page.tsx                │  │ /edit        │
│  (Auth check)            │  │  (Auth check)            │  │ page.tsx     │
│  → WorkflowsClient       │  │  → CreateWorkflowClient  │  │ (Auth check) │
│                          │  │                          │  │ → EditClient │
└──────────────────────────┘  └──────────────────────────┘  └──────────────┘

LAYER 2: CLIENT COMPONENTS (React 'use client')
┌────────────────────────────────────────────────────────────────────────┐
│ CreateWorkflowClient / EditWorkflowClient (Parent)                     │
│  ├─ State: isSubmitting                                                │
│  ├─ Handlers: handleSubmit (API call), handleBack (navigation)        │
│  └─ Props down: onSubmit, isSubmitting, mode                          │
└────────────────────────────────────────────────────────────────────────┘
                                    ↓

LAYER 3: BUILDER ORCHESTRATOR
┌────────────────────────────────────────────────────────────────────────┐
│ WorkflowBuilder (Main State Container)                                 │
│                                                                         │
│ State Management:                                                       │
│  ├─ formData: WorkflowFormData                                          │
│  │   ├─ name, description, documentType, stages, isDefault            │
│  │   └─ Updated by: handleFormChange, handleSaveStage, etc.           │
│  ├─ showStageDialog: boolean                                           │
│  │   └─ Updated by: handleAddStage, handleEditStage, handleSaveStage  │
│  ├─ editingStageId: string | null                                      │
│  │   └─ Updated by: handleEditStage, handleAddStage                   │
│  ├─ stageErrors: Record<string, string>                               │
│  │   └─ Updated by: handleSaveStage (validate), handleSaveStage       │
│  └─ formErrors: Record<string, string>                                │
│      └─ Updated by: handleSubmit (validate), handleFormChange (clear) │
│                                                                         │
│ Drag-Drop Setup:                                                        │
│  ├─ sensors: [PointerSensor, KeyboardSensor]                          │
│  ├─ collisionDetection: closestCenter                                  │
│  └─ handler: handleDragEnd                                             │
│                                                                         │
│ Handlers (Event Functions):                                             │
│  ├─ handleAddStage()           → Opens dialog in create mode          │
│  ├─ handleEditStage(id)        → Opens dialog in edit mode            │
│  ├─ handleDeleteStage(id)      → Removes stage, renumbers             │
│  ├─ handleSaveStage(stage)     → Adds/updates stage                   │
│  ├─ handleDragEnd(event)       → Reorders stages                      │
│  ├─ handleFormChange(key, val) → Updates form data                    │
│  ├─ validateStage(stage)       → Stage-level validation               │
│  ├─ validateForm()             → Form-level validation                │
│  └─ handleSubmit()             → Calls parent's onSubmit              │
└────────────────────────────────────────────────────────────────────────┘
        ↓                    ↓                    ↓                ↓

LAYER 4A: FORM INPUTS          LAYER 4B: STAGES LIST        LAYER 4C: DIALOG

┌──────────────────────────┐   ┌──────────────────────────┐  ┌─────────────┐
│WorkflowDetailsForm       │   │ Stages Card (DnD)        │  │ Dialog      │
│  ├─ Input: name          │   │  ├─ DndContext           │  │  └─Content: │
│  ├─ Textarea: desc       │   │  │  ├─ SortableContext   │  │    StageForm│
│  ├─ Select: docType      │   │  │  │  └─ StageItem[]    │  │             │
│  └─ Checkbox: isDefault  │   │  │  │     ├─ Drag Handle  │  │ Functions:  │
│                          │   │  │  │     ├─ Edit btn     │  │  ├─ onSave  │
│ Props:                   │   │  │  │     ├─ Delete btn   │  │  └─ onCanc  │
│  ├─ data (formData)      │   │  │  │     └─ Arrows       │  │             │
│  ├─ onChange             │   │  │  │                      │  │ Props:      │
│  └─ errors               │   │  │  └─ onDragEnd         │  │  ├─ stage   │
│                          │   │  │     callback           │  │  ├─ onSave  │
│ Render:                  │   │  └─ Empty state          │  │  ├─ onCancel│
│  ├─ Shadcn Input         │   │     (no stages)          │  │  └─ errors  │
│  ├─ Shadcn Textarea      │   │                          │  │             │
│  ├─ Shadcn Select        │   │ Render:                  │  │ Render:     │
│  └─ Shadcn Checkbox      │   │  ├─ Stages loop         │  │  ├─ Header  │
│                          │   │  ├─ Add Stage button    │  │  ├─ Form    │
│ Updates: formData.{name, │   │  └─ Validation error    │  │  │ ├─ Name  │
│           description,   │   │     display             │  │  │ ├─ Role  │
│           documentType,  │   │                          │  │  │ ├─ Approv│
│           isDefault}     │   │ Updates formData.stages │  │  │ └─ Perms │
│                          │   │  via handlers           │  │  └─ Buttons │
└──────────────────────────┘   └──────────────────────────┘  └─────────────┘

LAYER 5: DND-KIT INTEGRATION

┌─────────────────────────────────────────────────────────────────────────┐
│ dnd-kit (Drag-and-Drop Library)                                         │
│                                                                          │
│ Components:                                                              │
│  ├─ DndContext                                                          │
│  │  ├─ Props: sensors, collisionDetection, onDragEnd                   │
│  │  └─ Children: SortableContext                                       │
│  │                                                                       │
│  └─ SortableContext                                                     │
│     ├─ Props: items (array of IDs), strategy (vertical)               │
│     └─ Children: StageItem components (must use useSortable)          │
│                                                                          │
│ Within StageItem:                                                       │
│  └─ useSortable Hook                                                    │
│     ├─ Returns: { attributes, listeners, setNodeRef, transform,... }  │
│     ├─ attributes: {...} spread on drag handle                         │
│     ├─ listeners: {...} spread on drag handle                          │
│     ├─ setNodeRef: Register DOM node with dnd-kit                      │
│     ├─ transform: Current drag position                                │
│     ├─ isDragging: Visual feedback boolean                             │
│     └─ Used to: Apply CSS transform, opacity changes                   │
│                                                                          │
│ Drag Event Flow:                                                        │
│  1. User mouseDown on grip handle                                      │
│  2. dnd-kit detects drag start → active.id set                         │
│  3. User moves mouse → transform computed                              │
│  4. dnd-kit updates isDragging, collision detection                    │
│  5. User releases → onDragEnd callback fires                           │
│  6. handleDragEnd processes: { active, over }                          │
│  7. arrayMove reorders array, order recalculated                       │
│  8. State updates, component re-renders                                │
└─────────────────────────────────────────────────────────────────────────┘

LAYER 6: UI COMPONENTS (Shadcn)

┌─────────────────────────────────────────────────────────────────────────┐
│ Shadcn Components Used:                                                  │
│  ├─ Button (with variants: default, outline, ghost)                    │
│  ├─ Card (header, content)                                             │
│  ├─ Input (text)                                                        │
│  ├─ Textarea (multi-line)                                              │
│  ├─ Select (dropdown)                                                  │
│  ├─ Checkbox (boolean toggle)                                          │
│  ├─ Dialog (modal)                                                      │
│  ├─ AlertDialog (delete confirmation)                                  │
│  └─ Table (workflows list)                                             │
│                                                                          │
│ Icons from lucide-react:                                                │
│  ├─ Plus (add button)                                                  │
│  ├─ Edit2 (edit button)                                                │
│  ├─ Trash2 (delete button)                                             │
│  ├─ GripVertical (drag handle)                                         │
│  └─ ArrowRight (visual connector, rotated 90°)                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Diagram (Complete)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         DATA FLOW ARCHITECTURE                           │
└─────────────────────────────────────────────────────────────────────────┘

USER ACTION
    ↓
    ├─────────────────────────────────────────────────────────────────────┐
    │                                                                     │
    │ Add Stage              Edit Stage            Delete Stage          │
    │    ↓                      ↓                      ↓                   │
    │ handleAddStage()     handleEditStage()    handleDeleteStage()      │
    │    │                      │                      │                  │
    │    ├─ Check max?          ├─ Set editingId      └─ Filter array    │
    │    ├─ setEditingId=null   ├─ Show dialog            ├─ Remove      │
    │    └─ Show dialog         └─ Find stage             └─ Renumber    │
    │                              for pre-fill              order        │
    │         ↓                         ↓                      ↓          │
    │    Dialog opens         Dialog opens           formData.stages     │
    │    (empty form)         (pre-filled)           updated             │
    │         │                   │                        │             │
    │         ↓                   ↓                        ↓             │
    │    User fills          User modifies         UI re-renders       │
    │    form                form                                       │
    │         │                   │                                     │
    │         └───────┬───────────┘                                     │
    │                 ↓                                                  │
    │            User clicks save/update                                │
    │                 ↓                                                  │
    │         handleSaveStage(stage)                                    │
    │                 │                                                  │
    │                 ├─ validateStage()                                │
    │                 │   ├─ name.trim()?    ✓/✗                        │
    │                 │   ├─ role.trim()?    ✓/✗                        │
    │                 │   └─ approvals >= 1? ✓/✗                        │
    │                 │                                                  │
    │                 ├─ Errors? → setStageErrors, return              │
    │                 │   (dialog stays open)                           │
    │                 │                                                  │
    │                 └─ No errors → Continue                           │
    │                     ├─ editingStageId? YES                        │
    │                     │   └─ Update stage in array                  │
    │                     │   └─ toast.success('updated')               │
    │                     │                                              │
    │                     ├─ editingStageId? NO                         │
    │                     │   ├─ Generate ID: stage-${Date.now()}       │
    │                     │   ├─ Calculate order: length + 1            │
    │                     │   ├─ Add to array                           │
    │                     │   └─ toast.success('added')                 │
    │                     │                                              │
    │                     └─ setFormData({...})                         │
    │                        setShowStageDialog(false)                  │
    │                        setStageErrors({})                         │
    │                                                                     │
    │                            ↓                                        │
    │         Component re-renders (new state)                          │
    │         Dialog closes, stage visible in list                      │
    │                                                                     │
    └─────────────────────────────────────────────────────────────────────┘

DRAG-AND-DROP PATH
    ↓
    User clicks + drags grip handle
        ↓
    dnd-kit detects: active.id = "stage-2"
        ↓
    User drags over stage-1
        ↓
    dnd-kit detects: over.id = "stage-1"
        ↓
    User releases mouse
        ↓
    handleDragEnd({ active, over })
        ├─ active.id === over.id? NO → Continue
        │
        ├─ Find indices:
        │  ├─ oldIndex = formData.stages.findIndex(s => s.id === active.id)
        │  └─ newIndex = formData.stages.findIndex(s => s.id === over.id)
        │
        ├─ arrayMove(formData.stages, oldIndex, newIndex)
        │  └─ Returns: reordered array
        │
        ├─ Map with new order:
        │  └─ .map((stage, idx) => ({ ...stage, order: idx + 1 }))
        │
        ├─ setFormData({ ...formData, stages: newStages })
        │
        └─ UI re-renders with new order

FORM SUBMISSION PATH
    ↓
    User clicks "Create Workflow"
        ↓
    handleSubmit()
        │
        ├─ validateForm()
        │  ├─ name.trim() not empty?
        │  ├─ documentType selected?
        │  └─ stages.length > 0?
        │
        ├─ Has errors? → setFormErrors, toast.error, return
        │  (User stays on page)
        │
        └─ No errors → onSubmit(formData)
           (Call parent callback)
               ↓
           Parent (CreateWorkflowClient) receives formData
               │
               ├─ setIsSubmitting(true)
               │
               ├─ try {
               │   ├─ API call: POST /api/workflows
               │   │  └─ Send formData as JSON body
               │   │
               │   ├─ Response received
               │   │  ├─ Success?
               │   │  │  ├─ toast.success()
               │   │  │  └─ router.push('/admin/workflows')
               │   │  │
               │   │  └─ Error?
               │   │     └─ throw error
               │   │
               │   └─ finally {
               │       setIsSubmitting(false)
               │   }
               │
               └─ catch (error) {
                   ├─ console.error()
                   └─ toast.error()
                   (User stays on page, can retry)
```

---

## Component Dependency Graph

```
Page (Server)
    ↓
CreateWorkflowClient (Client Parent)
    │
    ├─ State:
    │  └─ isSubmitting
    │
    ├─ Props to WorkflowBuilder:
    │  ├─ onSubmit (callback)
    │  ├─ isSubmitting (boolean)
    │  ├─ mode="create" (string)
    │  └─ initialData (optional)
    │
    └─→ WorkflowBuilder (Main Container)
        │
        ├─ State:
        │  ├─ formData
        │  ├─ showStageDialog
        │  ├─ editingStageId
        │  ├─ stageErrors
        │  └─ formErrors
        │
        ├─ Setup: dnd-kit sensors
        │
        ├─→ WorkflowDetailsForm
        │   ├─ Receives: data, onChange, errors
        │   ├─ Renders: Input, Textarea, Select, Checkbox
        │   └─ Calls: handleFormChange on change
        │
        ├─→ Card (Stages Container)
        │   ├─→ DndContext
        │   │   └─→ SortableContext
        │   │       └─→ StageItem[] (children)
        │   │           ├─ Receives: stage, onEdit, onDelete
        │   │           ├─ Hook: useSortable
        │   │           └─ Calls: onEdit(), onDelete() on click
        │   │
        │   └─ Button "Add Stage"
        │      └─ Calls: handleAddStage()
        │
        └─→ Dialog
            │
            ├─ State: showStageDialog, editingStageId
            │
            └─→ StageForm
                ├─ Receives: stage (null or object), onSave, onCancel, errors
                ├─ Renders: Input, Textarea, Select, Checkbox, Button
                ├─ Buttons:
                │  ├─ Cancel → onCancel() (close dialog)
                │  └─ Add/Update → onSave(stage) (validate & save)
                └─ Managed state: formData (local copy of stage)

DATA FLOW SUMMARY:

Parent Props ───→ WorkflowBuilder State ───→ Children Props ───→ UI
                        ↑                            │
                        │ (callback)                 ↓
                        └─ User Events ← Child Handlers
```

---

## State Mutation Map

```
Which State Changes When?

formData.name
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ WorkflowDetailsForm (updates input value)
  └─ Input shows new text

formData.description
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ WorkflowDetailsForm (updates textarea value)
  └─ Textarea shows new text

formData.documentType
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ WorkflowDetailsForm (updates select value)
  └─ Select shows new selection

formData.stages (entire array)
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ Card container
  ├─ All StageItem children re-render
  ├─ New order badges visible
  ├─ New/updated/deleted stages shown
  └─ Arrow connectors recalculated

formData.isDefault
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ WorkflowDetailsForm (checkbox checked state)
  └─ Checkbox shows new state

showStageDialog
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ Dialog open state changes
  └─ Modal appears/disappears

editingStageId
  ↓ triggers re-render (indirect)
  ├─ WorkflowBuilder
  ├─ editingStage computed (find operation)
  ├─ Dialog title changes ("Add" vs "Edit")
  └─ StageForm receives different stage prop

stageErrors
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ Dialog re-renders
  └─ StageForm shows error messages inline

formErrors
  ↓ triggers re-render
  ├─ WorkflowBuilder
  ├─ WorkflowDetailsForm shows error messages
  └─ Validation error displayed below form
```

---

## Validation Flow Diagram

```
VALIDATION SYSTEM

                    ┌─ Stage-Level Validation
                    │  (When adding/editing stage)
                    │
                    ├─ Check: name not empty
                    ├─ Check: approverRole not empty
                    ├─ Check: requiredApprovals >= 1
                    │
                    └─ If error:
                       ├─ setStageErrors(errors)
                       ├─ Display errors in form
                       └─ Dialog stays open (retry)
                    │
                    ├─ If valid:
                       ├─ Clear errors
                       └─ Add/update stage

Form Validation
    ├─────────────────┤
                    │
                    ├─ Form-Level Validation
                    │  (Before submission)
                    │
                    ├─ Check: name not empty
                    ├─ Check: documentType selected
                    ├─ Check: stages.length > 0
                    │
                    └─ If error:
                       ├─ setFormErrors(errors)
                       ├─ Show toast warning
                       └─ Block submission
                    │
                    ├─ If valid:
                       └─ Call onSubmit(formData)
                           ├─ Parent API call
                           ├─ Success → redirect
                           └─ Error → show toast

Error Clearing
    ├─ On Stage Level:
    │  └─ After saving stage, clear stageErrors
    │
    ├─ On Form Level:
    │  ├─ When user changes field:
    │  │  └─ Delete error for that field
    │  │
    │  └─ This creates "live" validation feel
```

---

## Re-Render Trigger Map

```
STATE CHANGE → IMMEDIATE RE-RENDERS

formData → WorkflowBuilder + all children
showStageDialog → WorkflowBuilder + Dialog
editingStageId → WorkflowBuilder + Dialog (title changes)
stageErrors → WorkflowBuilder + Dialog + StageForm
formErrors → WorkflowBuilder + WorkflowDetailsForm

OPTIMIZATION OPPORTUNITIES:

❌ Currently: No memoization
   ├─ All children re-render on any parent state change
   └─ Fine for 5 stages, problematic for 50+

✅ Option 1: Memoize children
   ├─ React.memo(StageItem)
   ├─ React.memo(WorkflowDetailsForm)
   └─ React.memo(StageForm)

✅ Option 2: Memoize computed values
   ├─ useMemo for stageIds array
   └─ useMemo for editingStage lookup

✅ Option 3: useCallback for handlers
   ├─ Prevents child re-renders if props unchanged
   └─ Keep handlers stable across renders

✅ Option 4: Split state (separate contexts)
   ├─ Stages context (only updates when stages change)
   ├─ Dialog context (only updates when dialog state changes)
   ├─ Form context (only updates when form changes)
   └─ Each context updates independently

CURRENT COST:
- Adding a stage: 1 re-render of entire tree
- Reordering stages: 1 re-render of entire tree
- Changing form field: 1 re-render of entire tree
- Minimal since max 5 stages, but scales poorly
```

---

## Error Handling Flow

```
┌─────────────────────────────────────────────────────────┐
│              ERROR HANDLING ARCHITECTURE                 │
└─────────────────────────────────────────────────────────┘

ERROR TYPE 1: Invalid Stage
  ├─ Location: StageForm modal
  ├─ Handler: handleSaveStage → validateStage()
  ├─ Detection: Missing name, role, or approvals
  ├─ Response:
  │  ├─ setStageErrors(errorMap)
  │  ├─ Display inline in form
  │  ├─ Dialog stays open
  │  └─ User can fix and retry
  └─ User Experience: ✓ Good (can fix immediately)

ERROR TYPE 2: Invalid Workflow
  ├─ Location: Main form
  ├─ Handler: handleSubmit → validateForm()
  ├─ Detection: Missing name, docType, or stages
  ├─ Response:
  │  ├─ setFormErrors(errorMap)
  │  ├─ Show toast warning
  │  ├─ Block submission
  │  └─ User stays on page
  └─ User Experience: ✓ Good (clear feedback)

ERROR TYPE 3: API Error (Parent)
  ├─ Location: CreateWorkflowClient
  ├─ Handler: handleSubmit catch block
  ├─ Detection: Network failure, server error
  ├─ Response:
  │  ├─ console.error() for logging
  │  ├─ toast.error() for user notification
  │  ├─ Buttons re-enabled (finally block)
  │  └─ User stays on page to retry
  └─ User Experience: ✓ Good (can retry)

ERROR TYPE 4: Validation After DnD
  ├─ Location: handleDragEnd
  ├─ Detection: None - no validation here
  ├─ Response: Just reorder array and update order
  └─ Risk: ⚠️ Could leave invalid state (low risk, already validated)

ERROR RECOVERY STRATEGY:

1. Prevent invalid state (client-side validation)
2. Show clear error messages (inline + toast)
3. Keep form open to allow retry (modal)
4. Log errors for debugging (console)
5. Disable UI during submission (isSubmitting flag)
6. Re-enable UI on error (finally block)

MISSING FEATURES:

❌ No error boundary (React error boundary would help)
❌ No retry mechanism (just user manually retries)
❌ No error reporting (no error tracking service)
❌ No form recovery (progress lost on reload)
```

---

## Component Lifecycle

```
WorkflowBuilder Lifecycle

MOUNT
  ├─ useState initializes: formData, showStageDialog, editingStageId, errors
  ├─ useSensors runs: dnd-kit sensor setup
  └─ Initial render

RENDER TRIGGERS
  ├─ formData changes
  │  └─ User types in form or adds/edits/deletes/reorders stages
  ├─ showStageDialog changes
  │  └─ User opens or closes dialog
  ├─ editingStageId changes
  │  └─ User clicks edit or add
  └─ errors change
     └─ Validation runs or user fixes field

UPDATE CYCLE (Typical User Action)
  1. User interacts (click, drag, type, etc.)
  2. Event handler fires (handleAddStage, handleFormChange, etc.)
  3. setState called (setFormData, setShowStageDialog, etc.)
  4. Entire component re-renders
  5. Children re-render with new props
  6. React updates DOM where needed
  7. UI reflects new state
  8. (Next user action)

UNMOUNT
  └─ Component removed from DOM (navigation away)

RE-RENDER STRATEGY
  ├─ Every state change causes full tree re-render
  ├─ No optimization currently in place
  ├─ Acceptable for max 5 stages
  └─ Would need memo/useMemo for larger workflows

MEMORY USAGE
  ├─ formData: Object with nested array
  ├─ stages array: Up to 5 items
  └─ Each stage: ~10 properties
  Result: Very minimal memory footprint
```

---

## Data Mutation Timeline

```
TIMELINE: User Creates 3-Stage Workflow

T=0.0s   User lands on /admin/workflows/create
         ├─ Server renders page component
         ├─ Auth check passes
         ├─ ClientComponents hydrate
         └─ WorkflowBuilder mounts
            ├─ formData = { name: '', stages: [], ... }
            ├─ showStageDialog = false
            └─ editingStageId = null

T=2.3s   User types "Standard Approval"
         ├─ handleFormChange('name', 'Standard Approval')
         ├─ setFormData({ ...formData, name: 'Standard Approval' })
         └─ Component re-renders
            └─ formData.name = "Standard Approval"

T=4.1s   User selects "REQUISITION"
         ├─ handleFormChange('documentType', 'REQUISITION')
         ├─ setFormData({ ...formData, documentType: 'REQUISITION' })
         └─ Component re-renders
            └─ formData.documentType = "REQUISITION"

T=5.5s   User clicks "Add Stage"
         ├─ handleAddStage()
         ├─ Check: formData.stages.length >= 5? NO
         ├─ setEditingStageId(null)
         ├─ setShowStageDialog(true)
         └─ Component re-renders
            ├─ showStageDialog = true
            ├─ Dialog visible
            └─ editingStage = null

T=7.2s   User fills stage 1 in modal
         ├─ StageForm local state updates (not parent)
         ├─ No parent re-render yet
         └─ Dialog is modal (isolated)

T=9.3s   User clicks "Add Stage" button in modal
         ├─ handleSaveStage(stageData)
         ├─ validateStage() runs
         ├─ No errors
         ├─ editingStageId === null? YES
         ├─ Generate ID: "stage-1733328400000"
         ├─ Set order: 1
         ├─ setFormData({
         │    ...formData,
         │    stages: [stageData]
         │  })
         ├─ setShowStageDialog(false)
         └─ Component re-renders
            ├─ formData.stages = [stage-1]
            ├─ showStageDialog = false
            └─ Dialog closes, stage visible

T=11.0s  User adds stage 2
         ├─ (repeat same flow)
         ├─ Stage 2 order = 2
         └─ formData.stages = [stage-1, stage-2]

T=12.8s  User adds stage 3
         ├─ (repeat same flow)
         ├─ Stage 3 order = 3
         └─ formData.stages = [stage-1, stage-2, stage-3]

T=14.2s  User drags stage-3 to position 1
         ├─ handleDragEnd({ active: {id: 'stage-3'}, over: {id: 'stage-1'} })
         ├─ oldIndex = 2
         ├─ newIndex = 0
         ├─ arrayMove([s1, s2, s3], 2, 0) = [s3, s1, s2]
         ├─ Recalculate order: s3→1, s1→2, s2→3
         ├─ setFormData with new order
         └─ Component re-renders
            └─ formData.stages = [stage-3, stage-1, stage-2]
               with corrected order values

T=15.9s  User clicks "Create Workflow"
         ├─ handleSubmit()
         ├─ validateForm() runs
         ├─ All checks pass
         ├─ onSubmit(formData) called
         └─ Parent receives:
            {
              name: "Standard Approval",
              description: "",
              documentType: "REQUISITION",
              stages: [
                {id: "stage-1733328400000", order: 2, name: "..."},
                {id: "stage-1733328401000", order: 3, name: "..."},
                {id: "stage-1733328402000", order: 1, name: "..."}
              ],
              isDefault: false
            }

T=16.0s  Parent handles submission
         ├─ setIsSubmitting(true)
         ├─ API POST /api/workflows
         └─ Awaiting response...

T=17.1s  API returns success
         ├─ toast.success('Workflow created')
         ├─ router.push('/admin/workflows')
         └─ Navigation away
            ├─ WorkflowBuilder unmounts
            └─ Memory cleaned up
```

---

## Summary: The Big Picture

```
WORKFLOW BUILDER = Multi-Step Form + Drag-Drop List

The Builder uses:
  ├─ React Hooks for state management
  ├─ dnd-kit for drag-drop functionality
  ├─ Shadcn UI for components
  ├─ Client-side validation (two levels)
  ├─ Modal dialogs for sub-forms
  └─ Toast notifications for feedback

Core Operations:
  1. Add Stage → Modal opens → Save stage → Added to list
  2. Edit Stage → Modal opens with data → Save stage → Updated in list
  3. Delete Stage → Removed from list → Order recalculated
  4. Reorder → Drag-drop → Array reordered → Order recalculated
  5. Submit → Validate → Call parent → Parent submits to API

State Flow:
  User Input → Event Handler → setState → Re-render → UI Updates

Key Principle:
  Parent manages form data, children are presentational
  All state flows down as props, events flow up as callbacks

Production Gap:
  Mock API → Need real server actions for create/update/delete
  In-memory → Need database persistence
  No optimization → Could add memo/useMemo for performance
```
