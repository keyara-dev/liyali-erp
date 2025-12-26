# Backend Endpoint Requirements - Blockers #3 & #6
**Date**: 2025-12-26
**Priority**: CRITICAL - Blocking MVP
**Status**: Ready for implementation

---

## Overview

Two frontend blockers require backend API endpoints to function:
- **BLOCKER #3**: Admin metrics endpoints (4 endpoints)
- **BLOCKER #6**: Purchase order detail endpoint (1 endpoint)

This document specifies exact requirements for each endpoint.

---

## BLOCKER #3: Admin Metrics Endpoints

### Endpoint 1: System Health Metrics

**Route**: `GET /api/v1/admin/metrics/system-health`

**Authentication**: Required (JWT)
**Authorization**: Admin role required

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "approvals_24h": 45,
    "submissions_24h": 87,
    "rejections_24h": 12,
    "error_rate": 0.5,
    "uptime_percent": 99.8,
    "response_time_ms": 145,
    "last_updated": "2025-12-26T16:30:00Z"
  },
  "timestamp": "2025-12-26T16:30:00Z"
}
```

**Purpose**: Display current system metrics on monitoring dashboard
**Caching**: OK to cache for 30 seconds

---

### Endpoint 2: Hourly Metrics History

**Route**: `GET /api/v1/admin/metrics/hourly?hours=24`

**Query Parameters**:
- `hours` (optional, default 24): Number of hours of historical data to return
- Can be: 1, 6, 12, 24, 48, 72, 168 (7 days)

**Authentication**: Required (JWT)
**Authorization**: Admin role required

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "hour": "2025-12-25T16:00:00Z",
      "approvals": 5,
      "submissions": 8,
      "rejections": 1,
      "pending": 12,
      "error_count": 0,
      "avg_processing_time_ms": 132
    },
    {
      "hour": "2025-12-25T17:00:00Z",
      "approvals": 7,
      "submissions": 11,
      "rejections": 2,
      "pending": 15,
      "error_count": 1,
      "avg_processing_time_ms": 157
    },
    // ... one entry per hour
  ],
  "period_hours": 24,
  "total_approvals": 342,
  "total_submissions": 687,
  "total_rejections": 54,
  "timestamp": "2025-12-26T16:00:00Z"
}
```

**Purpose**: Display hourly metrics chart on monitoring dashboard
**Note**: Return data in chronological order (oldest first)
**Caching**: OK to cache for 60 seconds

---

### Endpoint 3: User Metrics (for User Details Page)

**Route**: `GET /api/v1/admin/users/{userId}/metrics`

**Path Parameters**:
- `userId` (required): ID of user to get metrics for

