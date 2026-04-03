# Audit Trail - Visual Guide

## 🎯 What We Built

A comprehensive audit logging system that provides **complete transparency** for all document changes.

## 📊 Audit Trail Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    USER MAKES A CHANGE                       │
│  (Update PO, Upload Quotation, Add Attachment, etc.)        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              CAPTURE OLD VALUES (Before)                     │
│  {                                                           │
│    "title": "Office Supplies",                              │
│    "priority": "MEDIUM",                                     │
│    "totalAmount": 3000.00                                    │
│  }                                                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  APPLY THE CHANGES                           │
│  order.Title = "Office Supplies - Updated"                  │
│  order.Priority = "HIGH"                                     │
│  order.TotalAmount = 5000.00                                 │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              CAPTURE NEW VALUES (After)                      │
│  {                                                           │
│    "title": "Office Supplies - Updated",                    │
│    "priority": "HIGH",                                       │
│    "totalAmount": 5000.00                                    │
│  }                                                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              COMPARE & BUILD CHANGES MAP                     │
│  {                                                           │
│    "title": {                                                │
│      "old": "Office Supplies",                              │
│      "new": "Office Supplies - Updated"                     │
│    },                                                        │
│    "priority": {                                             │
│      "old": "MEDIUM",                                        │
│      "new": "HIGH"                                           │
│    },                                                        │
│    "totalAmount": {                                          │
│      "old": 3000.00,                                         │
│      "new": 5000.00                                          │
│    }                                                         │
│  }                                                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│           CREATE COMPLETE DOCUMENT SNAPSHOT                  │
│  {                                                           │
│    "id": "po-789",                                           │
│    "documentNumber": "PO-2025-001",                          │
│    "title": "Office Supplies - Updated",                    │
│    "status": "DRAFT",                                        │
│    "priority": "HIGH",                                       │
│    "totalAmount": 5000.00,                                   │
│    "currency": "ZMW",                                        │
│    "vendorName": "Office Supplies Inc.",                    │
│    "items": [...],                                           │
│    "metadata": {...},                                        │
│    "snapshotTimestamp": "2025-04-03T10:30:00Z"              │
│  }                                                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              CREATE AUDIT LOG ENTRY                          │
│  {                                                           │
│    "id": "audit-123",                                        │
│    "organizationId": "org-456",                              │
│    "documentId": "po-789",                                   │
│    "documentType": "purchase_order",                         │
│    "userId": "user-101",                                     │
│    "actorName": "John Doe",                    ◄─────────── WHO
│    "actorRole": "PROCUREMENT_OFFICER",                       │
│    "action": "updated",                        ◄─────────── WHAT
│    "changes": {...},                           ◄─────────── CHANGES
│    "details": {                                              │
│      "snapshot": {...}                         ◄─────────── SNAPSHOT
│    },                                                        │
│    "createdAt": "2025-04-03T10:30:00Z"        ◄─────────── WHEN
│  }                                                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              SAVE TO DATABASE (Async)                        │
│  INSERT INTO audit_logs (...)                               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│           DISPLAY IN ACTIVITY LOG TAB                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 📝 John Doe updated purchase order                  │   │
│  │    PROCUREMENT_OFFICER • 2 hours ago                │   │
│  │                                                      │   │
│  │    Changes:                                          │   │
│  │    • title: "Office Supplies" → "Office Supplies... │   │
│  │    • priority: "MEDIUM" → "HIGH"                    │   │
│  │    • totalAmount: 3,000.00 → 5,000.00               │   │
│  │                                                      │   │
│  │    [View Full Snapshot]                             │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 🔍 What Gets Tracked

### 1. Document Updates

```
┌──────────────────────────────────────┐
│  📄 DOCUMENT UPDATE                  │
├──────────────────────────────────────┤
│  WHO:   John Doe (PROCUREMENT)       │
│  WHAT:  Updated purchase order       │
│  WHEN:  2025-04-03 10:30:00         │
│                                      │
│  CHANGES:                            │
│  ├─ title: "A" → "B"                │
│  ├─ priority: "MEDIUM" → "HIGH"     │
│  └─ amount: 3000 → 5000             │
│                                      │
│  SNAPSHOT: [Complete PO state]      │
└──────────────────────────────────────┘
```

