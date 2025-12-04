# Workflow Builder - Quick Summary

## What is the Workflow Builder?

A **React-based visual workflow designer** that lets administrators create multi-stage approval workflows for documents (requisitions, purchase orders, payment vouchers, etc.). Users drag-and-drop to reorder stages, and use modal dialogs to configure each stage's approval requirements.

---

## How It Works (30-Second Version)

```
User visits /admin/workflows/create
    ↓
Fills workflow name, description, document type
    ↓
Clicks "Add Stage" to open stage editor modal
    ↓
Fills stage details: name, role, approvals, permissions
    ↓
Saves stage → appears in list
    ↓
Can add up to 5 stages (drag to reorder)
    ↓
Clicks "Create Workflow" to submit
    ↓
Validation runs, API called, redirects to list
```

---

## Core Components

| Component | Purpose | Lines |
|-----------|---------|-------|
| **WorkflowBuilder** | Main orchestrator, manages state & handlers | 300 |
| **WorkflowDetailsForm** | Top-level form (name, description, doc type) | 104 |
| **StageForm** | Modal dialog for stage configuration | 187 |
| **StageItem** | Draggable stage card with dnd-kit | 123 |
| **WorkflowsClient** | List view of all workflows | 273 |

---

## Key Data Model

```typescript
WorkflowFormData {
  name: "Standard Requisition Approval"
  description: "4-stage approval process"
  documentType: "REQUISITION"
  stages: [
    {
      id: "stage-1733328400000"
      order: 1
      name: "Department Manager Review"
      approverRole: "DEPARTMENT_MANAGER"
      requiredApprovals: 1
      canReject: true
      canReassign: true
    },
    // ... more stages
  ]
  isDefault: true
}
```

---

## State Management

```typescript
const [formData, setFormData] = useState<WorkflowFormData>()      // Main data
const [showStageDialog, setShowStageDialog] = useState(false)     // Dialog open?
const [editingStageId, setEditingStageId] = useState(null)        // Edit mode?
const [stageErrors, setStageErrors] = useState({})                // Form errors
const [formErrors, setFormErrors] = useState({})                  // Stage errors
```

All state is local to WorkflowBuilder; parent component passes `onSubmit` callback.

---

## Drag-and-Drop System

**Library**: dnd-kit (`@dnd-kit/core`, `@dnd-kit/sortable`)

**How it works**:
1. User clicks grip handle on stage
2. dnd-kit detects drag start
3. Card fades to 0.5 opacity
4. User drags to new position
5. dnd-kit calculates collision (closest stage)
6. User releases mouse
7. `handleDragEnd()` fires
8. `arrayMove()` reorders array
9. Order property recalculated (1, 2, 3...)
10. UI re-renders

**Key Code**:
```typescript
const handleDragEnd = (event: DragEndEvent) => {
  if (over && active.id !== over.id) {
    const oldIndex = formData.stages.findIndex((s) => s.id === active.id)
    const newIndex = formData.stages.findIndex((s) => s.id === over.id)
    const newStages = arrayMove(formData.stages, oldIndex, newIndex).map(
      (stage, idx) => ({ ...stage, order: idx + 1 })
    )
    setFormData({ ...formData, stages: newStages })
  }
}
```

---

## Event Handlers Summary

| Handler | Triggered | Action |
|---------|-----------|--------|
| `handleAddStage()` | "Add Stage" button | Open dialog in create mode |
| `handleEditStage(id)` | Edit icon | Open dialog in edit mode |
| `handleDeleteStage(id)` | Delete icon | Remove stage, renumber order |
| `handleDragEnd(event)` | Drag release | Reorder stages via dnd-kit |
| `handleSaveStage(stage)` | Form submit | Add/update stage in array |
| `handleFormChange(key, value)` | Input change | Update formData, clear error |
| `handleSubmit()` | Submit button | Validate & call parent's onSubmit |

---

## Validation

**Two-level validation**:

1. **Stage-level** (when saving a stage):
   - Stage name not empty ✓
   - Approver role selected ✓
   - Required approvals ≥ 1 ✓

2. **Form-level** (before submission):
   - Workflow name not empty ✓
   - Document type selected ✓
   - At least 1 stage added ✓

**Errors are displayed inline**; if user corrects a field, error clears immediately.

---

## Create vs Edit Flow

### Create Mode
```
/admin/workflows/create
  → CreateWorkflowClient
    → WorkflowBuilder with initialData={undefined}
      → Empty form
      → User fills & adds stages
      → Submits → API creates new workflow
      → Redirects to /admin/workflows
```

### Edit Mode
```
/admin/workflows/{id}/edit
  → EditWorkflowClient (fetches workflow data)
    → WorkflowBuilder with initialData={existingWorkflow}
      → Pre-populated form
      → User modifies stages
      → Submits → API updates workflow
      → Redirects to /admin/workflows
```

---

## Limitations (Current MVP)

| Limitation | Impact | Fix |
|-----------|--------|-----|
| No backend | Data lost on refresh | Implement server actions |
| Stage ID from Date.now() | Could collide in < 1ms | Use UUID instead |
| No memoization | Unnecessary re-renders | Add React.memo, useMemo |
| Max 5 stages hardcoded | Artificial limit | Make configurable |
| No undo/redo | Changes permanent | Implement history stack |
| No draft save | Progress lost | Auto-save to localStorage |
| Limited stage config | Advanced features hidden | Expand UI (SLA, transitions) |
| No versioning UI | Can't select versions | Add version selector |

---

## Production Checklist

