# System Overview

**Status**: Phases 1-11 Complete | Ready for Demo
**Last Updated**: 2024-12-01

## What Is Liyali Gateway?

A workflow approval system for processing financial documents through multi-stage approval workflows with real-time analytics.

## Current Capabilities

- 5 workflow types (Requisition, Budget, PO, Payment Voucher, GRN)
- Multi-stage approvals (2-3 stages per type)
- Digital signature capture
- Bulk operations (approve/reject/reassign multiple items)
- Real-time analytics dashboard
- Data persistence across sessions

## Key Features

**Approval Workflows**: Documents route through 2-3 approval stages with digital signatures
**Bulk Operations**: Process multiple documents at once with validation
**Analytics**: Real-time dashboard with metrics, trends, bottleneck analysis
**Data**: All data persists in localStorage (Phase 12 adds PostgreSQL)

## Quick Start

```bash
npm install && npm run dev
# Open http://localhost:3000/workflows/tasks
```

## Next Steps

- Setup: Read `02-QUICK-START.md`
- Demo: Read `03-DEMO-GUIDE.md`
- Architecture: Read `04-ARCHITECTURE.md`
- Phase 12: Read `PHASE-12-PLAN.md`