### 2. Quotation Upload

```
┌──────────────────────────────────────┐
│  💰 QUOTATION UPLOADED               │
├──────────────────────────────────────┤
│  WHO:   Jane Smith (REQUESTER)       │
│  WHAT:  Uploaded quotation           │
│  WHEN:  2025-04-03 09:15:00         │
│                                      │
│  DETAILS:                            │
│  ├─ Vendor: Office Supplies Inc.    │
│  ├─ Amount: ZMW 5,000.00            │
│  ├─ File: quote.pdf (102 KB)        │
│  └─ Total Quotations: 3             │
│                                      │
│  SNAPSHOT: [Complete REQ state]      │
└──────────────────────────────────────┘
```

### 3. Attachment Upload

```
┌──────────────────────────────────────┐
│  📎 ATTACHMENT UPLOADED              │
├──────────────────────────────────────┤
│  WHO:   Mike Johnson (APPROVER)      │
│  WHAT:  Uploaded supporting document │
│  WHEN:  2025-04-03 11:45:00         │
│                                      │
│  DETAILS:                            │
│  ├─ File: invoice.pdf                │
│  ├─ Size: 256 KB                     │
│  ├─ Type: application/pdf            │
│  └─ Total Attachments: 2            │
│                                      │
│  SNAPSHOT: [Complete PO state]       │
└──────────────────────────────────────┘
```

### 4. Status Change

```
┌──────────────────────────────────────┐
│  🔄 STATUS CHANGED                   │
├──────────────────────────────────────┤
│  WHO:   Sarah Lee (FINANCE)          │
│  WHAT:  Changed status               │
│  WHEN:  2025-04-03 14:20:00         │
│                                      │
│  CHANGES:                            │
│  └─ status: "DRAFT" → "PENDING"     │
│                                      │
│  SNAPSHOT: [Complete PV state]       │
└──────────────────────────────────────┘
```

## 📱 Frontend Display

### Activity Log Tab

```
┌─────────────────────────────────────────────────────────────┐
│  Activity Log                                    [Filter ▼] │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 👤 John Doe                              2 hours ago│    │
│  │    PROCUREMENT_OFFICER                              │    │
│  │                                                      │    │
│  │    📝 Updated purchase order                        │    │
│  │                                                      │    │
│  │    Changes:                                          │    │
│  │    • title: "Office Supplies" → "Office Supplies... │    │
│  │    • priority: "MEDIUM" → "HIGH"                    │    │
│  │    • totalAmount: ZMW 3,000.00 → ZMW 5,000.00      │    │
│  │                                                      │    │
│  │    [View Full Snapshot] [View Details]              │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 👤 Jane Smith                            3 hours ago│    │
│  │    REQUESTER                                         │    │
│  │                                                      │    │
│  │    💰 Uploaded quotation                            │    │
│  │                                                      │    │
│  │    Details:                                          │    │
│  │    • Vendor: Office Supplies Inc.                   │    │
│  │    • Amount: ZMW 5,000.00                           │    │
│  │    • File: quote.pdf (102 KB)                       │    │
│  │                                                      │    │
│  │    [View Quotation] [View Snapshot]                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 👤 Mike Johnson                          5 hours ago│    │
│  │    APPROVER                                          │    │
│  │                                                      │    │
│  │    📎 Uploaded attachment                           │    │
│  │                                                      │    │
│  │    Details:                                          │    │
│  │    • File: invoice.pdf                              │    │
│  │    • Size: 256 KB                                   │    │
│  │                                                      │    │
│  │    [Download] [View Snapshot]                       │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 🎨 Change Visualization

### Before/After Comparison

```
┌─────────────────────────────────────────────────────────────┐
│  Field Changes                                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Title                                                       │
│  ┌──────────────────────┐      ┌──────────────────────┐   │
│  │ Office Supplies      │  →   │ Office Supplies -    │   │
│  │                      │      │ Updated              │   │
│  └──────────────────────┘      └──────────────────────┘   │
│                                                              │
│  Priority                                                    │
│  ┌──────────────────────┐      ┌──────────────────────┐   │
│  │ 🟡 MEDIUM            │  →   │ 🔴 HIGH              │   │
│  └──────────────────────┘      └──────────────────────┘   │
│                                                              │
│  Total Amount                                                │
│  ┌──────────────────────┐      ┌──────────────────────┐   │
│  │ ZMW 3,000.00         │  →   │ ZMW 5,000.00         │   │
│  └──────────────────────┘      └──────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 🔐 Security & Compliance

