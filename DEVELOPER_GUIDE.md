# Liyali Gateway - Developer Guide

**Last Updated**: February 23, 2026

Quick reference for developers working on the Liyali Gateway project.

---

## Project Structure

```
liyali-gateway/
├── backend/                    # Go/Fiber backend
│   ├── cmd/                   # CLI commands (migrate, seed)
│   ├── config/                # Configuration
│   ├── database/migrations/   # SQL migrations
│   ├── handlers/              # HTTP handlers
│   ├── middleware/            # Auth, tenant, logging
│   ├── models/                # Data models
│   ├── repository/            # Database layer
│   ├── services/              # Business logic
│   ├── routes/                # Route definitions
│   └── main.go               # Entry point
│
├── frontend/                  # Next.js 16 frontend
│   ├── src/
│   │   ├── app/              # App router pages
│   │   │   ├── (auth)/       # Auth pages (login, register)
│   │   │   ├── (private)/    # Protected pages
│   │   │   └── _actions/     # Server actions
│   │   ├── components/       # React components
│   │   ├── hooks/            # Custom hooks
│   │   ├── lib/              # Utilities
│   │   ├── stores/           # Zustand stores
│   │   └── types/            # TypeScript types
│   └── public/               # Static assets
│
└── .kiro/specs/              # Feature specifications
```

---

## Tech Stack

### Backend

- **Language**: Go 1.21+
- **Framework**: Fiber v2
- **Database**: PostgreSQL 15+
- **ORM**: database/sql (no ORM)
- **Auth**: JWT tokens

### Frontend

- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript 5+
- **State**: React Query + Zustand
- **UI**: shadcn/ui + Tailwind CSS
- **Forms**: React Hook Form + Zod
- **Charts**: Recharts

---

## Development Setup

### Prerequisites

```bash
# Required
- Go 1.21+
- Node.js 20+
- PostgreSQL 15+

# Optional
- Docker
- Fly.io CLI
```

### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env
# Edit .env with your database credentials

# Run migrations
export DATABASE_URL="postgres://..."
go run cmd/migrate/main.go

# Run backend
go run main.go
# Server runs on http://localhost:8081
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Copy environment file
cp .env.example .env
# Edit .env with API URL

# Run development server
npm run dev
# App runs on http://localhost:3000
```

---

## Code Patterns

### 1. Adding a New Feature

#### Backend (Go)

**Step 1: Create Model** (`backend/models/feature.go`)

```go
package models

type Feature struct {
    ID             string    `json:"id"`
    OrganizationID string    `json:"organizationId"`
    Name           string    `json:"name"`
    CreatedAt      time.Time `json:"createdAt"`
}
```

**Step 2: Create Repository** (`backend/repository/feature_repository.go`)

```go
package repository

type FeatureRepository struct {
    db *sql.DB
}

func NewFeatureRepository(db *sql.DB) *FeatureRepository {
    return &FeatureRepository{db: db}
}

func (r *FeatureRepository) GetByOrg(ctx context.Context, orgID string) ([]models.Feature, error) {
    query := `SELECT id, organization_id, name, created_at
              FROM features WHERE organization_id = $1`
    rows, err := r.db.QueryContext(ctx, query, orgID)
    // ... scan and return
}
```

**Step 3: Create Service** (`backend/services/feature_service.go`)

```go
package services

type FeatureService struct {
    repo *repository.FeatureRepository
}

func NewFeatureService(repo *repository.FeatureRepository) *FeatureService {
    return &FeatureService{repo: repo}
}

func (s *FeatureService) GetFeatures(ctx context.Context, orgID string) ([]models.Feature, error) {
    return s.repo.GetByOrg(ctx, orgID)
}
```

**Step 4: Create Handler** (`backend/handlers/feature.go`)

```go
package handlers

type FeatureHandler struct {
    service *services.FeatureService
}

func NewFeatureHandler(service *services.FeatureService) *FeatureHandler {
    return &FeatureHandler{service: service}
}

