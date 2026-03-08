# Dashboard & Analytics Audit Report

**Date**: 2026-03-08  
**Status**: ✅ Phase 1 Complete | ⏳ Phases 2-4 Pending  
**Scope**: Frontend Dashboard Analytics (Section 4.1 of TODO.md)

---

## Executive Summary

**Key Finding**: Backend has comprehensive analytics for ALL document types via admin reports endpoints, but frontend dashboard only uses requisition-specific analytics. The data exists - it just needs to be exposed to regular users with role-based filtering.

**Phase 1 Status**: ✅ COMPLETE (2026-03-08)

- Backend unified reports endpoints created
- Frontend updated to use comprehensive metrics
- All document types now visible on dashboard
- Average approval time and recent activity now working

**Remaining Work**: Phases 2-4 (Optional enhancements)

- Phase 2: Budget utilization calculation
- Phase 3: Role-based dashboard views
- Phase 4: Processing time tracking

---

## Backend Implementation Status

### ✅ Fully Implemented - Admin Reports (`/api/admin/reports/*`)

All endpoints are organization-scoped and fully functional:

#### 1. System Statistics

**Endpoint**: `GET /api/admin/reports/system-stats`  
**Handler**: `backend/handlers/reports.go:GetSystemStatistics`  
**Service**: `backend/services/reports_service.go:GetSystemStatistics`

**Returns**:

- Total documents (all types)
- Approved/rejected/draft/submitted/pending counts
- Average approval time (AvgApprovalDays) ✅
- Approval rate & rejection rate
- **Document type breakdown** ✅
  - Requisitions count
  - Purchase Orders count
  - Payment Vouchers count
  - GRNs count
  - Budgets count
- **Status breakdown** ✅
  - Draft, Submitted, InReview, Approved, Rejected

#### 2. Approval Metrics

**Endpoint**: `GET /api/admin/reports/approval-metrics`  
**Handler**: `backend/handlers/reports.go:GetApprovalMetrics`

**Returns**:

- Total approved/rejected/pending
- Approval rate
- **Recent approvals** (last 50 actions) ✅
  - Document ID, number, type
  - Action (approved/rejected)
  - Approver name & role
  - Comments
  - Timestamp

#### 3. User Activity Metrics

**Endpoint**: `GET /api/admin/reports/user-activity`  
**Handler**: `backend/handlers/reports.go:GetUserActivityMetrics`

**Returns**:

- Active users count
- Total actions
- Documents in progress
- Per-user statistics:
  - Approval count
  - Rejection count
  - Active documents
  - Last activity timestamp

#### 4. Analytics Dashboard

**Endpoint**: `GET /api/admin/reports/analytics`  
**Handler**: `backend/handlers/reports.go:GetAnalyticsDashboard`

**Returns**:

- Total pending/approved/rejected
- Average approval time
- SLA compliance percentage
- **Approval trends** (daily data) ✅
- **Document distribution** (by type with percentages) ✅
- **Stage metrics** (processing time per stage) ✅
- **Bottleneck identification** (slowest stage) ✅

### ⚠️ Partially Implemented - Tenant Analytics (`/api/v1/analytics/*`)

#### 1. Dashboard

**Endpoint**: `GET /api/v1/analytics/dashboard`  
**Handler**: `backend/handlers/analytics.go:GetDashboard`

**Returns**: Only requisition metrics

- Status counts (requisitions only)
- Rejection rate (requisitions only)
- Rejections over time (requisitions only)

**Missing**: PO, PV, GRN, Budget metrics

#### 2. Requisition Metrics

**Endpoint**: `GET /api/v1/analytics/requisitions/metrics`  
**Status**: ✅ Complete for requisitions

#### 3. Approval Metrics

**Endpoint**: `GET /api/v1/analytics/approvals/metrics`  
**Status**: ✅ Complete (extracts from requisition metrics)

---

## Data Availability Matrix

