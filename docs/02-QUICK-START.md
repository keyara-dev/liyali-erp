# Quick Start Guide

## Installation

```bash
npm install
```

## Development Server

```bash
npm run dev
```

Open [http://localhost:3000/workflows/tasks](http://localhost:3000/workflows/tasks)

## First Steps

1. **View Tasks**: Tasks tab shows all pending items
2. **View Approvals**: Click Approvals tab to see approval cards
3. **Approve a Task**: Click any card, draw signature, click Approve
4. **View Analytics**: Go to Admin Reports → Analytics tab

## Key Pages

| Page | URL | What You'll See |
|------|-----|-----------------|
| Tasks | `/workflows/tasks` | Tasks to do and approvals |
| PO Detail | `/workflows/purchase-orders/[id]` | Purchase order details |
| Approve PO | `/workflows/purchase-orders/[id]/approval` | Approval form |
| Admin Reports | `/admin/reports` | System analytics |

## Pre-loaded Demo Data

3 sample tasks are ready to test:
- REQ-2024-001 (K25,000 Requisition)
- BUD-2024-Q1-001 (K500,000 Budget)
- REQ-2024-002 (K5,000 Requisition)

## Commands

```bash
npm run build      # Production build
npm run test       # Run tests
npm run type-check # TypeScript check
```

## Troubleshooting

**No data visible?**
- Open DevTools (F12)
- Check Application → Local Storage
- Should have `approval_tasks_v1` key

**Page shows error?**
- Clear browser cache (Ctrl+Shift+Delete)
- Hard refresh (Ctrl+Shift+R)
- Restart dev server

## Next

Read `03-DEMO-GUIDE.md` to see all features.