**Authentication**: Required (JWT)
**Authorization**: Admin role required OR user viewing own metrics

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "user_id": "user-123",
    "total_documents": 47,
    "risk_score": 65,
    "compliance_status": "PASSING",
    "last_activity": "2025-12-26T14:30:00Z",
    "recent_activities": [
      {
        "id": "activity-1",
        "type": "requisition_approved",
        "action": "Approved requisition REQ-2025-001",
        "details": "Purchase request for office supplies",
        "timestamp": "2025-12-26T14:30:00Z",
        "metadata": {
          "document_id": "req-123",
          "document_number": "REQ-2025-001"
        }
      },
      {
        "id": "activity-2",
        "type": "budget_updated",
        "action": "Updated budget allocation",
        "details": "Increased IT department budget",
        "timestamp": "2025-12-26T10:15:00Z",
        "metadata": {
          "document_id": "budget-456",
          "amount": 50000
        }
      },
      // ... up to 10 recent activities (latest first)
    ],
    "audit_metrics": {
      "total_audits": 8,
      "completed_audits": 5,
      "in_progress_audits": 2,
      "upcoming_audits": 1,
      "findings_created": 24,
      "open_findings": 6,
      "resolution_rate_percent": 75
    },
    "risk_metrics": {
      "total_risks": 12,
      "critical_risks": 2,
      "high_risks": 5,
      "medium_risks": 3,
      "low_risks": 2,
      "risk_score": 65
    }
  },
  "timestamp": "2025-12-26T16:00:00Z"
}
```

**Purpose**: Display user metrics on user details admin page
**Caching**: OK to cache for 60 seconds

---

### Endpoint 4: Reports Analytics Data

**Route**: `GET /api/v1/admin/reports/analytics`

**Query Parameters**:
- `format` (optional, default "json"): Can be "json" or "csv"
- `start_date` (optional): ISO date string for report start
- `end_date` (optional): ISO date string for report end
- `department_id` (optional): Filter by department

**Authentication**: Required (JWT)
**Authorization**: Admin role required

**Response (JSON)** (200 OK):
```json
{
  "success": true,
  "data": {
    "total_pending": 24,
    "total_approved": 1247,
    "total_rejected": 83,
    "total_processed": 1330,
    "approval_rate_percent": 93.8,
    "rejection_rate_percent": 6.2,
    "avg_approval_time_hours": 3.2,
    "avg_approval_time_minutes": 192,
    "median_approval_time_hours": 2.5,
    "sla_compliance_percent": 94.5,
    "period": {
      "start_date": "2025-12-01T00:00:00Z",
      "end_date": "2025-12-26T23:59:59Z",
      "days": 26
    },
    "by_document_type": {
      "requisitions": {
        "total": 450,
        "approved": 425,
        "rejected": 25,
        "approval_rate": 94.4
      },
      "budgets": {
        "total": 180,
        "approved": 172,
        "rejected": 8,
        "approval_rate": 95.6
      },
      "purchase_orders": {
        "total": 320,
        "approved": 295,
        "rejected": 25,
        "approval_rate": 92.2
      },
      "payment_vouchers": {
        "total": 280,
        "approved": 255,
        "rejected": 25,
        "approval_rate": 91.1
      }
    },
    "by_department": [
      {
        "department_id": "dept-1",
        "department_name": "IT",
        "total_documents": 234,
        "approved": 220,
        "rejected": 14,
        "approval_rate": 94.0
      },
      // ... more departments
    ]
  },
  "timestamp": "2025-12-26T16:00:00Z"
}
```

**Response (CSV)** (200 OK):
```csv
Workflow Analytics Report
Generated: 2025-12-26T16:00:00Z

Total Pending,24
Total Approved,1247
Total Rejected,83
Total Processed,1330
Approval Rate %,93.8
Rejection Rate %,6.2
Average Approval Time (hours),3.2
Median Approval Time (hours),2.5
SLA Compliance %,94.5

By Document Type
Type,Total,Approved,Rejected,Approval Rate %
Requisitions,450,425,25,94.4
Budgets,180,172,8,95.6
Purchase Orders,320,295,25,92.2
Payment Vouchers,280,255,25,91.1

By Department
Department,Total,Approved,Rejected,Approval Rate %
IT,234,220,14,94.0
Finance,187,175,12,93.5
...
```

**Purpose**: Display analytics on reports dashboard and export reports
**Caching**: OK to cache for 5 minutes

---

## BLOCKER #6: Purchase Order Detail Endpoint

### Endpoint: Get Purchase Order Details

**Route**: `GET /api/v1/purchase-orders/{poId}`

**Path Parameters**:
- `poId` (required): ID of purchase order to retrieve

**Authentication**: Required (JWT)
**Authorization**: User must have access to org that PO belongs to

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": "po-12345",
    "po_number": "PO-2025-001234",
    "organization_id": "org-123",
    "vendor": {
      "id": "vendor-456",
      "name": "Global Supplies Inc.",
      "email": "contact@globalsupplies.com",
      "phone": "+1-555-0123",
      "address": {
        "country": "United States",
        "city": "New York"
      }
    },
    "items": [
      {
        "id": "item-1",
        "description": "Office Supplies Pack",
        "quantity": 50,
        "unit_price": 25.50,
        "total_price": 1275.00,
        "category": "Office Supplies"
      },
      {
        "id": "item-2",
        "description": "Printer Paper (500 sheets)",
        "quantity": 100,
        "unit_price": 5.00,
        "total_price": 500.00,
        "category": "Paper Products"
      }
    ],
    "delivery_date": "2025-12-31T00:00:00Z",
    "delivery_location": "Main Office",
    "total_amount": 1775.00,
    "currency": "USD",
    "status": "approved",
    "approval_stage": "final",
    "created_by": {
      "id": "user-123",
      "name": "John Requester",
      "email": "john@company.com"
    },
    "created_at": "2025-12-20T10:30:00Z",
    "updated_at": "2025-12-23T14:15:00Z",
    "linked_requisition": {
      "id": "req-789",
      "req_number": "REQ-2025-0456",
      "title": "Office Supplies Request"
    },
    "approval_history": [
      {
        "stage": "requester",
        "status": "approved",
        "approved_by": {
          "id": "user-123",
          "name": "John Requester",
          "email": "john@company.com"
        },
        "approved_at": "2025-12-20T10:30:00Z",
        "comments": "Request approved for processing"
      },
      {
        "stage": "manager",
        "status": "approved",
        "approved_by": {
          "id": "user-456",
          "name": "Jane Manager",
          "email": "jane@company.com"
        },
        "approved_at": "2025-12-21T09:00:00Z",
        "comments": "Approved by department manager"
      },
      {
        "stage": "finance",
        "status": "approved",
        "approved_by": {
          "id": "user-789",
          "name": "Bob Finance",
          "email": "bob@company.com"
        },
        "approved_at": "2025-12-23T14:15:00Z",
        "comments": "Budget verified and approved"
      }
    ]
  },
  "timestamp": "2025-12-26T16:00:00Z"
}
```

