# Liyali Gateway - Quick Reference

**Last Updated**: February 23, 2026

---

## Project Overview

Liyali Gateway is a procurement management system with approval workflows, document management, and reporting capabilities.

**Tech Stack**: Go/Fiber backend + Next.js 16 frontend + PostgreSQL

---

## Quick Start

```bash
# Backend
cd backend
cp .env.example .env  # Configure DATABASE_URL
go run cmd/migrate/main.go
go run main.go  # http://localhost:8081

# Frontend
cd frontend
cp .env.example .env  # Configure NEXT_PUBLIC_API_URL
npm install
npm run dev  # http://localhost:3000
```

---

## Key Directories

```
backend/
├── handlers/      # HTTP endpoints
├── services/      # Business logic
├── repository/    # Database queries
├── models/        # Data structures
└── database/migrations/  # SQL migrations

frontend/src/
├── app/_actions/  # Server actions (API calls)
├── hooks/         # React Query hooks
├── components/    # UI components
└── types/         # TypeScript interfaces
```

---

## Common Patterns

### Backend: Add Endpoint

```go
// 1. Model (models/)
type Item struct { ID string `json:"id"` }

// 2. Repository (repository/)
func (r *Repo) Get(ctx, orgID) (*Item, error) { /* SQL */ }

// 3. Service (services/)
func (s *Service) Get(ctx, orgID) (*Item, error) { return s.repo.Get(ctx, orgID) }

// 4. Handler (handlers/)
func (h *Handler) Get(c *fiber.Ctx) error {
    tenant, _ := middleware.GetTenantContext(c)
    item, _ := h.service.Get(c.Context(), tenant.OrganizationID)
    return c.JSON(item)
}

// 5. Route (routes/)
protected.Get("/items", handlers.Item.Get)
```

### Frontend: Fetch Data

```typescript
// 1. Type (types/)
export interface Item {
  id: string;
}

// 2. Server Action (app/_actions/)
("use server");
export async function getItems() {
  const res = await authenticatedApiClient({
    method: "GET",
    url: "/api/v1/items",
  });
  return res.data;
}

// 3. Hook (hooks/)
export function useItems() {
  return useQuery({ queryKey: ["items"], queryFn: getItems });
}

// 4. Component
const { data, isLoading } = useItems();
```

---

## Database

### Migrations

```bash
cd backend
export DATABASE_URL="postgres://..."
go run cmd/migrate/main.go
```

### Create Migration

```bash
cd backend/database/migrations
touch 015_feature.up.sql 015_feature.down.sql
```

### Query Pattern

```sql
-- Always filter by organization_id
SELECT * FROM table WHERE organization_id = $1 AND status = $2
```

---

## Authentication

### Backend: Get User Context

```go
tenant, err := middleware.GetTenantContext(c)
// tenant.OrganizationID, tenant.UserID, tenant.UserRole
```

### Frontend: Verify Session

```typescript
const { session } = await verifySession();
// session.user.id, session.user.role
```

---

## API Endpoints

### Core

```
POST   /api/v1/auth/login
POST   /api/v1/auth/register
GET    /api/v1/auth/me
```

### Documents

```
GET    /api/v1/requisitions
POST   /api/v1/requisitions
GET    /api/v1/requisitions/:id
PUT    /api/v1/requisitions/:id
DELETE /api/v1/requisitions/:id
POST   /api/v1/requisitions/:id/submit
```

### Admin

```
GET    /api/v1/admin/reports/system-stats
GET    /api/v1/admin/reports/approval-metrics
GET    /api/v1/admin/reports/user-activity
GET    /api/v1/admin/reports/analytics
```

---

## Environment Variables

### Backend (.env)

```env
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
JWT_SECRET=your-secret-key
APP_PORT=8081
FRONTEND_URL=https://your-frontend.com
```

### Frontend (.env)

```env
NEXT_PUBLIC_API_URL=https://your-backend.com
NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY=your_key
IMAGEKIT_PRIVATE_KEY=your_private_key
```

