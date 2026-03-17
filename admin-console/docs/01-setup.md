# Admin Console Setup

Super-admin portal for platform-level management (organizations, subscriptions, users, system health).

## Prerequisites

- Node.js 18+
- pnpm
- Backend running at `http://localhost:8080`

## Quick Start

```bash
cd admin-console
cp .env.example .env.local
pnpm install
pnpm dev        # http://localhost:3001
```

## Environment Variables

```env
# Baked at build time
NEXT_PUBLIC_API_URL=http://localhost:8080

# Runtime (server actions)
BASE_URL=http://localhost:8080
AUTH_SECRET=<min-32-char-secret>
```

## Dev Commands

```bash
pnpm dev          # dev server (port 3001)
pnpm build        # production build
pnpm lint         # ESLint
pnpm type-check   # tsc --noEmit
```

## App Structure

```
admin-console/src/app/
├── (auth)/              # login page
├── admin/
│   ├── dashboard/       # overview metrics
│   ├── organizations/   # all orgs + tier management
│   ├── subscriptions/   # subscription plans + trials
│   ├── users/           # all platform users
│   ├── admin-users/     # super-admin user management
│   ├── roles/           # global role management
│   ├── feature-flags/   # feature flag toggles
│   ├── audit-logs/      # immutable audit trail
│   ├── analytics/       # platform-wide analytics
│   ├── system-health/   # service health + metrics
│   ├── api-monitoring/  # API request logs
│   ├── database/        # DB health + query stats
│   ├── impersonation/   # user impersonation logs
│   ├── notifications/   # notification management
│   └── settings/        # system settings
└── _actions/            # server actions (API calls)
```

## Access

Only users with `is_super_admin = true` can log in. The seeded super-admin account:

| Field | Value |
|---|---|
| Email | `superadmin@liyali.com` |
| Password | `password` |