| Metric                        | Admin Reports       | Tenant Dashboard | Frontend Uses | Gap           |
| ----------------------------- | ------------------- | ---------------- | ------------- | ------------- |
| **Document Counts**           |
| All document types            | ✅ system-stats     | ❌               | ❌            | Admin has it  |
| Requisitions only             | ✅                  | ✅               | ✅            | Working       |
| Purchase Orders               | ✅ system-stats     | ❌               | ❌            | Admin has it  |
| Payment Vouchers              | ✅ system-stats     | ❌               | ❌            | Admin has it  |
| GRNs                          | ✅ system-stats     | ❌               | ❌            | Admin has it  |
| Budgets                       | ✅ system-stats     | ❌               | ❌            | Admin has it  |
| **Time Metrics**              |
| Average approval time         | ✅ system-stats     | ❌               | ❌            | Admin has it  |
| Stage processing time         | ✅ analytics        | ❌               | ❌            | Admin only    |
| SLA compliance                | ✅ analytics        | ❌               | ❌            | Admin only    |
| **Status & Rates**            |
| Status breakdown (all)        | ✅ system-stats     | ⚠️ Req only      | ⚠️ Req only   | Admin has all |
| Approval rate                 | ✅ system-stats     | ⚠️ Req only      | ⚠️ Req only   | Admin has all |
| Rejection rate                | ✅ system-stats     | ✅ Req           | ✅ Req        | Working       |
| **Activity & Trends**         |
| Recent activity feed          | ✅ approval-metrics | ❌               | ❌            | Admin has it  |
| Approval trends               | ✅ analytics        | ❌               | ❌            | Admin only    |
| Document distribution         | ✅ analytics        | ❌               | ❌            | Admin only    |
| Bottleneck analysis           | ✅ analytics        | ❌               | ❌            | Admin only    |
| **User Metrics**              |
| User activity                 | ✅ user-activity    | ❌               | ❌            | Admin only    |
| Active users                  | ✅ user-activity    | ❌               | ❌            | Admin only    |
| **Not Implemented**           |
| Budget utilization %          | ❌                  | ❌               | ❌            | Need to add   |
| Processing time (vs approval) | ❌                  | ❌               | ❌            | Need to add   |

---

## Frontend Implementation Status

### ✅ Phase 1 Complete (High Priority)

1. **Dashboard shows all document types** ✅ DONE
   - **Status**: Implemented
   - **Endpoint**: `/api/v1/reports/dashboard`
   - **Impact**: Users can now see PO, PV, GRN, Budget counts
   - **Files Modified**:
     - `backend/routes/routes.go`
     - `backend/handlers/reports.go`
     - `frontend/src/app/_actions/dashboard.ts`

2. **Document type breakdown** ✅ DONE
   - **Status**: Implemented
   - **Data**: Complete TypeBreakdown from backend
   - **Impact**: Full visibility into document type distribution

3. **Average approval time on dashboard** ✅ DONE
   - **Status**: Implemented
   - **Data**: AvgApprovalDays from system-stats
   - **Impact**: Real performance metrics visible

4. **Recent activity feed** ✅ DONE
   - **Status**: Implemented
   - **Data**: RecentApprovals from approval-metrics
   - **Impact**: Users can see last 50 approval actions

### ⏳ Phase 2 Pending (Medium Priority)

5. **Budget utilization metric** ❌ NOT IMPLEMENTED
   - **Status**: Needs backend calculation
   - **Required**: `SUM(allocated_amount) / SUM(total_budget)`
   - **Effort**: ~4 hours backend + 2 hours frontend
   - **Priority**: Medium

### ⏳ Phase 3 Pending (Medium Priority - Optional)

6. **Approval trends chart** ❌ NOT IMPLEMENTED
   - **Status**: Data available in analytics endpoint
   - **Required**: Frontend chart component
   - **Effort**: ~8 hours
   - **Priority**: Medium (nice to have)