---

## Key Features

1. **Admin Reports** - Live analytics at `/admin` (admin only)
2. **Workflow Selection** - Dynamic workflow assignment
3. **Configuration Checklist** - Pre-creation validation
4. **Logo Upload** - ImageKit CDN integration
5. **Session Management** - Auto-refresh + timeout warning
6. **PDF Generation** - All document types
7. **Subscription System** - Trial + paid plans

---

## Deployment

### Using Makefile (Recommended)

```bash
# Deploy all apps
make deploy

# Deploy individual apps
make deploy-backend    # Backend only
make deploy-web        # Web frontend only
make deploy-admin      # Admin console only

# Pre-deployment checks
make pre-deploy        # Verify env, build, test, migrate
```

### Fly.io (Direct)

```bash
cd backend && fly deploy
cd frontend && fly deploy
cd admin-console && fly deploy
```

### Manual

```bash
# Backend
cd backend && go build -o app .
# Deploy binary + set env vars

# Frontend
cd frontend && npm run build
# Deploy .next/ + set env vars

# Admin Console
cd admin-console && npm run build
# Deploy .next/ + set env vars
```

---

## Troubleshooting

### Backend won't start

- Check DATABASE_URL format
- Verify database is accessible
- Run migrations

### Frontend build fails

- Check TypeScript errors: `npm run build`
- Clear cache: `rm -rf .next node_modules && npm install`

### API calls fail

- Check NEXT_PUBLIC_API_URL
- Verify CORS settings (FRONTEND_URL in backend)
- Check authentication token

---

## Useful Commands

```bash
# Makefile Commands (Recommended)
make help                      # Show all commands
make deploy                    # Deploy all apps
make deploy-backend            # Deploy backend only
make deploy-web                # Deploy web only
make deploy-admin              # Deploy admin only
make build                     # Build all apps
make test                      # Run all tests
make migrate                   # Run migrations
make clean                     # Clean artifacts
make pre-deploy                # Pre-deployment checks

# Backend
cd backend
go run main.go                 # Dev server
go test ./...                  # Tests
go build                       # Build
go run cmd/migrate/main.go     # Run migrations

# Frontend (Web)
cd frontend
npm run dev                    # Dev server
npm run build                  # Build + type check
npm run lint                   # Lint

# Admin Console
cd admin-console
npm run dev                    # Dev server
npm run build                  # Build + type check
npm run lint                   # Lint

# Database
psql $DATABASE_URL             # Connect
psql $DATABASE_URL -c "SELECT 1"  # Test

# Fly.io
fly logs                       # View logs
fly ssh console                # SSH
```

---

## Code Style

- **Go**: Use `gofmt`, follow standard conventions
- **TypeScript**: Use Prettier, follow ESLint rules
- **SQL**: Uppercase keywords, parameterized queries
- **Components**: Kebab-case filenames, PascalCase exports

---

## Security Checklist

- [ ] All queries filter by organization_id
- [ ] Use parameterized SQL queries ($1, $2)
- [ ] Verify user role for admin endpoints
- [ ] Use authenticatedApiClient for API calls
- [ ] Never expose JWT_SECRET or private keys
- [ ] Validate all user inputs

---

## Performance Tips

- Add database indexes for foreign keys
- Use React Query caching (5-min stale time)
- Optimize SQL with CTEs and FILTER clauses
- Use Next.js Image component for images
- Lazy load heavy components

---

## Getting Help

1. Check `DEVELOPER_GUIDE.md` for detailed patterns
2. Check `FEATURES_IMPLEMENTED.md` for feature details
3. Check `.kiro/specs/` for feature specifications
4. Review existing code for examples

---

**Documentation**:

- `DEVELOPER_GUIDE.md` - Detailed development guide
- `FEATURES_IMPLEMENTED.md` - Feature documentation
- `.kiro/specs/` - Feature specifications