### Immutable Audit Trail

```
┌─────────────────────────────────────────────────────────────┐
│  ✅ AUDIT LOG GUARANTEES                                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  🔒 Immutable       - Cannot be modified or deleted         │
│  👤 Attributed      - Every change has an actor             │
│  ⏰ Timestamped     - Precise time for every action         │
│  📸 Snapshotted     - Complete state preserved              │
│  🏢 Org-Scoped      - Isolated per organization             │
│  🔍 Searchable      - Easy to find specific changes         │
│  📊 Reportable      - Export for compliance                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 📈 Benefits

### 1. Complete Transparency

```
Every change is visible with:
├─ What changed (field-level)
├─ Who changed it (user info)
├─ When it changed (timestamp)
└─ Complete snapshot (full state)
```

### 2. Accountability

```
For every action, we know:
├─ User ID
├─ User Name
├─ User Role
└─ Exact timestamp
```

### 3. Point-in-Time Recovery

```
Snapshots allow:
├─ View document at any point in time
├─ Reconstruct document history
├─ Compare states across time
└─ Recover from errors
```

### 4. Compliance Ready

```
Meets requirements for:
├─ Financial audits
├─ Regulatory compliance
├─ Internal controls
└─ Dispute resolution
```

## 🚀 Quick Start

### 1. Update a Document

```go
// Capture old values
oldValues := map[string]interface{}{
    "title": order.Title,
    "amount": order.TotalAmount,
}

// Make changes
order.Title = "New Title"
order.TotalAmount = 5000.00

// Capture new values
newValues := map[string]interface{}{
    "title": order.Title,
    "amount": order.TotalAmount,
}

// Log with snapshot
changes := services.CompareAndBuildChanges(oldValues, newValues)
snapshot := services.CreateDocumentSnapshot(order)

services.LogDocumentEvent(db, services.DocumentEvent{
    OrganizationID: orgID,
    DocumentID:     order.ID,
    DocumentType:   "purchase_order",
    UserID:         userID,
    ActorName:      userName,
    ActorRole:      userRole,
    Action:         "updated",
    Changes:        changes,
    Snapshot:       snapshot,
})
```

### 2. Upload Quotation

```go
services.LogQuotationUpload(
    db,
    orgID,
    reqID,
    "requisition",
    userID,
    userName,
    userRole,
    "Office Supplies Inc.",
    5000.00,
    "ZMW",
)
```

### 3. Upload Attachment

```go
services.LogAttachmentUpload(
    db,
    orgID,
    poID,
    "purchase_order",
    userID,
    userName,
    userRole,
    "invoice.pdf",
    256000,
)
```

## 📚 Documentation

- `backend/services/audit_service.go` - Core service
- `backend/utils/audit_helper.go` - Helper functions
- `backend/AUDIT_LOGGING_IMPLEMENTATION.md` - Full guide
- `backend/AUDIT_SNAPSHOT_IMPLEMENTATION.md` - Examples
- `AUDIT_TRAIL_TRANSPARENCY_SUMMARY.md` - Overview

---

**Status**: ✅ Ready for Implementation

**Next**: Integrate into document handlers and test!