7. **Role-based dashboard views** ❌ NOT IMPLEMENTED
   - **Status**: Backend filtering needed
   - **Required**:
     - Admin: See all organization data
     - Manager: See department data
     - User: See own documents + pending approvals
   - **Effort**: 2-3 days
   - **Priority**: Medium (UX enhancement)

8. **Stage metrics & bottleneck analysis** ❌ NOT IMPLEMENTED
   - **Status**: Data available in analytics endpoint
   - **Required**: Frontend visualization
   - **Effort**: ~8 hours
   - **Priority**: Low (admin feature)

### ⏳ Phase 4 Pending (Low Priority - Optional)

9. **Processing time (separate from approval time)** ❌ NOT IMPLEMENTED
   - **Status**: Not tracked anywhere
   - **Required**: Backend tracking from creation to completion
   - **Effort**: 1 day
   - **Priority**: Low

---

## Recommendations

### Phase 1: Expose Existing Data (High Priority)

**Goal**: Make admin reports available to all users with role-based filtering

#### Backend Changes

1. **Create unified reports endpoints** (`/api/v1/reports/*`)

   ```go
   // New routes in backend/routes/routes.go
   reports := tenant.Group("/reports")
   reports.Get("/dashboard", handlers.GetDashboardReports)      // All users
   reports.Get("/system-stats", handlers.GetSystemStatistics)   // All users (filtered)
   reports.Get("/approval-metrics", handlers.GetApprovalMetrics) // All users (filtered)
   reports.Get("/user-activity", middleware.RequireAdmin(), handlers.GetUserActivity) // Admin only
   reports.Get("/analytics", middleware.RequireAdmin(), handlers.GetAnalytics) // Admin only
   ```

2. **Add role-based filtering**

   ```go
   // In handlers, filter data based on role:
   // - Admin: See all organization data
   // - Manager: See department data
   // - User: See own documents + pending approvals for them
   ```

3. **Update tenant dashboard endpoint**
   ```go
   // Modify GET /api/v1/analytics/dashboard
   // Instead of calling GetRequisitionMetrics
   // Call GetSystemStatistics (returns all document types)
   ```

#### Frontend Changes

1. **Update dashboard actions** (`frontend/src/app/_actions/dashboard.ts`)

   ```typescript
   // Change from:
   const url = "/api/v1/analytics/dashboard"; // Requisitions only

   // To:
   const url = "/api/v1/reports/system-stats"; // All document types
   ```

2. **Update dashboard types** (`frontend/src/types/dashboard.ts`)

   ```typescript
   interface DashboardMetrics {
     // Add fields from SystemStatistics
     documentTypeBreakdown: {
       requisitions: number;
       purchaseOrders: number;
       paymentVouchers: number;
       grn: number;
       budgets: number;
     };
     averageApprovalTime: number; // From AvgApprovalDays
     recentActivity: ApprovalActivity[]; // From approval-metrics
     // ... existing fields
   }
   ```

3. **Update dashboard UI** to display:
   - Document type breakdown chart
   - Average approval time metric
   - Recent activity feed
   - All document status counts (not just requisitions)

### Phase 2: Add Budget Utilization (Medium Priority)

1. **Backend**: Add budget utilization calculation

   ```go
   // In reports_repository.go
   func (r *ReportsRepository) QueryBudgetUtilization(
     ctx context.Context,
     organizationID string,
   ) (float64, error) {
     query := `
       SELECT
         CASE
           WHEN SUM(total_budget) = 0 THEN 0
           ELSE (SUM(allocated_amount) / SUM(total_budget)) * 100
         END as utilization_percentage
       FROM budgets
       WHERE organization_id = $1
         AND status != 'rejected'
     `
     // ... implementation
   }
   ```

2. **Add to SystemStatistics model**

   ```go
   type SystemStatistics struct {
     // ... existing fields
     BudgetUtilization float64 `json:"budgetUtilization"`
   }
   ```

3. **Frontend**: Display budget utilization gauge/chart

