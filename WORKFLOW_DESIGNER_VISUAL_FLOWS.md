# Workflow Designer - Visual Interaction Flows

## Visual Map: Complete User Journey

### Journey 1: Creating a New Workflow

```
┌────────────────────────────────────────────────────────────────────────┐
│                    CREATE WORKFLOW FLOW                                 │
└────────────────────────────────────────────────────────────────────────┘

1. LANDING PAGE: /admin/workflows/create
   ┌──────────────────────────────────────┐
   │  Create Workflow                     │
   │  ────────────────────────────────────│
   │  [Workflow Details Section]          │
   │  • Name: [_____________]             │
   │  • Description: [_____________]      │
   │  • Document Type: [REQUISITION ▼]    │
   │  • Set as default: ☐                 │
   │                                      │
   │  [Approval Stages Section]           │
   │  ────────────────────────────────────│
   │  No stages added yet                 │
   │  [+ Add First Stage]                 │
   │                                      │
   │  [Action Buttons]                    │
   │  [Cancel] [Create Workflow]          │
   └──────────────────────────────────────┘

2. FILL WORKFLOW DETAILS
   User enters:
   • Name: "Standard Requisition Approval"
   • Description: "4-stage approval for requisitions"
   • Document Type: "REQUISITION"
   • isDefault: checked

   State updates: formData.name, formData.description, etc.
   Real-time validation clears error messages as user types

3. ADD FIRST STAGE
   User clicks [+ Add First Stage]
       ↓
   Dialog opens:
   ┌──────────────────────────────────────┐
   │  Add Stage                           │
   │  ────────────────────────────────────│
   │  Create a new approval stage...      │
   │                                      │
   │  Stage Name: [_____________]         │
   │  Description: [_____________]        │
   │  Approver Role: [DEPT_MGR ▼]         │
   │  Required Approvals: [1 ▼]           │
   │                                      │
   │  Permissions:                        │
   │  ☑ Approvers can reject documents    │
   │  ☑ Approvers can reassign to others  │
   │                                      │
   │  [Cancel] [Add Stage]                │
   └──────────────────────────────────────┘

4. FILL STAGE DETAILS
   User enters:
   • Stage Name: "Department Manager Review"
   • Description: "Manager reviews requisition"
   • Approver Role: "DEPARTMENT_MANAGER"
   • Required Approvals: "1"
   • canReject: true
   • canReassign: true

   State: stageErrors = {} (no errors)

5. SAVE FIRST STAGE
   User clicks [Add Stage]
       ↓
   validateStage() runs
       ├─ Check: name.trim() → "Department Manager Review" ✓
       ├─ Check: approverRole.trim() → "DEPARTMENT_MANAGER" ✓
       └─ Check: requiredApprovals >= 1 → 1 ✓
   Result: stageErrors = {}
       ↓
   editingStageId === null? YES
       ↓
   newStage = {
     id: "stage-1733328400000",
     order: 1,
     name: "Department Manager Review",
     description: "Manager reviews requisition",
     approverRole: "DEPARTMENT_MANAGER",
     requiredApprovals: 1,
     canReject: true,
     canReassign: true
   }
       ↓
   setFormData({
     ...formData,
     stages: [newStage]
   })
       ↓
   setShowStageDialog(false)
   Toast: "Stage added"

6. BACK TO MAIN VIEW
   Dialog closes, user sees:
   ┌──────────────────────────────────────┐
   │  Create Workflow                     │
   │  ────────────────────────────────────│
   │  [Workflow Details filled]           │
   │                                      │
   │  [Approval Stages Section]           │
   │  ┌────────────────────────────────┐  │
   │  │ 1 Department Manager Review    │  │
   │  │ Manager reviews requisition    │  │
   │  │ Role: Department Manager       │  │
   │  │ Approvals: 1                   │  │
   │  │ ✓ Can Reject ✓ Can Reassign    │  │
   │  │     [edit] [delete]            │  │
   │  └────────────────────────────────┘  │
   │            ↓                          │
   │  [+ Add Stage]                       │
   │                                      │
   │  [Cancel] [Create Workflow]          │
   └──────────────────────────────────────┘

7. ADD SECOND STAGE
   User clicks [+ Add Stage]
       ↓ (repeat steps 3-5)
       ↓
   New dialog opens, user adds:
   • "Finance Officer Review"
   • "CFO_FINANCE_OFFICER"
   • 1 approval
       ↓
   stageErrors = {}
   newStage = {
     id: "stage-1733328401000",
     order: 2,  ← Auto-calculated
     ...
   }
       ↓
   formData.stages = [stage-1, stage-2]
   Dialog closes

8. NOW VIEW WITH 2 STAGES
   ┌──────────────────────────────────────┐
   │  [Workflow Details]                  │
   │                                      │
   │  [Approval Stages]                   │
   │  ┌────────────────────────────────┐  │
   │  │≡ 1 Department Manager Review   │  │ ← Drag handle
   │  │   ...                          │  │
   │  │     [edit] [delete]            │  │
   │  └────────────────────────────────┘  │
   │            ↓ ← Arrow connector        │
   │  ┌────────────────────────────────┐  │
   │  │≡ 2 Finance Officer Review      │  │
   │  │   ...                          │  │
   │  │     [edit] [delete]            │  │
   │  └────────────────────────────────┘  │
   │                                      │
   │  [+ Add Stage]                       │
   │  [Cancel] [Create Workflow]          │
   └──────────────────────────────────────┘

9. SUBMIT WORKFLOW
   User clicks [Create Workflow]
       ↓
   handleSubmit() called
       ↓
   validateForm() runs:
       ├─ Check: name.trim() → "Standard..." ✓
       ├─ Check: documentType → "REQUISITION" ✓
       └─ Check: stages.length > 0 → 2 ✓
   Result: formErrors = {}
       ↓
   Validation passed? YES
       ↓
   onSubmit(formData) called (parent handler)
   Parent: setIsSubmitting(true)
       ↓
   API Call: POST /api/workflows
   Payload: {
     name: "Standard Requisition Approval",
     description: "4-stage approval for requisitions",
     documentType: "REQUISITION",
     stages: [stage-1, stage-2],
     isDefault: true
   }
       ↓
   Button text changes: "Creating..."
   All buttons disabled
       ↓
   API Response: 200 OK
       ↓
   Toast: "Workflow created successfully"
   router.push('/admin/workflows')
       ↓
   Redirect to workflow list
   ✓ Success!
```

