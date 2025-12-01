# Approval Workflow Guide

## How Approvals Work

### Basic Flow

```
Document Created
    ↓ Routed to First Approver
First Approver Reviews
    ↓ Can: Approve, Reject, Reassign
If Approved:
    ↓ Routed to Next Approver
    ↓ Continues through stages
Final Approval:
    ↓ Document Complete
If Rejected:
    ↓ Returned to Requester
    ↓ Reason provided
```

## Workflow Types Explained

### Requisition (3 stages)

**Purpose**: Request approval to buy something

**Stage 1: Department Manager**
- Reviews request
- Can approve if within budget
- Can reject if not needed
- Can reassign to someone else

**Stage 2: Finance Officer**
- Reviews department approval
- Checks budget availability
- Verifies cost
- Can approve or reject

**Stage 3: CFO**
- Final approval authority
- Reviews overall spend
- Releases purchase order
- Can approve or reject

**Time Estimate**: 3-5 business days

---

### Budget (3 stages)

**Purpose**: Allocate budget to departments

**Stage 1: Department Manager**
- Reviews allocation
- Confirms departmental needs

**Stage 2: Finance Officer**
- Checks total budget
- Verifies no overlap

**Stage 3: CFO**
- Final authorization
- Releases funds

**Time Estimate**: 2-3 business days

---

### Purchase Order (3 stages)

**Purpose**: Order products from vendor

**Special Features**: Shows vendor details, items, costs

**Stage 1: Department Manager**
- Confirms vendor and items
- Verifies specifications

**Stage 2: Finance Officer**
- Reviews pricing
- Checks GL codes

**Stage 3: CFO**
- Final approval
- Authorizes payment

**Time Estimate**: 3-5 business days

---

### Payment Voucher (3 stages)

**Purpose**: Approve payment to vendor

**Special Features**: Shows invoice, payment method, GL codes

**Stage 1: Department Manager**
- Verifies invoice matches PO
- Confirms goods received

**Stage 2: Finance Officer**
- Verifies payment details
- Checks GL coding

**Stage 3: CFO**
- Authorizes payment
- Releases check/transfer

**Time Estimate**: 1-2 business days

---

### GRN (2 stages) - UNIQUE

**Purpose**: Confirm goods received match order

**IMPORTANT**: Only 2 stages, not 3

**Stage 1: Warehouse Clerk**
- Receives goods
- Matches items to PO
- Documents any damage
- Notes variances
- Fills confirmation form
- Provides signature/name

**Stage 2: Department Manager**
- Reviews warehouse confirmation
- Verifies all items received
- Approves or rejects
- Provides signature

**Time Estimate**: Same day

---

## Approval Actions

### Approve

**What it means**: You approve and move to next stage

**What you provide**:
- Digital signature (draw on canvas)
- Optional remarks/comments

**Effect**:
- Document moves to next stage
- Next approver notified (Phase 12: email)
- Signature recorded in history

**Example Remarks**:
- "Approved for procurement"
- "Budget confirmed, proceed"
- "GL codes verified"

### Reject

**What it means**: You reject and send back to requester

**What you provide**:
- Digital signature (required)
- Rejection reason (required)
- Optional comments

**Effect**:
- Document returns to requester
- Stage resets to 0
- Requester sees rejection reason
- Can resubmit with changes

**Example Reasons**:
- "Cost exceeds budget"
- "Specs not compatible"
- "Need clarification on vendor"
- "GL code incorrect"

### Reassign

**What it means**: Pass to different approver at same stage

**What you provide**:
- New approver selection
- Optional reason

**Effect**:
- Document stays at same stage
- Goes to new approver
- Reason recorded in history

**Example Reasons**:
- "Manager on leave"
- "Better vendor knowledge"
- "Department authority"

---

## Multi-Stage Example

**Requisition for Office Supplies ($1,500)**

```
Day 1: Requester submits
↓ Goes to: Manager (Stage 1/3)

Manager reviews - "These supplies are needed"
Draws signature
Clicks APPROVE
↓ Goes to: Finance Officer (Stage 2/3)

Finance checks: "Budget available, approve"
Draws signature
Clicks APPROVE
↓ Goes to: CFO (Stage 3/3)

CFO final check: "All looks good"
Draws signature
Clicks APPROVE
↓
Status: APPROVED ✓
Purchase order released
Supplier notified
```

---

## Bulk Operations Guide

### When to Use Bulk Approve

**Use when**:
- Multiple similar items need approval
- All from same requester
- All at same stage
- All meeting same criteria

**How**:
1. Select items with checkboxes
2. Click "Approve All"
3. Dialog shows count
4. Add optional remarks
5. Submit

**Effect**: All selected items approved in one action

### When to Use Bulk Reject

**Use when**:
- Multiple items have same issue
- All missing same information
- All failing same criteria

**How**:
1. Select items
2. Click "Reject All"
3. Enter rejection reason (required)
4. Submit

**Effect**: All selected items rejected with reason

### When to Use Bulk Reassign

**Use when**:
- Multiple items need different approver
- You're unavailable
- Need specialized review

**How**:
1. Select items
2. Click "Reassign All"
3. Select new approver
4. Add optional reason
5. Submit

**Effect**: All items reassigned to new approver

---

## Signature Capture

### How to Sign

1. Click on signature canvas area
2. Draw signature with mouse or touch
3. Use natural handwriting style
4. Complete signing
5. Canvas shows your signature

### What Gets Stored

- Base64-encoded signature image
- Associated with approval action
- Timestamp recorded
- Approver name recorded

### Signature is Legal?

Phase 11: Simulated (proof of concept)
Phase 12: Will add proper digital certificate

---

## Common Questions

**Q: Can I change my approval?**
A: Not directly. You would need to contact admin for Phase 12.

**Q: What if I approve by mistake?**
A: Contact the next approver to reject it back to you.

**Q: How long do approvals take?**
A: 1-5 business days depending on workflow and stages.

**Q: Can I see approval history?**
A: Yes, click on any task to see all approvals/rejections.

**Q: What if I'm on vacation?**
A: Use Reassign to give to colleague, or ask admin.

---

## Tips for Approvers

1. **Act Promptly**: Don't leave tasks pending
2. **Add Remarks**: Help document your decision
3. **Check Details**: Review all fields before approval
4. **Use Bulk**: Approve multiple similar items together
5. **Watch Analytics**: Monitor bottlenecks
6. **Document**: Your signature proves approval

---

## Phase 12 Enhancements

In Phase 12:
- Email notifications when task assigned
- Digital certificates for signatures
- Audit log of all approvals
- Permission enforcement (role-based)
- Escalation if overdue
- Email notifications on rejection

---

**Need help?** Check the demo guide or contact support.