### Phase 3: Role-Based Dashboard Views (Medium Priority)

1. **Admin View**: Full organization visibility
   - All documents
   - All users
   - All departments
   - Bottleneck analysis
   - Stage metrics

2. **Manager View**: Department visibility
   - Department documents
   - Department users
   - Department metrics

3. **User View**: Personal visibility
   - Own documents
   - Pending approvals assigned to them
   - Personal metrics

### Phase 4: Processing Time Tracking (Low Priority)

1. **Add processing time calculation**
   - Track: `created_at` → `completed_at` (final status)
   - Different from approval time (workflow duration)

2. **Add to metrics**
   ```go
   type SystemStatistics struct {
     // ... existing fields
     AverageProcessingTime float64 `json:"averageProcessingTime"` // Days
   }
   ```

---

## Implementation Effort Estimate

| Phase               | Tasks                           | Effort   | Priority |
| ------------------- | ------------------------------- | -------- | -------- |
| Phase 1             | Expose existing data            | 2-3 days | High     |
| - Backend routes    | Add unified reports endpoints   | 4 hours  | High     |
| - Backend filtering | Role-based data filtering       | 4 hours  | High     |
| - Frontend actions  | Update API calls                | 2 hours  | High     |
| - Frontend types    | Update TypeScript types         | 2 hours  | High     |
| - Frontend UI       | Update dashboard components     | 8 hours  | High     |
| Phase 2             | Budget utilization              | 1 day    | Medium   |
| - Backend query     | Add utilization calculation     | 3 hours  | Medium   |
| - Backend model     | Update SystemStatistics         | 1 hour   | Medium   |
| - Frontend UI       | Add utilization display         | 4 hours  | Medium   |
| Phase 3             | Role-based views                | 2-3 days | Medium   |
| - Backend filtering | Implement role filters          | 8 hours  | Medium   |
| - Frontend views    | Create role-specific dashboards | 8 hours  | Medium   |
| Phase 4             | Processing time                 | 1 day    | Low      |
| - Backend tracking  | Add processing time calc        | 4 hours  | Low      |
| - Frontend display  | Show processing metrics         | 4 hours  | Low      |

**Total Estimated Effort**: 6-8 days for all phases

---

## Testing Checklist

### Phase 1 Testing

- [ ] Admin can see all document types on dashboard
- [ ] Manager can see department documents only
- [ ] User can see own documents + pending approvals
- [ ] Document type breakdown chart displays correctly
- [ ] Average approval time shows accurate data
- [ ] Recent activity feed displays last 50 actions
- [ ] All status counts include all document types
- [ ] Role-based filtering works correctly

### Phase 2 Testing

- [ ] Budget utilization calculates correctly
- [ ] Utilization updates when budgets change
- [ ] Handles zero total budget gracefully
- [ ] Excludes rejected budgets from calculation

### Phase 3 Testing

- [ ] Admin sees full organization view
- [ ] Manager sees only department data
- [ ] User sees only personal data
- [ ] Role changes reflect immediately
- [ ] Multi-department users see correct data

### Phase 4 Testing

- [ ] Processing time calculates correctly
- [ ] Different from approval time
- [ ] Handles incomplete documents
- [ ] Shows accurate averages

---

## Conclusion

The backend has comprehensive analytics infrastructure that covers all document types, approval metrics, user activity, and trends. The main gap is in exposing this data to regular users through the frontend dashboard.

**Key Actions**:

1. ✅ Backend has all the data (admin reports)
2. ❌ Frontend only uses requisition-specific analytics
3. 🎯 Solution: Expose admin reports to all users with role-based filtering
4. 📊 Impact: Complete dashboard visibility for all users

**Next Steps**:

1. Start with Phase 1 (expose existing data) - highest ROI
2. Add budget utilization (Phase 2) - fills remaining gap
3. Implement role-based views (Phase 3) - better UX
4. Add processing time tracking (Phase 4) - nice to have