- [ ] Implement `createWorkflow` server action
- [ ] Implement `updateWorkflow` server action
- [ ] Implement `deleteWorkflow` server action
- [ ] Switch to UUID for stage IDs
- [ ] Add `React.memo` to StageItem
- [ ] Add `useMemo` for stages array map
- [ ] Add `useCallback` for handlers
- [ ] Implement draft auto-save to localStorage
- [ ] Add loading skeleton for edit mode
- [ ] Add optimistic UI updates
- [ ] Add error recovery (rollback on API fail)
- [ ] Add workflow versioning UI
- [ ] Add SLA/escalation configuration
- [ ] Add stage transition customization
- [ ] Add import/export functionality
- [ ] Add workflow templates & cloning

---

## Integration Points

### Parent Component (CreateWorkflowClient)
```typescript
<WorkflowBuilder
  onSubmit={handleSubmit}          // Required
  isSubmitting={isSubmitting}      // Required
  mode="create"                    // Required
  initialData={undefined}          // Optional
/>
```

### Child Components
- **WorkflowDetailsForm**: Receives `data`, `onChange`, `errors`
- **StageForm**: Receives `stage`, `onSave`, `onCancel`, `errors`
- **StageItem**: Receives `stage`, `onEdit`, `onDelete`

---

## File Locations

```
src/app/(private)/admin/workflows/
├── page.tsx                          (workflows list page)
├── create/
│   ├── page.tsx                      (create page)
│   └── _components/create-workflow-client.tsx
├── [id]/edit/
│   ├── page.tsx                      (edit page)
│   └── _components/edit-workflow-client.tsx
└── _components/
    ├── workflow-builder.tsx          (★ MAIN)
    ├── workflow-details-form.tsx
    ├── stage-form.tsx
    ├── stage-item.tsx
    └── workflows-client.tsx
```

---

## Dependencies

```json
{
  "@dnd-kit/core": "^6.3.1",           // Drag-drop core
  "@dnd-kit/sortable": "^10.0.0",      // Sortable helper
  "@dnd-kit/utilities": "^3.2.2",      // Utilities
  "lucide-react": "^0.522.0",          // Icons
  "sonner": "^2.0.6",                  // Toast notifications
  "react-hook-form": "^7.58.1",        // Form management
  "zod": "^4.1.13"                     // Validation
}
```

---

## Type System

All types are TypeScript interfaces, defined in:
- `create-workflow-client.tsx`: `WorkflowFormData`, `WorkflowStage`
- `src/types/custom-workflow.ts`: Full comprehensive types
- `src/types/workflow.ts`: Base workflow types

---

## Testing Strategy

```typescript
// Unit tests for handlers
test('handleAddStage should open dialog', () => {
  const { getByRole } = render(<WorkflowBuilder ... />)
  fireEvent.click(getByRole('button', { name: /add stage/i }))
  expect(getByRole('dialog')).toBeInTheDocument()
})

// Integration tests for workflows
test('adding a stage should appear in list', () => {
  // Add stage with form
  // Verify stage renders in list
  // Verify order = 1
})

// Drag-drop tests with dnd-kit
test('dragging stage 3 to position 1 should reorder', () => {
  // Use dnd-kit testing utilities
  // Simulate drag event
  // Verify order property updated
})
```

---

## FAQ

**Q: Why is the stage ID using `Date.now()`?**
A: It's a temporary solution. Use UUID for production.

**Q: Can I add more than 5 stages?**
A: Not currently. The limit is hardcoded in `handleAddStage()`. Remove the check to allow more.

**Q: What happens if I refresh the page?**
A: All changes are lost. No backend persistence yet.

**Q: How is order calculated after reordering?**
A: After drag-and-drop or delete, the array is mapped with `idx + 1`.

**Q: Can two users edit the same workflow?**
A: Not concurrently. Last-write-wins. Add optimistic locking in production.

**Q: Is validation happening twice?**
A: Yes—once in stage form, once before submission. This is intentional (belt-and-suspenders).

**Q: Why use a modal for stage editing instead of inline?**
A: Modal provides focus and prevents accidental changes to the workflow.

**Q: Can I save drafts?**
A: Not currently. Consider localStorage auto-save for production.

**Q: What's the order property used for?**
A: Visual display (the numbered badge on each stage). Must stay in sync with array index.

---

## Next Steps

1. **Immediate**: Replace mock API with actual server actions
2. **Short-term**: Add database persistence layer
3. **Medium-term**: Implement undo/redo and draft saving
4. **Long-term**: Add advanced features (SLA, escalations, versioning)

---

## Documentation Files

This audit includes 4 comprehensive documents:

1. **WORKFLOW_BUILDER_DEEP_DIVE.md** - Architecture & detailed component analysis
2. **WORKFLOW_DESIGNER_VISUAL_FLOWS.md** - Visual flowcharts & interaction diagrams
3. **WORKFLOW_BUILDER_CODE_REFERENCE.md** - Code snippets & implementation details
4. **WORKFLOW_BUILDER_SUMMARY.md** - This file (quick reference)

---

## Key Takeaways

✅ **Well-architected**: Clear separation of concerns, good component structure
✅ **Type-safe**: Comprehensive TypeScript types
✅ **User-friendly**: Drag-and-drop with real-time validation
✅ **Extensible**: Data model supports advanced features
✅ **Production-ready UI**: Polished with Shadcn components

⚠️ **MVP Stage**: No backend integration, mock data only
⚠️ **Performance**: No memoization or optimization
⚠️ **Limited features**: Advanced stage config not exposed

🎯 **Ready for**: Adding backend API, database integration, advanced features