**Response (404 Not Found)**:
```json
{
  "success": false,
  "error": "Purchase order not found",
  "code": "PO_NOT_FOUND"
}
```

**Response (403 Forbidden)**:
```json
{
  "success": false,
  "error": "Access denied to this purchase order",
  "code": "UNAUTHORIZED"
}
```

**Purpose**: Display PO details in frontend po-detail-client component
**Caching**: OK to cache for 2 minutes

---

## Implementation Notes

### Error Handling
All endpoints should:
- Return 401 if not authenticated
- Return 403 if user lacks required permissions
- Return 400 for invalid query parameters
- Return 500 with error details for server errors

### Response Format
All endpoints must follow the standard response format:
```json
{
  "success": boolean,
  "data": object | array,
  "error": string | null,
  "timestamp": ISO 8601 datetime
}
```

### Performance Considerations
- All endpoints should be fast (< 500ms response time)
- Consider indexes on frequently queried fields
- Implement pagination where appropriate
- Cache where indicated (frontend will handle further caching)

### Security
- All endpoints require authentication (JWT)
- Verify organization/department access
- Don't return sensitive data (passwords, tokens, etc.)
- Log access to sensitive data

### Testing
Provide sample cURL commands for testing each endpoint:

```bash
# System metrics
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/admin/metrics/system-health

# Hourly metrics
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:8080/api/v1/admin/metrics/hourly?hours=24"

# User metrics
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/admin/users/user-123/metrics

# Reports analytics
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:8080/api/v1/admin/reports/analytics?format=json"

# Purchase order detail
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/purchase-orders/po-12345
```

---

## Timeline

**Estimated Effort Per Endpoint**:
- System Health Metrics: 2-3 hours
- Hourly Metrics: 3-4 hours (query optimization needed)
- User Metrics: 2-3 hours
- Reports Analytics: 3-4 hours
- PO Detail: 1-2 hours (likely already partially implemented)

**Total**: 11-16 hours (2-3 developer-days)

**Critical Path**:
1. **Day 1**: Implement System Health + Hourly Metrics (foundation)
2. **Day 1-2**: Implement User Metrics + Reports
3. **Day 2**: Test and verify with frontend team

---

## Frontend Integration Status

### Ready to Integrate Once Backend Complete
1. **Monitoring Client** - Waits for: System Health + Hourly Metrics
2. **User Details** - Waits for: User Metrics
3. **Admin Reports** - Waits for: Reports Analytics
4. **PO Detail** - Waits for: PO endpoint verification

### Frontend Files Awaiting Backend
- `frontend/src/app/(private)/admin/monitoring/_components/monitoring-client.tsx`
- `frontend/src/app/(private)/admin/users/[id]/user-details-client.tsx`
- `frontend/src/app/(private)/admin/reports/_components/admin-reports-client.tsx`
- `frontend/src/app/(private)/(main)/purchase-orders/[id]/_components/po-detail-client.tsx`

---

## Related Issues

**Blockers**:
- #3: Mock data in admin pages
- #6: PO using generated mock data

**Dependencies**:
- None (these are new endpoints)

**Tests Needed**:
- Unit tests for metric calculations
- Integration tests for auth/org access
- Performance tests for large datasets

---

## Sign-Off

**Prepared By**: Comprehensive Code Audit
**Date**: 2025-12-26
**Status**: Ready for Backend Implementation
**Blocking**: MVP Release (CRITICAL)

---

**Questions?** Contact frontend team for clarification on response format or data requirements.
