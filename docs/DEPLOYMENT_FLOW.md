# Deployment Flow Diagram

## 🔄 Automated Deployment Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Developer Workflow                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ git push origin develop
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    GitHub Actions Trigger                        │
│  • Push to develop branch                                        │
│  • Changes in: backend/**, frontend/**, admin-console/**         │
│  • Manual workflow dispatch                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Change Detection Job                          │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Backend    │  │   Frontend   │  │    Admin     │         │
│  │   Changed?   │  │   Changed?   │  │   Changed?   │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                  │                  │
│         ▼                 ▼                  ▼                  │
│    [Yes/No]          [Yes/No]           [Yes/No]               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Conditional Deployment                        │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  If Backend Changed:                                      │  │
│  │  1. Setup Fly.io CLI                                      │  │
│  │  2. Set secrets (DATABASE_URL, JWT_SECRET, CORS)         │  │
│  │  3. Deploy backend                                        │  │
│  │  4. Run database migrations                               │  │
│  │  5. Verify health check                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  If Frontend Changed:                                     │  │
│  │  1. Setup Fly.io CLI                                      │  │
│  │  2. Set secrets (API_URL, NEXTAUTH_SECRET)               │  │
│  │  3. Deploy frontend                                       │  │
│  │  4. Verify health check                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  If Admin Console Changed:                                │  │
│  │  1. Setup Fly.io CLI                                      │  │
│  │  2. Set secrets (API_URL, NEXTAUTH_SECRET)               │  │
│  │  3. Deploy admin console                                  │  │
│  │  4. Verify health check                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Deployment Complete Job                       │
│                                                                  │
│  • Generate deployment summary                                   │
│  • Report URLs for deployed apps                                 │
│  • Show efficiency metrics                                       │
│  • Set overall status (success/failure)                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Fly.io Platform                               │
│                                                                  │
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │   Frontend       │  │  Admin Console   │                    │
│  │   ✅ Running     │  │   ✅ Running     │                    │
│  └────────┬─────────┘  └────────┬─────────┘                    │
│           │                     │                               │
│           └──────────┬──────────┘                               │
│                      │                                          │
│            ┌─────────▼─────────┐                                │
│            │   Backend API     │                                │
│            │   ✅ Running      │                                │
│            └─────────┬─────────┘                                │
│                      │                                          │
│            ┌─────────▼─────────┐                                │
│            │   PostgreSQL      │                                │
│            │   ✅ Running      │                                │
│            └───────────────────┘                                │
└─────────────────────────────────────────────────────────────────┘
```

## 🎯 Selective Deployment Logic

```
┌─────────────────────────────────────────────────────────────────┐
│                    Change Detection Logic                        │
└─────────────────────────────────────────────────────────────────┘

Scenario 1: Only Admin Console Changed
├─ Backend: ⏭️  Skipped (no changes)
├─ Frontend: ⏭️  Skipped (no changes)
└─ Admin Console: ✅ Deployed

Scenario 2: Backend + Admin Console Changed
├─ Backend: ✅ Deployed
├─ Frontend: ⏭️  Skipped (no changes)
└─ Admin Console: ✅ Deployed

Scenario 3: All Changed
├─ Backend: ✅ Deployed
├─ Frontend: ✅ Deployed
└─ Admin Console: ✅ Deployed

Scenario 4: Workflow File Changed
├─ Backend: ✅ Deployed (forced)
├─ Frontend: ✅ Deployed (forced)
└─ Admin Console: ✅ Deployed (forced)

Scenario 5: No Changes
├─ Backend: ⏭️  Skipped
├─ Frontend: ⏭️  Skipped
└─ Admin Console: ⏭️  Skipped
```

## 🔍 Health Check Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Health Check Process                          │
└─────────────────────────────────────────────────────────────────┘

For Each Deployed App:
│
├─ Step 1: Wait for deployment to complete
│          └─ Timeout: 300-600 seconds
│
├─ Step 2: Wait for app to start
│          └─ Sleep: 30 seconds
│
├─ Step 3: Get app URL from Fly.io
│          └─ Command: flyctl status --json
│
├─ Step 4: Test health endpoint
│          ├─ Backend: /health
│          ├─ Frontend: / or /api/health
│          └─ Admin Console: /
│
├─ Step 5: Retry on failure
│          ├─ Max attempts: 5
│          ├─ Delay: 10 seconds
│          └─ On final failure: Show logs and exit
│
└─ Step 6: Report success
           └─ Add to deployment summary
```

## 🏗️ Application Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Production Architecture                       │
└─────────────────────────────────────────────────────────────────┘

Internet
   │
   ├─────────────────────────────────────────────────────────┐
   │                                                          │
   ▼                                                          ▼
┌──────────────────────┐                          ┌──────────────────────┐
│  liyali-gateway-     │                          │  liyali-admin-       │
│  frontend.fly.dev    │                          │  console.fly.dev     │
│                      │                          │                      │
│  Next.js 16          │                          │  Next.js 16          │
│  Port: 3000          │                          │  Port: 3001          │
│  Auto-scale: 0-N     │                          │  Auto-scale: 0-N     │
└──────────┬───────────┘                          └──────────┬───────────┘
           │                                                 │
           │                                                 │
           └─────────────────┬───────────────────────────────┘
                             │
                             │ HTTPS/REST API
                             │
                             ▼
                  ┌──────────────────────┐
                  │  liyali-gateway-     │
                  │  api.fly.dev         │
                  │                      │
                  │  Go/Fiber            │
                  │  Port: 8080          │
                  │  Auto-scale: 0-N     │
                  └──────────┬───────────┘
                             │
                             │ PostgreSQL Protocol
                             │
                             ▼
                  ┌──────────────────────┐
                  │  liyali-db           │
                  │                      │
                  │  PostgreSQL 15       │
                  │  Port: 5432          │
                  │  Persistent Volume   │
                  └──────────────────────┘

Region: jnb (Johannesburg, South Africa)
Network: Private Fly.io network
HTTPS: Enforced on all apps
CORS: Configured to allow frontend + admin console
```

## 🔐 Security Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Security Architecture                         │
└─────────────────────────────────────────────────────────────────┘

User Request
   │
   ▼
┌──────────────────────┐
│  HTTPS Termination   │  ← Fly.io handles SSL/TLS
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  CORS Check          │  ← Backend validates origin
│  - Frontend allowed  │
│  - Admin allowed     │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  JWT Validation      │  ← Backend validates token
│  - Check signature   │
│  - Check expiry      │
│  - Check claims      │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  RBAC Check          │  ← Backend checks permissions
│  - User role         │
│  - Required perms    │
│  - Admin access      │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Process Request     │  ← Execute business logic
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Audit Log           │  ← Log action for compliance
└──────────────────────┘
```

## 📊 Deployment Metrics

```
┌─────────────────────────────────────────────────────────────────┐
│                    Deployment Efficiency                         │
└─────────────────────────────────────────────────────────────────┘

Traditional Approach (Always deploy all):
├─ Backend: 5-8 minutes
├─ Frontend: 3-5 minutes
├─ Admin Console: 3-5 minutes
└─ Total: 11-18 minutes per deployment

Selective Approach (Only changed apps):
├─ Only Backend: 5-8 minutes
├─ Only Frontend: 3-5 minutes
├─ Only Admin Console: 3-5 minutes
├─ Backend + Frontend: 8-13 minutes
├─ Backend + Admin: 8-13 minutes
├─ Frontend + Admin: 6-10 minutes
└─ All three: 11-18 minutes

Average Time Saved: 40-60% per deployment
Cost Saved: ~50% (fewer build minutes)
```

## 🎯 Deployment Decision Tree

```
                    ┌─────────────────┐
                    │  Push to Develop│
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  Detect Changes │
                    └────────┬────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
                ▼            ▼            ▼
         ┌──────────┐ ┌──────────┐ ┌──────────┐
         │ Backend? │ │Frontend? │ │  Admin?  │
         └────┬─────┘ └────┬─────┘ └────┬─────┘
              │            │            │
         ┌────┴────┐  ┌────┴────┐  ┌────┴────┐
         │Yes │ No │  │Yes │ No │  │Yes │ No │
         └─┬──┴──┬─┘  └─┬──┴──┬─┘  └─┬──┴──┬─┘
           │     │      │     │      │     │
           ▼     │      ▼     │      ▼     │
        Deploy   │   Deploy   │   Deploy   │
        Backend  │   Frontend │   Admin    │
           │     │      │     │      │     │
           └─────┴──────┴─────┴──────┴─────┘
                       │
                       ▼
              ┌─────────────────┐
              │ Verify & Report │
              └─────────────────┘
```

---

**Visual Guide Version**: 1.0
**Last Updated**: February 8, 2026