---

## Visual Map: Drag-and-Drop Reordering

```
INITIAL STATE:
┌────────────────────────────────────────────────────┐
│  formData.stages = [stage-1, stage-2, stage-3]    │
│  formData.stages[0].order = 1                      │
│  formData.stages[1].order = 2                      │
│  formData.stages[2].order = 3                      │
└────────────────────────────────────────────────────┘

UI RENDERING:
┌──────────────────────────────────────┐
│ ┌────────────────────────────────┐   │
│ │ ≡ 1 Department Manager         │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
│            ↓                          │
│ ┌────────────────────────────────┐   │
│ │ ≡ 2 Finance Officer            │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │ CURRENT
│            ↓                          │ ORDER
│ ┌────────────────────────────────┐   │
│ │ ≡ 3 CFO                        │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
└──────────────────────────────────────┘

DRAG EVENT SEQUENCE:
1. User clicks grip handle on stage-3 (CFO)
   ├─ dnd-kit detects: active.id = "stage-3"
   ├─ isDragging = true
   ├─ Opacity changes to 0.5
   └─ Cursor changes to "grab"

2. User drags stage-3 between stage-1 and stage-2
   ├─ Mouse moves over stage-2
   ├─ collisionDetection algorithm activates
   ├─ over.id = "stage-2" (closest item)
   ├─ CSS transform applied for visual preview
   └─ User sees stage-3 in new position (preview)

3. User releases mouse on stage-2 position
   ├─ handleDragEnd({ active, over }) triggered
   ├─ active.id = "stage-3"
   ├─ over.id = "stage-2"
   ├─ active.id !== over.id? YES → Reorder needed
   │
   ├─ Find indices:
   │  ├─ oldIndex = 2 (stage-3 at position 2)
   │  └─ newIndex = 1 (stage-2 at position 1)
   │
   ├─ arrayMove([stage-1, stage-2, stage-3], 2, 1)
   │  └─ Result: [stage-1, stage-3, stage-2]
   │
   ├─ Renumber orders:
   │  ├─ stage-1: order = 1
   │  ├─ stage-3: order = 2 (was 3)
   │  └─ stage-2: order = 3 (was 2)
   │
   ├─ setFormData({
   │    ...formData,
   │    stages: [
   │      { ...stage-1, order: 1 },
   │      { ...stage-3, order: 2 },
   │      { ...stage-2, order: 3 }
   │    ]
   │  })
   │
   └─ isDragging = false, opacity = 1

FINAL STATE:
┌────────────────────────────────────────────────────┐
│  formData.stages = [stage-1, stage-3, stage-2]    │
│  formData.stages[0].order = 1                      │
│  formData.stages[1].order = 2                      │
│  formData.stages[2].order = 3                      │
└────────────────────────────────────────────────────┘

UI AFTER RE-RENDER:
┌──────────────────────────────────────┐
│ ┌────────────────────────────────┐   │
│ │ ≡ 1 Department Manager         │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
│            ↓                          │
│ ┌────────────────────────────────┐   │
│ │ ≡ 2 CFO                        │   │ NEW
│ │      [edit] [delete]           │   │ ORDER
│ └────────────────────────────────┘   │
│            ↓                          │
│ ┌────────────────────────────────┐   │
│ │ ≡ 3 Finance Officer            │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
└──────────────────────────────────────┘
```

