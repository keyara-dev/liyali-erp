# Workflows - Overview

**Chapter 7: Workflow System**

---

## What is the Workflow System?

The Liyali Gateway implements a flexible multi-stage approval workflow system that processes financial documents through configurable approval chains.

### Core Capabilities

- **5 Document Types**: Requisition, Budget, Purchase Order, Payment Voucher, GRN
- **Multi-stage Approvals**: 2-3 approval stages per workflow
- **Custom Workflows**: Admin-created workflows with drag-and-drop builder
- **Digital Signatures**: Approval capture with base64 signature storage
- **Bulk Operations**: Approve/reject/reassign multiple tasks simultaneously
- **Complete Audit Trail**: All workflow actions logged with timestamps

---

## Document Types

### 1. Requisition (REQUISITION)
- **Stages**: 2 (Department Manager → Finance Officer)
- **Purpose**: Purchase request approval
- **Fields**: Items, quantities, unit prices, total cost

### 2. Budget (BUDGET)
- **Stages**: 3 (Department Head → Finance Manager → CFO)
- **Purpose**: Budget allocation approval
- **Fields**: Line items, amounts, budget codes

### 3. Purchase Order (PURCHASE_ORDER)
- **Stages**: 2 (Procurement Officer → Approving Officer)
- **Purpose**: Vendor purchase authorization
- **Fields**: Vendor info, items, GL codes, delivery details

### 4. Payment Voucher (PAYMENT_VOUCHER)
- **Stages**: 3 (Approving Officer → Finance Officer → Bank Officer)
- **Purpose**: Payment processing authorization
- **Fields**: Payment method, bank details, invoice reference

### 5. GRN (Goods Received Note)
- **Stages**: 2 (Warehouse Officer → Quality Officer)
- **Purpose**: Goods receipt verification
- **Fields**: Received items, quantities, quality status

---

## Workflow Architecture

### Data Flow

```
Document Created (DRAFT)
    ↓
Submitted for Approval (SUBMITTED)
    ↓
Routed to First Stage Approver (IN_REVIEW)
    ↓
Stage 1 Approval/Rejection
    ↓
Routed to Next Stage OR Completed
    ↓
Final Status: APPROVED or REJECTED
```

### Database Tables

```sql
-- Core workflow tables
workflows              -- Workflow definitions
approval_tasks         -- Individual approval tasks
approval_history       -- Approval action history
documents              -- Financial documents
```

---

## Custom Workflows

Administrators can create custom workflows via the workflow builder:

1. Define workflow name and document type
2. Add approval stages (up to 5)
3. Configure stage approver roles
4. Set permissions (can reject, can reassign)
5. Set as default workflow (optional)

### Workflow JSON Structure

```json
{
  "id": "uuid",
  "name": "Standard Requisition Approval",
  "document_type": "REQUISITION",
  "stages": [
    {
      "stage": 1,
      "name": "Manager Review",
      "approver_role": "DEPARTMENT_MANAGER",
      "required_approvals": 1,
      "can_reject": true,
      "can_reassign": true
    }
  ],
  "is_active": true
}
```

---

## Approval Process

### 1. Task Assignment
When a document is submitted, approval tasks are created for each stage based on the workflow definition.

### 2. Approver Actions
Approvers can:
- **Approve**: Sign and move to next stage
- **Reject**: Provide reason and reject document
- **Reassign**: Transfer task to another user
- **Comment**: Add notes without changing status

### 3. Bulk Operations
Process multiple tasks simultaneously:
- Bulk approve with single signature
- Bulk reject with common reason
- Bulk reassign to specific user

### 4. Status Transitions

```
DRAFT → SUBMITTED → IN_REVIEW → APPROVED
                              ↘ REJECTED
```

---

## API Integration

### Key Endpoints

- `GET /api/workflows` - List all workflows
- `POST /api/workflows` - Create custom workflow
- `GET /api/approvals/tasks` - Get assigned tasks
- `POST /api/approvals/tasks/:id/approve` - Approve task
- `POST /api/approvals/tasks/:id/reject` - Reject task
- `POST /api/approvals/bulk/approve` - Bulk approve

See [Chapter 4: API Reference](../04-api-reference/) for complete endpoint documentation.

---

## Next Steps

- [Workflow Implementation](./implementation.md) - Technical implementation details
- [Approval Handlers](./approval-handlers.md) - Handler functions and business logic
- [Testing Workflows](../06-testing/workflows.md) - Testing strategies for workflows

---

**Last Updated**: December 26, 2025