func (h *FeatureHandler) GetFeatures(c *fiber.Ctx) error {
    tenant, err := middleware.GetTenantContext(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Unauthorized",
        })
    }

    features, err := h.service.GetFeatures(c.Context(), tenant.OrganizationID)
    if err != nil {
        return utils.SendInternalError(c, "Failed to get features", err)
    }

    return c.JSON(features)
}
```

**Step 5: Register Routes** (`backend/routes/routes.go`)

```go
func SetupRoutes(app *fiber.App, handlers *handlers.HandlerRegistry) {
    api := app.Group("/api/v1")

    // Protected routes
    protected := api.Use(middleware.AuthMiddleware())
    protected.Get("/features", handlers.Feature.GetFeatures)
}
```

**Step 6: Initialize in main.go**

```go
// Initialize repository
featureRepo := repository.NewFeatureRepository(config.PgxDB)

// Initialize service
featureService := services.NewFeatureService(featureRepo)

// Add to handler registry
handlers := handlers.NewHandlerRegistry(
    // ... other handlers
    featureService,
)
```

#### Frontend (TypeScript/React)

**Step 1: Create Types** (`frontend/src/types/feature.ts`)

```typescript
export interface Feature {
  id: string;
  organizationId: string;
  name: string;
  createdAt: string;
}
```

**Step 2: Create Server Action** (`frontend/src/app/_actions/features.ts`)

```typescript
"use server";

import authenticatedApiClient from "./api-config";
import { Feature } from "@/types/feature";

export async function getFeatures(): Promise<Feature[]> {
  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url: "/api/v1/features",
    });
    return response.data;
  } catch (error: any) {
    throw new Error(error.message || "Failed to fetch features");
  }
}
```

**Step 3: Create React Query Hook** (`frontend/src/hooks/use-feature-queries.ts`)

```typescript
"use client";

import { useQuery } from "@tanstack/react-query";
import { getFeatures } from "@/app/_actions/features";