---

## Visual Map: Delete Stage Flow

```
INITIAL STATE:
┌──────────────────────────────────────┐
│ formData.stages = [stage-1, stage-2, │
│                    stage-3]          │
│                                      │
│ stage-1: order=1, name="Dept Mgr"    │
│ stage-2: order=2, name="Finance"     │
│ stage-3: order=3, name="CFO"         │
└──────────────────────────────────────┘

USER ACTION: Click delete on stage-2 (Finance)

HANDLER CALLED: handleDeleteStage("stage-2")

STEP 1: Filter out stage-2
  formData.stages.filter(s => s.id !== "stage-2")
  Result: [stage-1, stage-3]

STEP 2: Renumber remaining stages
  [stage-1, stage-3].map((s, idx) => ({
    ...s,
    order: idx + 1
  }))

  ├─ idx=0: stage-1.order = 1 (no change)
  └─ idx=1: stage-3.order = 2 (was 3)

  Result: [
    { ...stage-1, order: 1 },
    { ...stage-3, order: 2 }
  ]

STEP 3: Update state
  setFormData({
    ...formData,
    stages: [stage-1, stage-3]  // stage-2 removed
  })

STEP 4: Show feedback
  toast.success('Stage removed')

FINAL STATE:
┌──────────────────────────────────────┐
│ formData.stages = [stage-1, stage-3] │
│                                      │
│ stage-1: order=1, name="Dept Mgr"    │
│ stage-3: order=2, name="CFO" ← order │
│          (was 3, now 2)              │
└──────────────────────────────────────┘

FINAL UI:
┌──────────────────────────────────────┐
│ ┌────────────────────────────────┐   │
│ │ ≡ 1 Department Manager         │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
│            ↓                          │
│ ┌────────────────────────────────┐   │
│ │ ≡ 2 CFO                        │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
│        (no more arrow below)          │
└──────────────────────────────────────┘
```

---

## Visual Map: Edit Stage Flow

