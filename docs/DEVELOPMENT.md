# Development Guide

## Development Workflow

### Project Structure
```
liyali-gateway/
├── backend/                 # Go Fiber API
│   ├── cmd/                # Application entry points
│   ├── handlers/           # HTTP handlers
│   ├── services/           # Business logic
│   ├── models/             # Data models
│   ├── middleware/         # HTTP middleware
│   ├── database/           # Migrations and queries
│   └── tests/              # Test files
├── frontend/               # Next.js application
│   ├── src/app/           # App router pages
│   ├── src/components/    # React components
│   ├── src/hooks/         # Custom hooks
│   ├── src/contexts/      # React contexts
│   └── src/lib/           # Utilities
└── docs/                  # Documentation
```

## Backend Development

### Running the Backend
```bash
cd backend
go mod tidy
go run main.go
```

### Code Generation
```bash
# Generate SQLC queries
sqlc generate

# Generate database models
go generate ./...
```

### Database Migrations
```bash
cd backend/database

# Run all migrations
./migrate.sh up

# Create new migration
./migrate.sh create add_new_feature

# Reset database
./migrate.sh reset

# Rollback one migration
./migrate.sh down 1
```

### Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./services -v

# Run specific test
go test ./services -run TestUserService
```

### Adding New Features

#### 1. Create Database Migration
```sql
-- migrations/004_add_feature.up.sql
CREATE TABLE new_feature (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. Update SQLC Queries
```sql
-- queries/new_feature.sql
-- name: CreateNewFeature :one
INSERT INTO new_feature (id, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetNewFeature :one
SELECT * FROM new_feature WHERE id = $1;
```

#### 3. Create Service
```go
// services/new_feature_service.go
type NewFeatureService struct {
    db *gorm.DB
}

func (s *NewFeatureService) Create(req CreateRequest) (*models.NewFeature, error) {
    // Implementation
}
```

#### 4. Create Handler
```go
// handlers/new_feature.go
func CreateNewFeature(c *fiber.Ctx) error {
    // Implementation
}
```

#### 5. Add Routes
```go
// routes/routes.go
api.Post("/new-features", handlers.CreateNewFeature)
```

## Frontend Development

### Running the Frontend
```bash
cd frontend
npm install
npm run dev
```

### Project Structure
```
src/
├── app/                   # App router pages
│   ├── (auth)/           # Authentication pages
│   ├── (private)/        # Protected pages
│   └── layout.tsx        # Root layout
├── components/           # Reusable components
│   ├── ui/              # Base UI components
│   ├── forms/           # Form components
│   └── layout/          # Layout components
├── hooks/               # Custom React hooks
├── contexts/            # React contexts
├── lib/                 # Utilities and configs
└── types/               # TypeScript types
```

### Adding New Pages

#### 1. Create Page Component
```typescript
// app/(private)/new-feature/page.tsx
export default function NewFeaturePage() {
  return (
    <div>
      <h1>New Feature</h1>
    </div>
  );
}
```

#### 2. Create API Actions
```typescript
// app/_actions/new-feature.ts
export async function createNewFeature(data: CreateRequest) {
  const response = await fetch('/api/v1/new-features', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  return response.json();
}
```

#### 3. Create Custom Hook
```typescript
// hooks/use-new-feature.ts
export function useNewFeature() {
  return useQuery({
    queryKey: ['new-features'],
    queryFn: () => fetchNewFeatures()
  });
}
```

### Component Development

#### Base Component Structure
```typescript
interface ComponentProps {
  title: string;
  onAction?: () => void;
}

export function Component({ title, onAction }: ComponentProps) {
  return (
    <div className="p-4">
      <h2>{title}</h2>
      {onAction && (
        <button onClick={onAction}>Action</button>
      )}
    </div>
  );
}
```

#### Form Components
```typescript
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Invalid email')
});

export function FormComponent() {
  const form = useForm({
    resolver: zodResolver(schema)
  });

  const onSubmit = (data: z.infer<typeof schema>) => {
    // Handle submission
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      {/* Form fields */}
    </form>
  );
}
```

## Testing

### Backend Testing

#### Unit Tests
```go
func TestUserService_Create(t *testing.T) {
    // Setup
    service := NewUserService(mockDB)
    
    // Test
    user, err := service.Create(CreateUserRequest{
        Name: "Test User",
        Email: "test@example.com"
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Test User", user.Name)
}
```

#### Integration Tests
```go
func TestUserHandler_Create(t *testing.T) {
    // Setup test server
    app := fiber.New()
    app.Post("/users", handlers.CreateUser)
    
    // Test request
    req := httptest.NewRequest("POST", "/users", strings.NewReader(`{
        "name": "Test User",
        "email": "test@example.com"
    }`))
    req.Header.Set("Content-Type", "application/json")
    
    resp, _ := app.Test(req)
    
    // Assert
    assert.Equal(t, 201, resp.StatusCode)
}
```

### Frontend Testing

#### Component Tests
```typescript
import { render, screen } from '@testing-library/react';
import { Component } from './component';

test('renders component with title', () => {
  render(<Component title="Test Title" />);
  expect(screen.getByText('Test Title')).toBeInTheDocument();
});
```

#### Hook Tests
```typescript
import { renderHook } from '@testing-library/react';
import { useNewFeature } from './use-new-feature';

test('fetches new features', async () => {
  const { result } = renderHook(() => useNewFeature());
  
  await waitFor(() => {
    expect(result.current.data).toBeDefined();
  });
});
```

### Manual Testing Checklist

#### Workflow Testing
- [ ] **Requisition Flow**: Create → Review → Approve → Complete
- [ ] **Budget Flow**: Create → Validate → Approve → Activate
- [ ] **Purchase Order Flow**: Create → Vendor Review → Approve → Send
- [ ] **Payment Voucher Flow**: Create → Financial Review → Approve → Pay
- [ ] **GRN Flow**: Receive → Quality Check → Approve → Complete

#### Authentication Testing
- [ ] **Login/Logout**: Test with all user roles
- [ ] **Permission Checks**: Verify access controls work
- [ ] **Token Refresh**: Test automatic token renewal
- [ ] **Session Management**: Test session timeout

#### Multi-tenancy Testing
- [ ] **Organization Isolation**: Verify data separation
- [ ] **Member Management**: Add/remove organization members
- [ ] **Role Assignment**: Test custom role creation and assignment

## Code Quality

### Linting and Formatting

#### Backend (Go)
```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

#### Frontend (TypeScript)
```bash
# Lint code
npm run lint

# Format code
npm run format

# Type check
npm run type-check
```

### Pre-commit Hooks
```bash
# Install pre-commit hooks
npm run prepare

# Run all checks
npm run pre-commit
```

## Environment Configuration

### Backend Environment Variables
```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/db
DATABASE_MAX_CONNECTIONS=25

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# Server
PORT=8080
ENVIRONMENT=development

# CORS
CORS_ORIGINS=http://localhost:3000

# Logging
LOG_LEVEL=info
```

### Frontend Environment Variables
```bash
# API
NEXT_PUBLIC_API_URL=http://localhost:8080

# Authentication
NEXT_PUBLIC_JWT_EXPIRATION=86400

# Features
NEXT_PUBLIC_ENABLE_ANALYTICS=true
```

## Debugging

### Backend Debugging
```go
// Add debug logging
log.Printf("Debug: %+v", data)

// Use debugger
import "github.com/go-delve/delve/service/debugger"
```

### Frontend Debugging
```typescript
// Console debugging
console.log('Debug:', data);

// React DevTools
// Install React Developer Tools browser extension

// Network debugging
// Use browser DevTools Network tab
```

## Performance Optimization

### Backend Optimization
- Use database indexes for frequent queries
- Implement connection pooling
- Add caching for expensive operations
- Use pagination for large datasets
- Optimize SQL queries

### Frontend Optimization
- Use React.memo for expensive components
- Implement code splitting with dynamic imports
- Optimize images and assets
- Use TanStack Query for efficient data fetching
- Minimize bundle size

## Deployment

### Development Deployment
```bash
# Backend
cd backend
go build -o liyali-gateway
./liyali-gateway

# Frontend
cd frontend
npm run build
npm start
```

### Docker Deployment
```bash
# Build and run with Docker Compose
docker-compose up --build

# Run migrations
docker-compose exec backend ./migrate.sh up
```

### Production Considerations
- Set secure JWT secrets
- Configure HTTPS
- Set up proper logging
- Configure database backups
- Set up monitoring and alerts
- Use environment-specific configurations

---

**Next**: See individual component documentation for specific implementation details