export function useFeatures() {
  return useQuery({
    queryKey: ["features"],
    queryFn: () => getFeatures(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
```

**Step 4: Create Component** (`frontend/src/components/features/feature-list.tsx`)

```typescript
"use client";

import { useFeatures } from "@/hooks/use-feature-queries";

export function FeatureList() {
  const { data: features, isLoading, error } = useFeatures();

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {features?.map((feature) => (
        <div key={feature.id}>{feature.name}</div>
      ))}
    </div>
  );
}
```

---

## Database Migrations

### Creating a Migration

```bash
cd backend/database/migrations

# Create files (use next number)
touch 015_add_feature_table.up.sql
touch 015_add_feature_table.down.sql
```

**Up Migration** (`015_add_feature_table.up.sql`)

```sql
CREATE TABLE IF NOT EXISTS features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_features_org ON features(organization_id);
```

**Down Migration** (`015_add_feature_table.down.sql`)

```sql
DROP TABLE IF EXISTS features CASCADE;
```

### Running Migrations

```bash
cd backend
export DATABASE_URL="postgres://user:pass@host:5432/db"
go run cmd/migrate/main.go
```

---

## Authentication & Authorization

### Backend: Protecting Routes

```go
// Require authentication
protected := api.Use(middleware.AuthMiddleware())
protected.Get("/endpoint", handler.Method)

// Require admin role
func (h *Handler) AdminOnly(c *fiber.Ctx) error {
    tenant, _ := middleware.GetTenantContext(c)
    if tenant.UserRole != "admin" && tenant.UserRole != "superadmin" {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "success": false,
            "message": "Admin access required",
        })
    }
    // ... handler logic
}
```

### Frontend: Protecting Pages

```typescript
// app/(private)/admin/page.tsx
import { verifySession } from "@/lib/auth";
import { redirect } from "next/navigation";

export default async function AdminPage() {
  const { session } = await verifySession();

  if (!session?.user) {
    redirect("/login");
  }

  if (session.user.role !== "admin") {
    redirect("/dashboard");
  }

  return <AdminContent />;
}
```

---

## Common Tasks

### Adding a New API Endpoint

1. Create model in `backend/models/`
2. Create repository method in `backend/repository/`
3. Create service method in `backend/services/`
4. Create handler in `backend/handlers/`
5. Register route in `backend/routes/routes.go`
6. Initialize in `backend/main.go`

### Adding a New Page

1. Create page in `frontend/src/app/(private)/feature/page.tsx`
2. Create server action in `frontend/src/app/_actions/feature.ts`
3. Create hook in `frontend/src/hooks/use-feature-queries.ts`
4. Create components in `frontend/src/components/feature/`

### Adding a New Component

1. Create in `frontend/src/components/`
2. Use shadcn/ui components: `npx shadcn@latest add button`
3. Follow naming: `feature-name.tsx` (kebab-case)
4. Export from index if needed

---

## Testing

### Backend Tests

```bash
cd backend
go test ./...
go test -v ./handlers  # Specific package
go test -cover ./...   # With coverage
```

### Frontend Type Checking

```bash
cd frontend
npm run build  # Runs TypeScript compiler
```

---

## Debugging

### Backend Logs

```go
import "github.com/liyali/liyali-gateway/logging"

logger := logging.FromContext(c)
logger.Info("message", "key", "value")
logger.Error("error", "error", err)
```

### Frontend Logs

```typescript
console.log("Debug:", data);
console.error("Error:", error);
```

### React Query DevTools

```typescript
// Already enabled in development
// Open browser and look for React Query icon
```

---

## Common Issues

### Backend: Database Connection

```bash
# Check DATABASE_URL format
postgres://user:password@host:port/database?sslmode=require

# Test connection
psql $DATABASE_URL -c "SELECT 1"
```

### Frontend: API Connection

```bash
# Check NEXT_PUBLIC_API_URL in .env
NEXT_PUBLIC_API_URL=http://localhost:8081

# Check CORS settings in backend
FRONTEND_URL=http://localhost:3000
```

### Build Errors

```bash
# Backend
go mod tidy
go clean -cache

# Frontend
rm -rf node_modules .next
npm install
npm run build
```

---

## Code Style

### Go

- Use `gofmt` for formatting
- Follow standard Go conventions
- Use meaningful variable names
- Add comments for exported functions

### TypeScript

- Use Prettier for formatting
- Follow ESLint rules
- Use TypeScript strict mode
- Prefer functional components

### SQL

- Use uppercase for keywords
- Use parameterized queries ($1, $2)
- Add indexes for foreign keys
- Include organization_id in WHERE clauses

---

## Deployment Checklist

- [ ] Run migrations
- [ ] Update environment variables
- [ ] Build backend: `go build`
- [ ] Build frontend: `npm run build`
- [ ] Test endpoints
- [ ] Check logs
- [ ] Verify database connections
- [ ] Test authentication
- [ ] Monitor performance

---

## Useful Commands

```bash
# Backend
go run main.go                    # Run server
go test ./...                     # Run tests
go build -o app .                 # Build binary
go run cmd/migrate/main.go        # Run migrations

# Frontend
npm run dev                       # Development server
npm run build                     # Production build
npm run lint                      # Lint code
npx shadcn@latest add button      # Add UI component

# Database
psql $DATABASE_URL                # Connect to DB
psql $DATABASE_URL -c "SELECT 1"  # Test connection
psql $DATABASE_URL -f file.sql    # Run SQL file

# Fly.io
fly deploy                        # Deploy app
fly logs                          # View logs
fly ssh console                   # SSH into app
```

---

## Resources

- **Backend Framework**: https://docs.gofiber.io/
- **Frontend Framework**: https://nextjs.org/docs
- **UI Components**: https://ui.shadcn.com/
- **React Query**: https://tanstack.com/query/latest
- **Tailwind CSS**: https://tailwindcss.com/docs

---

**For feature details, see**: `FEATURES_IMPLEMENTED.md`