```
INITIAL STATE:
Stage-2 data: {
  id: "stage-2",
  order: 2,
  name: "Finance Officer Review",
  description: "Budget validation",
  approverRole: "FINANCE_OFFICER",
  requiredApprovals: 1,
  canReject: true,
  canReassign: true
}

USER ACTION: Click edit icon on stage-2

┌──────────────────────────────────────┐
│ ≡ 2 Finance Officer Review           │
│ Budget validation                    │
│ Role: Finance Officer                │
│ Approvals: 1                         │
│ ✓ Can Reject ✓ Can Reassign          │
│      [✎ edit] [🗑 delete]             │
└──────────────────────────────────────┘
        ↑ clicks here

HANDLER CALLED: handleEditStage("stage-2")

STEP 1: Set edit state
  setEditingStageId("stage-2")
  setShowStageDialog(true)

STEP 2: Find stage data for pre-fill
  editingStage = formData.stages.find(
    s => s.id === "stage-2"
  )
  Result: {
    id: "stage-2",
    order: 2,
    name: "Finance Officer Review",
    description: "Budget validation",
    approverRole: "FINANCE_OFFICER",
    requiredApprovals: 1,
    canReject: true,
    canReassign: true
  }

STEP 3: Dialog opens with pre-filled data
  ┌──────────────────────────────────────┐
  │  Edit Stage                          │
  │  ────────────────────────────────────│
  │  Update the stage details            │
  │                                      │
  │  Stage Name: [Finance Officer Rev▌] │
  │  Description: [Budget validation▌]   │
  │  Approver Role: [FINANCE_OFFICER ▼]  │
  │  Required Approvals: [1 ▼]           │
  │                                      │
  │  Permissions:                        │
  │  ☑ Approvers can reject documents    │
  │  ☑ Approvers can reassign to others  │
  │                                      │
  │  [Cancel] [Update Stage]             │
  └──────────────────────────────────────┘

STEP 4: User modifies field
  Changes "Budget validation" to "Budget & compliance check"

STEP 5: User clicks "Update Stage"
  handleSaveStage(updatedStage) called
  where updatedStage = {
    id: "stage-2",
    order: 2,
    name: "Finance Officer Review",
    description: "Budget & compliance check",
    approverRole: "FINANCE_OFFICER",
    requiredApprovals: 1,
    canReject: true,
    canReassign: true
  }

STEP 6: Validate stage
  validateStage(updatedStage)
  ├─ name.trim()? "Finance Officer Review" ✓
  ├─ approverRole.trim()? "FINANCE_OFFICER" ✓
  └─ requiredApprovals >= 1? 1 ✓
  Result: stageErrors = {}

STEP 7: Check edit mode
  editingStageId !== null? YES → Edit mode

STEP 8: Update stages array
  updatedStages = formData.stages.map(s =>
    s.id === "stage-2" ? updatedStage : s
  )

  Result: [
    stage-1,
    {
      id: "stage-2",
      order: 2,
      name: "Finance Officer Review",
      description: "Budget & compliance check",  ← UPDATED
      approverRole: "FINANCE_OFFICER",
      requiredApprovals: 1,
      canReject: true,
      canReassign: true
    },
    stage-3
  ]

STEP 9: Update state & close
  setFormData({ ...formData, stages: updatedStages })
  setShowStageDialog(false)
  setStageErrors({})
  toast.success('Stage updated')

FINAL STATE:
All stages array with stage-2 updated

FINAL UI:
┌──────────────────────────────────────┐
│ ┌────────────────────────────────┐   │
│ │ ≡ 1 Department Manager...      │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
│            ↓                          │
│ ┌────────────────────────────────┐   │
│ │ ≡ 2 Finance Officer Review     │   │
│ │ Budget & compliance check      │   │ ← UPDATED
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
│            ↓                          │
│ ┌────────────────────────────────┐   │
│ │ ≡ 3 CFO                        │   │
│ │      [edit] [delete]           │   │
│ └────────────────────────────────┘   │
└──────────────────────────────────────┘
```

---

## State Machine Diagram

```
┌────────────────────────────────────────────────────────────────┐
│               WORKFLOW BUILDER STATE MACHINE                    │
└────────────────────────────────────────────────────────────────┘

                           START
                             │
                             ↓
                ┌─────────────────────────┐
                │  Initial State          │
                │  ─────────────────────  │
                │  formData: empty        │
                │  showDialog: false      │
                │  editingId: null        │
                │  errors: {}             │
                └─────────────────────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
            ↓                ↓                ↓
         EDIT         FILL DETAILS      ADD STAGE
      WORKFLOW            │                 │
            │              │                 │
            │              ↓                 ↓
            │    ┌──────────────────┐  Dialog Open
            │    │ User enters:     │  ├─ editingId: null
            │    │ • name           │  ├─ showDialog: true
            │    │ • description    │  └─ stageErrors: {}
            │    │ • documentType   │       │
            │    │ • isDefault      │       ↓
            │    └──────────────────┘  ┌──────────────────┐
            │              │           │ User fills       │
            │              │           │ stage form       │
            │              │           │ • name           │
            │              │           │ • role           │
            │              │           │ • approvals      │
            │              │           │ • permissions    │
            │              │           └──────────────────┘
            │              │                 │
            │              │        ┌────────┴─────────┐
            │              │        │                  │
            │              │        ↓                  ↓
            │              │      Invalid           VALID
            │              │        │                 │
            │              │        ↓                 ↓
            │              │  Show errors    Add/Update Stage
            │              │   (stay open)    Close dialog
            │              │        │          Toast ✓
            │              │        └─────┐    │
            │              │              │    ↓
            │              └──────┬───────┴──────────┐
            │                     │                  │
            │                     ↓                  ↓
            │            ┌──────────────────┐  Stages array
            │            │ Edit/Delete/     │  updated with
            │            │ Reorder stages   │  new/modified
            │            │                  │  stage
            │            └──────────────────┘  │
            │                     ↑             ↓
            │                     └─────────────┤
            │                                   │
            └───────────────────────────────────┤
                                                │
                                                ↓
                                  ┌──────────────────────┐
                                  │ Review workflow:     │
                                  │ • Details filled     │
                                  │ • All stages ready   │
                                  └──────────────────────┘
                                                │
                                                ↓
                                  User clicks "Create"
                                                │
                                        ┌───────┴─────────┐
                                        │                 │
                                        ↓                 ↓
                                      INVALID           VALID
                                        │                 │
                                        ↓                 ↓
                                  Show errors      Submit to
                                  (form errors)    parent
                                        │                 │
                                        │                 ↓
                                        │        API Call /
                                        │        Server Action
                                        │                 │
                                        │          ┌──────┴──────┐
                                        │          │             │
                                        │          ↓             ↓
                                        │        ERROR         SUCCESS
                                        │          │             │
                                        │          ↓             ↓
                                        │       Toast ✗      Toast ✓
                                        │       (show err)   Redirect
                                        │          │             │
                                        │          ↓             ↓
                                        │       USER BACK      SUCCESS
                                        │       CAN RETRY      STATE
                                        │          │
                                        └──────────┘
```

---

## Component Communication Flow

```
┌─────────────────────────────────────────────────────────────┐
│         COMPONENT COMMUNICATION DIAGRAM                       │
└─────────────────────────────────────────────────────────────┘

CreateWorkflowClient (Parent)
│
│ Props down: onSubmit, isSubmitting, mode
│ ↓↑ Events up: onSubmit(formData)
│
└─→ WorkflowBuilder (Main Container)
    │
    ├─ State: formData, showDialog, editingId, errors
    │
    ├─ Props down: data, onChange, errors
    │ ↓↑ Events up: onChange(key, value)
    ├─→ WorkflowDetailsForm (Presentational)
    │   │
    │   └─ UI: Input, Textarea, Select, Checkbox
    │
    ├─ Props down: data, onChange, errors
    │ ↓↑ Events up: onChange(key, value)
    ├─→ Card (Stages Container)
    │   │
    │   ├─→ DndContext (Drag context)
    │   │   │
    │   │   ├─→ SortableContext
    │   │   │   │
    │   │   │   └─→ StageItem[] (Presentational)
    │   │   │       │
    │   │   │       ├─ Props: stage, onEdit, onDelete
    │   │   │       └─ Events: onEdit(), onDelete()
    │   │   │           │ ↑
    │   │   │           │ └─ Calls parent handlers
    │   │   │
    │   │   └─ onDragEnd: handleDragEnd()
    │   │       └─ Updates formData.stages
    │
    └─ Props down: stage, onSave, onCancel, errors
      ↓↑ Events up: onSave(stage), onCancel()
      └─→ Dialog (Modal Container)
          │
          └─→ StageForm (Presentational)
              │
              └─ UI: Input, Textarea, Select, Checkbox
                 └─ Buttons: Cancel, Add/Update Stage
```

---

## Error Flow Visualization

```
┌────────────────────────────────────────────────────────┐
│              ERROR HANDLING FLOW                        │
└────────────────────────────────────────────────────────┘

1. STAGE-LEVEL VALIDATION
   ┌──────────────────────────────────┐
   │ User submits StageForm           │
   │ ├─ handleSaveStage(stage) called │
   │ ├─ validateStage(stage) runs     │
   │ │                                │
   │ │  Check:                        │
   │ │  1. stage.name.trim() not      │
   │ │     empty?                     │
   │ │  2. stage.approverRole.trim()  │
   │ │     not empty?                 │
   │ │  3. stage.requiredApprovals    │
   │ │     >= 1?                      │
   │ │                                │
   │ ├─ Result: errors object         │
   │ └─ Has errors?                   │
   │    ├─ YES:                       │
   │    │  ├─ setStageErrors(errors)  │
   │    │  ├─ return (exit early)     │
   │    │  └─ Dialog stays open       │
   │    │     Error shown on form     │
   │    │     User can fix & retry    │
   │    │                             │
   │    └─ NO:                        │
   │       └─ Continue to save        │
   └──────────────────────────────────┘

2. FORM-LEVEL VALIDATION
   ┌──────────────────────────────────┐
   │ User clicks "Create Workflow"    │
   │ ├─ handleSubmit() called         │
   │ ├─ validateForm() runs           │
   │ │                                │
   │ │  Check:                        │
   │ │  1. name not empty?            │
   │ │  2. documentType selected?     │
   │ │  3. stages.length > 0?         │
   │ │                                │
   │ ├─ Result: formErrors object     │
   │ └─ Has errors?                   │
   │    ├─ YES:                       │
   │    │  ├─ setFormErrors(errors)   │
   │    │  ├─ toast.error(...)        │
   │    │  └─ return (exit early)     │
   │    │     User stays on page      │
   │    │     Error message shown     │
   │    │                             │
   │    └─ NO:                        │
   │       └─ Call onSubmit()         │
   └──────────────────────────────────┘

3. ERROR CLEARING
   ┌──────────────────────────────────┐
   │ User changes a field             │
   │ ├─ handleFormChange(key, value)  │
   │ │  called                        │
   │ ├─ setFormData(...)              │
   │ ├─ if formErrors[key]?           │
   │ │  ├─ YES:                       │
   │ │  │  ├─ Delete error for key    │
   │ │  │  └─ Re-render (no error)    │
   │ │  │     Auto-cleared!           │
   │ │  └─ NO:                        │
   │ │     └─ Do nothing              │
   └──────────────────────────────────┘

4. API ERROR HANDLING (Parent level)
   ┌──────────────────────────────────┐
   │ onSubmit(formData) called        │
   │ ├─ setIsSubmitting(true)         │
   │ ├─ try {                         │
   │ │  ├─ API call POST /workflows   │
   │ │  ├─ Wait for response          │
   │ │  └─ Success:                   │
   │ │     ├─ toast.success()         │
   │ │     └─ router.push(...)        │
   │ ├─ } catch (error) {             │
   │ │  ├─ console.error()            │
   │ │  └─ toast.error(...)           │
   │ │     User stays on page         │
   │ │     Can try again              │
   │ ├─ } finally {                   │
   │ │  └─ setIsSubmitting(false)     │
   │ └─ }                             │
   └──────────────────────────────────┘
```

---

## Performance: Re-render Optimization

```
┌────────────────────────────────────────────────────┐
│     RE-RENDER TRIGGER MAP                          │
└────────────────────────────────────────────────────┘

STATE CHANGE → COMPONENT AFFECTED

formData.name changed
  ├─ WorkflowBuilder re-renders
  ├─ WorkflowDetailsForm re-renders
  └─ (Does NOT affect: StageItem, Dialog)

formData.stages changed (add/edit/delete/reorder)
  ├─ WorkflowBuilder re-renders
  ├─ Card container re-renders
  ├─ All StageItem children re-render
  └─ (Does NOT affect: WorkflowDetailsForm)

formData.documentType changed
  ├─ WorkflowBuilder re-renders
  ├─ WorkflowDetailsForm re-renders
  └─ (Does NOT affect: Stages)

showStageDialog changed to true
  ├─ WorkflowBuilder re-renders
  ├─ Dialog re-renders
  ├─ StageForm mounts
  └─ (Does NOT affect: Other components)

editingStageId changed
  ├─ WorkflowBuilder re-renders
  ├─ editingStage computed (find operation)
  └─ (Does NOT affect: UI until dialog opens)

formErrors changed
  ├─ WorkflowBuilder re-renders
  ├─ Error messages appear/disappear
  └─ (Does NOT affect: Stages)

stageErrors changed
  ├─ WorkflowBuilder re-renders
  ├─ Dialog re-renders
  └─ Error messages in StageForm update

┌────────────────────────────────────────────────────┐
│     OPTIMIZATION OPPORTUNITIES                      │
└────────────────────────────────────────────────────┘

Current: No memoization
├─ All children re-render on parent update
├─ Many unnecessary renders
└─ Fine for 5 stages, problematic for 50+

Option 1: useMemo for stages map
  const stageIds = useMemo(
    () => formData.stages.map((s) => s.id),
    [formData.stages]
  )

Option 2: React.memo for StageItem
  const MemoizedStageItem = React.memo(StageItem)

Option 3: useCallback for handlers
  const handleEditStage = useCallback((id) => {
    setEditingStageId(id)
    setShowStageDialog(true)
  }, [])

Option 4: Split state (separate contexts)
  - Stages context
  - Dialog context
  - Form errors context
  Each can update independently
```

---

## Testing Scenarios

```
┌────────────────────────────────────────────────────┐
│         TEST CASE SCENARIOS                        │
└────────────────────────────────────────────────────┘

1. ADD STAGE HAPPY PATH
   ✓ Click "Add Stage"
   ✓ Dialog opens
   ✓ Fill all required fields
   ✓ Click "Add Stage" button
   ✓ Dialog closes
   ✓ Stage appears in list
   ✓ Order = 1 (first stage)

2. ADD STAGE - VALIDATION
   ✓ Fill only stage name
   ✓ Click "Add Stage"
   ✓ Error: "Approver role is required" shown
   ✓ Dialog stays open
   ✓ Select approver role
   ✓ Error clears
   ✓ Click "Add Stage"
   ✓ Success

3. ADD STAGE - MAX LIMIT
   ✓ Add 5 stages successfully
   ✓ Click "Add Stage" (6th attempt)
   ✓ Toast: "Maximum 5 stages allowed"
   ✓ Dialog does NOT open
   ✓ Button disabled? (optional)

4. REORDER - DRAG DROP
   ✓ Stage 3 in position [1,2,3]
   ✓ Drag stage 3 to position 1
   ✓ Drop on stage 1
   ✓ Order becomes [3,1,2]
   ✓ stage-1.order = 2
   ✓ stage-3.order = 1

5. DELETE STAGE
   ✓ 3 stages visible
   ✓ Click delete on stage 2
   ✓ Stage removed
   ✓ Remaining 2 stages renumbered
   ✓ stage-3.order = 2 (was 3)

6. EDIT STAGE
   ✓ Stage-2 has name="Finance"
   ✓ Click edit
   ✓ Dialog opens with name="Finance"
   ✓ Change name to "Finance & Compliance"
   ✓ Click "Update Stage"
   ✓ Stage-2.name = "Finance & Compliance"
   ✓ List refreshes

7. SUBMIT WORKFLOW
   ✓ 2 stages added
   ✓ Name filled
   ✓ Document type selected
   ✓ Click "Create Workflow"
   ✓ Button shows "Creating..."
   ✓ API call sent
   ✓ Success: redirected to list
   ✓ Error: toast shown, stay on page

8. FORM VALIDATION
   ✓ No name, click submit
   ✓ Toast: "Please fix errors"
   ✓ No stages, click submit
   ✓ Toast: "At least one stage required"
   ✓ Empty form can't submit
```
